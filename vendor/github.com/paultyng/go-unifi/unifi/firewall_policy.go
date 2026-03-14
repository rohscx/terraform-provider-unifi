package unifi

import (
	"context"
	"fmt"
)

type FirewallPolicy struct {
	ID                    string                 `json:"_id,omitempty"`
	Name                  string                 `json:"name"`
	Action                string                 `json:"action"`
	Protocol              string                 `json:"protocol"`
	IPVersion             string                 `json:"ip_version"`
	Enabled               bool                   `json:"enabled"`
	Index                 int                    `json:"index"`
	Predefined            bool                   `json:"predefined,omitempty"`
	Logging               bool                   `json:"logging"`
	Description           string                 `json:"description,omitempty"`
	ConnectionStateType   string                 `json:"connection_state_type"`
	ConnectionStates      []string               `json:"connection_states"`
	CreateAllowRespond    bool                   `json:"create_allow_respond"`
	MatchIPSec            bool                   `json:"match_ip_sec"`
	MatchIPSecType        string                 `json:"match_ip_sec_type,omitempty"`
	MatchOppositeProtocol bool                   `json:"match_opposite_protocol"`
	ICMPTypename          string                 `json:"icmp_typename"`
	ICMPV6Typename        string                 `json:"icmp_v6_typename"`
	Schedule              FirewallPolicySchedule `json:"schedule"`
	Source                FirewallPolicyEndpoint `json:"source"`
	Destination           FirewallPolicyEndpoint `json:"destination"`
}

type FirewallPolicyEndpoint struct {
	ZoneID             string   `json:"zone_id"`
	MatchingTarget     string   `json:"matching_target"`
	MatchingTargetType string   `json:"matching_target_type,omitempty"`
	IPs                []string `json:"ips,omitempty"`
	IPGroupID          string   `json:"ip_group_id,omitempty"`
	MatchOppositeIPs   bool     `json:"match_opposite_ips,omitempty"`
	PortMatchingType   string   `json:"port_matching_type"`
	Port               string   `json:"port,omitempty"`
	PortGroupID        string   `json:"port_group_id,omitempty"`
	PortRanges         []string `json:"port_ranges,omitempty"`
	MatchOpositePorts  bool     `json:"match_opposite_ports"`
	NetworkID          string   `json:"network_id,omitempty"`
	MatchMAC           bool     `json:"match_mac,omitempty"`
}

type FirewallPolicySchedule struct {
	Mode           string   `json:"mode"`
	RepeatOnDays   []string `json:"repeat_on_days,omitempty"`
	TimeAllDay     bool     `json:"time_all_day,omitempty"`
	TimeRangeStart string   `json:"time_range_start,omitempty"`
	TimeRangeEnd   string   `json:"time_range_end,omitempty"`
}

func (c *Client) ListFirewallPolicies(ctx context.Context, site string) ([]FirewallPolicy, error) {
	var respBody []FirewallPolicy
	err := c.do(ctx, "GET", fmt.Sprintf("%s/site/%s/firewall-policies", c.apiV2Path, site), nil, &respBody)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func (c *Client) GetFirewallPolicy(ctx context.Context, site, id string) (*FirewallPolicy, error) {
	var respBody FirewallPolicy
	err := c.do(ctx, "GET", fmt.Sprintf("%s/site/%s/firewall-policies/%s", c.apiV2Path, site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *Client) CreateFirewallPolicy(ctx context.Context, site string, d *FirewallPolicy) (*FirewallPolicy, error) {
	var respBody FirewallPolicy
	err := c.do(ctx, "POST", fmt.Sprintf("%s/site/%s/firewall-policies", c.apiV2Path, site), d, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *Client) UpdateFirewallPolicy(ctx context.Context, site string, d *FirewallPolicy) (*FirewallPolicy, error) {
	var respBody FirewallPolicy
	err := c.do(ctx, "PUT", fmt.Sprintf("%s/site/%s/firewall-policies/%s", c.apiV2Path, site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *Client) DeleteFirewallPolicy(ctx context.Context, site, id string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("%s/site/%s/firewall-policies/%s", c.apiV2Path, site, id), nil, nil)
}
