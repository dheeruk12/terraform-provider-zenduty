---
page_title: "Zenduty Alert Rules"
subcategory: ""
description: |-
    " `zenduty_alertrules` is a resource to manage alert rules in a integration"
---
# zenduty_alertrules (Resource)
`zenduty_alertrules` is a resource to manage alert rules in a integration


## Example Usage

```hcl
resource "zenduty_teams" "exampleteam" {
    name = "exmaple team"
}

resource "zenduty_services" "exampleservice" {
    name = "example service"
    team_id = zenduty_teams.exampleteam.id 
    escalation_policy = zenduty_esp.example_esp.id e
}

resource "zenduty_integrations" "exampleintegration" {
    team_id = zenduty_teams.exampleteam.id
    service_id = zenduty_services.exampleservice.id
    application = ""
    name = "exampleintegration"
    summary = "This is the summary for the example integration"
}

```

```hcl 
resource "zenduty_alertrules" "example_alertrules" {
  
    description = "This is the description for the new alert rules"
    team_id = zenduty_teams.exampleteam.id
    service_id = zenduty_services.exampleservice.id
    integration_id = zenduty_integrations.exampleintegration.id
    rule_json = "" 
    #actions
}

```

## Argument Reference

* `team_id` (Required, Forces new resource) - The unique_id of the team to create the alert rule in.
* `service_id` (Required, Forces new resource) - The unique_id of the service to create the alert rule in.
* `integration_id` (Required, Forces new resource) - The unique_id of the integration to create the alert rule in.
* `description` (Required) - The description of the alert rule.
* `rule_json` (Required)(string) - The rule json of the alert rule.You cannot construct the rule json in terraform as of now.One can construct the rule json in Zenduty's UI.Create an dummy alert rule in Zenduty and copy the rule_json from the UI.
* `stop` (Optional)(bool) - Stop evaluating later alert rules when this rule matches. Defaults to `false`.
* `rule_type` (Optional)(Number) - `0` requires all conditions to match, `1` requires any condition to match.
* `position` (Optional)(Number) - Evaluation order of the rule. Assigned by the server when omitted.
* `conditions` (Optional) - Structured conditions evaluated against incoming alerts. (see [below for nested schema](#nestedblock--conditions))
* `actions` (Optional) - The actions to be performed when the rule matches. (see [below for nested schema](#nestedblock--actions))

<a id="nestedblock--conditions"></a>

## Conditions

```hcl
    conditions {
        alert_condition_type = 2
        alert_field          = "summary"
        pattern              = "CRITICAL"
    }
```

* `alert_condition_type` (Optional)(Number) - `1` matches the alert type, `2` matches a payload field. Defaults to `1`. New types may appear server-side; only the lower bound is validated in Terraform.
* `alert_field` (Required)(string) - The field the condition inspects: the alert attribute for type `1`, the payload field name for type `2`.
* `pattern` (Optional)(string) - The pattern the field is matched against.
* `unique_id` (Computed)(string) - The unique_id of the condition. The API replaces all conditions on every update, so this value changes between applies.

Condition order matters: the API assigns each condition's position from its
place in the list. When the config declares no `conditions` blocks, existing
conditions (e.g. created in the Zenduty UI) are preserved on update — which
also means they cannot be removed by simply omitting the blocks. Note that
most configurations express matching via `rule_json` instead — the API stores
`conditions` and `rule_json` independently, so keep them consistent with each
other if you use both.


<a id="nestedblock--actions"></a>

## Actions
```hcl
    actions {
        action_type = ""
        value = ""
        #key
    }

```
* `action_type` (Required) (Number):
    * `1` - change the alert type , value should be one of the following: `0` for info, `1` for warning, `2` for error, `3` for critical , `4` for acknowledged , `5` for resolved
    * `2` - add note , value will be the note summary to add
    * `3` - supress alert , value is not required
    * `4` - add escalation policy , set `escalation_policy` (or value) to the unique_id of the escalation policy
    * `5` - assign schedule , set `schedule` (or value) to the unique_id of the schedule
    * `6` - assign user , set `assign_to` (or value) to the username of the user
    * `7` - change urgency  , value should be one of the following: `0` for low, `1` for high
    * `8` - change message , value should be the message to change to 
    * `9` - change summary , value should be the summary to change to
    * `10` - change entry_id , value should be the entity to change to
    * `11` - assign role to user , `key` should be unique_id of the role , value should be the username of the user
    * `12` - Add tag. The value must be comma-separated.
        * For existing tags, use the unique_id.
        * For dynamic tags, use placeholders in {{ }} format.
    * `13` - create task , value should be the task description
    * `14` - add sla , set `sla` (or value) to the unique_id of the sla
    * `15` - add team priority , set `team_priority` (or value) to the unique_id of the team priority
    * `16` - add task template , set `task_template` (or value) to the unique_id of the task template
    * `17` - add assign incident responder , value should be the unique_id of the responder
    * `18` - hash entity_id, value is not required
    * `19` - delay notifications , value should be the delay in seconds

    New action types are added server-side over time; unrecognised types are passed
    through using `value` and validated by the API.

* `value` (Optional)(string) - The value of the action. Not required for `3` and `18`, and not required when the type-specific attribute below is used instead.
* `key`  (Optional)(string) - The key of the action. (required for `11`)

The following attributes are typed alternatives to the overloaded `value`. Each
applies to one action type, and setting both it and `value` is an error.

* `escalation_policy` (Optional)(string) - Escalation policy to assign (action type `4`).
* `schedule` (Optional)(string) - Schedule to assign (action type `5`).
* `assign_to` (Optional)(string) - User to assign (action type `6`).
* `sla` (Optional)(string) - SLA to assign (action type `14`).
* `team_priority` (Optional)(string) - Team priority to assign (action type `15`).
* `task_template` (Optional)(string) - Task template to assign (action type `16`).

```hcl
    actions {
        action_type       = 4
        escalation_policy = zenduty_esp.example_esp.id
    }

    actions {
        action_type = 5
        schedule    = zenduty_schedules.example_schedule.id
    }
```


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Alert Rule.

## Import

Integrations can be imported using the `team_id`(ie. unique_id of the team), `service_id`(ie. unique_id of the service),`integration_id`(ie. unique_id of the integration) and `alertrule_id`(ie. unique_id of the alert rule).

```hcl
resource "zenduty_alertrules" "rule1" {

}

```

`$ terraform import zenduty_alertrules.rule1 team_id/service_id/integration_id/alertrule_id` 

`$ terraform state show zenduty_alertrules.rule1`

`* copy the output data and paste inside zenduty_alertrules.rule1 resource block and remove the id attribute`

`$ terraform plan` to verify the import



    

