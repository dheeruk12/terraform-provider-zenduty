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

func resourceRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		ReadContext:   wrapReadWith404(resourceRoleRead),
		Importer: &schema.ResourceImporter{
			State: resourceIncidentRoleImporter,
		},
		Schema: map[string]*schema.Schema{
			"team": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"unique_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rank": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
				Default:      1,
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newrole := &client.Roles{}
	var diags diag.Diagnostics
	if v, ok := d.GetOk("team"); ok {
		newrole.Team = v.(string)

	}
	d.Set("team", newrole.Team)
	if v, ok := d.GetOk("description"); ok {
		newrole.Description = v.(string)

	}
	if v, ok := d.GetOk("title"); ok {
		newrole.Title = v.(string)
	}
	if v, ok := d.GetOk("rank"); ok {
		newrole.Rank = v.(int)
		if newrole.Rank == 0 {
			newrole.Rank = 1
		}
		if newrole.Rank <= 0 || newrole.Rank > 10 {
			return diag.Errorf("Rank should be between 1 and 10")
		}
	}

	role, err := apiclient.Roles.CreateRole(newrole.Team, newrole)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(role.UniqueID)
	return diags
}

func resourceRoleUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newrole := &client.Roles{}
	var teamID string
	id := d.Id()
	newrole.UniqueID = id
	var diags diag.Diagnostics
	if v, ok := d.GetOk("description"); ok {
		newrole.Description = v.(string)

	}
	if v, ok := d.GetOk("title"); ok {
		newrole.Title = v.(string)
	}
	if v, ok := d.GetOk("team"); ok {
		teamID = v.(string)
	}
	if v, ok := d.GetOk("rank"); ok {
		newrole.Rank = v.(int)
		if newrole.Rank == 0 {
			newrole.Rank = 1
		}
		if newrole.Rank <= 0 || newrole.Rank > 10 {
			return diag.Errorf("Rank should be between 1 and 10")
		}
	}
	_, err := apiclient.Roles.UpdateRoles(teamID, newrole)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	teamID := d.Get("team").(string)
	var diags diag.Diagnostics

	err := apiclient.Roles.DeleteRole(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	id := d.Id()
	teamID := d.Get("team").(string)
	var diags diag.Diagnostics
	role, err := apiclient.Roles.GetRolesByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("title", role.Title)
	d.Set("description", role.Description)
	d.Set("rank", role.Rank)
	return diags

}

func resourceIncidentRoleImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<role_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid role_id (%q)", parts[1])
	}
	d.Set("team", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
