package bitrise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/bitrise-io/go-utils/log"
)

var (
	buildInfoURL    = "https://api.bitrise.io/v0.1/apps/%s/builds/%s"
	bitriseAPIToken = os.Getenv("BITRISE_API_TOKEN")
	triggerURL      = "https://app.bitrise.io/app/%s/build/start.json"
)

type triggerRequestParams struct {
	appSlug    string
	buildToken string
}

type hookInfo struct {
	Type              string `json:"type"`
	BuildTriggerToken string `json:"build_trigger_token"`
}

type buildParams struct {
	WorkflowID string `json:"workflow_id"`
}

type triggerRequestBody struct {
	HookInfo    hookInfo    `json:"hook_info"`
	BuildParams buildParams `json:"build_params"`
}

// Build encapsulates data concerning a (triggered) build.
type Build struct {
	Title     string
	App       string
	Slug      string
	URL       string
	StartedAt time.Time
	Info      BuildInfo
}

// BuildInfo is used to unmarshal the bitrise api build info response
type BuildInfo struct {
	Status     int
	StatusText string `json:"status_text"`
}

func init() {
	if bitriseAPIToken == "" {
		log.Errorf("bitrise api token empty")
		os.Exit(1)
	}
}

// FetchBitriseYML fetches the yml configuration of a bitrise app 
func FetchBitriseYML(appSlug string) ([]byte, error) {
	url := fmt.Sprintf("https://api.bitrise.io/v0.1/apps/%s/bitrise.yml", appSlug)
	log.Printf("get bitrise yml from %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get bitrise yml request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", bitriseAPIToken))
	req.Header.Add("Content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http response %d %s", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bitrise yml: %s", err)
	}

	return body, nil
}

 // TriggerBuild triggers a build for a bitrise app
func TriggerBuild(appSlug string, workflow string, buildToken string) (map[string]interface{}, error) {
	reqBody := triggerRequestBody{
		HookInfo: hookInfo{
			Type:              "bitrise",
			BuildTriggerToken: buildToken,
		},
		BuildParams: buildParams{
			WorkflowID: workflow,
		},
	}
	marshalled, err := json.Marshal(reqBody)
	payload := bytes.NewReader(marshalled)
	url := fmt.Sprintf(triggerURL, appSlug)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("creating trigger request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", bitriseAPIToken))
	req.Header.Add("Content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("http response %d %s", resp.StatusCode, resp.Status)
	}

	var triggerResp map[string]interface{}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &triggerResp); err != nil {
		return nil, err
	}

	return triggerResp, nil
}

// FetchInfo retrieves a bitrise builds info
func FetchInfo(app string, build string) (BuildInfo, error) {
	// construct build info request
	url := fmt.Sprintf(buildInfoURL, app, build)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return BuildInfo{}, fmt.Errorf("construct build info request (url: %s): %s", url, err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", bitriseAPIToken))
	req.Header.Add("Content-type", "application/json")

	// send build info request
	log.Printf("send info request GET %s", url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return BuildInfo{}, err
	}

	if resp.StatusCode != 200 {
		return BuildInfo{}, fmt.Errorf("http response %d %s", resp.StatusCode, resp.Status)
	}

	// process build info response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BuildInfo{}, fmt.Errorf("read build info response: %s", err)
	}

	m := struct {
		Data BuildInfo
	}{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return BuildInfo{}, fmt.Errorf("unmarshal build info response: %s", err)
	}

	return m.Data, nil
}
