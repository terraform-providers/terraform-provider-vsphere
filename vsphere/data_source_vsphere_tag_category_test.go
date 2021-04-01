package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereTagCategory_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereTagCategoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_tag_category.testacc-category-data",
						"name",
						testAccDataSourceVSphereTagCategoryConfigName,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_tag_category.testacc-category-data",
						"description",
						testAccDataSourceVSphereTagCategoryConfigDescription,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_tag_category.testacc-category-data",
						"cardinality",
						testAccDataSourceVSphereTagCategoryConfigCardinality,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_tag_category.testacc-category-data",
						"associable_types.#",
						"1",
					),
					resource.TestCheckTypeSetElemAttr(
						"data.vsphere_tag_category.testacc-category-data",
						"associable_types.*",
						testAccDataSourceVSphereTagCategoryConfigAssociableType,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_tag_category.testacc-category-data", "id",
						"vsphere_tag_category.testacc-category", "id",
					),
				),
			},
		},
	})
}

const testAccDataSourceVSphereTagCategoryConfigName = "testacc-category"
const testAccDataSourceVSphereTagCategoryConfigDescription = "Managed by Terraform"
const testAccDataSourceVSphereTagCategoryConfigCardinality = vSphereTagCategoryCardinalitySingle
const testAccDataSourceVSphereTagCategoryConfigAssociableType = vSphereTagTypeVirtualMachine

func testAccDataSourceVSphereTagCategoryConfig() string {
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

resource "vsphere_tag_category" "testacc-category" {
  name        = "${var.tag_category_name}"
  description = "${var.tag_category_description}"
  cardinality = "${var.tag_category_cardinality}"

  associable_types = "${var.tag_category_associable_types}"
}

data "vsphere_tag_category" "testacc-category-data" {
  name = "${vsphere_tag_category.testacc-category.name}"
}
`,
		testAccDataSourceVSphereTagCategoryConfigName,
		testAccDataSourceVSphereTagCategoryConfigDescription,
		testAccDataSourceVSphereTagCategoryConfigCardinality,
		testAccDataSourceVSphereTagCategoryConfigAssociableType,
	)
}
