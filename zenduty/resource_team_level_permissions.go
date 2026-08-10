package zenduty

import (
	"context"
	"errors"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamLevelPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateTeamLeveLPermissions,
		ReadContext:   resourceReadTeamLeveLPermissions,
		UpdateContext: resourceUpdateTeamLeveLPermissions,
		DeleteContext: resourceDeleteTeamLeveLPermissions,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"permissions": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Permissions for the team. The API automatically adds the read permissions implied by the ones you list (e.g. incident_read pulls in team_read, service_read, ...); list the full stored set to avoid plan diffs.",
			},
		},
	}
}

func validateTeamLeveLPermissionss(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.TeamLevelPermissions, diag.Diagnostics) {
	permissions := d.Get("permissions").(*schema.Set).List()
	newPermission := &client.TeamLevelPermissions{}
	team_id := d.Get("team_id").(string)
	newPermission.UniqueID = team_id
	// The permission catalogue grows server-side and there is no endpoint to
	// fetch it, so unknown values are left to the API to reject.
	for _, permission := range permissions {
		if permission.(string) == "" {
			return nil, diag.FromErr(errors.New("permission must not be empty"))
		}
		newPermission.Permissions = append(newPermission.Permissions, permission.(string))
	}

	return newPermission, nil
}

func resourceCreateTeamLeveLPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newPermissions, validationerr := validateTeamLeveLPermissionss(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	updatedPermission, err := apiclient.Teams.UpdateTeamLevelPermissions(newPermissions.UniqueID, newPermissions)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(updatedPermission.UniqueID)
	return nil
}

func resourceUpdateTeamLeveLPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newPermissions, validationerr := validateTeamLeveLPermissionss(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	updatedPermission, err := apiclient.Teams.UpdateTeamLevelPermissions(newPermissions.UniqueID, newPermissions)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(updatedPermission.UniqueID)
	return nil
}

func resourceDeleteTeamLeveLPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newPermission := &client.TeamLevelPermissions{}
	newPermission.Permissions = []string{}

	_, err := apiclient.Teams.UpdateTeamLevelPermissions(d.Id(), newPermission)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceReadTeamLeveLPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	teamPermissions, err := apiclient.Teams.GetTeamLevelPermissions(d.Id())
	if err != nil {
		return handleReadError(d, err)
	}
	d.SetId(teamPermissions.UniqueID)
	d.Set("team_id", teamPermissions.UniqueID)
	d.Set("permissions", flattenPermissions(teamPermissions.Permissions))
	return nil
}
