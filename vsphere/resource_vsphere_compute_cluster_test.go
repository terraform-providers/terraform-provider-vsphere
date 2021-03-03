package vsphere

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	testAccResourceVSphereComputeClusterNameStandard = "testacc-compute-cluster"
	testAccResourceVSphereComputeClusterNameRenamed  = "testacc-compute-cluster-renamed"
	testAccResourceVSphereComputeClusterFolder       = "compute-cluster-folder-test"
)

func TestAccResourceVSphereComputeCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(false),
				),
			},
			{
				ResourceName:      "vsphere_compute_cluster.compute_cluster",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cluster, err := testGetComputeCluster(s, "compute_cluster")
					if err != nil {
						return "", err
					}
					return cluster.InventoryPath, nil
				},
				ImportStateVerifyIgnore: []string{"force_evacuate_on_destroy"},
				Config:                  testAccResourceVSphereComputeClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_haAdmissionControlPolicyDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigHAAdmissionControlPolicyDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(false),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_drsHAEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigDRSHABasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(true),
					testAccResourceVSphereComputeClusterCheckHAEnabled(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_explicitFailoverHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigDRSHABasicExplicitFailoverHost(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(true),
					testAccResourceVSphereComputeClusterCheckHAEnabled(true),
					testAccResourceVSphereComputeClusterCheckAdmissionControlMode(clusterAdmissionControlTypeFailoverHosts),
					testAccResourceVSphereComputeClusterCheckAdmissionControlFailoverHost(os.Getenv("TF_VAR_VSPHERE_ESXI_HOST1")),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigWithName(testAccResourceVSphereComputeClusterNameStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckName(testAccResourceVSphereComputeClusterNameStandard),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigWithName(testAccResourceVSphereComputeClusterNameRenamed),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckName(testAccResourceVSphereComputeClusterNameRenamed),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_inFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigWithFolder(testAccResourceVSphereComputeClusterFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterMatchInventoryPath(testAccResourceVSphereComputeClusterFolder),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_moveToFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterMatchInventoryPath(""),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigWithFolder(testAccResourceVSphereComputeClusterFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterMatchInventoryPath(testAccResourceVSphereComputeClusterFolder),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_multipleTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_switchTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_singleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_multipleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_switchCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func testAccResourceVSphereComputeClusterPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI1 to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI2") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI2 to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_virtual_machine acceptance tests")
	}
}

func testAccResourceVSphereComputeClusterCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetComputeCluster(s, "compute_cluster")
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected compute cluster to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckDRSEnabled(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		actual := *props.ConfigurationEx.(*types.ClusterConfigInfoEx).DrsConfig.Enabled
		if expected != actual {
			return fmt.Errorf("expected enabled to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckHAEnabled(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		actual := *props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.Enabled
		if expected != actual {
			return fmt.Errorf("expected enabled to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckAdmissionControlMode(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}

		var actual string
		switch props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.AdmissionControlPolicy.(type) {
		case *types.ClusterFailoverResourcesAdmissionControlPolicy:
			actual = clusterAdmissionControlTypeResourcePercentage
		case *types.ClusterFailoverLevelAdmissionControlPolicy:
			actual = clusterAdmissionControlTypeSlotPolicy
		case *types.ClusterFailoverHostAdmissionControlPolicy:
			actual = clusterAdmissionControlTypeFailoverHosts
		default:
			actual = clusterAdmissionControlTypeDisabled
		}
		if expected != actual {
			return fmt.Errorf("expected admission control policy to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckAdmissionControlFailoverHost(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}

		failoverHostsPolicy, ok := props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.AdmissionControlPolicy.(*types.ClusterFailoverHostAdmissionControlPolicy)
		if !ok {
			return fmt.Errorf(
				"admission control policy is not *types.ClusterFailoverHostAdmissionControlPolicy (actual: %T)",
				props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.AdmissionControlPolicy,
			)
		}

		// We just test the first host. The fixture this check is designed to be
		// used with currently only sets one failover host.
		if len(failoverHostsPolicy.FailoverHosts) < 1 {
			return errors.New("no failover hosts")
		}

		client := testAccProvider.Meta().(*VSphereClient).vimClient
		hs, err := hostsystem.FromID(client, failoverHostsPolicy.FailoverHosts[0].Value)
		if err != nil {
			return err
		}

		actual := hs.Name()
		if expected != actual {
			return fmt.Errorf("expected failover host name to be %s, got %s", expected, actual)
		}

		if failoverHostsPolicy.ResourceReductionToToleratePercent != structure.Int32Ptr(0) {
			return fmt.Errorf("expected ha_admission_control_performance_tolerance be 0, got %d", failoverHostsPolicy.ResourceReductionToToleratePercent)
		}

		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testGetComputeCluster(s, "compute_cluster")
		if err != nil {
			return err
		}
		actual := cluster.Name()
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterMatchInventoryPath(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testGetComputeCluster(s, "compute_cluster")
		if err != nil {
			return err
		}

		expected, err = folder.RootPathParticleHost.PathFromNewRoot(cluster.InventoryPath, folder.RootPathParticleHost, expected)
		actual := path.Dir(cluster.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckDRSDefaultAutomationLevel(expected types.DrsBehavior) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		actual := props.ConfigurationEx.(*types.ClusterConfigInfoEx).DrsConfig.DefaultVmBehavior
		if expected != actual {
			return fmt.Errorf("expected default automation level to be %q got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testGetComputeCluster(s, "compute_cluster")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*VSphereClient).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, cluster, tagResName)
	}
}

func testAccResourceVSphereComputeClusterCheckCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_compute_cluster", "compute_cluster", props.Entity())
	}
}

func testAccResourceVSphereComputeClusterConfigEmpty() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigHAAdmissionControlPolicyDisabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids             = [vsphere_host.nested-esxi1.id]
  ha_enabled                  = true
  ha_admission_control_policy = "disabled"

  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigResNestedEsxi(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "hosts" {
  count         = 1
  name          = vsphere_host.nested-esxi1.hostname
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [ vsphere_host.nested-esxi1.id ]

  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigResNestedEsxi(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigDataRootVMNet()),
	)
}

func testAccResourceVSphereComputeClusterConfigDRSHABasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [vsphere_host.nested-esxi1.id]

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = true
  
	force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigResNestedEsxi(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigDRSHABasicExplicitFailoverHost() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [vsphere_host.nested-esxi1.id, data.vsphere_host.roothost2.id]

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled                                    = true
  ha_admission_control_policy                   = "failoverHosts"
  ha_admission_control_failover_host_system_ids = [data.vsphere_host.roothost2.id]
  ha_admission_control_performance_tolerance    = 0

  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigResNestedEsxi(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigWithName(name string) string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		name,
	)
}

func testAccResourceVSphereComputeClusterConfigWithFolder(f string) string {
	return fmt.Sprintf(`
%s

variable "folder" {
  default = "%s"
}

resource "vsphere_folder" "compute_cluster_folder" {
  path          = "${var.folder}"
  type          = "host"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  folder        = "${vsphere_folder.compute_cluster_folder.path}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		f,
	)
}

func testAccResourceVSphereComputeClusterConfigSingleTag() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "ClusterComputeResource",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = [
    "${vsphere_tag.testacc-tag.id}",
  ]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigMultiTag() string {
	return fmt.Sprintf(`
%s

variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "ClusterComputeResource",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_tag" "testacc-tags-alt" {
  count       = "${length(var.extra_tags)}"
  name        = "${var.extra_tags[count.index]}"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigSingleCustomAttribute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "ClusterComputeResource"
}

locals {
  attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigMultiCustomAttributes() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "ClusterComputeResource"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "ClusterComputeResource"
}

locals {
  attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-2.id}" = "value-2"
  }
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
