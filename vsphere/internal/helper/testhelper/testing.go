package testhelper

import (
	"fmt"
	"os"
)

func CombineConfigs(configs ...string) string {
	var config string
	for _, c := range configs {
		config = fmt.Sprintf("%s\n%s", config, c)
	}
	return config
}

func ConfigDataDC1() string {
	return `
data "vsphere_datacenter" "dc1" {
  name = vsphere_datacenter.dc1.name
}
`
}

func ConfigDataDC2() string {
	return `
data "vsphere_datacenter" "dc2" {
  name = vsphere_datacenter.dc2.name
}
`
}

func ConfigResDC1() string {
	return `
resource "vsphere_datacenter" "dc1" {
  name = "testacc-dc1"
  tags = [vsphere_tag.tag1.id, vsphere_tag.tag2.id]
}
`
}

func ConfigResDC2() string {
	return `
resource "vsphere_datacenter" "dc2" {
  name = "testacc-dc2"
  tags = [ vsphere_tag.tag1.id, vsphere_tag.tag3.id ]
}
`
}

func ConfigResTagCat1() string {
	return `
resource "vsphere_tag_category" "category1" {
  name        = "testacc-cat1"
  description = "cat1"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datacenter"
  ]
}
`
}

func ConfigResTagCat2() string {
	return `
resource "vsphere_tag_category" "category2" {
  name        = "testacc-cat2"
  description = "cat2"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datacenter"
  ]
}
`
}

func ConfigDataRootDS1() string {
	return fmt.Sprintf(`
data "vsphere_datastore" "rootds1"{
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = "%s"
}
`, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME"))
}

func ConfigResTag1() string {
	return `
	resource "vsphere_tag" "tag1" {
	  name        = "testacc-tag1"
	  category_id = vsphere_tag_category.category1.id
	}
	`
}

func ConfigResTag2() string {
	return `
	resource "vsphere_tag" "tag2" {
	  name        = "testacc-tag2"
	  category_id = vsphere_tag_category.category2.id
	}
	`
}

func ConfigResTag3() string {
	return `
	resource "vsphere_tag" "tag3" {
	  name        = "testacc-tag2"
	  category_id = vsphere_tag_category.category1.id
	}`
}

func ConfigDataRootDC1() string {
	return fmt.Sprintf(`
data "vsphere_datacenter" "rootdc1" {
  name = "%s"
}
`, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
}

func ConfigDataRootHost1() string {
	return fmt.Sprintf(`
data "vsphere_host" "roothost1" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func ConfigDataRootHost2() string {
	return fmt.Sprintf(`
data "vsphere_host" "roothost2" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, os.Getenv("TF_VAR_VSPHERE_ESXI2"))
}

func ConfigDataRootHost3() string {
	return fmt.Sprintf(`
data "vsphere_host" "roothost3" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}

func ConfigDataRootHost4() string {
	return fmt.Sprintf(`
data "vsphere_host" "roothost4" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, os.Getenv("TF_VAR_VSPHERE_ESXI4"))
}

func ConfigResDS1() string {
	return fmt.Sprintf(`
resource "vsphere_nas_datastore" "ds1" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]
  type            = "NFS"
  remote_hosts    = ["%s"]
  remote_path     = "%s"
}
`, os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"), os.Getenv("TF_VAR_VSPHERE_NAS_HOST"), os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"))
}

func ConfigDataRootComputeCluster1() string {
	return fmt.Sprintf(`
data "vsphere_compute_cluster" "rootcompute_cluster1" {
	name          = "%s"
	datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, os.Getenv("TF_VAR_VSPHERE_CLUSTER"))
}

func ConfigResResourcePool1() string {
	return `
	resource "vsphere_resource_pool" "pool1" {
	  name                    = "testacc-resource-pool1"
	  parent_resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
	}
	`
}

func ConfigDataRootPortGroup1() string {
	return fmt.Sprintf(`
data "vsphere_network" "network1" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, os.Getenv("TF_VAR_VSPHERE_PG_NAME"))
}

func ConfigDataRootVMNet() string {
	return `
	data "vsphere_network" "vmnet" {
	  name          = "VM Network"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}
	`
}
