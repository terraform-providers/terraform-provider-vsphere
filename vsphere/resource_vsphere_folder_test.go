package vsphere

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/object"
)

const testAccResourceVSphereFolderConfigExpectedName = "testacc-folder"
const testAccResourceVSphereFolderConfigExpectedAltName = "terraform-renamed-folder"
const testAccResourceVSphereFolderConfigExpectedParentName = "terraform-test-parent"
const testAccResourceVSphereFolderConfigOOBName = "terraform-test-oob"

func TestAccResourceVSphereFolder_vmFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
			{
				ResourceName:      "vsphere_folder.folder",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					testFolder, err := testGetFolder(s, "folder")
					if err != nil {
						return "", err
					}
					return testFolder.InventoryPath, nil
				},
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_datastoreFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeDatastore,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeDatastore),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_networkFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeNetwork,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeNetwork),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_hostFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeHost,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeHost),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_datacenterFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeDatacenter,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeDatacenter),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedAltName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedAltName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_subfolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigSubFolder(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasParent(false, testAccResourceVSphereFolderConfigExpectedParentName),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_moveToSubfolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasParent(true, "vm"),
				),
			},
			{
				Config: testAccResourceVSphereFolderConfigSubFolder(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasParent(false, testAccResourceVSphereFolderConfigExpectedParentName),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_modifyTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereFolderConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_modifyTagsMultiStage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
			{
				Config: testAccResourceVSphereFolderConfigAllTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
			{
				Config: testAccResourceVSphereFolderConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_customAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_modifyCustomAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereFolderMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_removeAllCustomAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereFolderRemovedCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					testAccResourceVSphereFolderHasCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereFolder_preventDeleteIfNotEmpty(t *testing.T) {
	var s *terraform.State

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereFolderExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereFolderExists(true),
					testAccResourceVSphereFolderHasName(testAccResourceVSphereFolderConfigExpectedName),
					testAccResourceVSphereFolderHasType(folder.VSphereFolderTypeVM),
					copyStatePtr(&s),
				),
			},
			{
				PreConfig: func() {
					if err := testAccResourceVSphereFolderCreateOOB(s); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
				Destroy:     true,
				ExpectError: regexp.MustCompile("folder is not empty, please remove all items before deleting"),
			},
			{
				PreConfig: func() {
					if err := testAccResourceVSphereFolderDeleteOOB(s); err != nil {
						panic(err)
					}
				},
				Config: testAccResourceVSphereFolderConfigBasic(
					testAccResourceVSphereFolderConfigExpectedName,
					folder.VSphereFolderTypeVM,
				),
			},
		},
	})
}

func testAccResourceVSphereFolderExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		testFolder, err := testGetFolder(s, "folder")
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected folder %q to be missing", testFolder.Reference().Value)
		}
		return nil
	}
}

func testAccResourceVSphereFolderHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetFolderProperties(s, "folder")
		if err != nil {
			return err
		}
		actual := props.Name
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereFolderHasType(expected folder.VSphereFolderType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		f, err := testGetFolder(s, "folder")
		if err != nil {
			return err
		}
		actual, err := folder.FindType(f)
		if err != nil {
			return err
		}
		if expected != actual {
			return fmt.Errorf("expected type to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereFolderHasParent(expectedRoot bool, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetFolderProperties(s, "folder")
		if err != nil {
			return err
		}
		if props.Parent.Type != "Folder" && !expectedRoot {
			return fmt.Errorf("folder %q is a root folder", props.Name)
		}
		client := testAccProvider.Meta().(*Client).vimClient
		pfolder, err := folder.FromID(client, props.Parent.Value)
		if err != nil {
			return err
		}
		pprops, err := folder.Properties(pfolder)
		if err != nil {
			return err
		}

		actual := pprops.Name
		if expectedName != actual {
			return fmt.Errorf("expected parent folder name to be %q, got %q", expectedName, actual)
		}
		return nil
	}
}

// testAccResourceVSphereFolderCheckTags is a check to ensure that any
// tags that have been created with the supplied resource name have been
// attached to the folder.
func testAccResourceVSphereFolderCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		testFolder, err := testGetFolder(s, "folder")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, testFolder, tagResName)
	}
}

