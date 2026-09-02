# Terraform Provider for Grafana Adaptive Logs

An unofficial, standalone Terraform provider for [Grafana Adaptive Logs](https://grafana.com/docs/grafana-cloud/observe-and-act/adaptive-telemetry/adaptive-logs/), built as a sibling to Grafana's own
[`terraform-provider-grafana-adaptive-metrics`](https://github.com/grafana/terraform-provider-grafana-adaptive-metrics), following the same structure and tooling. Adaptive Logs itself has no official
Terraform provider today, though it's on Grafana's roadmap. Not yet published to the Terraform Registry.

- Grafana website: https://grafana.com
- Grafana Cloud website: https://grafana.com/products/cloud/
- Grafana Adaptive Logs docs: https://grafana.com/docs/grafana-cloud/observe-and-act/adaptive-telemetry/adaptive-logs/

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Development

This repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

### Building the provider

Build the provider using the Go `install` command:

```shell
go install
```

This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

### Using the provider

Add the following to your `.terraformrc` to test with a local version of the provider:

```
provider_installation {
  dev_overrides {
      "registry.terraform.io/wojukasz/grafana-adaptive-logs" = "/$GOPATH/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

### Common usage patterns

#### Segment overrides

Adaptive Logs has no separate "segment override" resource. It's expressed with `grafana-adaptive-logs_drop_rule`: a rule with `segment_id` set to `__global__` applies tenant-wide, and the same resource with `segment_id` set to a real `grafana-adaptive-logs_segment.<name>.id` scopes it to just that segment. See `examples/resources/grafana-adaptive-logs_drop_rule/resource.tf`.

#### Deriving segments from label values

`grafana-adaptive-logs_label_values` looks up every distinct value Loki has seen for a label (e.g. every team name), via Loki's own query API rather than the Adaptive Logs management API — there's no "list teams" concept in Adaptive Logs itself. Combine it with `for_each` to keep segments in sync as teams are added or removed.

Segmentation has a cap: Grafana Cloud docs document [a maximum of 50 segments for Adaptive Metrics](https://grafana.com/docs/grafana-cloud/adaptive-telemetry/adaptive-metrics/additional-configuration/adaptive-metrics-rule-segmentation/) ("Adaptive Metrics allows for a maximum of 50 segments. To increase the limit, contact Customer Support."), and the same ~50 figure was confirmed for Adaptive Logs directly by Grafana Cloud support, though it isn't stated in the public Adaptive Logs docs as of writing. It's a soft, performance-driven limit rather than a hard technical one, and Grafana's guidance is to collapse related values first rather than request an increase — so `grafana-adaptive-logs_label_values` deliberately returns the raw value list with no grouping, leaving related values (e.g. a shared prefix like `sc-`) to be collapsed into one segment via ordinary Terraform locals/`for` expressions in your own config. See `examples/data-sources/grafana-adaptive-logs_label_values/data-source.tf`.

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

### Running acceptance tests

Acceptance tests (`internal/provider/*_test.go`) run full Terraform apply/destroy cycles through `terraform-plugin-testing`, but against an **in-memory fake** of the Adaptive Logs API (see `internal/provider/common_test.go`), never a live Grafana Cloud tenant. This means the following is safe to run anywhere, with no credentials and no risk of touching real segments/rules:

```shell
TF_ACC=1 go test ./...
```

Unit tests in `internal/client/*_test.go` cover request/response handling the same way, against `httptest` servers.

Real-API correctness (does the client actually parse what Grafana Cloud returns) is checked separately and manually via a **read-only** smoke test, gated behind a build tag so it never runs as part of normal CI or `go test ./...`:

```shell
go test -tags smoke ./internal/client/... -run TestSmoke -v
```

This requires `GRAFANA_AL_API_URL` (the Loki instance URL, e.g. `https://logs-prod-035.grafana.net`) and `GRAFANA_AL_API_KEY` (`<loki-tenant-id>:<token>`) to be set. This smoke test only ever calls `GET`; it never creates, updates, or deletes anything against the real tenant.

### Updating documentation

To generate or update documentation, run `go generate`.

### Releasing the provider

The Terraform registry automatically indexes all GitHub releases in this repo. To publish a new release, choose the appropriate version according to semver, then:

```shell
git tag <version>
git push origin <version>
```

A GitHub Action creates and signs the release. Publishing that release to the Terraform Registry under `registry.terraform.io/wojukasz/grafana-adaptive-logs` is a separate, deliberate step — not automatic on tag push.
