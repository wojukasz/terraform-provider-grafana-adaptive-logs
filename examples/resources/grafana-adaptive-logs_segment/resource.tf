resource "grafana-adaptive-logs_segment" "sre" {
  name     = "SRE team"
  selector = "{team=\"sre\"}"
}
