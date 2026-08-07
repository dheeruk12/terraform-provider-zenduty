package zenduty

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceEsp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateEsp,
		UpdateContext: resourceUpdateEsp,
		DeleteContext: resourceDeleteEsp,
		ReadContext:   wrapReadWith404(resourceReadEsp),
		Importer: &schema.ResourceImporter{
			State: resourceEscalationPolicyImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rules": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delay": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"position": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"targets": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_type": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"target_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"position": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"repeat_policy": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"move_to_next": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func CreateEsp(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.EscalationPolicy, diag.Diagnostics) {
	newEsp := &client.EscalationPolicy{}
	rules := d.Get("rules").([]interface{})
	if v, ok := d.GetOk("name"); ok {
		newEsp.Name = v.(string)
	}
	if v, ok := d.GetOk("summary"); ok {
		newEsp.Summary = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		newEsp.Description = v.(string)
		if emptyString(newEsp.Description) {
			return nil, diag.FromErr(errors.New("description is empty"))
		}
	}
	if v, ok := d.GetOk("team_id"); ok {
		newEsp.Team = v.(string)
	}
	if v, ok := d.GetOk("repeat_policy"); ok {
		newEsp.RepeatPolicy = v.(int)
	}
	if v, ok := d.GetOk("move_to_next"); ok {
		newEsp.MoveToNext = v.(bool)
	}
	newEsp.Rules = make([]client.Rules, len(rules))
	oldDelay := 0
	for i, rule := range rules {

		ruleMap := rule.(map[string]interface{})
		newRule := client.Rules{}
		if v, ok := ruleMap["delay"]; ok {
			if i == 0 && newRule.Delay != 0 {
				return nil, diag.FromErr(errors.New("delay is not 0 for first rule"))
			}
			newRule.Delay = v.(int)
			if newRule.Delay < oldDelay && i != 0 {
				return nil, diag.Errorf("delay must be greater than previous %d should be greater than %d", newRule.Delay, oldDelay)
			}
			oldDelay = newRule.Delay

		}
		// Check if position was explicitly set by user
		if v, exists := ruleMap["position"]; exists && v != nil {
			// Only use user-specified position if it's not 0 (which might be default)
			if v.(int) != 0 {
				newRule.Position = v.(int)
			} else {
				// Position is 0, which might be default, so auto-assign
				newRule.Position = i + 1
			}
		} else {
			// No position specified, auto-assign (starting from 1)
			newRule.Position = i + 1
		}

		if v, ok := ruleMap["targets"]; ok {
			targets := v.([]interface{})
			newRule.Targets = make([]client.Targets, len(targets))
			for j, target := range targets {
				targetMap := target.(map[string]interface{})
				newTarget := client.Targets{}
				if v, ok := targetMap["target_type"]; ok {
					newTarget.TargetType = v.(int)
				}
				if v, ok := targetMap["target_id"]; ok {
					newTarget.TargetID = v.(string)
				}
				// Check if target position was explicitly set by user
				if v, exists := targetMap["position"]; exists && v != nil {
					// Only use user-specified position if it's not 0 (which might be default)
					if v.(int) != 0 {
						newTarget.Position = v.(int)
					} else {
						// Position is 0, which might be default, so auto-assign
						newTarget.Position = j
					}
				} else {
					// No position specified, auto-assign
					newTarget.Position = j
				}

				newRule.Targets[j] = newTarget
			}
		}

		newEsp.Rules[i] = newRule
	}

	return newEsp, nil

}

func flattenRules(rules []client.Rules) []map[string]interface{} {
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Position < rules[j].Position
	})

	result := make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
		result[i] = map[string]interface{}{
			"delay":    rule.Delay,
			"position": rule.Position,
			"targets":  flattenTargets(rule.Targets),
		}
	}
	return result
}

func flattenTargets(targets []client.Targets) []map[string]interface{} {
	sort.SliceStable(targets, func(i, j int) bool {
		return targets[i].Position < targets[j].Position
	})

	result := make([]map[string]interface{}, len(targets))
	for i, target := range targets {
		result[i] = map[string]interface{}{
			"target_type": target.TargetType,
			"target_id":   target.TargetID,
			"position":    target.Position,
		}
	}
	return result
}

func resourceCreateEsp(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics
	newEsp, createErr := CreateEsp(Ctx, d, m)
	if createErr != nil {
		return createErr
	}

	esp, err := apiclient.Esp.CreateEscalationPolicy(newEsp.Team, newEsp)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(esp.UniqueID)

	readErr := resourceReadEsp(Ctx, d, m)
	if readErr != nil {
		return readErr
	}

	return diags

}

func resourceUpdateEsp(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newEsp, createErr := CreateEsp(Ctx, d, m)
	if createErr != nil {
		return createErr
	}
	id := d.Id()
	var diags diag.Diagnostics

	_, err := apiclient.Esp.UpdateEscalationPolicy(newEsp.Team, id, newEsp)

	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDeleteEsp(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	err := apiclient.Esp.DeleteEscalationPolicy(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
func resourceReadEsp(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	if v, ok := d.GetOk("team_id"); ok {
		teamID = v.(string)
	}

	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required "))
	}
	var diags diag.Diagnostics
	esp, err := apiclient.Esp.GetEscalationPolicyByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", esp.Name)
	d.Set("team_id", esp.Team)
	d.Set("summary", esp.Summary)
	d.Set("description", esp.Description)
	d.Set("repeat_policy", esp.RepeatPolicy)
	d.Set("move_to_next", esp.MoveToNext)
	if err := d.Set("rules", flattenRules(esp.Rules)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceEscalationPolicyImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<esp_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid escalationpolicyid (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
