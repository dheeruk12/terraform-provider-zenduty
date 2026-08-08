package zenduty

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGlobalRoutingRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateRoutingRules,
		UpdateContext: resourceUpdateRoutingRules,
		DeleteContext: resourceDeleteRoutingRules,
		ReadContext:   wrapReadWith404(resourceReadRoutingRules),
		Importer: &schema.ResourceImporter{
			State: resourceRouterRulesImporter,
		},
		Schema: map[string]*schema.Schema{
			"router_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_json": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"position": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Evaluation order of the rule within the router. Assigned by the server when omitted.",
			},
			"actions": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_type": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 1),
						},
						"integration": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
func CreateRoutingRuleAction(Ctx context.Context, d *schema.ResourceData, m interface{}, newAlertRule *client.GlobalRoutingRule) ([]client.GlobalRoutingRuleAction, diag.Diagnostics) {
	actions := d.Get("actions").([]interface{})
	newAlertRule.Actions = make([]client.GlobalRoutingRuleAction, len(actions))
	for i, action := range actions {
		ruleMap := action.(map[string]interface{})
		newAction := client.GlobalRoutingRuleAction{}

		if v, ok := ruleMap["action_type"]; ok {
			newAction.ActionType = v.(int)
		}

		if newAction.ActionType == 0 {
			if v, ok := ruleMap["integration"]; ok {
				newAction.Integration = v.(string)
			}
			if newAction.Integration == "" {
				return nil, diag.FromErr(errors.New("integration is required"))
			}

		}

		newAlertRule.Actions[i] = newAction

	}
	return newAlertRule.Actions, nil

}

func ValidateAndCreateRoutingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.GlobalRoutingRule, diag.Diagnostics) {
	newAlertRule := &client.GlobalRoutingRule{}

	if v, ok := d.GetOk("name"); ok {
		newAlertRule.Name = v.(string)

	}
	if v, ok := d.GetOk("rule_json"); ok {
		newAlertRule.RuleJSON = v.(string)

	}
	if newAlertRule.Name == "" {
		return nil, diag.FromErr(errors.New("name is required"))
	}
	if newAlertRule.RuleJSON == "" {
		return nil, diag.FromErr(errors.New("rule_json is required"))
	}
	if !isJSONString(newAlertRule.RuleJSON) {
		return nil, diag.FromErr(errors.New("rule_json is not valid JSON"))
	}
	newAlertRule.Position = d.Get("position").(int)
	actions, actionErr := CreateRoutingRuleAction(Ctx, d, m, newAlertRule)
	if actionErr != nil {
		return nil, actionErr
	}

	newAlertRule.Actions = actions

	return newAlertRule, nil
}

func resourceCreateRoutingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var routerID string

	newRule, ruleErr := ValidateAndCreateRoutingRules(Ctx, d, m)
	if ruleErr != nil {
		return ruleErr
	}
	if v, ok := d.GetOk("router_id"); ok {
		routerID = v.(string)
	}

	alertRule, err := apiclient.GlobalRouter.CreateGlobalRoutingRule(routerID, newRule)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(alertRule.UniqueID)
	return diags
}

func resourceUpdateRoutingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics

	var routerID string
	if v, ok := d.GetOk("router_id"); ok {
		routerID = v.(string)
	}

	newRule, ruleErr := ValidateAndCreateRoutingRules(Ctx, d, m)
	if ruleErr != nil {
		return ruleErr
	}

	_, err := apiclient.GlobalRouter.UpdateGlobalRoutingRule(routerID, d.Id(), newRule)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceReadRoutingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics

	routerID := d.Get("router_id").(string)
	rule, err := apiclient.GlobalRouter.GetGlobalRoutingRule(routerID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(rule.UniqueID)
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
	d.Set("actions", flattenRoutingActions(rule))
	d.Set("name", rule.Name)
	d.Set("position", rule.Position)

	return diags
}
func flattenRoutingActions(rule *client.GlobalRoutingRule) []map[string]interface{} {
	var actionsList []map[string]interface{}
	for _, action := range rule.Actions {
		newAction := map[string]interface{}{}
		newAction["action_type"] = action.ActionType
		newAction["integration"] = action.IntegrationObject.UniqueID
		actionsList = append(actionsList, newAction)
	}
	return actionsList
}

func resourceDeleteRoutingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	apiclient, _ := m.(*Config).Client()
	routerID, uniqueID := d.Get("router_id").(string), d.Id()

	err := apiclient.GlobalRouter.DeleteGlobalRoutingRule(routerID, uniqueID)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRouterRulesImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <router_id>/<rule_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid router_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid rule_id (%q)", parts[1])
	}
	d.SetId(parts[1])
	d.Set("router_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
