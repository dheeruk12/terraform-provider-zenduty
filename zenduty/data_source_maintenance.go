package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceManintenanceRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"maintenance_windows": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"repeat_interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"repeat_until": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"services": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"unique_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"service": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"time_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceManintenanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	teamID := d.Get("team_id").(string)

	var diags diag.Diagnostics

	maintenances, err := apiclient.MaintenanceWindow.GetMaintenanceWindows(teamID)
	if err != nil {
		return diag.FromErr(err)
	}
	items := make([]map[string]interface{}, len(maintenances))

	for i, maintenance := range maintenances {
		item := make(map[string]interface{})
		item["unique_id"] = maintenance.UniqueID
		item["name"] = maintenance.Name
		item["creation_date"] = maintenance.CreationDate
		item["start_time"] = maintenance.StartTime
		item["end_time"] = maintenance.EndTime
		item["repeat_interval"] = maintenance.RepeatInterval
		item["repeat_until"] = maintenance.RepeatUntil
		item["time_zone"] = maintenance.TimeZone
		services := make([]map[string]interface{}, len(maintenance.Services))
		for j, service := range maintenance.Services {
			services[j] = map[string]interface{}{
				"unique_id": service.UniqueID,
				"service":   service.Service,
			}
		}
		item["services"] = services

		items[i] = item

	}
	if err := d.Set("maintenance_windows", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags

}
