package zenduty

import (
	"context"
	"strings"
	"sync"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// every zenduty_applications data source lists the same catalog, and a plan
// can hold dozens of them reading concurrently; fetch it once per (base URL,
// token) and share the result. The mutex is held across the fetch so parallel
// first readers wait for the one in flight instead of stampeding the API.
var applicationsCatalogMu sync.Mutex
var applicationsCatalog = map[string][]client.CatalogApplication{}

func getApplicationsCatalog(apiclient *client.Client) ([]client.CatalogApplication, error) {
	key := apiclient.Config.BaseURL + "|" + apiclient.Config.Token
	applicationsCatalogMu.Lock()
	defer applicationsCatalogMu.Unlock()
	if apps, ok := applicationsCatalog[key]; ok {
		return apps, nil
	}
	apps, err := apiclient.Applications.GetApplications()
	if err != nil {
		return nil, err
	}
	applicationsCatalog[key] = apps
	return apps, nil
}

func dataSourceApplications() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the application in the integration catalog (e.g. \"Grafana v8\")." +
					" Application unique ids differ between Zenduty instances, so looking one up by" +
					" name keeps a config portable. The match is case-insensitive.",
			},
			"summary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceApplicationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	name := d.Get("name").(string)
	if emptyString(name) {
		return diag.Errorf("name must not be empty")
	}

	apps, err := getApplicationsCatalog(apiclient)
	if err != nil {
		return diag.Errorf("Error listing applications: %s", err)
	}

	for _, app := range apps {
		if strings.EqualFold(app.Name, name) {
			d.SetId(app.UniqueID)
			d.Set("summary", app.Summary)
			d.Set("application_type", app.ApplicationType)
			return nil
		}
	}
	return diag.Errorf("no application named %q in the catalog (%d applications available)", name, len(apps))
}
