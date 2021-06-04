package vsphere

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/govmomi/license"
)

const testAccVSphereLicenseInvalidConfig = `
resource "vsphere_license" "foo" {
 license_key = "HN422-47193-58V7M-03086-0JAN2"
}
`

func TestAccResourceVSphereLicense_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccVSpherePreLicenseBasicCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereLicenseBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLicenseExists("vsphere_license.foo"),
				),
			},
		},
	})
}

func TestAccResourceVSphereLicense_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereLicenseInvalidConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLicenseNotExists("vsphere_license.foo"),
				),
				ExpectError: regexp.MustCompile("License is not valid for this product"),
			},
		},
	})
}

func TestAccResourceVSphereLicense_withLabelsOnVCenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccVSpherePreLicenseBasicCheck(t)
			testAccVspherePreLicenseESXiServerIsNotSetCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  testAccVSphereLicenseWithLabelConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLicenseWithLabelExists("vsphere_license.foo"),
				),
			},
		},
	})
}

func TestAccResourceVSphereLicense_withLabelsOnESXiServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccVSpherePreLicenseBasicCheck(t)
			testAccSkipIfNotEsxi(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereLicenseWithLabelConfig(),
				// This error will be thrown when ran against an ESXi server.
				ExpectError: regexp.MustCompile("Labels are not allowed in ESXi"),
			},
		},
	})
}

func testAccVspherePreLicenseESXiServerIsNotSetCheck(t *testing.T) {
	key, err := strconv.ParseBool(os.Getenv("TF_VAR_VSPHERE_TEST_ESXI"))
	if err == nil && key {
		t.Skip("TF_VAR_VSPHERE_TEST_ESXI must not be set for this acceptance test")
	}
}

func testAccVSphereLicenseWithLabelConfig() string {
	return fmt.Sprintf(`
resource "vsphere_license" "foo" {
 license_key = "%s"
  labels = {
   VpxClientLicenseLabel = "Hello World"
   TestTitle = "fooBar"
  }
}
`, os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testAccVSphereLicenseBasicConfig() string {
	return fmt.Sprintf(`
resource "vsphere_license" "foo" {
 license_key = "%s"
}
`, os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testAccVSphereLicenseDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).vimClient
	manager := license.NewManager(client.Client)
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_license" {
			continue
		}

		key := rs.Primary.ID
		if isKeyPresent(key, manager) {
			message += fmt.Sprintf("%s is still present on the server", key)
		}
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

func testAccVSphereLicenseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		manager := license.NewManager(client.Client)

		if !isKeyPresent(rs.Primary.ID, manager) {
			return fmt.Errorf("%s key not found on the server", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVSphereLicenseNotExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]

		if ok {
			return fmt.Errorf("%s key should not be present on the server", name)
		}

		return nil
	}
}

func testAccVSpherePreLicenseBasicCheck(t *testing.T) {
	if key := os.Getenv("TF_VAR_VSPHERE_LICENSE"); key == "" {
		t.Fatal("TF_VAR_VSPHERE_LICENSE must be set for acceptance test")
	}
}

func testAccVSphereLicenseWithLabelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		manager := license.NewManager(client.Client)

		if !isKeyPresent(rs.Primary.ID, manager) {
			return fmt.Errorf("%s key not found on the server", rs.Primary.ID)
		}

		info := getLicenseInfoFromKey(rs.Primary.ID, manager)

		if len(info.Labels) == 0 {
			return fmt.Errorf("The labels were not set for the key %s", info.LicenseKey)
		}

		if len(info.Labels) != 2 {
			return fmt.Errorf(`The number of labels on the server are incorrect. Expected 2 Got %d`,
				len(info.Labels))
		}

		return nil
	}
}
