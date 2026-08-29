// SPDX-License-Identifier: MPL-2.0

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

func TestSegmentToTF(t *testing.T) {
	segment := model.Segment{
		ID:        "seg-1",
		Name:      "sre",
		Selector:  `{team="sre"}`,
		ManagedBy: "terraform",
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-02T00:00:00Z",
	}

	tf := segment.ToTF()

	assert.Equal(t, "seg-1", tf.ID.ValueString())
	assert.Equal(t, "sre", tf.Name.ValueString())
	assert.Equal(t, `{team="sre"}`, tf.Selector.ValueString())
	assert.Equal(t, "terraform", tf.ManagedBy.ValueString())
	assert.Equal(t, "2026-01-01T00:00:00Z", tf.CreatedAt.ValueString())
	assert.Equal(t, "2026-01-02T00:00:00Z", tf.UpdatedAt.ValueString())
}

func TestSegmentTFToCreateReq(t *testing.T) {
	tf := model.Segment{Name: "sre", Selector: `{team="sre"}`}.ToTF()

	req := tf.ToCreateReq()

	assert.Equal(t, "sre", req.Name)
	assert.Equal(t, `{team="sre"}`, req.Selector)
}
