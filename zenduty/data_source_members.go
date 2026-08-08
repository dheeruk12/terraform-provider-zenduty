package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMembers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMembersRead,
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UniqueID of the team to query members for",
			},
			"member_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The UniqueID of the specific team member to query",
			},
			"members": {
				Type:        schema.TypeList,
				Description: "List of team members",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:        schema.TypeString,
							Description: "The UniqueID of the team member",
							Computed:    true,
						},
						"team": {
							Type:        schema.TypeString,
							Description: "The UniqueID of the team",
							Computed:    true,
						},
						"user": {
							Type:        schema.TypeMap,
							Description: "The user details: username, first_name, last_name, email",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"role": {
							Type:        schema.TypeInt,
							Description: "The role of the member in the team",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceMembersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	memberID := d.Get("member_id").(string)

	if memberID != "" {
		// Get specific member
		var diags diag.Diagnostics
		member, err := apiclient.Members.GetTeamMembersByID(teamID, memberID)
		if err != nil {
			return diag.FromErr(err)
		}

		items := make([]map[string]interface{}, 1)
		item := make(map[string]interface{})
		item["unique_id"] = member.UniqueID
		item["team"] = member.Team
		item["user"] = map[string]interface{}{
			"username":   member.User.Username,
			"first_name": member.User.FirstName,
			"last_name":  member.User.LastName,
			"email":      member.User.Email,
		}
		item["role"] = member.Role

		items[0] = item
		if err := d.Set("members", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s", teamID, memberID))

		return diags
	} else {
		// Get all members for the team
		var diags diag.Diagnostics

		members, err := apiclient.Members.GetTeamMembers(teamID)
		if err != nil {
			return diag.FromErr(err)
		}

		items := make([]map[string]interface{}, len(members))
		for i, member := range members {
			item := make(map[string]interface{})
			item["unique_id"] = member.UniqueID
			item["team"] = member.Team
			item["user"] = map[string]interface{}{
				"username":   member.User.Username,
				"first_name": member.User.FirstName,
				"last_name":  member.User.LastName,
				"email":      member.User.Email,
			}
			item["role"] = member.Role

			items[i] = item
		}

		d.SetId(fmt.Sprintf("%s/%s", teamID, memberID))

		if err := d.Set("members", items); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
