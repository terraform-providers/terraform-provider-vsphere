package vsphere

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/virtualmachine"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereHAVMOverride_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHAVMOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHAVMOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHAVMOverrideConfigOverrideDefaults(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHAVMOverrideExists(true),
					testAccResourceVSphereHAVMOverrideMatchBase(
						string(types.ClusterDasVmSettingsIsolationResponseClusterIsolationResponse),
						string(types.ClusterDasVmSettingsRestartPriorityClusterRestartPriority),
						-1,
					),
					testAccResourceVSphereHAVMOverrideMatchVMCP(
						string(types.ClusterVmComponentProtectionSettingsVmReactionOnAPDClearedUseClusterDefault),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionClusterDefault),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionClusterDefault),
						-1,
					),
					testAccResourceVSphereHAVMOverrideMatchMonitoring(
						true,
						30,
						3,
						-1,
						120,
						string(types.ClusterDasConfigInfoVmMonitoringStateVmMonitoringDisabled),
					),
				),
			},
			{
				ResourceName:      "vsphere_ha_vm_override.ha_vm_override",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cluster, err := testGetComputeClusterFromDataSource(s, "rootcompute_cluster1")
					if err != nil {
						return "", err
					}
					vm, err := testGetVirtualMachine(s, "vm")
					if err != nil {
						return "", err
					}

					m := make(map[string]string)
					m["compute_cluster_path"] = cluster.InventoryPath
					m["virtual_machine_path"] = vm.InventoryPath
					b, err := json.Marshal(m)
					if err != nil {
						return "", err
					}

					return string(b), nil
				},
				Config: testAccResourceVSphereHAVMOverrideConfigOverrideDefaults(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHAVMOverrideExists(true),
					testAccResourceVSphereHAVMOverrideMatchBase(
						string(types.ClusterDasVmSettingsIsolationResponseClusterIsolationResponse),
						string(types.ClusterDasVmSettingsRestartPriorityClusterRestartPriority),
						-1,
					),
					testAccResourceVSphereHAVMOverrideMatchVMCP(
						string(types.ClusterVmComponentProtectionSettingsVmReactionOnAPDClearedUseClusterDefault),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionClusterDefault),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionClusterDefault),
						-1,
					),
					testAccResourceVSphereHAVMOverrideMatchMonitoring(
						true,
						30,
						3,
						-1,
						120,
						string(types.ClusterDasConfigInfoVmMonitoringStateVmMonitoringDisabled),
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereHAVMOverride_complete(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHAVMOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHAVMOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHAVMOverrideConfigOverrideComplete(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHAVMOverrideExists(true),
					testAccResourceVSphereHAVMOverrideMatchBase(
						string(types.ClusterDasVmSettingsIsolationResponseShutdown),
						string(types.ClusterDasVmSettingsRestartPriorityHighest),
						30,
					),
					testAccResourceVSphereHAVMOverrideMatchVMCP(
						string(types.ClusterVmComponentProtectionSettingsVmReactionOnAPDClearedReset),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionRestartConservative),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionRestartAggressive),
						60,
					),
					testAccResourceVSphereHAVMOverrideMatchMonitoring(
						false,
						60,
						5,
						600,
						300,
						string(types.ClusterDasConfigInfoVmMonitoringStateVmMonitoringOnly),
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereHAVMOverride_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereHAVMOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHAVMOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHAVMOverrideConfigOverrideDefaults(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHAVMOverrideExists(true),
					testAccResourceVSphereHAVMOverrideMatchBase(
						string(types.ClusterDasVmSettingsIsolationResponseClusterIsolationResponse),
						string(types.ClusterDasVmSettingsRestartPriorityClusterRestartPriority),
						-1,
					),
					testAccResourceVSphereHAVMOverrideMatchVMCP(
						string(types.ClusterVmComponentProtectionSettingsVmReactionOnAPDClearedUseClusterDefault),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionClusterDefault),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionClusterDefault),
						-1,
					),
					testAccResourceVSphereHAVMOverrideMatchMonitoring(
						true,
						30,
						3,
						-1,
						120,
						string(types.ClusterDasConfigInfoVmMonitoringStateVmMonitoringDisabled),
					),
				),
			},
			{
				Config: testAccResourceVSphereHAVMOverrideConfigOverrideComplete(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHAVMOverrideExists(true),
					testAccResourceVSphereHAVMOverrideMatchBase(
						string(types.ClusterDasVmSettingsIsolationResponseShutdown),
						string(types.ClusterDasVmSettingsRestartPriorityHighest),
						30,
					),
					testAccResourceVSphereHAVMOverrideMatchVMCP(
						string(types.ClusterVmComponentProtectionSettingsVmReactionOnAPDClearedReset),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionRestartConservative),
						string(types.ClusterVmComponentProtectionSettingsStorageVmReactionRestartAggressive),
						60,
					),
					testAccResourceVSphereHAVMOverrideMatchMonitoring(
						false,
						60,
						5,
						600,
						300,
						string(types.ClusterDasConfigInfoVmMonitoringStateVmMonitoringOnly),
					),
				),
			},
		},
	})
}

func testAccResourceVSphereHAVMOverridePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_ha_vm_override acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_ha_vm_override acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_ha_vm_override acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_ha_vm_override acceptance tests")
	}
}

func testAccResourceVSphereHAVMOverrideExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		info, err := testGetComputeClusterHaVMConfig(s, "ha_vm_override")
		if err != nil {
			if expected == false {
				switch {
				case viapi.IsManagedObjectNotFoundError(err):
					fallthrough
				case virtualmachine.IsUUIDNotFoundError(err):
					// This is not necessarily a missing override, but more than likely a
					// missing cluster, which happens during destroy as the dependent
					// resources will be missing as well, so want to treat this as a
					// deleted override as well.
					return nil
				}
			}
			return err
		}

		switch {
		case info == nil && !expected:
			// Expected missing
			return nil
		case info == nil && expected:
			// Expected to exist
			return errors.New("HA VM override missing when expected to exist")
		case !expected:
			return errors.New("HA VM override still present when expected to be missing")
		}

		return nil
	}
}

func testAccResourceVSphereHAVMOverrideMatchBase(
	isolationResponse string,
	restartPriority string,
	restartPriorityTimeout int32,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterHaVMConfig(s, "ha_vm_override")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("HA VM override missing")
		}

		expected := &types.ClusterDasVmConfigInfo{
			DasSettings: &types.ClusterDasVmSettings{
				IsolationResponse:      isolationResponse,
				RestartPriority:        restartPriority,
				RestartPriorityTimeout: restartPriorityTimeout,
			},
			Key: actual.Key,
		}

		actual.DasSettings.VmComponentProtectionSettings = nil
		actual.DasSettings.VmToolsMonitoringSettings = nil
		actual.PowerOffOnIsolation = nil
		actual.RestartPriority = ""

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereHAVMOverrideMatchVMCP(
	vmReactionOnAPDCleared string,
	vmStorageProtectionForAPD string,
	vmStorageProtectionForPDL string,
	vmTerminateDelayForAPDSec int32,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterHaVMConfig(s, "ha_vm_override")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("HA VM override missing")
		}

		expected := &types.ClusterDasVmConfigInfo{
			DasSettings: &types.ClusterDasVmSettings{
				VmComponentProtectionSettings: &types.ClusterVmComponentProtectionSettings{
					VmReactionOnAPDCleared:    vmReactionOnAPDCleared,
					VmStorageProtectionForAPD: vmStorageProtectionForAPD,
					VmStorageProtectionForPDL: vmStorageProtectionForPDL,
					VmTerminateDelayForAPDSec: vmTerminateDelayForAPDSec,
				},
			},
			Key: actual.Key,
		}

		actual.DasSettings.IsolationResponse = ""
		actual.DasSettings.RestartPriority = ""
		actual.DasSettings.RestartPriorityTimeout = 0
		actual.DasSettings.VmToolsMonitoringSettings = nil
		actual.PowerOffOnIsolation = nil
		actual.RestartPriority = ""

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereHAVMOverrideMatchMonitoring(
	clusterSettings bool,
	failureInterval int32,
	maxFailures int32,
	maxFailureWindow int32,
	minUpTime int32,
	vmMonitoring string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterHaVMConfig(s, "ha_vm_override")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("DRS VM override missing")
		}

		expected := &types.ClusterDasVmConfigInfo{
			DasSettings: &types.ClusterDasVmSettings{
				VmToolsMonitoringSettings: &types.ClusterVmToolsMonitoringSettings{
					ClusterSettings:  structure.BoolPtr(clusterSettings),
					FailureInterval:  failureInterval,
					MaxFailures:      maxFailures,
					MaxFailureWindow: maxFailureWindow,
					MinUpTime:        minUpTime,
					VmMonitoring:     vmMonitoring,
				},
			},
			Key: actual.Key,
		}

		actual.DasSettings.IsolationResponse = ""
		actual.DasSettings.RestartPriority = ""
		actual.DasSettings.RestartPriorityTimeout = 0
		actual.DasSettings.VmComponentProtectionSettings = nil
		actual.PowerOffOnIsolation = nil
		actual.RestartPriority = ""

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereHAVMOverrideConfigOverrideDefaults() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

	wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}

resource "vsphere_ha_vm_override" "ha_vm_override" {
  compute_cluster_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.id}"
  virtual_machine_id = "${vsphere_virtual_machine.vm.id}"
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigResDS1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigResResourcePool1(),
			testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereHAVMOverrideConfigOverrideComplete() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

	wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}

resource "vsphere_ha_vm_override" "ha_vm_override" {
  compute_cluster_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.id}"
  virtual_machine_id = "${vsphere_virtual_machine.vm.id}"

  ha_vm_restart_priority = "highest"
  ha_vm_restart_timeout  = 30

  ha_host_isolation_response = "shutdown"

  ha_datastore_pdl_response        = "restartAggressive"
  ha_datastore_apd_response        = "restartConservative"
  ha_datastore_apd_recovery_action = "reset"
  ha_datastore_apd_response_delay  = 60

  ha_vm_monitoring_use_cluster_defaults = false
  ha_vm_monitoring                      = "vmMonitoringOnly"
  ha_vm_failure_interval                = 60
  ha_vm_minimum_uptime                  = 300
  ha_vm_maximum_resets                  = 5
  ha_vm_maximum_failure_window          = 600
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
