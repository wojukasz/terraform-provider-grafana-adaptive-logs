// SPDX-License-Identifier: MPL-2.0

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

func TestExemptionToTFAndBack(t *testing.T) {
	exemption := model.Exemption{
		ID:             "ex-1",
		StreamSelector: `{service_name="login"}`,
		Reason:         "compliance",
		ExpiresAt:      "2026-01-01T00:00:00Z",
	}

	tf := exemption.ToTF()
	assert.Equal(t, "ex-1", tf.ID.ValueString())
	assert.Equal(t, "compliance", tf.Reason.ValueString())

	req := tf.ToAPIReq()
	assert.Equal(t, `{service_name="login"}`, req.StreamSelector)
	assert.Equal(t, "compliance", req.Reason)
	assert.Equal(t, "2026-01-01T00:00:00Z", req.ExpiresAt)
}
