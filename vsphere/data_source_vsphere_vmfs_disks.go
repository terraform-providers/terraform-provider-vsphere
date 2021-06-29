package vsphere

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func dataSourceVSphereVmfsDisks() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereVmfsDisksRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the host to search for disks on.",
				Required:    true,
			},
			"rescan": {
				Type:        schema.TypeBool,
				Description: "Rescan the system for disks before querying. This may lengthen the time it takes to gather information.",
				Optional:    true,
			},
			"filter": {
				Type:         schema.TypeString,
				Description:  "A regular expression to filter the disks against. Only disks with canonical names that match will be included.",
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"disks": {
				Type:        schema.TypeList,
				Description: "The names of the disks discovered by the search.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"disks_info": {
				Type:        schema.TypeList,
				Description: "The details of the disks discovered by the search.",
				Computed:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Display name of the disk",
					},
					"path": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Path of the physical volume of the disk.",
					},
					"capacity_in_gb": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Capacity in GB.",
					},
				}},
			},
		},
	}
}

func dataSourceVSphereVmfsDisksRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hsID := d.Get("host_system_id").(string)
	ss, err := hostStorageSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host storage system: %s", err)
	}

	if d.Get("rescan").(bool) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()
		if err := ss.RescanAllHba(ctx); err != nil {
			return err
		}
	}

	var hss mo.HostStorageSystem
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	if err := ss.Properties(ctx, ss.Reference(), nil, &hss); err != nil {
		return fmt.Errorf("error querying storage system properties: %s", err)
	}

	d.SetId(time.Now().UTC().String())

	var disks []string
	var disksInfo []map[string]interface{}
	for _, sl := range hss.StorageDeviceInfo.ScsiLun {
		if hsd, ok := sl.(*types.HostScsiDisk); ok {
			if matched, _ := regexp.MatchString(d.Get("filter").(string), hsd.CanonicalName); matched {
				disk := make(map[string]interface{})
				disk["name"] = hsd.DisplayName
				disk["path"] = hsd.DevicePath
				block := hsd.Capacity.Block
				blockSize := int64(hsd.Capacity.BlockSize)
				disk["capacity_in_gb"] = structure.ByteToGiB(block * blockSize)
				disksInfo = append(disksInfo, disk)
				disks = append(disks, hsd.CanonicalName)
			}
		}
	}
	sort.Strings(disks)

	if err := d.Set("disks", disks); err != nil {
		return fmt.Errorf("error saving results to state: %s", err)
	}

	if err := d.Set("disks_info", disksInfo); err != nil {
		return fmt.Errorf("error saving results to state: %s", err)
	}

	return nil
}
