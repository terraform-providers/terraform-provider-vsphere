package vsphere

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceVSphereTagCategory_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagCategoryExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagCategoryConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
					testAccResourceVSphereTagCategoryHasName("testacc-category"),
					testAccResourceVSphereTagCategoryHasCardinality(vSphereTagCategoryCardinalitySingle),
					testAccResourceVSphereTagCategoryHasTypes([]string{
						vSphereTagTypeVirtualMachine,
					}),
				),
			},
			{
				ResourceName:      "vsphere_tag_category.testacc-category",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cat, err := testGetTagCategory(s, "testacc-category")
					if err != nil {
						return "", err
					}
					return cat.Name, nil
				},
				Config: testAccResourceVSphereTagCategoryConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereTagCategory_addType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagCategoryExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagCategoryConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
					testAccResourceVSphereTagCategoryHasTypes([]string{
						vSphereTagTypeVirtualMachine,
					}),
				),
			},
			{
				Config: testAccResourceVSphereTagCategoryConfigMultiType,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
					testAccResourceVSphereTagCategoryHasTypes([]string{
						vSphereTagTypeVirtualMachine,
						vSphereTagTypeDatastore,
					}),
				),
			},
		},
	})
}

func TestAccResourceVSphereTagCategory_removeTypeShouldError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagCategoryExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagCategoryConfigMultiType,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
				),
			},
			{
				Config:      testAccResourceVSphereTagCategoryConfigBasic,
				ExpectError: regexp.MustCompile("removal of associable types is not supported"),
			},
		},
	})
}

func TestAccResourceVSphereTagCategory_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagCategoryExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagCategoryConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
				),
			},
			{
				Config: testAccResourceVSphereTagCategoryConfigAltName,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
					testAccResourceVSphereTagCategoryHasName("testacc-category-renamed"),
				),
			},
		},
	})
}

func TestAccResourceVSphereTagCategory_singleCardinality(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagCategoryExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagCategoryConfigSingleCardinality,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
					testAccResourceVSphereTagCategoryHasCardinality(vSphereTagCategoryCardinalitySingle),
				),
			},
		},
	})
}

func TestAccResourceVSphereTagCategory_multiCardinality(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereTagCategoryExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTagCategoryConfigMultiCardinality,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTagCategoryExists(true),
					testAccResourceVSphereTagCategoryHasCardinality(vSphereTagCategoryCardinalityMultiple),
				),
			},
		},
	})
}

func testAccResourceVSphereTagCategoryExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetTagCategory(s, "testacc-category")
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected tag category to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereTagCategoryHasTypes(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cat, err := testGetTagCategory(s, "testacc-category")
		if err != nil {
			return err
		}
		// We use a *schema.Set for types, so the list may not be in the order that
		// we expect it in. Sort both lists to make sure.
		actual := cat.AssociableTypes
		sort.Strings(expected)
		sort.Strings(actual)
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected types list to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereTagCategoryHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cat, err := testGetTagCategory(s, "testacc-category")
		if err != nil {
			return err
		}
		actual := cat.Name
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereTagCategoryHasCardinality(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cat, err := testGetTagCategory(s, "testacc-category")
		if err != nil {
			return err
		}
		actual := cat.Cardinality
		if expected != actual {
			return fmt.Errorf("expected cardinality to be %q, got %q", expected, actual)
		}
		return nil
	}
}

const testAccResourceVSphereTagCategoryConfigBasic = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  description = "Managed by Terraform"
  cardinality = "SINGLE"

  associable_types = [
    "VirtualMachine",
  ]
}
`

const testAccResourceVSphereTagCategoryConfigMultiType = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  description = "Managed by Terraform"
  cardinality = "SINGLE"

  associable_types = [
    "VirtualMachine",
    "Datastore",
  ]
}
`

const testAccResourceVSphereTagCategoryConfigMultiCardinality = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  description = "Managed by Terraform"
  cardinality = "MULTIPLE"

  associable_types = [
    "VirtualMachine",
  ]
}
`

const testAccResourceVSphereTagCategoryConfigAltName = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category-renamed"
  description = "Managed by Terraform"
  cardinality = "MULTIPLE"

  associable_types = [
    "VirtualMachine",
  ]
}
`

const testAccResourceVSphereTagCategoryConfigSingleCardinality = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-category"
  description = "Managed by Terraform"
  cardinality = "SINGLE"

  associable_types = [
    "VirtualMachine",
  ]
}
`
