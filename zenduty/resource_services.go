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

func resourceServices() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateServices,
		UpdateContext: resourceUpdateServices,
		DeleteContext: resourceDeleteServices,
		ReadContext:   resourceReadServices,
		Importer: &schema.ResourceImporter{
			State: resourceServiceImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"escalation_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"collation": {
				Type:     schema.TypeInt,
				Optional: true,
				// backend SERVICE_COLLATION_TYPES: 0 off, 1 time-based, 3 content-based (2 is unused)
				ValidateFunc: validation.IntInSlice([]int{0, 1, 3}),
			},
			"collation_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			"sla": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"task_template": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"team_priority": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"auto_resolve_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Seconds after which open incidents auto-resolve. 0 disables auto-resolution.",
			},
			"acknowledgement_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Seconds after which unacknowledged incidents re-trigger. 0 disables the timeout; the backend requires at least 600 seconds when enabled.",
			},
			"status": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"under_maintenance": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreateServices(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.Services, error) {
	newService := &client.Services{}

	if v, ok := d.GetOk("name"); ok {
		newService.Name = v.(string)
	}
	if v, ok := d.GetOk("escalation_policy"); ok {
		newService.EscalationPolicy = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		newService.Description = v.(string)
	}
	if v, ok := d.GetOk("summary"); ok {
		newService.Summary = v.(string)
	}
	if v, ok := d.GetOk("collation"); ok {
		newService.Collation = v.(int)
	}
	if v, ok := d.GetOk("collation_time"); ok {

		newService.CollationTime = v.(int)

	}
	if v, ok := d.GetOk("sla"); ok {
		newService.SLA = v.(string)
	}
	if v, ok := d.GetOk("task_template"); ok {
		newService.TaskTemplate = v.(string)
	}
	if v, ok := d.GetOk("team_priority"); ok {
		newService.TeamPriority = v.(string)
	}
	newService.AutoResolveTimeout = d.Get("auto_resolve_timeout").(int)
	newService.AcknowledgmentTimeout = d.Get("acknowledgement_timeout").(int)
	newService.AcknowledgementTimeoutEnabled = newService.AcknowledgmentTimeout > 0
	if newService.Collation != 0 && newService.CollationTime == 0 {
		return nil, fmt.Errorf("collation_time is required when collation is enabled")
	}
	if newService.Collation == 0 && newService.CollationTime != 0 {
		return nil, fmt.Errorf("collation_time cannot be set without enabling collation")
	}
	return newService, nil
}

func resourceCreateServices(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}

	newService, serviceErr := CreateServices(Ctx, d, m)
	if serviceErr != nil {
		return diag.FromErr(serviceErr)
	}

	var diags diag.Diagnostics
	service, err := apiclient.Services.CreateService(teamID, newService)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(service.UniqueID)
	d.Set("team_id", teamID)

	return diags
}

func resourceUpdateServices(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	newService, serviceErr := CreateServices(Ctx, d, m)
	if serviceErr != nil {
		return diag.FromErr(serviceErr)

	}

	_, err := apiclient.Services.UpdateService(teamID, id, newService)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceReadServices(Ctx, d, m)
}

func resourceDeleteServices(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	err := apiclient.Services.DeleteService(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceReadServices(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if emptyString(teamID) {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	service, err := apiclient.Services.GetServicesByID(teamID, id)
	if err != nil {
		return handleReadError(d, err)
	}
	d.Set("name", service.Name)
	d.Set("escalation_policy", service.EscalationPolicy)
	d.Set("description", service.Description)
	d.Set("summary", service.Summary)
	d.Set("collation", service.Collation)
	d.Set("collation_time", service.CollationTime)
	d.Set("sla", service.SLA)
	d.Set("task_template", service.TaskTemplate)
	d.Set("team_priority", service.TeamPriority)
	d.Set("team_id", teamID)
	d.Set("auto_resolve_timeout", service.AutoResolveTimeout)
	// a stored timeout only counts while the enabled flag is on; report the
	// effective value so 0 always means "disabled"
	if service.AcknowledgementTimeoutEnabled {
		d.Set("acknowledgement_timeout", service.AcknowledgmentTimeout)
	} else {
		d.Set("acknowledgement_timeout", 0)
	}
	d.Set("status", service.Status)
	d.Set("under_maintenance", service.UnderMaintenance)
	d.Set("creation_date", service.CreationDate)

	return diags
}

func resourceServiceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<service_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid service_id (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
