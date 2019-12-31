package azurerm

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMDataMigrationServiceProject_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_migration_service_project", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataMigrationServiceProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataMigrationServiceProject_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataMigrationServiceProjectExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "source_platform", "SQL"),
					resource.TestCheckResourceAttr(data.ResourceName, "target_platform", "SQLDB"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMDataMigrationServiceProject_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_migration_service_project", "test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataMigrationServiceProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataMigrationServiceProject_complete(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataMigrationServiceProjectExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "source_platform", "SQL"),
					resource.TestCheckResourceAttr(data.ResourceName, "target_platform", "SQLDB"),
					resource.TestCheckResourceAttr(data.ResourceName, "tags.name", "test"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMDataMigrationServiceProject_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_data_migration_service_project", "test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataMigrationServiceProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataMigrationServiceProject_basic(data),
			},
			data.RequiresImportErrorStep(testAccAzureRMDataMigrationServiceProject_requiresImport),
		},
	})
}

func TestAccAzureRMDataMigrationServiceProject_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_migration_service_project", "test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataMigrationServiceProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataMigrationServiceProject_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataMigrationServiceProjectExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMDataMigrationServiceProject_complete(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataMigrationServiceProjectExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "tags.name", "test"),
				),
			},
			data.ImportStep(),
		},
	})
}

func testCheckAzureRMDataMigrationServiceProjectExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Data Migration Service Project not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		serviceName := rs.Primary.Attributes["service_name"]

		client := acceptance.AzureProvider.Meta().(*clients.Client).DataMigration.ProjectsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		if resp, err := client.Get(ctx, resourceGroup, serviceName, name); err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q) does not exist", name, serviceName, resourceGroup)
			}
			return fmt.Errorf("Bad: Get on ProjectsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMDataMigrationServiceProjectDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).DataMigration.ProjectsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_data_migration_service_project" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		serviceName := rs.Primary.Attributes["service_name"]

		if resp, err := client.Get(ctx, resourceGroup, serviceName, name); err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Get on ProjectsClient: %+v", err)
			}
		}

		return nil
	}

	return nil
}

func testAccAzureRMDataMigrationServiceProject_basic(data acceptance.TestData) string {
	template := testAccAzureRMDataMigrationService_basic(data)

	return fmt.Sprintf(`
%s

resource "azurerm_data_migration_service_project" "test" {
	name                = "acctestDmsProject-%d"
	service_name        = azurerm_data_migration_service.test.name
	resource_group_name = azurerm_resource_group.test.name
	location            = azurerm_resource_group.test.location
	source_platform     = "SQL"
	target_platform     = "SQLDB"
}
`, template, data.RandomInteger)
}

func testAccAzureRMDataMigrationServiceProject_complete(data acceptance.TestData) string {
	template := testAccAzureRMDataMigrationService_basic(data)

	return fmt.Sprintf(`
%s

resource "azurerm_data_migration_service_project" "test" {
	name                = "acctestDmsProject-%d"
	service_name        = azurerm_data_migration_service.test.name
	resource_group_name = azurerm_resource_group.test.name
	location            = azurerm_resource_group.test.location
	source_platform     = "SQL"
	target_platform     = "SQLDB"
    tags = {
 		name = "test"
    }
}
`, template, data.RandomInteger)
}

func testAccAzureRMDataMigrationServiceProject_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMDataMigrationServiceProject_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_data_migration_service_project" "import" {
  name                = azurerm_data_migration_service_project.test.name
  service_name		  = azurerm_data_migration_service_project.test.service_name
  resource_group_name = azurerm_data_migration_service_project.test.resource_group_name
  location            = azurerm_data_migration_service_project.test.location
  source_platform     =  azurerm_data_migration_service_project.test.source_platform
  target_platform     =  azurerm_data_migration_service_project.test.target_platform
}
`, template)
}
