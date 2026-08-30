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

func TestCreateExemption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/adaptive-logs/exemptions", r.URL.Path)
		var body model.ExemptionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Exemption{ID: "ex-1", StreamSelector: body.StreamSelector, Reason: body.Reason})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	exemption, err := c.CreateExemption(model.ExemptionRequest{StreamSelector: `{service_name="login"}`, Reason: "audit"})
	require.NoError(t, err)
	assert.Equal(t, "ex-1", exemption.ID)
}

// TestListExemptionsWrappedResult exercises the real API's quirk: unlike
// segments/drop-rules, exemptions are listed wrapped as {"result": [...]}.
func TestListExemptionsWrappedResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []model.Exemption{{ID: "ex-1"}, {ID: "ex-2"}},
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	exemptions, err := c.ListExemptions()
	require.NoError(t, err)
	assert.Len(t, exemptions, 2)
}

func TestReadExemptionByPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/adaptive-logs/exemptions/ex-1", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.Exemption{ID: "ex-1", Reason: "test"})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	exemption, err := c.ReadExemption("ex-1")
	require.NoError(t, err)
	assert.Equal(t, "test", exemption.Reason)
}

func TestDeleteExemption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/adaptive-logs/exemptions/ex-1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	require.NoError(t, c.DeleteExemption("ex-1"))
}
