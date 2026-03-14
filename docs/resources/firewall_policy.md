---
page_title: "unifi_firewall_policy Resource - terraform-provider-unifi"
subcategory: ""
description: |- 
  unifi_firewall_policy manages v2 zone-based firewall policies.
---

# unifi_firewall_policy (Resource)

`unifi_firewall_policy` manages v2 zone-based firewall policies.

## Example Usage

```terraform
resource "unifi_firewall_policy" "allow_printer" {
  name       = "ALLOW PRINTER(VPN-SNET)"
  action     = "ALLOW"
  protocol   = "tcp"
  enabled    = true
  ip_version = "IPV4"

  source {
    zone_id              = data.unifi_firewall_zone.vpn.id
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["172.17.254.5"]
    port_matching_type   = "ANY"
  }

  destination {
    zone_id              = data.unifi_firewall_zone.snet.id
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["172.17.195.8"]
    port_matching_type   = "SPECIFIC"
    port                 = "631,9100"
  }

  schedule {
    mode = "ALWAYS"
  }
}
```
