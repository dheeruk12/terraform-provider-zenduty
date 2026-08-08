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

func resourcePriority() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreatePriority,
		UpdateContext: resourceUpdatePriority,
		DeleteContext: resourceDeletePriority,
		ReadContext:   wrapReadWith404(resourceReadPriority),
		Importer: &schema.ResourceImporter{
			State: resourcePriorityImporter,
		},
		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"color": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
		},
	}
}

func validatePriority(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.Priority, diag.Diagnostics) {
	name := d.Get("name").(string)
	color := d.Get("color").(string)
	team := d.Get("team_id").(string)
	description := d.Get("description").(string)

	newPriority := &client.Priority{}
	if !IsValidUUID(team) {
		return nil, diag.FromErr(errors.New("team_id must be a valid UUID"))
	}

	newPriority.Name = name
	newPriority.Color = color
	newPriority.Description = description

	return newPriority, nil
}

func resourceCreatePriority(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	team := d.Get("team_id").(string)
	apiclient, _ := m.(*Config).Client()
	newpriority, validationerr := validatePriority(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	tag, err := apiclient.Priority.CreatePriority(team, newpriority)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tag.UniqueID)
	return nil
}

func resourceUpdatePriority(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	team := d.Get("team_id").(string)
	apiclient, _ := m.(*Config).Client()
	newpriority, validationerr := validatePriority(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	tag, err := apiclient.Priority.UpdatePriority(team, d.Id(), newpriority)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tag.UniqueID)
	return nil
}

func resourceDeletePriority(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	team := d.Get("team_id").(string)
	apiclient, _ := m.(*Config).Client()

	err := apiclient.Priority.DeletePriority(team, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceReadPriority(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	team := d.Get("team_id").(string)
	tag, err := apiclient.Priority.GetPriorityByID(team, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tag.UniqueID)
	d.Set("name", tag.Name)
	d.Set("color", tag.Color)
	d.Set("team_id", tag.Team)
	d.Set("description", tag.Description)
	return nil
}

func resourcePriorityImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<incident_priority_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid priority (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
