package virtualdevice

import (
	"github.com/vmware/govmomi/vim25/types"
)

type TerraformVirtualDiskFlatVer2BackingInfo struct {
	*types.VirtualDiskFlatVer2BackingInfo
}

type TerraformVirtualDiskRawDiskMappingVer1BackingInfo struct {
	*types.VirtualDiskRawDiskMappingVer1BackingInfo
}

type TerraformVirtualMachineDiskBackingInfo interface {
	types.BaseVirtualDeviceBackingInfo
	GetDiskMode() string
	SetDiskMode(string)
	GetSplit() *bool
	SetSplit(*bool)
	GetWriteThrough() *bool
	SetWriteThrough(*bool)
	GetThinProvisioned() *bool
	SetThinProvisioned(*bool)
	GetEagerlyScrub() *bool
	SetEagerlyScrub(*bool)
	GetUuid() string
	GetContentId() string
	SetContentId(string)
	GetChangeId() string
	SetChangeId(string)
	GetParent() interface{}
	SetParent(interface{})
	GetDeltaDiskFormat() string
	SetDeltaDiskFormat(string)
	GetDigestEnabled() *bool
	SetDigestEnabled(*bool)
	GetDeltaDiskFormatVariant() string
	SetDeltaDiskFormatVariant(string)
	GetDeltaGrainSize() int32
	SetDeltaGrainSize(int32)
	GetSharing() string
	SetSharing(string)
	GetKeyId() *types.CryptoKeyId
	SetKeyId(*types.CryptoKeyId)
	GetLunUuid() string
	SetLunUuid(string)
	GetDeviceName() string
	SetDeviceName(string)
	GetCompatibilityMode() string
	SetCompatibilityMode(string)
	GetDatastore() *types.ManagedObjectReference
	SetDatastore(*types.ManagedObjectReference)
	GetFileName() string
	SetFileName(string)
	GetBackingObjectId() string
	SetBackingObjectId(string)
}

func GetBackingForDisk(disk *types.VirtualDisk) (TerraformVirtualMachineDiskBackingInfo, bool) {
	return GetBacking(disk.Backing)
}

func GetBacking(backingInfo types.BaseVirtualDeviceBackingInfo) (TerraformVirtualMachineDiskBackingInfo, bool) {
	switch backingInfo.(type) {
	case *types.VirtualDiskFlatVer2BackingInfo:
		return TerraformVirtualDiskFlatVer2BackingInfo{backingInfo.(*types.VirtualDiskFlatVer2BackingInfo)}, true
	case *types.VirtualDiskRawDiskMappingVer1BackingInfo:
		return TerraformVirtualDiskRawDiskMappingVer1BackingInfo{backingInfo.(*types.VirtualDiskRawDiskMappingVer1BackingInfo)}, true
	default:
		return nil, false
	}
}

