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

type segmentResource struct {
	client *client.Client
}

var (
	_ resource.Resource                = &segmentResource{}
	_ resource.ResourceWithConfigure   = &segmentResource{}
	_ resource.ResourceWithImportState = &segmentResource{}
)

func newSegmentResource() resource.Resource {
	return &segmentResource{}
}

func (r *segmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *segmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_segment", req.ProviderTypeName)
}

func (r *segmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Adaptive Logs segment: a named grouping of log streams defined by a LogQL selector, used to scope drop rules and exemptions to a team, service, or department.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the segment, set by the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the segment.",
			},
			"selector": schema.StringAttribute{
				Required:    true,
				Description: "A LogQL stream selector that defines which log streams belong to this segment. Supports equality or multi-literal regex matching.",
			},
			"managed_by": schema.StringAttribute{
				Computed:    true,
				Description: "The system that manages this segment, if any.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the segment was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the segment was last updated.",
			},
		},
	}
}

func (r *segmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.SegmentTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	segment, err := r.client.CreateSegment(plan.ToCreateReq())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create segment", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, segment.ToTF())...)
}

func (r *segmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.SegmentTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	segment, err := r.client.ReadSegment(state.ID.ValueString())
	if err != nil {
		if client.IsErrNotFound(err) {
			resp.Diagnostics.AddWarning("Segment not found", err.Error())
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read segment", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, segment.ToTF())...)
}

func (r *segmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.SegmentTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state model.SegmentTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	segment, err := r.client.UpdateSegment(state.ID.ValueString(), plan.ToCreateReq())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update segment", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, segment.ToTF())...)
}

func (r *segmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.SegmentTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSegment(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete segment", err.Error())
	}
}

func (r *segmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
