package vsphere

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/clustercomputeresource"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/contentlibrary"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/datacenter"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/datastore"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/dvportgroup"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/network"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/resourcepool"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/spbm"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/storagepod"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/vappcontainer"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/virtualdisk"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/virtualmachine"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// testAccResourceVSphereEmpty provides an empty provider config to pass some
// error tests with an empty state. This is to ensure there's no dangling
// resources on the destroy check if for some reason some state gets written.
const testAccResourceVSphereEmpty = `
provider vsphere{}
`

// testCheckVariables bundles common variables needed by various test checkers.
type testCheckVariables struct {
	// A client for various operations.
	client *govmomi.Client

	// REST client
	restClient *rest.Client

	// The client for tagging operations.
	tagsManager *tags.Manager

	// The subject resource's ID.
	resourceID string

	// The subject resource's attributes.
	resourceAttributes map[string]string

	// The ESXi host that a various API call is directed at.
	esxiHost string

	// The datacenter that a various API call is directed at.
	datacenter string

	// A timeout to pass to various context creation calls.
	timeout time.Duration
}

func testClientVariablesForResource(s *terraform.State, addr string) (testCheckVariables, error) {
	rs, ok := s.RootModule().Resources[addr]
	if !ok {
		return testCheckVariables{}, fmt.Errorf("%s not found in state", addr)
	}

	tm, err := testAccProvider.Meta().(*Client).TagsManager()
	if err != nil {
		return testCheckVariables{}, err
	}
	return testCheckVariables{
		client:             testAccProvider.Meta().(*Client).vimClient,
		restClient:         testAccProvider.Meta().(*Client).restClient,
		tagsManager:        tm,
		resourceID:         rs.Primary.ID,
		resourceAttributes: rs.Primary.Attributes,
		esxiHost:           testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
		datacenter:         testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		timeout:            time.Minute * 5,
	}, nil
}

// testAccESXiFlagSet returns true if TF_VAR_VSPHERE_TEST_ESXI is set.
func testAccESXiFlagSet() bool {
	return os.Getenv("TF_VAR_VSPHERE_TEST_ESXI") != ""
}

// testAccSkipIfNotEsxi skips a test if TF_VAR_VSPHERE_TEST_ESXI is not set.
func testAccSkipIfNotEsxi(t *testing.T) {
	if !testAccESXiFlagSet() {
		t.Skip("set TF_VAR_VSPHERE_TEST_ESXI to run ESXi-specific acceptance tests")
	}
}

// testAccSkipIfEsxi skips a test if TF_VAR_VSPHERE_TEST_ESXI is set.
func testAccSkipIfEsxi(t *testing.T) {
	if testAccESXiFlagSet() {
		t.Skip("test skipped as TF_VAR_VSPHERE_TEST_ESXI is set")
	}
}

// expectErrorIfNotVirtualCenter returns the error message that
// viapi.ValidateVirtualCenter returns if TF_VAR_VSPHERE_TEST_ESXI is set, to allow for test
// cases that will still run on ESXi, but will expect validation failure.
func expectErrorIfNotVirtualCenter() *regexp.Regexp {
	if testAccESXiFlagSet() {
		return regexp.MustCompile(viapi.ErrVirtualCenterOnly)
	}
	return nil
}

// copyStatePtr returns a TestCheckFunc that copies the reference to the test
// run's state to t. This allows access to the state data in later steps where
// it's not normally accessible (ie: in pre-config parts in another test step).
func copyStatePtr(t **terraform.State) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*t = s
		return nil
	}
}

// copyState returns a TestCheckFunc that returns a deep copy of the state.
// Unlike copyStatePtr, this state has de-coupled from the in-flight state, so
// it will not be modified on subsequent steps and hence will possibly drift.
// It can be used to access values of the state at a certain step.
func copyState(t **terraform.State) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*t = s.DeepCopy()
		return nil
	}
}

// testGetPortGroup is a convenience method to fetch a static port group
// resource for testing.
func testGetPortGroup(s *terraform.State, resourceName string) (*types.HostPortGroup, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_host_port_group.%s", resourceName))
	if err != nil {
		return nil, err
	}

	hsID, name, err := splitHostPortGroupID(tVars.resourceID)
	if err != nil {
		return nil, err
	}
	ns, err := hostNetworkSystemFromHostSystemID(tVars.client, hsID)
	if err != nil {
		return nil, fmt.Errorf("error loading host network system: %s", err)
	}

	return hostPortGroupFromName(tVars.client, ns, name)
}

