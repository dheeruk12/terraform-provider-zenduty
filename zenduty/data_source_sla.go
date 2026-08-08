package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSLA() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSLARead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"slas": &schema.Schema{
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
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"conditions": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"acknowledge_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"resolve_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"is_active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceSLARead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)

	slas, err := apiclient.Sla.GetSLAs(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(slas))
	for i, sla := range slas {
		items[i] = map[string]interface{}{
			"unique_id":        sla.UniqueID,
			"name":             sla.Name,
			"description":      sla.Description,
			"conditions":       sla.Conditions,
			"acknowledge_time": sla.AcknowledgeTime,
			"resolve_time":     sla.ResolveTime,
			"is_active":        sla.IsActive,
		}
	}

	if err := d.Set("slas", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags
}
