package tests

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/tracing"
	_ "github.com/Azure/go-autorest/tracing/opencensus"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/compute/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

var (
	ITERATION int // The amount of iteraions for each test.
	COUNT     int // The count of each resource to provision. Defined by "ARM_PERFTEST_COUNT" environment variable.
)

func TestMain(m *testing.M) {
	if !tracing.IsEnabled() {
		log.Fatal("tracing is not enabled")
	}
	count := os.Getenv("ARM_PERFTEST_COUNT")
	if count == "" {
		log.Fatal(`"ARM_PERFTEST_COUNT" is not defined!`)
	}
	iteration := os.Getenv("ARM_PERFTEST_ITERATION")
	if count == "" {
		log.Fatal(`"ARM_PERFTEST_ITERATION" is not defined!`)
	}
	var err error
	COUNT, err = strconv.Atoi(count)
	if err != nil {
		log.Fatalf(`"ARM_PERFTEST_COUNT" is not a valid integer: %v`, err)
	}
	ITERATION, err = strconv.Atoi(iteration)
	if err != nil {
		log.Fatalf(`"ARM_PERFTEST_ITERATION" is not a valid integer: %v`, err)
	}
	os.Exit(m.Run())
}

func TestAccVirtualMachinePerf_SingleLinux(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_linux_virtual_machine", "test")

	for i := 0; i < ITERATION; i++ {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acceptance.PreCheck(t) },
			Providers:    acceptance.SupportedProviders,
			CheckDestroy: checkLinuxVirtualMachineIsDestroyed,
			Steps: []resource.TestStep{
				{
					Config: testVirtualMachinePerf_SingleLinux(data, i, COUNT),
					Check: resource.ComposeTestCheckFunc(
						checkVirtualMachinesExist(data.ResourceName, COUNT),
					),
				},
			},
		})
	}
}

func TestAccVirtualMachinePerf_SingleWindows(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_windows_virtual_machine", "test")

	for i := 0; i < ITERATION; i++ {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acceptance.PreCheck(t) },
			Providers:    acceptance.SupportedProviders,
			CheckDestroy: checkWindowsVirtualMachineIsDestroyed,
			Steps: []resource.TestStep{
				{
					Config: testVirtualMachinePerf_SingleWindows(data, i, COUNT),
					Check: resource.ComposeTestCheckFunc(
						checkVirtualMachinesExist(data.ResourceName, COUNT),
					),
				},
			},
		})
	}
}

func TestAccVirtualMachinePerf_SingleLinuxBatch(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_linux_virtual_machine", "test")

	// Temporarily set acctest parallelism to the same amount as COUNT, so
	// that VMs start to create a the same point.
	oldParallelism := os.Getenv(resource.TestParallelism)
	defer os.Setenv(resource.TestParallelism, oldParallelism)
	os.Setenv(resource.TestParallelism, strconv.Itoa(COUNT))

	for i := 0; i < ITERATION; i++ {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acceptance.PreCheck(t) },
			Providers:    acceptance.SupportedProviders,
			CheckDestroy: checkLinuxVirtualMachineIsDestroyed,
			Steps: []resource.TestStep{
				{
					Config: testVirtualMachinePerf_SingleLinuxBatch(data, i, COUNT),
					Check: resource.ComposeTestCheckFunc(
						checkVirtualMachinesExist(data.ResourceName, COUNT),
					),
				},
			},
		})
	}
}

func TestAccVirtualMachinePerf_VMSS_20Linux(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_linux_virtual_machine_scale_set", "test")

	for i := 0; i < ITERATION; i++ {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acceptance.PreCheck(t) },
			Providers:    acceptance.SupportedProviders,
			CheckDestroy: testCheckAzureRMLinuxVirtualMachineScaleSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testVirtualMachinePerf_VMSS_20Linux(data, i, COUNT),
					Check: resource.ComposeTestCheckFunc(
						checkVirtualMachineScaleSetsExist(data.ResourceName, COUNT),
					),
				},
			},
		})
	}
}

func testVirtualMachinePerf_SingleLinux(data acceptance.TestData, i, n int) string {
	template := testVirtualMachinePerf_VMTemplate(data, i, n)
	return fmt.Sprintf(`
%s

resource "azurerm_linux_virtual_machine" "test" {
  count               = %d
  name                = "acctestVM-SingleLinux-${azurerm_resource_group.test.location}-${%d + count.index}"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_D2s_v3"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurerm_network_interface.test[count.index].id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, template, n, i*n)
}

func testVirtualMachinePerf_SingleWindows(data acceptance.TestData, i, n int) string {
	template := testVirtualMachinePerf_VMTemplate(data, i, n)
	return fmt.Sprintf(`
%s

resource "azurerm_windows_virtual_machine" "test" {
  count               = %d
  name                = "SW-${azurerm_resource_group.test.location}-${%d + count.index}"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_D2s_v3"
  admin_username      = "adminuser"
  admin_password      = "P@$$w0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.test[count.index].id,
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }
}
`, template, n, i*n)
}

func testVirtualMachinePerf_SingleLinuxBatch(data acceptance.TestData, i, n int) string {
	template := testVirtualMachinePerf_VMTemplate(data, i, n)

	// We explicitly introduce a dependency here to ensure the VMs are created all together at the same time.
	var dependencies []string
	for i := 0; i < n; i++ {
		dependencies = append(dependencies, fmt.Sprintf("azurerm_network_interface.test[%d]", i))
	}
	return fmt.Sprintf(`
%s

resource "azurerm_linux_virtual_machine" "test" {
  count               = %d
  name                = "acctestVM-SingleLinuxBatch-${azurerm_resource_group.test.location}-%d-${count.index}"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_D2s_v3"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurerm_network_interface.test[count.index].id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  depends_on = [%s]
}
`, template, n, i, strings.Join(dependencies, ","))
}

func testVirtualMachinePerf_VMSS_20Linux(data acceptance.TestData, i, n int) string {
	template := testVirtualMachinePerf_BasicTemplate(data)
	return fmt.Sprintf(`
%s

resource "azurerm_linux_virtual_machine_scale_set" "test" {
  count               = %d
  name                = "VMSS-${azurerm_resource_group.test.location}-${%d + count.index}"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "Standard_D2s_v3"
  instances           = 20
  admin_username      = "adminuser"

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurerm_subnet.test.id
    }
  }
}
`, template, n, i*n)
}

func testVirtualMachinePerf_VMTemplate(data acceptance.TestData, i, n int) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_public_ip" "test" {
  count               = %[2]d
  name                = "acctpip-%[3]d-${%[4]d + count.index}"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurerm_network_interface" "test" {
  count               = %[2]d
  name                = "acctestnic-%[3]d-${%[4]d + count.index}"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
	public_ip_address_id          = azurerm_public_ip.test[count.index].id
  }
}
`, testVirtualMachinePerf_BasicTemplate(data), n, data.RandomInteger, i*n)
}