// testGetVirtualMachine is a convenience method to fetch a virtual machine by
// resource name.
func testGetVirtualMachine(s *terraform.State, resourceName string) (*object.VirtualMachine, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_virtual_machine.%s", resourceName))
	if err != nil {
		return nil, err
	}
	uuid, ok := tVars.resourceAttributes["uuid"]
	if !ok {
		return nil, fmt.Errorf("resource %q has no UUID", resourceName)
	}
	return virtualmachine.FromUUID(tVars.client, uuid)
}

// testGetVirtualMachineProperties is a convenience method that adds an extra
// step to testGetVirtualMachine to get the properties of a virtual machine.
func testGetVirtualMachineProperties(s *terraform.State, resourceName string) (*mo.VirtualMachine, error) {
	vm, err := testGetVirtualMachine(s, resourceName)
	if err != nil {
		return nil, err
	}
	return virtualmachine.Properties(vm)
}

// testGetVirtualMachineHost returns the HostSystem for the host that this
// virtual machine is currently on.
func testGetVirtualMachineHost(s *terraform.State, resourceName string) (*object.HostSystem, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_virtual_machine.%s", resourceName))
	if err != nil {
		return nil, err
	}
	vprops, err := testGetVirtualMachineProperties(s, resourceName)
	if err != nil {
		return nil, err
	}
	return hostsystem.FromID(tVars.client, vprops.Runtime.Host.Value)
}

// testGetVirtualMachineResourcePool returns the ResourcePool object for the
// resource pool this VM is currently in.
func testGetVirtualMachineResourcePool(s *terraform.State, resourceName string) (*object.ResourcePool, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_virtual_machine.%s", resourceName))
	if err != nil {
		return nil, err
	}
	vprops, err := testGetVirtualMachineProperties(s, resourceName)
	if err != nil {
		return nil, err
	}
	return resourcepool.FromID(tVars.client, vprops.ResourcePool.Value)
}

// testGetVirtualMachineSCSIBusType reads the SCSI bus type for the supplied
// virtual machine.
func testGetVirtualMachineSCSIBusType(s *terraform.State, resourceName string) (string, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_virtual_machine.%s", resourceName))
	if err != nil {
		return "", err
	}
	vprops, err := testGetVirtualMachineProperties(s, resourceName)
	if err != nil {
		return "", err
	}
	count, err := strconv.Atoi(tVars.resourceAttributes["scsi_controller_count"])
	if err != nil {
		return "", err
	}
	l := object.VirtualDeviceList(vprops.Config.Hardware.Device)
	return virtualdevice.ReadSCSIBusType(l, count), nil
}

// testGetVirtualMachineSCSIBusSharing reads the SCSI bus sharing mode for the
// supplied virtual machine.
func testGetVirtualMachineSCSIBusSharing(s *terraform.State, resourceName string) (string, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_virtual_machine.%s", resourceName))
	if err != nil {
		return "", err
	}
	vprops, err := testGetVirtualMachineProperties(s, resourceName)
	if err != nil {
		return "", err
	}
	count, err := strconv.Atoi(tVars.resourceAttributes["scsi_controller_count"])
	if err != nil {
		return "", err
	}
	l := object.VirtualDeviceList(vprops.Config.Hardware.Device)
	return virtualdevice.ReadSCSIBusSharing(l, count), nil
}

func testGetDatacenter(s *terraform.State, resourceName string) (*object.Datacenter, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_datacenter.%s", resourceName))
	if err != nil {
		return nil, err
	}
	dcName, ok := tVars.resourceAttributes["name"]
	if !ok {
		return nil, fmt.Errorf("Datacenter resource %q has no name", resourceName)
	}
	return getDatacenter(tVars.client, dcName)
}

func testGetResourcePool(s *terraform.State, resourceName string) (*object.ResourcePool, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereResourcePoolName, resourceName))
	if err != nil {
		return nil, err
	}
	return resourcepool.FromID(vars.client, vars.resourceID)
}

func testGetResourcePoolProperties(s *terraform.State, resourceName string) (*mo.ResourcePool, error) {
	rp, err := testGetResourcePool(s, resourceName)
	if err != nil {
		return nil, err
	}
	return resourcepool.Properties(rp)
}

func testGetDatacenterCustomAttributes(s *terraform.State, resourceName string) (*mo.Datacenter, error) {
	dc, err := testGetDatacenter(s, resourceName)
	if err != nil {
		return nil, err
	}
	return datacenterCustomAttributes(dc)
}

func testGetVAppEntity(s *terraform.State, resourceName string) (*types.VAppEntityConfigInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereVAppEntityName, resourceName))
	if err != nil {
		return nil, err
	}
	return resourceVSphereVAppEntityFind(vars.client, vars.resourceID)
}

