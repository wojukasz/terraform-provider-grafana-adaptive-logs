// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
)

type labelValuesDataSource struct {
	client *client.Client
}

var (
	_ datasource.DataSource              = &labelValuesDataSource{}
	_ datasource.DataSourceWithConfigure = &labelValuesDataSource{}
)

func newLabelValuesDataSource() datasource.DataSource {
	return &labelValuesDataSource{}
}

func (d *labelValuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *labelValuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_label_values", req.ProviderTypeName)
}

func (d *labelValuesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up every distinct value Loki has seen for a label across log streams (e.g. every team name if `label = \"team\"`), via Loki's standard query API rather than the Adaptive Logs management API. Read-only: values simply reflect whatever is attached to ingested logs. Deliberately returns the raw value list with no grouping: segmentation has a cap - Grafana Cloud docs document a maximum of 50 segments for Adaptive Metrics (https://grafana.com/docs/grafana-cloud/adaptive-telemetry/adaptive-metrics/additional-configuration/adaptive-metrics-rule-segmentation/), and the same ~50 figure was confirmed for Adaptive Logs directly by Grafana Cloud support (not yet published in the Adaptive Logs docs). It's a soft, performance-driven limit, not a hard technical one - Grafana's guidance is to collapse related values (e.g. a shared prefix like \"sc-\") into one segment yourself in Terraform (e.g. with a `for` expression building a regex selector) rather than creating one segment per value or requesting a limit increase.",
		Attributes: map[string]schema.Attribute{
			"label": schema.StringAttribute{
				Required:    true,
				Description: "The label name to look up values for, e.g. \"team\". Adaptive Logs segments in a tenant conventionally all key off the same label once the first segment picks one.",
			},
			"values": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Every distinct value currently seen for this label.",
			},
		},
	}
}

type labelValuesModel struct {
	Label  types.String   `tfsdk:"label"`
	Values []types.String `tfsdk:"values"`
}

func (d *labelValuesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config labelValuesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values, err := d.client.ListLabelValues(config.Label.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read label values", err.Error())
		return
	}

	state := labelValuesModel{Label: config.Label}
	for _, v := range values {
		state.Values = append(state.Values, types.StringValue(v))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
