package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testAccDataSourceVSphereFolderExpectedRegexp = regexp.MustCompile("^group-v")

func TestAccDataSourceVSphereFolder_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereFolderPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereFolderConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_folder.folder",
						"id",
						testAccDataSourceVSphereFolderExpectedRegexp,
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereFolderPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_folder acceptance tests")
	}
}

func testAccDataSourceVSphereFolderConfig() string {
	return fmt.Sprintf(`
data "vsphere_datacenter" "dc" {
  name = "%s"
}

resource "vsphere_folder" "folder" {
  path          = "test-folder"
  type          = "vm"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_folder" "folder" {
  path = "/${data.vsphere_datacenter.dc.name}/vm/${vsphere_folder.folder.path}"
}
`, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
}
