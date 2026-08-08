package zenduty

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSchedules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateSchedule,
		UpdateContext: resourceUpdateSchedule,
		DeleteContext: resourceDeleteSchedule,
		ReadContext:   wrapReadWith404(resourceReadSchedule),
		Importer: &schema.ResourceImporter{
			State: resourceScheduleImporter,
		},
		CustomizeDiff: validateScheduleRestrictions,

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
			"time_zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"layers": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"shift_length": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(3600, 365*24*3600),
						},
						"rotation_start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rotation_end_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"users": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"restriction_type": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 2),
							Default:      0,
						},
						"restrictions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"duration": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 7*24*3600),
									},
									"start_day_of_week": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 7),
									},
									"start_time_of_day": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9]|0[0-9]|1[0-9]|2[0-3]):([0-9]|[0-5][0-9]):([0-9]|[0-5][0-9])$`), "must be in the format HH:MM:SS"),
										// the API stores zero-padded times; accept 9:5:0 in
										// config without diffing against 09:05:00
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											return padTimeOfDay(old) == padTimeOfDay(new)
										},
									},
								},
							},
						},
					},
				},
			},
			"overrides": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"user": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: ValidateUserName(),
						},
					},
				},
			},
		},
	}
}

// validateScheduleRestrictions enforces the cross-field restriction rules at
// plan time: restrictions require a restriction_type, and a restriction may
// not span more than one period of its type.
func validateScheduleRestrictions(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	// Values interpolated from resources that have not been created yet read
	// as their zero value here, which would report a bogus error. Leave those
	// to the apply-time checks.
	if !diff.NewValueKnown("layers") {
		return nil
	}
	layers, _ := diff.Get("layers").([]interface{})
	for i, layer := range layers {
		layerMap, ok := layer.(map[string]interface{})
		if !ok {
			continue
		}
		restrictionType, _ := layerMap["restriction_type"].(int)
		restrictions, _ := layerMap["restrictions"].([]interface{})
		if len(restrictions) > 0 && restrictionType == 0 {
			return fmt.Errorf("layers[%d]: restriction_type must be 1 (daily) or 2 (weekly) when restrictions are set", i)
		}
		for j, restriction := range restrictions {
			restrictionMap, ok := restriction.(map[string]interface{})
			if !ok {
				continue
			}
			duration, _ := restrictionMap["duration"].(int)
			if restrictionType == 1 && duration >= 86400 {
				return fmt.Errorf("layers[%d].restrictions[%d]: duration must be less than 86400 seconds (24 hours) for daily restrictions", i, j)
			}
			if restrictionType == 2 && duration >= 604800 {
				return fmt.Errorf("layers[%d].restrictions[%d]: duration must be less than 604800 seconds (7 days) for weekly restrictions", i, j)
			}
		}
	}
	return nil
}

// padTimeOfDay normalizes H:M:S to a zero-padded HH:MM:SS; values that do not
// parse are returned unchanged.
func padTimeOfDay(s string) string {
	t, err := time.Parse("15:4:5", s)
	if err != nil {
		return s
	}
	return t.Format("15:04:05")
}

func buildScheduleLayerRescrition(newLayer *client.CreateLayers, layerMap map[string]interface{}, d *schema.ResourceData) ([]client.Restrictions, diag.Diagnostics) {
	if v, ok := layerMap["restriction_type"]; ok {
		newLayer.RestrictionType = v.(int)
	}

	if v, ok := layerMap["restrictions"]; ok {
		restrictions := v.([]interface{})
		Restrictions := make([]client.Restrictions, len(restrictions))
		for j, restriction := range restrictions {
			if newLayer.RestrictionType == 0 {
				return nil, diag.FromErr(errors.New("restriction_type must be 1 (daily) or 2 (weekly) when restrictions are set"))
			}
			restrictionMap := restriction.(map[string]interface{})
			newRestriction := client.Restrictions{}
			if v, ok := restrictionMap["duration"]; ok {
				newRestriction.Duration = v.(int)
				if newLayer.RestrictionType == 1 && newRestriction.Duration >= 86400 {
					return nil, diag.FromErr(errors.New("duration must be less than 86400 for daily restriction ie 24 hours"))
				} else if newLayer.RestrictionType == 2 && newRestriction.Duration >= 604800 {
					return nil, diag.FromErr(errors.New("duration must be less than 604800 for weekly restriction ie 7 days"))
				}
			}
			if v, ok := restrictionMap["start_day_of_week"]; ok {
				newRestriction.StartDayOfWeek = v.(int)
			}
			if v, ok := restrictionMap["start_time_of_day"]; ok {
				newRestriction.StartTimeOfDay = padTimeOfDay(v.(string))
			}

			Restrictions[j] = newRestriction
		}
		return Restrictions, nil
	}
	return nil, nil
}

func buildScheduleLayer(ctx context.Context, d *schema.ResourceData, TimeZone string) ([]client.CreateLayers, diag.Diagnostics) {
	layers := d.Get("layers").([]interface{})
	Layers := make([]client.CreateLayers, len(layers))

	for i, layer := range layers {
		layerMap := layer.(map[string]interface{})
		newLayer := client.CreateLayers{}

		if v, ok := layerMap["name"]; ok {
			if v.(string) == "" {
				return nil, diag.FromErr(errors.New("name must not be empty"))
			}

			newLayer.Name = v.(string)
		}
		// if v, ok := layerMap["time_zone"]; ok {
		// 	newLayer.TimeZone = v.(string)
		// }
		if v, ok := layerMap["shift_length"]; ok {

			newLayer.ShiftLength = v.(int)
		}
		if v, ok := layerMap["rotation_start_time"]; ok {

			newLayer.RotationStartTime = v.(string)
			loc, _ := time.LoadLocation(TimeZone)

			parsedTime, parseErr := time.ParseInLocation("2006-01-02 15:04", v.(string), loc)
			if parseErr == nil {
				newLayer.RotationStartTime = parsedTime.In(time.UTC).Format(time.RFC3339)
			}

		}
		if v, ok := layerMap["rotation_end_time"]; ok {
			newLayer.RotationEndTime = v.(string)
			loc, _ := time.LoadLocation(TimeZone)
			parsedTime, parsedErr := time.ParseInLocation("2006-01-02 15:04", v.(string), loc)
			if parsedErr == nil {
				newLayer.RotationEndTime = parsedTime.In(time.UTC).Format(time.RFC3339)
			}

		}
		if v, ok := layerMap["users"]; ok {
			users := v.([]interface{})
			newLayer.Users = make([]client.CreateUserLayer, len(users))
			for j, user := range users {
				newUser := client.CreateUserLayer{}
				newUser.User = user.(string)
				newLayer.Users[j] = newUser
			}
		}
		newRestriction, restrictionErr := buildScheduleLayerRescrition(&newLayer, layerMap, d)
		if restrictionErr != nil {
			return nil, restrictionErr
		}
		if newRestriction != nil {
			newLayer.Restrictions = newRestriction
		}
		Layers[i] = newLayer
	}
	return Layers, nil

}

func buildScheduleOverride(newSchedule *client.CreateSchedule, d *schema.ResourceData) ([]client.Overrides, diag.Diagnostics) {
	overrides := d.Get("overrides").([]interface{})
	Overrides := make([]client.Overrides, len(overrides))

	for o, override := range overrides {
		override := override.(map[string]interface{})
		newOverride := client.Overrides{}

		if v, ok := override["name"]; ok {

			newOverride.Name = v.(string)
		}
		if v, ok := override["start_time"]; ok {

			newOverride.StartTime = v.(string)

			loc, zoneErr := time.LoadLocation(newSchedule.TimeZone)
			if zoneErr != nil {
				return nil, diag.FromErr(errors.New(zoneErr.Error()))
			}
			parsedTime, err := time.ParseInLocation("2006-01-02 15:04", v.(string), loc)
			if err != nil {
				return nil, diag.FromErr(errors.New(err.Error()))
			}

			newOverride.StartTime = parsedTime.In(time.UTC).Format(time.RFC3339)

		}
		if v, ok := override["end_time"]; ok {
			newOverride.EndTime = v.(string)

			loc, zoneErr := time.LoadLocation(newSchedule.TimeZone)
			if zoneErr != nil {
				return nil, diag.FromErr(errors.New(zoneErr.Error()))
			}
			parsedEndTime, err := time.ParseInLocation("2006-01-02 15:04", newOverride.EndTime, loc)

			if err != nil {
				return nil, diag.FromErr(errors.New(err.Error()))
			}

			newOverride.EndTime = parsedEndTime.In(time.UTC).Format(time.RFC3339)

		}
		if v, ok := override["user"]; ok {

			newOverride.User = v.(string)
		}
		Overrides[o] = newOverride
	}
	return Overrides, nil
}

func createSchedule(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.CreateSchedule, diag.Diagnostics) {
	newSchedule := &client.CreateSchedule{}

	if v, ok := d.GetOk("name"); ok {
		if v.(string) == "" {
			return nil, diag.FromErr(errors.New("name must not be empty"))
		}
		newSchedule.Name = v.(string)
	}
	if v, ok := d.GetOk("summary"); ok {
		newSchedule.Summary = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		newSchedule.Description = v.(string)
	}
	if v, ok := d.GetOk("time_zone"); ok {
		if emptyString(v.(string)) {
			return nil, diag.FromErr(errors.New("time_zone must not be empty"))
		}
		_, zoneErr := time.LoadLocation(v.(string))
		if zoneErr != nil {
			return nil, diag.FromErr(errors.New(zoneErr.Error()))
		}
		newSchedule.TimeZone = v.(string)

	}
	if v, ok := d.GetOk("team_id"); ok {
		if emptyString(v.(string)) {
			return nil, diag.FromErr(errors.New("team_id must not be empty"))
		}
		newSchedule.Team = v.(string)

	}

	Layers, layerErr := buildScheduleLayer(Ctx, d, newSchedule.TimeZone)
	if layerErr != nil {
		return nil, layerErr
	}
	Overrides, overrideErr := buildScheduleOverride(newSchedule, d)
	if overrideErr != nil {
		return nil, overrideErr
	}

	newSchedule.Overrides = Overrides
	newSchedule.Layers = Layers

	return newSchedule, nil
}

func resourceCreateSchedule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var diags diag.Diagnostics
	newSchedule, createError := createSchedule(Ctx, d, m)
	if createError != nil {
		return createError
	}
	schedule, err := apiclient.Schedules.CreateSchedule(newSchedule.Team, newSchedule)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schedule.UniqueID)
	return diags

}

func resourceUpdateSchedule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	newSchedule, createError := createSchedule(Ctx, d, m)
	if createError != nil {
		return createError
	}
	_, err := apiclient.Schedules.UpdateScheduleByID(teamID, id, newSchedule)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDeleteSchedule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	err := apiclient.Schedules.DeleteScheduleByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceReadSchedule(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	service, err := apiclient.Schedules.GetScheduleByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", service.Name)
	d.Set("summary", service.Summary)
	d.Set("description", service.Description)
	d.Set("time_zone", service.TimeZone)
	d.Set("team_id", service.Team)
	if err := d.Set("layers", flattenLayer(service.TimeZone, service.Layers)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("overrides", flattenScheduleOverrides(service.TimeZone, service.Overrides)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenLayer(TimeZone string, layers []client.Layers) []map[string]interface{} {

	var layerList []map[string]interface{}
	for _, layer := range layers {
		layerList = append(layerList, map[string]interface{}{
			"name":                layer.Name,
			"shift_length":        layer.ShiftLength,
			"rotation_start_time": createScheduleLayerTimeFormat(layer.RotationStartTime, TimeZone),
			"rotation_end_time":   createScheduleLayerTimeFormat(layer.RotationEndTime, TimeZone),
			"users":               flattenLayerUsers(layer.Users),
			"restriction_type":    layer.RestrictionType,
			"restrictions":        flattenLayerRestrictions(layer.Restrictions),
		})
	}
	return layerList

}

func createScheduleLayerTimeFormat(timestamp, Zone string) string {
	RFC3339local := "2006-01-02T15:04:05Z"
	loc, zoneErr := time.LoadLocation(Zone)
	if zoneErr != nil {
		return timestamp
	}
	t, parseErr := time.ParseInLocation(RFC3339local, timestamp, time.UTC)
	if parseErr != nil {
		return timestamp
	}

	return t.In(loc).Format("2006-01-02 15:04")

}

func flattenLayerUsers(users []client.Users) []string {

	var userList []string
	for _, user := range users {
		userList = append(userList, user.User)
	}
	return userList

}

func flattenLayerRestrictions(restrictions []client.Restrictions) []map[string]interface{} {

	var restrictionList []map[string]interface{}
	for _, restriction := range restrictions {
		restrictionList = append(restrictionList, map[string]interface{}{
			"duration":          restriction.Duration,
			"start_day_of_week": restriction.StartDayOfWeek,
			"start_time_of_day": restriction.StartTimeOfDay,
		})
	}
	return restrictionList

}

func flattenScheduleOverrides(TimeZone string, overrides []client.Overrides) []map[string]interface{} {

	var overrideList []map[string]interface{}
	for _, override := range overrides {
		overrideList = append(overrideList, map[string]interface{}{
			"name":       override.Name,
			"start_time": createScheduleLayerTimeFormat(override.StartTime, TimeZone),
			"end_time":   createScheduleLayerTimeFormat(override.EndTime, TimeZone),
			"user":       override.User,
		})
	}

	return overrideList
}

func resourceScheduleImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format. Expecting team_id/schedule_id, got: %s", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id: %s", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid schedule_id: %s", parts[1])
	}

	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
