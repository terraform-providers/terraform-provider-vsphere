package vsphere

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereResourcePool_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
				),
			},
			{
				ResourceName:      "vsphere_resource_pool.resource_pool",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rp, err := testGetResourcePool(s, "resource_pool")
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("/%s/host/%s/Resources/terraform-resource-pool-test-parent/%s",
						os.Getenv("TF_VAR_VSPHERE_DATACENTER"),
						os.Getenv("TF_VAR_VSPHERE_CLUSTER"),
						rp.Name(),
					), nil
				},
				Config: testAccResourceVSphereResourcePoolConfigNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckCPUReservation(10),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(false),
					testAccResourceVSphereResourcePoolCheckCPULimit(20),
					testAccResourceVSphereResourcePoolCheckCPUShareLevel("custom"),
					testAccResourceVSphereResourcePoolCheckCPUShares(10),
					testAccResourceVSphereResourcePoolCheckCPUReservation(10),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(false),
					testAccResourceVSphereResourcePoolCheckCPULimit(20),
					testAccResourceVSphereResourcePoolCheckMemoryShareLevel("custom"),
					testAccResourceVSphereResourcePoolCheckMemoryShares(10),
				),
			},
		},
	})
}

func TestAccResourceVSphereResourcePool_updateRename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),

		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckName("terraform-resource-pool-test"),
				),
			},
			{
				Config: testAccResourceVSphereResourcePoolConfigRename(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckName("terraform-resource-pool-test-rename"),
				),
			},
		},
	})
}

func TestAccResourceVSphereResourcePool_updateToCustom(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),

		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckCPUReservation(0),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(true),
					testAccResourceVSphereResourcePoolCheckCPULimit(-1),
					testAccResourceVSphereResourcePoolCheckCPUShareLevel("normal"),
					testAccResourceVSphereResourcePoolCheckCPUReservation(0),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(true),
					testAccResourceVSphereResourcePoolCheckCPULimit(-1),
					testAccResourceVSphereResourcePoolCheckMemoryShareLevel("normal"),
				),
			},
			{
				Config: testAccResourceVSphereResourcePoolConfigNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckCPUReservation(10),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(false),
					testAccResourceVSphereResourcePoolCheckCPULimit(20),
					testAccResourceVSphereResourcePoolCheckCPUShareLevel("custom"),
					testAccResourceVSphereResourcePoolCheckCPUShares(10),
					testAccResourceVSphereResourcePoolCheckCPUReservation(10),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(false),
					testAccResourceVSphereResourcePoolCheckCPULimit(20),
					testAccResourceVSphereResourcePoolCheckMemoryShareLevel("custom"),
					testAccResourceVSphereResourcePoolCheckMemoryShares(10),
				),
			},
		},
	})
}

func TestAccResourceVSphereResourcePool_updateToDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckCPUReservation(10),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(false),
					testAccResourceVSphereResourcePoolCheckCPULimit(20),
					testAccResourceVSphereResourcePoolCheckCPUShareLevel("custom"),
					testAccResourceVSphereResourcePoolCheckCPUShares(10),
					testAccResourceVSphereResourcePoolCheckCPUReservation(10),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(false),
					testAccResourceVSphereResourcePoolCheckCPULimit(20),
					testAccResourceVSphereResourcePoolCheckMemoryShareLevel("custom"),
					testAccResourceVSphereResourcePoolCheckMemoryShares(10),
				),
			},
			{
				Config: testAccResourceVSphereResourcePoolConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckCPUReservation(0),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(true),
					testAccResourceVSphereResourcePoolCheckCPULimit(-1),
					testAccResourceVSphereResourcePoolCheckCPUShareLevel("normal"),
					testAccResourceVSphereResourcePoolCheckCPUReservation(0),
					testAccResourceVSphereResourcePoolCheckCPUExpandable(true),
					testAccResourceVSphereResourcePoolCheckCPULimit(-1),
					testAccResourceVSphereResourcePoolCheckMemoryShareLevel("normal"),
				),
			},
		},
	})
}

func TestAccResourceVSphereResourcePool_esxiHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigEsxiHost(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereResourcePool_updateParent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolHasParent("parent_resource_pool"),
				),
			},
			{
				Config: testAccResourceVSphereResourcePoolConfigAltParent(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolHasParent("alt_parent_resource_pool"),
				),
			},
		},
	})
}

