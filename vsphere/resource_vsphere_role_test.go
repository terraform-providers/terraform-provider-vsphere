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

const ROLE_RESOURCE = "role1"
const PRIVILEGE_1 = "Alarm.Acknowledge"
const PRIVILEGE_2 = "Alarm.Create"
const PRIVILEGE_3 = "Datacenter.Create"
const PRIVILEGE_4 = "Datacenter.Move"

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
					resource.TestCheckResourceAttr("vsphere_role."+ROLE_RESOURCE, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+ROLE_RESOURCE, "role_privileges.0", PRIVILEGE_1),
					resource.TestCheckResourceAttr("vsphere_role."+ROLE_RESOURCE, "role_privileges.1", PRIVILEGE_2),
					resource.TestCheckResourceAttr("vsphere_role."+ROLE_RESOURCE, "role_privileges.2", PRIVILEGE_3),
					resource.TestCheckResourceAttr("vsphere_role."+ROLE_RESOURCE, "role_privileges.3", PRIVILEGE_4),
				),
			},
		},
	})
}

func testAccResourceVsphereRoleCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereRole(s, ROLE_RESOURCE)
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
`, ROLE_RESOURCE,
		roleName,
		PRIVILEGE_1,
		PRIVILEGE_2,
		PRIVILEGE_3,
		PRIVILEGE_4,
	)
}
