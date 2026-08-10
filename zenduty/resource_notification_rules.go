package zenduty

import (
	"context"
	"fmt"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNotificationRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateNotificationRule,
		ReadContext:   resourceReadNotificationRule,
		UpdateContext: resourceUpdateNotificationRule,
		DeleteContext: resourceDeleteNotificationRule,
		Importer: &schema.ResourceImporter{
			State: resourceNotificationRuleImporter,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"contact": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"urgency": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1),
			},
			"delay": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceCreateNotificationRule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newNotificationRule := &client.CreateNotificationRules{}
	var diags diag.Diagnostics
	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	if v, ok := d.GetOk("urgency"); ok {
		newNotificationRule.Urgency = v.(int)
	}
	if v, ok := d.GetOk("contact"); ok {
		newNotificationRule.Contact = v.(string)
	}
	if v, ok := d.GetOk("delay"); ok {
		newNotificationRule.StartDelay = v.(int)
	}
	notificationRule, err := apiclient.NotificationRules.CreateNotificationRules(username, newNotificationRule)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(notificationRule.UniqueID)
	return diags
}

func resourceUpdateNotificationRule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newNotificationRule := &client.NotificationRules{}
	var diags diag.Diagnostics
	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	if v, ok := d.GetOk("urgency"); ok {
		newNotificationRule.Urgency = v.(int)
	}
	if v, ok := d.GetOk("contact"); ok {
		newNotificationRule.Contact = v.(string)
	}
	if v, ok := d.GetOk("delay"); ok {
		newNotificationRule.StartDelay = v.(int)
	}
	notificationRule, err := apiclient.NotificationRules.UpdateNotificationRules(username, d.Id(), newNotificationRule)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(notificationRule.UniqueID)
	return diags
}

func resourceReadNotificationRule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	notificationRule, err := apiclient.NotificationRules.GetNotificationRulesByID(username, d.Id())
	if err != nil {
		return handleReadError(d, err)
	}
	d.Set("contact", notificationRule.Contact)
	d.Set("delay", notificationRule.StartDelay)
	d.Set("urgency", notificationRule.Urgency)

	return diags
}

func resourceDeleteNotificationRule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	err := apiclient.NotificationRules.DeleteNotificationRules(username, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceNotificationRuleImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format. Expecting username/notification_rule_id, got: %s", d.Id())
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid notification_rule: %s", parts[1])
	}

	d.Set("username", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
