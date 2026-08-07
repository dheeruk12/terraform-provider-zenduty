package zenduty

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAlertRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateAlertRules,
		UpdateContext: resourceUpdateAlertRules,
		DeleteContext: resourceDeleteAlertRules,
		ReadContext:   wrapReadWith404(resourceReadAlertRules),
		Importer: &schema.ResourceImporter{
			State: resourceAlertRulesImporter,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_json": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"service_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"integration_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},

			"actions": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_type": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
func AlertRuleAction(Ctx context.Context, d *schema.ResourceData, m interface{}, newAlertRule *client.AlertRule) ([]client.AlertAction, diag.Diagnostics) {
	actions := d.Get("actions").([]interface{})
	newAlertRule.Actions = make([]client.AlertAction, len(actions))
	for i, action := range actions {
		ruleMap := action.(map[string]interface{})
		newAction := client.AlertAction{}
		var value, key string

		if v, ok := ruleMap["action_type"]; ok {
			newAction.ActionType = v.(int)
		}

		if (newAction.ActionType > 18) || (newAction.ActionType < 1) {
			return nil, diag.FromErr(errors.New("action_type is not valid"))
		}

		if v, ok := ruleMap["value"]; ok {
			value = v.(string)
		}

		if ((newAction.ActionType != 3) && (newAction.ActionType != 18)) && (value == "") {
			return nil, diag.FromErr(errors.New("value is required"))
		}
		if ((newAction.ActionType == 4) || (newAction.ActionType == 14) || (newAction.ActionType == 15) || (newAction.ActionType == 16)) && (!IsValidUUID(value)) {
			return nil, diag.FromErr(errors.New(value + " is not a valid UUID"))
		}

		if (newAction.ActionType == 7) && (!(value == "0" || value == "1")) {
			return nil, diag.FromErr(errors.New("incident urgency should be 0 or 1"))
		}
		if newAction.ActionType == 1 {
			i, err := strconv.Atoi(value)
			if i < 0 || i > 5 {
				return nil, diag.FromErr(errors.New("value should be between 0 and 5"))
			}
			if err != nil {
				return nil, diag.FromErr(errors.New("value is not valid"))
			}
			newAction.Value = value
		} else if newAction.ActionType == 3 {
			value = ""
		} else if newAction.ActionType == 4 {
			newAction.EscalationPolicy = value
			value = ""
		} else if newAction.ActionType == 6 {

			newAction.AssignedTo = value
			value = ""
		} else if newAction.ActionType == 11 {
			if v, ok := ruleMap["key"]; ok {
				key = v.(string)
			}
			if key == "" {
				return nil, diag.FromErr(errors.New("key(ie..role_id) is required"))
			}
			if !IsValidUUID(key) {
				return nil, diag.FromErr(errors.New("key(ie..role_id) is not valid UUID"))
			}

			newAction.Key = key

		} else if newAction.ActionType == 14 {
			newAction.SLA = value
			value = ""
			if newAction.SLA == "" {
				return nil, diag.FromErr(errors.New("sla is required"))
			} else if !IsValidUUID(newAction.SLA) {
				return nil, diag.FromErr(errors.New("sla is not valid UUID"))
			}

		} else if newAction.ActionType == 15 {
			newAction.TeamPriority = value
			value = ""
			if newAction.TeamPriority == "" {
				return nil, diag.FromErr(errors.New("team_priority is required"))
			} else if !IsValidUUID(newAction.TeamPriority) {
				return nil, diag.FromErr(errors.New("team_priority is not valid UUID"))
			}
		} else if newAction.ActionType == 16 {
			newAction.TaskTemplates = value
			value = ""
		} else if newAction.ActionType == 18 {
			value = ""
		}

		newAction.Value = value

		newAlertRule.Actions[i] = newAction

	}
	return newAlertRule.Actions, nil

}

func ValidateAncCreateAlertRules(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.AlertRule, diag.Diagnostics) {
	newAlertRule := &client.AlertRule{}

	if v, ok := d.GetOk("description"); ok {
		newAlertRule.Description = v.(string)

	}
	if v, ok := d.GetOk("rule_json"); ok {
		newAlertRule.RuleJSON = v.(string)

	}
	if newAlertRule.Description == "" {
		return nil, diag.FromErr(errors.New("description is required"))
	}
	if newAlertRule.RuleJSON == "" {
		return nil, diag.FromErr(errors.New("rule_json is required"))
	}
	if !isJSONString(newAlertRule.RuleJSON) {
		return nil, diag.FromErr(errors.New("rule_json is not valid JSON"))
	}
	actions, actionErr := AlertRuleAction(Ctx, d, m, newAlertRule)
	if actionErr != nil {
		return nil, actionErr
	}

	newAlertRule.Actions = actions

	return newAlertRule, nil
}

func resourceCreateAlertRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var teamID, serviceID, integrationID string

	teamID = d.Get("team_id").(string)
	serviceID = d.Get("service_id").(string)
	integrationID = d.Get("integration_id").(string)

	newRule, ruleErr := ValidateAncCreateAlertRules(Ctx, d, m)
	if ruleErr != nil {
		return ruleErr
	}

	alertRule, err := apiclient.AlertRules.CreateAlertRule(teamID, serviceID, integrationID, newRule)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(alertRule.UniqueID)
	return diags
}

func resourceUpdateAlertRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var teamID, serviceID, integrationID string

	teamID = d.Get("team_id").(string)
	serviceID = d.Get("service_id").(string)
	integrationID = d.Get("integration_id").(string)

	newRule, ruleErr := ValidateAncCreateAlertRules(Ctx, d, m)
	if ruleErr != nil {
		return ruleErr
	}

	_, err := apiclient.AlertRules.UpdateAlertRule(teamID, serviceID, integrationID, d.Id(), newRule)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceReadAlertRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var teamID, serviceID, integrationID string

	teamID = d.Get("team_id").(string)
	serviceID = d.Get("service_id").(string)
	integrationID = d.Get("integration_id").(string)

	rule, err := apiclient.AlertRules.GetAlertRule(teamID, serviceID, integrationID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(rule.UniqueID)
	// Normalize JSON before setting to avoid formatting issues
	if rule.RuleJSON != "" {
		normalizedJSON, err := normalizeJSON(rule.RuleJSON)
		if err != nil {
			// If normalization fails, use original JSON
			d.Set("rule_json", rule.RuleJSON)
		} else {
			d.Set("rule_json", normalizedJSON)
		}
	} else {
		d.Set("rule_json", rule.RuleJSON)
	}

	d.Set("actions", flattenAlertActions(rule))
	d.Set("description", rule.Description)

	return diags
}
func flattenAlertActions(rule *client.AlertRule) []map[string]interface{} {
	var actionsList []map[string]interface{}
	for _, action := range rule.Actions {
		newAction := map[string]interface{}{}
		newAction["action_type"] = action.ActionType
		if action.ActionType != 3 {
			if action.ActionType == 4 {
				newAction["value"] = action.EscalationPolicy
			} else if action.ActionType == 6 {
				newAction["value"] = action.AssignedTo
			} else if action.ActionType == 14 {
				newAction["value"] = action.SLA
			} else if action.ActionType == 15 {
				newAction["value"] = action.TeamPriority
			} else if action.ActionType == 16 {
				newAction["value"] = action.TaskTemplates
			} else {
				newAction["value"] = action.Value
			}
		}
		if action.ActionType == 11 {
			newAction["key"] = action.Key
		}
		actionsList = append(actionsList, newAction)
	}
	return actionsList
}

func resourceDeleteAlertRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	apiclient, _ := m.(*Config).Client()
	teamID, serviceID, integrationID, uniqueID := d.Get("team_id").(string), d.Get("service_id").(string), d.Get("integration_id").(string), d.Id()

	err := apiclient.AlertRules.DeleteAlertRule(teamID, serviceID, integrationID, uniqueID)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceAlertRulesImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<service_id>/<integration_id>/<alert_rule_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid serviceid (%q)", parts[1])
	} else if !IsValidUUID(parts[2]) {
		return nil, fmt.Errorf("invalid integration_id (%q)", parts[2])
	} else if !IsValidUUID(parts[3]) {
		return nil, fmt.Errorf("invalid alert_rule_id (%q)", parts[3])
	}
	d.SetId(parts[3])
	d.Set("integration_id", parts[2])
	d.Set("team_id", parts[0])
	d.Set("service_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
