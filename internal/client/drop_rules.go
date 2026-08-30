// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

const dropRulesEndpoint = "/adaptive-logs/drop-rules"

func (c *Client) CreateDropRule(req model.DropRuleRequest) (model.DropRule, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return model.DropRule{}, err
	}

	var resp model.DropRule
	if err := c.request("POST", dropRulesEndpoint, nil, body, &resp); err != nil {
		return model.DropRule{}, err
	}

	return resp, nil
}

func (c *Client) ListDropRules() ([]model.DropRule, error) {
	resp := []model.DropRule{}
	if err := c.request("GET", dropRulesEndpoint, nil, nil, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ReadDropRule(id string) (model.DropRule, error) {
	var resp model.DropRule
	path := fmt.Sprintf("%s/%s", dropRulesEndpoint, id)
	if err := c.request("GET", path, nil, nil, &resp); err != nil {
		return model.DropRule{}, err
	}

	return resp, nil
}

func (c *Client) UpdateDropRule(id string, req model.DropRuleRequest) (model.DropRule, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return model.DropRule{}, err
	}

	var resp model.DropRule
	path := fmt.Sprintf("%s/%s", dropRulesEndpoint, id)
	if err := c.request("PUT", path, nil, body, &resp); err != nil {
		return model.DropRule{}, err
	}

	return resp, nil
}

func (c *Client) DeleteDropRule(id string) error {
	path := fmt.Sprintf("%s/%s", dropRulesEndpoint, id)
	return c.request("DELETE", path, nil, nil, nil)
}
