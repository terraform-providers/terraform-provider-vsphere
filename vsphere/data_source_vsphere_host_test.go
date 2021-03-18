package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereHost_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereHostPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host.host",
						"id",
						testAccDataSourceVSphereHostExpectedRegexp(),
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereHost_defaultHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereHostPreCheck(t)
			testAccSkipIfNotEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfigDefault(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host.host",
						"id",
						testAccDataSourceVSphereHostExpectedRegexp(),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_host acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI1 to run vsphere_host acceptance tests")
	}
}

func testAccDataSourceVSphereHostExpectedRegexp() *regexp.Regexp {
	if os.Getenv("TF_VAR_VSPHERE_TEST_ESXI") != "" {
		return regexp.MustCompile("^ha-host$")
	}
	return regexp.MustCompile("^host-")
}

func testAccDataSourceVSphereHostConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()), os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccDataSourceVSphereHostConfigDefault() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "host" {
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}`, testhelper.ConfigDataRootDC1())
}
