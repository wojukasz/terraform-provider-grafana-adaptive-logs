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

func TestListRecommendations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/adaptive-logs/recommendations", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]model.Recommendation{
			{
				Tokens:              []string{"GET ", "/health"},
				RecommendedDropRate: 90,
				Segments: map[string]model.RecommendationSegmentStats{
					`{team="sre"}`: {RecommendedDropRate: 35},
				},
			},
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, &client.Config{APIKey: "1:token"})
	require.NoError(t, err)

	recs, err := c.ListRecommendations()
	require.NoError(t, err)
	require.Len(t, recs, 1)
	assert.Equal(t, int64(90), recs[0].RecommendedDropRate)
	assert.Equal(t, int64(35), recs[0].Segments[`{team="sre"}`].RecommendedDropRate)
}
