// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"net/url"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

const (
	segmentEndpoint  = "/adaptive-logs/segment"
	segmentsEndpoint = "/adaptive-logs/segments"
)

func (c *Client) CreateSegment(req model.CreateSegmentRequest) (model.Segment, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return model.Segment{}, err
	}

	c.segmentMutex.Lock()
	defer c.segmentMutex.Unlock()

	var resp model.Segment
	if err := c.request("POST", segmentEndpoint, nil, body, &resp); err != nil {
		return model.Segment{}, err
	}

	return resp, nil
}

func (c *Client) ListSegments() ([]model.Segment, error) {
	c.segmentMutex.Lock()
	defer c.segmentMutex.Unlock()

	resp := []model.Segment{}
	if err := c.request("GET", segmentsEndpoint, nil, nil, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ReadSegment(id string) (model.Segment, error) {
	c.segmentMutex.Lock()
	defer c.segmentMutex.Unlock()

	params := url.Values{"segment": []string{id}}
	var resp model.Segment
	if err := c.request("GET", segmentEndpoint, params, nil, &resp); err != nil {
		return model.Segment{}, err
	}

	return resp, nil
}

func (c *Client) UpdateSegment(id string, req model.CreateSegmentRequest) (model.Segment, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return model.Segment{}, err
	}

	c.segmentMutex.Lock()
	defer c.segmentMutex.Unlock()

	params := url.Values{"segment": []string{id}}
	var resp model.Segment
	if err := c.request("PUT", segmentEndpoint, params, body, &resp); err != nil {
		return model.Segment{}, err
	}

	return resp, nil
}

func (c *Client) DeleteSegment(id string) error {
	c.segmentMutex.Lock()
	defer c.segmentMutex.Unlock()

	params := url.Values{"segment": []string{id}}
	return c.request("DELETE", segmentEndpoint, params, nil, nil)
}
