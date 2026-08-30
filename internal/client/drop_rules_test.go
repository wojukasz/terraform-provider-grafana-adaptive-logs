// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

func TestCreateDropRule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/adaptive-logs/drop-rules", r.URL.Path)

		var body model.DropRuleRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.DropRule{ID: "dr-1", SegmentID: body.SegmentID, Name: body.Name, Version: 1, Body: body.Body})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	rule, err := c.CreateDropRule(model.DropRuleRequest{
		SegmentID: "__global__",
		Name:      "drop healthchecks",
		Body:      model.DropRuleBody{StreamSelector: `{service_name="api-gateway"}`, DropRate: 90},
	})
	require.NoError(t, err)
	assert.Equal(t, "dr-1", rule.ID)
	assert.Equal(t, int64(1), rule.Version)
}

func TestListDropRulesBareArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]model.DropRule{{ID: "dr-1"}, {ID: "dr-2"}})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	rules, err := c.ListDropRules()
	require.NoError(t, err)
	assert.Len(t, rules, 2)
}

func TestReadDropRuleByPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/adaptive-logs/drop-rules/dr-1", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.DropRule{ID: "dr-1", Name: "test"})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	rule, err := c.ReadDropRule("dr-1")
	require.NoError(t, err)
	assert.Equal(t, "test", rule.Name)
}

func TestDeleteDropRule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/adaptive-logs/drop-rules/dr-1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	require.NoError(t, c.DeleteDropRule("dr-1"))
}