func ToVirtualDiskFlatVer2BackingInfo(backing TerraformVirtualMachineDiskBackingInfo) types.VirtualDiskFlatVer2BackingInfo {
	flatBacking := types.VirtualDiskFlatVer2BackingInfo{}

	flatBacking.FileName = backing.GetFileName()
	flatBacking.Datastore = backing.GetDatastore()
	flatBacking.BackingObjectId = backing.GetBackingObjectId()

	flatBacking.DiskMode = backing.GetDiskMode()
	flatBacking.Split = backing.GetSplit()
	flatBacking.WriteThrough = backing.GetWriteThrough()
	flatBacking.ThinProvisioned = backing.GetThinProvisioned()
	flatBacking.EagerlyScrub = backing.GetEagerlyScrub()
	flatBacking.Uuid = backing.GetUuid()
	flatBacking.ContentId = backing.GetContentId()
	flatBacking.ChangeId = backing.GetChangeId()
	if parentBacking, ok := backing.GetParent().(*types.VirtualDiskFlatVer2BackingInfo); ok {
		flatBacking.Parent = parentBacking
	}
	flatBacking.DeltaDiskFormat = backing.GetDeltaDiskFormat()
	flatBacking.DigestEnabled = backing.GetDigestEnabled()
	flatBacking.DeltaGrainSize = backing.GetDeltaGrainSize()
	flatBacking.DeltaDiskFormatVariant = backing.GetDeltaDiskFormatVariant()
	flatBacking.Sharing = backing.GetSharing()
	flatBacking.KeyId = backing.GetKeyId()

	return flatBacking
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDiskMode() string {
	return backing.DiskMode
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDiskMode(diskMode string) {
	backing.DiskMode = diskMode
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetSplit() *bool {
	return backing.Split
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetSplit(split *bool) {
	backing.Split = split
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetWriteThrough() *bool {
	return backing.WriteThrough
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetWriteThrough(writeThrough *bool) {
	backing.WriteThrough = writeThrough
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetThinProvisioned() *bool {
	return backing.ThinProvisioned
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetThinProvisioned(thinProvisioned *bool) {
	backing.ThinProvisioned = thinProvisioned
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetEagerlyScrub() *bool {
	return backing.EagerlyScrub
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetEagerlyScrub(eagerlyScrub *bool) {
	backing.EagerlyScrub = eagerlyScrub
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetUuid() string {
	return backing.Uuid
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetContentId() string {
	return backing.ContentId
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetContentId(contentId string) {
	backing.ContentId = contentId
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetChangeId() string {
	return backing.ChangeId
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetChangeId(changeId string) {
	backing.ChangeId = changeId

}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetParent() interface{} {
	return backing.Parent
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetParent(i interface{}) {
	backing.Parent = i.(*types.VirtualDiskFlatVer2BackingInfo)
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDeltaDiskFormat() string {
	return backing.DeltaDiskFormat
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDeltaDiskFormat(deltaDiskFormat string) {
	backing.DeltaDiskFormat = deltaDiskFormat
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDigestEnabled() *bool {
	return backing.DigestEnabled
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDigestEnabled(digestEnabled *bool) {
	backing.DigestEnabled = digestEnabled
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDeltaDiskFormatVariant() string {
	return backing.DeltaDiskFormatVariant

}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDeltaDiskFormatVariant(deltaDiskFormatVariant string) {
	backing.DeltaDiskFormatVariant = deltaDiskFormatVariant
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDeltaGrainSize() int32 {
	return backing.DeltaGrainSize
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDeltaGrainSize(size int32) {
	backing.DeltaGrainSize = size
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetSharing() string {
	return backing.Sharing
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetSharing(sharing string) {
	backing.Sharing = sharing
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetKeyId() *types.CryptoKeyId {
	return backing.KeyId
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetKeyId(keyId *types.CryptoKeyId) {
	backing.KeyId = keyId
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetLunUuid() string {
	return ""
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetLunUuid(string) {
	//no op
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDeviceName() string {
	return ""
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDeviceName(string) {
	//no op
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetCompatibilityMode() string {
	return ""
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetCompatibilityMode(string) {
	//no op
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetDatastore() *types.ManagedObjectReference {
	return backing.Datastore
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetDatastore(mor *types.ManagedObjectReference) {
	backing.Datastore = mor
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetFileName() string {
	return backing.FileName
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetFileName(fileName string) {
	backing.FileName = fileName
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) GetBackingObjectId() string {
	return backing.BackingObjectId
}

func (backing TerraformVirtualDiskFlatVer2BackingInfo) SetBackingObjectId(id string) {
	backing.BackingObjectId = id
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDiskMode() string {
	return backing.DiskMode
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDiskMode(diskMode string) {
	backing.DiskMode = diskMode
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetSplit() *bool {
	return nil
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetSplit(split *bool) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetWriteThrough() *bool {
	return nil
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetWriteThrough(writeThrough *bool) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetThinProvisioned() *bool {
	return nil
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetThinProvisioned(thinProvisioned *bool) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetEagerlyScrub() *bool {
	return nil
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetEagerlyScrub(eagerlyScrub *bool) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetUuid() string {
	return backing.LunUuid
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetContentId() string {
	return backing.ContentId
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetContentId(id string) {
	backing.ContentId = id
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetChangeId() string {
	return backing.ChangeId
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetChangeId(id string) {
	backing.ChangeId = id
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetParent() interface{} {
	return backing.Parent
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetParent(i interface{}) {
	backing.Parent = i.(*types.VirtualDiskRawDiskMappingVer1BackingInfo)
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDeltaDiskFormat() string {
	return backing.DeltaDiskFormat
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDeltaDiskFormat(format string) {
	backing.DeltaDiskFormat = format
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDigestEnabled() *bool {
	return nil
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDigestEnabled(*bool) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDeltaDiskFormatVariant() string {
	return ""

}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDeltaDiskFormatVariant(string) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDeltaGrainSize() int32 {
	return backing.DeltaGrainSize
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDeltaGrainSize(size int32) {
	backing.DeltaGrainSize = size
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetSharing() string {
	return backing.Sharing
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetSharing(sharing string) {
	backing.Sharing = sharing
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetKeyId() *types.CryptoKeyId {
	return nil
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetKeyId(*types.CryptoKeyId) {
	//no op
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetLunUuid() string {
	return backing.LunUuid
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetLunUuid(lunUuid string) {
	backing.LunUuid = lunUuid
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDeviceName() string {
	return backing.DeviceName
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDeviceName(deviceName string) {
	backing.DeviceName = deviceName
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetCompatibilityMode() string {
	return backing.CompatibilityMode
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetCompatibilityMode(compatibilityMode string) {
	backing.CompatibilityMode = compatibilityMode
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetDatastore() *types.ManagedObjectReference {
	return backing.Datastore
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetDatastore(mor *types.ManagedObjectReference) {
	backing.Datastore = mor
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetFileName() string {
	return backing.FileName
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetFileName(fileName string) {
	backing.FileName = fileName
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) GetBackingObjectId() string {
	return backing.BackingObjectId
}

func (backing TerraformVirtualDiskRawDiskMappingVer1BackingInfo) SetBackingObjectId(id string) {
	backing.BackingObjectId = id
}
