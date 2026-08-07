package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Your Zenduty API key",
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENDUTY_API_KEY", nil),
			},
			"base_url": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The base url of the Zenduty",
				Optional:    true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zenduty_teams":               resourceTeam(),
			"zenduty_roles":               resourceRoles(),
			"zenduty_services":            resourceServices(),
			"zenduty_integrations":        resourceIntegrations(),
			"zenduty_schedules":           resourceSchedules(),
			"zenduty_esp":                 resourceEsp(),
			"zenduty_incidents":           resourceIncidents(),
			"zenduty_invite":              resourceInvite(),
			"zenduty_member":              resourceMembers(),
			"zenduty_alertrules":          resourceAlertRules(),
			"zenduty_tags":                resourceTags(),
			"zenduty_priorities":          resourcePriority(),
			"zenduty_maintenance_window":  resourceMaintenanceWindow(),
			"zenduty_notification_rules":  resourceNotificationRules(),
			"zenduty_user":                resourceUser(),
			"zenduty_account_role":        resourceAccountRole(),
			"zenduty_assign_account_role": resourceAssignAccountRole(),
			"zenduty_globalrouter":        resourceGlobalRouter(),
			"zenduty_globalrouting_rule":  resourceGlobalRoutingRules(),
			"zenduty_sla":                 resourceSLA(),
			"zenduty_post_incident_tasks": resourcePostIncidentTasks(),
			"zenduty_task_templates":      resourceTaskTemplates(),
			"zenduty_task_template_tasks": resourceTaskTemplateTaskTasks(),
			"zenduty_team_permissions":    resourceTeamLevelPermissions(),
			"zenduty_outgoing_rules":      resourceOutgoingRules(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"zenduty_teams":                dataSourceTeams(),
			"zenduty_roles":                dataSourceRoles(),
			"zenduty_incidents":            dataSourceIncidents(),
			"zenduty_services":             dataSourceServices(),
			"zenduty_integrations":         dataSourceIntegrations(),
			"zenduty_schedules":            dataSourceSchedules(),
			"zenduty_esp":                  dataSourceEsp(),
			"zenduty_user":                 dataSourceUsers(),
			"zenduty_alertrules":           dataSourceAlertRules(),
			"zenduty_tags":                 dataSourceTags(),
			"zenduty_priorities":           dataSourcePriorities(),
			"zenduty_maintenance_window":   dataSourceMaintenanceWindow(),
			"zenduty_usercontact":          dataSourceUserContacts(),
			"zenduty_globalrouter":         dataSourceGlobalRouter(),
			"zenduty_global_routing_rules": dataSourceGlobalRoutingRules(),
			"zenduty_members":              dataSourceMembers(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	baseURL := d.Get("base_url").(string)
	var diags diag.Diagnostics
	if token != "" {
		client := Config{
			Token:   token,
			BaseURL: baseURL,
		}
		return &client, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create zenduty client provider",
		Detail:   "Unable to auth user for authenticated zenduty client",
	})

	return nil, diags
}
