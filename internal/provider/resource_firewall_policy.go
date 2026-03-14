package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/paultyng/go-unifi/unifi"
)

func resourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "`unifi_firewall_policy` manages a v2 zone-based firewall policy.",
		CreateContext: resourceFirewallPolicyCreate,
		ReadContext:   resourceFirewallPolicyRead,
		UpdateContext: resourceFirewallPolicyUpdate,
		DeleteContext: resourceFirewallPolicyDelete,
		Importer:      &schema.ResourceImporter{StateContext: importSiteAndID},
		Schema: map[string]*schema.Schema{
			"id":                      {Type: schema.TypeString, Computed: true},
			"site":                    {Type: schema.TypeString, Optional: true, Computed: true, ForceNew: true},
			"name":                    {Type: schema.TypeString, Required: true},
			"action":                  {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"ALLOW", "REJECT", "DROP"}, false)},
			"protocol":                {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"all", "tcp", "udp", "icmp", "tcp_udp"}, false)},
			"ip_version":              {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringInSlice([]string{"IPV4", "IPV6", "BOTH"}, false)},
			"enabled":                 {Type: schema.TypeBool, Optional: true, Default: true},
			"index":                   {Type: schema.TypeInt, Optional: true, Computed: true},
			"logging":                 {Type: schema.TypeBool, Optional: true, Computed: true},
			"description":             {Type: schema.TypeString, Optional: true},
			"connection_state_type":   {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringInSlice([]string{"ALL", "RESPOND_ONLY", "SPECIFIC"}, false)},
			"connection_states":       {Type: schema.TypeList, Optional: true, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"create_allow_respond":    {Type: schema.TypeBool, Optional: true, Computed: true},
			"match_ip_sec":            {Type: schema.TypeBool, Optional: true, Computed: true},
			"match_ip_sec_type":       {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringInSlice([]string{"MATCH_IP_SEC", "MATCH_NON_IP_SEC"}, false)},
			"match_opposite_protocol": {Type: schema.TypeBool, Optional: true, Computed: true},
			"icmp_typename":           {Type: schema.TypeString, Optional: true, Computed: true},
			"icmp_v6_typename":        {Type: schema.TypeString, Optional: true, Computed: true},
			"source":                  firewallPolicyEndpointSchema(),
			"destination":             firewallPolicyEndpointSchema(),
			"schedule": {
				Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"mode":             {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"ALWAYS", "EVERY_DAY", "EVERY_WEEK"}, false)},
					"repeat_on_days":   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
					"time_all_day":     {Type: schema.TypeBool, Optional: true, Computed: true},
					"time_range_start": {Type: schema.TypeString, Optional: true, Computed: true},
					"time_range_end":   {Type: schema.TypeString, Optional: true, Computed: true},
				}},
			},
		},
	}
}

func firewallPolicyEndpointSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Required: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
		"zone_id":              {Type: schema.TypeString, Required: true},
		"matching_target":      {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"ANY", "IP", "CLIENT", "NETWORK"}, false)},
		"matching_target_type": {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringInSlice([]string{"SPECIFIC", "OBJECT"}, false)},
		"ips":                  {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"ip_group_id":          {Type: schema.TypeString, Optional: true},
		"match_opposite_ips":   {Type: schema.TypeBool, Optional: true, Computed: true},
		"port_matching_type":   {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringInSlice([]string{"ANY", "SPECIFIC", "OBJECT"}, false)},
		"port":                 {Type: schema.TypeString, Optional: true},
		"port_group_id":        {Type: schema.TypeString, Optional: true},
		"port_ranges":          {Type: schema.TypeList, Optional: true, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"match_opposite_ports": {Type: schema.TypeBool, Optional: true, Computed: true},
		"network_id":               {Type: schema.TypeString, Optional: true},
		"network_ids":              {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"match_opposite_networks":  {Type: schema.TypeBool, Optional: true, Computed: true},
		"match_mac":                {Type: schema.TypeBool, Optional: true, Computed: true},
	}}}
}

