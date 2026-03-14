package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFirewallZone() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_firewall_zone` data source looks up a firewall zone by ID or name.",
		ReadContext: dataFirewallZoneRead,
		Schema: map[string]*schema.Schema{
			"id":          {Type: schema.TypeString, Optional: true, Computed: true, ConflictsWith: []string{"name"}},
			"name":        {Type: schema.TypeString, Optional: true, Computed: true, ConflictsWith: []string{"id"}},
			"site":        {Type: schema.TypeString, Optional: true, Computed: true},
			"network_ids": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		},
	}
}

func dataFirewallZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	id := d.Get("id").(string)
	name := d.Get("name").(string)
	for _, z := range zones {
		if (id != "" && z.ID == id) || (name != "" && z.Name == name) {
			d.SetId(z.ID)
			_ = d.Set("name", z.Name)
			_ = d.Set("site", site)
			_ = d.Set("network_ids", byZone[z.ID])
			return nil
		}
	}

	if id != "" {
		if ids, ok := byZone[id]; ok {
			d.SetId(id)
			_ = d.Set("site", site)
			_ = d.Set("network_ids", ids)
			return nil
		}
	}

	return diag.Errorf("firewall zone not found")
}
