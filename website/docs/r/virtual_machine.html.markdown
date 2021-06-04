---
subcategory: "Virtual Machine"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_virtual_machine"
sidebar_current: "docs-vsphere-resource-vm-virtual-machine-resource"
description: |-
  Provides a VMware vSphere virtual machine resource. This can be used to create, modify, and delete virtual machines.
---

# vsphere\_virtual\_machine

The `vsphere_virtual_machine` resource can be used to manage the complex
lifecycle of a virtual machine. It supports management of disk, network
interface, and CDROM devices, creation from scratch or cloning from template,
and migration through both host and storage vMotion.

For more details on working with virtual machines in vSphere, see [this
page][vmware-docs-vm-management].

[vmware-docs-vm-management]: https://docs.vmware.com/en/VMware-vSphere/6.5/com.vmware.vsphere.vm_admin.doc/GUID-55238059-912E-411F-A0E9-A7A536972A91.html

## About Working with Virtual Machines in Terraform

A high degree of control and flexibility is afforded to a vSphere user when it
comes to how to configure, deploy, and manage virtual machines - much more
control than given in a traditional cloud provider. As such, Terraform has to
make some decisions on how to manage the virtual machines it creates and
manages. This section documents things you need to know about your virtual
machine configuration that you should consider when setting up virtual
machines, creating templates to clone from, or migrating from previous versions
of this resource.

### Disks

The `vsphere_virtual_machine` resource currently only supports standard
VMDK-backed virtual disks - it does not support other special kinds of disk
devices like RDM disks.

