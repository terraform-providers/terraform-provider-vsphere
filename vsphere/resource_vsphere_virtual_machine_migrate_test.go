package vsphere

import (
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/datacenter"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/virtualmachine"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func testAccResourceVSphereVirtualMachineMigrateStatePreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC to run vsphere_virtual_machine state migration tests (provider connection is required)")
	}
	if os.Getenv("TF_VAR_VSPHERE_VM_V1_PATH") == "" {
		t.Skip("set TF_VAR_VSPHERE_VM_V1_PATH to run vsphere_virtual_machine state migration tests")
	}
}

func TestVSphereVirtualMachine_migrateStateV1(t *testing.T) {
	cases := map[string]struct {
		Attributes map[string]string
		Expected   map[string]string
	}{
		"skip_customization before 0.6.16": {
			Attributes: map[string]string{},
			Expected: map[string]string{
				"skip_customization": "false",
			},
		},
		"enable_disk_uuid before 0.6.16": {
			Attributes: map[string]string{},
			Expected: map[string]string{
				"enable_disk_uuid": "false",
			},
		},
		"disk controller_type": {
			Attributes: map[string]string{
				"disk.1234.size":            "0",
				"disk.5678.size":            "0",
				"disk.9999.size":            "0",
				"disk.9999.controller_type": "ide",
			},
			Expected: map[string]string{
				"disk.1234.size":            "0",
				"disk.1234.controller_type": "scsi",
				"disk.5678.size":            "0",
				"disk.5678.controller_type": "scsi",
				"disk.9999.size":            "0",
				"disk.9999.controller_type": "ide",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		if err := migrateVSphereVirtualMachineStateV1(is, nil); err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}

func TestAccResourceVSphereVirtualMachine_migrateStateV3_fromV2(t *testing.T) {
	testAccResourceVSphereVirtualMachineMigrateStatePreCheck(t)
	testAccPreCheck(t)

	meta, err := testAccProviderMeta(t)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	client := meta.(*Client).vimClient
	pth := os.Getenv("TF_VAR_VSPHERE_VM_V1_PATH")
	dc, err := datacenter.FromPath(client, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
	if err != nil {
		t.Fatalf("error while fetching datacenter: %s", err)
	}
	vm, err := virtualmachine.FromPath(client, pth, dc)
	if err != nil {
		t.Fatalf("error fetching virtual machine: %s", err)
	}
	props, err := virtualmachine.Properties(vm)
	if err != nil {
		t.Fatalf("error fetching virtual machine properties: %s", err)
	}

	disks := virtualdevice.SelectDisks(object.VirtualDeviceList(props.Config.Hardware.Device), 1, 0, 0)
	disk := disks[0].(*types.VirtualDisk)
	backing := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo)
	is := &terraform.InstanceState{
		ID: props.Config.Uuid,
		Attributes: map[string]string{
			"disk.#":     "1",
			"disk.0.key": strconv.Itoa(int(disk.Key)),
		},
	}
	is, err = resourceVSphereVirtualMachineMigrateState(2, is, meta)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if is.Attributes["disk.0.uuid"] != backing.Uuid {
		t.Fatalf("expected disk.0.uuid to be %q", backing.Uuid)
	}
}

func TestAccResourceVSphereVirtualMachine_migrateStateV3FromV1(t *testing.T) {
	testAccResourceVSphereVirtualMachineMigrateStatePreCheck(t)
	testAccPreCheck(t)

	meta, err := testAccProviderMeta(t)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	client := meta.(*Client).vimClient
	pth := os.Getenv("TF_VAR_VSPHERE_VM_V1_PATH")
	name := path.Base(pth)
	dc, err := datacenter.FromPath(client, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
	if err != nil {
		t.Fatalf("error while fetching datacenter: %s", err)
	}
	vm, err := virtualmachine.FromPath(client, pth, dc)
	if err != nil {
		t.Fatalf("error fetching virtual machine: %s", err)
	}
	props, err := virtualmachine.Properties(vm)
	if err != nil {
		t.Fatalf("error fetching virtual machine properties: %s", err)
	}

	is := &terraform.InstanceState{
		ID: name,
		Attributes: map[string]string{
			"uuid": props.Config.Uuid,
		},
	}
	is, err = resourceVSphereVirtualMachineMigrateState(1, is, meta)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if is.ID != props.Config.Uuid {
		t.Fatalf("expected ID to match %q, got %q", props.Config.Uuid, is.ID)
	}
	if is.Attributes["imported"] != "true" {
		t.Fatal("expected imported to be true")
	}
	if is.Attributes["disk.#"] != "1" {
		t.Fatal("expected disk count to be 1")
	}
	if is.Attributes["disk.0.key"] != "-1" {
		t.Fatal("expected disk.0.key to be -1")
	}
	if is.Attributes["disk.0.device_address"] != "scsi:0:0" {
		t.Fatal("expected disk.0.device_address to be scsi:0:0")
	}
	if is.Attributes["disk.0.label"] != "disk0" {
		t.Fatal("expected disk.0.label to be disk0")
	}
	if is.Attributes["disk.0.keep_on_remove"] != "true" {
		t.Fatal("expected disk.0.keep_on_remove to be true")
	}
}

func TestAccResourceVSphereVirtualMachine_migrateStateV2(t *testing.T) {
	testAccResourceVSphereVirtualMachineMigrateStatePreCheck(t)
	testAccPreCheck(t)

	meta, err := testAccProviderMeta(t)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	client := meta.(*Client).vimClient
	pth := os.Getenv("TF_VAR_VSPHERE_VM_V1_PATH")
	name := path.Base(pth)
	dc, err := datacenter.FromPath(client, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
	if err != nil {
		t.Fatalf("error while fetching datacenter: %s", err)
	}
	vm, err := virtualmachine.FromPath(client, pth, dc)
	if err != nil {
		t.Fatalf("error fetching virtual machine: %s", err)
	}
	props, err := virtualmachine.Properties(vm)
	if err != nil {
		t.Fatalf("error fetching virtual machine properties: %s", err)
	}

	is := &terraform.InstanceState{
		ID: name,
		Attributes: map[string]string{
			"uuid": props.Config.Uuid,
		},
	}
	// Start this at version 0 so we know it go through the whole path. There's
	// currently nothing from v0 to v2 that should hinder this.
	is, err = resourceVSphereVirtualMachineMigrateState(0, is, meta)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if is.ID != props.Config.Uuid {
		t.Fatalf("expected ID to match %q, got %q", props.Config.Uuid, is.ID)
	}
	if is.Attributes["imported"] != "true" {
		t.Fatal("expected imported to be true")
	}
	if is.Attributes["disk.#"] != "1" {
		t.Fatal("expected disk count to be 1")
	}
	if is.Attributes["disk.0.key"] != "-1" {
		t.Fatal("expected disk.0.key to be -1")
	}
	if is.Attributes["disk.0.device_address"] != "scsi:0:0" {
		t.Fatal("expected disk.0.device_address to be scsi:0:0")
	}
	if is.Attributes["disk.0.label"] != "disk0" {
		t.Fatal("expected disk.0.label to be disk0")
	}
	if is.Attributes["disk.0.keep_on_remove"] != "true" {
		t.Fatal("expected disk.0.keep_on_remove to be true")
	}
}

func TestComputeInstanceMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceVSphereVirtualMachineMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = resourceVSphereVirtualMachineMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
