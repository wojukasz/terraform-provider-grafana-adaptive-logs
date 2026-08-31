// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/wojukasz/terraform-provider-grafana-adaptive-logs/internal/model"
)

func TestAccRecommendationsDataSource(t *testing.T) {
	CheckAccTestsEnabled(t)

	fakeAPI := newFakeAdaptiveLogsAPI()
	fakeAPI.recommendations = []model.Recommendation{
		{
			Tokens:              []string{"GET ", "/health", "<*>"},
			RecommendedDropRate: 90,
			Levels:              []string{"info"},
			Segments: map[string]model.RecommendationSegmentStats{
				`{team="sre"}`: {Volume: 123, RecommendedDropRate: 35, IngestedLines: 10, QueriedLines: 5},
			},
		},
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
data "grafana-adaptive-logs_recommendations" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.grafana-adaptive-logs_recommendations.test", "recommendations.#", "1"),
					resource.TestCheckResourceAttr("data.grafana-adaptive-logs_recommendations.test", "recommendations.0.recommended_drop_rate", "90"),
					resource.TestCheckResourceAttr("data.grafana-adaptive-logs_recommendations.test", "recommendations.0.segments.%", "1"),
					func(s *terraform.State) error {
						attrs := s.RootModule().Resources["data.grafana-adaptive-logs_recommendations.test"].Primary.Attributes
						for k, v := range attrs {
							if k == `recommendations.0.segments.{team="sre"}.recommended_drop_rate` && v == "35" {
								return nil
							}
						}
						return fmt.Errorf("expected segments map to contain sre segment with recommended_drop_rate=35, got attrs: %v", attrs)
					},
				),
			},
		},
	})
}
