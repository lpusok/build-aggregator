package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-tools/go-steputils/tools"
	"github.com/google/go-github/github"
	"github.com/lszucs/build-aggregator/internal/bitrise"
	"github.com/lszucs/build-aggregator/internal/output"
	"github.com/lszucs/build-aggregator/internal/steplib"
	"github.com/lszucs/build-aggregator/internal/webhook"
	"github.com/lszucs/build-aggregator/internal/util/looputil"
	"golang.org/x/oauth2"
)

var (
	bitriseOrgs    []string
	steplibSpecURL string
	defaultDebug   = ""
	defaultBatchSize = 5
)

func waitUntilFinished(builds []bitrise.Build) error {
	count := 0
	for count < len(builds) {
		for _, bld := range builds {
			info, err := bitrise.FetchInfo(bld.App, bld.Slug)
			if err != nil {
				return fmt.Errorf("fetch %s build %s info: %s", bld.Title, bld.URL, err)
			}
			if info.Status != 0 {
				log.Donef("%s %s finished", bld.Title, bld.URL)
				count++
			}
			bld.Info = info
			time.Sleep(time.Second * 10)
		}
	}
	return nil
}

func triggerBuilds(batch []models.StepModel) (builds []bitrise.Build, skips []output.Skipped) {
	for _, step := range batch {
		log.Infof("trigger build for %s", step.Source.Git)
		bld, err := steplib.TriggerBuild(step)
		if err != nil {
			log.Warnf("could not trigger build for %s: %s", step.Source.Git, err)
			skips = append(skips, output.Skipped{
				Title:  *step.Title,
				URL:    step.Source.Git,
				Reason: fmt.Sprintf("trigger build: %s", err),
			})
		} else {
			builds = append(builds, bld)
		}
	}
	return builds, skips
}

func githubClient(ctx *context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(*ctx, ts)
	return github.NewClient(tc)
}

func main() {
	ghtoken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if ghtoken == "" {
		log.Errorf("error: GITHUB_ACCESS_TOKEN empty")
		os.Exit(1)
	}

	webhook.BgCtx = context.Background()
	webhook.GHClient = githubClient(&webhook.BgCtx, ghtoken)

	var githubOrgs string
	var batchSize int

	flag.StringVar(&steplibSpecURL, "steplib-spec-url", "", "--steplib-spec-url=http://localhost:8088/mysteplib/spec.json")
	flag.StringVar(&githubOrgs, "github-orgs", "", "--github-orgs=bitrise-io,bitrise-steplib")
	flag.IntVar(&batchSize, "batch-size", defaultBatchSize, "--batch-size=5")
	flag.Parse()

	if steplibSpecURL == "" {
		log.Errorf("steplib spec url empty")
		os.Exit(1)
	}

	if githubOrgs == "" {
		log.Errorf("github org list empty")
		os.Exit(1)
	}

	log.Infof("fetch steplib from %s", steplibSpecURL)
	steps, err := steplib.FetchSteps(steplibSpecURL)
	if err != nil {
		log.Errorf("error getting step lib: %s", err)
		os.Exit(1)
	}

	log.Printf("filter steps by orgs: %s", githubOrgs)
	bitriseSteps := steplib.FilterByOrg(steps, strings.Split(githubOrgs, ","))

	batcher := looputil.Batcher{Collection: bitriseSteps}
	finisheds, allSkips := []bitrise.Build{}, []output.Skipped{}
	for batcher.HasNext() {
		batch := batcher.Next(batchSize)

		builds, skips := triggerBuilds(batch)
		if err := waitUntilFinished(builds); err != nil {
			log.Errorf("error waiting for builds: %s", err)
			os.Exit(1)
		}

		finisheds = append(finisheds, builds...)
		allSkips = append(allSkips, skips...)
	}

	log.Infof("generate outputs")
	output, err := output.Generate(finisheds, allSkips)
	if err != nil {
		log.Errorf("error generating output: %s", err)
		os.Exit(1)
	}

	log.Infof("Exporting output variables")
	for k, v := range output {
		if err := tools.ExportEnvironmentWithEnvman(k, v); err != nil {
			log.Errorf("error exporting output var %s value %s: %s", k, v, err)
			os.Exit(1)
		}
		log.Printf("%s: %s", k, v)
	}

}
