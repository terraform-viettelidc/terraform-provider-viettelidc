terraform {
  required_providers {
    viettelidc = {
      source  = "viettelidc.com.vn/iac/viettelidc"
      version = ">= 0.1.0"
    }
  }
}

provider "viettelidc" {
  base_url    = "https://iac.viettelidc.com.vn"
  customer_id = "cust-001"
  token       = var.viettelidc_ovpc_token
  vpc_id      = "vpc-default"
}

variable "viettelidc_ovpc_token" {
  type      = string
  sensitive = true
}
