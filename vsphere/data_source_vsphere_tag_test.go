package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereTag_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereTagConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_tag.testacc-tag-data",
						"name",
						testAccDataSourceVSphereTagConfigName,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_tag.testacc-tag-data",
						"description",
						testAccDataSourceVSphereTagConfigDescription,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_tag.testacc-tag-data", "id",
						"vsphere_tag.testacc-tag", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_tag.testacc-tag-data", "category_id",
						"vsphere_tag_category.testacc-category", "id",
					),
				),
			},
		},
	})
}

const testAccDataSourceVSphereTagConfigName = "testacc-tag"
const testAccDataSourceVSphereTagConfigDescription = "Managed by Terraform"

func testAccDataSourceVSphereTagConfig() string {
	return fmt.Sprintf(`
variable "tag_category_name" {
  default = "%s"
}

variable "tag_category_description" {
  default = "%s"
}

variable "tag_category_cardinality" {
  default = "%s"
}

variable "tag_category_associable_types" {
  default = [
    "%s",
  ]
}

variable "tag_name" {
  default = "%s"
}

variable "tag_description" {
  default = "%s"
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "${var.tag_category_name}"
  description = "${var.tag_category_description}"
  cardinality = "${var.tag_category_cardinality}"

  associable_types = "${var.tag_category_associable_types}"
}

resource "vsphere_tag" "testacc-tag" {
  name        = "${var.tag_name}"
  description = "${var.tag_description}"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

data "vsphere_tag" "testacc-tag-data" {
  name        = "${vsphere_tag.testacc-tag.name}"
  category_id = "${vsphere_tag.testacc-tag.category_id}"
}
`,
		testAccDataSourceVSphereTagCategoryConfigName,
		testAccDataSourceVSphereTagCategoryConfigDescription,
		testAccDataSourceVSphereTagCategoryConfigCardinality,
		testAccDataSourceVSphereTagCategoryConfigAssociableType,
		testAccDataSourceVSphereTagConfigName,
		testAccDataSourceVSphereTagConfigDescription,
	)
}
