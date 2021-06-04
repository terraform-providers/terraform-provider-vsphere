package vsphere

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceVSphereHostVirtualSwitch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
			{
				ResourceName:      "vsphere_host_virtual_switch.switch",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_host_virtual_switch.%s", "switch"))
					if err != nil {
						return "", err
					}
					return vars.resourceID, err
				},
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_removeNIC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfigSingleNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_noNICs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfigNoNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
					testAccResourceVSphereHostVirtualSwitchNoBridge(),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_badActiveNICList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereHostVirtualSwitchConfigBadActive(),
				ExpectError: regexp.MustCompile(fmt.Sprintf("active NIC entry %q not present in network_adapters list", os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"))),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_badStandbyNICList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereHostVirtualSwitchConfigBadStandby(),
				ExpectError: regexp.MustCompile(fmt.Sprintf("standby NIC entry %q not present in network_adapters list", os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"))),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_removeAllNICs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHostVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfigNoNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
					testAccResourceVSphereHostVirtualSwitchNoBridge(),
				),
			},
		},
	})
}

func testAccResourceVSphereHostVirtualSwitchPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_HOST_NIC0") == "" {
		t.Skip("set TF_VAR_VSPHERE_HOST_NIC0 to run vsphere_host_virtual_switch acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_HOST_NIC1") == "" {
		t.Skip("set TF_VAR_VSPHERE_HOST_NIC1 to run vsphere_host_virtual_switch acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI_HOST to run vsphere_host_virtual_switch acceptance tests")
	}
}

func testAccResourceVSphereHostVirtualSwitchExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_host_virtual_switch.switch")
		if err != nil {
			if expected {
				return errors.New("vsphere_host_virtual_switch.switch not found in state")
			}
			return nil
		}

		hsID, name, err := splitHostVirtualSwitchID(vars.resourceID)
		if err != nil {
			return err
		}
		ns, err := hostNetworkSystemFromHostSystemID(vars.client, hsID)
		if err != nil {
			return fmt.Errorf("error loading host network system: %s", err)
		}

		_, err = hostVSwitchFromName(vars.client, ns, name)
		if err != nil {
			if err.Error() == fmt.Sprintf("could not find virtual switch %s", name) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected vSwitch %s to be missing", name)
		}
		return nil
	}
}

func testAccResourceVSphereHostVirtualSwitchNoBridge() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_host_virtual_switch.switch")
		if err != nil {
			return errors.New("vsphere_host_virtual_switch.switch not found in state")
		}

		hsID, name, err := splitHostVirtualSwitchID(vars.resourceID)
		if err != nil {
			return err
		}
		ns, err := hostNetworkSystemFromHostSystemID(vars.client, hsID)
		if err != nil {
			return fmt.Errorf("error loading host network system: %s", err)
		}

		sw, err := hostVSwitchFromName(vars.client, ns, name)
		if err != nil {
			return err
		}
		if sw.Spec.Bridge != nil {
			return fmt.Errorf("expected no bridge on switch, got %+v", sw.Spec.Bridge)
		}
		return nil
	}
}

func testAccResourceVSphereHostVirtualSwitchConfig() string {
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

  network_adapters = [var.host_nic0, var.host_nic1]

  active_nics  = [var.host_nic0]
  standby_nics = [var.host_nic1]
}
`, os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccResourceVSphereHostVirtualSwitchConfigSingleNIC() string {
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

  active_nics  = ["${var.host_nic0}"]
  standby_nics = []
}
`, os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccResourceVSphereHostVirtualSwitchConfigNoNIC() string {
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

  network_adapters = []

  active_nics  = []
  standby_nics = []
}
`, os.Getenv("TF_VAR_VSPHERE_ESXI_TRUNK_NIC"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccResourceVSphereHostVirtualSwitchConfigBadActive() string {
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

  network_adapters = []

  active_nics  = ["${var.host_nic0}"]
  standby_nics = []
}
`, os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccResourceVSphereHostVirtualSwitchConfigBadStandby() string {
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

  network_adapters = []

  active_nics  = []
  standby_nics = ["${var.host_nic0}"]
}
`, os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}
