# Unifi Terraform Provider (terraform-provider-unifi) — v2 Firewall Fork

Fork of [paultyng/terraform-provider-unifi](https://github.com/paultyng/terraform-provider-unifi) adding support for **v2 zone-based firewall policies** on UDM/UCG controllers.

> **Note:** You can't configure your network while connected to something that may disconnect (like WiFi). Use a hard-wired connection to your controller.

## What This Fork Adds

### New Resources
- **`unifi_firewall_policy`** — Create, read, update, and delete v2 zone-based firewall policies

### New Data Sources
- **`unifi_firewall_zone`** — Look up a firewall zone by ID or name
- **`unifi_firewall_zones`** — List all discoverable firewall zones and their associated networks
- **`unifi_firewall_policy`** (data) — Look up an existing firewall policy by ID or name

### Go-Unifi Client Additions
- `FirewallPolicy`, `FirewallPolicyEndpoint`, `FirewallPolicySchedule` types
- `FirewallZone` type with `ListFirewallZones()` (falls back to network scan if zones endpoint unavailable)
- Full CRUD: `CreateFirewallPolicy`, `GetFirewallPolicy`, `UpdateFirewallPolicy`, `DeleteFirewallPolicy`, `ListFirewallPolicies`

## Usage Example

```hcl
terraform {
  required_providers {
    unifi = {
      source  = "paultyng/unifi"
      version = "~> 0.41"
    }
  }
}

provider "unifi" {
  username       = var.unifi_username
  password       = var.unifi_password
  api_url        = "https://your-controller-ip"
  allow_insecure = true
}

# Discover available firewall zones
data "unifi_firewall_zones" "all" {}

# Create an ICMP allow policy between zones
resource "unifi_firewall_policy" "allow_ping" {
  name     = "Allow VPN to LAN ICMP"
  action   = "ALLOW"
  protocol = "icmp"
  enabled  = true

  source {
    zone_id              = "your-vpn-zone-id"
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["172.17.254.5"]
    port_matching_type   = "ANY"
  }

  destination {
    zone_id              = "your-lan-zone-id"
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["172.17.195.74"]
    port_matching_type   = "ANY"
  }
}
```

## Building from Source

```bash
git clone https://github.com/rohscx/terraform-provider-unifi.git
cd terraform-provider-unifi
go build -o terraform-provider-unifi .
```

### Dev Override (for local testing)

Add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "paultyng/unifi" = "/path/to/built/binary/directory"
  }
  direct {}
}
```

Then run `terraform plan` / `terraform apply` as normal — no `terraform init` needed with dev overrides.

### macOS Local Network Note

On macOS Sequoia (15+), Little Snitch or macOS Local Network privacy may block the provider binary from reaching your controller on the LAN. Workaround: run a local TCP proxy using system Python:

```bash
# Start proxy (system Python is pre-approved for local network)
/usr/bin/python3 -c "
import socket, threading
def fwd(s,d):
    try:
        while True:
            data = s.recv(4096)
            if not data: break
            d.sendall(data)
    except: pass
    finally: s.close(); d.close()
l = socket.socket(); l.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
l.bind(('127.0.0.1', 18443)); l.listen(5)
print('Proxy: 127.0.0.1:18443 -> YOUR_CONTROLLER:443')
while True:
    c,_ = l.accept(); r = socket.socket(); r.connect(('YOUR_CONTROLLER_IP', 443))
    threading.Thread(target=fwd, args=(c,r), daemon=True).start()
    threading.Thread(target=fwd, args=(r,c), daemon=True).start()
" &