func testGetVAppContainer(s *terraform.State, resourceName string) (*object.VirtualApp, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereVAppContainerName, resourceName))
	if err != nil {
		return nil, err
	}
	return vappcontainer.FromID(vars.client, vars.resourceID)
}

func testGetVAppContainerProperties(s *terraform.State, resourceName string) (*mo.VirtualApp, error) {
	vc, err := testGetVAppContainer(s, resourceName)
	if err != nil {
		return nil, err
	}
	return vappcontainer.Properties(vc)
}

func testGetContentLibrary(s *terraform.State, resourceName string) (*library.Library, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_content_library.%s", resourceName))
	if err != nil {
		return nil, err
	}
	return contentlibrary.FromID(tVars.restClient, tVars.resourceID)
}

func testGetContentLibraryItem(s *terraform.State, resourceName string) (*library.Item, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_content_library_item.%s", resourceName))
	if err != nil {
		return nil, err
	}
	return contentlibrary.ItemFromID(tVars.restClient, tVars.resourceID)
}

// testPowerOffVM does an immediate power-off of the supplied virtual machine
// resource defined by the supplied resource address name. It is used to help
// set up a test scenarios where a VM is powered off.
func testPowerOffVM(s *terraform.State, resourceName string) error {
	vm, err := testGetVirtualMachine(s, resourceName)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return fmt.Errorf("error powering off VM: %s", err)
	}
	tctx, tcancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer tcancel()
	if err := task.Wait(tctx); err != nil {
		return fmt.Errorf("error waiting for poweroff: %s", err)
	}
	return nil
}

// testRenameVMFirstDisk renames the first disk in a virtual machine
// configuration and re-attaches it to the virtual machine under the new name.
func testRenameVMFirstDisk(s *terraform.State, resourceName string, new string) error {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_virtual_machine.%s", resourceName))
	if err != nil {
		return err
	}
	vm, err := testGetVirtualMachine(s, resourceName)
	if err != nil {
		return err
	}
	vprops, err := testGetVirtualMachineProperties(s, resourceName)
	if err != nil {
		return err
	}
	if err := testPowerOffVM(s, resourceName); err != nil {
		return err
	}
	dcp, err := folder.RootPathParticleVM.SplitDatacenter(vm.InventoryPath)
	if err != nil {
		return err
	}
	dc, err := getDatacenter(tVars.client, dcp)
	if err != nil {
		return err
	}

	var dcSpec []types.BaseVirtualDeviceConfigSpec
	for _, d := range vprops.Config.Hardware.Device {
		if oldDisk, ok := d.(*types.VirtualDisk); ok {
			newFileName, err := virtualdisk.Move(
				tVars.client,
				oldDisk.Backing.(*types.VirtualDiskFlatVer2BackingInfo).FileName,
				dc,
				new,
				nil,
			)
			if err != nil {
				return err
			}
			newDisk := &types.VirtualDisk{
				VirtualDevice: types.VirtualDevice{
					Backing: &types.VirtualDiskFlatVer2BackingInfo{
						VirtualDeviceFileBackingInfo: types.VirtualDeviceFileBackingInfo{
							FileName: newFileName,
						},
						ThinProvisioned: oldDisk.Backing.(*types.VirtualDiskFlatVer2BackingInfo).ThinProvisioned,
						EagerlyScrub:    oldDisk.Backing.(*types.VirtualDiskFlatVer2BackingInfo).EagerlyScrub,
						DiskMode:        oldDisk.Backing.(*types.VirtualDiskFlatVer2BackingInfo).DiskMode,
					},
				},
			}
			newDisk.ControllerKey = oldDisk.ControllerKey
			newDisk.UnitNumber = oldDisk.UnitNumber

			dspec, err := object.VirtualDeviceList{oldDisk}.ConfigSpec(types.VirtualDeviceConfigSpecOperationRemove)
			if err != nil {
				return err
			}
			dspec[0].GetVirtualDeviceConfigSpec().FileOperation = ""
			aspec, err := object.VirtualDeviceList{newDisk}.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
			if err != nil {
				return err
			}
			aspec[0].GetVirtualDeviceConfigSpec().FileOperation = ""
			dcSpec = append(dcSpec, dspec...)
			dcSpec = append(dcSpec, aspec...)
			break
		}
	}
	if len(dcSpec) < 1 {
		return fmt.Errorf("could not find a virtual disk on virtual machine %q", vm.InventoryPath)
	}
	spec := types.VirtualMachineConfigSpec{
		DeviceChange: dcSpec,
	}
	return virtualmachine.Reconfigure(vm, spec)
}

