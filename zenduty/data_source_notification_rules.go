package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNotificationRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNotificationRulesRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUserName(),
			},
			"notification_rules": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"contact": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"urgency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"start_delay": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNotificationRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	username := d.Get("username").(string)

	rules, err := apiclient.NotificationRules.GetNotificationRules(username)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
		items[i] = map[string]interface{}{
			"unique_id":   rule.UniqueID,
			"contact":     rule.Contact,
			"urgency":     rule.Urgency,
			"start_delay": rule.StartDelay,
			"type":        rule.Type,
		}
	}

	if err := d.Set("notification_rules", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(username)

	return diags
}
