// SPDX-License-Identifier: MPL-2.0

package model_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

func TestDropRuleToTFAndBack(t *testing.T) {
	rule := model.DropRule{
		ID:        "dr-1",
		SegmentID: "__global__",
		Name:      "drop healthchecks",
		Version:   3,
		Disabled:  true,
		Body: model.DropRuleBody{
			StreamSelector:  `{service_name="api-gateway"}`,
			DropRate:        90,
			Levels:          []string{"info", "debug"},
			LogLineContains: []string{"healthcheck"},
		},
	}

	tf := rule.ToTF()
	assert.Equal(t, "dr-1", tf.ID.ValueString())
	assert.Equal(t, int64(3), tf.Version.ValueInt64())
	assert.True(t, tf.Disabled.ValueBool())
	assert.Equal(t, []types.String{types.StringValue("info"), types.StringValue("debug")}, tf.Body.Levels)

	req := tf.ToAPIReq()
	assert.Equal(t, "__global__", req.SegmentID)
	assert.Equal(t, int64(90), req.Body.DropRate)
	assert.Equal(t, []string{"info", "debug"}, req.Body.Levels)
	assert.Equal(t, []string{"healthcheck"}, req.Body.LogLineContains)
}
