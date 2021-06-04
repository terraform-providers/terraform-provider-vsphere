# vSphere Provider for Terraform [![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/hashicorp/terraform-provider-vsphere?label=release)](https://github.com/hashicorp/terraform-provider-vsphere/releases) [![license](https://img.shields.io/github/license/hashicorp/terraform-provider-vsphere.svg)]()


<a href="https://terraform.io">
    <img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" alt="Terraform logo" title="Terrafpr," align="right" height="50" />
</a>

* [Getting Started & Documentation](https://www.terraform.io/docs/providers/vsphere/index.html)
* Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)


This is the repository for the vSphere Provider for Terraform, which one can use
with Terraform to work with VMware vSphere Products, notably [vCenter
Server][vmware-vcenter] and [ESXi][vmware-esxi].

[vmware-vcenter]: https://www.vmware.com/products/vcenter-server.html
[vmware-esxi]: https://www.vmware.com/products/esxi-and-esx.html

For general information about Terraform, visit the [official
website][tf-website] and the [GitHub project page][tf-github].

[tf-website]: https://terraform.io/
[tf-github]: https://github.com/hashicorp/terraform

This provider plugin is maintained by the Terraform team at [HashiCorp](https://www.hashicorp.com/).

## Requirements
-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
- vSphere version    
   -  This provider supports vSphere versions in accordance with the [VMware Product Lifecycle Matrix](https://lifecycle.vmware.com/#/), from General Availability until the End of General Support 
-	[Go](https://golang.org/doc/install) 1.16.x (to build the provider plugin)

## Building The Provider

Unless you are [contributing](_about/CONTRIBUTING.md) to the provider or require a
pre-release bugfix or feature, you will want to use an [officially released](https://github.com/hashicorp/terraform-provider-vsphere/releases)
version of the provider.


## Contributing to the provider

The vSphere Provider for Terraform is the work of many contributors. We appreciate your help!

To contribute, please read the [contribution guidelines](_about/CONTRIBUTING.md). You may also [report an issue](https://github.com/hashicorp/terraform-provider-vsphere/issues/new/choose). Once you've filed an issue, it will follow the [issue lifecycle](_about/ISSUES.md).

Also available are some answers to [Frequently Asked Questions](_about/FAQ.md).