// testAccResourceVSphereFolderCheckNoTags is a check to ensure that a folder
// has no tags on it. This is used by the vsphere_tag tests specifically to
// test to make sure that complete tag removal is explicitly working without
// having to rely on the simple empty diff test after the final step.
func testAccResourceVSphereFolderCheckNoTags() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		testFolder, err := testGetFolder(s, "folder")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasNoTags(tagsClient, testFolder)
	}
}

func testAccResourceVSphereFolderHasCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetFolderProperties(s, "folder")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_folder", "folder", props.Entity())
	}
}

// testAccResourceVSphereFolderCreateOOB creates an out-of-band folder that is
// not tracked by TF. This is used in deletion checks to make sure we don't
// perform unsafe recursive deletions.
func testAccResourceVSphereFolderCreateOOB(s *terraform.State) error {
	testFolder, err := testGetFolder(s, "folder")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	if _, err := testFolder.CreateFolder(ctx, testAccResourceVSphereFolderConfigOOBName); err != nil {
		return err
	}
	return nil
}

// testAccResourceVSphereFolderDeleteOOB wipes any child items in the test
// folder resource. This is used to reverse the actions of
// testAccResourceVSphereFolderCreateOOB so we can properly clean up the test.
func testAccResourceVSphereFolderDeleteOOB(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).vimClient
	testFolder, err := testGetFolder(s, "folder")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	refs, err := testFolder.Children(ctx)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		me := object.NewCommon(client.Client, ref.Reference())
		dctx, dcancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer dcancel()
		task, err := me.Destroy(dctx)
		if err != nil {
			return err
		}
		tctx, tcancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer tcancel()
		if err := task.Wait(tctx); err != nil {
			return err
		}
	}
	return nil
}

func testAccResourceVSphereFolderConfigBasic(name string, ft folder.VSphereFolderType) string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		name,
		ft,
	)
}

func testAccResourceVSphereFolderConfigSubFolder(name string, ft folder.VSphereFolderType) string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

variable "parent_name" {
  default = "%s"
}

resource "vsphere_folder" "parent" {
  path          = "${var.parent_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_folder" "folder" {
  path          = "${vsphere_folder.parent.path}/${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		name,
		ft,
		testAccResourceVSphereFolderConfigExpectedParentName,
	)
}

func testAccResourceVSphereFolderConfigTag() string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "Folder",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  tags          = ["${vsphere_tag.testacc-tag.id}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testAccResourceVSphereFolderConfigExpectedName,
		folder.VSphereFolderTypeVM,
	)
}

func testAccResourceVSphereFolderConfigAllTag() string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

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
    "Folder",
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

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  tags          = ["${vsphere_tag.testacc-tag.id}", "${vsphere_tag.testacc-tags-alt.0.id}", "${vsphere_tag.testacc-tags-alt.1.id}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testAccResourceVSphereFolderConfigExpectedName,
		folder.VSphereFolderTypeVM,
	)
}

func testAccResourceVSphereFolderConfigMultiTag() string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

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
    "Folder",
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

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  tags          = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testAccResourceVSphereFolderConfigExpectedName,
		folder.VSphereFolderTypeVM,
	)
}

func testAccResourceVSphereFolderCustomAttribute() string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Folder"
}

locals {
  folder_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  custom_attributes = "${local.folder_attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testAccResourceVSphereFolderConfigExpectedName,
		folder.VSphereFolderTypeVM,
	)
}

func testAccResourceVSphereFolderMultiCustomAttributes() string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Folder"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "Folder"
}

locals {
  folder_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-2.id}" = "value-2"
  }
}

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  custom_attributes = "${local.folder_attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testAccResourceVSphereFolderConfigExpectedName,
		folder.VSphereFolderTypeVM,
	)
}

func testAccResourceVSphereFolderRemovedCustomAttributes() string {
	return fmt.Sprintf(`
%s

variable "folder_name" {
  default = "%s"
}

variable "folder_type" {
  default = "%s"
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Folder"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "Folder"
}

resource "vsphere_folder" "folder" {
  path          = "${var.folder_name}"
  type          = "${var.folder_type}"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testAccResourceVSphereFolderConfigExpectedName,
		folder.VSphereFolderTypeVM,
	)
}
