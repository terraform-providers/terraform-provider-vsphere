package vsphere

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceVSphereContentLibraryItem_localOva(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereContentLibraryItemPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryItemCheckExists(false),
		Steps: []resource.TestStep{
			{
				PreConfig: testAccResourceVSphereContentLibraryItemGetOva,
				Config:    testAccResourceVSphereContentLibraryItemConfig_localOva(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "description", regexp.MustCompile("TestAcc Description"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "name", regexp.MustCompile("testacc-item"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "type", regexp.MustCompile("ovf"),
					),
					testAccResourceVSphereContentLibraryItemDescription(regexp.MustCompile("TestAcc Description")),
					testAccResourceVSphereContentLibraryItemName(regexp.MustCompile("testacc-item")),
					testAccResourceVSphereContentLibraryItemType(regexp.MustCompile("ovf")),
					testAccResourceVSphereContentLibraryItemDestroyFile("./testdata/test.ova"),
				),
			},
		},
	})
}

func TestAccResourceVSphereContentLibraryItem_remoteOvf(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereContentLibraryItemPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryItemCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereContentLibraryItemConfig_remoteOvf(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "description", regexp.MustCompile("TestAcc Description"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "name", regexp.MustCompile("testacc-item"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "type", regexp.MustCompile("ovf"),
					),
					testAccResourceVSphereContentLibraryItemDescription(regexp.MustCompile("TestAcc Description")),
					testAccResourceVSphereContentLibraryItemName(regexp.MustCompile("testacc-item")),
					testAccResourceVSphereContentLibraryItemType(regexp.MustCompile("ovf")),
				),
			},
			{
				ResourceName:      "vsphere_content_library_item.item",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"file_url",
				},
				Config: testAccResourceVSphereContentLibraryItemConfig_remoteOvf(),
			},
		},
	})
}

func TestAccResourceVSphereContentLibraryItem_remoteOva(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereContentLibraryItemPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryItemCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereContentLibraryItemConfig_remoteOva(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "description", regexp.MustCompile("TestAcc Description"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "name", regexp.MustCompile("testacc-item"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "type", regexp.MustCompile("ovf"),
					),
					testAccResourceVSphereContentLibraryItemDescription(regexp.MustCompile("TestAcc Description")),
					testAccResourceVSphereContentLibraryItemName(regexp.MustCompile("testacc-item")),
					testAccResourceVSphereContentLibraryItemType(regexp.MustCompile("ovf")),
				),
			},
			{
				ResourceName:      "vsphere_content_library_item.item",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"file_url",
				},
				Config: testAccResourceVSphereContentLibraryItemConfig_remoteOva(),
			},
		},
	})
}

func testAccResourceVSphereContentLibraryItemGetOva() {
	testAccResourceVSphereContentLibraryItemGetFile(os.Getenv("TF_VAR_VSPHERE_TEST_OVA"), "./testdata/test.ova")
}

func testAccResourceVSphereContentLibraryItemGetFile(url, file string) {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	out, _ := os.Create(file)
	defer out.Close()

	io.Copy(out, resp.Body)
}

func testAccResourceVSphereContentLibraryItemDestroyFile(file string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		os.Remove(file)
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_content_library acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_content_library acceptance tests")
	}
}

func testAccResourceVSphereContentLibraryItemDescription(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Description) {
			return fmt.Errorf("Content Library item description does not match. expected: %s, got %s", expected.String(), library.Description)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemName(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Name) {
			return fmt.Errorf("Content Library item name does not match. expected: %s, got %s", expected.String(), library.Name)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemType(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Type) {
			return fmt.Errorf("Content Library item type does not match. expected: %s, got %s", expected.String(), library.Type)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemConfig_localOva() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "testacc-item"
  description = "TestAcc Description"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = "./testdata/test.ova"
}
`,
		testAccResourceVSphereContentLibraryItemConfig_base(),
	)
}

func testAccResourceVSphereContentLibraryItemConfig_remoteOvf() string {
	return fmt.Sprintf(`
%s

variable "file" {
  default = "%s" 
}

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "testacc-item"
  description = "TestAcc Description"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}
`,
		testAccResourceVSphereContentLibraryItemConfig_base(),
		os.Getenv("TF_VAR_VSPHERE_TEST_OVF"),
	)
}

func testAccResourceVSphereContentLibraryItemConfig_remoteOva() string {
	return fmt.Sprintf(`
%s

variable "file" {
  default = "%s" 
}

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [ data.vsphere_datastore.rootds1.id ]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "testacc-item"
  description = "TestAcc Description"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}
`,
		testAccResourceVSphereContentLibraryItemConfig_base(),
		os.Getenv("TF_VAR_VSPHERE_TEST_OVA"),
	)
}

func testAccResourceVSphereContentLibraryItemConfig_base() string {
	return testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1())
}

func testAccResourceVSphereContentLibraryItemCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetContentLibraryItem(s, "item")
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
			return fmt.Errorf("expected Content Library item to be missing")
		}
		return nil
	}
}
