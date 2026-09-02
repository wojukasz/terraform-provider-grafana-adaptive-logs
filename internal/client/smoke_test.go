//go:build smoke

// SPDX-License-Identifier: MPL-2.0

// This file is a read-only smoke check against a real Adaptive Logs tenant.
// It never creates, updates, or deletes anything - it only confirms the
// client can authenticate and parse real API responses. It is excluded from
// normal `go test ./...` runs by the "smoke" build tag; run it explicitly
// with:
//
//	go test -tags smoke ./internal/client/... -run TestSmokeListSegments -v
//
// and GRAFANA_AL_API_URL / GRAFANA_AL_API_KEY set to real credentials
// (e.g. sourced from adaptive-things-segmentation's .env, using the Loki
// tenant ID - not the metrics instance ID - as the username half of
// GRAFANA_AL_API_KEY).
package client_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
)

func TestSmokeListSegments(t *testing.T) {
	apiURL := os.Getenv("GRAFANA_AL_API_URL")
	apiKey := os.Getenv("GRAFANA_AL_API_KEY")
	if apiURL == "" || apiKey == "" {
		t.Skip("GRAFANA_AL_API_URL and GRAFANA_AL_API_KEY must be set for the smoke test")
	}

	c, err := client.New(apiURL, &client.Config{APIKey: apiKey})
	require.NoError(t, err)

	segments, err := c.ListSegments()
	require.NoError(t, err)
	t.Logf("read %d real segment(s) from %s", len(segments), apiURL)
}

func TestSmokeListDropRules(t *testing.T) {
	apiURL := os.Getenv("GRAFANA_AL_API_URL")
	apiKey := os.Getenv("GRAFANA_AL_API_KEY")
	if apiURL == "" || apiKey == "" {
		t.Skip("GRAFANA_AL_API_URL and GRAFANA_AL_API_KEY must be set for the smoke test")
	}

	c, err := client.New(apiURL, &client.Config{APIKey: apiKey})
	require.NoError(t, err)

	rules, err := c.ListDropRules()
	require.NoError(t, err)
	t.Logf("read %d real drop rule(s) from %s", len(rules), apiURL)
}

func TestSmokeListExemptions(t *testing.T) {
	apiURL := os.Getenv("GRAFANA_AL_API_URL")
	apiKey := os.Getenv("GRAFANA_AL_API_KEY")
	if apiURL == "" || apiKey == "" {
		t.Skip("GRAFANA_AL_API_URL and GRAFANA_AL_API_KEY must be set for the smoke test")
	}

	c, err := client.New(apiURL, &client.Config{APIKey: apiKey})
	require.NoError(t, err)

	exemptions, err := c.ListExemptions()
	require.NoError(t, err)
	t.Logf("read %d real exemption(s) from %s", len(exemptions), apiURL)
}

func TestSmokeListRecommendations(t *testing.T) {
	apiURL := os.Getenv("GRAFANA_AL_API_URL")
	apiKey := os.Getenv("GRAFANA_AL_API_KEY")
	if apiURL == "" || apiKey == "" {
		t.Skip("GRAFANA_AL_API_URL and GRAFANA_AL_API_KEY must be set for the smoke test")
	}

	c, err := client.New(apiURL, &client.Config{APIKey: apiKey})
	require.NoError(t, err)

	recs, err := c.ListRecommendations()
	require.NoError(t, err)
	t.Logf("read %d real recommendation(s) from %s", len(recs), apiURL)
}

func TestSmokeListLabelValues(t *testing.T) {
	apiURL := os.Getenv("GRAFANA_AL_API_URL")
	apiKey := os.Getenv("GRAFANA_AL_API_KEY")
	if apiURL == "" || apiKey == "" {
		t.Skip("GRAFANA_AL_API_URL and GRAFANA_AL_API_KEY must be set for the smoke test")
	}

	c, err := client.New(apiURL, &client.Config{APIKey: apiKey})
	require.NoError(t, err)

	values, err := c.ListLabelValues("team")
	require.NoError(t, err)
	t.Logf("read %d real team value(s) from %s", len(values), apiURL)
}
