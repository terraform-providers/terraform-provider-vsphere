package vsphere

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/virtualmachine"
)

func TestAccResourceVSphereVirtualMachineSnapshot_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachineSnapshotPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineSnapshotExists("vsphere_virtual_machine_snapshot.snapshot", false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineSnapshotConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualMachineSnapshotExists("vsphere_virtual_machine_snapshot.snapshot", true),
					resource.TestCheckResourceAttr(
						"vsphere_virtual_machine_snapshot.snapshot", "snapshot_name", "terraform-test-snapshot"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineSnapshotConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualMachineHasNoSnapshots("vsphere_virtual_machine.vm"),
				),
			},
		},
	})
}

func testAccResourceVSphereVirtualMachineSnapshotPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL") == "" {
		t.Skip("set TF_VAR_VSPHERE_RESOURCE_POOL to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_IPV4_ADDRESS") == "" {
		t.Skip("set TF_VAR_VSPHERE_IPV4_ADDRESS to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_IPV4_PREFIX") == "" {
		t.Skip("set TF_VAR_VSPHERE_IPV4_PREFIX to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_IPV4_GATEWAY") == "" {
		t.Skip("set TF_VAR_VSPHERE_IPV4_GATEWAY to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_virtual_machine_snapshot acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE to run vsphere_virtual_machine_snapshot acceptance tests")
	}
}

func snapshotExists(n string, s *terraform.State) (bool, error) {
	rs, ok := s.RootModule().Resources[n]

	if !ok {
		return false, nil
	}

	if rs.Primary.ID == "" {
		return false, fmt.Errorf("No Vm Snapshot ID is set")
	}
	client := testAccProvider.Meta().(*Client).vimClient

	vm, err := virtualmachine.FromUUID(client, rs.Primary.Attributes["virtual_machine_uuid"])
	if err != nil {
		return false, fmt.Errorf("error %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout) // This is 5 mins
	defer cancel()
	snapshot, err := vm.FindSnapshot(ctx, rs.Primary.ID)
	if err != nil {
		return false, fmt.Errorf("Error while getting the snapshot %v", snapshot)
	}

	return true, nil
}

func testAccCheckVirtualMachineSnapshotExists(n string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found, err := snapshotExists(n, s)
		if err != nil {
			return err
		}
		if found != exists {
			return fmt.Errorf("Snapshot exists error. expected state: %t, actual state: %t", exists, found)
		}

		return nil
	}
}

func testAccCheckVirtualMachineHasNoSnapshots(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VM ID is set")
		}
		client := testAccProvider.Meta().(*Client).vimClient

		vm, err := virtualmachine.FromUUID(client, rs.Primary.Attributes["uuid"])
		if err != nil {
			return fmt.Errorf("error %s", err)
		}
		props, err := virtualmachine.Properties(vm)
		if err != nil {
			return fmt.Errorf("cannot get properties for virtual machine: %s", err)
		}
		if props.Snapshot != nil {
			return fmt.Errorf("expected VM to not have snapshots, got %#v", props.Snapshot)
		}

		return nil
	}
}

func testAccResourceVSphereVirtualMachineSnapshotConfig(enabled bool) string {
	return fmt.Sprintf(`
%s

variable "ipv4_address" {
  default = "%s"
}

variable "ipv4_netmask" {
  default = "%s"
}

variable "ipv4_gateway" {
  default = "%s"
}

variable "template" {
  default = "%s"
}

variable "snapshot_enabled" {
  default = "%t"
}

data "vsphere_resource_pool" "pool" {
  name          = "${var.resource_pool}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_virtual_machine" "template" {
  name          = "${var.template}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 1024
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label = "disk0"
    size  = "${data.vsphere_virtual_machine.template.disks.0.size}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = true



      network_interface {
        ipv4_address = "${var.ipv4_address}"
        ipv4_netmask = "${var.ipv4_netmask}"
      }

      ipv4_gateway = "${var.ipv4_gateway}"
    }
  }
}

resource "vsphere_virtual_machine_snapshot" "snapshot" {
  count                = "${var.snapshot_enabled == "true" ? 1 : 0 }"
  virtual_machine_uuid = "${vsphere_virtual_machine.vm.uuid}"
  snapshot_name        = "terraform-test-snapshot"
  description          = "Managed by Terraform"
  memory               = true
  quiesce              = true
}
`,
		os.Getenv("TF_VAR_VSPHERE_IPV4_ADDRESS"),
		os.Getenv("TF_VAR_VSPHERE_IPV4_PREFIX"),
		os.Getenv("TF_VAR_VSPHERE_IPV4_GATEWAY"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		enabled,
	)
}
