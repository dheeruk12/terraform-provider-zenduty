---
page_title: "Zenduty: SLAs"
subcategory: ""
description: |-
    "`zenduty_sla` is a data source that lists the SLAs of a team."
---

# Data Source : zenduty_sla

`zenduty_sla` lists the SLAs of a team.

## Example Usage

```hcl
data "zenduty_sla" "team_slas" {
  team_id = zenduty_teams.exampleteam.id
}

output "sla_names" {
  value = data.zenduty_sla.team_slas.slas[*].name
}
```

## Argument Reference

* `team_id` (Required) - The unique_id of the team whose SLAs to list.

## Attributes Reference

* `slas` - The list of SLAs, each with:
    * `unique_id` - The unique_id of the SLA.
    * `name` - The name of the SLA.
    * `creation_date` - When the SLA was created.
    * `acknowledge_time` - Seconds before the acknowledgement SLA is breached.
    * `resolve_time` - Seconds before the resolution SLA is breached.
    * `is_active` - Whether the SLA is active.

The list endpoint does not return `description`, `conditions` or `escalations`.
Reference the `zenduty_sla` resource, or import it, if you need those.
