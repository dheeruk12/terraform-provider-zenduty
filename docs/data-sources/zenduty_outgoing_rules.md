---
page_title: "Zenduty: Outgoing Rules"
subcategory: ""
description: |-
    "`zenduty_outgoing_rules` is a data source that lists the outgoing rules of an integration."
---

# Data Source : zenduty_outgoing_rules

`zenduty_outgoing_rules` lists the outgoing rules of an integration.

## Example Usage

```hcl
data "zenduty_outgoing_rules" "integration_rules" {
  team_id        = zenduty_teams.exampleteam.id
  service_id     = zenduty_services.exampleservice.id
  integration_id = zenduty_integrations.exampleintegration.id
}
```

## Argument Reference

* `team_id` (Required) - The unique_id of the team.
* `service_id` (Required) - The unique_id of the service.
* `integration_id` (Required) - The unique_id of the integration whose outgoing rules to list.

## Attributes Reference

* `rules` - The list of outgoing rules, each with:
    * `unique_id` - The unique_id of the rule.
    * `rule_json` - The rule definition as a JSON string.
    * `enabled` - Whether the rule is enabled.
