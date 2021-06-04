package vsphere

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	resource.AddTestSweepers("tags", &resource.Sweeper{
		Name:         "tag_cleanup",
		Dependencies: nil,
		F:            tagSweep,
	})
	resource.AddTestSweepers("datacenters", &resource.Sweeper{
		Name:         "datacenter_cleanup",
		Dependencies: nil,
		F:            dcSweep,
	})
	resource.AddTestSweepers("vms", &resource.Sweeper{
		Name:         "vm_cleanup",
		Dependencies: nil,
		F:            vmSweep,
	})
	resource.AddTestSweepers("rps", &resource.Sweeper{
		Name:         "rp_cleanup",
		Dependencies: nil,
		F:            rpSweep,
	})
	resource.AddTestSweepers("net", &resource.Sweeper{
		Name:         "net_cleanup",
		Dependencies: nil,
		F:            netSweep,
	})
	resource.AddTestSweepers("folder", &resource.Sweeper{
		Name:         "folder_cleanup",
		Dependencies: nil,
		F:            folderSweep,
	})
	resource.AddTestSweepers("dss", &resource.Sweeper{
		Name:         "ds_cleanup",
		Dependencies: nil,
		F:            dsSweep,
	})
	resource.AddTestSweepers("dsps", &resource.Sweeper{
		Name:         "dsp_cleanup",
		Dependencies: nil,
		F:            dspSweep,
	})
	resource.AddTestSweepers("ccs", &resource.Sweeper{
		Name:         "cc_cleanup",
		Dependencies: nil,
		F:            ccSweep,
	})
}

func testAccClientPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC to run vsphere_virtual_machine state migration tests (provider connection is required)")
	}
	testAccPreCheck(t)
}

func testAccClientGenerateConfig() *Config {
	insecure, _ := strconv.ParseBool(os.Getenv("TF_VAR_VSPHERE_ALLOW_UNVERIFIED_SSL"))
	debug, _ := strconv.ParseBool(os.Getenv("TF_VAR_VSPHERE_CLIENT_DEBUG"))

	return &Config{
		InsecureFlag:  insecure,
		Debug:         debug,
		User:          os.Getenv("TF_VAR_VSPHERE_USER"),
		Password:      os.Getenv("TF_VAR_VSPHERE_PASSWORD"),
		VSphereServer: os.Getenv("TF_VAR_VSPHERE_SERVER"),
		DebugPath:     os.Getenv("TF_VAR_VSPHERE_CLIENT_DEBUG_PATH"),
		DebugPathRun:  os.Getenv("TF_VAR_VSPHERE_CLIENT_DEBUG_PATH_RUN"),
	}
}

func testAccClientGenerateData(t *testing.T, c *Config) string {
	_, err := c.Client()
	if err != nil {
		t.Fatalf("error setting up client: %s", err)
	}

	vimSessionFile, err := c.vimSessionFile()
	if err != nil {
		t.Fatalf("error computing VIM session file: %s", err)
	}

	vimData, err := ioutil.ReadFile(vimSessionFile)
	if err != nil {
		t.Fatalf("error reading VIM session file: %s", err)
	}

	return string(vimData)
}

func testAccClientCheckStatNoExist(t *testing.T, p string) {
	_, err := os.Stat(p)
	switch {
	case err == nil:
		t.Fatalf("expected session file %q to not exist", p)
	case os.IsNotExist(err):
		return
	case err != nil:
		t.Fatalf("could not stat path %q: %s", p, err)
	}
}

func TestAccClient_persistence(t *testing.T) {
	testAccClientPreCheck(t)

	vimSessionDir, err := ioutil.TempDir("", "tf-vsphere-test-vimsessiondir")
	if err != nil {
		t.Fatalf("error creating VIM session temp directory: %s", err)
	}
	restSessionDir, err := ioutil.TempDir("", "tf-vsphere-test-restsessiondir")
	if err != nil {
		t.Fatalf("error creating REST session temp directory: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(vimSessionDir); err != nil {
			log.Printf("[DEBUG] Error removing test VIM session directory %q: %s", vimSessionDir, err)
		}
	}()
	defer func() {
		if err = os.RemoveAll(restSessionDir); err != nil {
			log.Printf("[DEBUG] Error removing test REST session directory %q: %s", restSessionDir, err)
		}
	}()

	c := testAccClientGenerateConfig()
	c.Persist = true
	c.VimSessionPath = vimSessionDir

	expectedVim := testAccClientGenerateData(t, c)

	// This will create a brand new session under normal circumstances
	actualVim := testAccClientGenerateData(t, c)

	if expectedVim != actualVim {
		t.Fatalf("VIM session data mismatch.\n\n\n\nExpected:\n\n %s\n\nActual:\n\n%s\n\n", expectedVim, actualVim)
	}
}

func TestAccClient_noPersistence(t *testing.T) {
	testAccClientPreCheck(t)

	vimSessionDir, err := ioutil.TempDir("", "tf-vsphere-test-vimsessiondir")
	if err != nil {
		t.Fatalf("error creating VIM session temp directory: %s", err)
	}
	restSessionDir, err := ioutil.TempDir("", "tf-vsphere-test-restsessiondir")
	if err != nil {
		t.Fatalf("error creating REST session temp directory: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(vimSessionDir); err != nil {
			log.Printf("[DEBUG] Error removing test VIM session directory %q: %s", vimSessionDir, err)
		}
	}()
	defer func() {
		if err = os.RemoveAll(restSessionDir); err != nil {
			log.Printf("[DEBUG] Error removing test REST session directory %q: %s", restSessionDir, err)
		}
	}()

	c := testAccClientGenerateConfig()
	// Just to be explicit on intent
	c.Persist = false
	c.VimSessionPath = vimSessionDir

	_, err = c.Client()
	if err != nil {
		t.Fatalf("error setting up client: %s", err)
	}

	vimSessionFile, err := c.vimSessionFile()
	if err != nil {
		t.Fatalf("error computing VIM session file: %s", err)
	}

	testAccClientCheckStatNoExist(t, vimSessionFile)
}

func TestNewConfig(t *testing.T) {
	expected := &Config{
		User:           "foo",
		Password:       "bar",
		InsecureFlag:   true,
		VSphereServer:  "vsphere.foo.internal",
		Debug:          true,
		DebugPathRun:   "./foo",
		DebugPath:      "./bar",
		Persist:        true,
		VimSessionPath: "./baz",
	}

	r := &schema.Resource{Schema: Provider().Schema}
	d := r.Data(nil)
	_ = d.Set("user", expected.User)
	_ = d.Set("password", expected.Password)
	_ = d.Set("allow_unverified_ssl", expected.InsecureFlag)
	_ = d.Set("vsphere_server", expected.VSphereServer)
	_ = d.Set("client_debug", expected.Debug)
	_ = d.Set("client_debug_path_run", expected.DebugPathRun)
	_ = d.Set("client_debug_path", expected.DebugPath)
	_ = d.Set("persist_session", expected.Persist)
	_ = d.Set("vim_session_path", expected.VimSessionPath)

	actual, err := NewConfig(d)
	if err != nil {
		t.Fatalf("error creating new configuration: %s", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
