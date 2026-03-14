package unifi

import (
	"context"
	"fmt"
)

type FirewallZone struct {
	ID   string `json:"_id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (c *Client) ListFirewallZones(ctx context.Context, site string) ([]FirewallZone, error) {
	var zones []FirewallZone

	err := c.do(ctx, "GET", fmt.Sprintf("%s/site/%s/firewall-zones", c.apiV2Path, site), nil, &zones)
	if err == nil {
		return zones, nil
	}

	networks, netErr := c.ListNetwork(ctx, site)
	if netErr != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	zones = make([]FirewallZone, 0)
	for _, n := range networks {
		if n.FirewallZoneID == "" {
			continue
		}
		if _, ok := seen[n.FirewallZoneID]; ok {
			continue
		}
		seen[n.FirewallZoneID] = struct{}{}
		zones = append(zones, FirewallZone{ID: n.FirewallZoneID})
	}

	return zones, nil
}
