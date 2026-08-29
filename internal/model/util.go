// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// stringOrNull returns a null String for an empty value rather than
// types.StringValue(""), so optional fields the API omits round-trip as
// null instead of an inconsistent empty string.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
