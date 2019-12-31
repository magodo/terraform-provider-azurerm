---
subcategory: "Data Migration"
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
}

data "azurerm_data_migration_service_project" "example" {
  name                  = "example-dms-project"
  resource_group_name   = "example-rg"
  service_name          = "example-dms"
}

output "name" {
  value = "${data.azurerm_data_migration_service_project.example.name}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the data migration service project.

* `resource_group_name` - (Required) Name of the resource group where resource belongs to.

* `service_name` - (Required) Name of the data migration service where resource belongs to.


## Attributes Reference

The following attributes are exported:

* `id` - Resource ID of data migration service project.

* `location` - Azure location where the resource exists.

* `source_platform` - Platform type of migration source.

* `target_platform` - Platform type of migration target.

* `tags` - A mapping of tags to assigned to the resource.