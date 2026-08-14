---
page_title: "Zenduty: Account Roles"
subcategory: ""
description: |-
    "`zenduty_account_role` is a data source that lists the custom account roles of an account."
---

# Data Source : zenduty_account_role

`zenduty_account_role` lists the custom account roles of the account.

## Example Usage

```hcl
data "zenduty_account_role" "account_roles" {}

output "role_names" {
  value = data.zenduty_account_role.account_roles.roles[*].name
}
```

## Argument Reference

This data source takes no arguments.

## Attributes Reference

* `roles` - The list of custom account roles, each with:
    * `unique_id` - The unique_id of the role.
    * `name` - The name of the role.
    * `description` - The description of the role.
    * `permissions` - The list of permissions granted to the role.
