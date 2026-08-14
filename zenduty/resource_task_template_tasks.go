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

func resourceTaskTemplateTaskTasks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateTaskTemplateTaskTasks,
		UpdateContext: resourceUpdateTaskTemplateTaskTasks,
		DeleteContext: resourceDeleteTaskTemplateTaskTasks,
		ReadContext:   resourceReadTaskTemplateTaskTasks,
		Importer: &schema.ResourceImporter{
			State: resourceTaskTemplateTaskTasksImporter,
		},
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"task_template_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the task. Zenduty allows tasks without a description, so this may be omitted or empty.",
			},
			"due_in": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validation.IntBetween(-1, 10080),
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"position": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func CreateTaskTemplateTask(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.TaskTemplateTaskObj, diag.Diagnostics) {

	newTaskTemplateTask := &client.TaskTemplateTaskObj{}

	if v, ok := d.GetOk("role"); ok {
		newTaskTemplateTask.Role = v.(string)

	}
	if v, ok := d.GetOk("description"); ok {
		newTaskTemplateTask.Description = v.(string)

	}
	if v, ok := d.GetOk("title"); ok {
		newTaskTemplateTask.Title = v.(string)
	}
	if v, ok := d.GetOk("due_in"); ok {
		newTaskTemplateTask.DueIn = v.(int)
	}
	if v, ok := d.GetOk("task_template_id"); ok {
		newTaskTemplateTask.TaskTemplate = v.(string)
	}
	position := d.Get("position").(int)
	newTaskTemplateTask.Position = position

	return newTaskTemplateTask, nil

}

func resourceCreateTaskTemplateTaskTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}

	task, createErr := CreateTaskTemplateTask(Ctx, d, m)
	if createErr != nil {
		return createErr
	}

	var diags diag.Diagnostics
	incidenttask, err := apiclient.TaskTemplate.CreateTaskTemplateTask(teamID, task.TaskTemplate, task)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(incidenttask.UniqueID)

	return diags
}

func resourceUpdateTaskTemplateTaskTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}
	task, createErr := CreateTaskTemplateTask(Ctx, d, m)
	if createErr != nil {
		return createErr

	}

	_, err := apiclient.TaskTemplate.UpdateTaskTemplateTaskByID(teamID, task.TaskTemplate, id, task)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceReadTaskTemplateTaskTasks(Ctx, d, m)
}

func resourceDeleteTaskTemplateTaskTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	task_id := d.Get("task_template_id").(string)

	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required"))
	}

	var diags diag.Diagnostics
	err := apiclient.TaskTemplate.DeleteTaskTemplateTaskByID(teamID, task_id, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceReadTaskTemplateTaskTasks(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	task_id := d.Get("task_template_id").(string)
	id := d.Id()
	if emptyString(teamID) {
		return diag.FromErr(errors.New("team_id is required"))
	}
	var diags diag.Diagnostics
	tasktemplatetask, err := apiclient.TaskTemplate.GetTaskTemplateTaskByID(teamID, task_id, id)
	if err != nil {
		return handleReadError(d, err)
	}
	d.Set("unique_id", tasktemplatetask.UniqueID)
	d.Set("title", tasktemplatetask.Title)
	d.Set("description", tasktemplatetask.Description)
	d.Set("creation_date", tasktemplatetask.CreationDate)
	d.Set("team_id", teamID)
	d.Set("role", tasktemplatetask.Role)
	d.Set("task_template_id", tasktemplatetask.TaskTemplate)
	d.Set("position", tasktemplatetask.Position)
	d.Set("due_in", tasktemplatetask.DueIn)
	return diags
}

func resourceTaskTemplateTaskTasksImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<task_template_id>/<task_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid task_template_id (%q)", parts[1])
	} else if !IsValidUUID(parts[2]) {
		return nil, fmt.Errorf("invalid task_id (%q)", parts[2])
	}
	d.Set("team_id", parts[0])
	d.Set("task_template_id", parts[1])
	d.SetId(parts[2])
	return []*schema.ResourceData{d}, nil
}
