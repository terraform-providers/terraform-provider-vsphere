---
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vmfs_disks"
sidebar_current: "docs-vsphere-data-source-vmfs-disks"
description: |-
  A data source that can be used to discover storage devices that can be used for VMFS datastores.
---

# vsphere\_vmfs\_disks

The `vsphere_vmfs_disks` data source can be used to discover the storage
devices available on an ESXi host. This data source can be combined with the
[`vsphere_vmfs_datastore`][data-source-vmfs-datastore] resource to create VMFS
datastores based off a set of discovered disks.

[data-source-vmfs-datastore]: /docs/providers/vsphere/r/vmfs_datastore.html

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc1"
}

data "vsphere_host" "host" {
  name          = "esxi1"
  datacenter_id = "${data.vsphere_datacenter.datacenter.id}"
}

data "vsphere_vmfs_disks" "available" {
  host_system_id = "${data.vsphere_host.host.id}"
  rescan         = true
  filter         = "mpx.vmhba1:C0:T[12]:L0"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) The [managed object ID][docs-about-morefs] of
  the host to look for disks on.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

* `rescan` - (Optional) Whether or not to rescan storage adapters before
  searching for disks. This may lengthen the time it takes to perform the
  search. Default: `false`.
* `filter` - (Optional) A regular expression to filter the disks against. Only
  disks with canonical names that match will be included. 

~> **NOTE:** Using a `filter` is recommended if there is any chance the host
will have any specific storage devices added to it that may affect the order of
the output `disks` attribute below, which is lexicographically sorted.

## Attribute Reference

* `disks` - A lexicographically sorted list of devices discovered by the
  operation, matching the supplied `filter`, if provided.
  
* `disks_info` - List of disks discovered by the operation with more details about them.
  * `name`           - Display name of the disk 
  * `path`           - The path of the volume of the disk.
  * `capacity_in_gb` - Capacity of the disk in GB.
