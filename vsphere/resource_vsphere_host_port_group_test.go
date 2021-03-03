package vsphere

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereHostPortGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostPortGroupPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostPortGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostPortGroup_complexWithOverrides(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostPortGroupPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostPortGroupConfigWithOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
					testAccResourceVSphereHostPortGroupCheckVlan(1000),
					testAccResourceVSphereHostPortGroupCheckEffectiveActive([]string{os.Getenv("TF_VAR_VSPHERE_HOST_NIC0")}),
					testAccResourceVSphereHostPortGroupCheckEffectiveStandby([]string{os.Getenv("TF_VAR_VSPHERE_HOST_NIC1")}),
					testAccResourceVSphereHostPortGroupCheckEffectivePromisc(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostPortGroup_basicToComplex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostPortGroupPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostPortGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
				),
			},
			{
				Config: testAccResourceVSphereHostPortGroupConfigWithOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
					testAccResourceVSphereHostPortGroupCheckVlan(1000),
					testAccResourceVSphereHostPortGroupCheckEffectiveActive([]string{os.Getenv("TF_VAR_VSPHERE_HOST_NIC0")}),
					testAccResourceVSphereHostPortGroupCheckEffectiveStandby([]string{os.Getenv("TF_VAR_VSPHERE_HOST_NIC1")}),
					testAccResourceVSphereHostPortGroupCheckEffectivePromisc(true),
				),
			},
		},
	})
}

func testAccResourceVSphereHostPortGroupPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_HOST_NIC0") == "" {
		t.Skip("set TF_VAR_VSPHERE_HOST_NIC0 to run vsphere_host_port_group acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_HOST_NIC1") == "" {
		t.Skip("set TF_VAR_VSPHERE_HOST_NIC1 to run vsphere_host_port_group acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI_HOST to run vsphere_host_port_group acceptance tests")
	}
}

func testAccResourceVSphereHostPortGroupExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := "PGTerraformTest"
		id := "pg"
		_, err := testGetPortGroup(s, id)
		if err != nil {
			if err.Error() == fmt.Sprintf("could not find port group %s", name) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if expected == false {
			return fmt.Errorf("expected port group %s to still be missing", name)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckVlan(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := pg.Spec.VlanId
		if expected != actual {
			return fmt.Errorf("expected VLAN ID to be %d, got %d", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckEffectiveActive(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := pg.ComputedPolicy.NicTeaming.NicOrder.ActiveNic
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected effective active NICs to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckEffectiveStandby(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := pg.ComputedPolicy.NicTeaming.NicOrder.StandbyNic
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected effective standby NICs to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckEffectivePromisc(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := *pg.ComputedPolicy.Security.AllowPromiscuous
		if expected != actual {
			return fmt.Errorf("expected effective allow promiscuous to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupConfig() string {
	return fmt.Sprintf(`
variable "host_nic0" {
  default = "%s"
}

%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_host_virtual_switch" "switch" {
  name           = "vSwitchTerraformTest"
  host_system_id = "${data.vsphere_host.esxi_host.id}"

  network_adapters = ["${var.host_nic0}"]
  active_nics      = ["${var.host_nic0}"]
  standby_nics     = []
}

resource "vsphere_host_port_group" "pg" {
  name                = "PGTerraformTest"
  host_system_id      = "${data.vsphere_host.esxi_host.id}"
  virtual_switch_name = "${vsphere_host_virtual_switch.switch.name}"
}
`, os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccResourceVSphereHostPortGroupConfigWithOverrides() string {
	return fmt.Sprintf(`
variable "host_nic0" {
  default = "%s"
}

variable "host_nic1" {
  default = "%s"
}

%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_host_virtual_switch" "switch" {
  name           = "vSwitchTerraformTest"
  host_system_id = "${data.vsphere_host.esxi_host.id}"

  network_adapters  = ["${var.host_nic0}", "${var.host_nic1}"]
  active_nics       = ["${var.host_nic0}"]
  standby_nics      = ["${var.host_nic1}"]
  allow_promiscuous = false
}

resource "vsphere_host_port_group" "pg" {
  name                = "PGTerraformTest"
  host_system_id      = "${data.vsphere_host.esxi_host.id}"
  virtual_switch_name = "${vsphere_host_virtual_switch.switch.name}"

  vlan_id           = 1000
  active_nics       = ["${var.host_nic0}"]
  standby_nics      = ["${var.host_nic1}"]
  allow_promiscuous = true
}
`, os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}
