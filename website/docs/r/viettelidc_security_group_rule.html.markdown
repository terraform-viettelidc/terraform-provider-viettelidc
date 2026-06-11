---
layout: "vcloud"
page_title: "Viettel IDC Cloud: viettelidc_security_group_rule"
sidebar_current: "docs-viettelidc-resource-security-group-rule"
description: |-
  Provides a ViettelIDC Security Group Rule.
---

# viettelidc\_security\_group\_rule

Provides an inbound or outbound rule for a Security Group on ViettelIDC. All attributes are immutable — any change forces a new resource.

~> **Note:** The API has no separate delete endpoint for rules. Destroying this resource calls the update endpoint with `action = "Delete"`.

## Example Usage

```hcl
# Inbound SSH rule
resource "viettelidc_security_group_rule" "ssh" {
  security_group_id = viettelidc_security_group.web.id
  direction         = "in"
  rule_type         = "SSH"
  source_ip         = "0.0.0.0/0"
  vpc_id            = viettelidc_vpc.main.id
}

# Custom inbound TCP rule
resource "viettelidc_security_group_rule" "app" {
  security_group_id = viettelidc_security_group.web.id
  direction         = "in"
  rule_type         = "Custom TCP"
  port              = "8080"
  source_ip         = "10.0.0.0/8"
  vpc_id            = viettelidc_vpc.main.id
}
```

## Argument Reference

* `security_group_id` - (Required, ForceNew) ID of the Security Group.
* `direction` - (Required, ForceNew) Traffic direction: `"in"` (inbound/ingress) or `"out"` (outbound/egress).
* `rule_type` - (Required, ForceNew) Rule type: `"Custom TCP"`, `"Custom UDP"`, `"All TCP"`, `"All UDP"`, `"All ICMP - IPv4"`, `"SSH"`, `"DNS"`, `"HTTP"`, `"HTTPS"`, `"IMAP"`.
* `protocol_name` - (Optional, ForceNew) Protocol name. Auto-populated from `rule_type` if not set.
* `port` - (Optional, ForceNew) Port or port range (e.g. `"80"` or `"8000-9000"`).
* `source_ip` - (Optional, ForceNew) Source CIDR (for inbound rules).
* `destination_ip` - (Optional, ForceNew) Destination CIDR (for outbound rules).
* `action` - (Optional) Action, defaults to `"New"`. Set to `"Delete"` to remove the rule.
* `is_valid` - (Optional) Whether the rule is valid.
* `vpc_id` - (Optional) VPC ID. Falls back to the provider default `vpc_id` when unset.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Rule ID.

## Import

Security Group Rules can be imported using `<security_group_id>/<rule_id>`:

```
terraform import viettelidc_security_group_rule.ssh <security_group_id>/<rule_id>
```
