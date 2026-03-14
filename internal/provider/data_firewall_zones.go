package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFirewallZones() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_firewall_zones` lists all discoverable firewall zones.",
		ReadContext: dataFirewallZonesRead,
		Schema: map[string]*schema.Schema{
			"site": {Type: schema.TypeString, Optional: true, Computed: true},
			"zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":          {Type: schema.TypeString, Computed: true},
					"name":        {Type: schema.TypeString, Computed: true},
					"network_ids": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				}},
			},
		},
	}
}

func dataFirewallZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	site := d.Get("site").(string)
	if site == "" {
		site = c.site
	}

	zones, err := c.c.ListFirewallZones(ctx, site)
	if err != nil {
		return diag.FromErr(err)
	}
	networks, err := c.c.ListNetwork(ctx, site)
	if err != nil {
		return diag.FromErr(err)
	}

	byZone := map[string][]string{}
	for _, n := range networks {
		if n.FirewallZoneID != "" {
			byZone[n.FirewallZoneID] = append(byZone[n.FirewallZoneID], n.ID)
		}
	}

	result := make([]map[string]interface{}, 0, len(zones))
	seen := map[string]struct{}{}
	for _, z := range zones {
		seen[z.ID] = struct{}{}
		result = append(result, map[string]interface{}{"id": z.ID, "name": z.Name, "network_ids": byZone[z.ID]})
	}
	for zid, nets := range byZone {
		if _, ok := seen[zid]; ok {
			continue
		}
		result = append(result, map[string]interface{}{"id": zid, "name": "", "network_ids": nets})
	}

	d.SetId(site)
	_ = d.Set("site", site)
	_ = d.Set("zones", result)
	return nil
}
