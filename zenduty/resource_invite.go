package zenduty

import (
	"context"
	"time"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceInvite() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "zenduty_invite is deprecated and will be removed in a future release: it only fires invitations on create and tracks no server state. Use the zenduty_user resource instead.",
		CreateContext:      resourceInviteCreate,
		UpdateContext:      resourceInviteUpdate,
		DeleteContext:      resourceInviteDelete,
		ReadContext:        wrapReadWith404(resourceInviteRead),
		Schema: map[string]*schema.Schema{
			"team": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email_accounts": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:     schema.TypeString,
							Required: true,
						},
						"first_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"last_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeInt,
							Required: true,
							// account roles: 2 admin, 3 user (1 owner is not assignable)
							ValidateFunc: validation.IntBetween(2, 3),
							Description:  "Account role of the invited user: 2 (admin) or 3 (user).",
						},
					},
				},
			},
		},
	}
}

func resourceInviteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newinvite := &client.Invite{}
	if v, ok := d.GetOk("team"); ok {
		newinvite.Team = v.(string)
	}
	emailAccounts := d.Get("email_accounts").([]interface{})
	var diags diag.Diagnostics
	newinvite.EmailAccounts = make([]client.EmailAccounts, len(emailAccounts))
	for i, user := range emailAccounts {
		emailAccount := user.(map[string]interface{})
		if v, ok := emailAccount["email"]; ok {
			newinvite.EmailAccounts[i].Email = v.(string)
		}
		if v, ok := emailAccount["first_name"]; ok {
			newinvite.EmailAccounts[i].FirstName = v.(string)
		}
		if v, ok := emailAccount["last_name"]; ok {
			newinvite.EmailAccounts[i].LastName = v.(string)
		}
		if v, ok := emailAccount["role"]; ok {
			newinvite.EmailAccounts[i].Role = v.(int)
		}

	}
	_, err := apiclient.Invite.CreateInvite(newinvite)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(time.Now().String())

	return diags
}

func resourceInviteUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	return diags
}

func resourceInviteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	return diags
}

func resourceInviteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags

}
