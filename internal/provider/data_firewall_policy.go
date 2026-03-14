package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_firewall_policy` data source retrieves an existing custom firewall policy by ID or name.",
		ReadContext: dataFirewallPolicyRead,
		Schema: map[string]*schema.Schema{
			"id":         {Type: schema.TypeString, Optional: true, Computed: true, ConflictsWith: []string{"name"}},
			"name":       {Type: schema.TypeString, Optional: true, Computed: true, ConflictsWith: []string{"id"}},
			"site":       {Type: schema.TypeString, Optional: true, Computed: true},
			"action":     {Type: schema.TypeString, Computed: true},
			"protocol":   {Type: schema.TypeString, Computed: true},
			"ip_version": {Type: schema.TypeString, Computed: true},
			"enabled":    {Type: schema.TypeBool, Computed: true},
			"index":      {Type: schema.TypeInt, Computed: true},
		},
	}
}

func dataFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	site := d.Get("site").(string)
	if site == "" {
		site = c.site
	}

	var policyID string
	if id := d.Get("id").(string); id != "" {
		policyID = id
	} else {
		name := d.Get("name").(string)
		policies, err := c.c.ListFirewallPolicies(ctx, site)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, p := range policies {
			if !p.Predefined && p.Name == name {
				policyID = p.ID
				break
			}
		}
		if policyID == "" {
			return diag.Errorf("firewall policy not found with name %q", name)
		}
	}

	resp, err := c.c.GetFirewallPolicy(ctx, site, policyID)
	if err != nil {
		return diag.FromErr(err)
	}
	if resp.Predefined {
		return diag.Errorf("predefined firewall policies cannot be referenced by this data source")
	}

	d.SetId(resp.ID)
	_ = d.Set("site", site)
	_ = d.Set("name", resp.Name)
	_ = d.Set("action", resp.Action)
	_ = d.Set("protocol", resp.Protocol)
	_ = d.Set("ip_version", resp.IPVersion)
	_ = d.Set("enabled", resp.Enabled)
	_ = d.Set("index", resp.Index)

	return nil
}
