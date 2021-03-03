package vsphere

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereVAppEntity_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppEntityPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppEntityCheckExists("vapp_entity", false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppEntityConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity", "powerOn"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity", 120),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity", "powerOff"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity", 120),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity", false),
				),
			},
			{
				ResourceName:      "vsphere_vapp_entity.vapp_entity",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceVSphereVAppEntity_nonDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppEntityPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppEntityCheckExists("vapp_entity", false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppEntityConfigNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity", "none"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity", 5),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity", "guestShutdown"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity", 5),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppEntity_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppEntityPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVAppEntityCheckExists("vapp_entity", false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppEntityConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity", "powerOn"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity", 120),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity", "powerOff"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity", 120),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity", false),
				),
			},
			{
				Config: testAccResourceVSphereVAppEntityConfigNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity", "none"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity", 5),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity", "guestShutdown"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity", 5),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppEntity_multi(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppEntityPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceVSphereVAppEntityCheckExists("vapp_entity1", false),
			testAccResourceVSphereVAppEntityCheckExists("vapp_entity2", false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppEntityConfigMultipleNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity1", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity1", "none"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity1", 5),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity1", "guestShutdown"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity1", 5),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity1", 2),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity1", true),
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity2", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity2", "none"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity2", 5),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity2", "guestShutdown"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity2", 5),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity2", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity2", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVAppEntity_multiUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVAppEntityPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceVSphereVAppEntityCheckExists("vapp_entity1", false),
			testAccResourceVSphereVAppEntityCheckExists("vapp_entity2", false),
		), Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVAppEntityConfigMultipleDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity1", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity1", "powerOn"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity1", 120),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity1", "powerOff"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity1", 120),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity1", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity1", false),
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity2", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity2", "powerOn"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity2", 120),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity2", "powerOff"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity2", 120),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity2", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity2", false),
				),
			},
			{
				Config: testAccResourceVSphereVAppEntityConfigMultipleNonDefault(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity1", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity1", "none"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity1", 5),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity1", "guestShutdown"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity1", 5),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity1", 2),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity1", true),
					testAccResourceVSphereVAppEntityCheckExists("vapp_entity2", true),
					testAccResourceVSphereVAppEntityStartAction("vapp_entity2", "none"),
					testAccResourceVSphereVAppEntityStartDelay("vapp_entity2", 5),
					testAccResourceVSphereVAppEntityStopAction("vapp_entity2", "guestShutdown"),
					testAccResourceVSphereVAppEntityStopDelay("vapp_entity2", 5),
					testAccResourceVSphereVAppEntityStartOrder("vapp_entity2", 1),
					testAccResourceVSphereVAppEntityWaitForGuest("vapp_entity2", true),
				),
			},
		},
	})
}

func testAccResourceVSphereVAppEntityPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_vapp_entity acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_vapp_entity acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI2") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI2 to run vsphere_vapp_entity acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_vapp_entity acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_vapp_entity acceptance tests")
	}
}

