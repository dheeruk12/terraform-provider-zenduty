package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAlertRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAlertRulesRead,
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"integration_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alert_rule_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"alertrules": {
				Type:        schema.TypeList,
				Description: "List of Alert Rules",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"position": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"rule_json": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rule_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"stop": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"conditions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"unique_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"alert_condition_type": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"alert_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"pattern": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"actions": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"unique_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"action_type": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"escalation_policy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"assign_to": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"schedule": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"sla": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"team_priority": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceAlertRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	serviceID := d.Get("service_id").(string)
	integrationID := d.Get("integration_id").(string)
	alertRuleID := d.Get("alert_rule_id").(string)

	var diags diag.Diagnostics
	if alertRuleID != "" {
		rule, err := apiclient.AlertRules.GetAlertRule(teamID, serviceID, integrationID, alertRuleID)
		if err != nil {
			return diag.FromErr(err)
		}
		items := make([]map[string]interface{}, 1)

		item := make(map[string]interface{})

		item["position"] = rule.Position
		item["rule_json"] = rule.RuleJSON
		item["rule_type"] = rule.RuleType
		item["stop"] = rule.Stop
		item["unique_id"] = rule.UniqueID
		item["description"] = rule.Description
		item["conditions"] = flattenAlertRuleConditions(rule.Conditions)
		actions := make([]map[string]interface{}, len(rule.Actions))
		for j, action := range rule.Actions {
			rule := make(map[string]interface{})

			rule["unique_id"] = action.UniqueID
			rule["action_type"] = action.ActionType
			rule["key"] = action.Key
			rule["value"] = action.Value
			rule["escalation_policy"] = action.EscalationPolicy
			rule["assign_to"] = action.AssignedTo
			rule["schedule"] = action.Schedule
			rule["sla"] = action.SLA
			rule["team_priority"] = action.TeamPriority
			actions[j] = rule
		}
		item["actions"] = actions
		items[0] = item

		if err := d.Set("alertrules", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", teamID, serviceID, integrationID, alertRuleID))

		return diags
	} else {

		rules, err := apiclient.AlertRules.GetAlertRules(teamID, serviceID, integrationID)
		if err != nil {
			return diag.FromErr(err)
		}
		items := make([]map[string]interface{}, len(rules))
		for i, rule := range rules {
			item := make(map[string]interface{})

			item["position"] = rule.Position
			item["rule_json"] = rule.RuleJSON
			item["rule_type"] = rule.RuleType
			item["stop"] = rule.Stop
			item["unique_id"] = rule.UniqueID
			item["description"] = rule.Description
			item["conditions"] = flattenAlertRuleConditions(rule.Conditions)
			actions := make([]map[string]interface{}, len(rule.Actions))
			for j, action := range rule.Actions {
				rule := make(map[string]interface{})

				rule["unique_id"] = action.UniqueID
				rule["action_type"] = action.ActionType
				rule["key"] = action.Key
				rule["value"] = action.Value
				rule["escalation_policy"] = action.EscalationPolicy
				rule["assign_to"] = action.AssignedTo
				rule["schedule"] = action.Schedule
				rule["sla"] = action.SLA
				rule["team_priority"] = action.TeamPriority
				actions[j] = rule
			}
			item["actions"] = actions
			items[i] = item
		}
		if err := d.Set("alertrules", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", teamID, serviceID, integrationID, alertRuleID))

		return diags
	}
}
