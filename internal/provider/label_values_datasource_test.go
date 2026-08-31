// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLabelValuesDataSource(t *testing.T) {
	CheckAccTestsEnabled(t)

	fakeAPI := newFakeAdaptiveLogsAPI()
	fakeAPI.labelValues = map[string][]string{
		"team": {"sre", "checkout", "sc", "sc-anduin", "sc-oms"},
	}
	server := fakeAPI.Server()
	t.Cleanup(server.Close)

	t.Setenv("GRAFANA_AL_API_URL", server.URL)
	t.Setenv("GRAFANA_AL_API_KEY", "1:fake-token")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "grafana-adaptive-logs_label_values" "teams" {
	label = "team"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.grafana-adaptive-logs_label_values.teams", "label", "team"),
					resource.TestCheckResourceAttr("data.grafana-adaptive-logs_label_values.teams", "values.#", "5"),
					resource.TestCheckResourceAttr("data.grafana-adaptive-logs_label_values.teams", "values.0", "sre"),
				),
			},
		},
	})
}
