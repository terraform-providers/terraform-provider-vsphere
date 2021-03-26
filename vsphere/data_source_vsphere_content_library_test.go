package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereContentLibrary_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereContentLibraryPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereContentLibraryPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_CONTENT_LIBRARY_FILES") == "" {
		t.Skip("set TF_VAR_VSPHERE_CONTENT_LIBRARY_FILES to run vsphere_content_library acceptance tests")
	}
}

func testAccDataSourceVSphereContentLibraryConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_test"
  storage_backing = [ vsphere_nas_datastore.ds1.id ]
  description     = "Library Description"
}

data "vsphere_content_library" "library" {
  name = vsphere_content_library.library.name
}

`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
