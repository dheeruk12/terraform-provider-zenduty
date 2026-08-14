---
page_title: "Zenduty: Team Permissions"
subcategory: ""
description: |-
    "`zenduty_team_permissions` is a data source that reads the team-level permissions of a team."
---

# Data Source : zenduty_team_permissions

`zenduty_team_permissions` reads the permissions a team grants to non-team members.

## Example Usage

```hcl
data "zenduty_team_permissions" "team_perms" {
  team_id = zenduty_teams.exampleteam.id
}
```

## Argument Reference

* `team_id` (Required) - The unique_id of the team whose permissions to read.

## Attributes Reference

* `permissions` - The list of team-level permissions granted by the team.
