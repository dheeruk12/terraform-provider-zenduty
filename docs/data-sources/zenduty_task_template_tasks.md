---
page_title: "Zenduty: Task Template Tasks"
subcategory: ""
description: |-
    "`zenduty_task_template_tasks` is a data source that lists the tasks of a task template."
---

# Data Source : zenduty_task_template_tasks

`zenduty_task_template_tasks` lists the tasks of a task template.

## Example Usage

```hcl
data "zenduty_task_template_tasks" "template_tasks" {
  team_id          = zenduty_teams.exampleteam.id
  task_template_id = zenduty_task_templates.example_template.id
}
```

## Argument Reference

* `team_id` (Required) - The unique_id of the team.
* `task_template_id` (Required) - The unique_id of the task template whose tasks to list.

## Attributes Reference

* `tasks` - The list of tasks, each with:
    * `unique_id` - The unique_id of the task.
    * `title` - The title of the task.
    * `description` - The description of the task.
    * `role` - The unique_id of the incident role the task is assigned to.
    * `due_in` - Minutes until the task is due, or `-1` for no due date.
    * `position` - The position of the task within the template.
    * `creation_date` - When the task was created.
