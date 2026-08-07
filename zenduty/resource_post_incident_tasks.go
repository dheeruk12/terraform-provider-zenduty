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

func resourcePostIncidentTasks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreatePostIncidentTasks,
		UpdateContext: resourceUpdatePostIncidentTasks,
		DeleteContext: resourceDeletePostIncidentTasks,
		ReadContext:   wrapReadWith404(resourceReadPostIncidentTasks),
		Importer: &schema.ResourceImporter{
			State: resourcePostIncidentTasksImporter,
		},
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"assigned_to": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2),
				Default:      0,
			},
			"due_in_time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func parseDueInTime(timestamp string) string {
	RFC3339local := "2006-01-02T15:04:05Z"
	t, parseErr := time.Parse(RFC3339local, timestamp)
	if parseErr != nil {
		return timestamp
	}
	return t.Format("2006-01-02 15:04")

}

func CreatePostIncidentTask(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.PostIncidentTaskObj, diag.Diagnostics) {

	newpostincidenttask := &client.PostIncidentTaskObj{}

	if v, ok := d.GetOk("team_id"); ok {
		newpostincidenttask.Team = v.(string)

	}
	if v, ok := d.GetOk("description"); ok {
		newpostincidenttask.Description = v.(string)

	}
	if v, ok := d.GetOk("title"); ok {
		newpostincidenttask.Title = v.(string)
	}
	if v, ok := d.GetOk("assigned_to"); ok {
		newpostincidenttask.AssignedTo = v.(string)
	}
	if v, ok := d.GetOk("status"); ok {
		newpostincidenttask.Status = v.(int)
	}
	if v, ok := d.GetOk("due_in_time"); ok {
		DueInTime := v.(string)
		newpostincidenttask.DueInTime = &DueInTime
		parsedTime, parsedErr := time.Parse("2006-01-02 15:04", v.(string))
		if parsedErr == nil {
			formattedTime := parsedTime.In(time.UTC).Format(time.RFC3339)
			newpostincidenttask.DueInTime = &formattedTime
		} else {
			return nil, diag.FromErr(parsedErr)
		}
	}

	return newpostincidenttask, nil

}

func resourceCreatePostIncidentTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}

	task, createErr := CreatePostIncidentTask(Ctx, d, m)
	if createErr != nil {
		return createErr
	}

	var diags diag.Diagnostics
	incidenttask, err := apiclient.PostIncidentTask.CreatePostIncidentTask(teamID, task)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(incidenttask.UniqueID)

	return diags
}

func resourceUpdatePostIncidentTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	task, createErr := CreatePostIncidentTask(Ctx, d, m)
	if createErr != nil {
		return createErr

	}

	_, err := apiclient.PostIncidentTask.UpdatePostIncidentTaskByID(teamID, id, task)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceReadPostIncidentTasks(Ctx, d, m)
}

func resourceDeletePostIncidentTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	err := apiclient.PostIncidentTask.DeletePostIncidentTaskByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceReadPostIncidentTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if emptyString(teamID) {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	postincidenttask, err := apiclient.PostIncidentTask.GetPostIncidentTaskByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("title", postincidenttask.Title)
	d.Set("description", postincidenttask.Description)
	d.Set("assigned_to", postincidenttask.AssignedTo)
	d.Set("status", postincidenttask.Status)
	d.Set("team_id", teamID)
	if postincidenttask.DueInTime != nil {
		d.Set("due_in_time", parseDueInTime(*postincidenttask.DueInTime))
	}
	d.Set("creation_date", postincidenttask.CreationDate)

	return diags
}

func resourcePostIncidentTasksImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<task_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid task_id (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
