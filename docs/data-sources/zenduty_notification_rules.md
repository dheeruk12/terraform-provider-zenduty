---
page_title: "Zenduty: Notification Rules"
subcategory: ""
description: |-
    "`zenduty_notification_rules` is a data source that lists a user's notification rules."
---

# Data Source : zenduty_notification_rules

`zenduty_notification_rules` lists the notification rules of a user.

## Example Usage

```hcl
data "zenduty_user" "user1" {
  email = "demouser@gmail.com"
}

data "zenduty_notification_rules" "user_rules" {
  username = data.zenduty_user.user1.users[0].username
}
```

## Argument Reference

* `username` (Required) - The username of the user whose notification rules to list.

## Attributes Reference

* `notification_rules` - The list of notification rules, each with:
    * `unique_id` - The unique_id of the notification rule.
    * `contact` - The unique_id of the contact method to notify.
    * `urgency` - The urgency the rule applies to: `0` for low, `1` for high.
    * `start_delay` - Seconds to wait before notifying.
    * `type` - The type of the notification rule.
