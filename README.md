# Terraform Provider for Grafana Adaptive Logs

An unofficial, standalone Terraform provider for [Grafana Adaptive Logs](https://grafana.com/docs/grafana-cloud/observe-and-act/adaptive-telemetry/adaptive-logs/), built as a sibling to Grafana's own
[`terraform-provider-grafana-adaptive-metrics`](https://github.com/grafana/terraform-provider-grafana-adaptive-metrics), following the same structure and tooling. Adaptive Logs itself has no official
Terraform provider today, though it's on Grafana's roadmap.

- Grafana website: https://grafana.com
- Grafana Cloud website: https://grafana.com/products/cloud/
- Adaptive Logs docs: https://grafana.com/docs/grafana-cloud/observe-and-act/adaptive-telemetry/adaptive-logs/

## Status

Early / in development. Not yet published to the Terraform Registry.

### Resources

- [x] `grafana-adaptive-logs_segment`
- [x] `grafana-adaptive-logs_drop_rule` — also how "segment overrides" are expressed: a drop rule with `segment_id` set to a real segment instead of `__global__`.
- [x] `grafana-adaptive-logs_exemption`

### Data sources

- [x] `grafana-adaptive-logs_recommendations` — read-only, mirrors the equivalent data source in the Adaptive Metrics provider. A non-empty `segments` entry on a recommendation is the read-only signal for where a segment-override drop rule would make sense.

### Phase 2

- [x] `grafana-adaptive-logs_label_values` — looks up every distinct value Loki has seen for a label (e.g. every team name), via Loki's own query API rather than the Adaptive Logs management API (there's no "list teams" concept in Adaptive Logs itself). Combine with `for_each` to keep segments in sync with reality as teams are added or removed. Deliberately returns the raw value list with no grouping: **there is a segment cap** - [documented as a maximum of 50 segments for Adaptive Metrics](https://grafana.com/docs/grafana-cloud/adaptive-telemetry/adaptive-metrics/additional-configuration/adaptive-metrics-rule-segmentation/) ("Adaptive Metrics allows for a maximum of 50 segments. To increase the limit, contact Customer Support."); the same ~50 figure was confirmed for Adaptive Logs directly by Grafana Cloud support in a support thread, though it isn't stated in the public Adaptive Logs docs as of writing. It's a soft, performance-driven limit rather than a hard technical one, and Grafana's stated guidance is to collapse related values first rather than request an increase - so related values (e.g. a shared prefix like `sc-`) should be collapsed into one segment via ordinary Terraform locals/`for` expressions in your own config - see `examples/data-sources/grafana-adaptive-logs_label_values/data-source.tf`. Grouping policy is a judgment call for the team, not something baked into the provider.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Development

This repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

### Building the provider

```shell
go install
```

### Using a local build

Add the following to your `.terraformrc` to test with a local version of the provider:

```
provider_installation {
  dev_overrides {
      "registry.terraform.io/wojukasz/grafana-adaptive-logs" = "/$GOPATH/bin"
  }

  direct {}
}
```

### Testing philosophy

Acceptance tests (`internal/provider/*_test.go`) run full Terraform apply/destroy cycles through `terraform-plugin-testing`, but against an **in-memory fake** of the Adaptive Logs API (see `internal/provider/common_test.go`), never a live Grafana Cloud tenant. This means `make testacc` is safe to run anywhere, with no credentials and no risk of touching real segments/rules.

Unit tests in `internal/client/*_test.go` cover request/response handling the same way, against `httptest` servers.

Real-API correctness (does the client actually parse what Grafana Cloud returns) is checked separately and manually via a **read-only** smoke test, gated behind a build tag so it never runs as part of normal CI or `go test ./...`:

```shell
go test -tags smoke ./internal/client/... -run TestSmoke -v
```

This requires `GRAFANA_AL_API_URL` (the Loki instance URL, e.g. `https://logs-prod-035.grafana.net`) and `GRAFANA_AL_API_KEY` (`<loki-tenant-id>:<token>`) to be set. The existing Adaptive Metrics access policy token in the `adaptive-things-segmentation` project's `.env` already carries the `adaptive-logs` scope too — reuse it with the Loki tenant ID (`1497381`), not the metrics instance ID (`3003412`), as the username half of `api_key`. This smoke test only ever calls `GET`; it never creates, updates, or deletes anything against the real tenant.

Run unit tests + the fake-backed acceptance suite together:

```shell
make testacc
```

### Debugging the provider

1. Build the provider:
    ```shell
    go build -gcflags "all=-N -l" -o terraform-provider-grafana-adaptive-logs .
    ```
2. Run w/ delve:
    ```shell
    dlv exec --accept-multiclient --listen=:2345 --continue --headless ./terraform-provider-grafana-adaptive-logs -- -debug
    ```
3. Connect your IDE debugger to the delve instance (listening on port 2345).
4. The `dlv` command outputs a `TF_REATTACH_PROVIDERS` value; prepend it to the terraform command you're testing.

### Updating documentation

To generate or update documentation, run `go generate`.

### Releasing the provider

Choose the appropriate version according to semver, then:

```shell
git tag <version>
git push origin <version>
```

A GitHub Action creates and signs the release. Publishing that release to the Terraform Registry under `registry.terraform.io/wojukasz/grafana-adaptive-logs` is a separate, deliberate step — not automatic on tag push.
