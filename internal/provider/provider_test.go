// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// We expect GRAFANA_AL_API_URL and GRAFANA_AL_API_KEY to be set per-test
	// (via t.Setenv), pointing at a local fakeSegmentsAPI instance.
	providerConfig = `provider "grafana-adaptive-logs" {}`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"grafana-adaptive-logs": providerserver.NewProtocol6WithError(New("test", "unknown")()),
}
