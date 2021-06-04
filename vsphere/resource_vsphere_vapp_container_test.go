package vsphere

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/vappcontainer"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereVAppContainer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerCheckFolder("parent_folder"),
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerCheckCPUReservation(10),
					testAccResourceVSphereVAppContainerCheckCPUExpandable(false),
					testAccResourceVSphereVAppContainerCheckCPULimit(20),
					testAccResourceVSphereVAppContainerCheckCPUShareLevel("custom"),
					testAccResourceVSphereVAppContainerCheckCPUShares(10),
					testAccResourceVSphereVAppContainerCheckCPUReservation(10),
					testAccResourceVSphereVAppContainerCheckCPUExpandable(false),
					testAccResourceVSphereVAppContainerCheckCPULimit(20),
					testAccResourceVSphereVAppContainerCheckMemoryShareLevel("custom"),
					testAccResourceVSphereVAppContainerCheckMemoryShares(10),
				),
			},
			{
				ResourceName:      "vsphere_vapp_container.vapp_container",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vc, err := testGetVAppContainer(s, "vapp_container")
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("/%s/host/%s/Resources/resource-pool-parent/%s",
						os.Getenv("TF_VAR_VSPHERE_DATACENTER"),
						os.Getenv("TF_VAR_VSPHERE_CLUSTER"),
						vc.Name(),
					), nil
				},
				Config: testAccResourceVSphereVAppContainerConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_childImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExistsInner("child", false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigChildImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExistsInner("parent", true),
				),
			},
			{
				ResourceName:      "vsphere_vapp_container.child",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vc, err := testGetVAppContainer(s, "child")
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("/%s/host/%s/Resources/parentVApp/%s",
						os.Getenv("TF_VAR_VSPHERE_DATACENTER"),
						os.Getenv("TF_VAR_VSPHERE_CLUSTER"),
						vc.Name(),
					), nil
				},
				Config: testAccResourceVSphereVAppContainerConfigChildImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExistsInner("child", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_vmBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigVM(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_vmMoveIntoVApp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigVMOutsideVApp(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVAppContainerConfigVM(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_vmSDRS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigVMSdrs(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_vmClone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigVMClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_vmCloneSDRS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigVMSdrsClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppContainer_vmMoveIntoVAppSDRS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppContainerPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppContainerCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppContainerConfigVMSdrsNoVApp(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
			{
				Config: testAccResourceVSphereVAppContainerConfigVMSdrs(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppContainerCheckExists(true),
					testAccResourceVSphereVAppContainerContainsVM("vm"),
				),
			},
		},
	})
}

func testAccResourceVSphereVAppContainerPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_vapp_container acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_vapp_container acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_vapp_container acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_vapp_container acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_PATH") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_PATH to run vsphere_vapp_container acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE to run vsphere_vapp_container acceptance tests")
	}
}

func testAccResourceVSphereVAppContainerCheckExists(expected bool) resource.TestCheckFunc {
	return testAccResourceVSphereVAppContainerCheckExistsInner("vapp_container", expected)
}

