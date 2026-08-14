package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePostIncidentTasks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostIncidentTasksRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"tasks": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"assigned_to": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"due_in_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePostIncidentTasksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)

	tasks, err := apiclient.PostIncidentTask.GetAllPostIncidentTasks(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		item := map[string]interface{}{
			"unique_id":     task.UniqueID,
			"title":         task.Title,
			"description":   task.Description,
			"status":        task.Status,
			"assigned_to":   task.AssignedTo,
			"creation_date": task.CreationDate,
		}
		if task.DueInTime != nil {
			item["due_in_time"] = *task.DueInTime
		}
		items[i] = item
	}

	if err := d.Set("tasks", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags
}
