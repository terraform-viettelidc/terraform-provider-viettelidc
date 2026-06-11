// Package main is the Terraform provider entrypoint for the ViettelIDC IaC
// provider served at address "viettelidc.com.vn/iac/viettelidc".
package main

import (
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"

	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc/iac"
)

const providerAddress = "viettelidc.com.vn/iac/viettelidc"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	serveOpts := []tf6server.ServeOpt{}
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err := tf6server.Serve(
		providerAddress,
		providerserver.NewProtocol6(iac.New()()),
		serveOpts...,
	); err != nil {
		log.Fatalf("failed to serve provider: %s", err)
	}
}
