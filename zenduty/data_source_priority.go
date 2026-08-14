package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePriorities() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePriorityRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"priorities": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"team": {
							Type:     schema.TypeString,
							Computed: true,
						},
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
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"color": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePriorityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)

	priorities, err := apiclient.Priority.GetPriority(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(priorities))
	for i, priority := range priorities {
		item := make(map[string]interface{})
		item["team"] = priority.Team
		item["unique_id"] = priority.UniqueID
		item["name"] = priority.Name
		item["color"] = priority.Color
		item["creation_date"] = priority.CreationDate
		item["description"] = priority.Description
		items[i] = item
	}

	if err := d.Set("priorities", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags

}
