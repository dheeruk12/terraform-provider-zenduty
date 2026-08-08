package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTaskTemplates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTaskTemplatesRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"task_templates": &schema.Schema{
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
						"summary": {
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

func dataSourceTaskTemplatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)

	templates, err := apiclient.TaskTemplate.GetTaskTemplates(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(templates))
	for i, template := range templates {
		items[i] = map[string]interface{}{
			"unique_id":     template.UniqueID,
			"name":          template.Name,
			"summary":       template.Summary,
			"creation_date": template.CreationDate,
		}
	}

	if err := d.Set("task_templates", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags
}
