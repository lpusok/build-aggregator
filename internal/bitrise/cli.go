package bitrise

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	// "time"

	"github.com/bitrise-io/go-utils/log"
)

// DeterminePRWorkflow returns the workflow mapped to run
// when a pr is opened to master
func DeterminePRWorkflow(configBase64 string) (string, error) {
	prSrcBr := "--pr-source-branch=whatever"
	prTargetBr := "--pr-target-branch=master"
	slice := []string{"bitrise", "trigger-check", prSrcBr, prTargetBr, fmt.Sprintf("--config-base64=%s", configBase64), "--format=json"}

	// use native golang Cmd: need to override env, so as to not inherit DEBUG env var
	cmd := exec.Command(slice[0], slice[1:]...)
	cmd.Env = []string{}
	out, err := cmd.CombinedOutput()
	cmdStr := strings.Join(slice, " ")
	if err != nil {
		return "", fmt.Errorf("running command %s: %s", cmdStr, err)
	}
	log.Donef("$ %s", cmdStr)
	log.Printf("out: %s", out)

	var m map[string]string
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		return "", fmt.Errorf("reading command output: %s", err)
	}
	if _, ok := m["workflow"]; !ok {
		return "", fmt.Errorf("no workflow for specified options %s %s", prSrcBr, prTargetBr)
	}

	return m["workflow"], nil
}