func TestAccResourceVSphereResourcePool_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereResourcePoolPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereResourcePoolCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereResourcePoolConfigTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereResourcePoolCheckExists(true),
					testAccResourceVSphereResourcePoolCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func testAccResourceVSphereResourcePoolPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_resource_pool acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_resource_pool acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI2") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI2 to run vsphere_resource_pool acceptance tests")
	}
}

func testAccResourceVSphereResourcePoolCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetResourcePool(s, "resource_pool")
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected resource pool to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolHasParent(parent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		prp, err := testGetResourcePool(s, parent)
		if err != nil {
			return err
		}
		if *props.Parent != prp.Reference() {
			return fmt.Errorf("resource pool has wrong parent. Expected: %s, got: %s", parent, prp.Name())
		}
		return nil
	}
}

// testAccResourceVSphereResourcePoolCheckTags is a check to ensure that any
// tags that have been created with the supplied resource name have been
// attached to the resource pool.
func testAccResourceVSphereResourcePoolCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rp, err := testGetResourcePool(s, "resource_pool")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, rp, tagResName)
	}
}

func testAccResourceVSphereResourcePoolCheckCPUReservation(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if *props.Config.CpuAllocation.Reservation != *structure.Int64Ptr(int64(value)) {
			return fmt.Errorf("CpuAllocation.Reservation check failed. Expected: %d, got: %d", *props.Config.CpuAllocation.Reservation, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckCPUExpandable(value bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if *props.Config.CpuAllocation.ExpandableReservation != *structure.BoolPtr(value) {
			return fmt.Errorf("CpuAllocation.Expandable check failed. Expected: %t, got: %t", *props.Config.CpuAllocation.ExpandableReservation, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckCPULimit(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if *props.Config.CpuAllocation.Limit != *structure.Int64Ptr(int64(value)) {
			return fmt.Errorf("CpuAllocation.Limit check failed. Expected: %d, got: %d", *props.Config.CpuAllocation.Limit, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckCPUShareLevel(value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if string(props.Config.CpuAllocation.Shares.Level) != value {
			return fmt.Errorf("CpuAllocation.Shares.Level check failed. Expected: %s, got: %s", props.Config.CpuAllocation.Shares.Level, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckCPUShares(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if props.Config.CpuAllocation.Shares.Shares != int32(value) {
			return fmt.Errorf("CpuAllocation.Shares.Shares check failed. Expected: %d, got: %d", props.Config.CpuAllocation.Shares.Shares, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckMemoryShareLevel(value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if string(props.Config.MemoryAllocation.Shares.Level) != value {
			return fmt.Errorf("MemoryAllocation.Shares.Level check failed. Expected: %s, got: %s", props.Config.MemoryAllocation.Shares.Level, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckMemoryShares(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetResourcePoolProperties(s, "resource_pool")
		if err != nil {
			return err
		}
		if props.Config.MemoryAllocation.Shares.Shares != int32(value) {
			return fmt.Errorf("MemoryAllocation.Shares.Shares check failed. Expected: %d, got: %d", props.Config.MemoryAllocation.Shares.Shares, value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolCheckName(value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rp, err := testGetResourcePool(s, "resource_pool")
		if err != nil {
			return err
		}
		if rp.Name() != value {
			return fmt.Errorf("name check failed. Expected: %s, got: %s", rp.Name(), value)
		}
		return nil
	}
}

func testAccResourceVSphereResourcePoolConfigAltParent() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_resource_pool" "alt_parent_resource_pool" {
  name                    = "alt-terraform-resource-pool-test-paren"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.alt_parent_resource_pool.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}

func testAccResourceVSphereResourcePoolConfigNonDefault() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  cpu_share_level         = "custom"
  cpu_shares              = 10
  cpu_reservation         = 10
  cpu_expandable          = false
  cpu_limit               = 20
  memory_share_level      = "custom"
  memory_shares           = 10
  memory_reservation      = 10
  memory_expandable       = false
  memory_limit            = 20
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}

func testAccResourceVSphereResourcePoolConfigEsxiHost() string {
	return fmt.Sprintf(`
%s

variable "host" {
  default = "%s"
}

data "vsphere_host" "host" {
  name          = "${var.host}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${data.vsphere_host.host.resource_pool_id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI2"),
	)
}

func testAccResourceVSphereResourcePoolConfigTags() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "ResourcePool",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
  tags                    = ["${vsphere_tag.testacc-tag.id}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}

func testAccResourceVSphereResourcePoolConfigRename() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "terraform-resource-pool-test-rename"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}

func testAccResourceVSphereResourcePoolConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}
