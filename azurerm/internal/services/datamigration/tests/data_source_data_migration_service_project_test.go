package azurerm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
)

func TestAccDataSourceAzureRMDataMigrationServiceProject_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_migration_service_project", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDataMigrationServiceProject_basic(data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data.ResourceName, "source_platform", "SQL"),
					resource.TestCheckResourceAttr(data.ResourceName, "target_platform", "SQLDB"),
				),
			},
		},
	})
}

func testAccDataSourceDataMigrationServiceProject_basic(data acceptance.TestData) string {
	config := testAccAzureRMDataMigrationServiceProject_basic(data)
	return fmt.Sprintf(`
%s

data "azurerm_data_migration_service_project" "test" {
  name                  = azurerm_data_migration_service_project.test.name
  service_name          = azurerm_data_migration_service_project.test.service_name
  resource_group_name   = azurerm_data_migration_service_project.test.resource_group_name
}
`, config)
}
