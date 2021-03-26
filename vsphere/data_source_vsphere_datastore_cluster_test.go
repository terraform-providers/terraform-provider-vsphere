package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereDatastoreCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore_cluster.datastore_cluster_data", "id",
						"vsphere_datastore_cluster.datastore_cluster", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatastoreCluster_absolutePathNoDatacenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreClusterConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore_cluster.datastore_cluster_data", "id",
						"vsphere_datastore_cluster.datastore_cluster", "id",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatastoreClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_datastore_cluster" "datastore_cluster_data" {
  name          = "${vsphere_datastore_cluster.datastore_cluster.name}"
  datacenter_id = "${vsphere_datastore_cluster.datastore_cluster.datacenter_id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccDataSourceVSphereDatastoreClusterConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_datastore_cluster" "datastore_cluster_data" {
  name          = "/${data.vsphere_datacenter.rootdc1.name}/datastore/${vsphere_datastore_cluster.datastore_cluster.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
