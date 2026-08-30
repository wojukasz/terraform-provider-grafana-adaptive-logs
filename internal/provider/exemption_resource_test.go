// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
)

func TestAccExemptionResource(t *testing.T) {
	CheckAccTestsEnabled(t)

	fakeAPI := newFakeAdaptiveLogsAPI()
	server := fakeAPI.Server()
	t.Cleanup(server.Close)

	t.Setenv("GRAFANA_AL_API_URL", server.URL)
	t.Setenv("GRAFANA_AL_API_KEY", "1:fake-token")

	var exemptionID string
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create + Read.
			{
				Config: providerConfig + `
resource "grafana-adaptive-logs_exemption" "test" {
	stream_selector = "{service_name=\"login\"}"
	reason          = "compliance audit trail"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafana-adaptive-logs_exemption.test", "stream_selector", `{service_name="login"}`),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_exemption.test", "reason", "compliance audit trail"),
					resource.TestCheckResourceAttrSet("grafana-adaptive-logs_exemption.test", "id"),
					func(s *terraform.State) error {
						exemptionID = s.RootModule().Resources["grafana-adaptive-logs_exemption.test"].Primary.ID
						return nil
					},
				),
			},
			// ImportState.
			{
				ResourceName:      "grafana-adaptive-logs_exemption.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update + Read.
			{
				Config: providerConfig + `
resource "grafana-adaptive-logs_exemption" "test" {
	stream_selector = "{service_name=\"login\"}"
	reason          = "updated reason"
}
`,
				Check: resource.TestCheckResourceAttr("grafana-adaptive-logs_exemption.test", "reason", "updated reason"),
			},
			// External delete; Terraform should recreate it.
			{
				PreConfig: func() {
					c, err := client.New(server.URL, &client.Config{APIKey: "1:fake-token"})
					require.NoError(t, err)
					require.NoError(t, c.DeleteExemption(exemptionID))
				},
				Config: providerConfig + `
resource "grafana-adaptive-logs_exemption" "test" {
	stream_selector = "{service_name=\"login\"}"
	reason          = "updated reason"
}
`,
				Check: resource.TestCheckResourceAttr("grafana-adaptive-logs_exemption.test", "reason", "updated reason"),
			},
		},
	})
}
