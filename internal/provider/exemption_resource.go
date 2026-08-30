// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

type exemptionResource struct {
	client *client.Client
}

var (
	_ resource.Resource                = &exemptionResource{}
	_ resource.ResourceWithConfigure   = &exemptionResource{}
	_ resource.ResourceWithImportState = &exemptionResource{}
)

func newExemptionResource() resource.Resource {
	return &exemptionResource{}
}

func (r *exemptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *exemptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_exemption", req.ProviderTypeName)
}

func (r *exemptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Adaptive Logs exemption: a LogQL stream selector whose matching logs are never dropped, regardless of drop rules or recommendations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the exemption, set by the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stream_selector": schema.StringAttribute{
				Required:    true,
				Description: "A LogQL stream selector matching logs that should never be dropped.",
			},
			"reason": schema.StringAttribute{
				Optional:    true,
				Description: "Business reason for this exemption.",
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Description: "RFC 3339 timestamp after which this exemption no longer applies.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the exemption was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the exemption was last updated.",
			},
		},
	}
}

func (r *exemptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ExemptionTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exemption, err := r.client.CreateExemption(plan.ToAPIReq())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create exemption", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, exemption.ToTF())...)
}

func (r *exemptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.ExemptionTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exemption, err := r.client.ReadExemption(state.ID.ValueString())
	if err != nil {
		if client.IsErrNotFound(err) {
			resp.Diagnostics.AddWarning("Exemption not found", err.Error())
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read exemption", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, exemption.ToTF())...)
}

func (r *exemptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ExemptionTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state model.ExemptionTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exemption, err := r.client.UpdateExemption(state.ID.ValueString(), plan.ToAPIReq())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update exemption", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, exemption.ToTF())...)
}

func (r *exemptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.ExemptionTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteExemption(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete exemption", err.Error())
	}
}

func (r *exemptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
