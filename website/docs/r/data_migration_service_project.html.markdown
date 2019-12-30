---
subcategory: ""
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_data_migration_service_project"
sidebar_current: "docs-azurerm-resource-data-migration-service-project"
description: |-
  Manage Azure Project instance.
---

# azurerm_data_migration_service_project

Manage Azure Project instance.


## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-rg"
  location = "%s"
}

resource "DataMigrationServiceProject" "example" {
}
```

## Argument Reference

The following arguments are supported:

* `group_name` - (Required) Name of the resource group Changing this forces a new resource to be created.

* `location` - (Required) Resource location. Changing this forces a new resource to be created.

* `project_name` - (Required) Name of the project Changing this forces a new resource to be created.

* `service_name` - (Required) Name of the service Changing this forces a new resource to be created.

* `source_platform` - (Required) Source platform for the project

* `target_platform` - (Required) Target platform for the project

* `databases_info` - (Optional) One or more `databases_info` block defined below.

* `delete_running_tasks` - (Optional) Delete the resource even if it contains running tasks Changing this forces a new resource to be created.

* `source_connection_info` - (Optional) One `source_connection_info` block defined below.

* `target_connection_info` - (Optional) One `target_connection_info` block defined below.

* `tags` - (Optional) Resource tags. Changing this forces a new resource to be created.

---

The `databases_info` block supports the following:

* `source_database_name` - (Required) Name of the database

---

The `source_connection_info` block supports the following:

* `user_name` - (Optional) User name

* `password` - (Optional) Password credential.

---

The `target_connection_info` block supports the following:

* `user_name` - (Optional) User name

* `password` - (Optional) Password credential.

## Attributes Reference

The following attributes are exported:

* `creation_time` - UTC Date and time when project was created

* `provisioning_state` - The project's provisioning state

* `id` - Resource ID.

* `name` - Resource name.

* `type` - Resource type.


## Import

Data Migration Service Project can be imported using the `resource id`, e.g.

```shell
$ terraform import azurerm_data_migration_service_project.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.DataMigration/services//projects/
```
