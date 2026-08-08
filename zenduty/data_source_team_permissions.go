package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamPermissions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamPermissionsRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceTeamPermissionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)

	teamPermissions, err := apiclient.Teams.GetTeamLevelPermissions(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("permissions", flattenPermissions(teamPermissions.Permissions)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(teamID)

	return diags
}