func testVirtualMachinePerf_BasicTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`
# note: whilst these aren't used in all tests, it saves us redefining these everywhere
locals {
  first_public_key  = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+wWK73dCr+jgQOAxNsHAnNNNMEMWOHYEccp6wJm2gotpr9katuF/ZAdou5AaW1C61slRkHRkpRRX9FA9CYBiitZgvCCz+3nWNN7l/Up54Zps/pHWGZLHNJZRYyAB6j5yVLMVHIHriY49d/GZTZVNB8GoJv9Gakwc/fuEZYYl4YDFiGMBP///TzlI4jhiJzjKnEvqPFki5p2ZRJqcbCiF4pJrxUQR/RXqVFQdbRLZgYfJ8xGB878RENq3yQ39d8dVOkq4edbkzwcUmwwwkYVPIoDGsYLaRHnG+To7FvMeyO7xDVQkMKzopTQV8AuKpyvpqu0a9pWOMaiCyDytO7GGN you@me.com"
  second_public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0/NDMj2wG6bSa6jbn6E3LYlUsYiWMp1CQ2sGAijPALW6OrSu30lz7nKpoh8Qdw7/A4nAJgweI5Oiiw5/BOaGENM70Go+VM8LQMSxJ4S7/8MIJEZQp5HcJZ7XDTcEwruknrd8mllEfGyFzPvJOx6QAQocFhXBW6+AlhM3gn/dvV5vdrO8ihjET2GoDUqXPYC57ZuY+/Fz6W3KV8V97BvNUhpY5yQrP5VpnyvvXNFQtzDfClTvZFPuoHQi3/KYPi6O0FSD74vo8JOBZZY09boInPejkm9fvHQqfh0bnN7B6XJoUwC1Qprrx+XIy7ust5AEn5XL7d4lOvcR14MxDDKEp you@me.com"
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestnw-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.128.0/17"
}
`, data.RandomInteger, data.Locations.Primary)
}

func checkVirtualMachinesExist(resourceBaseName string, n int) resource.TestCheckFunc {
	if n == 1 {
		// Though below function is checking only for linux VM, but also workds for windows VM.
		return checkLinuxVirtualMachineExists(resourceBaseName)
	}

	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Compute.VMClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		ch := make(chan error)
		for i := 0; i < n; i++ {
			go func(i int) {
				resourceName := fmt.Sprintf("%s.%d", resourceBaseName, i)
				rs, ok := s.RootModule().Resources[resourceName]
				if !ok {
					ch <- fmt.Errorf("Not found: %s", resourceName)
					return
				}

				id, err := parse.VirtualMachineID(rs.Primary.ID)
				if err != nil {
					ch <- err
					return
				}

				resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
				if err != nil {
					if utils.ResponseWasNotFound(resp.Response) {
						ch <- fmt.Errorf("Bad: Virtual Machine %q (Resource Group: %q) does not exist", id.Name, id.ResourceGroup)
						return
					}

					ch <- fmt.Errorf("Bad: Get on VMClient: %+v", err)
					return
				}

				ch <- nil
				return
			}(i)
		}

		for i := 0; i < n; i++ {
			if err := <-ch; err != nil {
				return err
			}
		}
		close(ch)
		return nil
	}
}

func checkVirtualMachineScaleSetsExist(resourceBaseName string, n int) resource.TestCheckFunc {
	if n == 1 {
		return testCheckAzureRMLinuxVirtualMachineScaleSetExists(resourceBaseName)
	}

	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Compute.VMScaleSetClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		ch := make(chan error)
		for i := 0; i < n; i++ {
			go func(i int) {
				resourceName := fmt.Sprintf("%s.%d", resourceBaseName, i)
				rs, ok := s.RootModule().Resources[resourceName]
				if !ok {
					ch <- fmt.Errorf("Not found: %s", resourceName)
					return
				}

				id, err := parse.VirtualMachineScaleSetID(rs.Primary.ID)
				if err != nil {
					ch <- err
					return
				}

				resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
				if err != nil {
					if utils.ResponseWasNotFound(resp.Response) {
						ch <- fmt.Errorf("Bad: Virtual Machine Scale Set %q (Resource Group: %q) does not exist", id.Name, id.ResourceGroup)
						return
					}

					ch <- fmt.Errorf("Bad: Get on VMScaleSetClient: %+v", err)
					return
				}

				ch <- nil
				return
			}(i)
		}

		for i := 0; i < n; i++ {
			if err := <-ch; err != nil {
				return err
			}
		}
		close(ch)
		return nil
	}
}
