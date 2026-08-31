// SPDX-License-Identifier: MPL-2.0

package client

import "github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"

const recommendationsEndpoint = "/adaptive-logs/recommendations"

// ListRecommendations returns the system-generated pattern recommendations.
// This is a read-only endpoint, regenerated asynchronously roughly every 24
// hours - there is no create/update/delete for recommendations.
func (c *Client) ListRecommendations() ([]model.Recommendation, error) {
	resp := []model.Recommendation{}
	if err := c.request("GET", recommendationsEndpoint, nil, nil, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