// testDeleteVMDisk deletes a VMDK file from the virtual machine directory. It
// doesn't check configuration other than to look for the directory the VMX
// file is in and is mainly meant to serve as a cleanup method.
func testDeleteVMDisk(s *terraform.State, name string) error {
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
	p := &object.DatastorePath{
		Datastore: vmxPath.Datastore,
		Path:      path.Join(path.Dir(vmxPath.Path), name),
	}
	return virtualdisk.Delete(tVars.client, p.String(), dc)
}

// testDeleteVM deletes the virtual machine. This is used to test resource
// re-creation if TF cannot locate a VM that is in state any more.
func testDeleteVM(s *terraform.State, resourceName string) error {
	if err := testPowerOffVM(s, resourceName); err != nil {
		return err
	}
	vm, err := testGetVirtualMachine(s, resourceName)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	task, err := vm.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("error destroying virtual machine: %s", err)
	}
	tctx, tcancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer tcancel()
	return task.Wait(tctx)
}

// testGetTagCategory gets a tag category by name.
func testGetTagCategory(s *terraform.State, resourceName string) (*tags.Category, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_tag_category.%s", resourceName))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	category, err := tVars.tagsManager.GetCategory(ctx, tVars.resourceID)
	if err != nil {
		return nil, fmt.Errorf("could not get tag category for ID %q: %s", tVars.resourceID, err)
	}

	return category, nil
}

// testGetTag gets a tag by name.
func testGetTag(s *terraform.State, resourceName string) (*tags.Tag, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_tag.%s", resourceName))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	tag, err := tVars.tagsManager.GetTag(ctx, tVars.resourceID)
	if err != nil {
		return nil, fmt.Errorf("could not get tag for ID %q: %s", tVars.resourceID, err)
	}

	return tag, nil
}

// testObjectHasTags checks an object to see if it has the tags that currently
// exist in the Terrafrom state under the resource with the supplied name.
func testObjectHasTags(s *terraform.State, tm *tags.Manager, obj object.Reference, tagResName string) error {
	var expectedIDs []string
	if tagRS, ok := s.RootModule().Resources[fmt.Sprintf("vsphere_tag.%s", tagResName)]; ok {
		expectedIDs = append(expectedIDs, tagRS.Primary.ID)
	} else {
		var n int
		for {
			multiTagRS, ok := s.RootModule().Resources[fmt.Sprintf("vsphere_tag.%s.%d", tagResName, n)]
			if !ok {
				break
			}
			expectedIDs = append(expectedIDs, multiTagRS.Primary.ID)
			n++
		}
	}
	if len(expectedIDs) < 1 {
		return fmt.Errorf("could not find state for vsphere_tag.%s or vsphere_tag.%s.*", tagResName, tagResName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	actualIDs, err := tm.ListAttachedTags(ctx, obj)
	if err != nil {
		return err
	}

	for _, expectedID := range expectedIDs {
		var found bool
		for _, actualID := range actualIDs {
			if expectedID == actualID {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("could not find expected tag ID %q attached to object %q", expectedID, obj.Reference().Value)
		}
	}

	return nil
}

// testObjectHasNoTags checks to make sure that an object has no tags attached
// to it. The parameters are the same as testObjectHasTags, but no tag resource
// needs to be supplied.
func testObjectHasNoTags(tm *tags.Manager, obj object.Reference) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	actualIDs, err := tm.ListAttachedTags(ctx, obj)
	if err != nil {
		return err
	}
	if len(actualIDs) > 0 {
		return fmt.Errorf("object %q still has tags (%#v)", obj.Reference().Value, actualIDs)
	}
	return nil
}

// testGetDatastore gets the datastore at the supplied full address. This
// function works for multiple datastore resources (example:
// vsphere_nas_datastore and vsphere_vmfs_datastore), hence the need for the
// full resource address including the resource type.
func testGetDatastore(s *terraform.State, resAddr string) (*object.Datastore, error) {
	vars, err := testClientVariablesForResource(s, resAddr)
	if err != nil {
		return nil, err
	}
	return datastore.FromID(vars.client, vars.resourceID)
}

// testGetDatastoreProperties is a convenience method that adds an extra step
// to testGetDatastore to get the properties of a datastore.
func testGetDatastoreProperties(s *terraform.State, datastoreType string, resourceName string) (*mo.Datastore, error) {
	ds, err := testGetDatastore(s, "vsphere_"+datastoreType+"_datastore."+resourceName)
	if err != nil {
		return nil, err
	}
	return datastore.Properties(ds)
}

// testAccResourceVSphereDatastoreCheckTags is a check to ensure that the
// supplied datastore has had the tags that have been created with the supplied
// tag resource name attached.
//
// The full datastore resource address is needed as this functions across
// multiple datastore resource types.
func testAccResourceVSphereDatastoreCheckTags(dsResAddr, tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, err := testGetDatastore(s, dsResAddr)
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, ds, tagResName)
	}
}

