package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/lszucs/build-aggregator/internal/util/httputil"
)

var (
	// BgCtx ...
	BgCtx    context.Context
	// GHClient ...
	GHClient *github.Client
)

// FetchAll returns the webhooks of a github repo
func FetchAll(repoURL string) ([]*github.Hook, error) {
	hooks, resp, err := GHClient.Repositories.ListHooks(BgCtx, parseRepoOwner(repoURL), parseRepoName(repoURL), nil)
	if ok, err := httputil.Check(resp.Response, err, 200); !ok {
		return nil, fmt.Errorf("get webhooks for %s: %s", repoURL, err)
	}

	return hooks, nil
}

// HasMultipleApps tells if a set of github webhooks reference
// different bitrise apps.
func HasMultipleApps(hooks []*github.Hook) (bool, error) {
	slugs := make(map[string]bool)
	if len(hooks) > 0 {
		url, ok := hooks[0].Config["url"].(string)
		if !ok {
			return false, fmt.Errorf("could not read webhook url")
		}

		slugs[ParseAppSlug(url)] = true
	}
	for _, hook := range hooks[1:] {
		url, ok := hook.Config["url"].(string)
		if !ok {
			return false, fmt.Errorf("could not read webhook url")
		}
		if _, found := slugs[ParseAppSlug(url)]; !found {
			return true, nil
		}
	}
	return false, nil
}

// DeterminePRHook returns the webhook which will be called
// during the event of a pull request
func DeterminePRHook(hooks []*github.Hook) *github.Hook {
	switch len(hooks) {
	case 0:
		return nil
	case 1:
		return hooks[0]
	default:
		for _, hook := range hooks {
			if contains(hook.Events, "pull_request") {
				return hook
			}
		}
	}

	return nil

}

// ParseAppSlug returns a bitrise app slug parsed from a webhook url
func ParseAppSlug(webhookURL string) string {
	relevantSubstr := strings.SplitAfter(webhookURL, "github/")[1]
	pieces := strings.Split(relevantSubstr, "/")
	
	return pieces[0]
}

// ParseBuildToken returns a bitrise build token parsed from a webhook url
func ParseBuildToken(webhookURL string) string {
	relevantSubstr := strings.SplitAfter(webhookURL, "github/")[1]
	pieces := strings.Split(relevantSubstr, "/")

	return pieces[1]
}

func fragments(url string) []string {
	pathInfo := strings.SplitAfter(url, "github.com/")[1]
	noSuffix := strings.TrimSuffix(pathInfo, ".git")

	return strings.Split(noSuffix, "/")
}

func parseRepoOwner(url string) string {
	return fragments(url)[0]
}

func parseRepoName(url string) string {
	return fragments(url)[1]
}

func contains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}

	return false
}
