package zenduty

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOutgoingRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateOutgoingRules,
		UpdateContext: resourceUpdateOutgoingRules,
		DeleteContext: resourceDeleteOutgoingRules,
		ReadContext:   wrapReadWith404(resourceReadOutgoingRules),
		Importer: &schema.ResourceImporter{
			State: resourceOutgoingRulesImporter,
		},
		Schema: map[string]*schema.Schema{
			"rule_json": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
		},
	}
}

func ValidateAncCreateOutgoingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.OutgoingRule, diag.Diagnostics) {
	newAlertRule := &client.OutgoingRule{}

	if v, ok := d.GetOk("rule_json"); ok {
		newAlertRule.RuleJSON = v.(string)
	}
	enabled := d.Get("enabled")
	newAlertRule.Enabled = enabled.(bool)

	if newAlertRule.RuleJSON == "" {
		return nil, diag.FromErr(errors.New("rule_json is required"))
	}
	if !isJSONString(newAlertRule.RuleJSON) {
		return nil, diag.FromErr(errors.New("rule_json is not valid JSON"))
	}

	return newAlertRule, nil
}

func resourceCreateOutgoingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var teamID, serviceID, integrationID string

	teamID = d.Get("team_id").(string)
	serviceID = d.Get("service_id").(string)
	integrationID = d.Get("integration_id").(string)

	newRule, ruleErr := ValidateAncCreateOutgoingRules(Ctx, d, m)
	if ruleErr != nil {
		return ruleErr
	}

	alertRule, err := apiclient.OutgoingRules.CreateOutgoingRule(teamID, serviceID, integrationID, newRule)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(alertRule.UniqueID)
	return diags
}

func resourceUpdateOutgoingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var teamID, serviceID, integrationID string

	teamID = d.Get("team_id").(string)
	serviceID = d.Get("service_id").(string)
	integrationID = d.Get("integration_id").(string)

	newRule, ruleErr := ValidateAncCreateOutgoingRules(Ctx, d, m)
	if ruleErr != nil {
		return ruleErr
	}

	_, err := apiclient.OutgoingRules.UpdateOutgoingRule(teamID, serviceID, integrationID, d.Id(), newRule)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceReadOutgoingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var teamID, serviceID, integrationID string

	teamID = d.Get("team_id").(string)
	serviceID = d.Get("service_id").(string)
	integrationID = d.Get("integration_id").(string)

	rule, err := apiclient.OutgoingRules.GetOutgoingRule(teamID, serviceID, integrationID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(rule.UniqueID)
	// normalize like alertrules/routing rules so formatting differences in
	// the stored JSON never show up as diffs
	if normalizedJSON, normErr := normalizeJSON(rule.RuleJSON); normErr == nil {
		d.Set("rule_json", normalizedJSON)
	} else {
		d.Set("rule_json", rule.RuleJSON)
	}
	d.Set("enabled", rule.Enabled)

	return diags
}

func resourceDeleteOutgoingRules(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	apiclient, _ := m.(*Config).Client()
	teamID, serviceID, integrationID, uniqueID := d.Get("team_id").(string), d.Get("service_id").(string), d.Get("integration_id").(string), d.Id()

	err := apiclient.OutgoingRules.DeleteOutgoingRule(teamID, serviceID, integrationID, uniqueID)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceOutgoingRulesImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
