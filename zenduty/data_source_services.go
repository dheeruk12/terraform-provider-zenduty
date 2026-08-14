package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServicesRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"services": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auto_resolve_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"created_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"acknowledgement_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"under_maintenance": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"escalation_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"summary": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"collation": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"collation_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"sla": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_template": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team_priority": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceServicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Get("service_id").(string)
	if teamID == "" {
		return diag.Errorf("team_id is required")
	}
	var diags diag.Diagnostics
	if id != "" {
		service, err := apiclient.Services.GetServicesByID(teamID, id)
		if err != nil {
			return diag.FromErr(err)
		}
		items := make([]map[string]interface{}, 1)

		items[0] = map[string]interface{}{
			"name":                    service.Name,
			"escalation_policy":       service.EscalationPolicy,
			"team":                    service.Team,
			"description":             service.Description,
			"summary":                 service.Summary,
			"collation":               service.Collation,
			"collation_time":          service.CollationTime,
			"sla":                     service.SLA,
			"task_template":           service.TaskTemplate,
			"team_priority":           service.TeamPriority,
			"creation_date":           service.CreationDate,
			"unique_id":               service.UniqueID,
			"auto_resolve_timeout":    service.AutoResolveTimeout,
			"created_by":              service.CreatedBy,
			"acknowledgement_timeout": service.AcknowledgmentTimeout,
			"status":                  service.Status,
			"under_maintenance":       service.UnderMaintenance,
		}
		if err := d.Set("services", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s", teamID, id))
		return diags
	} else {

		// id := d.Get("id").(string)
		services, err := apiclient.Services.GetServices(teamID)
		if err != nil {
			return diag.FromErr(err)
		}
		items := make([]map[string]interface{}, len(services))
		for i, service := range services {
			items[i] = map[string]interface{}{
				"name":                    service.Name,
				"escalation_policy":       service.EscalationPolicy,
				"team":                    service.Team,
				"description":             service.Description,
				"summary":                 service.Summary,
				"collation":               service.Collation,
				"collation_time":          service.CollationTime,
				"sla":                     service.SLA,
				"task_template":           service.TaskTemplate,
				"team_priority":           service.TeamPriority,
				"creation_date":           service.CreationDate,
				"unique_id":               service.UniqueID,
				"auto_resolve_timeout":    service.AutoResolveTimeout,
				"created_by":              service.CreatedBy,
				"acknowledgement_timeout": service.AcknowledgmentTimeout,
				"status":                  service.Status,
				"under_maintenance":       service.UnderMaintenance,
			}
		}
		if err := d.Set("services", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s", teamID, id))

		return diags
	}
}
