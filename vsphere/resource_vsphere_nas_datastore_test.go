package vsphere

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/datastore"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereNasDatastore_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
			{
				Config:            testAccResourceVSphereNasDatastoreConfigBasic(),
				ImportState:       true,
				ResourceName:      "vsphere_nas_datastore.datastore",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_multiHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigMultiHost(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_basicToMultiHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
			{
				Config:      testAccResourceVSphereNasDatastoreConfigMultiHost(),
				ExpectError: expectErrorIfNotVirtualCenter(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_multiHostToBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigMultiHost(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_renameDatastore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasicAltName(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreHasName(fmt.Sprintf("%s-renamed", os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_inFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
			// NOTE: This test can't run on ESXi without giving a "dangling
			// resource" error during testing - "move to folder after" hits the
			// error on the same path of the call stack that triggers an error in
			// both create and update and should provide adequate coverage
			// barring manual testing.
			testAccSkipIfEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasicFolder(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreMatchInventoryPath(os.Getenv("TF_VAR_VSPHERE_DS_FOLDER")),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_moveToFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
			{
				Config:      testAccResourceVSphereNasDatastoreConfigBasicFolder(),
				ExpectError: expectErrorIfNotVirtualCenter(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreMatchInventoryPath(os.Getenv("TF_VAR_VSPHERE_DS_FOLDER")),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_inDatastoreCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigDatastoreCluster(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreMatchInventoryPath(testAccResourceVSphereDatastoreClusterNameStandard),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_moveToDatastoreCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
				),
			},
			{
				Config: testAccResourceVSphereNasDatastoreConfigDatastoreCluster(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreMatchInventoryPath(testAccResourceVSphereDatastoreClusterNameStandard),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasicTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereDatastoreCheckTags("vsphere_nas_datastore.datastore", "testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_modifyTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigBasicTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereDatastoreCheckTags("vsphere_nas_datastore.datastore", "testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereNasDatastoreConfigMultiTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereDatastoreCheckTags("vsphere_nas_datastore.datastore", "testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_singleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreHasCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereNasDatastore_multiCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereNasDatastorePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereNasDatastoreExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNasDatastoreConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreHasCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereNasDatastoreConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereNasDatastoreExists(true),
					testAccResourceVSphereNasDatastoreHasCustomAttributes(),
				),
			},
		},
	})
}

func testAccResourceVSphereNasDatastorePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI_HOST to run vsphere_vmfs_disks acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NAS_HOST") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAS_HOST to run vsphere_nas_datastore acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_PATH") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_PATH to run vsphere_nas_datastore acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_DS_FOLDER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DS_FOLDER to run vsphere_nas_datastore acceptance tests")
	}
}

func testAccResourceVSphereNasDatastoreExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, err := testGetDatastore(s, "vsphere_nas_datastore.datastore")
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected datastore %q to be missing", ds.Reference().Value)
		}
		return nil
	}
}

func testAccResourceVSphereNasDatastoreHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, err := testGetDatastore(s, "vsphere_nas_datastore.datastore")
		if err != nil {
			return err
		}

		props, err := datastore.Properties(ds)
		if err != nil {
			return err
		}

		actual := props.Summary.Name
		if expected != actual {
			return fmt.Errorf("expected datastore name to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereNasDatastoreMatchInventoryPath(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, err := testGetDatastore(s, "vsphere_nas_datastore.datastore")
		if err != nil {
			return err
		}

		expected, err = folder.RootPathParticleDatastore.PathFromNewRoot(ds.InventoryPath, folder.RootPathParticleDatastore, expected)
		actual := path.Dir(ds.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereNasDatastoreHasCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDatastoreProperties(s, "nas", "datastore")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_nas_datastore", "datastore", props.Entity())
	}
}

func testAccResourceVSphereNasDatastoreConfigBasic() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
		),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}

func testAccResourceVSphereNasDatastoreConfigMultiHost() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

%s 

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"),
	)
}

func testAccResourceVSphereNasDatastoreConfigBasicAltName() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s-renamed"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"), os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}

func testAccResourceVSphereNasDatastoreConfigBasicFolder() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

variable "folder" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_folder" "folder" {
  path          = var.folder
  type          = "datastore"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]
  folder          = vsphere_folder.folder.path

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		os.Getenv("TF_VAR_VSPHERE_DS_FOLDER"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}

func testAccResourceVSphereNasDatastoreConfigBasicTags() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datastore",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"

  tags = ["${vsphere_tag.testacc-tag.id}"]
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}

func testAccResourceVSphereNasDatastoreConfigMultiTags() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datastore",
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

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"

  tags = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}

func testAccResourceVSphereNasDatastoreConfigSingleCustomAttribute() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Datastore"
}

locals {
  nas_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"

  custom_attributes = "${local.nas_attrs}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"), os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}

func testAccResourceVSphereNasDatastoreConfigMultiCustomAttributes() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Datastore"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "Datastore"
}

locals {
  nas_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-2.id}" = "value-2"
  }
}

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"

  custom_attributes = "${local.nas_attrs}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"), os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"),
	)
}

func testAccResourceVSphereNasDatastoreConfigDatastoreCluster() string {
	return fmt.Sprintf(`
variable "nfs_host" {
  type    = string
  default = "%s"
}

variable "nfs_path" {
  type    = string
  default = "%s"
}

variable "folder" {
  type    = string
  default = "%s"
}

%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_nas_datastore" "datastore" {
  name            = "%s"
  host_system_ids = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]
  datastore_cluster_id = "${vsphere_datastore_cluster.datastore_cluster.id}"

  type         = "NFS"
  remote_hosts = ["${var.nfs_host}"]
  remote_path  = "${var.nfs_path}"
}
`, os.Getenv("TF_VAR_VSPHERE_NAS_HOST"),
		os.Getenv("TF_VAR_VSPHERE_NFS_PATH2"),
		os.Getenv("TF_VAR_VSPHERE_DS_FOLDER"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME2"))
}