// testGetFolder is a convenience method to fetch a folder by resource name.
func testGetFolder(s *terraform.State, resourceName string) (*object.Folder, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_folder.%s", resourceName))
	if err != nil {
		return nil, err
	}
	return folder.FromID(tVars.client, tVars.resourceID)
}

// testGetFolderProperties is a convenience method that adds an extra step to
// testGetFolder to get the properties of a folder.
func testGetFolderProperties(s *terraform.State, resourceName string) (*mo.Folder, error) {
	f, err := testGetFolder(s, resourceName)
	if err != nil {
		return nil, err
	}
	return folder.Properties(f)
}

// testGetDVS is a convenience method to fetch a DVS by resource name.
func testGetDVS(s *terraform.State, resourceName string) (*object.VmwareDistributedVirtualSwitch, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_distributed_virtual_switch.%s", resourceName))
	if err != nil {
		return nil, err
	}

	return dvsFromUUID(tVars.client, tVars.resourceID)
}

// testGetDVSProperties is a convenience method that adds an extra step to
// testGetDVS to get the properties of a DVS.
func testGetDVSProperties(s *terraform.State, resourceName string) (*mo.VmwareDistributedVirtualSwitch, error) {
	dvs, err := testGetDVS(s, resourceName)
	if err != nil {
		return nil, err
	}
	return dvsProperties(dvs)
}

// testGetDVPortgroup is a convenience method to fetch a DV portgroup by resource name.
func testGetDVPortgroup(s *terraform.State, resourceName string) (*object.DistributedVirtualPortgroup, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_distributed_port_group.%s", resourceName))
	if err != nil {
		return nil, err
	}
	dvsID := tVars.resourceAttributes["distributed_virtual_switch_uuid"]
	return dvportgroup.FromKey(tVars.client, dvsID, tVars.resourceID)
}

// testGetDVPortgroupProperties is a convenience method that adds an extra step to
// testGetDVPortgroup to get the properties of a DV portgroup.
func testGetDVPortgroupProperties(s *terraform.State, resourceName string) (*mo.DistributedVirtualPortgroup, error) {
	dvs, err := testGetDVPortgroup(s, resourceName)
	if err != nil {
		return nil, err
	}
	return dvportgroup.Properties(dvs)
}

// testCheckResourceNotAttr is an inverse check of TestCheckResourceAttr. It
// checks to make sure the resource attribute does *not* match a certain value.
func testCheckResourceNotAttr(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := resource.TestCheckResourceAttr(name, key, value)(s)
		if err != nil {
			if regexp.MustCompile("[-_.a-zA-Z0-9]\\: Attribute '.*' expected .*, got .*").MatchString(err.Error()) {
				return nil
			}
			return err
		}
		return fmt.Errorf("%s: Attribute '%s' expected to not match %#v", name, key, value)
	}
}

// testGetCustomAttribute gets a custom attribute by name.
func testGetCustomAttribute(s *terraform.State, resourceName string) (*types.CustomFieldDef, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_custom_attribute.%s", resourceName))
	if err != nil {
		return nil, err
	}

	key, err := strconv.ParseInt(tVars.resourceID, 10, 32)
	if err != nil {
		return nil, err
	}
	fm, err := object.GetCustomFieldsManager(tVars.client.Client)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	fields, err := fm.Field(ctx)
	if err != nil {
		return nil, err
	}
	field := fields.ByKey(int32(key))

	return field, nil
}

func testResourceHasCustomAttributeValues(s *terraform.State, resourceType string, resourceName string, entity *mo.ManagedEntity) error {
	testVars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceType, resourceName))
	if err != nil {
		return err
	}
	expectedAttrs := make(map[string]string)
	re := regexp.MustCompile(`custom_attributes\.(\d+)`)
	for key, value := range testVars.resourceAttributes {
		if m := re.FindStringSubmatch(key); m != nil {
			expectedAttrs[m[1]] = value
		}
	}

	actualAttrs := make(map[string]string)
	for _, fv := range entity.CustomValue {
		value := fv.(*types.CustomFieldStringValue).Value
		if value != "" {
			actualAttrs[fmt.Sprint(fv.GetCustomFieldValue().Key)] = value
		}
	}

	if !reflect.DeepEqual(expectedAttrs, actualAttrs) {
		return fmt.Errorf("expected custom attributes to be %q, got %q", expectedAttrs, actualAttrs)
	}
	return nil
}

