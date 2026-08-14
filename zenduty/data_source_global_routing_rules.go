package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGlobalRoutingRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalRoutingRulesRead,
		Schema: map[string]*schema.Schema{
			"router_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UniqueID of the global router to query rules for",
			},
			"rule_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The UniqueID of the specific routing rule to query",
			},
			"routing_rules": {
				Type:        schema.TypeList,
				Description: "List of global routing rules",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:        schema.TypeString,
							Description: "The UniqueID of the routing rule",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the routing rule",
							Computed:    true,
						},
						"rule_json": {
							Type:        schema.TypeString,
							Description: "The JSON rule configuration",
							Computed:    true,
						},
						"actions": {
							Type:        schema.TypeList,
							Description: "The list of actions for this rule",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action_type": {
										Type:        schema.TypeInt,
										Description: "The type of action (0 for integration, 1 for other)",
										Computed:    true,
									},
									"integration": {
										Type:        schema.TypeString,
										Description: "The integration ID for action_type 0",
										Computed:    true,
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

func dataSourceGlobalRoutingRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	routerID := d.Get("router_id").(string)
	ruleID := d.Get("rule_id").(string)

	if ruleID != "" {
		// Get specific rule
		var diags diag.Diagnostics
		rule, err := apiclient.GlobalRouter.GetGlobalRoutingRule(routerID, ruleID)
		if err != nil {
			return diag.FromErr(err)
		}

		items := make([]map[string]interface{}, 1)
		item := make(map[string]interface{})
		item["unique_id"] = rule.UniqueID
		item["name"] = rule.Name
		item["rule_json"] = rule.RuleJSON

		// Flatten actions
		actions := make([]map[string]interface{}, len(rule.Actions))
		for j, action := range rule.Actions {
			actions[j] = map[string]interface{}{
				"action_type": action.ActionType,
				"integration": action.IntegrationObject.UniqueID,
			}
		}
		item["actions"] = actions

		items[0] = item
		if err := d.Set("routing_rules", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s", routerID, ruleID))

		return diags
	} else {
		// Get all rules for the router
		var diags diag.Diagnostics

		rules, err := apiclient.GlobalRouter.GetGlobalRoutingRules(routerID)
		if err != nil {
			return diag.FromErr(err)
		}

		items := make([]map[string]interface{}, len(rules))
		for i, rule := range rules {
			item := make(map[string]interface{})
			item["unique_id"] = rule.UniqueID
			item["name"] = rule.Name
			item["rule_json"] = rule.RuleJSON

			// Flatten actions
			actions := make([]map[string]interface{}, len(rule.Actions))
			for j, action := range rule.Actions {
				actions[j] = map[string]interface{}{
					"action_type": action.ActionType,
					"integration": action.IntegrationObject.UniqueID,
				}
			}
			item["actions"] = actions

			items[i] = item
		}

		if err := d.Set("routing_rules", items); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(fmt.Sprintf("%s/%s", routerID, ruleID))
		return diags
	}
}
