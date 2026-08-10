package zenduty

import (
	"context"
	"regexp"
	"strconv"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// The API appends "\n\n[Created by: First Last(email)]" to the summary of
// every incident created through it, so the value read back never matches
// the configured one.
var createdBySuffix = regexp.MustCompile(`\n\n\[Created by: [^\]]*\]$`)

func resourceIncidents() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIncidentsCreate,
		UpdateContext: resourceIncidentUpdate,
		DeleteContext: resourceIncidentDelete,
		ReadContext:   resourceIncidentRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"escalation_policy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// The API treats user as the acting user on create and never
				// returns it, so an imported incident has no value in state;
				// suppress that one-sided diff or every import plans a replacement.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "" && d.Id() != ""
				},
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// ignore the server-appended "[Created by: ...]" suffix or
				// every plan after create forces a replacement
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return createdBySuffix.ReplaceAllString(old, "") == new
				},
			},
			"status": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 3),
			},
			"urgency": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 1),
				Description:  "Urgency of the incident: 0 (low) or 1 (high).",
			},
		},
	}
}

func resourceIncidentsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newIncident := &client.Incident{}
	var diags diag.Diagnostics
	summary := d.Get("summary").(string)
	if summary != "" {
		newIncident.Summary = summary
	}
	if v, ok := d.GetOk("title"); ok {
		newIncident.Title = v.(string)
	}
	if v, ok := d.GetOk("user"); ok {
		newIncident.User = v.(string)
	}
	if v, ok := d.GetOk("escalation_policy"); ok {
		newIncident.EscalationPolicy = v.(string)
	}
	if v, ok := d.GetOk("service"); ok {
		newIncident.Service = v.(string)
	}
	// raw-config check because 0 (low) is a meaningful urgency, unlike the
	// zero value of most attributes; unset keeps the service default.
	if !d.GetRawConfig().GetAttr("urgency").IsNull() {
		urgency := d.Get("urgency").(int)
		newIncident.Urgency = &urgency
	}

	incident, err := apiclient.Incidents.CreateIncident(newIncident)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(incident.IncidentNumber))

	if v, ok := d.GetOk("status"); ok && v.(int) != incident.Status {
		newStatus := &client.IncidentStatus{Status: v.(int)}
		if _, err := apiclient.Incidents.UpdateIncident(d.Id(), newStatus); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceIncidentUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChange("status") || d.HasChange("urgency") {
		apiclient, _ := m.(*Config).Client()

		patch := &client.IncidentStatus{}
		if d.HasChange("status") {
			patch.Status = d.Get("status").(int)
		}
		if d.HasChange("urgency") {
			urgency := d.Get("urgency").(int)
			patch.Urgency = &urgency
		}
		_, err := apiclient.Incidents.UpdateIncident(d.Id(), patch)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

// Incidents cannot be deleted via the Zenduty API; deleting the resource only
// removes it from Terraform state. The incident itself is left untouched.
func resourceIncidentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	return diags
}

func resourceIncidentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics
	incident, err := apiclient.Incidents.GetIncidentByNumber(d.Id())
	if err != nil {
		return handleReadError(d, err)
	}
	d.Set("title", incident.Title)
	d.Set("summary", incident.Summary)
	d.Set("status", incident.Status)
	d.Set("urgency", incident.Urgency)
	d.Set("service", incident.Service)
	d.Set("escalation_policy", incident.EscalationPolicy)

	return diags
}
