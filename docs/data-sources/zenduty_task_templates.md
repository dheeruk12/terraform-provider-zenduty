---
page_title: "Zenduty: Task Templates"
subcategory: ""
description: |-
    "`zenduty_task_templates` is a data source that lists the task templates of a team."
---

# Data Source : zenduty_task_templates

`zenduty_task_templates` lists the task templates of a team.

## Example Usage

```hcl
data "zenduty_task_templates" "team_templates" {
  team_id = zenduty_teams.exampleteam.id
}
```

## Argument Reference

* `team_id` (Required) - The unique_id of the team whose task templates to list.

## Attributes Reference

* `task_templates` - The list of task templates, each with:
    * `unique_id` - The unique_id of the task template.
    * `name` - The name of the task template.
    * `summary` - The summary of the task template.
    * `creation_date` - When the task template was created.
