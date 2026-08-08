package zenduty

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAccountRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountRolesRead,

		Schema: map[string]*schema.Schema{
			"roles": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"permissions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceAccountRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var diags diag.Diagnostics

	roles, err := apiclient.AccountRole.GetAccountRoles()
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		items[i] = map[string]interface{}{
			"unique_id":   role.UniqueID,
			"name":        role.Name,
			"description": role.Description,
			"permissions": role.Permissions,
		}
	}

	if err := d.Set("roles", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("account_roles")

	return diags
}
