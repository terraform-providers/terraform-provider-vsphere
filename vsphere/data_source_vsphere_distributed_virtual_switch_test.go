package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereDistributedVirtualSwitch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.0",
						os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.1",
						os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_distributed_virtual_switch.dvs-data", "id",
						"vsphere_distributed_virtual_switch.dvs", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDistributedVirtualSwitch_absolutePathNoDatacenterSpecified(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDistributedVirtualSwitchConfigAbsolute(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.0",
						os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.1",
						os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_distributed_virtual_switch.dvs-data", "id",
						"vsphere_distributed_virtual_switch.dvs", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDistributedVirtualSwitch_CreatePortgroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDistributedVirtualSwitchConfigWithPortgroup(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"active_uplinks.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"standby_uplinks.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"active_uplinks.0",
						os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
					),
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"standby_uplinks.0",
						os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDistributedVirtualSwitchConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  uplinks       = ["%s", "%s"]
}

data "vsphere_distributed_virtual_switch" "dvs-data" {
  name          = "${vsphere_distributed_virtual_switch.dvs.name}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
	)
}

func testAccDataSourceVSphereDistributedVirtualSwitchConfigWithPortgroup() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  uplinks       = ["%s", "%s"]
}

data "vsphere_distributed_virtual_switch" "dvs-data" {
  name          = "${vsphere_distributed_virtual_switch.dvs.name}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = "${data.vsphere_distributed_virtual_switch.dvs-data.id}"

  active_uplinks  = ["${data.vsphere_distributed_virtual_switch.dvs-data.uplinks[0]}"]
  standby_uplinks = ["${data.vsphere_distributed_virtual_switch.dvs-data.uplinks[1]}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
	)
}

func testAccDataSourceVSphereDistributedVirtualSwitchConfigAbsolute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  uplinks       = ["%s", "%s"]
}

data "vsphere_distributed_virtual_switch" "dvs-data" {
  name          = "/${data.vsphere_datacenter.rootdc1.name}/network/${vsphere_distributed_virtual_switch.dvs.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
	)
}
