package vsphere

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const EntityPermissionResource = "entity_permission1"

func TestAccResourcevsphereEntityPermissions_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccResourceVSphereEntityPermissionsPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceEntityPermissionsCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereEntityPermissionsConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceEntityPermissionsCheckExists(true),
					resource.TestCheckResourceAttrSet("vsphere_entity_permissions."+EntityPermissionResource, "permissions.0.user_or_group"),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+EntityPermissionResource, "permissions.0.propagate", "true"),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+EntityPermissionResource, "permissions.0.is_group", "true"),
				),
			},
		},
	})
}

func testAccResourceVSphereEntityPermissionsPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_USER_GROUP") == "" {
		t.Skip("set TF_VAR_VSPHERE_ENTITY_PERMISSION_USER_GROUP to run vsphere_entity_permission acceptance tests")
	}
}

func testAccResourceEntityPermissionsCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereEntityPermission(s, EntityPermissionResource)
		if err != nil {
			if strings.Contains(err.Error(), "permissions not found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected permissions to be missing")
		}
		return nil
	}
}

func testAccResourceVsphereEntityPermissionsConfigBasic() string {
	return fmt.Sprintf(`
%s

	data "vsphere_virtual_machine" "vm" {
		datacenter_id = data.vsphere_datacenter.rootdc1.id
        name          = "%s"
	}

	data "vsphere_role" "role1" {
	  label = "Administrator"
	}

   resource vsphere_entity_permissions "%s" {
	   entity_id = data.vsphere_virtual_machine.vm.id
	   entity_type = "VirtualMachine"
	   permissions {
		 user_or_group = "%s"
		 propagate = true
		 is_group = true
		 role_id = data.vsphere_role.role1.id
	   }
   }
`,
		testhelper.ConfigDataRootDC1(),
		os.Getenv("TF_VAR_VSPHERE_VM_V1_PATH"),
		EntityPermissionResource,
		os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_USER_GROUP"),
	)
}
