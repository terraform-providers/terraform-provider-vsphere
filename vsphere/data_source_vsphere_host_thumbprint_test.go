package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereHostThumbprint_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccDataSourceVSphereHostThumbprintPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostThumbprintConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_host_thumbprint.thumb", "id", regexp.MustCompile("([0-9A-F]{2}:)+[0-9A-F]{2}")),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostThumbprintPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI1 to run vsphere_host_thumbprint acceptance tests")
	}
}

func testAccDataSourceVSphereHostThumbprintConfig() string {
	return fmt.Sprintf(`
data "vsphere_host_thumbprint" "thumb" {
  address = "%s"
}`,
		os.Getenv("TF_VAR_VSPHERE_ESXI1"),
	)
}
