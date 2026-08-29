// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/client"
)

// TestAccSegmentResource exercises the full create/read/update/import/delete
// lifecycle via real terraform apply/destroy cycles, but against an
// in-memory fakeSegmentsAPI rather than a live Grafana Cloud tenant.
func TestAccSegmentResource(t *testing.T) {
	CheckAccTestsEnabled(t)

	fakeAPI := newFakeAdaptiveLogsAPI()
	server := fakeAPI.Server()
	t.Cleanup(server.Close)

	t.Setenv("GRAFANA_AL_API_URL", server.URL)
	t.Setenv("GRAFANA_AL_API_KEY", "1:fake-token")

	var segmentID string
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create + Read.
			{
				Config: providerConfig + `
resource "grafana-adaptive-logs_segment" "test" {
	name     = "test segment"
	selector = "{team=\"sre\"}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafana-adaptive-logs_segment.test", "name", "test segment"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_segment.test", "selector", "{team=\"sre\"}"),
					resource.TestCheckResourceAttrSet("grafana-adaptive-logs_segment.test", "id"),
					func(s *terraform.State) error {
						segmentID = s.RootModule().Resources["grafana-adaptive-logs_segment.test"].Primary.ID
						return nil
					},
				),
			},
			// ImportState.
			{
				ResourceName:      "grafana-adaptive-logs_segment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update + Read.
			{
				Config: providerConfig + `
resource "grafana-adaptive-logs_segment" "test" {
	name     = "test segment 2"
	selector = "{team=\"sre\"}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafana-adaptive-logs_segment.test", "name", "test segment 2"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_segment.test", "selector", "{team=\"sre\"}"),
				),
			},
			// External delete of the resource; Terraform should recreate it.
			{
				PreConfig: func() {
					c, err := client.New(server.URL, &client.Config{APIKey: "1:fake-token"})
					require.NoError(t, err)
					require.NoError(t, c.DeleteSegment(segmentID))
				},
				Config: providerConfig + `
resource "grafana-adaptive-logs_segment" "test" {
	name     = "test segment 2"
	selector = "{team=\"sre\"}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("grafana-adaptive-logs_segment.test", "name", "test segment 2"),
					resource.TestCheckResourceAttr("grafana-adaptive-logs_segment.test", "selector", "{team=\"sre\"}"),
				),
			},
			// Delete is verified automatically by the test framework at the end.
		},
	})
}
