package zenduty

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceIncidents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIncidentRead,
		Schema: map[string]*schema.Schema{
			"number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Fetch a single incident by its incident number.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9]+$`), "must be an incident number",
				),
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter incidents by status, applied server-side: -1 open (triggered + acknowledged), 1 triggered, 2 acknowledged, 3 resolved.",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum number of incidents to fetch (the API pages 10 at a time). 0 fetches every page — use with care on accounts with a large incident history.",
			},
			"results": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"summary": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"incident_number": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"creation_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_creation_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_summary": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_auto_resolve_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_object_created_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_team_priority": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_task_template": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_acknowledgement_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_object_status": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_object_escalation_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_team": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_sla": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_object_collation_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_object_collation": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"incident_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"urgency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"merged_with": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assigned_to": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"escalation_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"escalation_policy_object_unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"escalation_policy_object_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assigned_to_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resolved_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"acknowledged_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"context_window_start": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"context_window_end": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// "tags": {
						// 	Type:     schema.TypeList,
						// 	Computed: true,
						// },
						"sla": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team_priority": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team_priority_object": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIncidentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	number := d.Get("number").(string)
	statusFilter := d.Get("status").(string)
	limit := d.Get("limit").(int)

	status := 0
	if statusFilter != "" {
		var err error
		status, err = strconv.Atoi(statusFilter)
		if err != nil || status < -1 || status > 3 {
			return diag.Errorf("status must be one of -1 (open), 1 (triggered), 2 (acknowledged), 3 (resolved), got %q", statusFilter)
		}
	}

	var results []client.Incidents
	if number != "" {
		incident, err := apiclient.Incidents.GetIncidentByNumber(number)
		if err != nil {
			return diag.FromErr(err)
		}
		results = []client.Incidents{*incident}
		if status != 0 {
			open := incident.Status == 1 || incident.Status == 2
			if incident.Status != status && !(status == -1 && open) {
				results = nil
			}
		}
	} else {
		// The list endpoint filters by status server-side and pages 10
		// results at a time; follow Next until done or the limit is reached.
		for page := 1; ; page++ {
			pageResult, err := apiclient.Incidents.ListIncidents(&client.ListIncidentsOptions{Page: page, Status: status})
			if err != nil {
				return diag.FromErr(err)
			}
			results = append(results, pageResult.Results...)
			if pageResult.Next == "" || len(pageResult.Results) == 0 || (limit > 0 && len(results) >= limit) {
				break
			}
		}
		if limit > 0 && len(results) > limit {
			results = results[:limit]
		}
	}

	items := make([]map[string]interface{}, len(results))
	for i, result := range results {

		item := make(map[string]interface{})
		item["summary"] = result.Summary
		item["incident_number"] = result.IncidentNumber
		item["creation_date"] = result.CreationDate
		item["status"] = result.Status
		item["unique_id"] = result.UniqueID
		item["service_object_name"] = result.ServiceObject.Name
		item["service_object_unique_id"] = result.ServiceObject.UniqueID
		item["service_object_creation_date"] = result.ServiceObject.CreationDate
		item["service_object_status"] = result.ServiceObject.Status
		item["service_object_team"] = result.ServiceObject.Team
		item["service_object_summary"] = result.ServiceObject.Summary
		item["service_object_description"] = result.ServiceObject.Description
		item["service_object_acknowledgement_timeout"] = result.ServiceObject.AcknowledgmentTimeout
		item["service_object_auto_resolve_timeout"] = result.ServiceObject.AutoResolveTimeouts
		item["service_object_created_by"] = result.ServiceObject.CreatedBy
		item["service_object_team_priority"] = result.ServiceObject.TeamPriority
		item["service_object_task_template"] = result.ServiceObject.TaskTemplate
		item["service_object_escalation_policy"] = result.ServiceObject.EscalationPolicy
		item["service_object_sla"] = result.ServiceObject.SLA
		item["service_object_collation_time"] = result.ServiceObject.CollationTime
		item["service_object_collation"] = result.ServiceObject.Collation
		item["team_priority"] = result.TeamPriority
		item["team_priority_object"] = result.TeamPriorityObject
		item["title"] = result.Title
		item["incident_key"] = result.IncidentKey
		item["service"] = result.Service
		item["urgency"] = result.Urgency
		item["merged_with"] = result.MergedWith
		item["assigned_to"] = result.AssignedTo
		item["escalation_policy"] = result.EscalationPolicy
		item["escalation_policy_object_unique_id"] = result.EscalationPolicyObject.UniqueID
		item["escalation_policy_object_name"] = result.EscalationPolicyObject.Name
		item["assigned_to_name"] = result.AssignedToName
		item["resolved_date"] = result.ResolvedDate
		item["acknowledged_date"] = result.AcknowledgedDate
		item["context_window_start"] = result.ContextWindowStart
		item["context_window_end"] = result.ContextWindowEnd
		item["sla"] = result.SLA
		items[i] = item
	}
	if err := d.Set("results", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("incidents/%s/%s/%d", number, statusFilter, limit))
	return diags

}
