data "grafana-adaptive-logs_label_values" "teams" {
  label = "team"
}

locals {
  # Segmentation has a soft cap (informally ~50 today) enforced for
  # performance reasons, not a hard technical limit - but the guidance from
  # Grafana is to collapse related values into one segment rather than push
  # for a higher cap. Here, every "sc-*" team collapses into a single "sc"
  # segment; everything else gets its own segment.
  sc_teams    = [for t in data.grafana-adaptive-logs_label_values.teams.values : t if startswith(t, "sc")]
  other_teams = [for t in data.grafana-adaptive-logs_label_values.teams.values : t if !startswith(t, "sc")]
}

resource "grafana-adaptive-logs_segment" "sc" {
  name     = "sc"
  selector = "{team=~\"${join("|", local.sc_teams)}\"}"
}

resource "grafana-adaptive-logs_segment" "team" {
  for_each = toset(local.other_teams)
  name     = each.value
  selector = "{team=\"${each.value}\"}"
}
