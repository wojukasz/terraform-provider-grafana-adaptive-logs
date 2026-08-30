// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// DropRuleBody is the rule definition nested under a DropRule.
type DropRuleBody struct {
	StreamSelector  string   `json:"stream_selector"`
	DropRate        int64    `json:"drop_rate"`
	Levels          []string `json:"levels,omitempty"`
	LogLineContains []string `json:"log_line_contains,omitempty"`
}

// DropRule mirrors the Adaptive Logs drop-rule API resource.
type DropRule struct {
	ID        string       `json:"id"`
	SegmentID string       `json:"segment_id"`
	Name      string       `json:"name"`
	Version   int64        `json:"version,omitempty"`
	Disabled  bool         `json:"disabled,omitempty"`
	ExpiresAt string       `json:"expires_at,omitempty"`
	CreatedAt string       `json:"created_at,omitempty"`
	UpdatedAt string       `json:"updated_at,omitempty"`
	Body      DropRuleBody `json:"body"`
}

// DropRuleRequest is the body accepted by POST/PUT on the drop-rules endpoint.
type DropRuleRequest struct {
	SegmentID string       `json:"segment_id"`
	Name      string       `json:"name"`
	Version   int64        `json:"version,omitempty"`
	Disabled  bool         `json:"disabled"`
	ExpiresAt string       `json:"expires_at,omitempty"`
	Body      DropRuleBody `json:"body"`
}

func (d DropRule) ToTF() DropRuleTF {
	// Nil (not an empty, non-nil slice) so absent optional lists round-trip
	// as null rather than an inconsistent empty list.
	var levels []types.String
	for _, l := range d.Body.Levels {
		levels = append(levels, types.StringValue(l))
	}
	var contains []types.String
	for _, l := range d.Body.LogLineContains {
		contains = append(contains, types.StringValue(l))
	}

	return DropRuleTF{
		ID:        types.StringValue(d.ID),
		SegmentID: types.StringValue(d.SegmentID),
		Name:      types.StringValue(d.Name),
		Version:   types.Int64Value(d.Version),
		Disabled:  types.BoolValue(d.Disabled),
		ExpiresAt: stringOrNull(d.ExpiresAt),
		CreatedAt: types.StringValue(d.CreatedAt),
		UpdatedAt: types.StringValue(d.UpdatedAt),
		Body: &DropRuleBodyTF{
			StreamSelector:  types.StringValue(d.Body.StreamSelector),
			DropRate:        types.Int64Value(d.Body.DropRate),
			Levels:          levels,
			LogLineContains: contains,
		},
	}
}

// DropRuleBodyTF is the Terraform-side nested representation of a drop rule's body.
type DropRuleBodyTF struct {
	StreamSelector  types.String   `tfsdk:"stream_selector"`
	DropRate        types.Int64    `tfsdk:"drop_rate"`
	Levels          []types.String `tfsdk:"levels"`
	LogLineContains []types.String `tfsdk:"log_line_contains"`
}

// DropRuleTF is the Terraform-side representation of a drop rule.
//
// Body is a pointer because "body" is a Required nested attribute: during
// ResourceWithImportState's brief window between ImportStatePassthroughID
// (which sets only "id") and the following Read, the framework decodes a
// state where "body" is still null - a plain (non-pointer) struct can't
// represent that.
type DropRuleTF struct {
	ID        types.String    `tfsdk:"id"`
	SegmentID types.String    `tfsdk:"segment_id"`
	Name      types.String    `tfsdk:"name"`
	Version   types.Int64     `tfsdk:"version"`
	Disabled  types.Bool      `tfsdk:"disabled"`
	ExpiresAt types.String    `tfsdk:"expires_at"`
	CreatedAt types.String    `tfsdk:"created_at"`
	UpdatedAt types.String    `tfsdk:"updated_at"`
	Body      *DropRuleBodyTF `tfsdk:"body"`
}

func (d DropRuleTF) ToAPIReq() DropRuleRequest {
	req := DropRuleRequest{
		SegmentID: d.SegmentID.ValueString(),
		Name:      d.Name.ValueString(),
		Version:   d.Version.ValueInt64(),
		Disabled:  d.Disabled.ValueBool(),
		ExpiresAt: d.ExpiresAt.ValueString(),
	}
	if d.Body == nil {
		return req
	}

	levels := make([]string, 0, len(d.Body.Levels))
	for _, l := range d.Body.Levels {
		levels = append(levels, l.ValueString())
	}
	contains := make([]string, 0, len(d.Body.LogLineContains))
	for _, l := range d.Body.LogLineContains {
		contains = append(contains, l.ValueString())
	}

	req.Body = DropRuleBody{
		StreamSelector:  d.Body.StreamSelector.ValueString(),
		DropRate:        d.Body.DropRate.ValueInt64(),
		Levels:          levels,
		LogLineContains: contains,
	}
	return req
}
