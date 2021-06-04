package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceVSphereContentLibrary_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereContentLibraryPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereContentLibraryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "description", regexp.MustCompile("Library Description"),
					),
					testAccResourceVSphereContentLibraryDescription(regexp.MustCompile("Library Description")),
					testAccResourceVSphereContentLibraryName(regexp.MustCompile("testacc_content_library")),
				),
			},
			{
				ResourceName:      "vsphere_content_library.library",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testAccResourceVSphereContentLibraryConfig(),
			},
		},
	})
}

func TestAccResourceVSphereContentLibrary_subscribed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereContentLibraryPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherecontentlibraryconfigSubscribed(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "description", regexp.MustCompile("Library Description"),
					),
					testAccResourceVSphereContentLibraryDescription(regexp.MustCompile("Library Description")),
					testAccResourceVSphereContentLibraryName(regexp.MustCompile("testacc_subscribed")),
				),
			},
		},
	})
}
func TestAccResourceVSphereContentLibrary_authenticated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereContentLibraryPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherecontentlibraryconfigAuthenticated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
				),
			},
		},
	})
}

func testAccResourceVSphereContentLibraryPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_content_library acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_content_library acceptance tests")
	}
}

func testAccResourceVSphereContentLibraryDescription(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibrary(s, "library")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Description) {
			return fmt.Errorf("Content Library description does not match. expected: %s, got %s", expected.String(), library.Description)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryName(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibrary(s, "library")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Name) {
			return fmt.Errorf("Content Library name does not match. expected: %s, got %s", expected.String(), library.Name)
		}
		return nil
	}
}

func testaccresourcevspherecontentlibraryconfigBase() string {
	return testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1())
}

func testaccresourcevspherecontentlibraryconfigAuthenticated() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library_published" {
  name            = "testacc_published"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
	publication {
	  authentication_method = "BASIC"
		username = "vcsp"
		password = "Password123!"
	  published = true
	}
}

resource "vsphere_content_library" "library" {
  name            = "testacc_subscribed"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
	subscription {
	  authentication_method = "BASIC"
		username = "vcsp"
		password = "Password123!"
	  subscription_url = vsphere_content_library.library_published.publication.0.publish_url
	}
}
`,
		testaccresourcevspherecontentlibraryconfigBase())
}
func testaccresourcevspherecontentlibraryconfigSubscribed() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library_published" {
  name            = "testacc_published"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
	publication {
	  published = true
	}
}

resource "vsphere_content_library" "library" {
  name            = "testacc_subscribed"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
	subscription {
	  subscription_url = vsphere_content_library.library_published.publication.0.publish_url
	}
}
`,
		testaccresourcevspherecontentlibraryconfigBase())
}

func testAccResourceVSphereContentLibraryConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
}
`,
		testaccresourcevspherecontentlibraryconfigBase())
}

func testAccResourceVSphereContentLibraryCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetContentLibrary(s, "library")
		if err != nil {
			missingState, _ := regexp.MatchString("not found in state", err.Error())
			missingVSphere, _ := regexp.MatchString("404 Not Found", err.Error())
			if missingState && !expected || missingVSphere && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected Content Library to be missing")
		}
		return nil
	}
}
