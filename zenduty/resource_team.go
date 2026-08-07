package zenduty

import (
	"context"
	"time"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   wrapReadWith404(resourceTeamRead),
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newteam := &client.CreateTeams{}

	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		newteam.Name = v.(string)

	}
	retryErr := resource.Retry(1*time.Minute, func() *resource.RetryError {

		task, err := apiclient.Teams.CreateTeam(newteam)
		if err != nil {
			if isRetryableError(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		} else if task != nil {
			d.SetId(task.UniqueID)
		}
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}
	return diags
}

func resourceTeamUpdate(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	newteam := &client.CreateTeams{}
	id := d.Id()
	newteam.UniqueID = id
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		newteam.Name = v.(string)

	}
	retryErr := resource.Retry(1*time.Minute, func() *resource.RetryError {

		task, err := apiclient.Teams.UpdateTeam(id, newteam)
		if err != nil {
			if isRetryableError(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		} else if task != nil {
			d.SetId(task.UniqueID)
		}
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}
	return diags

}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	var diags diag.Diagnostics
	err := apiclient.Teams.DeleteTeam(id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	id := d.Id()
	var diags diag.Diagnostics

	t, err := apiclient.Teams.GetTeamByID(id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", t.Name)

	return diags
}
