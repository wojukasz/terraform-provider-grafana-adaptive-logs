data "grafana-adaptive-logs_recommendations" "all" {}

output "top_drop_candidates" {
  value = [
    for r in data.grafana-adaptive-logs_recommendations.all.recommendations :
    r.tokens if r.recommended_drop_rate > 50
  ]
}
