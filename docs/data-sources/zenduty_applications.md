---
page_title: "zenduty_applications Data Source - terraform-provider-zenduty"
subcategory: ""
description: |-
  Look up an application in the integration catalog by name.
---

# zenduty_applications (Data Source)

Looks up an application in the account's integration application catalog by name and exposes its unique id.

Application unique ids differ between Zenduty instances, so hardcoding one makes a config unportable — an id exported from one account will not exist on another. Referencing the application through this data source resolves the id on whatever instance the config is applied to.

## Example Usage

```hcl
data "zenduty_applications" "grafana" {
  name = "Grafana v8"
}

resource "zenduty_integrations" "grafana" {
  team_id     = zenduty_teams.devops.id
  service_id  = zenduty_services.frontend.id
  application = data.zenduty_applications.grafana.id
  name        = "Grafana v8"
}
```

## Argument Reference

* `name` - (Required) The name of the application in the integration catalog (e.g. `"Grafana v8"`). The match is case-insensitive.

## Attribute Reference

* `id` - The unique id of the application on this instance.
* `summary` - The catalog summary of the application.
* `application_type` - The application's type.