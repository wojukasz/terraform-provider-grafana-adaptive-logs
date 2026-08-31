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
)

func TestListLabelValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/loki/api/v1/label/team/values", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []string{"sre", "checkout"},
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	values, err := c.ListLabelValues("team")
	require.NoError(t, err)
	assert.Equal(t, []string{"sre", "checkout"}, values)
}
