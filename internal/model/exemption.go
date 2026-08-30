// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// Exemption mirrors the Adaptive Logs exemption API resource.
type Exemption struct {
	ID             string `json:"id"`
	StreamSelector string `json:"stream_selector"`
	Reason         string `json:"reason,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// ExemptionRequest is the body accepted by POST/PUT on the exemptions endpoint.
type ExemptionRequest struct {
	StreamSelector string `json:"stream_selector"`
	Reason         string `json:"reason,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
}

func (e Exemption) ToTF() ExemptionTF {
	return ExemptionTF{
		ID:             types.StringValue(e.ID),
		StreamSelector: types.StringValue(e.StreamSelector),
		Reason:         stringOrNull(e.Reason),
		ExpiresAt:      stringOrNull(e.ExpiresAt),
		CreatedAt:      types.StringValue(e.CreatedAt),
		UpdatedAt:      types.StringValue(e.UpdatedAt),
	}
}

// ExemptionTF is the Terraform-side representation of an exemption.
type ExemptionTF struct {
	ID             types.String `tfsdk:"id"`
	StreamSelector types.String `tfsdk:"stream_selector"`
	Reason         types.String `tfsdk:"reason"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (e ExemptionTF) ToAPIReq() ExemptionRequest {
	return ExemptionRequest{
		StreamSelector: e.StreamSelector.ValueString(),
		Reason:         e.Reason.ValueString(),
		ExpiresAt:      e.ExpiresAt.ValueString(),
	}
}
