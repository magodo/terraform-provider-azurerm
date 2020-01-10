---
subcategory: "Data Migration"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_data_migration_service_project"
sidebar_current: "docs-azurerm-resource-data-migration-service-project"
description: |-
  Manage Azure Data Migration Project instance.
---

# azurerm_data_migration_service_project

Manage a Azure Data Migration Project.

~> **NOTE on destroy behavior of Data Migration Service Project:** Destroy a Data Migration Service Project will leave any outstanding tasks untouched. This is to avoid unexpectedly delete any tasks managed out of terraform.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-rg"
  location = "West Europe"
}

resource "azurerm_virtual_network" "example" {
  name                = "example-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "example" {
  name                 = "example-subnet"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurerm_data_migration_service" "example" {
	name                = "example-dms"
	location            = azurerm_resource_group.example.location
	resource_group_name = azurerm_resource_group.example.name
	virtual_subnet_id   = azurerm_subnet.example.id
	sku_name            = "Standard_1vCores"
}

resource "azurerm_data_migration_service_project" "example" {
	name                = "example-dms-project"
	service_name        = azurerm_data_migration_service.example.name
	resource_group_name = azurerm_resource_group.example.name
	location            = zurerm_resource_group.example.location
	source_platform     = "SQL"
	target_platform     = "SQLDB"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specify the name of the data migration service project. Changing this forces a new resource to be created.

* `service_name` - (Required) Name of the data migration service where resource belongs to. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) Name of the resource group in which to create the data migration service project. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `source_platform` - (Required) Platform type of migration source. Currently only support: `SQL`(on-premises SQL Server). Changing this forces a new resource to be created.

* `target_platform` - (Required) Platform type of migration target. Currently only support: `SQLDB`(Azure SQL Database). Changing this forces a new resource to be created.

* `tags` - (Optional) A mapping of tags to assigned to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of Data Migration Service Project.

## Import

Data Migration Service Projects can be imported using the `resource id`, e.g.

```shell
$ terraform import azurerm_data_migration_service_project.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg/providers/Microsoft.DataMigration/services/example-dms/projects/project1
```
