package zenduty

import (
	"context"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGlobalRouter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlobalRouterCreate,
		UpdateContext: resourceGlobalRouterUpdate,
		DeleteContext: resourceGlobalRouterDelete,
		ReadContext:   resourceGlobalRouterRead,
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
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceGlobalRouterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newGlobalRouter := &client.GlobalRouterPayload{}
	var diags diag.Diagnostics

	if v, ok := d.GetOk("name"); ok {
		newGlobalRouter.Name = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		newGlobalRouter.Description = v.(string)
	}
	newGlobalRouter.IsEnabled = d.Get("is_enabled").(bool)

	router, err := apiclient.GlobalRouter.CreateGlobalRouter(newGlobalRouter)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(router.UniqueID)
	// added integration_key in response output
	d.Set("integration_key", router.IntegrationKey)
	d.Set("is_enabled", router.IsEnabled)
	d.Set("description", router.Description)

	return diags
}

func resourceGlobalRouterUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newGlobalRouter := &client.GlobalRouterPayload{}
	var diags diag.Diagnostics

	if v, ok := d.GetOk("name"); ok {
		newGlobalRouter.Name = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		newGlobalRouter.Description = v.(string)
	}
	newGlobalRouter.IsEnabled = d.Get("is_enabled").(bool)

	router, err := apiclient.GlobalRouter.UpdateGlobalRouter(d.Id(), newGlobalRouter)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(router.UniqueID)
	// added integration_key in response output
	d.Set("integration_key", router.IntegrationKey)
	d.Set("is_enabled", router.IsEnabled)
	d.Set("description", router.Description)

	return diags
}

func resourceGlobalRouterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	var diags diag.Diagnostics
	err := apiclient.GlobalRouter.DeleteGlobalRouter(id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceGlobalRouterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	id := d.Id()
	apiclient, _ := m.(*Config).Client()

	router, err := apiclient.GlobalRouter.GetGlobalRouter(id)
	if err != nil {
		return handleReadError(d, err)
	}
	d.Set("name", router.Name)
	d.Set("integration_key", router.IntegrationKey)
	d.Set("is_enabled", router.IsEnabled)
	d.Set("description", router.Description)

	return diags
}
