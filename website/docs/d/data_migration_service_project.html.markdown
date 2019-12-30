---
subcategory: ""
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_data_migration_service_project"
sidebar_current: "docs-azurerm-datasource-data-migration-service-project"
description: |-
  Gets information about an existing Data Migration Service Project
---

# Data Source: azurerm_data_migration_service_project

Use this data source to access information about an existing Data Migration Service Project.


## Example Usage

```hcl
data "azurerm_data_migration_service_project" "example" {
  group_name   = "example-rg"
  // TODO: Unsupported property "service_name" value 
}

output "name" {
  value = "${data.azurerm_data_migration_service_project.example.name}"
}
```


## Argument Reference

The following arguments are supported:

* `group_name` - (Required) Name of the resource group

* `service_name` - (Required) Name of the service


## Attributes Reference

The following attributes are exported:

* `location` - Resource location.

* `creation_time` - UTC Date and time when project was created

* `databases_info` - One or more `databases_info` block defined below.

* `delete_running_tasks` - Delete the resource even if it contains running tasks

* `id` - Resource ID.

* `name` - Resource name.

* `project_name` - Name of the project

* `provisioning_state` - The project's provisioning state

* `source_connection_info` - One `source_connection_info` block defined below.

* `source_platform` - Source platform for the project

* `target_connection_info` - One `target_connection_info` block defined below.

* `target_platform` - Target platform for the project

* `type` - Resource type.

* `tags` - Resource tags.


---

The `databases_info` block contains the following:

* `source_database_name` - Name of the database

---

The `source_connection_info` block contains the following:

* `user_name` - User name

* `password` - Password credential.

---

The `target_connection_info` block contains the following:

* `user_name` - User name

* `password` - Password credential.