func testAccResourceVSphereVAppContainerCheckExistsInner(containerName string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVAppContainer(s, containerName)
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected vapp_container to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckFolder(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vc, err := testGetVAppContainer(s, "vapp_container")
		if err != nil {
			return err
		}
		f, err := testGetFolder(s, expected)
		if err != nil {
			return err
		}
		vcp, err := vappcontainer.Properties(vc)
		if err != nil {
			return err
		}
		if *vcp.ParentFolder != f.Reference() {
			return fmt.Errorf("expected path to be %s, got %s", expected, vcp.ParentFolder)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerContainsVM(vmName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testGetVirtualMachineProperties(s, vmName)
		if err != nil {
			return err
		}
		vc, err := testGetVAppContainer(s, "vapp_container")
		if err != nil {
			return err
		}
		if vm.ParentVApp != nil && vm.ParentVApp.Reference() != vc.Reference() {
			return fmt.Errorf("VM is not a part of vApp container")
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckCPUReservation(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if *props.Config.CpuAllocation.Reservation != *structure.Int64Ptr(int64(value)) {
			return fmt.Errorf("CpuAllocation.Reservation check failed. Expected: %d, got: %d", *props.Config.CpuAllocation.Reservation, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckCPUExpandable(value bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if *props.Config.CpuAllocation.ExpandableReservation != *structure.BoolPtr(value) {
			return fmt.Errorf("CpuAllocation.Expandable check failed. Expected: %t, got: %t", *props.Config.CpuAllocation.ExpandableReservation, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckCPULimit(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if *props.Config.CpuAllocation.Limit != *structure.Int64Ptr(int64(value)) {
			return fmt.Errorf("CpuAllocation.Limit check failed. Expected: %d, got: %d", *props.Config.CpuAllocation.Limit, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckCPUShareLevel(value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if string(props.Config.CpuAllocation.Shares.Level) != value {
			return fmt.Errorf("CpuAllocation.Shares.Level check failed. Expected: %s, got: %s", props.Config.CpuAllocation.Shares.Level, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckCPUShares(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if props.Config.CpuAllocation.Shares.Shares != int32(value) {
			return fmt.Errorf("CpuAllocation.Shares.Shares check failed. Expected: %d, got: %d", props.Config.CpuAllocation.Shares.Shares, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckMemoryShareLevel(value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if string(props.Config.MemoryAllocation.Shares.Level) != value {
			return fmt.Errorf("MemoryAllocation.Shares.Level check failed. Expected: %s, got: %s", props.Config.MemoryAllocation.Shares.Level, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerCheckMemoryShares(value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVAppContainerProperties(s, "vapp_container")
		if err != nil {
			return err
		}
		if props.Config.MemoryAllocation.Shares.Shares != int32(value) {
			return fmt.Errorf("MemoryAllocation.Shares.Shares check failed. Expected: %d, got: %d", props.Config.MemoryAllocation.Shares.Shares, value)
		}
		return nil
	}
}

func testAccResourceVSphereVAppContainerConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
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
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereVAppContainerConfigVMSdrsNoVApp() string {
	return fmt.Sprintf(`
%s

variable "nfs_path" {
  default = "%s"
}

variable "nas_host" {
  default = "%s"
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled  = true
}

resource "vsphere_nas_datastore" "datastore1" {
  name                 = "terraform-datastore-test1"
  host_system_ids      = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  type         = "NFS"
  remote_hosts = ["${var.nas_host}"]
  remote_path  = "${var.nfs_path}"
}

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name                 = "terraform-virtual-machine-test"
  resource_pool_id     = "${vsphere_resource_pool.parent_resource_pool.id}"
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1
  depends_on                 = ["vsphere_nas_datastore.datastore1"]

  disk {
    label = "disk0"
    size  = "1"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1()),

		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
	)
}

func testAccResourceVSphereVAppContainerConfigVMSdrs() string {
	return fmt.Sprintf(`
%s

variable "nfs_path" {
  default = "%s"
}

variable "nas_host" {
  default = "%s"
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled  = true
}

resource "vsphere_nas_datastore" "datastore1" {
  name                 = "terraform-datastore-test1"
  host_system_ids      = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  type         = "NFS"
  remote_hosts = ["${var.nas_host}"]
  remote_path  = "${var.nfs_path}"
}

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name                 = "terraform-virtual-machine-test"
  resource_pool_id     = "${vsphere_vapp_container.vapp_container.id}"
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1
  depends_on                 = ["vsphere_nas_datastore.datastore1"]

  disk {
    label = "disk0"
    size  = "1"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),

		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
	)
}

func testAccResourceVSphereVAppContainerConfigVMSdrsClone() string {
	x := fmt.Sprintf(`
%s

variable "nfs_path" {
  default = "%s"
}

variable "nas_host" {
  default = "%s"
}

variable "template" {
  default = "%s"
}

data "vsphere_virtual_machine" "template" {
  name          = "${var.template}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  sdrs_enabled  = true
}

resource "vsphere_nas_datastore" "datastore1" {
  name                 = "terraform-datastore-test1"
  host_system_ids      = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  type         = "NFS"
  remote_hosts = ["${var.nas_host}"]
  remote_path  = "${var.nfs_path}"
}

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name                 = "terraform-virtual-machine-test"
  resource_pool_id     = "${vsphere_vapp_container.vapp_container.id}"
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "ubuntu64Guest"
  wait_for_guest_net_timeout = -1
  depends_on                 = ["vsphere_nas_datastore.datastore1"]

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  disk {
    label = "disk0"
    size  = "32"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigResDS1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootVMNet(),
			testhelper.ConfigDataRootPortGroup1()),

		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
	)

	return x
}

func testAccResourceVSphereVAppContainerConfigVMClone() string {
	return fmt.Sprintf(`
%s

variable "template" {
  default = "%s"
}

data "vsphere_virtual_machine" "template" {
  name          = "${var.template}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-virtual-machine-test"
  resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "ubuntu64Guest"
  wait_for_guest_net_timeout = -1

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = true
  }

  disk {
    label = "disk0"
    size  = "%s"
  }

  cdrom {
    client_device = true
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }
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
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_CLONED_VM_DISK_SIZE"),
	)
}

func testAccResourceVSphereVAppContainerConfigVM() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-virtual-machine-test"
  resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  disk {
    label = "disk0"
    size  = "1"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereVAppContainerConfigVMOutsideVApp() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "resource-pool-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "vapp-container-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-virtual-machine-test"
  resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  disk {
    label = "disk0"
    size  = "1"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereVAppContainerConfigChildImport() string {
	return fmt.Sprintf(`
%s

resource "vsphere_folder" "parent_folder" {
  path          = "terraform-test-parent-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "parent" {
  name                    = "parentVApp"
  parent_resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
  parent_folder_id        = vsphere_folder.parent_folder.id
}

resource "vsphere_vapp_container" "child" {
  name                    = "childVApp"
  parent_resource_pool_id = vsphere_vapp_container.parent.id
}`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}
