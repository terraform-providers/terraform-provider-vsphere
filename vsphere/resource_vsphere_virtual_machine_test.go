package vsphere

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/computeresource"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/datastore"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/resourcepool"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/virtualdisk"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	testAccResourceVSphereVirtualMachineDiskNameExtraVmdk = "terraform-test-vm-extra-disk.vmdk"
	testAccResourceVSphereVirtualMachineStaticMacAddr     = "06:5c:89:2b:a0:64"
	testAccResourceVSphereVirtualMachineAnnotation        = "Managed by Terraform"
)

func TestAccResourceVSphereVirtualMachine_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "moid", regexp.MustCompile("^vm-")),
				),
			},
			{
				ResourceName:      "vsphere_virtual_machine.vm",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disk",
					"imported",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vm, err := testGetVirtualMachine(s, "vm")
					if err != nil {
						return "", err
					}
					return vm.InventoryPath, nil
				},
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_TestAccResourceVSphereVirtualMachine_hardwareVersionBare(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(15),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "hardware_version", regexp.MustCompile("^15$")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_hardwareVersionUpgrade(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(14),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "hardware_version", regexp.MustCompile("^14$")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(15),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "hardware_version", regexp.MustCompile("^15$")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_hardwareVersionInvalidVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				ExpectError: regexp.MustCompile("expected hardware_version to be in the range"),
				Config:      testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(1),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "hardware_version", regexp.MustCompile("^14$")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_hardwareVersionDowngrade(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(14),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "hardware_version", regexp.MustCompile("^14$")),
				),
			},
			{
				ExpectError: regexp.MustCompile("cannot downgrade virtual machine hardware version"),
				Config:      testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(13),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_TestAccResourceVSphereVirtualMachine_hardwareVersionClone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigHardwareVersionClone(15),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "hardware_version", regexp.MustCompile("^15$")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachineContentLibrary_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherevirtualmachineconfigcontentlibraryBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestMatchResourceAttr("vsphere_virtual_machine.vm", "moid", regexp.MustCompile("^vm-")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_ignoreValidationOnComputedValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:             testAccResourceVSphereVirtualMachineConfigComputedValue(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_highLatencySensitivity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigHighSensitivity(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckLatencySensitivity(types.LatencySensitivitySensitivityLevelHigh),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_ESXiOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
			testAccSkipIfNotEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasicESXiOnly(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_shutdownOK(t *testing.T) {
	var state *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				PreConfig: func() {
					if err := testPowerOffVM(state, "vm"); err != nil {
						panic(err)
					}
				},
				PlanOnly: true,
				Config:   testAccResourceVSphereVirtualMachineConfigBasic(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_reCreateOnDeletion(t *testing.T) {
	var state *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					copyState(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				PreConfig: func() {
					if err := testDeleteVM(state, "vm"); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					func(s *terraform.State) error {
						oldID := state.RootModule().Resources["vsphere_virtual_machine.vm"].Primary.ID
						return testCheckResourceNotAttr("vsphere_virtual_machine.vm", "id", oldID)(s)
					},
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_multiDevice(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiDevice(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckMultiDevice([]bool{true, true, true}, []bool{true, true, true}),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_addDevices(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiDevice(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckMultiDevice([]bool{true, true, true}, []bool{true, true, true}),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_removeMiddleDevices(t *testing.T) {
	var state *terraform.State
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiDevice(),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckMultiDevice([]bool{true, true, true}, []bool{true, true, true}),
				),
			},
			{
				PreConfig: func() {
					// As sometimes the OS image that we are using to test "bare metal"
					// changes in how well it integrates VMware tools, we power down the
					// VM for this operation. This is not necessarily checking that a
					// hot-remove operation happened so it's not essential it's powered
					// on.
					if err := testPowerOffVM(state, "vm"); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereVirtualMachineConfigRemoveMiddle(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckMultiDevice([]bool{true, false, true}, []bool{true, false, true}),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_removeMiddleDevicesChangeDiskUnit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiDevice(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckMultiDevice([]bool{true, true, true}, []bool{true, true, true}),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigRemoveMiddleChangeUnit(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckMultiDevice([]bool{true, false, true}, []bool{true, false, true}),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_highDiskUnitNumbers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiHighBus(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test.vmdk", 0, 0),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_1.vmdk", 0, 1),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_2.vmdk", 1, 0),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_3.vmdk", 1, 1),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_4.vmdk", 2, 0),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_5.vmdk", 2, 1),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_highDiskUnitInsufficientBus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigMultiHighBusInsufficientBus(),
				ExpectError: regexp.MustCompile("unit_number on disk \"disk1\" too high \\(15\\) - maximum value is 14 with 1 SCSI controller\\(s\\)"),
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_highDiskUnitsToRegularSingleController(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiHighBus(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test.vmdk", 0, 0),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_1.vmdk", 0, 1),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_2.vmdk", 1, 0),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiDevice(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test.vmdk", 0, 0),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_1.vmdk", 0, 1),
					testAccResourceVSphereVirtualMachineCheckDiskBus("testacc-test_2.vmdk", 0, 2),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_scsiBusSharing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigSharedSCSIBus(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckSCSIBusSharing(string(types.VirtualSCSISharingPhysicalSharing)),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_scsiBusSharingUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckSCSIBusSharing(string(types.VirtualSCSISharingNoSharing)),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigSharedSCSIBus(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckSCSIBusSharing(string(types.VirtualSCSISharingPhysicalSharing)),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_disksKeepOnRemove(t *testing.T) {
	var disks []map[string]string
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigKeepDisksOnRemove(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachinePersistentDiskInfo(&disks),
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(false),
					testAccResourceVSphereVirtualMachineDeletePersistentDisks(&disks),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cdromClientMapping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdrom(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckClientCdrom(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppIsoBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSpherevirtualMachineCheckHostname("custom-hostname"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppIsoNoVApp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromClone(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppIsoNoCdrom(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigNoCdromCloneIsoVApp(),
				ExpectError: regexp.MustCompile("this virtual machine requires a client CDROM device to deliver vApp properties"),
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppIsoConfigIsoIgnored(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppIsoChangeCdromBacking(t *testing.T) {
	var state *terraform.State
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp(),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSpherevirtualMachineCheckHostname("custom-hostname"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppIsoPoweredOffCdromRead(t *testing.T) {
	var state *terraform.State
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp(),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				PreConfig: func() {
					if err := testPowerOffVM(state, "vm"); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vvtdAndVbs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigVbsEnabledAndVvtdEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckVVTD(true),
					testAccResourceVSphereVirtualMachineCheckVBS(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cdromNoParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigNoCdromParameters(),
				ExpectError: regexp.MustCompile("Either client_device or datastore_id and path must be set"),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClientCdrom(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cdromConflictingParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigConflictingCdromParameters(),
				ExpectError: regexp.MustCompile("Cannot have both client_device parameter and ISO file parameters"),
			},
			{
				Config: testAccResourceVSphereEmpty,
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_maximumNumberOfNICs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMaxNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckNICCount(10),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_upgradeCPUAndRam(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBeefy(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCPUMem(4, 8192),
					// Since hot-add should be off, we expect that the VM was powered
					// off as a part of this step. This helps check the functionality
					// of the check for later tests as well.
					testAccResourceVSphereVirtualMachineCheckPowerOffEvent(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_modifyAnnotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasicAnnotation(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckAnnotation(),
					testAccResourceVSphereVirtualMachineCheckPowerOffEvent(false),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_growDisk(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigGrowDisk(10),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckDiskSize(10),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigGrowDisk(20),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckDiskSize(20),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_swapSCSIBus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckSCSIBus(virtualdevice.SubresourceControllerTypeParaVirtual),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigLsiLogicSAS(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckSCSIBus(virtualdevice.SubresourceControllerTypeLsiLogicSAS),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_extraConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigExtraConfig("foo", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckExtraConfig("foo", "bar"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_extraConfigSwapKeys(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigExtraConfig("foo", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckExtraConfig("foo", "bar"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigExtraConfig("baz", "qux"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckExtraConfig("baz", "qux"),
					testAccResourceVSphereVirtualMachineCheckExtraConfigKeyMissing("foo"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_attachExistingVmdk(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigExistingVmdk(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckExistingVmdk(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_attachExistingVmdkTaint(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigExistingVmdk(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckExistingVmdk(),
				),
			},
			{
				Taint:  []string{"vsphere_virtual_machine.vm"},
				Config: testAccResourceVSphereVirtualMachineConfigExistingVmdk(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckExistingVmdk(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_resourcePoolMove(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckResourcePool("testacc-resource-pool1"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigNewResourcePool(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckResourcePool("terraform-test-new-resource-pool"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppContainerAndFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigVAppAndFolder(),
				ExpectError: regexp.MustCompile("cannot set folder while VM is in a vApp container"),
			},
			{
				Config: testAccResourceVSphereEmpty,
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_vAppContainerMove(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigOutOfVAppContainer(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckFolder("terraform-test-vms"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigInVAppContainer(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckResourcePool("terraform-vapp-test"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigOutOfVAppContainer(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckFolder("terraform-test-vms"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_inFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigInFolder(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckFolder("terraform-test-vms"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_moveToFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigInFolder(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckFolder("terraform-test-vms"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_staticMAC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigStaticMAC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckStaticMACAddr(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_multipleTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_switchTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_renamedDiskInPlaceOfExisting(t *testing.T) {
	var state *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				PreConfig: func() {
					if err := testRenameVMFirstDisk(state, "vm", "foobar.vmdk"); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					// The only real way we can check to see if this is actually
					// functional in the current test framework is by checking that
					// the file we renamed to was not deleted (this is due to a lack
					// of ability to check diff in the test framework right now).
					testCheckVMDiskFileExists("testacc-test.vmdk"),
					testCheckVMDiskFileExists("foobar.vmdk"),
				),
			},
			// The last step is a cleanup step. This assumes the test is
			// functional as the orphaned disk will be now detached and not
			// deleted when the VM is destroyed.
			{
				PreConfig: func() {
					if err := testDeleteVMDisk(state, "foobar.vmdk"); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_blockComputedDiskName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigComputedDisk(),
				ExpectError: regexp.MustCompile("disk label or name must be defined and cannot be computed"),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_blockVAppSettingsOnNonClones(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineVAppPropertiesNonClone(),
				ExpectError: regexp.MustCompile("vApp properties can only be set on cloned virtual machines"),
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_blockVAppSettingsOnNonClonesAfterCreation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config:      testAccResourceVSphereVirtualMachineVAppPropertiesNonClone(),
				ExpectError: regexp.MustCompile("this VM lacks a vApp configuration and cannot have vApp properties set on it"),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_blockDiskLabelStartingWithOrphanedPrefix(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigBadOrphanedLabel(),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`disk label "orphaned_disk_0" cannot start with "orphaned_disk_"`)),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_createIntoEmptyClusterNoEnvironmentBrowser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigBasicEmptyCluster(),
				ExpectError: regexp.MustCompile("compute resource .* is missing an Environment Browser\\. Check host, cluster, and vSphere license health of all associated resources and try again"),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneFromTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_clonePoweredOn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClonePoweredOn(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneCustomizeWithNewResourcePool(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneWithNewResourcePool(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneCustomizeForceNewWithDatastore(t *testing.T) {
	var state *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneParameterized(
					os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME"),
					"testacc-test",
				),
				Check: resource.ComposeTestCheckFunc(
					copyState(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHostname("testacc-test"),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneParameterized(
					os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"),
					"terraform-test-renamed",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHostname("terraform-test-renamed"),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					func(s *terraform.State) error {
						oldID := state.RootModule().Resources["vsphere_virtual_machine.vm"].Primary.ID
						return testCheckResourceNotAttr("vsphere_virtual_machine.vm", "id", oldID)(s)
					},
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneModifyDiskAndSCSITypeAtSameTime(t *testing.T) {
	var state *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneChangeDiskAndSCSI(),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
					func(s *terraform.State) error {
						oldSize, _ := strconv.Atoi(state.RootModule().Resources["data.vsphere_virtual_machine.template"].Primary.Attributes["disks.0.size"])
						newSize := oldSize * 2
						return resource.TestCheckResourceAttr("vsphere_virtual_machine.vm", "disk.0.size", strconv.Itoa(newSize))(s)
					},
					func(s *terraform.State) error {
						oldBus := state.RootModule().Resources["data.vsphere_virtual_machine.template"].Primary.Attributes["scsi_type"]
						var expected string
						if oldBus == virtualdevice.SubresourceControllerTypeParaVirtual {
							expected = virtualdevice.SubresourceControllerTypeLsiLogicSAS
						} else {
							expected = virtualdevice.SubresourceControllerTypeParaVirtual
						}
						return resource.TestCheckResourceAttr("vsphere_virtual_machine.vm", "scsi_type", expected)(s)
					},
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneMultiNICFromSingleNICTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneMultiNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneWithDifferentTimezone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneTimeZone("America/Vancouver"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneBlockESXi(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
			testAccSkipIfNotEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigClone(),
				ExpectError: regexp.MustCompile("use of the clone sub-resource block requires vCenter"),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneWithBadTimezone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigCloneTimeZone("Pacific Standard Time"),
				ExpectError: regexp.MustCompile("must be similar to America/Los_Angeles or other Linux/Unix TZ format"),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

// Temporarily removed until https://github.com/hashicorp/terraform/issues/21225 is resolved.
// func TestAccResourceVSphereVirtualMachine_cloneWithBadThinProvisionedWithLinkedClone(t *testing.T) {
//        t.Cleanup(RunSweepers)
//        resource.Test(t, resource.TestCase{
//                PreCheck: func() {
//			testAccPreCheck(t)
//			testAccResourceVSphereVirtualMachinePreCheck(t)
//		},
//		Providers: testAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config:      testAccResourceVSphereVirtualMachineConfigBadThin(),
//				ExpectError: regexp.MustCompile("must have same value for thin_provisioned as source when using linked_clone"),
//				PlanOnly:    true,
//			},
//			{
//				Config: testAccResourceVSphereEmpty,
//				Check:  resource.ComposeTestCheckFunc(),
//			},
//		},
//	})
//}

func TestAccResourceVSphereVirtualMachine_cloneWithBadSizeWithLinkedClone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigBadSizeLinked(),
				ExpectError: regexp.MustCompile("must be the exact size of source when using linked_clone"),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneWithBadSizeWithoutLinkedClone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereVirtualMachineConfigBadSizeUnlinked(),
				ExpectError: regexp.MustCompile("must be at least the same size of source when cloning"),
				PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneIntoEmptyCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneEmptyClusterNoVM(),
			},
			{
				Config:      testAccResourceVSphereVirtualMachineConfigCloneEmptyCluster(),
				ExpectError: regexp.MustCompile("compute resource .* is missing an Environment Browser\\. Check host, cluster, and vSphere license health of all associated resources and try again"),
				//PlanOnly:    true,
			},
			{
				Config: testAccResourceVSphereEmpty,
				Check:  resource.ComposeTestCheckFunc(),
			}},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneWithDifferentHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigCloneHostname(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHostname("terraform-test-renamed"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cpuHotAdd(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				// Starting config
				Config: testAccResourceVSphereVirtualMachineConfigWithHotAdd(2, 2048, true, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCPUMem(2, 2048),
				),
			},
			{
				// Add CPU w/hot-add
				Config: testAccResourceVSphereVirtualMachineConfigWithHotAdd(4, 2048, true, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCPUMem(4, 2048),
					testAccResourceVSphereVirtualMachineCheckPowerOffEvent(false),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_memoryHotAdd(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				// Starting config
				Config: testAccResourceVSphereVirtualMachineConfigWithHotAdd(2, 2048, true, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCPUMem(2, 2048),
				),
			},
			{
				// Add memory with hot-add
				Config: testAccResourceVSphereVirtualMachineConfigWithHotAdd(2, 3072, true, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCPUMem(2, 3072),
					testAccResourceVSphereVirtualMachineCheckPowerOffEvent(false),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_dualStackIPv4AndIPv6(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigDualStack(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckNet("fd00::2", "32", "fd00::1"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_hostCheck(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigHostCheck(os.Getenv("TF_VAR_VSPHERE_ESXI1")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHost(os.Getenv("TF_VAR_VSPHERE_ESXI1")),
				),
			},
			{
				Config: testAccResourceVSphereEmpty,
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigHostCheck(os.Getenv("TF_VAR_VSPHERE_ESXI2")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHost(os.Getenv("TF_VAR_VSPHERE_ESXI2")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_hostVMotion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigHostVMotion(os.Getenv("TF_VAR_VSPHERE_ESXI1")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHost(os.Getenv("TF_VAR_VSPHERE_ESXI1")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigHostVMotion(os.Getenv("TF_VAR_VSPHERE_ESXI2")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckHost(os.Getenv("TF_VAR_VSPHERE_ESXI2")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_resourcePoolVMotion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigResourcePoolVMotion(os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckResourcePool(os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigResourcePoolVMotion(fmt.Sprintf("%s/Resources", os.Getenv("TF_VAR_VSPHERE_CLUSTER"))),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckResourcePool(fmt.Sprintf("%s/Resources", os.Getenv("TF_VAR_VSPHERE_CLUSTER"))),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_storageVMotionGlobalSetting(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionGlobal(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionGlobal(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_storageVMotionSingleDisk(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionSingleDisk(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(1, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionSingleDisk(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(1, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_storageVMotionPinDatastore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionPinDatastore("vsphere_nas_datastore.ds1.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(1, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionPinDatastore("data.vsphere_datastore.rootds1.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(1, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_storageVMotionRenamedVirtualMachine(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionRename("testacc-test", os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionRename("testacc-foobar-test", os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionRename("foobar-test", os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_storageVMotionLinkedClones(t *testing.T) {
	var state *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionLinkedClone("data.vsphere_datastore.rootds1.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionLinkedClone("vsphere_nas_datastore.ds1.id"),
				Check: resource.ComposeTestCheckFunc(
					copyStatePtr(&state),
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
					func(s *terraform.State) error {
						return testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))(s)
					},
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_storageVMotionBlockExternallyAttachedDisks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionAttachedDisk(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckVmxDatastore(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
					testAccResourceVSphereVirtualMachineCheckVmdkDatastore(0, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigStorageVMotionAttachedDisk(os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2")),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(
					fmt.Sprintf("externally attached disk %q cannot be migrated", testAccResourceVSphereVirtualMachineDiskNameExtraVmdk),
				)),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_singleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigWithCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_multiCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigWithMultiCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_switchCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigWithCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigWithMultiCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccResourceVSphereVirtualMachineCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_multipleDisksAtDifferentSCSISlotsImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigMultiHighBus(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				ResourceName:      "vsphere_virtual_machine.vm",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disk",
					"imported",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vm, err := testGetVirtualMachine(s, "vm")
					if err != nil {
						return "", err
					}
					return vm.InventoryPath, nil
				},
				Config: testAccResourceVSphereVirtualMachineConfigMultiHighBus(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_cloneImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				ResourceName:      "vsphere_virtual_machine.vm",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disk",
					"imported",
					"clone",
					"cdrom",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vm, err := testGetVirtualMachine(s, "vm")
					if err != nil {
						return "", err
					}
					return vm.InventoryPath, nil
				},
				Config: testAccResourceVSphereVirtualMachineConfigClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					func(s *terraform.State) error {
						// This simulates an import scenario, as ImportStateVerify does not
						// actually do a full TF run after import, and hence the above import
						// check does not actually test to see Terraform will be able to
						// plan. Hence we actually remove the clone configuration from the
						// state and ensure that imported is flagged. This allows the next
						// step to properly simulate the post-imported state.
						rs, ok := s.RootModule().Resources["vsphere_virtual_machine.vm"]
						if !ok {
							return errors.New("vsphere_virtual_machine.vm not found in state")
						}
						for k := range rs.Primary.Attributes {
							if strings.HasPrefix(k, "clone") {
								delete(rs.Primary.Attributes, k)
							}
						}
						rs.Primary.Attributes["imported"] = "true"

						return nil
					},
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_interpolatedDisk(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
			{
				Config: testAccResourceVSphereVirtualMachineTestPathInterpolation(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_deployOvfFromUrl(t *testing.T) {
	vmName := "terraform_test_vm_" + acctest.RandStringFromCharSet(4, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineDeployOvfFromURL(vmName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_virtual_machine.vm", "name", vmName),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualMachine_deployOvaFromUrl(t *testing.T) {
	vmName := "terraform_test_vm_" + acctest.RandStringFromCharSet(4, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccResourceVSphereVirtualMachinePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVirtualMachineDeployOvaFromURL(vmName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_virtual_machine.vm", "name", vmName),
				),
			},
			{
				Config: testAccResourceVSphereVirtualMachineConfigBase(),
			},
		},
	})
}

func testAccResourceVSphereVirtualMachinePreCheck(t *testing.T) {
	// Note that TF_VAR_VSPHERE_USE_LINKED_CLONE is also a variable and its presence
	// speeds up tests greatly, but it's not a necessary variable, so we don't
	// enforce it here.
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_CLUSTER to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_RESOURCE_POOL") == "" {
		t.Skip("set TF_VAR_VSPHERE_RESOURCE_POOL to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NETWORK_LABEL to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE_ISO_TRANSPORT to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE_WINDOWS to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE_NONUSER_VAPP to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_TEMPLATE") == "" {
		t.Skip("set TF_VAR_VSPHERE_TEMPLATE_COREOS to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI_HOST to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI2") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI_HOST2 to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NAS_HOST") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAS_HOST to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_PATH") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_PATH to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_CONTENT_LIBRARY_FILES") == "" {
		t.Skip("set TF_VAR_VSPHERE_CONTENT_LIBRARY_FILES to run vsphere_virtual_machine acceptance tests")
	}
}

func testAccResourceVSphereVirtualMachineCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVirtualMachine(s, "vm")
		if err != nil {
			missingState, _ := regexp.MatchString("not found in state", err.Error())
			missingVSphere, _ := regexp.MatchString("virtual machine with UUID \"[-a-f0-9]+\" not found", err.Error())
			if missingState && !expected || missingVSphere && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected VM to be missing")
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckPowerState is a check to check for
// a VirtualMachine's power state.

// testAccResourceVSphereVirtualMachineCheckHostname is a check to check for a
// VirtualMachine's hostname. The check uses guest info, so VMware tools needs
// to be installed.
func testAccResourceVSphereVirtualMachineCheckHostname(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		actual := props.Guest.HostName
		if expected != actual {
			return fmt.Errorf("expected hostname to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckExtraDisks is a check for proper
// parameters on the vsphere_virtual_machine extra disks test. This is a very
// specific check that checks for the specific disk devices and respective
// backings, and expects them in the exact order outlined in the function.

// testAccResourceVSphereVirtualMachineCheckDiskBus is a check that looks for a
// disk with a specific name at a specific SCSI bus number and unit number.
func testAccResourceVSphereVirtualMachineCheckDiskBus(name string, expectedBus, expectedUnit int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}

		for _, dev := range props.Config.Hardware.Device {
			if disk, ok := dev.(*types.VirtualDisk); ok {
				if info, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
					dp := new(object.DatastorePath)
					if ok := dp.FromString(info.FileName); !ok {
						return fmt.Errorf("could not parse datastore path %q", info.FileName)
					}
					if path.Base(dp.Path) != name {
						continue
					}
					l := object.VirtualDeviceList(props.Config.Hardware.Device)
					ctlr := l.FindByKey(disk.ControllerKey)
					if ctlr == nil {
						return fmt.Errorf("could not find controller with key %d for disk %q", disk.ControllerKey, name)
					}
					sc, ok := ctlr.(types.BaseVirtualSCSIController)
					if !ok {
						return fmt.Errorf("disk %q not attached to a SCSI controller (actual: %T)", name, ctlr)
					}
					if sc.GetVirtualSCSIController().BusNumber != int32(expectedBus) {
						return fmt.Errorf("disk %q: Expected controller bus to be %d, got %d", name, expectedBus, sc.GetVirtualSCSIController().BusNumber)
					}
					if disk.UnitNumber == nil {
						return fmt.Errorf("disk %q has no unit number", name)
					}
					if *disk.UnitNumber != int32(expectedUnit) {
						return fmt.Errorf("disk %q: Expected unit number to be %d, got %d", name, expectedUnit, *disk.UnitNumber)
					}
					return nil
				}
			}
		}

		return fmt.Errorf("could not find disk path %q", name)
	}
}

// testAccResourceVSphereVirtualMachineCheckFolder checks to make sure a
// virtual machine's folder matches the folder supplied with expected.
func testAccResourceVSphereVirtualMachineCheckFolder(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testGetVirtualMachine(s, "vm")
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		expected, err := folder.RootPathParticleVM.PathFromNewRoot(vm.InventoryPath, folder.RootPathParticleVM, expected)
		actual := path.Dir(vm.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckExistingVmdk is a check to make
// sure that the appropriate disk is attached in the existing VMDK test.
func testAccResourceVSphereVirtualMachineCheckExistingVmdk() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}

		expected := testAccResourceVSphereVirtualMachineDiskNameExtraVmdk

		for _, dev := range props.Config.Hardware.Device {
			if disk, ok := dev.(*types.VirtualDisk); ok {
				if info, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
					if strings.HasSuffix(info.FileName, expected) {
						return nil
					}
				}
			}
		}

		return fmt.Errorf("could not find attached disk matching %q", expected)
	}
}

// testAccResourceVSphereVirtualMachineCheckCPUMem checks the CPU and RAM for a
// VM.
func testAccResourceVSphereVirtualMachineCheckCPUMem(expectedCPU, expectedMem int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		actualCPU := props.Summary.Config.NumCpu
		actualMem := props.Summary.Config.MemorySizeMB
		if expectedCPU != actualCPU {
			return fmt.Errorf("expected CPU count to be %d, got %d", expectedCPU, actualCPU)
		}
		if expectedMem != actualMem {
			return fmt.Errorf("expected memory size to be %d, got %d", expectedMem, actualMem)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckNet checks to make sure a virtual
// machine's primary NIC has the given IP address and netmask assigned to it,
// and that the appropriate gateway is present.
//
// This uses VMware tools to check this, so it needs to be installed on the
// guest.
func testAccResourceVSphereVirtualMachineCheckNet(expectedAddr, expectedPrefix, expectedGW string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, err := strconv.ParseInt(expectedPrefix, 10, 32)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		expectedPrefixInt := int32(res)

		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		var v4gw, v6gw net.IP
		for _, s := range props.Guest.IpStack {
			if s.IpRouteConfig != nil {
				for _, r := range s.IpRouteConfig.IpRoute {
					switch r.Network {
					case "0.0.0.0":
						v4gw = net.ParseIP(r.Gateway.IpAddress)
					case "::":
						v6gw = net.ParseIP(r.Gateway.IpAddress)
					}
				}
			}
		}
		for _, n := range props.Guest.Net {
			if n.IpConfig != nil {
				for _, addr := range n.IpConfig.IpAddress {
					ip := net.ParseIP(addr.IpAddress)
					if !ip.Equal(net.ParseIP(expectedAddr)) {
						continue
					}
					if addr.PrefixLength != expectedPrefixInt {
						continue
					}
					var mask net.IPMask
					if ip.To4() != nil {
						mask = net.CIDRMask(int(addr.PrefixLength), 32)
					} else {
						mask = net.CIDRMask(int(addr.PrefixLength), 128)
					}
					switch {
					case v4gw != nil && ip.Mask(mask).Equal(v4gw.Mask(mask)):
						if net.ParseIP(expectedGW).Equal(v4gw) {
							return nil
						}
					case v6gw != nil && ip.Mask(mask).Equal(v6gw.Mask(mask)):
						if net.ParseIP(expectedGW).Equal(v6gw) {
							return nil
						}
					}
				}
			}
		}

		return fmt.Errorf("could not find IP %s/%s, gateway %s", expectedAddr, expectedPrefix, expectedGW)
	}
}

// testAccResourceVSphereVirtualMachineCheckNetDeviceOrder checks to make sure a virtual
// machine with multiple NICs has the given IP address and netmask assigned to it,
// and that the order of the NICs correspond to the declared order.
//
// This uses VMware tools to check this, so it needs to be installed on the
// guest.

func testAccResourceVSpherevirtualMachineCheckHostname(hostname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		if props.Guest.HostName != hostname {
			return fmt.Errorf("expected host name: %s, got %s", hostname, props.Guest.HostName)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckStaticMACAddr is a check to look
// for the MAC address defined in the static MAC address test on the first
// network interface.
func testAccResourceVSphereVirtualMachineCheckStaticMACAddr() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		l := object.VirtualDeviceList(props.Config.Hardware.Device)
		devices := l.Select(func(device types.BaseVirtualDevice) bool {
			if _, ok := device.(types.BaseVirtualEthernetCard); ok {
				return true
			}
			return false
		})
		if devices[0].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().AddressType != string(types.VirtualEthernetCardMacTypeManual) {
			return fmt.Errorf("first network interface is not set to manual address type")
		}
		actual := devices[0].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().MacAddress
		expected := testAccResourceVSphereVirtualMachineStaticMacAddr
		if expected != actual {
			return fmt.Errorf("expected MAC address to be %q, got %q", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckAnnotation is a check to ensure
// that a VM's annotation is correctly set in the annotation test.
func testAccResourceVSphereVirtualMachineCheckAnnotation() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		expected := testAccResourceVSphereVirtualMachineAnnotation
		actual := props.Config.Annotation
		if expected != actual {
			return fmt.Errorf("expected annotation to be %q, got %q", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckCustomizationSucceeded is a check
// to ensure that events have been received for customization success on a VM.

// testAccResourceVSphereVirtualMachineCheckTags is a check to ensure that any
// tags that have been created with supplied resource name have been attached
// to the virtual machine.
func testAccResourceVSphereVirtualMachineCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testGetVirtualMachine(s, "vm")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, vm, tagResName)
	}
}

// testAccResourceVSphereVirtualMachineCheckMultiDevice is a check for proper
// parameters on the vsphere_virtual_machine multi-device test. This is a very
// specific check that checks for the specific disk and network devices. The
// configuration that this test asserts should be in the
// testAccResourceVSphereVirtualMachineConfigMultiDevice resource.
func testAccResourceVSphereVirtualMachineCheckMultiDevice(expectedD, expectedN []bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}

		actualD := make([]bool, 3)
		actualN := make([]bool, 3)
		expectedDisk0Size := structure.GiBToByte(20)
		expectedDisk1Size := structure.GiBToByte(10)
		expectedDisk2Size := structure.GiBToByte(5)
		expectedNet0Level := types.SharesLevelNormal
		expectedNet1Level := types.SharesLevelHigh
		expectedNet2Level := types.SharesLevelLow

		for _, dev := range props.Config.Hardware.Device {
			if disk, ok := dev.(*types.VirtualDisk); ok {
				switch {
				case disk.CapacityInBytes == expectedDisk0Size:
					actualD[0] = true
				case disk.CapacityInBytes == expectedDisk1Size:
					actualD[1] = true
				case disk.CapacityInBytes == expectedDisk2Size:
					actualD[2] = true
				}
			}
			if bvec, ok := dev.(types.BaseVirtualEthernetCard); ok {
				card := bvec.GetVirtualEthernetCard()
				switch {
				case card.ResourceAllocation.Share.Level == expectedNet0Level:
					actualN[0] = true
				case card.ResourceAllocation.Share.Level == expectedNet1Level:
					actualN[1] = true
				case card.ResourceAllocation.Share.Level == expectedNet2Level:
					actualN[2] = true
				}
			}
		}

		for n, actual := range actualD {
			if actual != expectedD[n] {
				return fmt.Errorf("expected disk at index %d to be %t, got %t", n, expectedD[n], actual)
			}
		}
		for n, actual := range actualN {
			if actual != expectedN[n] {
				return fmt.Errorf("expected network interface at index %d to be %t, got %t", n, expectedN[n], actual)
			}
		}

		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckIsoCdrom checks to make sure that the
// subject VM has a CDROM device configured with iso backing and is connected.

// testAccResourceVSphereVirtualMachineCheckClientCdrom checks to make sure that the
// subject VM has a CDROM device mapped to a client device.
func testAccResourceVSphereVirtualMachineCheckClientCdrom() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}

		for _, dev := range props.Config.Hardware.Device {
			if cdrom, ok := dev.(*types.VirtualCdrom); ok {
				if backing, ok := cdrom.Backing.(*types.VirtualCdromRemoteAtapiBackingInfo); ok {
					useAutoDetect := false
					expected := &types.VirtualCdromRemoteAtapiBackingInfo{
						VirtualDeviceRemoteDeviceBackingInfo: types.VirtualDeviceRemoteDeviceBackingInfo{
							UseAutoDetect: &useAutoDetect,
							DeviceName:    "",
						},
					}
					if !reflect.DeepEqual(expected, backing) {
						return fmt.Errorf("expected %#v, got %#v", expected, backing)
					}
					return nil
				}
				return errors.New("could not find CDROM with correct backing device")
			}
		}
		return errors.New("could not locate CDROM device on VM")
	}
}

// testAccResourceVSphereVirtualMachinePersistentDiskInfo goes through the
// current state and creates a slice of maps containing information on disks
// which have `keep_on_remove` set to true.  This list can later be used to
// examine disks that have been removed from the virtual machine configuration.
func testAccResourceVSphereVirtualMachinePersistentDiskInfo(disks *[]map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vms := s.RootModule().Resources["vsphere_virtual_machine.vm"].Primary.Attributes
		dn, err := strconv.Atoi(vms["disk.#"])
		if err != nil {
			return err
		}
		for i := 0; i < dn; i++ {
			if vms[fmt.Sprintf("disk.%s.keep_on_remove", strconv.Itoa(i))] == "true" {
				disk := map[string]string{
					"path":         vms[fmt.Sprintf("disk.%s.path", strconv.Itoa(i))],
					"datastore_id": vms[fmt.Sprintf("disk.%s.datastore_id", strconv.Itoa(i))],
				}
				*disks = append(*disks, disk)
			}
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineDeletePersistentDisks goes through a
// list of disks and deletes the files backing those disks. This process also
// checks that the backing files exist in the deletion process. If the files
// don't exist, an error will be raised during deletion. The folder containing
// the disks will also be deleted. If the folder is not empty, all remaining
// files will be deleted.
func testAccResourceVSphereVirtualMachineDeletePersistentDisks(disks *[]map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client).vimClient
		reFlat := regexp.MustCompile("\\.vmdk$")
		reVM := regexp.MustCompile("\\/.*?\\.vmdk$")
		var vmFolder string
		var dsID string
		for _, disk := range *disks {
			ds, err := datastore.FromID(client, disk["datastore_id"])
			if err != nil {
				return err
			}
			dsFilePath := fmt.Sprintf("[%s] %s", ds.Name(), disk["path"])
			flat := reFlat.ReplaceAllString(dsFilePath, "-flat.vmdk")
			vmFolder = reVM.ReplaceAllString(dsFilePath, "")
			dsID = disk["datastore_id"]
			err = testDeleteDatastoreFile(client, dsID, dsFilePath)
			if err != nil {
				return err
			}
			// Ignore errors here as the _flat files only exist for thin provisioned
			// disks.
			_ = testDeleteDatastoreFile(client, dsID, flat)
		}
		// Delete the VM folder now that we've checked and cleaned up the disk.
		err := testDeleteDatastoreFile(client, dsID, vmFolder)
		if err != nil {
			return err
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckPowerOffEvent is a check to see if
// the VM has been powered off at any point in time.
func testAccResourceVSphereVirtualMachineCheckPowerOffEvent(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testGetVirtualMachine(s, "vm")
		if err != nil {
			return err
		}
		client := testAccProvider.Meta().(*Client).vimClient
		actual, err := selectEventsForReference(client, vm.Reference(), []string{eventTypeVMPoweredOffEvent})
		if err != nil {
			return err
		}
		switch {
		case len(actual) < 1 && expected:
			return errors.New("expected power off, VM was not powered off")
		case len(actual) > 1 && !expected:
			return errors.New("VM was powered off when it should not have been")
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckDiskSize checks the first
// VirtualDisk it encounters for a specific size in GiB. It should only be used
// with test configurations with a single disk attached.
func testAccResourceVSphereVirtualMachineCheckDiskSize(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}

		expectedBytes := structure.GiBToByte(expected)

		for _, dev := range props.Config.Hardware.Device {
			if disk, ok := dev.(*types.VirtualDisk); ok {
				if expectedBytes != disk.CapacityInBytes {
					return fmt.Errorf("expected disk size to be %d, got %d", expectedBytes, disk.CapacityInBytes)
				}
			}
		}

		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckSCSIBus checks to make sure the
// test VM's SCSI bus is all of the specified SCSI type.
func testAccResourceVSphereVirtualMachineCheckSCSIBus(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetVirtualMachineSCSIBusType(s, "vm")
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected SCSI bus to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckSCSIBusSharing checks to make sure the
// test VM's SCSI bus is all of the specified sharing type.
func testAccResourceVSphereVirtualMachineCheckSCSIBusSharing(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetVirtualMachineSCSIBusSharing(s, "vm")
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected SCSI bus sharing to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckHost checks to make sure the
// test VM is currently located on a specific host.
func testAccResourceVSphereVirtualMachineCheckHost(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hs, err := testGetVirtualMachineHost(s, "vm")
		if err != nil {
			return err
		}
		actual := hs.Name()
		if expected != actual {
			return fmt.Errorf("expected host to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckResourcePool checks to make sure the
// test VM is currently located in a specific resource pool.
func testAccResourceVSphereVirtualMachineCheckResourcePool(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pool, err := testGetVirtualMachineResourcePool(s, "vm")
		if err != nil {
			return err
		}

		actual := pool.Name()
		if actual == "Resources" && path.Base(expected) == "Resources" {
			client := testAccProvider.Meta().(*Client).vimClient
			expectedCluster, err := computeresource.BaseFromPath(client, path.Dir(expected))
			if err != nil {
				return err
			}
			pprops, err := resourcepool.Properties(pool)
			if err != nil {
				return err
			}
			actualCluster, err := computeresource.BaseFromReference(client, *pprops.Parent)
			if err != nil {
				return err
			}
			if expectedCluster.Reference().Value != actualCluster.Reference().Value {
				return fmt.Errorf("expected default resource pool of %q, got default resource pool of %q", expectedCluster.Reference().Value, actualCluster.Reference().Value)
			}
			return nil
		}
		if expected != actual {
			return fmt.Errorf("expected resource pool or to be %q, got %q", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckExtraConfig checks a key/expected
// value combination in a VM's config.
func testAccResourceVSphereVirtualMachineCheckExtraConfig(key, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		for _, bov := range props.Config.ExtraConfig {
			ov := bov.GetOptionValue()
			if ov.Key == key {
				if ov.Value.(string) == expected {
					return nil
				}
				return fmt.Errorf("expected key %s to be %s, got %s", key, expected, ov.Value.(string))
			}
		}
		return fmt.Errorf("key %s not found", key)
	}
}

// testAccResourceVSphereVirtualMachineCheckExtraConfigKeyMissing checks to
// make sure that a key is missing in the VM's extraConfig.
func testAccResourceVSphereVirtualMachineCheckExtraConfigKeyMissing(key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		for _, bov := range props.Config.ExtraConfig {
			ov := bov.GetOptionValue()
			if ov.Key == key {
				return fmt.Errorf("expected key %s to be missing", key)
			}
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckVmxDatastore checks the datastore
// that the virtual machine's configuration is currently located.
func testAccResourceVSphereVirtualMachineCheckVmxDatastore(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		var dsPath object.DatastorePath
		if ok := dsPath.FromString(props.Config.Files.VmPathName); !ok {
			return fmt.Errorf("could not parse datastore path %q", props.Config.Files.VmPathName)
		}
		actual := dsPath.Datastore
		if expected != actual {
			return fmt.Errorf("expected VM configuration to be in datastore %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckVmdkDatastore checks the datastore
// that a specific VMDK file is in.
func testAccResourceVSphereVirtualMachineCheckVmdkDatastore(diskIndex int, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tVars, err := testClientVariablesForResource(s, "vsphere_virtual_machine.vm")
		if err != nil {
			return err
		}
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		name := tVars.resourceAttributes[fmt.Sprintf("disk.%d.path", diskIndex)]
		for _, dev := range props.Config.Hardware.Device {
			if disk, ok := dev.(*types.VirtualDisk); ok {
				if info, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
					var dsPath object.DatastorePath
					if ok := dsPath.FromString(info.FileName); !ok {
						return fmt.Errorf("could not parse datastore path %q", info.FileName)
					}
					if dsPath.Path == name {
						actual := dsPath.Datastore
						if expected == actual {
							return nil
						}
						return fmt.Errorf("expected disk name %q to be on datastore %q, got %q", name, expected, actual)
					}
				}
			}
		}
		return fmt.Errorf("could not find disk %q", name)
	}
}

// testAccResourceVSphereVirtualMachineCheckVmxDatastoreCluster checks the
// datastore cluster that the virtual machine's configuration is currently
// located.

func testAccResourceVSphereVirtualMachineCheckVVTD(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		vvtdEnabled := *props.Config.Flags.VvtdEnabled
		if vvtdEnabled != expected {
			return fmt.Errorf("vvtd flag was %t, expected: %t", vvtdEnabled, expected)
		}
		return nil
	}
}

func testAccResourceVSphereVirtualMachineCheckVBS(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		vbsEnabled := *props.Config.Flags.VbsEnabled
		if vbsEnabled != expected {
			return fmt.Errorf("vbs flag was %t, expected: %t", vbsEnabled, expected)
		}
		return nil
	}
}

// testAccResourceVSphereVirtualMachineCheckVmdkDatastoreCluster checks the
// datastore cluster that a specific VMDK file is in.

// testAccResourceVSphereVirtualMachineCheckNICCount checks the number of NICs
// on the virtual machine.
func testAccResourceVSphereVirtualMachineCheckNICCount(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}

		var actual int
		for _, dev := range props.Config.Hardware.Device {
			if _, ok := dev.(types.BaseVirtualEthernetCard); ok {
				actual++
			}
		}
		if expected != actual {
			return fmt.Errorf("expected %d number of NICs, got %d", expected, actual)
		}
		return nil
	}
}

// testCheckVMDiskFileExists takes a VMDK filename and checks to see if it
// exists within the same directory as the virtual machine's VMX file.
func testCheckVMDiskFileExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tVars, err := testClientVariablesForResource(s, "vsphere_virtual_machine.vm")
		if err != nil {
			return err
		}
		vm, err := testGetVirtualMachine(s, "vm")
		if err != nil {
			return err
		}
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		vmxPath, success := virtualdisk.DatastorePathFromString(props.Config.Files.VmPathName)
		if !success {
			return fmt.Errorf("could not parse VMX path %q", props.Config.Files.VmPathName)
		}
		dcp, err := folder.RootPathParticleVM.SplitDatacenter(vm.InventoryPath)
		if err != nil {
			return err
		}
		dc, err := getDatacenter(tVars.client, dcp)
		if err != nil {
			return err
		}
		ds, err := datastore.FromPath(tVars.client, vmxPath.Datastore, dc)
		if err != nil {
			return err
		}
		p := path.Join(path.Dir(vmxPath.Path), name)
		exists, err := datastore.FileExists(ds, p)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("file %q does not exist in datastore %q", p, ds.Name())
		}
		return nil
	}
}

func testAccResourceVSphereVirtualMachineCheckCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_virtual_machine", "vm", props.Entity())
	}
}

// testAccResourceVSphereVirtualMachineCheckLatencySensitivity checks a virtual
// machine's latency sensitivity value.
func testAccResourceVSphereVirtualMachineCheckLatencySensitivity(
	expected types.LatencySensitivitySensitivityLevel,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		actual := props.Config.LatencySensitivity.Level
		if expected != actual {
			return fmt.Errorf("expected latency sensitivity to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereVirtualMachineConfigBase() string {
	return testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigDataRootDS1(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
		testhelper.ConfigDataRootPortGroup1())
}

func testAccResourceVSphereVirtualMachineConfigComputedValue() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = "${vsphere_nas_datastore.ds1.id}"

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigHardwareVersionClone(hw int) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus         = 2
  memory           = 2048
  guest_id         = "${data.vsphere_virtual_machine.template.guest_id}"
  hardware_version = %d

  wait_for_guest_net_timeout = -1

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		hw,
	)
}

func testAccResourceVSphereVirtualMachineConfigBareHardwareVersion(hw int) string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus         = 2
  memory           = 2048
  guest_id         = "other3xLinux64Guest"
  hardware_version = %d

  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		hw,
	)
}

func testAccResourceVSphereVirtualMachineConfigBasic() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = vsphere_resource_pool.pool1.id
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = data.vsphere_network.network1.id
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigSharedSCSIBus() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  scsi_bus_sharing = "physicalSharing"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigKeepDisksOnRemove() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  wait_for_guest_net_timeout = -1

  disk {
    label            = "disk0"
    unit_number      = 0
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  disk {
    label            = "disk1"
    unit_number      = 1
    size             = 1
    thin_provisioned = true
    keep_on_remove   = true
  }

  disk {
    label            = "disk2"
    unit_number      = 2
    size             = 1
    thin_provisioned = true
    keep_on_remove   = true
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigBasicESXiOnly() string {
	return fmt.Sprintf(`

%s  // Mix and match config

data "vsphere_datacenter" "dc" {}

data "vsphere_resource_pool" "pool" {}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigMultiDevice() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "normal"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "high"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "low"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  disk {
    label       = "disk1"
    unit_number = 1
    size        = 10
  }

  disk {
    label       = "disk2"
    unit_number = 2
    size        = 5
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigMultiHighBusInsufficientBus() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  scsi_controller_count = 1

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "normal"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "high"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "low"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  disk {
    label       = "disk1"
    unit_number = 15
    size        = 10
  }

  disk {
    label       = "disk2"
    unit_number = 31
    size        = 5
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigMultiHighBus() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  scsi_controller_count = 3

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "normal"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "high"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "low"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  disk {
    label       = "disk1"
    unit_number = 1
    size        = 10
  }

  disk {
    label       = "disk2"
    unit_number = 15
    size        = 5
  }

  disk {
    label       = "disk3"
    unit_number = 16
    size        = 5
  }
  
  disk {
    label       = "disk4"
    unit_number = 30
    size        = 5
  }
  
  disk {
    label       = "disk5"
    unit_number = 31
    size        = 5
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigRemoveMiddle() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "normal"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "low"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  disk {
    label       = "disk2"
    unit_number = 2
    size        = 5
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigRemoveMiddleChangeUnit() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "normal"
  }

  network_interface {
    network_id            = "${data.vsphere_network.network1.id}"
    bandwidth_share_level = "low"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  disk {
    label       = "disk2"
    unit_number = 1
    size        = 5
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigNoCdromCloneIsoVApp() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  wait_for_guest_net_timeout = 10

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  vapp {
    properties = {
      hostname  = "custom-hostname"
    }
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigClientCdromCloneIsoVApp() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  wait_for_guest_net_timeout = 10

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  cdrom {
    client_device = true
  }

  vapp {
    properties = {
      hostname  = "custom-hostname"
    }
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigClientCdromClone() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  cdrom {
    client_device = true
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigClientCdrom() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
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

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigNoCdromParameters() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
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

  cdrom {}
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigConflictingCdromParameters() string {
	return fmt.Sprintf(`


%s  // Mix and match config

variable "iso_datastore" {
  default = "%s"
}

variable "iso_path" {
  default = "%s"
}

data "vsphere_datastore" "iso_datastore" {
  name          = "${var.iso_datastore}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
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

  cdrom {
    datastore_id  = "${data.vsphere_datastore.iso_datastore.id}"
    path          = "${var.iso_path}"
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_ISO_DATASTORE"),
		os.Getenv("TF_VAR_VSPHERE_ISO_FILE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigBeefy() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 4
  memory   = 8192
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigMaxNIC() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigBasicAnnotation() string {
	return fmt.Sprintf(`


%s  // Mix and match config

variable "annotation" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus   = 2
  memory     = 2048
  guest_id   = "other3xLinux64Guest"
  annotation = "${var.annotation}"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		testAccResourceVSphereVirtualMachineAnnotation,
	)
}

func testAccResourceVSphereVirtualMachineConfigGrowDisk(size int) string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = %d
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		size,
	)
}

func testAccResourceVSphereVirtualMachineConfigLsiLogicSAS() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  scsi_type = "lsilogic-sas"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigExtraConfig(k, v string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  extra_config = {
    %s = "%s"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		k, v,
	)
}

func testAccResourceVSphereVirtualMachineConfigExistingVmdk() string {
	return fmt.Sprintf(`


%s  // Mix and match config

variable "extra_vmdk_name" {
  default = "%s"
}

resource "vsphere_virtual_disk" "disk" {
  size         = 1
  vmdk_path    = "${var.extra_vmdk_name}"
  datacenter   = data.vsphere_datacenter.rootdc1.name
  datastore    = data.vsphere_datastore.rootds1.name
  type         = "thin"
  adapter_type = "lsiLogic"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  disk {
    label        = "disk1"
    datastore_id = data.vsphere_datastore.rootds1.id
    path         = "${vsphere_virtual_disk.disk.vmdk_path}"
    disk_mode    = "independent_persistent"
    attach       = true
    unit_number  = 1
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		testAccResourceVSphereVirtualMachineDiskNameExtraVmdk,
	)
}

func testAccResourceVSphereVirtualMachineConfigInFolder() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_folder" "folder" {
  path          = "terraform-test-vms"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id
  folder           = "${vsphere_folder.folder.path}"

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigInVAppContainer() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_folder" "folder" {
  path          = "terraform-test-vms"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp-container" {
  name = "terraform-vapp-test"

  parent_resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  parent_folder_id        = "${vsphere_folder.folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_vapp_container.vapp-container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigNewResourcePool() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_resource_pool" "pool" {
  name                    = "terraform-test-new-resource-pool"
  parent_resource_pool_id = "${vsphere_resource_pool.pool1.id}"
}

resource "vsphere_folder" "folder" {
  path          = "terraform-test-vms"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id
  folder           = "${vsphere_folder.folder.path}"

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigVAppAndFolder() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_folder" "folder" {
  path          = "terraform-test-vms"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp-container" {
  name = "terraform-vapp-test"

  parent_resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  parent_folder_id        = "${vsphere_folder.folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_vapp_container.vapp-container.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id
  folder           = "${vsphere_folder.folder.path}"

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigVbsEnabledAndVvtdEnabled() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  vbs_enabled             = true
  firmware                = "efi"
  vvtd_enabled            = true
  nested_hv_enabled       = true
  efi_secure_boot_enabled = true

  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,
		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigOutOfVAppContainer() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_folder" "folder" {
  path          = "terraform-test-vms"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_vapp_container" "vapp-container" {
  name = "terraform-vapp-test"

  parent_resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  parent_folder_id        = "${vsphere_folder.folder.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id
  folder           = "${vsphere_folder.folder.path}"

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "other3xLinux64Guest"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigStaticMAC() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id     = "${data.vsphere_network.network1.id}"
    use_static_mac = true
    mac_address    = "%s"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		testAccResourceVSphereVirtualMachineStaticMacAddr,
	)
}

func testAccResourceVSphereVirtualMachineConfigSingleTag() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VirtualMachine",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  tags = [
    "${vsphere_tag.testacc-tag.id}",
  ]
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigMultiTag() string {
	return fmt.Sprintf(`


%s  // Mix and match config

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
    "VirtualMachine",
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

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  tags = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigComputedDisk() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "random_pet" "pet" {}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-test-${random_pet.pet.id}"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "terraform-test-${random_pet.pet.id}.vmdk"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineVAppPropertiesNonClone() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }

  vapp {
    properties = {
      foo = "bar"
    }
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigBadOrphanedLabel() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "orphaned_disk_0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneWithNewResourcePool() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_resource_pool" "pool" {
  name                    = "terraform-test-resource-pool"
  parent_resource_pool_id = "${vsphere_resource_pool.pool1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  wait_for_guest_net_timeout = -1

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"

	customize {
      linux_options {
        host_name = "testacc-test"
        domain   = "testdomain.internal"
      }
      network_interface {}
    }
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigClone() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigClonePoweredOn() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm_source" {
  name             = "terraform-test1"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "${data.vsphere_virtual_machine.template.guest_id}"
	wait_for_guest_net_timeout = -1

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-test2"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "${data.vsphere_virtual_machine.template.guest_id}"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${vsphere_virtual_machine.vm_source.id}"
    linked_clone  = "false"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigBadSizeLinked() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = 999
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = true
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigBadSizeUnlinked() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = 1
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneMultiNIC() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneTimeZone(zone string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "time_zone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"

    customize {
      linux_options {
        host_name = "testacc-test"
        domain    = "test.internal"
        time_zone = "${var.time_zone}"
      }
      network_interface {}
    }
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		zone,
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneHostname() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"

    customize {
      linux_options {
        host_name = "terraform-test-renamed"
        domain    = "test.internal"
      }
      network_interface {}
    }
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigWithHotAdd(nc, nm int, cha, chr, mha bool) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                  = %d
  memory                    = %d
  cpu_hot_add_enabled       = %t
  cpu_hot_remove_enabled    = %t
  memory_hot_add_enabled    = %t
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		nc,
		nm,
		cha,
		chr,
		mha,
	)
}

func testAccResourceVSphereVirtualMachineConfigDualStack() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
    customize {
      linux_options {
        host_name = "ipv6test"
        domain    = "testdomain.com"
      }
      network_interface {
        ipv6_address = "fd00::2"
        ipv6_netmask = "32"
      }
      ipv6_gateway = "fd00::1"
    }
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigHostCheck(host string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

variable "linked_clone" {
  default = "%s"
}

variable "host" {
  default = "%s"
}

data "vsphere_host" "host" {
  name          = "${var.host}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  host_system_id   = data.vsphere_host.host.id
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus                   = 2
  memory                     = 2048
  guest_id                   = "ubuntu64Guest"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label            = "disk0"
    size             = "1"
    eagerly_scrub    = false
    thin_provisioned = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		host,
	)
}

func testAccResourceVSphereVirtualMachineConfigHostVMotion(host string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "host" {
  default = "%s"
}

data "vsphere_host" "host" {
  name          = "${var.host}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  host_system_id   = data.vsphere_host.host.id
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		host,
	)
}

func testAccResourceVSphereVirtualMachineConfigResourcePoolVMotion(pool string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "resource_pool" {
  default = "%s"
}

variable "linked_clone" {
  default = "%s"
}

data "vsphere_resource_pool" "pool" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = var.resource_pool
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = data.vsphere_resource_pool.pool.id
  datastore_id     = data.vsphere_datastore.rootds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  wait_for_guest_net_timeout = -1

  network_interface {
    network_id   = "${data.vsphere_network.vmnet.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }
}
`,

		testhelper.CombineConfigs(
			testAccResourceVSphereVirtualMachineConfigBase(),
			testhelper.ConfigDataRootVMNet()),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		pool,
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigStorageVMotionGlobal(datastore string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		datastore,
	)
}

func testAccResourceVSphereVirtualMachineConfigStorageVMotionSingleDisk(datastore string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "disk_datastore" {
  default = "%s"
}

data "vsphere_datastore" "disk_datastore" {
  name          = "${var.disk_datastore}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = data.vsphere_datastore.rootds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  disk {
    label        = "disk1"
    datastore_id = "${data.vsphere_datastore.disk_datastore.id}"
    size         = 1
    unit_number  = 1
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		datastore,
	)
}

func testAccResourceVSphereVirtualMachineConfigStorageVMotionPinDatastore(datastoreAddress string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "ds_id" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = var.ds_id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"
  wait_for_guest_net_timeout = -1

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
    datastore_id     = var.ds_id
  }

  disk {
    label        = "disk1"
    datastore_id = %s
    size         = 1
    unit_number  = 1
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = true
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		datastoreAddress, datastoreAddress,
	)
}

func testAccResourceVSphereVirtualMachineConfigStorageVMotionRename(name, datastore string) string {
	return fmt.Sprintf(`
variable "vm_name" {
  default = "%s"
}

%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "${var.vm_name}"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
	datastore_id     = data.vsphere_datastore.ds.id
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  cdrom {
    client_device = true
  }
}
`,

		name,
		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		datastore,
	)
}

func testAccResourceVSphereVirtualMachineConfigStorageVMotionLinkedClone(datastoreAddress string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "ds" {
  default = %s
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = var.ds

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
    datastore_id     = var.ds
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
    linked_clone  = true
  }

  cdrom {
    client_device = true
  }
}
`,
		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		datastoreAddress,
	)
}

func testAccResourceVSphereVirtualMachineConfigStorageVMotionAttachedDisk(datastore string) string {
	return fmt.Sprintf(`


%s  // Mix and match config

variable "datastore" {
  default = "%s"
}

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "extra_vmdk_name" {
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = var.datastore  
}

resource "vsphere_virtual_disk" "disk" {
  size         = 1
  vmdk_path    = "${var.extra_vmdk_name}"
  datacenter   = data.vsphere_datacenter.rootdc1.name
  datastore    = data.vsphere_datastore.ds.name
  type         = "thin"
  adapter_type = "lsiLogic"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = 2
  memory   = 2048
  guest_id = "ubuntu64Guest"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  disk {
    label        = "disk1"
    datastore_id = data.vsphere_datastore.ds.id
    path         = "${vsphere_virtual_disk.disk.vmdk_path}"
    disk_mode    = "independent_persistent"
    attach       = true
    unit_number  = 1
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  cdrom {
    client_device = true
  }
}
`,
		testAccResourceVSphereVirtualMachineConfigBase(),
		datastore,
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		testAccResourceVSphereVirtualMachineDiskNameExtraVmdk,
	)
}

func testAccResourceVSphereVirtualMachineConfigWithCustomAttribute() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "VirtualMachine"
}

locals {
  vm_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id
  num_cpus         = 2
  memory           = 2048
  guest_id         = "${data.vsphere_virtual_machine.template.guest_id}"
  scsi_type        = "${data.vsphere_virtual_machine.template.scsi_type}"

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
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }

  custom_attributes = "${local.vm_attrs}"
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigWithMultiCustomAttribute() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "custom_attrs" {
  default = [
    "testacc-attribute-1",
    "terraform-test-attriubute-2",
  ]
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "VirtualMachine"
}

resource "vsphere_custom_attribute" "testacc-attribute-alt" {
  count               = "${length(var.custom_attrs)}"
  name                = "${var.custom_attrs[count.index]}"
  managed_object_type = "VirtualMachine"
}

locals {
  vm_attrs = {
    "${vsphere_custom_attribute.testacc-attribute-alt.0.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-alt.1.id}" = "value-2"
  }
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id
  num_cpus         = 2
  memory           = 2048
  guest_id         = "${data.vsphere_virtual_machine.template.guest_id}"
  scsi_type        = "${data.vsphere_virtual_machine.template.scsi_type}"

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
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"
  }

  cdrom {
    client_device = true
  }

  custom_attributes = "${local.vm_attrs}"
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneChangeDiskAndSCSI() string {
	return fmt.Sprintf(`


%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  scsi_type = "${data.vsphere_virtual_machine.template.scsi_type == "pvscsi" ? "lsilogic-sas" : "pvscsi"}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label = "disk0"
    size  = "${data.vsphere_virtual_machine.template.disks.0.size * 2}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigBasicEmptyCluster() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "empty_cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  host_managed  = true

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = true
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = vsphere_compute_cluster.compute_cluster.resource_pool_id
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneEmptyClusterNoVM() string {
	return fmt.Sprintf(`
%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "empty_cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  host_managed  = true

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = true
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneEmptyCluster() string {
	return fmt.Sprintf(`
%s  // Mix and match config

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "empty_cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  host_managed  = true

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = true
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = vsphere_compute_cluster.compute_cluster.resource_pool_id
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  network_interface {
    network_id   = "${data.vsphere_network.network1.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    label            = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  cdrom {
    client_device = true
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
	)
}

func testAccResourceVSphereVirtualMachineConfigCloneParameterized(datastore, hostname string) string {
	return fmt.Sprintf(`
%s  // Mix and match config

data "vsphere_datastore" "ds" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

data "vsphere_virtual_machine" "template" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

variable "linked_clone" {
  default = "%s"
}

variable "hostname" {
  default = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = vsphere_resource_pool.pool1.id
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = 2
  memory   = 2048
  guest_id = data.vsphere_virtual_machine.template.guest_id

  network_interface {
    network_id   = data.vsphere_network.network1.id
    adapter_type = data.vsphere_virtual_machine.template.network_interface_types[0]
  }

  disk {
    label            = "disk0"
    size             = data.vsphere_virtual_machine.template.disks.0.size
    eagerly_scrub    = data.vsphere_virtual_machine.template.disks.0.eagerly_scrub
    thin_provisioned = data.vsphere_virtual_machine.template.disks.0.thin_provisioned
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
    linked_clone  = "${var.linked_clone != "" ? "true" : "false" }"

    customize {
      linux_options {
        host_name = var.hostname
        domain    = "test.internal"
      }
    network_interface {}
    }
  }

  cdrom {
    client_device = true
  }
  depends_on = ["vsphere_nas_datastore.ds1"]
}
`,
		testAccResourceVSphereVirtualMachineConfigBase(),
		datastore,
		os.Getenv("TF_VAR_VSPHERE_TEMPLATE"),
		os.Getenv("TF_VAR_VSPHERE_USE_LINKED_CLONE"),
		hostname,
	)
}

func testAccResourceVSphereVirtualMachineConfigHighSensitivity() string {
	return fmt.Sprintf(`


%s  // Mix and match config

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = "${vsphere_resource_pool.pool1.id}"
  datastore_id     = vsphere_nas_datastore.ds1.id

  num_cpus            = 2
  cpu_reservation     = 6192
  memory              = 2048
  memory_reservation  = 2048
  latency_sensitivity = "high"
  guest_id            = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network1.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
`,

		testAccResourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccResourceVSphereVirtualMachineTestPathInterpolation() string {
	return fmt.Sprintf(`


%s  // Mix and match config


resource "vsphere_virtual_disk" "d" {
  count      = 2
  size       = 1
  vmdk_path  = "disk-${count.index}.vmdk"
  datastore  = vsphere_nas_datastore.ds1.name
  datacenter = data.vsphere_datacenter.rootdc1.name
}

resource "vsphere_virtual_machine" "vm" {
  name                       = "testacc-test"
  resource_pool_id           = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
  datastore_id               = vsphere_nas_datastore.ds1.id
  guest_id                   = "ubuntu64Guest"
	wait_for_guest_net_timeout = -1

  network_interface { 
    network_id = data.vsphere_network.network1.id
  }
    disk {
      attach       = true
      label        = "disk0"
      path         = vsphere_virtual_disk.d[0].vmdk_path
      unit_number  = 0
      datastore_id = vsphere_nas_datastore.ds1.id
    }

    disk {
      attach       = true
      label        = "disk1"
      path         = vsphere_virtual_disk.d[1].vmdk_path
      unit_number  = 1
      datastore_id = vsphere_nas_datastore.ds1.id
   }
}`,

		testAccResourceVSphereVirtualMachineConfigBase())
}

func testaccresourcevspherevirtualmachineconfigcontentlibraryBasic() string {
	return fmt.Sprintf(`
%s


resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_test"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name = "ubuntu"
  description = "Ubuntu Description"
  library_id = vsphere_content_library.library.id
  file_url = "%s"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "testacc-test"
  resource_pool_id = vsphere_resource_pool.pool1.id
  datastore_id     = data.vsphere_datastore.rootds1.id
  annotation       = "Name: yVM (a very small virtual machine)\nRelease date: 11th November 2015\nFor more information, please visit: cloudarchitectblog.wordpress.com"

  num_cpus = 1
  memory   = 2048

  wait_for_guest_net_timeout = -1
  guest_id                   = "otherLinuxGuest"

  network_interface {
    network_id = data.vsphere_network.network1.id
  }

  clone {
    template_uuid = vsphere_content_library_item.item.id
  }

  cdrom {
    client_device = true
  }

  disk {
    label            = "disk0"
    thin_provisioned = true
    size             = 20
  }
}

`,
		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_CONTENT_LIBRARY_FILES"),
	)
}

func testAccResourceVSphereVirtualMachineDeployOvfFromURL(vmName string) string {
	return fmt.Sprintf(`
%s

variable "ovf_url" {
	default = "%s"
}

data "vsphere_ovf_vm_template" "ovf" {
  name              = "%s"
  resource_pool_id  = vsphere_resource_pool.pool1.id
  datastore_id      = vsphere_nas_datastore.ds1.id
  host_system_id    = data.vsphere_host.roothost1.id
  remote_ovf_url    = var.ovf_url

  ovf_network_map   = {
    "Production_DVS - Mgmt": data.vsphere_network.network1.id
  }
}


resource "vsphere_virtual_machine" "vm" {
  datacenter_id    = data.vsphere_datacenter.rootdc1.id

  annotation       = data.vsphere_ovf_vm_template.ovf.annotation
  name             = data.vsphere_ovf_vm_template.ovf.name
  num_cpus         = data.vsphere_ovf_vm_template.ovf.num_cpus
  memory           = data.vsphere_ovf_vm_template.ovf.memory
  guest_id         = data.vsphere_ovf_vm_template.ovf.guest_id
  resource_pool_id = data.vsphere_ovf_vm_template.ovf.resource_pool_id
  datastore_id     = data.vsphere_ovf_vm_template.ovf.datastore_id
  host_system_id   = data.vsphere_ovf_vm_template.ovf.host_system_id

  dynamic "network_interface" {
    for_each = data.vsphere_ovf_vm_template.ovf.ovf_network_map
    content {
        network_id = network_interface.value
    }
  }

  ovf_deploy {
	  remote_ovf_url  = var.ovf_url
	  ovf_network_map = data.vsphere_ovf_vm_template.ovf.ovf_network_map
  }
}

`,
		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEST_OVF"),
		vmName,
	)
}

func testAccResourceVSphereVirtualMachineDeployOvaFromURL(vmName string) string {
	return fmt.Sprintf(`
%s // Mix and match config

variable "ova_url" {
	default = "%s"
}

data "vsphere_ovf_vm_template" "ovf" {
  name              = "%s"
  resource_pool_id  = vsphere_resource_pool.pool1.id
  datastore_id      = vsphere_nas_datastore.ds1.id
  host_system_id    = data.vsphere_host.roothost1.id
  remote_ovf_url    = var.ova_url

  ovf_network_map   = {
    "Production_DVS - Mgmt": data.vsphere_network.network1.id,
  }
}


resource "vsphere_virtual_machine" "vm" {
  datacenter_id    = data.vsphere_datacenter.rootdc1.id

  annotation            = data.vsphere_ovf_vm_template.ovf.annotation
  name                  = data.vsphere_ovf_vm_template.ovf.name
  num_cpus              = data.vsphere_ovf_vm_template.ovf.num_cpus
  memory                = data.vsphere_ovf_vm_template.ovf.memory
  guest_id              = data.vsphere_ovf_vm_template.ovf.guest_id
  resource_pool_id      = vsphere_resource_pool.pool1.id
  datastore_id          = vsphere_nas_datastore.ds1.id
  host_system_id        = data.vsphere_ovf_vm_template.ovf.host_system_id

  dynamic "network_interface" {
    for_each = data.vsphere_ovf_vm_template.ovf.ovf_network_map
    content {
        network_id = network_interface.value
    }
  }

  ovf_deploy {
	  remote_ovf_url  = var.ova_url
	  ovf_network_map = data.vsphere_ovf_vm_template.ovf.ovf_network_map
  }
}
`,
		testAccResourceVSphereVirtualMachineConfigBase(),
		os.Getenv("TF_VAR_VSPHERE_TEST_OVA"),
		vmName,
	)
}

// Tests to skip until new features are developed.

// Needs storage policy resource

// Needs vsphere_file to support remote sources

// Same as above

// Needs ability to set up SCSI adapter and disks

// Require vApp enabled source

// Must be able to manage datastore cluster membership outside of datastore
