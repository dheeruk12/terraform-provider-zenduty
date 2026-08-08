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

func resourceMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberCreate,
		ReadContext:   wrapReadWith404(resourceMemberRead),
		UpdateContext: resourceMemberUpdate,
		DeleteContext: resourceMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMemberImporter,
		},
		Schema: map[string]*schema.Schema{
			"team": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
				// team roles, not account roles: 1 manager, 2 user
				ValidateFunc: validation.IntBetween(1, 2),
				Description:  "Team role of the member: 1 (manager) or 2 (user).",
			},
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newMembers := &client.Member{}
	role := d.Get("role").(int)
	if role == 0 {
		newMembers.Role = 2
	} else {
		newMembers.Role = role
	}
	var diags diag.Diagnostics
	if v, ok := d.GetOk("team"); ok {
		newMembers.Team = v.(string)

	}
	if v, ok := d.GetOk("user"); ok {
		newMembers.User = v.(string)
	}

	member, err := apiclient.Members.CreateTeamMember(newMembers.Team, newMembers)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(member.UniqueID)
	return diags
}

func resourceMemberUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newMembers := &client.Member{}
	id := d.Id()
	newMembers.UniqueID = id
	var diags diag.Diagnostics
	if v, ok := d.GetOk("user"); ok {
		newMembers.User = v.(string)
	}
	if v, ok := d.GetOk("role"); ok {

		if v.(int) == 0 {
			newMembers.Role = 2
		} else {
			newMembers.Role = v.(int)
		}
	}
	if v, ok := d.GetOk("team"); ok {
		newMembers.Team = v.(string)
	}
	_, err := apiclient.Members.UpdateTeamMember(newMembers)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	team := d.Get("team").(string)
	var diags diag.Diagnostics
	err := apiclient.Members.DeleteTeamMember(team, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	team := d.Get("team").(string)
	var diags diag.Diagnostics
	member, err := apiclient.Members.GetTeamMembersByID(team, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("team", member.Team)
	// The API accepts a username, email, or id in "user" but always returns
	// the username. Keep the config's spelling when it still identifies the
	// same person, so the choice of identifier is not a diff — but fall
	// through to the username otherwise, so a genuine change of member is
	// still reported as drift.
	prior := d.Get("user").(string)
	if prior != member.User.Username && prior != member.User.Email {
		d.Set("user", member.User.Username)
	}
	d.Set("role", member.Role)

	return diags
}

func resourceMemberImporter(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Import format: <team_id>/<member_id>
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<member_id>", d.Id())
	}

	teamID := parts[0]
	memberID := parts[1]

	// Validate UUIDs
	if !IsValidUUID(teamID) {
		return nil, fmt.Errorf("invalid team_id (%q)", teamID)
	}
	if !IsValidUUID(memberID) {
		return nil, fmt.Errorf("invalid member_id (%q)", memberID)
	}

	d.SetId(memberID)
	d.Set("team", teamID)

	return []*schema.ResourceData{d}, nil
}
