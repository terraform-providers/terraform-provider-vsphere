package vsphere

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	testAccResourceVSphereDatastoreClusterNameStandard = "testacc-datastore-cluster"
	testAccResourceVSphereDatastoreClusterNameRenamed  = "testacc-datastore-cluster-renamed"
	testAccResourceVSphereDatastoreClusterFolder       = "datastore-cluster-folder-test"
)

func TestAccResourceVSphereDatastoreCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(false),
				),
			},
			{
				ResourceName:            "vsphere_datastore_cluster.datastore_cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"datacenter_id", "sdrs_free_space_threshold"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					pod, err := testGetDatastoreCluster(s, "datastore_cluster")
					if err != nil {
						return "", err
					}
					return pod.InventoryPath, nil
				},
				Config: testAccResourceVSphereDatastoreClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(false),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_sdrsEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSDRSBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigWithName(testAccResourceVSphereDatastoreClusterNameStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckName(testAccResourceVSphereDatastoreClusterNameStandard),
				),
			},
			{
				Config: testAccResourceVSphereDatastoreClusterConfigWithName(testAccResourceVSphereDatastoreClusterNameRenamed),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckName(testAccResourceVSphereDatastoreClusterNameRenamed),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_inFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigWithFolder(testAccResourceVSphereDatastoreClusterFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterMatchInventoryPath(testAccResourceVSphereDatastoreClusterFolder),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_moveToFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterMatchInventoryPath(""),
				),
			},
			{
				Config: testAccResourceVSphereDatastoreClusterConfigWithFolder(testAccResourceVSphereDatastoreClusterFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterMatchInventoryPath(testAccResourceVSphereDatastoreClusterFolder),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_sdrsOverrides(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSDRSOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSDefaultAutomationLevel(string(types.StorageDrsPodConfigInfoBehaviorManual)),
					testAccResourceVSphereDatastoreClusterCheckSDRSOverrides(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_miscTweaks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSDRSMiscTweaks(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSDefaultIntraVMAffinity(false),
					testAccResourceVSphereDatastoreClusterCheckSDRSIoLatencyThreshold(5),
					testAccResourceVSphereDatastoreClusterCheckSDRSSpaceThresholdMode(
						string(types.StorageDrsSpaceLoadBalanceConfigSpaceThresholdModeUtilization),
					),
					testAccResourceVSphereDatastoreClusterCheckSDRSSpaceUtilizationThreshold(50),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_reservableIops(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigReservableIopsManual(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSReservableIopsThresholdMode(
						string(types.StorageDrsPodConfigInfoBehaviorManual),
					),
					testAccResourceVSphereDatastoreClusterCheckSDRSReservableIopsThreshold(5000),
				),
			},
			{
				Config: testAccResourceVSphereDatastoreClusterConfigReservableIopsAutomatic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSReservableIopsThresholdMode(
						string(types.StorageDrsPodConfigInfoBehaviorAutomated),
					),
					testAccResourceVSphereDatastoreClusterCheckSDRSReservablePercentThreshold(40),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_freeSpace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSpaceManual(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSSpaceThresholdMode(
						string(types.StorageDrsSpaceLoadBalanceConfigSpaceThresholdModeFreeSpace),
					),
					testAccResourceVSphereDatastoreClusterCheckSDRSFreeSpaceThresholdGB(500),
				),
			},
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSDRSMiscTweaks(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(true),
					testAccResourceVSphereDatastoreClusterCheckSDRSSpaceThresholdMode(
						string(types.StorageDrsSpaceLoadBalanceConfigSpaceThresholdModeUtilization),
					),
					testAccResourceVSphereDatastoreClusterCheckSDRSSpaceUtilizationThreshold(50),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_multipleTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_switchTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereDatastoreClusterConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_singleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_multipleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatastoreCluster_switchCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDatastoreClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatastoreClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatastoreClusterConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereDatastoreClusterConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatastoreClusterCheckExists(true),
					testAccResourceVSphereDatastoreClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func testAccResourceVSphereDatastoreClusterPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_datastore_cluster acceptance tests")
	}
}

func testAccResourceVSphereDatastoreClusterCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetDatastoreCluster(s, "datastore_cluster")
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected datastore cluster to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSEnabled(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}
		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.Enabled
		if expected != actual {
			return fmt.Errorf("expected enabled to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pod, err := testGetDatastoreCluster(s, "datastore_cluster")
		if err != nil {
			return err
		}
		actual := pod.Name()
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDatastoreClusterMatchInventoryPath(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pod, err := testGetDatastoreCluster(s, "datastore_cluster")
		if err != nil {
			return err
		}

		expected, err = folder.RootPathParticleDatastore.PathFromNewRoot(pod.InventoryPath, folder.RootPathParticleDatastore, expected)
		actual := path.Dir(pod.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSDefaultAutomationLevel(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}
		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.DefaultVmBehavior
		if expected != actual {
			return fmt.Errorf("expected default automation level to be %q got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSOverrides() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}
		expected := &types.StorageDrsAutomationConfig{
			IoLoadBalanceAutomationMode:     string(types.StorageDrsPodConfigInfoBehaviorAutomated),
			PolicyEnforcementAutomationMode: string(types.StorageDrsPodConfigInfoBehaviorAutomated),
			RuleEnforcementAutomationMode:   string(types.StorageDrsPodConfigInfoBehaviorAutomated),
			SpaceLoadBalanceAutomationMode:  string(types.StorageDrsPodConfigInfoBehaviorAutomated),
			VmEvacuationAutomationMode:      string(types.StorageDrsPodConfigInfoBehaviorAutomated),
		}
		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.AutomationOverrides
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSDefaultIntraVMAffinity(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := *props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.DefaultIntraVmAffinity

		if expected != actual {
			return fmt.Errorf("expected DefaultIntraVmAffinity to be %t, got %t", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSIoLatencyThreshold(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.IoLoadBalanceConfig.IoLatencyThreshold

		if expected != actual {
			return fmt.Errorf("expected IoLatencyThreshold to be %d, got %d", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSSpaceUtilizationThreshold(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.SpaceLoadBalanceConfig.SpaceUtilizationThreshold

		if expected != actual {
			return fmt.Errorf("expected SpaceUtilizationThreshold to be %d, got %d", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSSpaceThresholdMode(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.SpaceLoadBalanceConfig.SpaceThresholdMode

		if expected != actual {
			return fmt.Errorf("expected SpaceThresholdMode to be %q, got %q", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSReservableIopsThresholdMode(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.IoLoadBalanceConfig.ReservableThresholdMode

		if expected != actual {
			return fmt.Errorf("expected SpaceThresholdMode to be %q, got %q", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSReservableIopsThreshold(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.IoLoadBalanceConfig.ReservableIopsThreshold

		if expected != actual {
			return fmt.Errorf("expected SpaceThresholdMode to be %d, got %d", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSReservablePercentThreshold(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.IoLoadBalanceConfig.ReservablePercentThreshold

		if expected != actual {
			return fmt.Errorf("expected SpaceThresholdMode to be %d, got %d", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckSDRSFreeSpaceThresholdGB(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}

		actual := props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.SpaceLoadBalanceConfig.FreeSpaceThresholdGB

		if expected != actual {
			return fmt.Errorf("expected SpaceUtilizationThreshold to be %d, got %d", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDatastoreClusterCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pod, err := testGetDatastoreCluster(s, "datastore_cluster")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, pod, tagResName)
	}
}

func testAccResourceVSphereDatastoreClusterCheckCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreClusterProperties(s, "datastore_cluster")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_datastore_cluster", "datastore_cluster", props.Entity())
	}
}

func testAccResourceVSphereDatastoreClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigSDRSBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled  = true
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigWithName(name string) string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		name,
	)
}

func testAccResourceVSphereDatastoreClusterConfigWithFolder(f string) string {
	return fmt.Sprintf(`
%s

variable "folder" {
  default = "%s"
}

resource "vsphere_folder" "datastore_cluster_folder" {
  path          = "${var.folder}"
  type          = "datastore"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  folder        = "${vsphere_folder.datastore_cluster_folder.path}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		f,
	)
}

func testAccResourceVSphereDatastoreClusterConfigSDRSOverrides() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name                                     = "testacc-datastore-cluster"
  datacenter_id                            = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled                             = true
  sdrs_automation_level                    = "manual"
  sdrs_space_balance_automation_level      = "automated"
  sdrs_io_balance_automation_level         = "automated"
  sdrs_rule_enforcement_automation_level   = "automated"
  sdrs_policy_enforcement_automation_level = "automated"
  sdrs_vm_evacuation_automation_level      = "automated"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigSDRSMiscTweaks() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name                             = "testacc-datastore-cluster"
  datacenter_id                    = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled                     = true
  sdrs_default_intra_vm_affinity   = false
  sdrs_io_latency_threshold        = 5
  sdrs_space_utilization_threshold = 50
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigReservableIopsManual() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name                              = "testacc-datastore-cluster"
  datacenter_id                     = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled                      = true
  sdrs_io_reservable_threshold_mode = "manual"
  sdrs_io_reservable_iops_threshold = 5000
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigReservableIopsAutomatic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name                                 = "testacc-datastore-cluster"
  datacenter_id                        = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled                         = true
  sdrs_io_reservable_percent_threshold = 40
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigSpaceManual() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name                           = "testacc-datastore-cluster"
  datacenter_id                  = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled                   = true
  sdrs_free_space_threshold_mode = "freeSpace"
  sdrs_free_space_threshold      = 500
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigSingleTag() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "StoragePod",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = [
    "${vsphere_tag.testacc-tag.id}",
  ]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigMultiTag() string {
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
    "StoragePod",
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

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigSingleCustomAttribute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "StoragePod"
}

locals {
  attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDatastoreClusterConfigMultiCustomAttributes() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "StoragePod"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "StoragePod"
}

locals {
  attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-2.id}" = "value-2"
  }
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
