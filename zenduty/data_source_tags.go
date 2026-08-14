package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTagsRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"tags": &schema.Schema{
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

func dataSourceTagsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)

	tags, err := apiclient.Tags.GetTags(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(tags))
	for i, tag := range tags {
		item := make(map[string]interface{})
		item["team"] = tag.Team
		item["unique_id"] = tag.UniqueID
		item["name"] = tag.Name
		item["color"] = tag.Color
		item["creation_date"] = tag.CreationDate
		items[i] = item
	}

	if err := d.Set("tags", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags

}
