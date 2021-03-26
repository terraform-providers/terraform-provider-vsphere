package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceVSphereVmfsDisks_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereVmfsDisksPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVmfsDisksConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputBool("found", "true"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereVmfsDisksPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI1 to run vsphere_vmfs_disks acceptance tests")
	}
}

// testCheckOutputBool checks an output in the Terraform configuration
func testCheckOutputBool(name string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Value.(string) != value {
			return fmt.Errorf(
				"Output '%s': expected %#v, got %#v",
				name,
				value,
				rs)
		}

		return nil
	}
}

func testAccDataSourceVSphereVmfsDisksConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_vmfs_disks" "available" {
  host_system_id = "${data.vsphere_host.esxi_host.id}"
  rescan         = true
}

output "found" {
  value = "${length(data.vsphere_vmfs_disks.available.disks) >= 1 ? "true" : "false" }"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"),
	)
}
