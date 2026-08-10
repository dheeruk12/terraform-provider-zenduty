package zenduty

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateManintenances,
		UpdateContext: resourceUpdateManintenances,
		DeleteContext: resourceDeleteManintenances,
		ReadContext:   resourceReadManintenances,
		Importer: &schema.ResourceImporter{
			State: resourceMaintenanceImporter,
		},
		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"start_time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repeat_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"repeat_until": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"services": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func ValidateMaintenanceWindow(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.MaintenanceWindow, diag.Diagnostics) {
	newManintence := &client.MaintenanceWindow{}
	services := d.Get("services").([]interface{})

	if v, ok := d.GetOk("name"); ok {
		if v.(string) == "" {
			return nil, diag.FromErr(errors.New("name must not be empty"))
		}
		newManintence.Name = v.(string)
	}
	if v, ok := d.GetOk("timezone"); ok {
		if v.(string) == "" {
			return nil, diag.FromErr(errors.New("timezone must not be empty"))
		}
		newManintence.TimeZone = v.(string)
	}

	if v, ok := d.GetOk("start_time"); ok {
		newManintence.StartTime = v.(string)

		loc, zoneErr := time.LoadLocation(newManintence.TimeZone)
		if zoneErr != nil {
			return nil, diag.FromErr(errors.New(zoneErr.Error()))
		}
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04", v.(string), loc)
		if err != nil {
			return nil, diag.FromErr(errors.New(err.Error()))
		}
		newManintence.StartTime = parsedTime.In(time.UTC).Format(time.RFC3339)

	}
	if v, ok := d.GetOk("end_time"); ok {
		newManintence.EndTime = v.(string)

		loc, zoneErr := time.LoadLocation(newManintence.TimeZone)
		if zoneErr != nil {
			return nil, diag.FromErr(errors.New(zoneErr.Error()))
		}
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04", v.(string), loc)
		if err != nil {
			return nil, diag.FromErr(errors.New(err.Error()))
		}

		newManintence.EndTime = parsedTime.In(time.UTC).Format(time.RFC3339)
	}

	if v, ok := d.GetOk("repeat_interval"); ok {
		newManintence.RepeatInterval = v.(int)
	}
	if v, ok := d.GetOk("repeat_until"); ok {
		if v.(string) == "" {
			return nil, diag.FromErr(errors.New("repeat_until must not be empty"))
		}

		loc, zoneErr := time.LoadLocation(newManintence.TimeZone)
		if zoneErr != nil {
			return nil, diag.FromErr(errors.New(zoneErr.Error()))
		}

		parsedTime, err := time.ParseInLocation("2006-01-02 15:04", v.(string), loc)
		if err != nil {
			return nil, diag.FromErr(errors.New(err.Error()))
		}
		newManintence.RepeatUntil = parsedTime.In(time.UTC).Format(time.RFC3339)
	}
	for _, service := range services {
		if service.(string) == "" {
			return nil, diag.FromErr(errors.New("services must not be empty"))
		}
		if !IsValidUUID(service.(string)) {
			return nil, diag.FromErr(errors.New("services must be a valid UUID"))
		}
		newManintence.Services = append(newManintence.Services, client.ServiceMaintenance{Service: service.(string)})
	}
	return newManintence, nil
}

func resourceCreateManintenances(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if !IsValidUUID(v.(string)) {
			return diag.FromErr(errors.New("team_id must be a valid UUID"))

		}
		teamID = v.(string)
	}
	apiclient, _ := m.(*Config).Client()
	newManintence, diags := ValidateMaintenanceWindow(ctx, d, m)
	if diags != nil {
		return diags
	}
	maintenance, err := apiclient.MaintenanceWindow.CreateMaintenanceWindow(teamID, newManintence)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(maintenance.UniqueID)
	return nil
}

func resourceUpdateManintenances(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if !IsValidUUID(v.(string)) {
			return diag.FromErr(errors.New("team_id must be a valid UUID"))
		}
		teamID = v.(string)
	}
	apiclient, _ := m.(*Config).Client()
	newManintence, diags := ValidateMaintenanceWindow(ctx, d, m)
	if diags != nil {
		return diags
	}
	maintenance, err := apiclient.MaintenanceWindow.UpdateMaintenanceWindow(teamID, d.Id(), newManintence)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(maintenance.UniqueID)
	return nil
}

func resourceReadManintenances(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if !IsValidUUID(v.(string)) {
			return diag.FromErr(errors.New("team_id must be a valid UUID"))
		}
		teamID = v.(string)
	}
	apiclient, _ := m.(*Config).Client()
	maintenance, err := apiclient.MaintenanceWindow.GetMaintenanceWindowByID(teamID, d.Id())
	if err != nil {
		return handleReadError(d, err)
	}
	d.Set("name", maintenance.Name)
	d.Set("repeat_interval", maintenance.RepeatInterval)
	startTime, timeErr := createMaintenanceTimeFormat(maintenance.StartTime, maintenance.TimeZone)
	endTime, endTimeErr := createMaintenanceTimeFormat(maintenance.EndTime, maintenance.TimeZone)
	RepeatUntil, RepeatUntilErr := createMaintenanceTimeFormat(maintenance.RepeatUntil, maintenance.TimeZone)
	if RepeatUntilErr == nil {
		d.Set("repeat_until", RepeatUntil)
	}

	if timeErr == nil && endTimeErr == nil {
		d.Set("start_time", startTime)
		d.Set("end_time", endTime)
	}
	d.Set("services", flattenServices(maintenance.Services))
	d.Set("timezone", maintenance.TimeZone)

	return nil
}

func flattenServices(services []client.ServiceMaintenance) []interface{} {
	var servicesList []interface{}
	for _, service := range services {
		servicesList = append(servicesList, service.Service)
	}
	return servicesList
}

func createMaintenanceTimeFormat(timestamp, Zone string) (string, error) {
	RFC3339local := "2006-01-02T15:04:05Z"
	loc, zoneErr := time.LoadLocation(Zone)
	if zoneErr != nil {
		return "", zoneErr
	}
	t, parseErr := time.ParseInLocation(RFC3339local, timestamp, time.UTC)
	if parseErr != nil {
		return "", parseErr
	}

	return t.In(loc).Format("2006-01-02 15:04"), nil

}

func resourceDeleteManintenances(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if !IsValidUUID(v.(string)) {
			return diag.FromErr(errors.New("team_id must be a valid UUID"))
		}
		teamID = v.(string)
	}
	apiclient, _ := m.(*Config).Client()
	err := apiclient.MaintenanceWindow.DeleteMaintenanceWindow(teamID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceMaintenanceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<maintenance_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid maintenance (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
