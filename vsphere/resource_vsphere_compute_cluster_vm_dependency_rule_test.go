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
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereComputeClusterVMDependencyRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterVMDependencyRulePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterVMDependencyRuleExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						true,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group",
						"terraform-test-cluster-vm-group",
					),
				),
			},
			{
				ResourceName:      "vsphere_compute_cluster_vm_dependency_rule.cluster_vm_dependency_rule",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cluster, err := testGetComputeCluster(s, "rootcompute_cluster1", "data.vsphere_compute_cluster")
					if err != nil {
						return "", err
					}

					rs, ok := s.RootModule().Resources["vsphere_compute_cluster_vm_dependency_rule.cluster_vm_dependency_rule"]
					if !ok {
						return "", errors.New("no resource at address vsphere_compute_cluster_vm_dependency_rule.cluster_vm_dependency_rule")
					}
					name, ok := rs.Primary.Attributes["name"]
					if !ok {
						return "", errors.New("vsphere_compute_cluster_vm_dependency_rule.cluster_vm_dependency_rule has no name attribute")
					}

					m := make(map[string]string)
					m["compute_cluster_path"] = cluster.InventoryPath
					m["name"] = name
					b, err := json.Marshal(m)
					if err != nil {
						return "", err
					}

					return string(b), nil
				},
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						true,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group",
						"terraform-test-cluster-vm-group",
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeClusterVMDependencyRule_altGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterVMDependencyRulePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterVMDependencyRuleExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigAltGroup(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						true,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group2",
						"terraform-test-cluster-vm-group",
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeClusterVMDependencyRule_updateEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterVMDependencyRulePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterVMDependencyRuleExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						true,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group",
						"terraform-test-cluster-vm-group",
					),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						false,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group",
						"terraform-test-cluster-vm-group",
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeClusterVMDependencyRule_updateGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterVMDependencyRulePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterVMDependencyRuleExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						true,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group",
						"terraform-test-cluster-vm-group",
					),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterVMDependencyRuleConfigAltGroup(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMDependencyRuleExists(true),
					testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
						true,
						false,
						"terraform-test-cluster-vm-dependency-rule",
						"terraform-test-cluster-dependent-vm-group2",
						"terraform-test-cluster-vm-group",
					),
				),
			},
		},
	})
}

func testAccResourceVSphereComputeClusterVMDependencyRulePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_compute_cluster_vm_dependency_rule acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI1 to run vsphere_compute_cluster_vm_dependency_rule acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI2") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI2 to run vsphere_compute_cluster_vm_dependency_rule acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_compute_cluster_vm_dependency_rule acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_compute_cluster_vm_dependency_rule acceptance tests")
	}
}

func testAccResourceVSphereComputeClusterVMDependencyRuleExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		info, err := testGetComputeClusterVMDependencyRule(s, "cluster_vm_dependency_rule")
		if err != nil {
			if expected == false {
				if viapi.IsManagedObjectNotFoundError(err) {
					// This is not necessarily a missing rule, but more than likely a
					// missing cluster, which happens during destroy as the dependent
					// resources will be missing as well, so want to treat this as a
					// deleted rule as well.
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
			return errors.New("cluster rule missing when expected to exist")
		case !expected:
			return errors.New("cluster rule still present when expected to be missing")
		}

		return nil
	}
}

func testAccResourceVSphereComputeClusterVMDependencyRuleMatch(
	enabled bool,
	mandatory bool,
	name string,
	dependsOnVMGroup string,
	vmGroup string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterVMDependencyRule(s, "cluster_vm_dependency_rule")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("cluster rule missing")
		}

		expected := &types.ClusterDependencyRuleInfo{
			ClusterRuleInfo: types.ClusterRuleInfo{
				Enabled:      structure.BoolPtr(enabled),
				Mandatory:    structure.BoolPtr(mandatory),
				Name:         name,
				UserCreated:  structure.BoolPtr(true),
				InCompliance: actual.InCompliance,
				Key:          actual.Key,
				RuleUuid:     actual.RuleUuid,
				Status:       actual.Status,
			},
			DependsOnVmGroup: dependsOnVMGroup,
			VmGroup:          vmGroup,
		}

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereComputeClusterVMDependencyRuleConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_virtual_machine" "dependent_vm" {
  name             = "terraform-test-dependency"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_compute_cluster_vm_group" "cluster_vm_group" {
  name                = "terraform-test-cluster-vm-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.vm.id}"]
}

resource "vsphere_compute_cluster_vm_group" "dependent_vm_group" {
  name                = "terraform-test-cluster-dependent-vm-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.dependent_vm.id}"]
}

resource "vsphere_compute_cluster_vm_dependency_rule" "cluster_vm_dependency_rule" {
  compute_cluster_id       = data.vsphere_compute_cluster.rootcompute_cluster1.id
  name                     = "terraform-test-cluster-vm-dependency-rule"
  dependency_vm_group_name = "${vsphere_compute_cluster_vm_group.dependent_vm_group.name}"
  vm_group_name            = "${vsphere_compute_cluster_vm_group.cluster_vm_group.name}"
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigResResourcePool1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
			testhelper.ConfigResDS1()),
	)
}

func testAccResourceVSphereComputeClusterVMDependencyRuleConfigAltGroup() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_virtual_machine" "dependent_vm" {
  name             = "terraform-test-dependency"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_virtual_machine" "second_dependent_vm" {
  name             = "terraform-test-dependency2"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_compute_cluster_vm_group" "cluster_vm_group" {
  name                = "terraform-test-cluster-vm-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.vm.id}"]
}

resource "vsphere_compute_cluster_vm_group" "dependent_vm_group" {
  name                = "terraform-test-cluster-dependent-vm-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.dependent_vm.id}"]
}

resource "vsphere_compute_cluster_vm_group" "second_dependent_vm_group" {
  name                = "terraform-test-cluster-dependent-vm-group2"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.second_dependent_vm.id}"]
}

resource "vsphere_compute_cluster_vm_dependency_rule" "cluster_vm_dependency_rule" {
  compute_cluster_id       = data.vsphere_compute_cluster.rootcompute_cluster1.id
  name                     = "terraform-test-cluster-vm-dependency-rule"
  dependency_vm_group_name = "${vsphere_compute_cluster_vm_group.second_dependent_vm_group.name}"
  vm_group_name            = "${vsphere_compute_cluster_vm_group.cluster_vm_group.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterVMDependencyRuleConfigDisabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_virtual_machine" "dependent_vm" {
  name             = "terraform-test-dependency"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
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

resource "vsphere_compute_cluster_vm_group" "cluster_vm_group" {
  name                = "terraform-test-cluster-vm-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.vm.id}"]
}

resource "vsphere_compute_cluster_vm_group" "dependent_vm_group" {
  name                = "terraform-test-cluster-dependent-vm-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = ["${vsphere_virtual_machine.dependent_vm.id}"]
}

resource "vsphere_compute_cluster_vm_dependency_rule" "cluster_vm_dependency_rule" {
  compute_cluster_id       = data.vsphere_compute_cluster.rootcompute_cluster1.id
  name                     = "terraform-test-cluster-vm-dependency-rule"
  dependency_vm_group_name = "${vsphere_compute_cluster_vm_group.dependent_vm_group.name}"
  vm_group_name            = "${vsphere_compute_cluster_vm_group.cluster_vm_group.name}"
  enabled                  = false
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
