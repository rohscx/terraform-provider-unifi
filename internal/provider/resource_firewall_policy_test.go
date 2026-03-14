package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallPolicy_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc-fwpol")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{Config: testAccFirewallPolicyConfig(name)},
			importStep("unifi_firewall_policy.test"),
		},
	})
}

func testAccFirewallPolicyConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_policy" "test" {
  name       = %[1]q
  action     = "ALLOW"
  protocol   = "tcp"
  ip_version = "IPV4"

  source {
    zone_id              = "internal"
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["10.0.0.10"]
    port_matching_type   = "ANY"
  }

  destination {
    zone_id              = "external"
    matching_target      = "IP"
    matching_target_type = "SPECIFIC"
    ips                  = ["8.8.8.8"]
    port_matching_type   = "SPECIFIC"
    port                 = "443"
  }

  schedule {
    mode = "ALWAYS"
  }
}
`, name)
}
