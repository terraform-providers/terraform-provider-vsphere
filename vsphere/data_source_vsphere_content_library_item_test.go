package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereContentLibraryItem_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereContentLibraryItemPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereContentLibraryItemPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_CONTENT_LIBRARY") == "" {
		t.Skip("set TF_VAR_VSPHERE_CONTENT_LIBRARY to run vsphere_content_library acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CONTENT_LIBRARY_ITEM") == "" {
		t.Skip("set TF_VAR_VSPHERE_CONTENT_LIBRARY_ITEM to run vsphere_content_library_item acceptance tests")
	}
}

func testAccDataSourceVSphereContentLibraryItemConfig() string {
	return fmt.Sprintf(`
variable "datacenter" {
  type    = string
  default = "%s"
}

variable "file_list" {
  type    = list(string)
  default = %s 
}

data "vsphere_datacenter" "dc" {
  name = data.vsphere_datacenter.rootdc1.name
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name = var.datastore
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_test"
  storage_backing = [ data.vsphere_datastore.ds.id ]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "ubuntu"
  description = "Ubuntu Description"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file_list
}

data "vsphere_content_library_item" "item" {
  name       = vsphere_content_library_item.item.name
  library_id = vsphere_content_library.library.id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_CONTENT_LIBRARY_FILES"),
	)
}
