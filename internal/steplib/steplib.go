package steplib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bitrise-io/stepman/models"
	"github.com/lszucs/build-aggregator/internal/bitrise"
	"github.com/lszucs/build-aggregator/internal/webhook"
)

// FetchSteps loads the step specs found at the configured path
func FetchSteps(steplibSpecURL string) ([]models.StepModel, error) {
	resp, err := http.Get(steplibSpecURL)
	if err != nil {
		return []models.StepModel{}, fmt.Errorf("getting step spec.json: %s", err)
	}

	if resp.StatusCode != 200 {
		return []models.StepModel{}, fmt.Errorf("getting step spec.json: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []models.StepModel{}, fmt.Errorf("reading spec.json: %s", err)
	}

	var spec models.StepCollectionModel
	if err := json.Unmarshal(body, &spec); err != nil {
		return []models.StepModel{}, fmt.Errorf("deserializing spec.json %s: %s", body, err)
	}

	latests := []models.StepModel{}
	for _, step := range spec.Steps {
		latests = append(latests, getLatestVersion(step))
	}

	return latests, nil
}

func getLatestVersion(step models.StepGroupModel) models.StepModel {
	var latest models.StepModel
	for _, stepver := range step.Versions {
		latest = stepver // we know, that this will always be a map with one element
	}
	return latest
}

// FilterByOrg returns a subslice of steps filtered by the parameter, according
// the repos bgithub owner
func FilterByOrg(steps []models.StepModel, includedOrgs []string) []models.StepModel {
	filtered := []models.StepModel{}
	for _, stp := range steps {
		for _, org := range includedOrgs {
			if strings.Contains(stp.Source.Git, fmt.Sprintf("github.com/%s", org)) {
				filtered = append(filtered, stp)
				break
			}
		}
	}
	return filtered
}

 // TriggerBuild triggers a build for a bitrise app
func TriggerBuild(step models.StepModel) (bitrise.Build, error) {
	repoURL := step.Source.Git

	hooks, err := webhook.FetchAll(repoURL)
	if err != nil {
		return bitrise.Build{}, fmt.Errorf("get webhooks for %s: %s", repoURL, err)
	}

	if ambiguous, err := webhook.HasMultipleApps(hooks); err != nil {
		return bitrise.Build{}, fmt.Errorf("determine if multiple apps are connected to %s: %s", repoURL, err)
	} else if ambiguous {
		return bitrise.Build{}, fmt.Errorf("%s has multiple hooks for different bitrise apps", repoURL)
	}

	hook := webhook.DeterminePRHook(hooks)

	webhookURL, ok := hook.Config["url"].(string)
	if !ok {
		return bitrise.Build{}, fmt.Errorf("type assert error: read hook.config.url from %s", hook.Config)
	}

	appSlug := webhook.ParseAppSlug(webhookURL)

	yml, err := bitrise.FetchBitriseYML(appSlug)
	encoded := base64.StdEncoding.EncodeToString(yml)
	wf, err := bitrise.DeterminePRWorkflow(encoded)
	if err != nil {
		return bitrise.Build{}, fmt.Errorf("determine workflow based on yml %s: %s", yml, err)
	}

	buildToken := webhook.ParseBuildToken(webhookURL)
	trigResp, err := bitrise.TriggerBuild(appSlug, wf, buildToken)
	if err != nil {
		return bitrise.Build{}, fmt.Errorf("trigger %s workflow on app %s: %s", wf, appSlug, err)
	}

	if status, exists := trigResp["status"]; !exists || status != "ok" {
		return bitrise.Build{}, fmt.Errorf("trigger %s workflow on app %s: response body: %s", wf, appSlug, trigResp)
	}

	return bitrise.Build{
		Title:     *step.Title,
		App:       appSlug,
		Slug:      trigResp["build_slug"].(string),
		URL:       trigResp["build_url"].(string),
		StartedAt: time.Now(),
	}, nil
}
