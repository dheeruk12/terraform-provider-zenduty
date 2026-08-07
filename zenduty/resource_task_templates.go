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

func resourceTaskTemplates() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateTaskTemplates,
		UpdateContext: resourceUpdateTaskTemplates,
		DeleteContext: resourceDeleteTaskTemplates,
		ReadContext:   wrapReadWith404(resourceReadTaskTemplates),
		Importer: &schema.ResourceImporter{
			State: resourceTaskTemplatesImporter,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Required: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreateTaskTemplate(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.TaskTemplateObj, diag.Diagnostics) {

	newtasktemplate := &client.TaskTemplateObj{}

	if v, ok := d.GetOk("team_id"); ok {
		newtasktemplate.Team = v.(string)

	}
	if v, ok := d.GetOk("summary"); ok {
		newtasktemplate.Summary = v.(string)

	}
	if v, ok := d.GetOk("name"); ok {
		newtasktemplate.Name = v.(string)

	}

	return newtasktemplate, nil

}

func resourceCreateTaskTemplates(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}

	task, createErr := CreateTaskTemplate(Ctx, d, m)
	if createErr != nil {
		return createErr
	}

	var diags diag.Diagnostics
	incidenttask, err := apiclient.TaskTemplate.CreateTaskTemplate(teamID, task)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(incidenttask.UniqueID)

	return diags
}

func resourceUpdateTaskTemplates(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	task, createErr := CreateTaskTemplate(Ctx, d, m)
	if createErr != nil {
		return createErr

	}

	_, err := apiclient.TaskTemplate.UpdateTaskTemplateByID(teamID, id, task)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceReadTaskTemplates(Ctx, d, m)
}

func resourceDeleteTaskTemplates(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	err := apiclient.TaskTemplate.DeleteTaskTemplateByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceReadTaskTemplates(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if emptyString(teamID) {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	postincidenttask, err := apiclient.TaskTemplate.GetTaskTemplateByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", postincidenttask.Name)
	d.Set("summary", postincidenttask.Summary)
	d.Set("creation_date", postincidenttask.CreationDate)
	d.Set("team_id", teamID)

	return diags
}

func resourceTaskTemplatesImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<task_template_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid task_template_id (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
