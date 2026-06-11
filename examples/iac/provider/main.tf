# Story 1.1 smoke test - verifies the muxed provider responds to terraform init
# and accepts the provider block (no resources defined yet).
#
# Usage (after `make install`):
#   cd examples/iac/provider
#   terraform init
#   terraform plan   # should report "No changes" (empty config)

terraform {
  required_providers {
    viettelidc = {
      source  = "viettelidc.com.vn/iac/viettelidc"
      version = "0.1.0"
    }
  }
}

provider "viettelidc" {
  # Story 1.3 will add: base_url, customer_id, token, vpc_id
}
