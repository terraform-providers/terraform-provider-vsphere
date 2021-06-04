package vsphere

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceVSphereHost_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereHostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
			{
				Config: testaccvspherehostconfigImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceVSphereHost_rootFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigRootfolder(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_connection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigConnection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostConnected("vsphere_host.h1", false),
				),
			},
			{
				Config: testaccvspherehostconfigConnection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostConnected("vsphere_host.h1", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_maintenance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigMaintenance(true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostMaintenanceState("vsphere_host.h1", true),
				),
			},
			{
				Config: testaccvspherehostconfigMaintenance(false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostMaintenanceState("vsphere_host.h1", false),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_lockdown(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigLockdown("strict"),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostLockdownState("vsphere_host.h1", "strict"),
				),
			},
			{
				Config: testaccvspherehostconfigLockdown("normal"),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostLockdownState("vsphere_host.h1", "normal"),
				),
			},
			{
				Config: testaccvspherehostconfigLockdown("disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostLockdownState("vsphere_host.h1", "disabled"),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_lockdown_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testaccvspherehostconfigLockdown("invalidvalue"),
				ExpectError: regexp.MustCompile(`be one of \[disabled normal strict\], got invalidvalue`),
			},
		},
	})
}

func TestAccResourceVSphereHost_emptyLicense(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigEmptylicense(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
		},
	})
}

func testAccVSphereHostExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostExists(client, hostID)
		if err != nil {
			return err
		}

		if !res {
			return fmt.Errorf("Host with ID %s not found", hostID)
		}

		return nil
	}
}

func testAccVSphereHostConnected(name string, shouldBeConnected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostConnected(client, hostID)
		if err != nil {
			return err
		}

		if res != shouldBeConnected {
			return fmt.Errorf("Host with ID %s connection: %t, expected %t", hostID, res, shouldBeConnected)
		}

		return nil
	}
}

func testAccVSphereHostMaintenanceState(name string, inMaintenance bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostInMaintenance(client, hostID)
		if err != nil {
			return err
		}

		if res != inMaintenance {
			return fmt.Errorf("Host with ID %s in maintenance : %t, expected %t", hostID, res, inMaintenance)
		}

		return nil
	}
}

func testAccVSphereHostLockdownState(name string, lockdown string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := checkHostLockdown(client, hostID, lockdown)
		if err != nil {
			return err
		}

		if !res {
			return fmt.Errorf("Host with ID %s not in desired lockdown state. Current state: %s", hostID, lockdown)
		}

		return nil
	}
}

func testAccVSphereHostDestroy(s *terraform.State) error {
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_host" {
			continue
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostExists(client, hostID)
		if err != nil {
			return err
		}

		if res {
			message += fmt.Sprintf("Host with ID %s was found", hostID)
		}
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

func hostExists(client *govmomi.Client, hostID string) (bool, error) {
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	if hs.Reference().Value != hostID {
		return false, nil
	}
	return true, nil
}

func hostConnected(client *govmomi.Client, hostID string) (bool, error) {
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return false, err
	}

	connectionState, err := hostsystem.GetConnectionState(hs)
	if err != nil {
		return false, err
	}
	return connectionState == types.HostSystemConnectionStateConnected, nil
}

func hostInMaintenance(client *govmomi.Client, hostID string) (bool, error) {
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return false, err
	}

	maintenanceState, err := hostsystem.HostInMaintenance(hs)
	if err != nil {
		return false, err
	}
	return maintenanceState, nil
}

func checkHostLockdown(client *govmomi.Client, hostID, lockdownMode string) (bool, error) {
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return false, err
	}
	hostProps, err := hostsystem.Properties(host)
	if err != nil {
		return false, err
	}

	lockdownModes := map[types.HostLockdownMode]string{
		types.HostLockdownModeLockdownDisabled: "disabled",
		types.HostLockdownModeLockdownNormal:   "normal",
		types.HostLockdownModeLockdownStrict:   "strict",
	}

	modeString, ok := lockdownModes[hostProps.Config.LockdownMode]
	if !ok {
		return false, fmt.Errorf("Unknown lockdown mode found: %s", hostProps.Config.LockdownMode)
	}

	return modeString == lockdownMode, nil
}

func testAccVSphereHostConfig() string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}

	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  # Makes sense to update
	  license = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testaccvspherehostconfigRootfolder() string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  # Makes sense to update
	  license = "%s"
	  datacenter = data.vsphere_datacenter.rootdc1.id
	}
	`, testhelper.ConfigDataRootDC1(), os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testaccvspherehostconfigEmptylicense() string {
	return fmt.Sprintf(`
	%s 
	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  # Makes sense to update
	  datacenter = data.vsphere_datacenter.rootdc1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
	)
}

func testaccvspherehostconfigImport() string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}
		
	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id
	
	  # Makes sense to update
	  license = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}	  
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testaccvspherehostconfigConnection(connection bool) string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}
		
	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id
	
	  license = "%s"
	  connected = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}	  
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"),
		strconv.FormatBool(connection))
}

func testaccvspherehostconfigMaintenance(maintenance bool) string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}
		
	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id
	
	  license = "%s"
	  connected = "true"
	  maintenance = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}	  
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"),
		strconv.FormatBool(maintenance))
}

func testaccvspherehostconfigLockdown(lockdown string) string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}
		
	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id
	
	  license = "%s"
	  connected = "true"
	  maintenance = "false"
	  lockdown = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}	  
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"),
		lockdown)
}
