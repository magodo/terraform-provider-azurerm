package azurerm

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"

	"github.com/Azure/azure-sdk-for-go/services/datamigration/mgmt/2018-04-19/datamigration"

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
					resource.TestCheckResourceAttr(data.ResourceName, "source_databases.#", "2"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.additional_settings", "foo"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.authentication", string(datamigration.SQLAuthentication)),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.data_source", `tcp:localhost\sourceServer,12345`),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.encrypt_connection", "true"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.password", "secret"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.platform", string(datamigration.SQLOnPrem)),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.trust_server_certificate", "true"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_source_connection_info.user_name", "root"),
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
					testCheckAzureRMDataMigrationServiceExists(data.ResourceName),
				),
			},
			{
				Config: testAccAzureRMDataMigrationServiceProject_complete(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataMigrationServiceProjectExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "source_platform", "SQL"),
					resource.TestCheckResourceAttr(data.ResourceName, "target_platform", "SQLDB"),
					resource.TestCheckResourceAttr(data.ResourceName, "source_databases.#", "2"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.additional_settings", "bar"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.authentication", string(datamigration.SQLAuthentication)),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.data_source", `tcp:localhost\targetServer,12345`),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.encrypt_connection", "true"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.password", "secret"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.platform", string(datamigration.SQLOnPrem)),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.trust_server_certificate", "true"),
					resource.TestCheckResourceAttr(data.ResourceName, "sql_target_connection_info.user_name", "root"),
					resource.TestCheckResourceAttr(data.ResourceName, "tags.name", "test"),
				),
			},
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
    source_databases    = ["db1", "db2"]
	sql_source_connection_info {
		additional_settings = "foo"
		authentication = "SqlAuthentication"
		data_source = "tcp:localhost\\sourceServer,12345"
		encrypt_connection = true
		password = "secret"
		platform = "SqlOnPrem"
		trust_server_certificate = true
		user_name = "root"
	}
	sql_source_connection_info {
		additional_settings = "bar"
		authentication = "SqlAuthentication"
		data_source = "tcp:localhost\\targetServer,12345"
		encrypt_connection = true
		password = "secret"
		platform = "SqlOnPrem"
		trust_server_certificate = true
		user_name = "root"
	}
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
