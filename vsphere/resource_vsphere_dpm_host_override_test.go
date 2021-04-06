package vsphere

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereDPMHostOverride_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDPMHostOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDPMHostOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDPMHostOverrideConfigDefaults(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDPMHostOverrideExists(true),
					testAccResourceVSphereDPMHostOverrideMatch(types.DpmBehaviorManual, false),
				),
			},
			{
				ResourceName:      "vsphere_dpm_host_override.dpm_host_override",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cluster, err := testGetComputeCluster(s, "rootcompute_cluster1", "data.vsphere_compute_cluster")
					if err != nil {
						return "", err
					}
					host, err := testGetHostFromDataSource(s, "roothost1")
					if err != nil {
						return "", err
					}

					m := make(map[string]string)
					m["compute_cluster_path"] = cluster.InventoryPath
					m["host_path"] = host.InventoryPath
					b, err := json.Marshal(m)
					if err != nil {
						return "", err
					}

					return string(b), nil
				},
				Config: testAccResourceVSphereDPMHostOverrideConfigOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDPMHostOverrideExists(true),
					testAccResourceVSphereDPMHostOverrideMatch(types.DpmBehaviorAutomated, true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDPMHostOverride_overrides(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDPMHostOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDPMHostOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDPMHostOverrideConfigOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDPMHostOverrideExists(true),
					testAccResourceVSphereDPMHostOverrideMatch(types.DpmBehaviorAutomated, true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDPMHostOverride_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDPMHostOverridePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDPMHostOverrideExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDPMHostOverrideConfigDefaults(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDPMHostOverrideExists(true),
					testAccResourceVSphereDPMHostOverrideMatch(types.DpmBehaviorManual, false),
				),
			},
			{
				Config: testAccResourceVSphereDPMHostOverrideConfigOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDPMHostOverrideExists(true),
					testAccResourceVSphereDPMHostOverrideMatch(types.DpmBehaviorAutomated, true),
				),
			},
		},
	})
}

func testAccResourceVSphereDPMHostOverridePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI1") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI1 to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI2") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI2 to run vsphere_compute_cluster acceptance tests")
	}
}

func testAccResourceVSphereDPMHostOverrideExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		info, err := testGetComputeClusterDPMHostConfig(s, "dpm_host_override")

		if err != nil {
			if expected == false {
				if viapi.IsManagedObjectNotFoundError(err) {
					// This is not necessarily a missing override, but more than likely a
					// missing cluster, which happens during destroy as the dependent
					// resources will be missing as well, so want to treat this as a
					// deleted override as well.
					return nil
				}
			}
			return err
		}

		switch {
		case info == nil && !expected:
			// Expected missing
			return nil
		case info == nil && expected:
			// Expected to exist
			return errors.New("DPM host override missing when expected to exist")
		case !expected:
			return errors.New("DPM host override still present when expected to be missing")
		}

		return nil
	}
}

func testAccResourceVSphereDPMHostOverrideMatch(behavior types.DpmBehavior, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterDPMHostConfig(s, "dpm_host_override")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("DPM host override missing")
		}

		expected := &types.ClusterDpmHostConfigInfo{
			Behavior: behavior,
			Enabled:  structure.BoolPtr(enabled),
			Key:      actual.Key,
		}

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereDPMHostOverrideConfigDefaults() string {
	return fmt.Sprintf(`
%s

resource "vsphere_dpm_host_override" "dpm_host_override" {
  compute_cluster_id   = "${data.vsphere_compute_cluster.rootcompute_cluster1.id}"
  host_system_id       = "${data.vsphere_host.roothost1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootComputeCluster1(),
		),
	)
}

func testAccResourceVSphereDPMHostOverrideConfigOverrides() string {
	return fmt.Sprintf(`
%s


resource "vsphere_dpm_host_override" "dpm_host_override" {
  compute_cluster_id   = "${data.vsphere_compute_cluster.rootcompute_cluster1.id}"
  host_system_id       = data.vsphere_host.roothost1.id
  dpm_enabled          = true
  dpm_automation_level = "automated"
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
		),
	)
}
