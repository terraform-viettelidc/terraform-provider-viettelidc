//go:build smoke

// Package smoke serves only the CSA half of the provider, so that local
// smoke tests do not require a real VCD endpoint. Build with:
//
//	go build -tags smoke -o terraform-provider-viettelidc.exe ./cmd/smoke
package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"

	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac"
)

func main() {
	if err := tf6server.Serve(
		"viettelidc.com.vn/iac/viettelidc",
		providerserver.NewProtocol6(iac.New()()),
	); err != nil {
		log.Fatal(err)
	}
}
