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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

			"stop": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Stop processing further alert rules when this rule matches.",
			},
			"rule_type": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 1),
				Description:  "0 requires all conditions to match, 1 requires any condition to match.",
			},
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"actions": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_type": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"escalation_policy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: ValidateUUID(),
							Description:      "Escalation policy to assign (action_type 4). Alternative to value.",
						},
						"schedule": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: ValidateUUID(),
							Description:      "Schedule to assign (action_type 5). Alternative to value.",
						},
						"sla": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: ValidateUUID(),
							Description:      "SLA to assign (action_type 14). Alternative to value.",
						},
						"team_priority": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: ValidateUUID(),
							Description:      "Team priority to assign (action_type 15). Alternative to value.",
						},
						"task_template": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: ValidateUUID(),
							Description:      "Task template to assign (action_type 16). Alternative to value.",
						},
						"assign_to": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "User to assign (action_type 6). Alternative to value.",
						},
					},
				},
			},
		},
	}
}

// actionTypeAttr maps action types that reference another Zenduty object to
// the dedicated schema attribute for that object. These types accept either
// the dedicated attribute or the legacy overloaded "value" attribute.
var actionTypeAttr = map[int]string{
	4:  "escalation_policy",
	5:  "schedule",
	6:  "assign_to",
	14: "sla",
	15: "team_priority",
	16: "task_template",
}

// actionTypesWithoutValue take no payload at all.
var actionTypesWithoutValue = map[int]bool{
	3:  true, // suppress alert
	18: true, // change entity id hash
}

func AlertRuleAction(Ctx context.Context, d *schema.ResourceData, m interface{}, newAlertRule *client.AlertRule) ([]client.AlertAction, diag.Diagnostics) {
	actions := d.Get("actions").([]interface{})
	newAlertRule.Actions = make([]client.AlertAction, len(actions))
	for i, action := range actions {
		ruleMap := action.(map[string]interface{})
		newAction := client.AlertAction{}

		if v, ok := ruleMap["action_type"]; ok {
			newAction.ActionType = v.(int)
		}
		// New action types appear server-side over time; only the lower bound
		// is enforced here and unknown types pass through via "value".
		if newAction.ActionType < 1 {
			return nil, diag.FromErr(errors.New("action_type is not valid"))
		}

		value, _ := ruleMap["value"].(string)

		// Resolve the effective value: the type-specific attribute wins, the
		// legacy "value" attribute is kept for backwards compatibility.
		if attr, hasAttr := actionTypeAttr[newAction.ActionType]; hasAttr {
			attrValue, _ := ruleMap[attr].(string)
			if attrValue != "" && value != "" {
				return nil, diag.FromErr(fmt.Errorf("action %d: %s and value are mutually exclusive", i, attr))
			}
			if attrValue != "" {
				value = attrValue
			}
		}

		if !actionTypesWithoutValue[newAction.ActionType] && value == "" {
			attr := actionTypeAttr[newAction.ActionType]
			if attr != "" {
				return nil, diag.FromErr(fmt.Errorf("action %d: %s (or value) is required", i, attr))
			}
			return nil, diag.FromErr(fmt.Errorf("action %d: value is required", i))
		}

		switch newAction.ActionType {
		case 1: // change alert type
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, diag.FromErr(fmt.Errorf("action %d: value %q is not a number", i, value))
			}
			if n < 0 || n > 5 {
				return nil, diag.FromErr(fmt.Errorf("action %d: value should be between 0 and 5", i))
			}
			newAction.Value = value
		case 3, 18: // no payload
		case 4: // assign escalation policy
			if !IsValidUUID(value) {
				return nil, diag.FromErr(fmt.Errorf("action %d: escalation_policy %q is not a valid UUID", i, value))
			}
			newAction.EscalationPolicy = value
		case 5: // assign schedule
			if !IsValidUUID(value) {
				return nil, diag.FromErr(fmt.Errorf("action %d: schedule %q is not a valid UUID", i, value))
			}
			newAction.Schedule = value
		case 6: // assign user
			newAction.AssignedTo = value
		case 7: // change urgency
			if value != "0" && value != "1" {
				return nil, diag.FromErr(fmt.Errorf("action %d: incident urgency should be 0 or 1", i))
			}
			newAction.Value = value
		case 11: // assign incident role to user
			key, _ := ruleMap["key"].(string)
			if key == "" {
				return nil, diag.FromErr(fmt.Errorf("action %d: key (the role id) is required", i))
			}
			if !IsValidUUID(key) {
				return nil, diag.FromErr(fmt.Errorf("action %d: key (the role id) is not a valid UUID", i))
			}
			newAction.Key = key
			newAction.Value = value
		case 14: // assign SLA
			if !IsValidUUID(value) {
				return nil, diag.FromErr(fmt.Errorf("action %d: sla %q is not a valid UUID", i, value))
			}
			newAction.SLA = value
		case 15: // assign team priority
			if !IsValidUUID(value) {
				return nil, diag.FromErr(fmt.Errorf("action %d: team_priority %q is not a valid UUID", i, value))
			}
			newAction.TeamPriority = value
		case 16: // assign task template
			if !IsValidUUID(value) {
				return nil, diag.FromErr(fmt.Errorf("action %d: task_template %q is not a valid UUID", i, value))
			}
			newAction.TaskTemplates = value
		default:
			newAction.Value = value
		}

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
	newAlertRule.Stop = d.Get("stop").(bool)
	newAlertRule.RuleType = d.Get("rule_type").(int)
	newAlertRule.Position = d.Get("position").(int)
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

	d.Set("actions", flattenAlertActions(rule, d.Get("actions").([]interface{})))
	d.Set("description", rule.Description)
	d.Set("stop", rule.Stop)
	d.Set("rule_type", rule.RuleType)
	d.Set("position", rule.Position)

	return diags
}

// flattenAlertActions maps API actions back to state. For action types that
// accept either the dedicated attribute or the legacy "value" attribute, the
// API value is written into whichever shape the existing state (i.e. the
// config) used, so switching representations never produces a phantom diff.
func flattenAlertActions(rule *client.AlertRule, prior []interface{}) []map[string]interface{} {
	var actionsList []map[string]interface{}
	for i, action := range rule.Actions {
		newAction := map[string]interface{}{}
		newAction["action_type"] = action.ActionType

		var apiValue string
		switch action.ActionType {
		case 4:
			apiValue = action.EscalationPolicy
		case 5:
			apiValue = action.Schedule
		case 6:
			apiValue = action.AssignedTo
		case 14:
			apiValue = action.SLA
		case 15:
			apiValue = action.TeamPriority
		case 16:
			apiValue = action.TaskTemplates
		case 3, 18:
			apiValue = ""
		default:
			apiValue = action.Value
		}

		attr := actionTypeAttr[action.ActionType]
		if attr != "" && priorActionUsedAttr(prior, i, attr) {
			newAction[attr] = apiValue
		} else if action.ActionType != 3 && action.ActionType != 18 {
			newAction["value"] = apiValue
		}
		if action.ActionType == 11 {
			newAction["key"] = action.Key
		}
		actionsList = append(actionsList, newAction)
	}
	return actionsList
}

// priorActionUsedAttr reports whether the action at index i in the prior
// state carried a non-empty value for the given attribute.
func priorActionUsedAttr(prior []interface{}, i int, attr string) bool {
	if i >= len(prior) {
		return false
	}
	m, ok := prior[i].(map[string]interface{})
	if !ok {
		return false
	}
	v, _ := m[attr].(string)
	return v != ""
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
