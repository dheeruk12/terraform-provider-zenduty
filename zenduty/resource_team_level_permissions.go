package zenduty

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamLevelPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateTeamLeveLPermissions,
		ReadContext:   wrapReadWith404(resourceReadTeamLeveLPermissions),
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
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func validateTeamLeveLPermissionss(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.TeamLevelPermissions, diag.Diagnostics) {
	permissionsList := []string{
		"analytics_read",
		"escalation_policy_attach",
		"escalation_policy_read",
		"incident_read",
		"incident_role_read",
		"incident_write",
		"integration_read",
		"maintenance_read",
		"member_read",
		"post_incident_task_read",
		"postmortem_read",
		"priority_read",
		"schedule_attach",
		"schedule_read",
		"service_read",
		"sla_read",
		"stakeholder_template_read",
		"tag_read",
		"task_template_read",
		"team_read",
	}

	permissions := d.Get("permissions").([]interface{})
	newPermission := &client.TeamLevelPermissions{}
	team_id := d.Get("team_id").(string)
	newPermission.UniqueID = team_id
	for _, permission := range permissions {
		if permission.(string) == "" {
			return nil, diag.FromErr(errors.New("permission must not be empty"))
		}
		if !checkList(permission.(string), permissionsList) {
			return nil, diag.FromErr(fmt.Errorf("invalid permission received %s", permission.(string)))
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
		return diag.FromErr(err)
	}
	d.SetId(teamPermissions.UniqueID)
	d.Set("team_id", teamPermissions.UniqueID)
	d.Set("permissions", flattenPermissions(teamPermissions.Permissions))
	return nil
}
