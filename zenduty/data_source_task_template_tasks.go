package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTaskTemplateTasks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTaskTemplateTasksRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"task_template_id": {
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
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"due_in": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"position": {
							Type:     schema.TypeInt,
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

func dataSourceTaskTemplateTasksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)
	taskTemplateID := d.Get("task_template_id").(string)

	tasks, err := apiclient.TaskTemplate.GetTaskTemplateTasks(teamID, taskTemplateID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		items[i] = map[string]interface{}{
			"unique_id":     task.UniqueID,
			"title":         task.Title,
			"description":   task.Description,
			"role":          task.Role,
			"due_in":        task.DueIn,
			"position":      task.Position,
			"creation_date": task.CreationDate,
		}
	}

	if err := d.Set("tasks", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s", teamID, taskTemplateID))

	return diags
}