Disks are managed by an arbitrary label supplied to the [`label`](#label)
attribute of a [`disk` block](#disk-options). This is separate from the
automatic naming that vSphere picks for you when creating a virtual machine.
Control over a virtual disk's name is not supported unless you are attaching an
external disk with the [`attach`](#attach) attribute.

Virtual disks can be SCSI, SATA or IDE. The storage controllers managed by Terraform
can vary, depending on the value supplied to
[`scsi_controller_count`](#scsi_controller_count), 
[`sata_controller_count`](#sata_controller_count), or 
[`ide_controller_count`](#ide_controller_count). This also dictates the
controllers that are checked when looking for disks during a cloning process.
SCSI controllers are all configured with the controller type defined by the 
[`scsi_type`](#scsi_type) setting. If you are cloning from a template, devices
will be added or re-configured as necessary.

When cloning from a template, you must specify disks of either the same or
greater size than the disks in the source template when creating a traditional
clone, or exactly the same size when cloning from snapshot (also known as a
linked clone). For more details, see the section on [creating a virtual machine
from a template](#creating-a-virtual-machine-from-a-template).

### Customization and network waiters

Terraform waits during various parts of a virtual machine deployment to ensure
that it is in a correct expected state before proceeding. These happen when a
VM is created, or also when it's updated, depending on the waiter.

Two waiters of note are:

* **The customization waiter:** This waiter watches events in vSphere to
  monitor when customization on a virtual machine completes during VM creation.
  Depending on your vSphere or VM configuration it may be necessary to change
  the timeout or turn it off. This can be controlled by the
  [`timeout`](#timeout-1) setting in the [customization
  settings](#virtual-machine-customization) block.
* **The network waiter:** This waiter waits for interfaces to show up on a
  guest virtual machine close to the end of both VM creation and update. This
  waiter is necessary to ensure that correct IP information gets reported to
  the guest virtual machine, mainly to facilitate the availability of a valid,
  reachable default IP address for any [provisioners][tf-docs-provisioners].
  The behavior of the waiter can be controlled with the
  [`wait_for_guest_net_timeout`](#wait_for_guest_net_timeout),
  [`wait_for_guest_net_routable`](#wait_for_guest_net_routable),
  [`wait_for_guest_ip_timeout`](#wait_for_guest_ip_timeout), and
  [`ignored_guest_ips`](#ignored_guest_ips) settings.

[tf-docs-provisioners]: /docs/provisioners/index.html

### Migrating from a previous version of this resource

~> **NOTE:** This section only applies to versions of this resource available
in versions v0.4.2 of this provider or earlier.

The path for migrating to the current version of this resource is very similar
to the [import](#importing) path, with the exception that the `terraform
import` command does not need to be run. See that section for details on what
is required before you run `terraform plan` on a state that requires migration.

A successful migration usually only results in a configuration-only diff - that
is, Terraform reconciles some configuration settings that cannot be set during
the migration process with state. In this event, no reconfiguration operations
are sent to the vSphere server during the next `terraform apply`.  See the
[importing](#importing) section for more details.

## Example Usage

### Creating a virtual machine from scratch

The following block contains all that is necessary to create a new virtual
machine, with a single disk and network interface. 

The resource makes use of the following data sources to do its job:
[`vsphere_datacenter`][tf-vsphere-datacenter] to locate the datacenter,
[`vsphere_datastore`][tf-vsphere-datastore] to locate the default datastore to
put the virtual machine in, [`vsphere_resource_pool`][tf-vsphere-resource-pool]
to locate a resource pool located in a cluster or standalone host, and
[`vsphere_network`][tf-vsphere-network] to locate a network.

[tf-vsphere-datacenter]: /docs/providers/vsphere/d/datacenter.html
[tf-vsphere-datastore]: /docs/providers/vsphere/d/datastore.html
[tf-vsphere-resource-pool]: /docs/providers/vsphere/d/resource_pool.html
[tf-vsphere-network]: /docs/providers/vsphere/d/network.html

```hcl
data "vsphere_datacenter" "dc" {
  name = "dc1"
}

data "vsphere_datastore" "datastore" {
  name          = "datastore1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_network" "network" {
  name          = "public"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-test"
  resource_pool_id = "${data.vsphere_compute_cluster.cluster.resource_pool_id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  num_cpus = 2
  memory   = 1024
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
```

### Cloning and customization example

Building on the above example, the below configuration creates a VM by cloning
it from a template, fetched via the
[`vsphere_virtual_machine`][tf-vsphere-virtual-machine-ds] data source. This
allows us to locate the UUID of the template we want to clone, along with
settings for network interface type, SCSI bus type (especially important on
Windows machines), and disk attributes.

[tf-vsphere-virtual-machine-ds]: /docs/providers/vsphere/d/virtual_machine.html

~> **NOTE:** Cloning requires vCenter and is not supported on direct ESXi
connections.

```hcl
data "vsphere_datacenter" "dc" {
  name = "dc1"
}

data "vsphere_datastore" "datastore" {
  name          = "datastore1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_network" "network" {
  name          = "public"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_virtual_machine" "template" {
  name          = "ubuntu-16.04"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-test"
  resource_pool_id = "${data.vsphere_compute_cluster.cluster.resource_pool_id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  num_cpus = 2
  memory   = 1024
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  scsi_type = "${data.vsphere_virtual_machine.template.scsi_type}"

  network_interface {
    network_id   = "${data.vsphere_network.network.id}"
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

    customize {
      linux_options {
        host_name = "terraform-test"
        domain    = "test.internal"
      }

      network_interface {
        ipv4_address = "10.0.0.10"
        ipv4_netmask = 24
      }

      ipv4_gateway = "10.0.0.1"
    }
  }
}
```
### Deploying VM from an OVF/OVA template
Ovf and ova templates can be deployed both from local system and remote URL into the 
vcenter using the `ovf_deploy` property. When deploying from local system, the 
path to the ovf or ova template needs to be provided. While deploying ovf, all other 
necessary files like vmdk files also should be present in the same directory as the ovf file.

~> **NOTE:** Only the vApp properties which are pre-defined in the ovf template can be overwritten. 
vApp properties from scratch cannot be created.

~> **NOTE:** ovf deployment requires vCenter and is not supported on direct ESXi
connections.

```hcl
data "vsphere_datacenter" "dc" {
  name = "DC"
}

data "vsphere_datastore" "datastore" {
  name          = "datastore"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_resource_pool" "pool" {
  name          = "Cluster1/Resources"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_host" "host" {
  name          = "hostip"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_network" "network" {
  name          = "VM Network"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_virtual_machine" "vmFromLocalOvf" {
  name                       = "vm1"
  resource_pool_id           = data.vsphere_resource_pool.pool.id
  datastore_id               = data.vsphere_datastore.datastore.id
  host_system_id             = data.vsphere_host.host.id
  wait_for_guest_net_timeout = 0
  wait_for_guest_ip_timeout  = 0
  datacenter_id              = data.vsphere_datacenter.dc.id
  ovf_deploy {
    local_ovf_path       = "Full Path to local ovf/ova file"
    disk_provisioning    = "thin"
    ip_protocol          = "IPV4"
    ip_allocation_policy = "STATIC_MANUAL"
    ovf_network_map = {
      "ESX-port-1" = data.vsphere_network.network.id
      "ESX-port-2" = data.vsphere_network.network.id
    }
  }
  vapp {
    properties = {
      "guestinfo.tf.internal.id" = "42"
    }
  }
}

resource "vsphere_virtual_machine" "vmFromRemoteOvf" {
  name                       = "vm2"
  resource_pool_id           = data.vsphere_resource_pool.pool.id
  datastore_id               = data.vsphere_datastore.datastore.id
  host_system_id             = data.vsphere_host.host.id
  wait_for_guest_net_timeout = 0
  wait_for_guest_ip_timeout  = 0
  datacenter_id              = data.vsphere_datacenter.dc.id
  ovf_deploy {
    // Url to remote ovf/ova file
    remote_ovf_url = "https://download3.vmware.com/software/vmw-tools/nested-esxi/Nested_ESXi7.0_Appliance_Template_v1.ova"
  }
}
```

Using the OVF/OVA to deploy virtual machines can result in situations where the resulting VM does
match what's specified in the OVF specification because the provider will attempt to apply its own
settings on top of that. For example if `memory` is not set in a `vsphere_virtual_machine` resource,
then terraform will assume that it has to apply the default. There are also cases where the hardware
specifications may not be applicable and changes need to be made. The way to work around these cases
is to use the `vsphere_ovf_vm_template` data source that will parse the OVF template and export the
settings as outputs, that are then passed as parameters to the `vsphere_virtual_machine` resource.
Please note that the parameters to `ovf_deploy` are still required, in order for vSphere to do the
actual deployment, as shown in the example below:

```hcl
provider "vsphere" {}

data "vsphere_datacenter" "dc" {
  name = "hashidc"
}

data "vsphere_host" "hs" {
  name = "172.16.12.125"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_compute_cluster" "compute_cluster" {
  name          = "c1"
  datacenter_id = data.vsphere_datacenter.dc.id
}

resource "vsphere_resource_pool" "rp" {
  name = "rp1"
  parent_resource_pool_id = data.vsphere_compute_cluster.compute_cluster.resource_pool_id
}

data "vsphere_network" "net" {
  name = "VM Network"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_datastore" "ds" {
  name          = "nfs-vol1"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_ovf_vm_template" "ovf" {
  name             = "testOVF"
  resource_pool_id = vsphere_resource_pool.rp.id
  datastore_id     = data.vsphere_datastore.ds.id
  host_system_id   = data.vsphere_host.hs.id
  remote_ovf_url   = "https://download3.vmware.com/software/vmw-tools/nested-esxi/Nested_ESXi7.0_Appliance_Template_v1.ova"
  
  ovf_network_map = {
    "Network 1": data.vsphere_network.net.id
  }
}

resource "vsphere_virtual_machine" "vm" {
  datacenter_id = data.vsphere_datacenter.dc.id

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
    ovf_network_map = data.vsphere_ovf_vm_template.ovf.ovf_network_map
    remote_ovf_url  = data.vsphere_ovf_vm_template.ovf.remote_ovf_url
  }
}


```


### Cloning from an OVF/OVA-created template with vApp properties

This alternate example details how to clone a VM from a template that came from
an OVF/OVA file. This leverages the resource's [vApp
properties](#using-vapp-properties-to-supply-ovf-ova-configuration) capabilities to
set appropriate keys that control various configuration settings on the virtual
machine or virtual appliance. In this scenario, using `customize` is not
recommended as the functionality has tendency to overlap.

[ext-packer-io]: https://www.packer.io/
[ext-govc]: https://github.com/vmware/govmomi/tree/master/govc
[ext-ovftool]: https://code.vmware.com/web/dp/tool/ovf

```hcl
data "vsphere_datacenter" "dc" {
  name = "dc1"
}

data "vsphere_datastore" "datastore" {
  name          = "datastore1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_network" "network" {
  name          = "public"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_virtual_machine" "template_from_ovf" {
  name          = "template_from_ovf"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name             = "terraform-test"
  resource_pool_id = "${data.vsphere_compute_cluster.cluster.resource_pool_id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  num_cpus = 2
  memory   = 1024
  guest_id = "${data.vsphere_virtual_machine.template.guest_id}"

  scsi_type = "${data.vsphere_virtual_machine.template.scsi_type}"

  network_interface {
    network_id   = "${data.vsphere_network.network.id}"
    adapter_type = "${data.vsphere_virtual_machine.template.network_interface_types[0]}"
  }

  disk {
    name             = "disk0"
    size             = "${data.vsphere_virtual_machine.template.disks.0.size}"
    eagerly_scrub    = "${data.vsphere_virtual_machine.template.disks.0.eagerly_scrub}"
    thin_provisioned = "${data.vsphere_virtual_machine.template.disks.0.thin_provisioned}"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template_from_ovf.id}"
  }

  vapp {
    properties = {
      "guestinfo.tf.internal.id" = "42"
    }
  }
}
```

### Using Storage DRS

The `vsphere_virtual_machine` resource also supports Storage DRS, allowing the
assignment of virtual machines to datastore clusters. When assigned to a
datastore cluster, changes to a virtual machine's underlying datastores are
ignored unless disks drift outside of the datastore cluster. The example below
makes use of the [`vsphere_datastore_cluster` data
source][tf-vsphere-datastore-cluster-data-source], and the
[`datastore_cluster_id`](#datastore_cluster_id) configuration setting. Note
that the [`vsphere_datastore_cluster`
resource][tf-vsphere-datastore-cluster-resource] also exists to allow for
management of datastore clusters directly in Terraform.

[tf-vsphere-datastore-cluster-data-source]: /docs/providers/vsphere/d/datastore_cluster.html
[tf-vsphere-datastore-cluster-resource]: /docs/providers/vsphere/r/datastore_cluster.html

~> **NOTE:** When managing datastore clusters, member datastores, and virtual
machines within the same Terraform configuration, race conditions can apply.
This is because datastore clusters must be created before datastores can be
assigned to them, and the respective `vsphere_virtual_machine` resources will
no longer have an implicit dependency on the specific datastore resources. Use
[`depends_on`][tf-docs-depends-on] to create an explicit dependency on the
datastores in the cluster, or manage datastore clusters and datastores in a
separate configuration.

[tf-docs-depends-on]: /docs/configuration/resources.html#depends_on

```hcl
data "vsphere_datacenter" "dc" {
  name = "dc1"
}

data "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "datastore-cluster1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_network" "network" {
  name          = "public"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_virtual_machine" "vm" {
  name                 = "terraform-test"
  resource_pool_id     = "${data.vsphere_compute_cluster.cluster.resource_pool_id}"
  datastore_cluster_id = "${data.vsphere_datastore_cluster.datastore_cluster.id}"

  num_cpus = 2
  memory   = 1024
  guest_id = "other3xLinux64Guest"

  network_interface {
    network_id = "${data.vsphere_network.network.id}"
  }

  disk {
    label = "disk0"
    size  = 20
  }
}
```

## Argument Reference

The following arguments are supported:

### General options

The following options are general virtual machine and Terraform workflow
options:

* `name` - (Required) The name of the virtual machine.
* `resource_pool_id` - (Required) The [managed object reference
  ID][docs-about-morefs] of the resource pool to put this virtual machine in.
  See the section on [virtual machine migration](#virtual-machine-migration)
  for details on changing this value.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

~> **NOTE:** All clusters and standalone hosts have a resource pool, even if
one has not been explicitly created. For more information, see the section on
[specifying the root resource pool ][docs-resource-pool-cluster-default] in the
`vsphere_resource_pool` data source documentation. This resource does not take
a cluster or standalone host resource directly.

[docs-resource-pool-cluster-default]: /docs/providers/vsphere/d/resource_pool.html#specifying-the-root-resource-pool-for-a-standalone-host

* `datastore_id` - (Optional) The [managed object reference
  ID][docs-about-morefs] of the virtual machine's datastore. The virtual
  machine configuration is placed here, along with any virtual disks that are
  created where a datastore is not explicitly specified. See the section on
  [virtual machine migration](#virtual-machine-migration) for details on
  changing this value.
* `datastore_cluster_id` - (Optional) The [managed object reference
  ID][docs-about-morefs] of the datastore cluster ID to use. This setting
  applies to entire virtual machine and implies that you wish to use Storage
  DRS with this virtual machine. See the section on [virtual machine
  migration](#virtual-machine-migration) for details on changing this value.

~> **NOTE:** One of `datastore_id` or `datastore_cluster_id` must be specified.

~> **NOTE:** Use of `datastore_cluster_id` requires Storage DRS to be enabled
on that cluster. 

~> **NOTE:** The `datastore_cluster_id` setting applies to the entire virtual
machine - you cannot assign individual datastore clusters to individual disks.
In addition to this, you cannot use the [`attach`](#attach) setting to attach
external disks on virtual machines that are assigned to datastore clusters.

* `datacenter_id` - (Optional) The datacenter id. Required only when deploying
   an ovf template.
* `folder` - (Optional) The path to the folder to put this virtual machine in,
  relative to the datacenter that the resource pool is in.
* `host_system_id` - (Optional) An optional [managed object reference
  ID][docs-about-morefs] of a host to put this virtual machine on. See the
  section on [virtual machine migration](#virtual-machine-migration) for
  details on changing this value. If a `host_system_id` is not supplied,
  vSphere will select a host in the resource pool to place the virtual machine,
  according to any defaults or DRS policies in place. 
* `disk` - (Required) A specification for a virtual disk device on this virtual
  machine. See [disk options](#disk-options) below.
* `network_interface` - (Required) A specification for a virtual NIC on this
  virtual machine. See [network interface options](#network-interface-options)
  below.
* `cdrom` - (Optional) A specification for a CDROM device on this virtual
  machine. See [CDROM options](#cdrom-options) below.
* `clone` - (Optional) When specified, the VM will be created as a clone of a
  specified template. Optional customization options can be submitted as well.
  See [creating a virtual machine from a
  template](#creating-a-virtual-machine-from-a-template) for more details.
 * `hardware_version` - (Optional) The hardware version number. Valid range
   is from 4 to 15. The hardware version cannot be downgraded. See [virtual
   machine hardware compatibility][virtual-machine-hardware-compatibility] for
   more details.
 * `pci_device_id` - (Optional) List of host PCI device IDs to create PCI 
   passthroughs for.
   
[virtual-machine-hardware-compatibility]: https://kb.vmware.com/s/article/2007240

~> **NOTE:** Cloning requires vCenter and is not supported on direct ESXi
connections.

* `ovf_deploy` - (Optional) When specified, the VM will be deployed from the
   provided ovf/ova template. See [creating a virtual machine from a 
   ovf/ova template](#creating-vm-from-deploying-a-ovf-ova-template) for more details.
* `vapp` - (Optional) Optional vApp configuration. The only sub-key available
  is `properties`, which is a key/value map of properties for virtual machines
  imported from OVF or OVA files. See [Using vApp properties to supply OVF/OVA
  configuration](#using-vapp-properties-to-supply-ovf-ova-configuration) for
  more details.
* `guest_id` - (Optional) The guest ID for the operating system type. For a
  full list of possible values, see [here][vmware-docs-guest-ids]. Default: `other-64`.

[vmware-docs-guest-ids]: https://code.vmware.com/apis/358/doc/vim.vm.GuestOsDescriptor.GuestOsIdentifier.html

* `alternate_guest_name` - (Optional) The guest name for the operating system
  when `guest_id` is `other` or `other-64`.
* `annotation` - (Optional) A user-provided description of the virtual machine.
  The default is no annotation.
* `firmware` - (Optional) The firmware interface to use on the virtual machine.
  Can be one of `bios` or `EFI`. Default: `bios`.
* `extra_config` - (Optional) Extra configuration data for this virtual
  machine. Can be used to supply advanced parameters not normally in
  configuration, such as instance metadata. 

~> **NOTE:** Do not use `extra_config` when working with a template imported
from OVF or OVA as more than likely your settings will be ignored. Use the
`vapp` block's `properties` section as outlined in [Using vApp properties to
supply OVF/OVA
configuration](#using-vapp-properties-to-supply-ovf-ova-configuration).

* `replace_trigger` - (Optional) Triggers replacement of resource whenever it changes.
  `replace_trigger = sha256(format("%s-%s",data.template_file.cloud_init_metadata.rendered,data.template_file.cloud_init_userdata.rendered))`
  will fingerprint the changes in cloud_init metadata and userdata templates. This will enable a replacement
  of the resource whenever the dependant template renders a new configuration. (Forces a replacement)
* `scsi_type` - (Optional) The type of SCSI bus this virtual machine will have.
  Can be one of lsilogic (LSI Logic Parallel), lsilogic-sas (LSI Logic SAS) or
  pvscsi (VMware Paravirtual). Defualt: `pvscsi`.
* `scsi_bus_sharing` - (Optional) Mode for sharing the SCSI bus. The modes are
  physicalSharing, virtualSharing, and noSharing. Default: `noSharing`.
* `tags` - (Optional) The IDs of any tags to attach to this resource. See
  [here][docs-applying-tags] for a reference on how to apply tags.

[docs-applying-tags]: /docs/providers/vsphere/r/tag.html#using-tags-in-a-supported-resource

~> **NOTE:** Tagging support is unsupported on direct ESXi connections and
requires vCenter 6.0 or higher.

* `custom_attributes` - (Optional) Map of custom attribute ids to attribute
  value strings to set for virtual machine. See 
  [here][docs-setting-custom-attributes] for a reference on how to set values 
  for custom attributes.

[docs-setting-custom-attributes]: /docs/providers/vsphere/r/custom_attribute.html#using-custom-attributes-in-a-supported-resource

~> **NOTE:** Custom attributes are unsupported on direct ESXi connections 
and require vCenter.

* `storage_policy_id` - (Optional) The UUID of the storage policy to assign to VM home directory.

### CPU and memory options

The following options control CPU and memory settings on the virtual machine:

* `num_cpus` - (Optional) The total number of virtual processor cores to assign
  to this virtual machine. Default: `1`.
* `num_cores_per_socket` - (Optional) The number of cores per socket in this
  virtual machine. The number of vCPUs on the virtual machine will be
  `num_cpus` divided by `num_cores_per_socket`. If specified, the value
  supplied to `num_cpus` must be evenly divisible by this value. Default: `1`.
* `cpu_hot_add_enabled` - (Optional) Allow CPUs to be added to this virtual
  machine while it is running.
* `cpu_hot_remove_enabled` - (Optional) Allow CPUs to be removed to this
  virtual machine while it is running.
* `memory` - (Optional) The size of the virtual machine's memory, in MB.
  Default: `1024` (1 GB).
* `memory_hot_add_enabled` - (Optional) Allow memory to be added to this
  virtual machine while it is running.

~> **NOTE:** Certain CPU and memory hot-plug options are not available on every
operating system. Check the [VMware Guest OS Compatibility
Guide][vmware-docs-compat-guide] first to see what settings your guest
operating system is eligible for. In addition, at least one `terraform apply`
must be executed before being able to take advantage of CPU and memory hot-plug
settings, so if you want the support, enable it as soon as possible.

[vmware-docs-compat-guide]: http://partnerweb.vmware.com/comp_guide2/pdf/VMware_GOS_Compatibility_Guide.pdf

### Boot options 

The following options control boot settings on the virtual machine:

* `boot_delay` - (Optional) The number of milliseconds to wait before starting
  the boot sequence. The default is no delay.
* `efi_secure_boot_enabled` - (Optional) When the `firmware` type is set to is
  `efi`, this enables EFI secure boot. Default: `false`.

~> **NOTE:** EFI secure boot is only available on vSphere 6.5 and higher.

* `boot_retry_delay` - (Optional) The number of milliseconds to wait before
  retrying the boot sequence. This only valid if `boot_retry_enabled` is true.
  Default: `10000` (10 seconds).
* `boot_retry_enabled` - (Optional) If set to true, a virtual machine that
  fails to boot will try again after the delay defined in `boot_retry_delay`.
  Default: `false`.

### VMware Tools options

The following options control VMware tools options on the virtual machine:

* `sync_time_with_host` - (Optional) Enable guest clock synchronization with
  the host. Requires VMware tools to be installed. Default: `false`.
* `run_tools_scripts_after_power_on` - (Optional) Enable the execution of
  post-power-on scripts when VMware tools is installed. Default: `true`.
* `run_tools_scripts_after_resume` - (Optional) Enable the execution of
  post-resume scripts when VMware tools is installed. Default: `true`.
* `run_tools_scripts_before_guest_reboot` - (Optional) Enable the execution of
  pre-reboot scripts when VMware tools is installed. Default: `false`.
* `run_tools_scripts_before_guest_shutdown` - (Optional) Enable the execution
  of pre-shutdown scripts when VMware tools is installed. Default: `true`.
* `run_tools_scripts_before_guest_standby` - (Optional) Enable the execution of
  pre-standby scripts when VMware tools is installed. Default: `true`.

### Resource allocation options

The following options allow control over CPU and memory allocation on the
virtual machine. Note that the resource pool that this VM is in may affect
these options.

* `cpu_limit` - (Optional) The maximum amount of CPU (in MHz) that this virtual
  machine can consume, regardless of available resources. The default is no
  limit.
* `cpu_reservation` - (Optional) The amount of CPU (in MHz) that this virtual
  machine is guaranteed. The default is no reservation.
* `cpu_share_level` - (Optional) The allocation level for CPU resources. Can be
  one of `high`, `low`, `normal`, or `custom`. Default: `custom`.
* `cpu_share_count` - (Optional) The number of CPU shares allocated to the
  virtual machine when the `cpu_share_level` is `custom`.
* `memory_limit` - (Optional) The maximum amount of memory (in MB) that this
  virtual machine can consume, regardless of available resources. The default
  is no limit.
* `memory_reservation` - (Optional) The amount of memory (in MB) that this
  virtual machine is guaranteed. The default is no reservation.
* `memory_share_level` - (Optional) The allocation level for memory resources.
  Can be one of `high`, `low`, `normal`, or `custom`. Default: `custom`.
* `memory_share_count` - (Optional) The number of memory shares allocated to
  the virtual machine when the `memory_share_level` is `custom`.

### Advanced options

The following options control advanced operation of the virtual machine, or
control various parts of Terraform workflow, and should not need to be modified
during basic operation of the resource. Only change these options if they are
explicitly required, or if you are having trouble with Terraform's default
behavior.

* `enable_disk_uuid` - (Optional) Expose the UUIDs of attached virtual disks to
  the virtual machine, allowing access to them in the guest. Default: `false`.
* `vbs_enabled` - (Optional) Enable Virtualization Based Security. Requires
  `firmware` to be `efi`, and `vvtd_enabled`, `nested_hv_enabled` and
  `efi_secure_boot_enabled` must all have a value of `true`. Supported on
  vSphere 6.7 and higher. Default: `false`.
* `vvtd_enabled` - (Optional) Flag to specify if Intel Virtualization Technology 
  for Directed I/O is enabled for this virtual machine (_I/O MMU_ in the
  vSphere Client). Supported on vSphere 6.7 and higher. Default: `false`.
* `hv_mode` - (Optional) The (non-nested) hardware virtualization setting for
  this virtual machine. Can be one of `hvAuto`, `hvOn`, or `hvOff`. Default:
  `hvAuto`.
* `ept_rvi_mode` - (Optional) The EPT/RVI (hardware memory virtualization)
  setting for this virtual machine. Can be one of `automatic`, `on`, or `off`.
  Default: `automatic`.
* `nested_hv_enabled` - (Optional) Enable nested hardware virtualization on
  this virtual machine, facilitating nested virtualization in the guest.
  Default: `false`.
* `enable_logging` - (Optional) Enable logging of virtual machine events to a
  log file stored in the virtual machine directory. Default: `false`.
* `cpu_performance_counters_enabled` - (Optional) Enable CPU performance
  counters on this virtual machine. Default: `false`.
* `swap_placement_policy` - (Optional) The swap file placement policy for this
  virtual machine. Can be one of `inherit`, `hostLocal`, or `vmDirectory`.
  Default: `inherit`.
* `latency_sensitivity` - (Optional) Controls the scheduling delay of the
  virtual machine. Use a higher sensitivity for applications that require lower
  latency, such as VOIP, media player applications, or applications that
  require frequent access to mouse or keyboard devices. Can be one of `low`,
  `normal`, `medium`, or `high`.

~> **NOTE:** Do not use a `latency_sensitivity` setting of `low` or `medium` on
hosts running ESXi 6.0 or older. Doing so may result in virtual machine startup
issues or spurious diffs in Terraform. In addition, on higher sensitivities,
you may have to adjust [`memory_reservation`](#memory_reservation) to the full
amount of memory provisioned for the virtual machine.

* `wait_for_guest_net_timeout` - (Optional) The amount of time, in minutes, to
  wait for an available IP address on this virtual machine's NICs. Older
  versions of VMware Tools do not populate this property. In those cases, this
  waiter can be disabled and the
  [`wait_for_guest_ip_timeout`](#wait_for_guest_ip_timeout) waiter can be used
  instead. A value less than 1 disables the waiter. Default: 5 minutes.
* `wait_for_guest_net_routable` - (Optional) Controls whether or not the guest
  network waiter waits for a routable address. When `false`, the waiter does
  not wait for a default gateway, nor are IP addresses checked against any
  discovered default gateways as part of its success criteria. This property is
  ignored if the [`wait_for_guest_ip_timeout`](#wait_for_guest_ip_timeout)
  waiter is used. Default: `true`.
* `wait_for_guest_ip_timeout` - (Optional) The amount of time, in minutes, to
  wait for an available guest IP address on this virtual machine. This should
  only be used if your version of VMware Tools does not allow the
  [`wait_for_guest_net_timeout`](#wait_for_guest_net_timeout) waiter to be
  used. A value less than 1 disables the waiter. Default: 0.
* `ignored_guest_ips` - (Optional) List of IP addresses and CIDR networks to
  ignore while waiting for an available IP address using either of the waiters.
  Any IP addresses in this list will be ignored if they show up so that the
  waiter will continue to wait for a real IP address. Default: [].
* `shutdown_wait_timeout` - (Optional) The amount of time, in minutes, to wait
  for a graceful guest shutdown when making necessary updates to the virtual
  machine. If `force_power_off` is set to true, the VM will be force powered-off
  after this timeout, otherwise an error is returned. Default: 3 minutes.
* `migrate_wait_timeout` - (Optional) The amount of time, in minutes, to wait
  for a virtual machine migration to complete before failing. Default: 10
  minutes. Also see the section on [virtual machine
  migration](#virtual-machine-migration).
* `force_power_off` - (Optional) If a guest shutdown failed or timed out while
  updating or destroying (see
  [`shutdown_wait_timeout`](#shutdown_wait_timeout)), force the power-off of
  the virtual machine. Default: `true`.
* `scsi_controller_count` - (Optional) The number of SCSI controllers that
  Terraform manages on this virtual machine. This directly affects the amount
  of disks you can add to the virtual machine and the maximum disk unit number.
  Note that lowering this value does not remove controllers. Default: `1`.
* `sata_controller_count` - (Optional) The number of SATA controllers that
  Terraform manages on this virtual machine. This directly affects the amount
  of disks you can add to the virtual machine and the maximum disk unit number.
  Note that lowering this value does not remove controllers. Default: `0`.
* `ide_controller_count` - (Optional) The number of IDE controllers that
  Terraform manages on this virtual machine. This directly affects the amount
  of disks you can add to the virtual machine and the maximum disk unit number.
  Note that lowering this value does not remove controllers. Default: `2`.

### Disk options

Virtual disks are managed by adding an instance of the `disk` block.

At the very least, there must be `name` and `size` attributes. `unit_number` is
required for any disk other than the first, and there must be at least one
resource with the implicit number of 0.

An abridged multi-disk example is below:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  disk {
    label = "disk0"
    size  = "10"
  }

  disk {
    label       = "disk1"
    size        = "100"
    unit_number = 1
  }

  # ... other configuration ...
}
```

The options are:

* `label` - (Required) A label for the disk. Forces a new disk if changed.

~> **NOTE:** It's recommended that you set the disk label to a format matching
`diskN`, where `N` is the number of the disk, starting from disk number 0. This
will ensure that your configuration is compatible when importing a virtual
machine. For more information, see the section on [importing](#importing).

~> **NOTE:** Do not choose a label that starts with `orphaned_disk_` (example:
`orphaned_disk_0`), as this prefix is reserved for disks that Terraform does
not recognize, such as disks that are attached externally. Terraform will issue
an error if you try to label a disk with this prefix. 

* `size` - (Required) The size of the disk, in GB.
* `unit_number` - (Optional) The disk number on the storage bus. The maximum 
  value for this setting is the value of the controller count times the 
  controller capacity (15 for SCSI, 30 for SATA, and 2 for IDE). 
  The default is `0`, for which one disk must be set to. Duplicate unit numbers 
  are not allowed.
* `datastore_id` - (Optional) A [managed object reference
  ID][docs-about-morefs] to the datastore for this virtual disk. The default is
  to use the datastore of the virtual machine. See the section on [virtual
  machine migration](#virtual-machine-migration) for details on changing this
  value.

~> **NOTE:** Datastores cannot be assigned to individual disks when
[`datastore_cluster_id`](#datastore_cluster_id) is in use.

* `attach` - (Optional) Attach an external disk instead of creating a new one.
  Implies and conflicts with `keep_on_remove`. If set, you cannot set `size`,
  `eagerly_scrub`, or `thin_provisioned`. Must set `path` if used.

~> **NOTE:** External disks cannot be attached when
[`datastore_cluster_id`](#datastore_cluster_id) is in use.

* `path` - (Optional) When using `attach`, this parameter controls the path of
  a virtual disk to attach externally. Otherwise, it is a computed attribute
  that contains the virtual disk's current filename.
* `keep_on_remove` - (Optional) Keep this disk when removing the device or
  destroying the virtual machine. Default: `false`.
* `disk_mode` - (Optional) The mode of this this virtual disk for purposes of
  writes and snapshotting. Can be one of `append`, `independent_nonpersistent`,
  `independent_persistent`, `nonpersistent`, `persistent`, or `undoable`.
  Default: `persistent`. For an explanation of options, click
  [here][vmware-docs-disk-mode].

[vmware-docs-disk-mode]: https://pubs.vmware.com/vsphere-6-5/topic/com.vmware.wssdk.apiref.doc/vim.vm.device.VirtualDiskOption.DiskMode.html

* `eagerly_scrub` - (Optional) If set to `true`, the disk space is zeroed out
  on VM creation. This will delay the creation of the disk or virtual machine.
  Cannot be set to `true` when `thin_provisioned` is `true`.  See the section
  on [picking a disk type](#picking-a-disk-type).  Default: `false`.
* `thin_provisioned` - (Optional) If `true`, this disk is thin provisioned,
  with space for the file being allocated on an as-needed basis. Cannot be set
  to `true` when `eagerly_scrub` is `true`. See the section on [picking a disk
  type](#picking-a-disk-type). Default: `true`. 
* `disk_sharing` - (Optional) The sharing mode of this virtual disk. Can be one
  of `sharingMultiWriter` or `sharingNone`. Default: `sharingNone`.

~> **NOTE:** Disk sharing is only available on vSphere 6.0 and higher.

* `write_through` - (Optional) If `true`, writes for this disk are sent
  directly to the filesystem immediately instead of being buffered. Default:
  `false`.
* `io_limit` - (Optional) The upper limit of IOPS that this disk can use. The
  default is no limit.
* `io_reservation` - (Optional) The I/O reservation (guarantee) that this disk
  has, in IOPS.  The default is no reservation.
* `io_share_level` - (Optional) The share allocation level for this disk. Can
  be one of `low`, `normal`, `high`, or `custom`. Default: `normal`.
* `io_share_count` - (Optional) The share count for this disk when the share
  level is `custom`.
* `storage_policy_id` - (Optional) The UUID of the storage policy to assign to this disk.
* `controller_type` - (Optional) The type of storage controller to attach the 
  disk to. Can be `scsi`, `sata`, or `ide`. You must have the appropriate
  number of controllers enabled for the selected type. Default `scsi`.

#### Computed disk attributes

* `uuid` - The UUID of the virtual disk's VMDK file. This is used to track the
  virtual disk on the virtual machine.

#### Picking a disk type

The `eagerly_scrub` and `thin_provisioned` options control the space allocation
type of a virtual disk. These show up in the vSphere console as a unified
enumeration of options, the equivalents of which are explained below. The
defaults in Terraform are the equivalent of thin provisioning.

* **Thick provisioned lazy zeroed:** Both `eagerly_scrub` and
  `thin_provisioned` should be set to `false`.
* **Thick provisioned eager zeroed:** `eagerly_scrub` should be set to true,
  and `thin_provisioned` should be set to `false`.
* **Thin provisioned:** `eagerly_scrub` should be set to `false`, and
  `thin_provisioned` should be set to `true`.

For the technical details of each virtual disk provisioning policy, click
[here][docs-vmware-vm-disk-provisioning].

[docs-vmware-vm-disk-provisioning]: https://docs.vmware.com/en/VMware-vSphere/6.5/com.vmware.vsphere.vm_admin.doc/GUID-4C0F4D73-82F2-4B81-8AA7-1DD752A8A5AC.html

~> **NOTE:** Not all disk types are available on some types of datastores.
Attempting to set options inappropriate for a datastore that a disk is deployed
to will result in a successful initial apply, but vSphere will silently correct
the options, and subsequent plans will fail with an appropriate error message
until the settings are corrected.

~> **NOTE:** The disk type cannot be changed once set.

### Network interface options

Network interfaces are managed by adding an instance of the `network_interface`
block.

Interfaces are assigned to devices in the specific order they are declared.
This has different implications for different operating systems.

Given the following example:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  network_interface {
    network_id = "${data.vsphere_network.public.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.private.id}"
  }
}
```

The first interface with the `public` network assigned to it would show up in
order before the interface assigned to `private`. On some Linux systems, this
might mean that the first interface would show up as `eth0` and the second
would show up as `eth1`.

The options are:

* `network_id` - (Required) The [managed object reference
  ID][docs-about-morefs] of the network to connect this interface to.
* `adapter_type` - (Optional) The network interface type. Can be one of
  `e1000`, `e1000e`, or `vmxnet3`. Default: `vmxnet3`.
* `use_static_mac` - (Optional) If true, the `mac_address` field is treated as
  a static MAC address and set accordingly. Setting this to `true` requires
  `mac_address` to be set. Default: `false`.
* `mac_address` - (Optional) The MAC address of this network interface. Can
  only be manually set if `use_static_mac` is true, otherwise this is a
  computed value that gives the current MAC address of this interface.
* `bandwidth_limit` - (Optional) The upper bandwidth limit of this network
  interface, in Mbits/sec. The default is no limit.
* `bandwidth_reservation` - (Optional) The bandwidth reservation of this
  network interface, in Mbits/sec. The default is no reservation.
* `bandwidth_share_level` - (Optional) The bandwidth share allocation level for
  this interface. Can be one of `low`, `normal`, `high`, or `custom`. Default:
  `normal`.
* `bandwidth_share_count` - (Optional) The share count for this network
  interface when the share level is `custom`.
* `ovf_mapping` - (Optional) Specifies which OVF NIC the `network_interface`
  should be associated with. Only applies at creation and only when deploying
  from an OVF source.

### CDROM options

A single virtual CDROM device can be created and attached to the virtual
machine. The resource supports attaching a CDROM from a datastore ISO or
using a remote client device.

An example is below:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  cdrom {
    datastore_id = "${data.vsphere_datastore.iso_datastore.id}"
    path         = "ISOs/os-livecd.iso"
  }
}
```

The options are:

* `client_device` - (Optional) Indicates whether the device should be backed by
  remote client device. Conflicts with `datastore_id` and `path`.
* `datastore_id` - (Optional) The datastore ID that the ISO is located in.
  Requried for using a datastore ISO. Conflicts with `client_device`.
* `path` - (Optional) The path to the ISO file. Required for using a datastore
  ISO. Conflicts with `client_device`.

~> **NOTE:** Either `client_device` (for a remote backed CDROM) or `datastore_id`
and path (for a datastore ISO backed CDROM) are required.

~> **NOTE:** Some CDROM drive types are currently unsupported by this resource,
such as pass-through devices. If these drives are present in a cloned template,
or added outside of Terraform, they will have their configurations corrected to
that of the defined device, or removed if no `cdrom` block is present.

### Virtual device computed options

Configured virtual devices (`disk`, `network_interface`, and `cdrom`) all
export the following attributes. These options help locate the device on future
Terraform runs. The options are:

* `key` - The ID of the device within the virtual machine.
* `device_address` - An address internal to Terraform that helps locate the
  device when `key` is unavailable. This follows a convention of
  `CONTROLLER_TYPE:BUS_NUMBER:UNIT_NUMBER`. Example: `scsi:0:1` means device
  unit 1 on SCSI bus 0.

## Creating a Virtual Machine from a Template

The `clone` block can be used to create a new virtual machine from an existing
virtual machine or template. The resource supports both making a complete copy
of a virtual machine, or cloning from a snapshot (otherwise known as a linked
clone).

See the [cloning and customization
example](#cloning-and-customization-example) for a usage synopsis.

~> **NOTE:** Changing any option in `clone` after creation forces a new
resource.

~> **NOTE:** Cloning requires vCenter and is not supported on direct ESXi
connections.

The options available in the `clone` block are:

* `template_uuid` - (Required) The UUID of the source virtual machine or
  template.
* `linked_clone` - (Optional) Clone this virtual machine from a snapshot or
  a template. Default: `false`.
* `timeout` - (Optional) The timeout, in minutes, to wait for the cloning
  process to complete. Default: 30 minutes.
* `customize` - (Optional) The customization spec for this clone. This allows
  the user to configure the virtual machine post-clone. For more details, see
  [virtual machine customization](#virtual-machine-customization).

### Virtual machine customization

As part of the `clone` operation, a virtual machine can be
[customized][vmware-docs-customize] to configure host, network, or licensing
settings.

[vmware-docs-customize]: https://docs.vmware.com/en/VMware-vSphere/6.5/com.vmware.vsphere.vm_admin.doc/GUID-58E346FF-83AE-42B8-BE58-253641D257BC.html

To perform virtual machine customization as a part of the clone process,
specify the `customize` block with the respective customization options, nested
within the `clone` block. Windows guests are customized using Sysprep, which
will result in the machine SID being reset. Before using customization, check
is that your source VM meets the [requirements](https://pubs.vmware.com/vsphere-50/index.jsp?topic=%2Fcom.vmware.vsphere.vm_admin.doc_50%2FGUID-80F3F5B5-F795-45F1-B0FA-3709978113D5.html)
for guest OS customization on vSphere. See the [cloning and customization
example](#cloning-and-customization-example) for a usage synopsis.

The settings for `customize` are as follows:

#### Customization timeout settings

* `timeout` - (Optional) The time, in minutes that Terraform waits for
  customization to complete before failing. The default is 10 minutes, and
  setting the value to 0 or a negative value disables the waiter altogether.

#### Network interface settings

These settings, which should be specified in nested `network_interface` blocks
within [`customize`](#virtual-machine-customization), configure network
interfaces on a per-interface basis and are matched up to
[`network_interface`](#network-interface-options) devices in the order they are
declared.

Given the following example:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  network_interface {
    network_id = "${data.vsphere_network.public.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.private.id}"
  }

  clone {
    # ... other configuration ...

    customize {
      # ... other configuration ...

      network_interface {
        ipv4_address = "10.0.0.10"
        ipv4_netmask = 24
      }

      network_interface {
        ipv4_address = "172.16.0.10"
        ipv4_netmask = 24
      }

      ipv4_gateway = "10.0.0.1"
    }
  }
}
```

The first set of `network_interface` data would be assigned to the `public`
interface, and the second to the `private` interface.

To use DHCP, declare an empty `network_interface` block for each interface
being configured. So the above example would look like:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  network_interface {
    network_id = "${data.vsphere_network.public.id}"
  }

  network_interface {
    network_id = "${data.vsphere_network.private.id}"
  }

  clone {
    # ... other configuration ...

    customize {
      # ... other configuration ...

      network_interface {}

      network_interface {}
    }
  }
}
```

The options are:

* `dns_server_list` - (Optional) Network interface-specific DNS server settings
  for Windows operating systems. Ignored on Linux and possibly other operating
  systems - for those systems, please see the [global DNS
  settings](#global-dns-settings) section.
* `dns_domain` - (Optional) Network interface-specific DNS search domain for
  Windows operating systems. Ignored on Linux and possibly other operating
  systems - for those systems, please see the [global DNS
  settings](#global-dns-settings) section.
* `ipv4_address` - (Optional) The IPv4 address assigned to this network adapter. If left
  blank or not included, DHCP is used.
* `ipv4_netmask` The IPv4 subnet mask, in bits (example: `24` for
  255.255.255.0).
* `ipv6_address` - (Optional) The IPv6 address assigned to this network adapter. If left
  blank or not included, auto-configuration is used.
* `ipv6_netmask` - (Optional) The IPv6 subnet mask, in bits (example: `32`).

~> **NOTE:** The minimum setting for IPv4 in a customization specification is
DHCP. If you are setting up an IPv6-exclusive network without DHCP, you might
need to set [`wait_for_guest_net_timeout`](#wait_for_guest_net_timeout) to a
high enough value to cover the DHCP timeout of your virtual machine, or turn it
off altogether by supplying a zero or negative value. Keep in mind that turning
off `wait_for_guest_net_timeout` will more than likely mean that IP addresses
will not be reported to any provisioners you may have configured on the
resource.

#### Global routing settings

VM customization under the `vsphere_virtual_machine` resource does not take a
per-interface gateway setting, but rather default routes are configured on a
global basis. For an example, see the [network interface settings
section](#network-interface-settings).

The settings here must match the IP/mask of at least one `network_interface`
supplied to customization.

The options are:

* `ipv4_gateway` - (Optional) The IPv4 default gateway when using
  `network_interface` customization on the virtual machine.
* `ipv6_gateway` - (Optional) The IPv6 default gateway when using
  `network_interface` customization on the virtual machine.

#### Global DNS settings

The following settings configure DNS globally, generally for Linux systems. For
Windows systems, this is done per-interface, see [network interface
settings](#network-interface-settings).

* `dns_server_list` - The list of DNS servers to configure on a virtual
  machine. 
* `dns_suffix_list` - A list of DNS search domains to add to the DNS
  configuration on the virtual machine.

#### Linux customization options

The settings in the `linux_options` block pertain to Linux guest OS
customization. If you are customizing a Linux operating system, this section
must be included.

Example:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  clone {
    # ... other configuration ...

    customize {
      # ... other configuration ...

      linux_options {
        host_name = "terraform-test"
        domain    = "test.internal"
      }
    }
  }
}
```

The options are:

* `host_name` - (Required) The host name for this machine. This, along with
  `domain`, make up the FQDN of this virtual machine.
* `domain` - (Required) The domain name for this machine. This, along with
  `host_name`, make up the FQDN of this virtual machine.
* `hw_clock_utc` - (Optional) Tells the operating system that the hardware
  clock is set to UTC. Default: `true`.
* `time_zone` - (Optional) Sets the time zone. For a list of possible
  combinations, click [here][vmware-docs-valid-linux-tzs]. The default is UTC.

[vmware-docs-valid-linux-tzs]: https://pubs.vmware.com/vsphere-6-5/topic/com.vmware.wssdk.apiref.doc/timezone.html

#### Windows customization options

The settings in the `windows_options` block pertain to Windows guest OS
customization. If you are customizing a Windows operating system, this section
must be included.

Example:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  clone {
    # ... other configuration ...

    customize {
      # ... other configuration ...

      windows_options {
        computer_name  = "terraform-test"
        workgroup      = "test"
        admin_password = "VMw4re"
      }
    }
  }
}
```

The options are:

* `computer_name` - (Required) The computer name of this virtual machine.
* `admin_password` - (Optional) The administrator password for this virtual
  machine.

~> **NOTE:** `admin_password` is a sensitive field in Terraform and will not be
output on-screen, but is stored in state and sent to the VM in plain text -
keep this in mind when provisioning your infrastructure.

* `workgroup` - (Optional) The workgroup name for this virtual machine. One of
  this or `join_domain` must be included.
* `join_domain` - (Optional) The domain to join for this virtual machine. One
  of this or `workgroup` must be included.
* `domain_admin_user` - (Optional) The user of the domain administrator used to
  join this virtual machine to the domain. Required if you are setting `join_domain`.
* `domain_admin_password` - (Optional) The password of the domain administrator
  used to join this virtual machine to the domain. Required if you are setting
  `join_domain`.

~> **NOTE:** `domain_admin_password` is a sensitive field in Terraform and will
not be output on-screen, but is stored in state and sent to the VM in plain
text - keep this in mind when provisioning your infrastructure.

* `full_name` - (Optional) The full name of the user of this virtual machine.
  This populates the "user" field in the general Windows system information.
  Default: `Administrator`.
* `organization_name` - (Optional) The organization name this virtual machine
  is being installed for.  This populates the "organization" field in the
  general Windows system information.  Default: `Managed by Terraform`.
* `product_key` - (Optional) The product key for this virtual machine. The
  default is no key.
* `run_once_command_list` - (Optional) A list of commands to run at first user
  logon, after guest customization. Each command is limited by the API to 260
  characters.
* `auto_logon` - (Optional) Specifies whether or not the VM automatically logs
  on as Administrator. Default: `false`.
* `auto_logon_count` - (Optional) Specifies how many times the VM should auto-logon
  the Administrator account when `auto_logon` is true. This should be set
  accordingly to ensure that all of your commands that run in
  `run_once_command_list` can log in to run. Default: `1`.
* `time_zone` - (Optional) The new time zone for the virtual machine. This is a
  numeric, sysprep-dictated, timezone code. For a list of codes, click
  [here][ms-docs-valid-sysprep-tzs]. The default is `85` (GMT/UTC).

[ms-docs-valid-sysprep-tzs]: https://msdn.microsoft.com/en-us/library/ms912391(v=winembedded.11).aspx

#### Supplying your own SysPrep file

Alternative to the `windows_options` supplied above, you can instead supply
your own `sysprep.xml` file contents via the `windows_sysprep_text` option.
This allows full control of the customization process out-of-band of vSphere.
Example below:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  clone {
    # ... other configuration ...

    customize {
      # ... other configuration ...

      windows_sysprep_text = "${file("${path.module}/sysprep.xml")}"
    }
  }
}
```

Note this option is mutually exclusive to `windows_options` - one must not be
included if the other is specified.

### Creating VM from deploying a OVF/OVA template

The `ovf_deploy` block can be used to create a new virtual machine from an ovf/ova
template either from local system or remote URL. While deploying, the VM
properties are taken from ovf and setting them in configuration file is not necessary.

See the [Deploying from OVF example](#deploying-vm-from-an-ovf-ova-template) for a usage synopsis.

~> **NOTE:** Changing any option in `ovf_deploy` after creation forces a new
resource.

The options available in the `ovf_deploy` block are:

* `local_ovf_path` - (Optional) The absolute path to the ovf/ova file in the local system. While deploying from ovf,
   make sure the other necessary files like the .vmdk files are also in the same directory as the given ovf file.
* `remote_ovf_url` - (Optional) URL to the remote ovf/ova file to be deployed.

~> **NOTE:** Either `local_ovf_path` or `remote_ovf_url` is required, both can't be empty.

* `ip_allocation_policy` - (Optional) The IP allocation policy.
* `ip_protocol` - (Optional) The IP protocol.
* `disk_provisioning` - (Optional) The disk provisioning. If set, all the disks in the deployed OVF will have 
   the same specified disk type (accepted values {thin, flat, thick, sameAsSource}).
* `deployment_option` - (Optional) The key of the chosen deployment option. If empty, the default option is chosen.   
* `ovf_network_map` - (Optional) The mapping of name of network identifiers from the ovf descriptor to network UUID in the 
   VI infrastructure.
* `allow_unverified_ssl_cert` - (Optional) Allow unverified ssl certificates while deploying ovf/ova from url.
   Defaults true.   

### Using vApp properties to supply OVF/OVA configuration

Alternative to the settings in `customize`, one can use the settings in the
`properties` section of the `vapp` block to supply configuration parameters to
a virtual machine cloned from a template that came from an imported OVF or OVA
file. Both GuestInfo and ISO transport methods are supported. For templates
that use ISO transport, a CDROM backed by client device is required. See [CDROM
options](#cdrom-options) for details. 

~> **NOTE:** The only supported usage path for vApp properties is for existing
user-configurable keys. These generally come from an existing template that was
created from an imported OVF or OVA file. You cannot set values for vApp
properties on virtual machines created from scratch, virtual machines lacking a
vApp configuration, or on property keys that do not exist.

The configuration looks similar to the one below:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template_from_ovf.id}"
  }

  vapp {
    properties = {
      guestinfo.tf.internal.id = "42"
    }
  }
}
```

### Additional requirements and notes for cloning

Note that when cloning from a template, there are additional requirements in
both the resource configuration and source template:

* The virtual machine must not be powered on at the time of cloning.
* All disks on the virtual machine must be SCSI disks.
* You must specify at least the same number of `disk` devices as there are
  disks that exist in the template. These devices are ordered and lined up by
  the `unit_number` attribute. Additional disks can be added past this.
* The `size` of a virtual disk must be at least the same size as its
  counterpart disk in the template.
* When using `linked_clone`, the `size`, `thin_provisioned`, and
  `eagerly_scrub` settings for each disk must be an exact match to the
  individual disk's counterpart in the source template.
* The storage controller count settings should be
  configured as necessary to cover all of the disks on the template. For best
  results, only configure this setting for the amount of controllers you will
  need to cover your disk quantity and bandwidth needs, and configure your
  template accordingly. For most workloads, this setting should be kept at its
  default of `1` SCSI controller, and all disks in the template should reside 
  on the single, primary controller.
* Some operating systems (such as Windows) do not respond well to a change in
  disk controller type, so when using such OSes, take care to ensure that
  `scsi_type` is set to an exact match of the template's controller set. For
  maximum compatibility, make sure the SCSI controllers on the source template
  are all the same type.

To ease the gathering of some of these options, you can use the
[`vsphere_virtual_machine` data source][tf-vsphere-virtual-machine-ds], which
will give you disk attributes, network interface types, SCSI bus types, and
also the guest ID of the source template.  See the [cloning and customization
example](#cloning-and-customization-example) for usage details.

## Virtual Machine Migration

The `vsphere_virtual_machine` resource supports live migration (otherwise known
as vMotion) both on the host and storage level. One can migrate the entire VM
to another host, cluster, resource pool, or datastore, and migrate or pin a
single disk to a specific datastore.

### Host, cluster, and resource pool migration 

To migrate the virtual machine to another host or resource pool, change the
`host_system_id` or `resource_pool_id` to the manged object IDs of the new host
or resource pool accordingly. To change the virtual machine's cluster or
standalone host, select a resource pool within the specific target.

The same rules apply for migration as they do for VM creation - any host
specified needs to be a part of the resource pool supplied. Also keep in mind
the implications of moving the virtual machine to a resource pool in another
cluster or standalone host, namely ensuring that all hosts in the cluster (or
the single standalone host) have access to the datastore that the virtual
machine is in.

### Storage migration

Storage migration can be done on two levels:

* Global datastore migration can be handled by changing the global
  `datastore_id` attribute. This triggers a storage migration for all disks
  that do not have an explicit `datastore_id` specified.
* When using Storage DRS through the `datastore_cluster_id` attribute, the
  entire virtual machine can be migrated from one datastore cluster to another
  by changing the value of this setting. In addition, when
  `datastore_cluster_id` is in use, any disks that drift to datastores outside
  of the datastore cluster via such actions as manual modification will be
  migrated back to the datastore cluster on the next apply.
* An individual `disk` device can be migrated by manually specifying the
  `datastore_id` in its configuration block. This also pins it to the specific
  datastore that is specified - if at a later time the VM and any unpinned
  disks migrate to another host, the disk will stay on the specified datastore.

An example of datastore pinning is below. As long as the datastore in the
`pinned_datastore` data source does not change, any change to the standard
`vm_datastore` data source will not affect the data disk - the disk will stay
where it is.

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  datastore_id = "${data.vsphere_datastore.vm_datastore.id}"

  disk {
    label = "disk0"
    size  = 10
  }

  disk {
    datastore_id = "${data.vsphere_datastore.pinned_datastore.id}"
    label        = "disk1"
    size         = 100
    unit_number  = 1
  }

  # ... other configuration ...
}
```

#### Storage migration restrictions

Note that you cannot migrate external disks added with the `attach` parameter.
As these disks have usually been created and assigned to a datastore outside of
the scope of the `vsphere_virtual_machine` resource in question, such as by
using the [`vsphere_virtual_disk` resource][tf-vsphere-virtual-disk],
management of such disks would render their configuration unstable.

[tf-vsphere-virtual-disk]: /docs/providers/vsphere/r/virtual_disk.html

## Attribute Reference

The following attributes are exported on the base level of this resource:

* `id` - The UUID of the virtual machine.
* `reboot_required` - Value internal to Terraform used to determine if a
  configuration set change requires a reboot. This value is only useful during
  an update process and gets reset on refresh.
* `vmware_tools_status` - The state of VMware tools in the guest. This will
  determine the proper course of action for some device operations.
* `vmx_path` - The path of the virtual machine's configuration file in the VM's
  datastore.
* `imported` - This is flagged if the virtual machine has been imported, or the
  state has been migrated from a previous version of the resource. It
  influences the behavior of the first post-import apply operation. See the
  section on [importing](#importing) below.
* `change_version` - A unique identifier for a given version of the last
  configuration applied, such the timestamp of the last update to the
  configuration.
* `uuid` - The UUID of the virtual machine. Also exposed as the `id` of the
  resource.
* `default_ip_address` - The IP address selected by Terraform to be used with
  any [provisioners][tf-docs-provisioners] configured on this resource.
  Whenever possible, this is the first IPv4 address that is reachable through
  the default gateway configured on the machine, then the first reachable IPv6
  address, and then the first general discovered address if neither exist. If
  VMware tools is not running on the virtual machine, or if the VM is powered
  off, this value will be blank.
* `guest_ip_addresses` - The current list of IP addresses on this machine,
  including the value of `default_ip_address`. If VMware tools is not running
  on the virtual machine, or if the VM is powered off, this list will be empty.
* `moid`: The [managed object reference ID][docs-about-morefs] of the created
  virtual machine.
* `vapp_transport` - Computed value which is only valid for cloned virtual
  machines. A list of vApp transport methods supported by the source virtual
  machine or template.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

## Importing 

An existing virtual machine can be [imported][docs-import] into this resource
via supplying the full path to the virtual machine. An example is below:

[docs-import]: /docs/import/index.html

```
terraform import vsphere_virtual_machine.vm /dc1/vm/srv1
```

The above would import the virtual machine named `srv1` that is located in the
`dc1` datacenter.

### Additional requirements and notes for importing

Many of the same requirements for
[cloning](#additional-requirements-and-notes-for-cloning) apply to importing,
although since importing writes directly to state, a lot of these rules cannot
be enforced at import time, so every effort should be made to ensure the
correctness of the configuration before the import.

In addition to these rules, the following extra rules apply to importing:

* Disks need to have their [`label`](#label) argument assigned in a convention
  matching `diskN`, starting with disk number 0, based on each disk's order on
  the SCSI bus. As an example, a disk on SCSI controller 0 with a unit number
  of 0 would be labeled `disk0`, a disk on the same controller with a unit
  number of 1 would be `disk1`, but the next disk, which is on SCSI controller
  1 with a unit number of 0, still becomes `disk2`.
* Disks always get imported with [`keep_on_remove`](#keep_on_remove) enabled
  until the first `terraform apply` runs, which will remove the setting for
  known disks. This is an extra safeguard against naming or accounting mistakes
  in the disk configuration.
* The storage controller count for the resource is set to the number of 
  contiguous storage controllers found, starting with the controller at bus 
  number 0. If no storage controllers are found, the VM is not eligible for 
  import. To ensure maximum compatibility, make sure your virtual machine has
  the exact number of storage controllers it needs, and set the storage
   controller counts accordingly.

After importing, you should run `terraform plan`. Unless you have changed
anything else in configuration that would be causing other attributes to
change, the only difference should be configuration-only changes, usually
comprising of:

* The [`imported`](#imported) flag will transition from `true` to `false`.
* [`keep_on_remove`](#keep_on_remove) of known disks will transition from
  `true` to `false`. 
* Configuration supplied in the [`clone`](#clone) block, if present, will be
  persisted to state. This initial persistence operation does not perform any
  cloning or customization actions, nor does it force a new resource. After the
  first apply operation, further changes to `clone` will force a new resource
  as per normal operation.

~> **NOTE:** Further to the above, do not make any configuration changes to
`clone` after importing or upgrading from a legacy version of the provider
before doing an initial `terraform apply` as these changes will not correctly
force a new resource, and your changes will have persisted to state, preventing
further plans from correctly triggering a diff.

These changes only update Terraform state when applied, hence it is safe to run
when the virtual machine is running. If more settings are being modified, you
may need to plan maintenance accordingly for any necessary re-configuration of
the virtual machine.
