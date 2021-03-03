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

func TestAccResourceVSphereDRSVMOverride_drs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDRSVMOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDRSVMOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDRSVMOverrideConfigOverrideDRSEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDRSVMOverrideExists(true),
					testAccResourceVSphereDRSVMOverrideMatch(types.DrsBehaviorManual, false),
				),
			},
			{
				ResourceName:      "vsphere_drs_vm_override.drs_vm_override",
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
				Config: testAccResourceVSphereDRSVMOverrideConfigOverrideDRSEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDRSVMOverrideExists(true),
					testAccResourceVSphereDRSVMOverrideMatch(types.DrsBehaviorManual, false),
				),
			},
		},
	})
}

func TestAccResourceVSphereDRSVMOverride_automationLevel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDRSVMOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDRSVMOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDRSVMOverrideConfigOverrideAutomationLevel(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDRSVMOverrideExists(true),
					testAccResourceVSphereDRSVMOverrideMatch(types.DrsBehaviorFullyAutomated, true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDRSVMOverride_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDRSVMOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDRSVMOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDRSVMOverrideConfigOverrideDRSEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDRSVMOverrideExists(true),
					testAccResourceVSphereDRSVMOverrideMatch(types.DrsBehaviorManual, false),
				),
			},
			{
				Config: testAccResourceVSphereDRSVMOverrideConfigOverrideAutomationLevel(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDRSVMOverrideExists(true),
					testAccResourceVSphereDRSVMOverrideMatch(types.DrsBehaviorFullyAutomated, true),
				),
			},
		},
	})
}

func testAccResourceVSphereDRSVMOverridePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_drs_vm_override acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_drs_vm_override acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_drs_vm_override acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_drs_vm_override acceptance tests")
	}
}

func testAccResourceVSphereDRSVMOverrideExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		info, err := testGetComputeClusterDRSVMConfig(s, "drs_vm_override")
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
			return errors.New("DRS VM override missing when expected to exist")
		case !expected:
			return errors.New("DRS VM override still present when expected to be missing")
		}

		return nil
	}
}

func testAccResourceVSphereDRSVMOverrideMatch(behavior types.DrsBehavior, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterDRSVMConfig(s, "drs_vm_override")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("DRS VM override missing")
		}

		expected := &types.ClusterDrsVmConfigInfo{
			Behavior: behavior,
			Enabled:  structure.BoolPtr(enabled),
			Key:      actual.Key,
		}

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDRSVMOverrideConfigOverrideDRSEnabled() string {
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

resource "vsphere_drs_vm_override" "drs_vm_override" {
  compute_cluster_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.id}"
  virtual_machine_id = "${vsphere_virtual_machine.vm.id}"
  drs_enabled        = false
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigResDS1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigResResourcePool1(),
			testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDRSVMOverrideConfigOverrideAutomationLevel() string {
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

resource "vsphere_drs_vm_override" "drs_vm_override" {
  compute_cluster_id   = "${data.vsphere_compute_cluster.rootcompute_cluster1.id}"
  virtual_machine_id   = "${vsphere_virtual_machine.vm.id}"
  drs_enabled          = true
  drs_automation_level = "fullyAutomated"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
