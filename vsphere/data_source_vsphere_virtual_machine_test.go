package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereVirtualMachine_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereVirtualMachinePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVirtualMachineConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_virtual_machine.template",
						"id",
						regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "scsi_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "num_cores_per_socket"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "firmware"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "hardware_version"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.eagerly_scrub"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.thin_provisioned"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.unit_number"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "network_interface_types.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "firmware"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVirtualMachine_noDatacenterAndAbsolutePath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereVirtualMachinePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVirtualMachineConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_virtual_machine.template",
						"id",
						regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "scsi_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "num_cores_per_socket"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "firmware"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "hardware_version"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.eagerly_scrub"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.thin_provisioned"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.unit_number"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "network_interface_types.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.template", "firmware"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereVirtualMachinePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_virtual_machine data source acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE to run vsphere_virtual_machine data source acceptance tests")
	}
}

func testAccDataSourceVSphereVirtualMachineConfig() string {
	return fmt.Sprintf(`
%s

variable "template" {
  default = "%s"
}

data "vsphere_virtual_machine" "template" {
  name          = "${var.template}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
	)
}

func testAccDataSourceVSphereVirtualMachineConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

variable "template" {
  default = "%s"
}

data "vsphere_virtual_machine" "template" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/vm/${var.template}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
	)
}
