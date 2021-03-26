package vsphere

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const ENTITY_PERMISSION_RESOURCE = "entity_permission1"

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
					resource.TestCheckResourceAttrSet("vsphere_entity_permissions."+ENTITY_PERMISSION_RESOURCE, "permissions.0.user_or_group"),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+ENTITY_PERMISSION_RESOURCE, "permissions.0.role_id",
						os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ROLE_ID")),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+ENTITY_PERMISSION_RESOURCE, "permissions.0.propagate", "true"),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+ENTITY_PERMISSION_RESOURCE, "permissions.0.is_group", "true"),
				),
			},
		},
	})
}

func testAccResourceVSphereEntityPermissionsPreCheck(t *testing.T) {

	if os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ENTITY_ID") == "" {
		t.Skip("set TF_VAR_VSPHERE_ENTITY_PERMISSION_ENTITY_ID to run vsphere_entity_permission tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ENTITY_TYPE") == "" {
		t.Skip("set TF_VAR_VSPHERE_ENTITY_PERMISSION_ENTITY_TYPE to run vsphere_entity_permission acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_USER_GROUP") == "" {
		t.Skip("set TF_VAR_VSPHERE_ENTITY_PERMISSION_USER_GROUP to run vsphere_entity_permission acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ROLE_ID") == "" {
		t.Skip("set TF_VAR_VSPHERE_ENTITY_PERMISSION_ROLE_ID  to run vsphere_entity_permission acceptance tests")
	}
}

func testAccResourceEntityPermissionsCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereEntityPermission(s, ENTITY_PERMISSION_RESOURCE)
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
   resource vsphere_entity_permissions "%s" {
   entity_id = "%s"
   entity_type = "%s"
   permissions {
     user_or_group = "%s"
     propagate = true
     is_group = true
     role_id = "%s"
   }
 }
`, ENTITY_PERMISSION_RESOURCE,
		os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ENTITY_ID"),
		os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ENTITY_TYPE"),
		os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_USER_GROUP"),
		os.Getenv("TF_VAR_VSPHERE_ENTITY_PERMISSION_ROLE_ID"),
	)
}
