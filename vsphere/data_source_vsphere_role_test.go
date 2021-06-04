package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereRoleConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "id",
						"vsphere_role.test-role", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "name",
						"vsphere_role.test-role", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.0",
						"vsphere_role.test-role", "role_privileges.0",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.1",
						"vsphere_role.test-role", "role_privileges.1",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.2",
						"vsphere_role.test-role", "role_privileges.2",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.3",
						"vsphere_role.test-role", "role_privileges.3",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereRoleConfig() string {
	return fmt.Sprintf(`
resource "vsphere_role" test-role {
  name = "terraform-test-role1"
  role_privileges = ["%s", "%s","%s","%s"]
}

data "vsphere_role" "role1" {
  label = vsphere_role.test-role.label
}
`,
		Privilege1,
		Privilege2,
		Privilege3,
		Privilege4,
	)
}