func resourceFirewallPolicyGetResourceData(d *schema.ResourceData) *unifi.FirewallPolicy {
	policy := &unifi.FirewallPolicy{
		ID:                    d.Id(),
		Name:                  d.Get("name").(string),
		Action:                d.Get("action").(string),
		Protocol:              d.Get("protocol").(string),
		IPVersion:             d.Get("ip_version").(string),
		Enabled:               d.Get("enabled").(bool),
		Index:                 d.Get("index").(int),
		Logging:               d.Get("logging").(bool),
		Description:           d.Get("description").(string),
		ConnectionStateType:   d.Get("connection_state_type").(string),
		ConnectionStates:      interfaceToStringList(d.Get("connection_states")),
		CreateAllowRespond:    d.Get("create_allow_respond").(bool),
		MatchIPSec:            d.Get("match_ip_sec").(bool),
		MatchIPSecType:        d.Get("match_ip_sec_type").(string),
		MatchOppositeProtocol: d.Get("match_opposite_protocol").(bool),
		ICMPTypename:          d.Get("icmp_typename").(string),
		ICMPV6Typename:        d.Get("icmp_v6_typename").(string),
	}
	if policy.IPVersion == "" {
		policy.IPVersion = "IPV4"
	}
	policy.Source = expandFirewallPolicyEndpoint(d.Get("source").([]interface{}))
	policy.Destination = expandFirewallPolicyEndpoint(d.Get("destination").([]interface{}))
	policy.Schedule = expandFirewallPolicySchedulePtr(d.Get("schedule").([]interface{}))
	if policy.Schedule == nil {
		policy.Schedule = &unifi.FirewallPolicySchedule{Mode: "ALWAYS"}
	}
	return policy
}

func expandFirewallPolicyEndpoint(raw []interface{}) unifi.FirewallPolicyEndpoint {
	if len(raw) == 0 || raw[0] == nil {
		return unifi.FirewallPolicyEndpoint{}
	}
	m := raw[0].(map[string]interface{})
	return unifi.FirewallPolicyEndpoint{
		ZoneID: m["zone_id"].(string), MatchingTarget: m["matching_target"].(string), MatchingTargetType: m["matching_target_type"].(string),
		IPs: interfaceToStringList(m["ips"]), IPGroupID: m["ip_group_id"].(string), MatchOppositeIPs: m["match_opposite_ips"].(bool),
		PortMatchingType: m["port_matching_type"].(string), Port: m["port"].(string), PortGroupID: m["port_group_id"].(string),
		PortRanges: interfaceToStringList(m["port_ranges"]), MatchOpositePorts: m["match_opposite_ports"].(bool), NetworkID: m["network_id"].(string), NetworkIDs: interfaceToStringList(m["network_ids"]), MatchOppositeNetworks: m["match_opposite_networks"].(bool), MatchMAC: m["match_mac"].(bool),
	}
}

func expandFirewallPolicySchedulePtr(raw []interface{}) *unifi.FirewallPolicySchedule {
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]interface{})
	return &unifi.FirewallPolicySchedule{Mode: m["mode"].(string), RepeatOnDays: interfaceToStringList(m["repeat_on_days"]), TimeAllDay: m["time_all_day"].(bool), TimeRangeStart: m["time_range_start"].(string), TimeRangeEnd: m["time_range_end"].(string)}
}

