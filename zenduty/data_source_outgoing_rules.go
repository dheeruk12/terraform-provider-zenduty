package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOutgoingRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOutgoingRulesRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"service_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"integration_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"rules": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rule_json": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOutgoingRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	teamID := d.Get("team_id").(string)
	serviceID := d.Get("service_id").(string)
	integrationID := d.Get("integration_id").(string)

	rules, err := apiclient.OutgoingRules.GetOutgoingRules(teamID, serviceID, integrationID)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
		items[i] = map[string]interface{}{
			"unique_id": rule.UniqueID,
			"rule_json": rule.RuleJSON,
			"enabled":   rule.Enabled,
		}
	}

	if err := d.Set("rules", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s/%s", teamID, serviceID, integrationID))

	return diags
}
