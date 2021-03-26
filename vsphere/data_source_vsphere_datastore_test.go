package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereDatastore_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereDatastorePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore.datastore_data", "id",
						"vsphere_nas_datastore.ds1", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatastore_noDatacenterAndAbsolutePath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereDatastorePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore.datastore_data", "id",
						"vsphere_nas_datastore.ds1", "id",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatastorePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_nas_datastore acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_nas_datastore acceptance tests")
	}
}

func testAccDataSourceVSphereDatastoreConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datastore" "datastore_data" {
  name          = vsphere_nas_datastore.ds1.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccDataSourceVSphereDatastoreConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

data "vsphere_datastore" "datastore_data" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/datastore/${vsphere_nas_datastore.ds1.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
