package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Storage policy resource is needed.
func skipTestAccDataSourceVSphereStoragePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereStoragePolicyPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereStoragePolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_storage_policy.storage_policy", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
				),
			},
		},
	})
}

func testAccDataSourceVSphereStoragePolicyPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_STORAGE_POLICY") == "" {
		t.Skip("set TF_VAR_VSPHERE_STORAGE_POLICY to run vsphere_storage_policy acceptance tests")
	}
}

func testAccDataSourceVSphereStoragePolicyConfig() string {
	return fmt.Sprintf(`
variable "storage_policy" {
  default = "%s"
}

data "vsphere_storage_policy" "storage_policy" {
  name          = "${var.storage_policy}"
}
`,
		os.Getenv("TF_VAR_VSPHERE_STORAGE_POLICY"),
	)
}
