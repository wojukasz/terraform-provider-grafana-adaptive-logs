// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
)

var _ provider.Provider = &AdaptiveLogsProvider{}

// AdaptiveLogsProvider defines the provider implementation.
type AdaptiveLogsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and run locally, and "test" when running acceptance
	// testing.
	version string

	// commit is set to the provider commit on release, "unknown" when the
	// provider is built and run locally or when running acceptance testing.
	commit string
}

// AdaptiveLogsProviderModel describes the provider data model.
type AdaptiveLogsProviderModel struct {
	URL         types.String `tfsdk:"url"`
	APIKey      types.String `tfsdk:"api_key"`
	HTTPHeaders types.Map    `tfsdk:"http_headers"`
	Retries     types.Int64  `tfsdk:"retries"`
	Debug       types.Bool   `tfsdk:"debug"`
}

func getStringOverriddenByEnvOrDefault(s types.String, envKey string, valDefault string) string {
	if val, ok := os.LookupEnv(envKey); ok {
		return val
	}
	if !s.IsNull() {
		return s.ValueString()
	}
	return valDefault
}

func getIntOverriddenByEnvOrDefault(s types.Int64, envKey string, valDefault int) (int, error) {
	if val, ok := os.LookupEnv(envKey); ok {
		return strconv.Atoi(val)
	}
	if !s.IsNull() {
		return int(s.ValueInt64()), nil
	}
	return valDefault, nil
}

func getBooleanOverriddenByEnvOrDefault(s types.Bool, envKey string, valDefault bool) (bool, error) {
	if val, ok := os.LookupEnv(envKey); ok {
		return strconv.ParseBool(val)
	}
	if !s.IsNull() {
		return s.ValueBool(), nil
	}
	return valDefault, nil
}

func (p *AdaptiveLogsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "grafana-adaptive-logs"
	resp.Version = p.version
}

func (p *AdaptiveLogsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages [Grafana Adaptive Logs](https://grafana.com/docs/grafana-cloud/observe-and-act/adaptive-telemetry/adaptive-logs/) segments, drop rules, and exemptions as code. Unofficial and standalone, built as a sibling to Grafana's own `terraform-provider-grafana-adaptive-metrics`.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Grafana Cloud's Loki API URL, e.g. `https://logs-prod-035.grafana.net`. May alternatively be set via the `GRAFANA_AL_API_URL` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Tenant ID and Access Policy Token for Grafana Cloud in the format '<tenant-id>:<token>'. The token needs the `adaptive-logs:admin` scope. May alternatively be set via the `GRAFANA_AL_API_KEY` environment variable.",
			},
			"http_headers": schema.MapAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "HTTP headers mapping keys to values used for accessing Grafana Cloud APIs. May alternatively be set via the `GRAFANA_AL_HTTP_HEADERS` environment variable in JSON format.",
				ElementType:         types.StringType,
			},
			"retries": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The amount of retries to use for Grafana Cloud API calls. Defaults to 3. May alternatively be set via the `GRAFANA_AL_RETRIES` environment variable.",
			},
			"debug": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to enable debug logging. Defaults to false.",
			},
		},
	}
}

func (p *AdaptiveLogsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg AdaptiveLogsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := getStringOverriddenByEnvOrDefault(cfg.URL, "GRAFANA_AL_API_URL", "")
	if apiURL == "" {
		resp.Diagnostics.AddError("Missing required attribute 'url'", "This may alternatively be set via the `GRAFANA_AL_API_URL` environment variable.")
		return
	}

	apiKey := getStringOverriddenByEnvOrDefault(cfg.APIKey, "GRAFANA_AL_API_KEY", "")
	debug, err := getBooleanOverriddenByEnvOrDefault(cfg.Debug, "GRAFANA_AL_DEBUG", false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse GRAFANA_AL_DEBUG", err.Error())
		return
	}
	retries, err := getIntOverriddenByEnvOrDefault(cfg.Retries, "GRAFANA_AL_RETRIES", 3)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse GRAFANA_AL_RETRIES", err.Error())
		return
	}

	httpClient := cleanhttp.DefaultClient()
	if retries > 0 {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = retries
		httpClient = retryClient.StandardClient()
	}

	httpHeaders := make(map[string]string)
	if envHeaders := os.Getenv("GRAFANA_AL_HTTP_HEADERS"); envHeaders != "" {
		if err := json.Unmarshal([]byte(envHeaders), &httpHeaders); err != nil {
			resp.Diagnostics.AddError("Failed to parse GRAFANA_AL_HTTP_HEADERS", err.Error())
			return
		}
	} else if !cfg.HTTPHeaders.IsNull() {
		for k, v := range cfg.HTTPHeaders.Elements() {
			vStr, ok := v.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Non-string value in http_headers", fmt.Sprintf("got %v for key %s", v, k))
				continue
			}
			httpHeaders[k] = vStr.ValueString()
		}
	}

	c, err := client.New(apiURL, &client.Config{
		APIKey:      apiKey,
		HTTPHeaders: httpHeaders,
		Debug:       debug,
		HttpClient:  httpClient,
		UserAgent:   fmt.Sprintf("Terraform/%s grafana-adaptive-logs-provider/%s (commit:%s)", req.TerraformVersion, p.version, p.commit),
	})
	if err != nil {
		resp.Diagnostics.AddError("Could not instantiate the API client.", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *AdaptiveLogsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newSegmentResource,
	}
}

func (p *AdaptiveLogsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string, commit string) func() provider.Provider {
	return func() provider.Provider {
		return &AdaptiveLogsProvider{
			version: version,
			commit:  commit,
		}
	}
}
