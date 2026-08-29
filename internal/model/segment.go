// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// Segment mirrors the Adaptive Logs segment API resource.
type Segment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Selector  string `json:"selector"`
	ManagedBy string `json:"managed_by,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	IsEarly   bool   `json:"is_early,omitempty"`
}

// CreateSegmentRequest is the body accepted by POST /adaptive-logs/segment.
type CreateSegmentRequest struct {
	Name     string `json:"name"`
	Selector string `json:"selector"`
}

func (s Segment) ToTF() SegmentTF {
	return SegmentTF{
		ID:        types.StringValue(s.ID),
		Name:      types.StringValue(s.Name),
		Selector:  types.StringValue(s.Selector),
		ManagedBy: types.StringValue(s.ManagedBy),
		CreatedAt: types.StringValue(s.CreatedAt),
		UpdatedAt: types.StringValue(s.UpdatedAt),
	}
}

// SegmentTF is the Terraform-side representation of a segment.
type SegmentTF struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Selector  types.String `tfsdk:"selector"`
	ManagedBy types.String `tfsdk:"managed_by"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (s SegmentTF) ToCreateReq() CreateSegmentRequest {
	return CreateSegmentRequest{
		Name:     s.Name.ValueString(),
		Selector: s.Selector.ValueString(),
	}
}