# Set api_url to the proxy
# unifi_api_url = "https://127.0.0.1:18443"
```

## API Schema Reference

### `unifi_firewall_policy` Resource

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Policy name |
| `action` | string | yes | `ALLOW`, `REJECT`, or `DROP` |
| `protocol` | string | yes | `all`, `tcp`, `udp`, `icmp`, or `tcp_udp` |
| `enabled` | bool | no | Default: `true` |
| `ip_version` | string | no | `IPV4`, `IPV6`, or `BOTH`. Default: `IPV4` |
| `index` | int | no | Rule ordering index (computed if omitted) |
| `logging` | bool | no | Enable logging |
| `description` | string | no | Description |
| `connection_state_type` | string | no | `ALL`, `RESPOND_ONLY`, or `SPECIFIC` |
| `connection_states` | list(string) | no | e.g. `["ESTABLISHED", "RELATED"]` |
| `match_ip_sec` | bool | no | Match IPsec traffic |
| `match_ip_sec_type` | string | no | `MATCH_IP_SEC` or `MATCH_NON_IP_SEC` |
| `icmp_typename` | string | no | ICMP type filter |
| `source` | block | yes | Source endpoint (see below) |
| `destination` | block | yes | Destination endpoint (see below) |
| `schedule` | block | no | Schedule (defaults to `ALWAYS`) |

### Endpoint Block (`source` / `destination`)

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `zone_id` | string | yes | Firewall zone ID |
| `matching_target` | string | yes | `ANY`, `IP`, `CLIENT`, or `NETWORK` |
| `matching_target_type` | string | no | `SPECIFIC` (direct IPs) or `OBJECT` (firewall group) |
| `ips` | list(string) | no | IP addresses (when `matching_target_type = "SPECIFIC"`) |
| `ip_group_id` | string | no | Firewall group ID (when `matching_target_type = "OBJECT"`) |
| `port_matching_type` | string | no | `ANY`, `SPECIFIC`, or `OBJECT` |
| `port` | string | no | Comma-separated ports (e.g. `"80,443"`) |
| `port_group_id` | string | no | Port group ID |
| `network_id` | string | no | Network ID |

### Schedule Block

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `mode` | string | yes | `ALWAYS`, `EVERY_DAY`, or `EVERY_WEEK` |
| `repeat_on_days` | list(string) | no | Days for `EVERY_WEEK` |
| `time_all_day` | bool | no | All day |
| `time_range_start` | string | no | Start time |
| `time_range_end` | string | no | End time |

## Known Limitations

- **Zone names not available on UCG/UDM controllers**: The v2 `/firewall-zones` endpoint returns **404** on UCG Max (Network 10.x) and likely other UDM-based controllers. The provider falls back to discovering zone IDs from network configs, but network configs do not contain zone names — only zone IDs. This means `data.unifi_firewall_zones` returns zones with empty `name` fields, and name-based lookups in `data.unifi_firewall_zone` will not work.
- **Predefined policies are read-only**: The provider rejects management of predefined (system) policies.
- **Enum values are UPPERCASE**: `IPV4` not `IPv4`, `SPECIFIC` not `specific`, `ALLOW` not `allow`.

### Workaround: Zone ID Mapping

Since zone IDs are stable but not discoverable by name, define them in a `locals` block. You can find zone IDs by inspecting existing firewall policies in the Unifi UI or via the API (`/proxy/network/v2/api/site/default/firewall-policies`).

```hcl
locals {
  zones = {
    internal = "6787fdf03cdacc306b0322fe"
    wan      = "6787fdf03cdacc306b0322ff"
    gateway  = "6787fdf03cdacc306b032300"
    vpn      = "6787fdf03cdacc306b032301"
    guest    = "6787fdf03cdacc306b032302"
    # Add custom zones as needed
  }
}

resource "unifi_firewall_policy" "example" {
  name     = "Allow VPN to LAN"
  action   = "ALLOW"
  protocol = "icmp"
  enabled  = true

  source {
    zone_id              = local.zones.vpn
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["172.17.254.5"]
    port_matching_type   = "ANY"
  }

  destination {
    zone_id              = local.zones.internal
    matching_target      = "ANY"
    port_matching_type   = "ANY"
  }
}
```

To discover your zone IDs, query the firewall policies API and correlate zone IDs with networks:

```bash
# Get all policies (zone IDs are in source.zone_id / destination.zone_id)
curl -sk "https://YOUR_CONTROLLER/proxy/network/v2/api/site/default/firewall-policies" \
  -b "TOKEN=..." | python3 -m json.tool

# Get networks (each has a firewall_zone_id field)
curl -sk "https://YOUR_CONTROLLER/proxy/network/api/s/default/rest/networkconf" \
  -b "TOKEN=..." | python3 -c "
import json, sys
from collections import defaultdict
data = json.load(sys.stdin).get('data', [])
zones = defaultdict(list)
for n in data:
    zid = n.get('zone_id') or n.get('firewall_zone_id', '')
    if zid: zones[zid].append(n.get('name', 'unnamed'))
for zid, nets in sorted(zones.items()):
    print(f'{zid} -> {\", \".join(nets)}')
"
```

## Tested On

- UCG Max running UniFi Network 10.1.85
- Terraform 1.14.3
- macOS (arm64) and Linux (arm64)

## Upstream

Based on [paultyng/terraform-provider-unifi](https://github.com/paultyng/terraform-provider-unifi) v0.41. All existing resources and data sources from upstream remain functional.

For upstream documentation, see the [Terraform provider registry](https://registry.terraform.io/providers/paultyng/unifi/latest/docs).
