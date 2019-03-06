package webhook_test

import (
	"testing"
	"github.com/lszucs/build-aggregator/internal/webhook"
)

func TestParseAppSlug(t *testing.T) {
	want := "82ac2cda76a2755f"
	got := webhook.ParseAppSlug("https://hooks.bitrise.io/h/github/82ac2cda76a2755f/0WJcNWSv-0rO-RRAd_6ZGw")
	if got != want {
		t.Errorf("want %s got %s", want, got)
	}
}