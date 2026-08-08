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

func resourceIntegrations() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIntegrationCreate,
		UpdateContext: resourceIntegrationUpdate,
		DeleteContext: resourceIntegrationDelete,
		ReadContext:   wrapReadWith404(resourceIntegrationRead),
		Importer: &schema.ResourceImporter{
			State: resourceIntegrationImporter,
		},
		Schema: map[string]*schema.Schema{
			"application": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"webhook_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"create_incident_for": {
				Type:     schema.TypeInt,
				Optional: true,
				// 0 none, 1 critical, 2 critical+error, 3 critical+error+warning
				ValidateFunc: validation.IntBetween(0, 3),
				Default:      1,
			},
			"default_urgency": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 1),
				Default:      1,
			},
		},
	}
}

func resourceIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newIntegration := &client.IntegrationCreate{}
	var diags diag.Diagnostics
	teamID := d.Get("team_id").(string)
	serviceID := d.Get("service_id").(string)
	summary := d.Get("summary").(string)
	if summary != "" {
		newIntegration.Summary = summary
	}

	if v, ok := d.GetOk("name"); ok {
		newIntegration.Name = v.(string)
	}
	if v, ok := d.GetOk("application"); ok {
		newIntegration.Application = v.(string)
	}
	newIntegration.IsEnabled = d.Get("is_enabled").(bool)
	newIntegration.CreateIncidentFor = d.Get("create_incident_for").(int)
	newIntegration.DefaultUrgency = d.Get("default_urgency").(int)

	integration, err := apiclient.Integrations.CreateIntegration(teamID, serviceID, newIntegration)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(integration.UniqueID)
	// added integration_key in response output
	d.Set("integration_key", integration.IntegrationKey)
	d.Set("is_enabled", integration.IsEnabled)
	d.Set("webhook_url", integration.WebhookURL)

	return diags
}

func resourceIntegrationUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	id := d.Id()

	newIntegration := &client.IntegrationCreate{}
	var diags diag.Diagnostics
	teamID := d.Get("team_id").(string)
	serviceID := d.Get("service_id").(string)
	summary := d.Get("summary").(string)
	if summary != "" {
		newIntegration.Summary = summary
	}

	if v, ok := d.GetOk("name"); ok {
		newIntegration.Name = v.(string)
	}
	if v, ok := d.GetOk("application"); ok {
		newIntegration.Application = v.(string)
	}
	newIntegration.IsEnabled = d.Get("is_enabled").(bool)
	newIntegration.CreateIncidentFor = d.Get("create_incident_for").(int)
	newIntegration.DefaultUrgency = d.Get("default_urgency").(int)

	integration, err := apiclient.Integrations.UpdateIntegration(teamID, serviceID, id, newIntegration)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("integration_key", integration.IntegrationKey)
	d.Set("is_enabled", integration.IsEnabled)
	d.Set("webhook_url", integration.WebhookURL)
	// added integration_key in response output

	return diags
}

func resourceIntegrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	teamID := d.Get("team_id").(string)
	serviceID := d.Get("service_id").(string)
	var diags diag.Diagnostics
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	if serviceID == "" {
		return diag.FromErr(errors.New("service_id is required"))
	}
	err := apiclient.Integrations.DeleteIntegration(teamID, serviceID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	id := d.Id()
	teamID := d.Get("team_id").(string)
	serviceID := d.Get("service_id").(string)
	apiclient, _ := m.(*Config).Client()

	integration, err := apiclient.Integrations.GetIntegrationByID(teamID, serviceID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", integration.Name)
	d.Set("application", integration.Application)
	d.Set("summary", integration.Summary)
	d.Set("integration_key", integration.IntegrationKey)
	d.Set("webhook_url", integration.WebhookURL)
	d.Set("is_enabled", integration.IsEnabled)
	d.Set("create_incident_for", integration.CreateIncidentFor)
	d.Set("default_urgency", integration.DefaultUrgency)

	return diags
}

func resourceIntegrationImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<service_id>/<integration_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid serviceid (%q)", parts[1])
	} else if !IsValidUUID(parts[2]) {
		return nil, fmt.Errorf("invalid integration_id (%q)", parts[2])
	}
	d.SetId(parts[2])
	d.Set("team_id", parts[0])
	d.Set("service_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
