package zenduty

import (
	"context"
	"errors"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAccountRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateAccountRole,
		ReadContext:   resourceReadAccountRole,
		UpdateContext: resourceUpdateAccountRole,
		DeleteContext: resourceDeleteAccountRole,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Permissions for the role. The API automatically adds the read permissions implied by the ones you list (e.g. incident_read pulls in team_read, service_read, ...); list the full stored set to avoid plan diffs.",
			},
		},
	}
}

func validateAccountRoles(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.AccountRole, diag.Diagnostics) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	permissions := d.Get("permissions").(*schema.Set).List()
	newRole := &client.AccountRole{}

	newRole.Name = name
	newRole.Description = description
	// The permission catalogue grows server-side and there is no endpoint to
	// fetch it, so unknown values are left to the API to reject.
	for _, permission := range permissions {
		if permission.(string) == "" {
			return nil, diag.FromErr(errors.New("permission must not be empty"))
		}
		newRole.Permissions = append(newRole.Permissions, permission.(string))
	}

	return newRole, nil
}

func resourceCreateAccountRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newrole, validationerr := validateAccountRoles(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	role, err := apiclient.AccountRole.CreateAccountRole(newrole)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(role.UniqueID)
	return nil
}

func resourceUpdateAccountRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	newrole, validationerr := validateAccountRoles(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	role, err := apiclient.AccountRole.UpdateAccountRole(d.Id(), newrole)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(role.UniqueID)
	return nil
}

func resourceDeleteAccountRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	err := apiclient.AccountRole.DeleteAccountRole(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceReadAccountRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	role, err := apiclient.AccountRole.GetAccountRoleByID(d.Id())
	if err != nil {
		return handleReadError(d, err)
	}
	d.SetId(role.UniqueID)
	d.Set("name", role.Name)
	d.Set("description", role.Description)
	d.Set("permissions", flattenPermissions(role.Permissions))
	return nil
}

func flattenPermissions(permissions []string) []string {
	var permissionList []string
	permissionList = append(permissionList, permissions...)
	return permissionList
}
