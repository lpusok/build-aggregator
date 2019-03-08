package output

import (
	"fmt"
	"strings"

	"github.com/lszucs/build-aggregator/internal/bitrise"
)

// Skipped encapsulates info on why a build was skipped in the trigger phase
type Skipped struct {
	Title  string
	URL    string
	Reason string
}

// Generate returns the env vars to export
func Generate(reports []bitrise.Build, skips []Skipped) (map[string]string, error) {
	faileds := []bitrise.Build{}
	for _, r := range reports {
		if r.Info.StatusText != "success" {
			faileds = append(faileds, r)
		}
	}

	var sb strings.Builder
	for _, f := range faileds {
		sb.WriteString(fmt.Sprintf("[FAIL] %s|build: %s\n", f.Title, f.URL))
	}

	for _, s := range skips {
		sb.WriteString(fmt.Sprintf("[SKIP] %s|reason: %s\n", s.Title, s.Reason))
	}

	msg := "All scheduled builds successful"
	pretext := "*Build Succeeded!*"
	color := "#f0741f"
	if sb.Len() > 0 {
		pretext = "*Scheduled build failures!*"
		color = "#f0741f"
		msg = sb.String()
	}

	return map[string]string{
		"REPORT_TEXT":    msg,
		"REPORT_PRETEXT": pretext,
		"REPORT_COLOR":   color,
	}, nil
}
