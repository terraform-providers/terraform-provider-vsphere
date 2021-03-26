package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereComputeCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereComputeClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "id",
						"vsphere_compute_cluster.compute_cluster", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "resource_pool_id",
						"vsphere_compute_cluster.compute_cluster", "resource_pool_id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereComputeCluster_absolutePathNoDatacenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereComputeClusterConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "id",
						"vsphere_compute_cluster.compute_cluster", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "resource_pool_id",
						"vsphere_compute_cluster.compute_cluster", "resource_pool_id",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereComputeClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_compute_cluster" "compute_cluster_data" {
  name          = "${vsphere_compute_cluster.compute_cluster.name}"
  datacenter_id = "${vsphere_compute_cluster.compute_cluster.datacenter_id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccDataSourceVSphereComputeClusterConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_compute_cluster" "compute_cluster_data" {
  name          = "/${data.vsphere_datacenter.rootdc1.name}/host/${vsphere_compute_cluster.compute_cluster.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
