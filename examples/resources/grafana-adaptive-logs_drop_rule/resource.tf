resource "grafana-adaptive-logs_segment" "sre" {
  name     = "SRE team"
  selector = "{team=\"sre\"}"
}

# A tenant-wide drop rule.
resource "grafana-adaptive-logs_drop_rule" "healthchecks" {
  segment_id = "__global__"
  name       = "drop noisy health checks"

  body = {
    stream_selector   = "{service_name=\"api-gateway\"}"
    drop_rate         = 90
    levels            = ["info", "debug"]
    log_line_contains = ["healthcheck"]
  }
}

# A "segment override": the same kind of rule, scoped to one segment instead
# of __global__.
resource "grafana-adaptive-logs_drop_rule" "sre_override" {
  segment_id = grafana-adaptive-logs_segment.sre.id
  name       = "SRE team override"

  body = {
    stream_selector = "{service_name=\"api-gateway\"}"
    drop_rate       = 35
  }
}
