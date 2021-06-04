package vsphere

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi"
)

type genTfConfig func(string) string

func generateSteps(cfgFunc genTfConfig, netstack string) []resource.TestStep {
	out := make([]resource.TestStep, 0)
	for _, ipv4 := range []string{"dhcp", "192.0.2.10|255.255.255.0|192.0.2.1", ""} {
		for _, ipv6 := range []string{"autoconfig", "dhcp", "2001:DB8::10/32|2001:DB8::1", ""} {
			if ipv4 == "" && ipv6 == "" {
				continue
			}
			cfg := combineSnippets(ipv4Snippet(ipv4),
				ipv6Snippet(ipv6),
				netstackSnippet(netstack))
			out = append(out, []resource.TestStep{
				{
					Config: cfgFunc(cfg),
					Check: resource.ComposeTestCheckFunc(
						testAccVsphereVNicNetworkSettings("vsphere_vnic.v1", ipv4, ipv6, netstack),
					),
				},
			}...)
		}
	}
	return out
}

func TestAccResourceVSphereVNic_dvs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVNicDestroy,
		Steps:        generateSteps(testaccvspherevnicconfigDvs, "defaultTcpipStack"),
	})
}

func TestAccResourceVSphereVNic_dvs_vmotion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVNicDestroy,
		Steps:        generateSteps(testaccvspherevnicconfigDvs, "vmotion"),
	})
}

func TestAccResourceVSphereVNic_hvs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVNicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherevnicconfigHvs(combineSnippets(
					ipv4Snippet("192.0.2.10|255.255.255.0|192.0.2.1"),
					"",
					netstackSnippet("defaultTcpipStack"))),
				Check: resource.ComposeTestCheckFunc(
					testAccVsphereVNicNetworkSettings("vsphere_vnic.v1",
						"192.0.2.10|255.255.255.0|192.0.2.1",
						"",
						"defaultTcpipStack"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVNic_hvs_vmotion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVNicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherevnicconfigHvs(combineSnippets(
					ipv4Snippet("192.0.2.10|255.255.255.0|192.0.2.1"),
					"",
					netstackSnippet("vmotion"))),
				Check: resource.ComposeTestCheckFunc(
					testAccVsphereVNicNetworkSettings("vsphere_vnic.v1",
						"192.0.2.10|255.255.255.0|192.0.2.1",
						"",
						"vmotion"),
				),
			},
		},
	})
}

func testAccVsphereVNicNetworkSettings(name, ipv4State, ipv6State, netstack string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		idParts := strings.Split(rs.Primary.ID, "_")
		hostID := idParts[0]
		vmnicID := idParts[1]
		client := testAccProvider.Meta().(*Client).vimClient
		vnic, err := getVnicFromHost(context.TODO(), client, hostID, vmnicID)
		if err != nil {
			return err
		}

		switch ipv4State {
		case "dhcp":
			if !vnic.Spec.Ip.Dhcp {
				return fmt.Errorf("ipv4 network error, expected dhcp")
			}
		case "":
			break
		default:
			if vnic.Spec.Ip.Dhcp {
				return fmt.Errorf("ipv4 network error, expected static")
			}
			if ipv4State == "static" {
				break
			}
			addrBits := strings.Split(ipv4State, "|")
			if len(addrBits) != 3 {
				return fmt.Errorf("ipv4 test error, invalid parameter: %s", ipv4State)
			}
			ip := addrBits[0]
			netmask := addrBits[1]
			gw := addrBits[2]

			routeConfig := vnic.Spec.IpRouteSpec.IpRouteConfig.GetHostIpRouteConfig()
			if ip != vnic.Spec.Ip.IpAddress || netmask != vnic.Spec.Ip.SubnetMask || gw != routeConfig.DefaultGateway {
				return fmt.Errorf(
					"ipv4 network error, static config mismatch. ip %s vs %s, netmask %s vs %s, gw %s vs %s",
					ip, vnic.Spec.Ip.IpAddress,
					netmask, vnic.Spec.Ip.SubnetMask,
					gw, routeConfig.DefaultGateway)
			}
		}

		switch ipv6State {
		case "dhcp":
			if !*vnic.Spec.Ip.IpV6Config.DhcpV6Enabled {
				return fmt.Errorf("ipv6 network error, expected dhcp")
			}
		case "autoconfig":
			if !*vnic.Spec.Ip.IpV6Config.AutoConfigurationEnabled {
				return fmt.Errorf("ipv6 network error, expected autoconfig")
			}
		case "":
			break
		default:
			if *vnic.Spec.Ip.IpV6Config.AutoConfigurationEnabled || *vnic.Spec.Ip.IpV6Config.DhcpV6Enabled {
				return fmt.Errorf("ipv6 network error, expected static configuration")
			}
			if ipv6State == "static" {
				break
			}
			addrBits := strings.Split(ipv6State, "|")
			if len(addrBits) != 2 {
				return fmt.Errorf("ipv4 test error, invalid parameter: %s", ipv6State)
			}
			ipParts := strings.Split(addrBits[0], "/")
			ip := ipParts[0]
			prefix, err := strconv.ParseInt(ipParts[1], 10, 32)
			if err != nil {
				return fmt.Errorf("error while parsing prefix: %s", err)
			}
			gw := addrBits[1]
			routeConfig := vnic.Spec.IpRouteSpec.IpRouteConfig.GetHostIpRouteConfig()

			// loop through ipv6 IPs until we find our own
			addrFound := false
			for _, ipv6addr := range vnic.Spec.Ip.IpV6Config.IpV6Address {
				if strings.EqualFold(ipv6addr.IpAddress, ip) {
					if ipv6addr.PrefixLength != int32(prefix) ||
						strings.EqualFold(routeConfig.IpV6DefaultGateway, gw) {
						return fmt.Errorf(
							"ipv6 network error, static config mismatch. prefix length %d vs %d, gw %s vs %s",
							prefix, ipv6addr.PrefixLength,
							gw, routeConfig.IpV6DefaultGateway)
					}
					addrFound = true
					break
				}
			}
			if !addrFound {
				return fmt.Errorf("ipv6 network error, could not find %s assigned to the interface", ip)
			}
		}

		configuredNetstack := vnic.Spec.NetStackInstanceKey
		if netstack != configuredNetstack {
			return fmt.Errorf("netstack mismatch. expected %s found %s", netstack, configuredNetstack)
		}

		return nil
	}
}

