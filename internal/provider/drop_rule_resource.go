// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

type dropRuleResource struct {
	client *client.Client
}

var (
	_ resource.Resource                = &dropRuleResource{}
	_ resource.ResourceWithConfigure   = &dropRuleResource{}
	_ resource.ResourceWithImportState = &dropRuleResource{}
)

func newDropRuleResource() resource.Resource {
	return &dropRuleResource{}
}

func (r *dropRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dropRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_drop_rule", req.ProviderTypeName)
}

func (r *dropRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Adaptive Logs drop rule: a deterministic rule that drops a percentage of log lines matching a stream selector. Scope it to a segment via `segment_id` to express a \"segment override\", or use `__global__` for a tenant-wide rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the drop rule, set by the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"segment_id": schema.StringAttribute{
				Required:    true,
				Description: "The segment this rule belongs to. Use `__global__` for a tenant-wide rule, or a `grafana-adaptive-logs_segment` resource's `id` to scope it to that segment.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the drop rule.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Optimistic concurrency version, set by the server and incremented on every update.",
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the drop rule is currently disabled.",
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Description: "RFC 3339 timestamp after which this drop rule no longer applies.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the drop rule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the drop rule was last updated.",
			},
			"body": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The rule definition.",
				Attributes: map[string]schema.Attribute{
					"stream_selector": schema.StringAttribute{
						Required:    true,
						Description: "A LogQL stream selector defining which log streams this rule applies to.",
					},
					"drop_rate": schema.Int64Attribute{
						Required:    true,
						Description: "Percentage (0-100) of matching lines to drop.",
					},
					"levels": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Log levels to match, e.g. [\"info\", \"debug\"]. Matches all levels if unset.",
					},
					"log_line_contains": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Strings that must appear in the log line for this rule to apply.",
					},
				},
			},
		},
	}
}

func (r *dropRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.DropRuleTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.CreateDropRule(plan.ToAPIReq())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create drop rule", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, rule.ToTF())...)
}

func (r *dropRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.DropRuleTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.ReadDropRule(state.ID.ValueString())
	if err != nil {
		if client.IsErrNotFound(err) {
			resp.Diagnostics.AddWarning("Drop rule not found", err.Error())
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read drop rule", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, rule.ToTF())...)
}

func (r *dropRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.DropRuleTF
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state model.DropRuleTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := plan.ToAPIReq()
	apiReq.Version = state.Version.ValueInt64()

	rule, err := r.client.UpdateDropRule(state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update drop rule", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, rule.ToTF())...)
}

func (r *dropRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.DropRuleTF
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDropRule(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete drop rule", err.Error())
	}
}

func (r *dropRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
