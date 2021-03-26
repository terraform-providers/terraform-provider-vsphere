package vsphere

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceVSphereTag_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
					testAccResourceVSphereTagHasName("testacc-tag"),
					testAccResourceVSphereTagHasDescription("Managed by Terraform"),
					testAccResourceVSphereTagHasCategory(),
				),
			},
			{
				ResourceName:      "vsphere_tag.testacc-tag",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cat, err := testGetTagCategory(s, "testacc-category")
					if err != nil {
						return "", err
					}
					tag, err := testGetTag(s, "testacc-tag")
					if err != nil {
						return "", err
					}
					m := make(map[string]string)
					m["category_name"] = cat.Name
					m["tag_name"] = tag.Name
					b, err := json.Marshal(m)
					if err != nil {
						return "", err
					}

					return string(b), nil
				},
				Config: testAccResourceVSphereTagConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereTag_changeName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
				),
			},
			{
				Config: testAccResourceVSphereTagConfigAltName,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
					testAccResourceVSphereTagHasName("testacc-tag-renamed"),
				),
			},
		},
	})
}

func TestAccResourceVSphereTag_changeDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
				),
			},
			{
				Config: testAccResourceVSphereTagConfigAltDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
					testAccResourceVSphereTagHasDescription("Still managed by Terraform"),
				),
			},
		},
	})
}

func TestAccResourceVSphereTag_detachAllTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagConfigOnFolderAttached(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
				),
			},
			{
				Config: testAccResourceVSphereTagConfigOnFolderNotAttached(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagExists(true),
					testAccResourceVSphereFolderCheckNoTags(),
				),
			},
		},
	})
}

func testAccResourceVSphereTagExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetTag(s, "testacc-tag")
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected tag to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereTagHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tag, err := testGetTag(s, "testacc-tag")
		if err != nil {
			return err
		}
		actual := tag.Name
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereTagHasDescription(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tag, err := testGetTag(s, "testacc-tag")
		if err != nil {
			return err
		}
		actual := tag.Description
		if expected != actual {
			return fmt.Errorf("expected description to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereTagHasCategory() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tag, err := testGetTag(s, "testacc-tag")
		if err != nil {
			return err
		}
		category, err := testGetTagCategory(s, "testacc-category")
		if err != nil {
			return err
		}

		expected := category.ID
		actual := tag.CategoryID
		if expected != actual {
			return fmt.Errorf("expected ID to be %q, got %q", expected, actual)
		}
		return nil
	}
}

const testAccResourceVSphereTagConfigBasic = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  cardinality = "SINGLE"

  associable_types = [
    "All",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  description = "Managed by Terraform"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}
`

const testAccResourceVSphereTagConfigAltName = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  cardinality = "SINGLE"

  associable_types = [
    "All",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag-renamed"
  description = "Managed by Terraform"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}
`

const testAccResourceVSphereTagConfigAltDescription = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  cardinality = "SINGLE"

  associable_types = [
    "All",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  description = "Still managed by Terraform"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}
`

func testAccResourceVSphereTagConfigOnFolderAttached() string {
	return fmt.Sprintf(`
  %s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  cardinality = "SINGLE"

  associable_types = [
    "Folder",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  description = "Managed by Terraform"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_folder" "folder" {
  path          = "testacc-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = ["${vsphere_tag.testacc-tag.id}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereTagConfigOnFolderNotAttached() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  cardinality = "SINGLE"

  associable_types = [
    "Folder",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  description = "Managed by Terraform"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_folder" "folder" {
  path          = "testacc-folder"
  type          = "vm"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
