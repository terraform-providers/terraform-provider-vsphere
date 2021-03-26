package vsphere

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceVSphereCustomAttribute_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereCustomAttributeExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereCustomAttributeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
					testAccResourceVSphereCustomAttributeHasName("testacc-attribute"),
					testAccResourceVSphereCustomAttributeHasType(""),
				),
			},
			{
				ResourceName:      "vsphere_custom_attribute.testacc-attribute",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					attr, err := testGetCustomAttribute(s, "testacc-attribute")
					if err != nil {
						return "", err
					}
					if attr == nil {
						return "", errors.New("custom attribute does not exist")
					}
					return attr.Name, nil
				},
				Config: testAccResourceVSphereCustomAttributeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
				),
			},
		},
	})
}
func TestAccResourceVSphereCustomAttribute_withType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereCustomAttributeExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereCustomAttributeConfigType,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
					testAccResourceVSphereCustomAttributeHasName("testacc-attribute"),
					testAccResourceVSphereCustomAttributeHasType("VirtualMachine"),
				),
			},
		},
	})
}
func TestAccResourceVSphereCustomAttribute_rename(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereCustomAttributeExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereCustomAttributeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
				),
			},
			{
				Config: testAccResourceVSphereCustomAttributeConfigAltName,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
					testAccResourceVSphereCustomAttributeHasName("testacc-attribute-renamed"),
				),
			},
		},
	})
}

func TestAccResourceVSphereCustomAttribute_changeType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereCustomAttributeExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereCustomAttributeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
					testAccResourceVSphereCustomAttributeHasType(""),
				),
			},
			{
				Config: testAccResourceVSphereCustomAttributeConfigType,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereCustomAttributeExists(true),
					testAccResourceVSphereCustomAttributeHasType("VirtualMachine"),
				),
			},
		},
	})
}

func testAccResourceVSphereCustomAttributeExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attr, err := testGetCustomAttribute(s, "testacc-attribute")
		if err != nil {
			return err
		}
		if attr == nil && expected {
			return errors.New("expected custom attribute to exist")
		} else if attr != nil && !expected {
			return errors.New("expected custom attribute to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereCustomAttributeHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attr, err := testGetCustomAttribute(s, "testacc-attribute")
		if err != nil {
			return err
		}
		actual := attr.Name
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereCustomAttributeHasType(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attr, err := testGetCustomAttribute(s, "testacc-attribute")
		if err != nil {
			return err
		}
		actual := attr.ManagedObjectType
		if expected != actual {
			return fmt.Errorf("expected managed object type to be %q, got %q", expected, actual)
		}
		return nil
	}
}

const testAccResourceVSphereCustomAttributeConfigBasic = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name = "testacc-attribute"
}
`

const testAccResourceVSphereCustomAttributeConfigType = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "VirtualMachine"
}
`

const testAccResourceVSphereCustomAttributeConfigAltName = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name = "testacc-attribute-renamed"
}
`
