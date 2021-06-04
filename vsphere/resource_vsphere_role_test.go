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

const RoleResource = "role1"
const Privilege1 = "Alarm.Acknowledge"
const Privilege2 = "Alarm.Create"
const Privilege3 = "Datacenter.Create"
const Privilege4 = "Datacenter.Move"

func TestAccResourceVsphereRole_basic(t *testing.T) {
	roleName := "terraform_role" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVsphereRoleCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereRoleConfigBasic(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVsphereRoleCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.0", Privilege1),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.1", Privilege2),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.2", Privilege3),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.3", Privilege4),
				),
			},
		},
	})
}

func testAccResourceVsphereRoleCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereRole(s, RoleResource)
		if err != nil {
			if strings.Contains(err.Error(), "role not found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected role to be missing")
		}
		return nil
	}
}

func testAccResourceVsphereRoleConfigBasic(roleName string) string {
	return fmt.Sprintf(`
  resource "vsphere_role" "%s" {
  name = "%s"
  role_privileges = ["%s", "%s","%s","%s"]
}
`, RoleResource,
		roleName,
		Privilege1,
		Privilege2,
		Privilege3,
		Privilege4,
	)
}
