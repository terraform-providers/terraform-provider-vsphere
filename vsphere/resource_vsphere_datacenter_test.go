package vsphere

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/govmomi/find"
)

const testAccCheckVSphereDatacenterResourceName = "vsphere_datacenter.testDC"

const testAccCheckVSphereDatacenterConfig = `
resource "vsphere_datacenter" "testDC" {
  name = "testDC"
}
`

const testAccCheckVSphereDatacenterConfigSubfolder = `
resource "vsphere_folder" "folder" {
  path = "%s"
  type = "datacenter"
}

resource "vsphere_datacenter" "testDC" {
  name   = "testDC"
  folder = vsphere_folder.folder.path
}
`

const testAccCheckVSphereDatacenterConfigTags = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datacenter",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_datacenter" "testDC" {
  name = "testDC"
  tags = ["${vsphere_tag.testacc-tag.id}"]
}
`

const testAccCheckVSphereDatacenterConfigMultiTags = `
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
    "Datacenter",
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

resource "vsphere_datacenter" "testDC" {
  name = "testDC"
  tags = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`

const testAccCheckVSphereDatacenterConfigCustomAttributes = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Datacenter"
}

locals {
  dc_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_datacenter" "testDC" {
  name = "testDC"
  custom_attributes = "${local.dc_attrs}"
}
`
const testAccCheckVSphereDatacenterConfigMultiCustomAttributes = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Datacenter"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "Datacenter"
}

locals {
  dc_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-2.id}" = "value-2"
  }
}

resource "vsphere_datacenter" "testDC" {
  name = "testDC"
  custom_attributes = "${local.dc_attrs}"
}
`

// Create a datacenter on the root folder
func TestAccResourceVSphereDatacenter_createOnRootFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfig,
				Check:  resource.ComposeTestCheckFunc(testAccCheckVSphereDatacenterExists(testAccCheckVSphereDatacenterResourceName, true)),
			},
		},
	})
}

// Create a datacenter on a subfolder
func TestAccResourceVSphereDatacenter_createOnSubfolder(t *testing.T) {
	dcFolder := os.Getenv("TF_VAR_VSPHERE_DC_FOLDER")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckVSphereDatacenterConfigSubfolder, dcFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_modifyTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccCheckVSphereDatacenterConfigMultiTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_singleCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigCustomAttributes,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterHasCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_modifyCustomAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigCustomAttributes,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterHasCustomAttributes(),
				),
			},
			{
				Config: testAccCheckVSphereDatacenterConfigMultiCustomAttributes,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterHasCustomAttributes(),
				),
			},
		},
	})
}

func testAccCheckVSphereDatacenterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).vimClient
	finder := find.NewFinder(client.Client, true)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_datacenter" {
			continue
		}

		path := rs.Primary.Attributes["name"]
		if _, ok := rs.Primary.Attributes["folder"]; ok {
			path = rs.Primary.Attributes["folder"] + "/" + path
		}
		_, err := finder.Datacenter(context.TODO(), path)
		if err != nil {
			switch err.(type) {
			case *find.NotFoundError:
				return nil
			default:
				return err
			}
		} else {
			return fmt.Errorf("datacenter '%s' still exists", path)
		}
	}

	return nil
}

func testAccCheckVSphereDatacenterExists(n string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*Client).vimClient
		finder := find.NewFinder(client.Client, true)

		path := rs.Primary.Attributes["name"]
		if _, ok := rs.Primary.Attributes["folder"]; ok {
			path = rs.Primary.Attributes["folder"] + "/" + path
		}
		_, err := finder.Datacenter(context.TODO(), path)
		if err != nil {
			switch e := err.(type) {
			case *find.NotFoundError:
				if exists {
					return fmt.Errorf("datacenter does not exist: %s", e.Error())
				}
				return nil
			default:
				return err
			}
		}
		return nil
	}
}

// testAccResourceVSphereDatacenterCheckTags is a check to ensure that any tags
// that have been created with supplied resource name have been attached to the
// datacenter.
func testAccResourceVSphereDatacenterCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_datacenter.testDC")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}

		finder := find.NewFinder(vars.client.Client, true)

		path := vars.resourceAttributes["name"]
		if _, ok := vars.resourceAttributes["folder"]; ok {
			path = vars.resourceAttributes["folder"] + "/" + path
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()
		dc, err := finder.Datacenter(ctx, path)
		if err != nil {
			return err
		}

		return testObjectHasTags(s, tagsClient, dc, tagResName)
	}
}

func testAccResourceVSphereDatacenterHasCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ca, err := testGetDatacenterCustomAttributes(s, "testDC")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_datacenter", "testDC", ca.Entity())
	}
}
