---
page_title: "Zenduty: Post Incident Tasks"
subcategory: ""
description: |-
    "`zenduty_post_incident_tasks` is a data source that lists the post-incident tasks of a team."
---

# Data Source : zenduty_post_incident_tasks

`zenduty_post_incident_tasks` lists the post-incident tasks of a team.

## Example Usage

```hcl
data "zenduty_post_incident_tasks" "team_tasks" {
  team_id = zenduty_teams.exampleteam.id
}
```

## Argument Reference

* `team_id` (Required) - The unique_id of the team whose post-incident tasks to list.

All pages of the API's paginated response are fetched, so the list is complete.

## Attributes Reference

* `tasks` - The list of tasks, each with:
    * `unique_id` - The unique_id of the task.
    * `title` - The title of the task.
    * `description` - The description of the task.
    * `status` - The status of the task.
    * `assigned_to` - The username of the assignee.
    * `due_in_time` - When the task is due.
    * `creation_date` - When the task was created.
