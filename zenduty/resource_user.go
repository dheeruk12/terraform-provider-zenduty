package zenduty

import (
	"context"
	"fmt"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateUser,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUpdateUser,
		DeleteContext: resourceDeleteUser,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"team": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
				// The API treats team as the invite destination on create and
				// never returns it, so an imported user has no value in state;
				// suppress that one-sided diff or every import plans a
				// replacement.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "" && d.Id() != ""
				},
			},
			"first_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateRequired(),
			},
			"last_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateRequired(),
			},
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateEmail(),
			},
			"role": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(2, 3),
				Default:      3,
			},
		},
	}
}

func resourceCreateUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	team := d.Get("team").(string)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)
	email := d.Get("email").(string)
	role := d.Get("role").(int)
	apiclient, _ := m.(*Config).Client()
	newUser := &client.UserObj{FirstName: firstName, LastName: lastName, Email: email, Role: role}
	newUserobj := &client.CreateUser{Team: team, User: *newUser}

	user, err := apiclient.Users.CreateUser(newUserobj)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.User.Username)
	// The create response is a team-member object whose role is the TEAM role
	// enum (1 manager, 2 user), not the account role — keep the planned value;
	// Read reports the account role from the account-member endpoint.
	return nil
}

func resourceUpdateUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	role := d.Get("role").(int)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)
	email := d.Get("email").(string)
	apiclient, _ := m.(*Config).Client()

	newUser := &client.UserObj{FirstName: firstName, LastName: lastName, Email: email, Role: role}

	user, err := apiclient.Users.UpdateUser(d.Id(), newUser)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.User.Username)
	d.Set("role", user.Role)
	d.Set("first_name", user.User.FirstName)
	d.Set("last_name", user.User.LastName)
	d.Set("email", user.User.Email)
	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	apiclient, _ := m.(*Config).Client()
	user, err := apiclient.Users.GetUser(d.Id())
	if err != nil {
		return handleReadError(d, err)
	}
	d.SetId(user.User.Username)
	d.Set("role", user.Role)
	d.Set("first_name", user.User.FirstName)
	d.Set("last_name", user.User.LastName)
	d.Set("email", user.User.Email)
	return nil
}

// The Zenduty API has no endpoint to delete or deactivate an account member,
// so destroy can only forget the user from state.
func resourceDeleteUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "zenduty_user cannot be deleted via the API",
		Detail:   fmt.Sprintf("User %s was removed from Terraform state, but the account member still exists in Zenduty and must be removed from the web console.", d.Id()),
	}}
}