func testAccVSphereVNicDestroy(s *terraform.State) error {
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_vnic" {
			continue
		}
		nicID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := nicExists(client, nicID)
		if err != nil {
			return err
		}

		if res {
			message += fmt.Sprintf("vNic with ID %s was found", nicID)
		}
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

func nicExists(client *govmomi.Client, nicID string) (bool, error) {
	toks := strings.Split(nicID, "_")
	vmnicID := toks[1]
	hostID := toks[0]
	_, err := getVnicFromHost(context.TODO(), client, hostID, vmnicID)
	if err != nil {
		if err.Error() == fmt.Sprintf("vNic interface with id %s not found", vmnicID) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func testaccvspherevnicconfigHvs(netConfig string) string {
	return fmt.Sprintf(`
%s

	data "vsphere_host" "h1" {
	  name          = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}
	
	
	resource "vsphere_host_virtual_switch" "hvs1" {
	  name             = "hashi-dc_HPG0"
	  host_system_id   = data.vsphere_host.h1.id
	  network_adapters = ["%s", "%s"]
	  active_nics      = ["%s"]
	  standby_nics     = ["%s"]
	}
	
	resource "vsphere_host_port_group" "p1" {
	  name                     = "ko-pg"
	  virtual_switch_name = vsphere_host_virtual_switch.hvs1.name
	  host_system_id   = data.vsphere_host.h1.id
	}
	
	resource "vsphere_vnic" "v1" {
	  host      = data.vsphere_host.h1.id
	  portgroup = vsphere_host_port_group.p1.name
	  %s
	}
	`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC0"),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		netConfig)
}

func testaccvspherevnicconfigDvs(netConfig string) string {
	return fmt.Sprintf(`
%s

	resource "vsphere_distributed_virtual_switch" "d1" {
	  name          = "hashi-dc_DVPG0"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	  host {
		host_system_id = data.vsphere_host.roothost1.id
		devices        = ["%s"]
	  }
	}
	
	resource "vsphere_distributed_port_group" "p1" {
	  name                            = "ko-pg"
	  vlan_id                         = 1234
	  distributed_virtual_switch_uuid = vsphere_distributed_virtual_switch.d1.id
	}
	
	resource "vsphere_vnic" "v1" {
	  host                    = data.vsphere_host.roothost1.id
	  distributed_switch_port = vsphere_distributed_virtual_switch.d1.id
	  distributed_port_group  = vsphere_distributed_port_group.p1.id
	  %s
	}
	`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
	),
		os.Getenv("TF_VAR_VSPHERE_HOST_NIC1"),
		netConfig)
}

func combineSnippets(snippets ...string) string {
	out := ""
	for _, snippet := range snippets {
		out = fmt.Sprintf("%s\n%s", out, snippet)
	}
	return out
}

func ipv4Snippet(payload string) string {
	switch payload {
	case "dhcp":
		return ipv4DHCPSnippet()
	case "":
		return ""
	default:
		addrBits := strings.Split(payload, "|")
		ip := addrBits[0]
		netmask := addrBits[1]
		gw := addrBits[2]
		return ipv4StaticSnippet(ip, netmask, gw)
	}
}

func ipv4StaticSnippet(ip, netmask, gw string) string {
	return fmt.Sprintf(`
	  ipv4 {
        ip = "%s"
        netmask = "%s"
		gw = "%s"
      }
	`, ip, netmask, gw)
}

func ipv4DHCPSnippet() string {
	return `
	  ipv4 {
	    dhcp = true
	  }`
}

func ipv6Snippet(payload string) string {
	switch payload {
	case "dhcp":
		return ipv6DHCPSnippet()
	case "autoconfig":
		return ipv6AutoconfigSnippet()
	case "":
		return ""
	default:
		addrBits := strings.Split(payload, "|")
		ip := addrBits[0]
		gw := addrBits[1]
		return ipv6StaticSnippet(ip, gw)
	}
}

func ipv6DHCPSnippet() string {
	return `
	  ipv6 {
	    dhcp = true
	  }`
}

func ipv6AutoconfigSnippet() string {
	return `
	  ipv6 {
		autoconfig = true
      }`
}

func ipv6StaticSnippet(ip, gw string) string {
	return fmt.Sprintf(`
	  ipv6 {
        addresses = ["%s"]
        gw = "%s"
      }`, ip, gw)
}

func netstackSnippet(stack string) string {
	if stack == "" {
		stack = "defaultTcpipStack"
	}
	return fmt.Sprintf(`
	  netstack = "%s"
	`, stack)
}
