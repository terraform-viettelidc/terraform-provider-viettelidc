// Package csa — provider_data.go re-exports providerdata.ProviderData under
// the csa package name for backward compatibility with code that references
// csa.ProviderData. New code should import providerdata directly.
package iac

import "github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac/providerdata"

// ProviderData is the shared config struct injected into resources and data
// sources by CSAProvider.Configure(). See viettelidc/csa/providerdata for
// the canonical definition.
type ProviderData = providerdata.ProviderData
