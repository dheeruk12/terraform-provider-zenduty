package zenduty

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserReads,
		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"users": {
				Type:        schema.TypeList,
				Description: "List of User",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"first_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUserReads(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	email := d.Get("email").(string)

	var diags diag.Diagnostics

	users, err := apiclient.Users.GetUsers(email)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(users) == 0 {
		return diag.FromErr(fmt.Errorf("no users found with email %s", email))
	}
	items := make([]map[string]interface{}, len(users))
	for i, user := range users {
		item := make(map[string]interface{})
		item["unique_id"] = user.UniqueID
		item["email"] = user.User.Email
		item["first_name"] = user.User.FirstName
		item["last_name"] = user.User.LastName
		item["username"] = user.User.Username
		items[i] = item

	}
	if err := d.Set("users", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(email)

	return diags
}
