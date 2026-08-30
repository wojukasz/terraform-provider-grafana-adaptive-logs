// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

const exemptionsEndpoint = "/adaptive-logs/exemptions"

func (c *Client) CreateExemption(req model.ExemptionRequest) (model.Exemption, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return model.Exemption{}, err
	}

	var resp model.Exemption
	if err := c.request("POST", exemptionsEndpoint, nil, body, &resp); err != nil {
		return model.Exemption{}, err
	}

	return resp, nil
}

func (c *Client) ListExemptions() ([]model.Exemption, error) {
	resp := []model.Exemption{}
	if err := c.request("GET", exemptionsEndpoint, nil, nil, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ReadExemption(id string) (model.Exemption, error) {
	var resp model.Exemption
	path := fmt.Sprintf("%s/%s", exemptionsEndpoint, id)
	if err := c.request("GET", path, nil, nil, &resp); err != nil {
		return model.Exemption{}, err
	}

	return resp, nil
}

func (c *Client) UpdateExemption(id string, req model.ExemptionRequest) (model.Exemption, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return model.Exemption{}, err
	}

	var resp model.Exemption
	path := fmt.Sprintf("%s/%s", exemptionsEndpoint, id)
	if err := c.request("PUT", path, nil, body, &resp); err != nil {
		return model.Exemption{}, err
	}

	return resp, nil
}

func (c *Client) DeleteExemption(id string) error {
	path := fmt.Sprintf("%s/%s", exemptionsEndpoint, id)
	return c.request("DELETE", path, nil, nil, nil)
}