// testDeleteDatastoreFile deletes the specified file from a datastore. If the
// file does not exist, an error is returned.
func testDeleteDatastoreFile(client *govmomi.Client, dsID string, path string) error {
	ds, err := datastore.FromID(client, dsID)
	if err != nil {
		return err
	}
	var dc *object.Datacenter
	if ds.DatacenterPath != "" {
		dc, err = getDatacenter(client, ds.DatacenterPath)
		if err != nil {
			return err
		}
	} else {
		dc, err = datacenter.FromInventoryPath(client, ds.InventoryPath)
		if err != nil {
			return err
		}
	}
	fm := object.NewFileManager(client.Client)

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	task, err := fm.DeleteDatastoreFile(ctx, path, dc)
	if err != nil {
		return err
	}
	return task.Wait(context.TODO())
}

// testGetDatastoreCluster is a convenience method to fetch a datastore cluster by
// resource name.
func testGetDatastoreCluster(s *terraform.State, resourceName string) (*object.StoragePod, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereDatastoreClusterName, resourceName))
	if err != nil {
		return nil, err
	}
	return storagepod.FromID(vars.client, vars.resourceID)
}

// testGetDatastoreClusterProperties is a convenience method that adds an extra
// step to testGetDatastoreCluster to get the properties of a StoragePod.
func testGetDatastoreClusterProperties(s *terraform.State, resourceName string) (*mo.StoragePod, error) {
	pod, err := testGetDatastoreCluster(s, resourceName)
	if err != nil {
		return nil, err
	}
	return storagepod.Properties(pod)
}

// testGetDatastoreClusterSDRSVMConfig is a convenience method to fetch a VM's
// SDRS override in a datastore cluster.
func testGetDatastoreClusterSDRSVMConfig(s *terraform.State, resourceName string) (*types.StorageDrsVmConfigInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereStorageDrsVMOverrideName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	podID, vmID, err := resourceVSphereStorageDrsVMOverrideParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	pod, err := storagepod.FromID(vars.client, podID)
	if err != nil {
		return nil, err
	}

	vm, err := virtualmachine.FromUUID(vars.client, vmID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereStorageDrsVMOverrideFindEntry(pod, vm)
}

// testGetComputeCluster is a convenience method to fetch a compute cluster by
// resource name.
func testGetComputeCluster(s *terraform.State, resourceName string, resourceType string) (*object.ClusterComputeResource, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceType, resourceName))
	if err != nil {
		return nil, err
	}
	return clustercomputeresource.FromID(vars.client, vars.resourceID)
}

// testGetComputeClusterFromDataSource is a convenience method to fetch a
// compute cluster via the data in a vsphere_compute_cluster data source.
func testGetComputeClusterFromDataSource(s *terraform.State, resourceName string) (*object.ClusterComputeResource, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("data.%s.%s", resourceVSphereComputeClusterName, resourceName))
	if err != nil {
		return nil, err
	}
	return clustercomputeresource.FromID(vars.client, vars.resourceID)
}

// testGetComputeClusterProperties is a convenience method that adds an extra
// step to testGetComputeCluster to get the properties of a
// ClusterComputeResource.
func testGetComputeClusterProperties(s *terraform.State, resourceName string) (*mo.ClusterComputeResource, error) {
	cluster, err := testGetComputeCluster(s, resourceName, resourceVSphereComputeClusterName)
	if err != nil {
		return nil, err
	}
	return clustercomputeresource.Properties(cluster)
}

