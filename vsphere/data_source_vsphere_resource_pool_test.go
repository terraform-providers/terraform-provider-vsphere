package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereResourcePool_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereResourcePoolPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereResourcePoolConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_resource_pool.pool", "id", regexp.MustCompile("^resgroup-")),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereResourcePool_noDatacenterAndAbsolutePath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereResourcePoolPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereResourcePoolConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_resource_pool.pool", "id", regexp.MustCompile("^resgroup-")),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereResourcePool_defaultResourcePoolForESXi(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereResourcePoolPreCheck(t)
			testAccSkipIfNotEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereResourcePoolConfigDefault,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_resource_pool.pool", "id", regexp.MustCompile("^ha-root-pool$")),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereResourcePool_emptyNameOnVCenterShouldError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereResourcePoolPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceVSphereResourcePoolConfigDefault,
				ExpectError: regexp.MustCompile("name cannot be empty when using vCenter"),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccDataSourceVSphereResourcePoolPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_resource_pool data source acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_resource_pool data source acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL") == "" {
		t.Skip("set TF_VAR_VSPHERE_RESOURCE_POOL to run vsphere_resource_pool data source acceptance tests")
	}
}

func testAccDataSourceVSphereResourcePoolConfig() string {
	return fmt.Sprintf(`
%s

variable "resource_pool" {
  default = "%s"
}

data "vsphere_resource_pool" "pool" {
  name          = "${data.vsphere_compute_cluster.rootcompute_cluster1.name}/Resources/${var.resource_pool}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootComputeCluster1()),

		os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL"),
	)
}

func testAccDataSourceVSphereResourcePoolConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

variable "resource_pool" {
  default = "%s"
}

data "vsphere_resource_pool" "pool" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/host/${data.vsphere_compute_cluster.rootcompute_cluster1.name}/Resources/${var.resource_pool}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootComputeCluster1()),

		os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL"),
	)
}

const testAccDataSourceVSphereResourcePoolConfigDefault = `
data "vsphere_resource_pool" "pool" {}
`
