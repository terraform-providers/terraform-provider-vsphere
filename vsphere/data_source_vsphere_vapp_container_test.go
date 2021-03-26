package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereVAppContainer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereVAppContainerPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: true,
				Config:             testAccDataSourceVSphereVAppContainerConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_vapp_container.container", "id", regexp.MustCompile("^resgroup-")),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVAppContainer_path(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereVAppContainerPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: true,
				Config:             testAccDataSourceVSphereVAppContainerPathConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_vapp_container.container", "id", regexp.MustCompile("^resgroup-")),
				),
			},
		},
	})
}

func testAccDataSourceVSphereVAppContainerPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_vapp_container acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_vapp_container acceptance tests")
	}
}

func testAccDataSourceVSphereVAppContainerConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datacenter" "dc" {
  name = data.vsphere_datacenter.rootdc1.name
}

resource "vsphere_vapp_container" "vapp" {
  name                    = "vapp-test"
  parent_resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
}

data "vsphere_vapp_container" "container" {
  name          = vsphere_vapp_container.vapp.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,

		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}

func testAccDataSourceVSphereVAppContainerPathConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datacenter" "dc" {
  name = data.vsphere_datacenter.rootdc1.name
}

resource "vsphere_vapp_container" "vapp" {
  name                    = "vapp-test"
  parent_resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
}

data "vsphere_vapp_container" "container" {
  name          = "/${data.vsphere_datacenter.rootdc1.name}/vm/${vsphere_vapp_container.vapp.name}"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}
