package trafficmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type TrafficManagerUserMetricsKeyResource struct{}

func TestAccAzureRMTrafficManagerUserMetricsKey_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_traffic_manager_user_metrics_key", "test")
	r := TrafficManagerUserMetricsKeyResource{}

	data.ResourceSequentialTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("key").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMTrafficManagerUserMetricsKey_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_traffic_manager_user_metrics_key", "test")
	r := TrafficManagerUserMetricsKeyResource{}

	data.ResourceSequentialTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (r TrafficManagerUserMetricsKeyResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	client := clients.TrafficManager.UserMetricsKeysClient

	if resp, err := client.Get(ctx); err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Traffic Manager User Metrics Key: %+v", err)
	}

	return utils.Bool(true), nil
}

func (r TrafficManagerUserMetricsKeyResource) basic(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_traffic_manager_user_metrics_key" "test" {}
`, template)
}

func (r TrafficManagerUserMetricsKeyResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_traffic_manager_user_metrics_key" "import" {}
`, template)
}

func (r TrafficManagerUserMetricsKeyResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
`)
}
