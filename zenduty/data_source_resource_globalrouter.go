package zenduty

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGlobalRouter() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalRouterRead,
		Schema: map[string]*schema.Schema{
			"router_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The UniqueID of the global router to query",
			},
			"global_routers": {
				Type:        schema.TypeList,
				Description: "List of global routers",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:        schema.TypeString,
							Description: "The UniqueID of the global router",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the global router",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "The description of the global router",
							Computed:    true,
						},
						"integration_key": {
							Type:        schema.TypeString,
							Description: "The integration key of the global router",
							Computed:    true,
							Sensitive:   true,
						},
						"is_enabled": {
							Type:        schema.TypeBool,
							Description: "Whether the global router is enabled",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGlobalRouterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	routerID := d.Get("router_id").(string)
	if routerID != "" {
		var diags diag.Diagnostics
		router, err := apiclient.GlobalRouter.GetGlobalRouter(routerID)
		if err != nil {
			return diag.FromErr(err)
		}
		items := make([]map[string]interface{}, 1)
		item := make(map[string]interface{})
		item["unique_id"] = router.UniqueID
		item["name"] = router.Name
		item["description"] = router.Description
		item["integration_key"] = router.IntegrationKey
		item["is_enabled"] = router.IsEnabled
		items[0] = item
		if err := d.Set("global_routers", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(time.Now().String())

		return diags
	} else {
		var diags diag.Diagnostics

		routers, err := apiclient.GlobalRouter.GetGlobalRouters()
		if err != nil {
			return diag.FromErr(err)
		}

		items := make([]map[string]interface{}, len(routers))
		for i, router := range routers {
			item := make(map[string]interface{})
			item["unique_id"] = router.UniqueID
			item["name"] = router.Name
			item["description"] = router.Description
			item["integration_key"] = router.IntegrationKey
			item["is_enabled"] = router.IsEnabled
			items[i] = item
		}

		if err := d.Set("global_routers", items); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(time.Now().String())

		return diags
	}
}