func resourceFirewallPolicySetResourceData(resp *unifi.FirewallPolicy, d *schema.ResourceData, site string) diag.Diagnostics {
	d.SetId(resp.ID)
	_ = d.Set("site", site)
	_ = d.Set("name", resp.Name)
	_ = d.Set("action", resp.Action)
	_ = d.Set("protocol", resp.Protocol)
	_ = d.Set("ip_version", resp.IPVersion)
	_ = d.Set("enabled", resp.Enabled)
	_ = d.Set("index", resp.Index)
	_ = d.Set("logging", resp.Logging)
	_ = d.Set("description", resp.Description)
	_ = d.Set("connection_state_type", resp.ConnectionStateType)
	_ = d.Set("connection_states", resp.ConnectionStates)
	_ = d.Set("create_allow_respond", resp.CreateAllowRespond)
	_ = d.Set("match_ip_sec", resp.MatchIPSec)
	_ = d.Set("match_ip_sec_type", resp.MatchIPSecType)
	_ = d.Set("match_opposite_protocol", resp.MatchOppositeProtocol)
	_ = d.Set("icmp_typename", resp.ICMPTypename)
	_ = d.Set("icmp_v6_typename", resp.ICMPV6Typename)
	_ = d.Set("source", []map[string]interface{}{flattenFirewallPolicyEndpoint(resp.Source)})
	_ = d.Set("destination", []map[string]interface{}{flattenFirewallPolicyEndpoint(resp.Destination)})
	if resp.Schedule != nil {
		_ = d.Set("schedule", []map[string]interface{}{flattenFirewallPolicySchedule(*resp.Schedule)})
	}
	return nil
}

func flattenFirewallPolicyEndpoint(e unifi.FirewallPolicyEndpoint) map[string]interface{} {
	return map[string]interface{}{"zone_id": e.ZoneID, "matching_target": e.MatchingTarget, "matching_target_type": e.MatchingTargetType, "ips": e.IPs, "ip_group_id": e.IPGroupID, "match_opposite_ips": e.MatchOppositeIPs, "port_matching_type": e.PortMatchingType, "port": e.Port, "port_group_id": e.PortGroupID, "port_ranges": e.PortRanges, "match_opposite_ports": e.MatchOpositePorts, "network_id": e.NetworkID, "network_ids": e.NetworkIDs, "match_opposite_networks": e.MatchOppositeNetworks, "match_mac": e.MatchMAC}
}

func flattenFirewallPolicySchedule(s unifi.FirewallPolicySchedule) map[string]interface{} {
	return map[string]interface{}{"mode": s.Mode, "repeat_on_days": s.RepeatOnDays, "time_all_day": s.TimeAllDay, "time_range_start": s.TimeRangeStart, "time_range_end": s.TimeRangeEnd}
}

func resourceFirewallPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	site := d.Get("site").(string)
	if site == "" {
		site = c.site
	}
	resp, err := c.c.CreateFirewallPolicy(ctx, site, resourceFirewallPolicyGetResourceData(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create firewall policy (ensure controller supports v2 firewall policies): %w", err))
	}
	if resp.Predefined {
		return diag.Errorf("predefined firewall policies cannot be managed")
	}
	return resourceFirewallPolicySetResourceData(resp, d, site)
}

func resourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	site := d.Get("site").(string)
	if site == "" {
		site = c.site
	}
	resp, err := c.c.GetFirewallPolicy(ctx, site, d.Id())
	if err != nil {
		if _, ok := err.(*unifi.NotFoundError); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if resp.Predefined {
		return diag.Errorf("predefined firewall policies cannot be managed")
	}
	return resourceFirewallPolicySetResourceData(resp, d, site)
}

func resourceFirewallPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	site := d.Get("site").(string)
	if site == "" {
		site = c.site
	}
	resp, err := c.c.UpdateFirewallPolicy(ctx, site, resourceFirewallPolicyGetResourceData(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if resp.Predefined {
		return diag.Errorf("predefined firewall policies cannot be managed")
	}
	return resourceFirewallPolicySetResourceData(resp, d, site)
}

func resourceFirewallPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	site := d.Get("site").(string)
	if site == "" {
		site = c.site
	}
	if err := c.c.DeleteFirewallPolicy(ctx, site, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func interfaceToStringList(v interface{}) []string {
	in, ok := v.([]interface{})
	if !ok || len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.(string)
	}
	return out
}
