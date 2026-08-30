resource "grafana-adaptive-logs_exemption" "login_audit" {
  stream_selector = "{service_name=\"login\"}"
  reason          = "compliance audit trail - never drop"
}