// testGetComputeClusterDRSVMConfig is a convenience method to fetch a VM's DRS
// override in a (compute) cluster.
func testGetComputeClusterDRSVMConfig(s *terraform.State, resourceName string) (*types.ClusterDrsVmConfigInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereDRSVMOverrideName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, vmID, err := resourceVSphereDRSVMOverrideParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	vm, err := virtualmachine.FromUUID(vars.client, vmID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereDRSVMOverrideFindEntry(cluster, vm)
}

// testGetComputeClusterHaVMConfig is a convenience method to fetch a VM's HA
// override in a (compute) cluster.
func testGetComputeClusterHaVMConfig(s *terraform.State, resourceName string) (*types.ClusterDasVmConfigInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereHAVMOverrideName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, vmID, err := resourceVSphereHAVMOverrideParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	vm, err := virtualmachine.FromUUID(vars.client, vmID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereHAVMOverrideFindEntry(cluster, vm)
}

// testGetComputeClusterDPMHostConfig is a convenience method to fetch a host's
// DPM override in a (compute) cluster.
func testGetComputeClusterDPMHostConfig(s *terraform.State, resourceName string) (*types.ClusterDpmHostConfigInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereDPMHostOverrideName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, hostID, err := resourceVSphereDPMHostOverrideParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	host, err := hostsystem.FromID(vars.client, hostID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereDPMHostOverrideFindEntry(cluster, host)
}

// testGetHostFromDataSource is a convenience method to fetch a host via the
// data in a vsphere_host data source.
func testGetHostFromDataSource(s *terraform.State, resourceName string) (*object.HostSystem, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("data.vsphere_host.%s", resourceName))
	if err != nil {
		return nil, err
	}
	return hostsystem.FromID(vars.client, vars.resourceID)
}

// testGetComputeClusterVMGroup is a convenience method to fetch a virtual
// machine group in a (compute) cluster.
func testGetComputeClusterVMGroup(s *terraform.State, resourceName string) (*types.ClusterVmGroup, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereComputeClusterVMGroupName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, name, err := resourceVSphereComputeClusterVMGroupParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereComputeClusterVMGroupFindEntry(cluster, name)
}

// testGetComputeClusterHostGroup is a convenience method to fetch a host group
// in a (compute) cluster.
func testGetComputeClusterHostGroup(s *terraform.State, resourceName string) (*types.ClusterHostGroup, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereComputeClusterHostGroupName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, name, err := resourceVSphereComputeClusterHostGroupParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereComputeClusterHostGroupFindEntry(cluster, name)
}

// testGetComputeClusterVMHostRule is a convenience method to fetch a VM/host
// rule from a (compute) cluster.
func testGetComputeClusterVMHostRule(s *terraform.State, resourceName string) (*types.ClusterVmHostRuleInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereComputeClusterVMHostRuleName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, name, err := resourceVSphereComputeClusterVMHostRuleParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereComputeClusterVMHostRuleFindEntry(cluster, name)
}

// testGetComputeClusterVMDependencyRule is a convenience method to fetch a VM
// dependency rule from a (compute) cluster.
func testGetComputeClusterVMDependencyRule(s *terraform.State, resourceName string) (*types.ClusterDependencyRuleInfo, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereComputeClusterVMDependencyRuleName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, name, err := resourceVSphereComputeClusterVMDependencyRuleParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereComputeClusterVMDependencyRuleFindEntry(cluster, name)
}

// testGetComputeClusterVMAffinityRule is a convenience method to fetch a VM
// affinity rule from a (compute) cluster.
func testGetComputeClusterVMAffinityRule(s *terraform.State, resourceName string) (*types.ClusterAffinityRuleSpec, error) {
	vars, err := testClientVariablesForResource(s, fmt.Sprintf("%s.%s", resourceVSphereComputeClusterVMAffinityRuleName, resourceName))
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, name, err := resourceVSphereComputeClusterVMAffinityRuleParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereComputeClusterVMAffinityRuleFindEntry(cluster, name)
}

// testGetComputeClusterVMAntiAffinityRule is a convenience method to fetch a
// VM anti-affinity rule from a (compute) cluster.
func testGetComputeClusterVMAntiAffinityRule(s *terraform.State, resourceName string) (*types.ClusterAntiAffinityRuleSpec, error) {
	vars, err := testClientVariablesForResource(
		s,
		fmt.Sprintf("%s.%s", resourceVSphereComputeClusterVMAntiAffinityRuleName, resourceName),
	)
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	clusterID, name, err := resourceVSphereComputeClusterVMAntiAffinityRuleParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	cluster, err := clustercomputeresource.FromID(vars.client, clusterID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereComputeClusterVMAntiAffinityRuleFindEntry(cluster, name)
}

// testGetDatastoreClusterVMAntiAffinityRule is a convenience method to fetch a
// VM anti-affinity rule from a datastore cluster.
func testGetDatastoreClusterVMAntiAffinityRule(s *terraform.State, resourceName string) (*types.ClusterAntiAffinityRuleSpec, error) {
	vars, err := testClientVariablesForResource(
		s,
		fmt.Sprintf("%s.%s", resourceVSphereDatastoreClusterVMAntiAffinityRuleName, resourceName),
	)
	if err != nil {
		return nil, err
	}

	if vars.resourceID == "" {
		return nil, errors.New("resource ID is empty")
	}

	podID, key, err := resourceVSphereDatastoreClusterVMAntiAffinityRuleParseID(vars.resourceID)
	if err != nil {
		return nil, err
	}

	pod, err := storagepod.FromID(vars.client, podID)
	if err != nil {
		return nil, err
	}

	return resourceVSphereDatastoreClusterVMAntiAffinityRuleFindEntry(pod, key)
}

func testGetVMStoragePolicy(s *terraform.State, resourceName string) (string, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_vm_storage_policy.%s", resourceName))
	if err != nil {
		return "", err
	}
	policyID, ok := tVars.resourceAttributes["id"]
	if !ok {
		return "", fmt.Errorf("resource %q has no id", resourceName)
	}

	return spbm.PolicyNameByID(tVars.client, policyID)
}

func RunSweepers() {
	_ = tagSweep("")
	_ = dcSweep("")
	_ = vmSweep("")
	_ = rpSweep("")
	_ = dsSweep("")
	_ = netSweep("")
	_ = folderSweep("")
}

func tagSweep(string) error {
	ctx := context.TODO()
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	tm, err := client.TagsManager()
	if err != nil {
		return err
	}
	cats, err := tm.GetCategories(ctx)
	if err != nil {
		return err
	}
	for _, cat := range cats {
		if regexp.MustCompile("testacc").Match([]byte(cat.Name)) {
			_ = tm.DeleteCategory(ctx, &cat)
		}
	}
	return nil
}

func dcSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	dcs, err := listDatacenters(client.vimClient)
	if err != nil {
		return err
	}
	for _, dc := range dcs {
		if regexp.MustCompile("testacc").Match([]byte(dc.Name())) {
			_, err := dc.Destroy(context.TODO())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func vmSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	vms, err := virtualmachine.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, vm := range vms {
		if regexp.MustCompile("testacc").Match([]byte(vm.Name())) {
			_ = virtualmachine.PowerOff(vm)
			_ = virtualmachine.Destroy(vm)
		}
	}
	return nil
}

func rpSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	rps, err := resourcepool.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, rp := range rps {
		if regexp.MustCompile("testacc").Match([]byte(rp.Name())) {
			return resourcepool.Delete(rp)
		}
	}
	return nil
}

func dsSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	dss, err := datastore.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, ds := range dss {
		if regexp.MustCompile("testacc").Match([]byte(ds.Name())) {
			return datastore.Unmount(client.vimClient, ds)
		}
	}
	return nil
}

func dspSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	dsps, err := storagepod.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, dsp := range dsps {
		if regexp.MustCompile("testacc").Match([]byte(dsp.Name())) {
			return storagepod.Delete(dsp)
		}
	}
	return nil
}

func ccSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	dsps, err := clustercomputeresource.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, dsp := range dsps {
		if regexp.MustCompile("testacc").Match([]byte(dsp.Name())) {
			return clustercomputeresource.Delete(dsp)
		}
	}
	return nil
}

func netSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	nets, err := network.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, net := range nets {
		if regexp.MustCompile("testacc").Match([]byte(net.Name())) {
			_, err = net.Destroy(context.TODO())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func folderSweep(string) error {
	client, err := sweepVSphereClient()
	if err != nil {
		return err
	}
	folders, err := folder.List(client.vimClient)
	if err != nil {
		return err
	}
	for _, f := range folders {
		if regexp.MustCompile("testacc").Match([]byte(f.Name())) {
			_, err = f.Destroy(context.TODO())
			return err
		}
	}
	return nil
}

func testGetVsphereEntityPermission(s *terraform.State, resourceName string) (string, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_entity_permissions.%s", resourceName))
	if err != nil {
		return "", err
	}
	id, ok := tVars.resourceAttributes["id"]
	if !ok {
		return "", fmt.Errorf("resource %q has no id", resourceName)
	}
	entityType, ok := tVars.resourceAttributes["entity_type"]
	if !ok {
		return "", fmt.Errorf("resource %q has no entity_type", resourceName)
	}
	entityMor := types.ManagedObjectReference{
		Type:  entityType,
		Value: id,
	}
	authorizationManager := object.NewAuthorizationManager(tVars.client.Client)
	permissionsArr, err := authorizationManager.RetrieveEntityPermissions(context.Background(), entityMor, false)
	if err != nil {
		return "", err
	}
	if len(permissionsArr) == 0 {
		return "", fmt.Errorf("permissions not found for entity id %s", id)
	}
	return strconv.Itoa(len(permissionsArr)), nil
}

func testGetVsphereRole(s *terraform.State, resourceName string) (string, error) {
	tVars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_role.%s", resourceName))
	if err != nil {
		return "", err
	}
	id, ok := tVars.resourceAttributes["id"]
	if !ok {
		return "", fmt.Errorf("resource %q has no id", resourceName)
	}

	roleIDInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return "", fmt.Errorf("error while coverting role id %s from string to int %s", id, err)
	}
	roleID := int32(roleIDInt)
	authorizationManager := object.NewAuthorizationManager(tVars.client.Client)
	roleList, err := authorizationManager.RoleList(context.Background())

	if err != nil {
		return "", fmt.Errorf("error while reading the role list %s", err)
	}
	role := roleList.ById(roleID)
	if role == nil {
		return "", fmt.Errorf("role not found")
	}
	return role.Name, nil
}
