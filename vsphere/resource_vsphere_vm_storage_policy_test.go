package vsphere

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const policyResource = "policy1"

func TestAccResourceVMStoragePolicy_basic(t *testing.T) {
	policyName := "terraform_test_policy" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVMStoragePolicyCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVMStoragePolicyonfigBasic(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVMStoragePolicyCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_vm_storage_policy."+policyResource, "name", policyName),
					resource.TestCheckResourceAttr("vsphere_vm_storage_policy."+policyResource, "tag_rules.0.tag_category", "cat1"),
					resource.TestCheckResourceAttr("vsphere_vm_storage_policy."+policyResource, "tag_rules.0.tags.0", "tag1"),
					resource.TestCheckResourceAttr("vsphere_vm_storage_policy."+policyResource, "tag_rules.1.tag_category", "cat2"),
					resource.TestCheckResourceAttr("vsphere_vm_storage_policy."+policyResource, "tag_rules.1.tags.0", "tag2"),
					resource.TestCheckResourceAttr("vsphere_vm_storage_policy."+policyResource, "tag_rules.1.tags.1", "tag3"),
				),
			},
		},
	})
}

func testAccResourceVMStoragePolicyCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVMStoragePolicy(s, policyResource)
		if err != nil {
			if strings.Contains(err.Error(), "Profile not found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected policy profile to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereVMStoragePolicyonfigBasic(policyName string) string {
	return fmt.Sprintf(`
resource "vsphere_tag_category" "category1" {
  name = "cat1"
  cardinality = "SINGLE"
  associable_types = ["Datastore"]
}

resource "vsphere_tag_category" "category2" {
  name = "cat2"
  cardinality = "SINGLE"
  associable_types = ["Datastore"]
}

resource "vsphere_tag" "tag1" {
  name        = "tag1"
  category_id = "${vsphere_tag_category.category1.id}"
}

resource "vsphere_tag" "tag2" {
  name        = "tag2"
  category_id = "${vsphere_tag_category.category2.id}"
}

resource "vsphere_tag" "tag3" {
  name        = "tag3"
  category_id = "${vsphere_tag_category.category2.id}"
}

resource "vsphere_vm_storage_policy" "%s" {
  name = "%s"
  description = "description"

  tag_rules {
    tag_category = vsphere_tag_category.category1.name
    tags = [vsphere_tag.tag1.name]
  }

 tag_rules {
    tag_category = vsphere_tag_category.category2.name
    tags = [vsphere_tag.tag2.name, vsphere_tag.tag3.name]
  }
  
}
`, policyResource,
		policyName,
	)
}
