// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

type recommendationsDataSource struct {
	client *client.Client
}

var (
	_ datasource.DataSource              = &recommendationsDataSource{}
	_ datasource.DataSourceWithConfigure = &recommendationsDataSource{}
)

func newRecommendationsDataSource() datasource.DataSource {
	return &recommendationsDataSource{}
}

func (d *recommendationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected datasource configure type",
			fmt.Sprintf("Got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *recommendationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_recommendations", req.ProviderTypeName)
}

func segmentStatsAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"volume": schema.Int64Attribute{
			Computed:    true,
			Description: "Bytes ingested for this pattern within this segment.",
		},
		"ingested_lines": schema.Int64Attribute{
			Computed:    true,
			Description: "Log lines ingested for this pattern within this segment.",
		},
		"queried_lines": schema.Int64Attribute{
			Computed:    true,
			Description: "Log lines queried for this pattern within this segment.",
		},
		"recommended_drop_rate": schema.Int64Attribute{
			Computed:    true,
			Description: "Recommended drop rate (0-100) for this pattern within this segment - the basis for a \"segment override\" drop rule.",
		},
		"locked": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether this segment's recommendation is locked against automatic changes.",
		},
	}
}

func (d *recommendationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves Adaptive Logs' system-generated pattern recommendations: detected log patterns with a suggested tenant-wide and per-segment drop rate. Read-only; regenerated asynchronously roughly every 24 hours.",
		Attributes: map[string]schema.Attribute{
			"recommendations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tokens": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The tokenized log pattern this recommendation was detected for.",
						},
						"locked": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this recommendation is locked against automatic changes.",
						},
						"configured_drop_rate": schema.Int64Attribute{
							Computed:    true,
							Description: "The drop rate currently configured for this pattern, if any.",
						},
						"volume": schema.Int64Attribute{
							Computed:    true,
							Description: "Total bytes ingested for this pattern.",
						},
						"ingested_lines": schema.Int64Attribute{
							Computed:    true,
							Description: "Total log lines ingested for this pattern.",
						},
						"queried_lines": schema.Int64Attribute{
							Computed:    true,
							Description: "Total log lines queried for this pattern.",
						},
						"recommended_drop_rate": schema.Int64Attribute{
							Computed:    true,
							Description: "Recommended tenant-wide drop rate (0-100) for this pattern.",
						},
						"superseded": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this recommendation has been superseded by a newer one.",
						},
						"levels": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Log levels observed for this pattern.",
						},
						"segments": schema.MapNestedAttribute{
							Computed:    true,
							Description: "Per-segment breakdown, keyed by the segment's LogQL selector. A non-empty entry here is the read-only signal for where a \"segment override\" drop rule would apply.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: segmentStatsAttributes(),
							},
						},
					},
				},
			},
		},
	}
}

func (d *recommendationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	recs, err := d.client.ListRecommendations()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read recommendations", err.Error())
		return
	}

	state := struct {
		Recommendations []model.RecommendationTF `tfsdk:"recommendations"`
	}{}
	for _, rec := range recs {
		state.Recommendations = append(state.Recommendations, rec.ToTF())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
