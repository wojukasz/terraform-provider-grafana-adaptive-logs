// SPDX-License-Identifier: MPL-2.0

package client

import "fmt"

// lokiLabelValuesResponse is Loki's standard query-API envelope for
// GET /loki/api/v1/label/<name>/values - a different API surface than
// /adaptive-logs/*, but reachable with the same tenant/token on the same
// host.
type lokiLabelValuesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// ListLabelValues returns every distinct value Loki has seen for the given
// label across log streams, e.g. every team name if label is "team". This
// is read-only: there is no way to create/update/delete label values, they
// simply reflect whatever labels are attached to ingested logs.
func (c *Client) ListLabelValues(label string) ([]string, error) {
	path := fmt.Sprintf("/loki/api/v1/label/%s/values", label)

	var resp lokiLabelValuesResponse
	if err := c.request("GET", path, nil, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
