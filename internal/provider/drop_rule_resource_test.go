// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
)

func TestAccDropRuleResource(t *testing.T) {
	CheckAccTestsEnabled(t)

	fakeAPI := newFakeAdaptiveLogsAPI()
	server := fakeAPI.Server()
	t.Cleanup(server.Close)

	t.Setenv("GRAFANA_AL_API_URL", server.URL)
	t.Setenv("GRAFANA_AL_API_KEY", "1:fake-token")

	var ruleID string
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create + Read.
			{
				Config: providerConfig + `
resource "grafana-adaptive-logs_drop_rule" "test" {
	segment_id = "__global__"
	name       = "drop healthchecks"
	body = {
		stream_selector = "{service_name=\"api-gateway\"}"
		drop_rate       = 90
		levels          = ["info", "debug"]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "segment_id", "__global__"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "name", "drop healthchecks"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "body.drop_rate", "90"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "disabled", "false"),
					resource.TestCheckResourceAttrSet("grafana-adaptive-logs_drop_rule.test", "id"),
					func(s *terraform.State) error {
						ruleID = s.RootModule().Resources["grafana-adaptive-logs_drop_rule.test"].Primary.ID
						return nil
					},
				),
			},
			// ImportState.
			{
				ResourceName:      "grafana-adaptive-logs_drop_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update + Read.
			{
				Config: providerConfig + `
resource "grafana-adaptive-logs_drop_rule" "test" {
	segment_id = "__global__"
	name       = "drop healthchecks v2"
	disabled   = true
	body = {
		stream_selector = "{service_name=\"api-gateway\"}"
		drop_rate       = 50
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "name", "drop healthchecks v2"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "disabled", "true"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "body.drop_rate", "50"),
				),
			},
			// External delete; Terraform should recreate it.
			{
				PreConfig: func() {
					c, err := client.New(server.URL, &client.Config{APIKey: "1:fake-token"})
					require.NoError(t, err)
					require.NoError(t, c.DeleteDropRule(ruleID))
				},
				Config: providerConfig + `
resource "grafana-adaptive-logs_drop_rule" "test" {
	segment_id = "__global__"
	name       = "drop healthchecks v2"
	disabled   = true
	body = {
		stream_selector = "{service_name=\"api-gateway\"}"
		drop_rate       = 50
	}
}
`,
				Check: resource.TestCheckResourceAttr("grafana-adaptive-logs_drop_rule.test", "name", "drop healthchecks v2"),
			},
		},
	})
}
