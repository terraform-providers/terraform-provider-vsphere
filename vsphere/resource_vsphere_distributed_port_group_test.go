package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereDistributedPortGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
				),
			},
			{
				ResourceName:            "vsphere_distributed_port_group.pg",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vlan_range"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					pg, err := testGetDVPortgroup(s, "pg")
					if err != nil {
						return "", err
					}
					return pg.InventoryPath, nil
				},
				Config: testAccResourceVSphereDistributedPortGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_inheritPolicyDiffCheck(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigPolicyInherit(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_inheritPolicyDiffCheckVlanRangeTypeSetEdition(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigPolicyInheritVLANRange(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_overrideVlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigOverrideVLAN(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(1000, 1999),
					testAccResourceVSphereDistributedPortGroupHasVlanRange(3000, 3999),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
					testAccResourceVSphereDistributedPortGroupCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_multiTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
					testAccResourceVSphereDistributedPortGroupCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_singleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
					testAccResourceVSphereDistributedPortGroupCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedPortGroup_multiCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedPortGroupPreCheck()
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
					testAccResourceVSphereDistributedPortGroupCheckCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereDistributedPortGroupConfigMultiCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedPortGroupExists(true),
					testAccResourceVSphereDistributedPortGroupCheckCustomAttributes(),
				),
			},
		},
	})
}

func testAccResourceVSphereDistributedPortGroupPreCheck() {
}

func testAccResourceVSphereDistributedPortGroupExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVPortgroup(s, "pg")
		if err != nil {
			if viapi.IsAnyNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected DVS %s to be missing", dvs.Reference().Value)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedPortGroupHasVlanRange(emin, emax int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVPortgroupProperties(s, "pg")
		if err != nil {
			return err
		}
		pc := props.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
		ranges := pc.Vlan.(*types.VmwareDistributedVirtualSwitchTrunkVlanSpec).VlanId
		var found bool
		for _, rng := range ranges {
			if rng.Start == emin && rng.End == emax {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("could not find start %d and end %d in %#v", emin, emax, ranges)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedPortGroupCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVPortgroup(s, "pg")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, dvs, tagResName)
	}
}

func testAccResourceVSphereDistributedPortGroupCheckCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVPortgroupProperties(s, "pg")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_distributed_port_group", "pg", props.Entity())
	}
}

func testAccResourceVSphereDistributedPortGroupConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigPolicyInherit() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  vlan_id = 1000

  active_uplinks  = ["uplink1", "uplink2"]
  standby_uplinks = ["uplink3", "uplink4"]
  check_beacon    = true
  failback        = true
  notify_switches = true
  teaming_policy  = "failover_explicit"

  lacp_enabled = true
  lacp_mode    = "active"

  allow_forged_transmits = true
  allow_mac_changes      = true
  allow_promiscuous      = true

  ingress_shaping_enabled           = true
  ingress_shaping_average_bandwidth = 1000000
  ingress_shaping_peak_bandwidth    = 10000000
  ingress_shaping_burst_size        = 5000000

  egress_shaping_enabled           = true
  egress_shaping_average_bandwidth = 1000000
  egress_shaping_peak_bandwidth    = 10000000
  egress_shaping_burst_size        = 5000000

  block_all_ports = true
  netflow_enabled = true
  tx_uplink       = true
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigPolicyInheritVLANRange() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  vlan_range {
		min_vlan = 1000
		max_vlan = 1999
	}
  
	vlan_range {
		min_vlan = 3000
		max_vlan = 3999
	}

  active_uplinks  = ["uplink1", "uplink2"]
  standby_uplinks = ["uplink3", "uplink4"]
  check_beacon    = true
  failback        = true
  notify_switches = true
  teaming_policy  = "failover_explicit"

  lacp_enabled = true
  lacp_mode    = "active"

  allow_forged_transmits = true
  allow_mac_changes      = true
  allow_promiscuous      = true

  ingress_shaping_enabled           = true
  ingress_shaping_average_bandwidth = 1000000
  ingress_shaping_peak_bandwidth    = 10000000
  ingress_shaping_burst_size        = 5000000

  egress_shaping_enabled           = true
  egress_shaping_average_bandwidth = 1000000
  egress_shaping_peak_bandwidth    = 10000000
  egress_shaping_burst_size        = 5000000

  block_all_ports = true
  netflow_enabled = true
  tx_uplink       = true
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigOverrideVLAN() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  
	vlan_range {
		min_vlan = 1000
		max_vlan = 1999
	}
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"

	vlan_range {
		min_vlan = 3000
		max_vlan = 3999
	}
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigSingleTag() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VmwareDistributedVirtualPortgroup",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"
  tags                            = ["${vsphere_tag.testacc-tag.id}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigMultiTag() string {
	return fmt.Sprintf(`
%s

variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VmwareDistributedVirtualPortgroup",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_tag" "testacc-tags-alt" {
  count       = "${length(var.extra_tags)}"
  name        = "${var.extra_tags[count.index]}"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"
  tags                            = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigSingleCustomAttribute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "DistributedVirtualPortgroup" 
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

locals {
  pg_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"

  custom_attributes = "${local.pg_attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedPortGroupConfigMultiCustomAttribute() string {
	return fmt.Sprintf(`
%s

variable "custom_attrs" {
  default = [
    "testacc-attribute-1",
    "testacc-attribute-2"
  ]
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "DistributedVirtualPortgroup"
}

resource "vsphere_custom_attribute" "testacc-attribute-alt" {
  count               = "${length(var.custom_attrs)}"
  name                = "${var.custom_attrs[count.index]}"
  managed_object_type = "DistributedVirtualPortgroup"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

locals {
  pg_attrs = {
    "${vsphere_custom_attribute.testacc-attribute-alt.0.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-alt.1.id}" = "value-2"
  }
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${vsphere_distributed_virtual_switch.dvs.id}"

  custom_attributes = "${local.pg_attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