func testAccResourceVSphereVAppEntityCheckExists(name string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVAppEntity(s, name)
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected vapp_entity to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityStartAction(name string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ve, err := testGetVAppEntity(s, name)
		if err != nil {
			return err
		}
		if ve.StartAction != value {
			return fmt.Errorf("StartAction check failed. Expected: %s, got: %s", value, ve.StartAction)
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityStartDelay(name string, value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ve, err := testGetVAppEntity(s, name)
		if err != nil {
			return err
		}
		if ve.StartDelay != int32(value) {
			return fmt.Errorf("StartDelay check failed. Expected: %d, got: %d", value, ve.StartDelay)
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityStopAction(name string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ve, err := testGetVAppEntity(s, name)
		if err != nil {
			return err
		}
		if ve.StopAction != value {
			return fmt.Errorf("StopAction check failed. Expected: %s, got: %s", value, ve.StopAction)
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityStopDelay(name string, value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ve, err := testGetVAppEntity(s, name)
		if err != nil {
			return err
		}
		if ve.StopDelay != int32(value) {
			return fmt.Errorf("StopDelay check failed. Expected: %d, got: %d", value, ve.StartDelay)
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityStartOrder(name string, value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ve, err := testGetVAppEntity(s, name)
		if err != nil {
			return err
		}
		if ve.StartOrder != int32(value) {
			return fmt.Errorf("StartOrder check failed. Expected: %d, got: %d", value, ve.StartOrder)
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityWaitForGuest(name string, value bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ve, err := testGetVAppEntity(s, name)
		if err != nil {
			return err
		}
		if *ve.WaitingForGuest != value {
			return fmt.Errorf("WaitForGuest check failed. Expected: %t, got: %t", value, *ve.WaitingForGuest)
		}
		return nil
	}
}

func testAccResourceVSphereVAppEntityConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
	path = "parent_folder"
	type = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_vapp_entity" "vapp_entity" {
	target_id    = "${vsphere_virtual_machine.vm.moid}"
	container_id = "${vsphere_vapp_container.vapp_container.id}"
}

resource "vsphere_virtual_machine" "vm" {
	name = "terraform-virtual-machine-test"
	resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
	datastore_id = vsphere_nas_datastore.ds1.id

	num_cpus = 2
	memory   = 2048
	guest_id = "other3xLinux64Guest"
	wait_for_guest_net_timeout = -1

	
	disk {
		label = "disk0"
		size = "1"
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
	)
}

func testAccResourceVSphereVAppEntityConfigNonDefault() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "parent_folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_vapp_entity" "vapp_entity" {
  target_id      = "${vsphere_virtual_machine.vm.moid}"
  container_id   = "${vsphere_vapp_container.vapp_container.id}"
	start_action   = "none"
	start_delay    = 5
	stop_action    = "guestShutdown"
	stop_delay     = 5
	start_order    = 1
	wait_for_guest = true
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

func testAccResourceVSphereVAppEntityConfigMultipleDefault() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "parent_folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_vapp_entity" "vapp_entity1" {
  target_id      = "${vsphere_virtual_machine.vm1.moid}"
  container_id   = "${vsphere_vapp_container.vapp_container.id}"
}

resource "vsphere_vapp_entity" "vapp_entity2" {
  target_id      = "${vsphere_virtual_machine.vm2.moid}"
  container_id   = "${vsphere_vapp_container.vapp_container.id}"
}

resource "vsphere_virtual_machine" "vm1" {
  name             = "terraform-virtual-machine-test-1"
  resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 1024
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

resource "vsphere_virtual_machine" "vm2" {
  name             = "terraform-virtual-machine-test-2"
  resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 1024
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
func testAccResourceVSphereVAppEntityConfigMultipleNonDefault() string {
	return fmt.Sprintf(`
%s

resource "vsphere_resource_pool" "parent_resource_pool" {
  name                    = "terraform-resource-pool-test-parent"
  parent_resource_pool_id = "${data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id}"
}

resource "vsphere_folder" "parent_folder" {
  path          = "parent_folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp_container" {
  name                    = "terraform-resource-pool-test"
  parent_resource_pool_id = "${vsphere_resource_pool.parent_resource_pool.id}"
  parent_folder_id        = "${vsphere_folder.parent_folder.id}"
}

resource "vsphere_vapp_entity" "vapp_entity1" {
  target_id      = "${vsphere_virtual_machine.vm1.moid}"
  container_id   = "${vsphere_vapp_container.vapp_container.id}"
	start_action   = "none"
	start_delay    = 5
	stop_action    = "guestShutdown"
	stop_delay     = 5
	start_order    = 2
	wait_for_guest = true
}

resource "vsphere_vapp_entity" "vapp_entity2" {
  target_id      = "${vsphere_virtual_machine.vm2.moid}"
  container_id   = "${vsphere_vapp_container.vapp_container.id}"
	start_action   = "none"
	start_delay    = 5
	stop_action    = "guestShutdown"
	stop_delay     = 5
	start_order    = 1
	wait_for_guest = true
}

resource "vsphere_virtual_machine" "vm1" {
  name             = "terraform-virtual-machine-test-1"
  resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 1024
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

resource "vsphere_virtual_machine" "vm2" {
  name             = "terraform-virtual-machine-test-2"
  resource_pool_id = "${vsphere_vapp_container.vapp_container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 1024
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
