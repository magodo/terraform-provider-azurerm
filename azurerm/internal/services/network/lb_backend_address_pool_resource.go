package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-05-01/network"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmLoadBalancerBackendAddressPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmLoadBalancerBackendAddressPoolCreateOrUpdate,
		Update: resourceArmLoadBalancerBackendAddressPoolCreateOrUpdate,
		Read:   resourceArmLoadBalancerBackendAddressPoolRead,
		Delete: resourceArmLoadBalancerBackendAddressPoolDelete,

		Importer: loadBalancerSubResourceImporter(func(input string) (*parse.LoadBalancerId, error) {
			id, err := parse.LoadBalancerBackendAddressPoolID(input)
			if err != nil {
				return nil, err
			}

			lbId := parse.NewLoadBalancerID(id.ResourceGroup, id.LoadBalancerName)
			return &lbId, nil
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			// TODO 3.0: remove this as "loadbalancer_id" already provide the resource group info
			"resource_group_name": azure.SchemaResourceGroupNameDeprecated(),

			"loadbalancer_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"backend_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"virtual_network_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validate.VirtualNetworkID,
						},

						"ip_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPAddress,
						},

						"network_interface_ip_configuration": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"backend_ip_configurations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Set: schema.HashString,
			},

			"load_balancing_rules": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Set: schema.HashString,
			},

			"outbound_rules": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceArmLoadBalancerBackendAddressPoolCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.LoadBalancerBackendAddressPoolsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return fmt.Errorf("parsing Load Balancer Name and Group: %+v", err)
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %v",
					name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_lb_backend_address_pool", *existing.ID)
		}
	}

	param := network.BackendAddressPool{
		Name: &name,
		BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{
			LoadBalancerBackendAddresses: expandArmLoadBalancerBackendAddresses(d.Get("backend_addresses").(*schema.Set).List()),
		},
	}

	future, err := client.CreateOrUpdate(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, name, param)
	if err != nil {
		return fmt.Errorf("creating/updating Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
	}

	read, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
	}
	if read.ID == nil || *read.ID == "" {
		return fmt.Errorf("nil or empty ID of Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
	}

	poolId, err := parse.LoadBalancerBackendAddressPoolID(*read.ID)
	if err != nil {
		return err
	}

	d.SetId(poolId.ID(subscriptionId))

	return resourceArmLoadBalancerBackendAddressPoolRead(d, meta)
}

func resourceArmLoadBalancerBackendAddressPoolRead(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Network.LoadBalancerBackendAddressPoolsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerBackendAddressPoolID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", id.Name, id.ResourceGroup, id.LoadBalancerName, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("loadbalancer_id", parse.NewLoadBalancerID(id.ResourceGroup, id.LoadBalancerName).ID(subscriptionId))

	var backendIpConfigurations []string
	var loadBalancingRules []string
	var outboudRules []string

	if props := resp.BackendAddressPoolPropertiesFormat; props != nil {
		if err := d.Set("backend_addresses", flattenArmLoadBalancerBackendAddresses(props.LoadBalancerBackendAddresses)); err != nil {
			return fmt.Errorf("setting `backend_address`: %v", err)
		}

		if configs := props.BackendIPConfigurations; configs != nil {
			for _, backendConfig := range *configs {
				backendIpConfigurations = append(backendIpConfigurations, *backendConfig.ID)
			}
		}

		if rules := props.LoadBalancingRules; rules != nil {
			for _, rule := range *rules {
				loadBalancingRules = append(loadBalancingRules, *rule.ID)
			}
		}

		if rules := props.OutboundRules; rules != nil {
			for _, rule := range *rules {
				outboudRules = append(outboudRules, *rule.ID)
			}
		}
	}
	d.Set("backend_ip_configurations", backendIpConfigurations)
	d.Set("load_balancing_rules", loadBalancingRules)
	d.Set("outbound_rules", outboudRules)

	return nil
}

func resourceArmLoadBalancerBackendAddressPoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.LoadBalancerBackendAddressPoolsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerBackendAddressPoolID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.LoadBalancerName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", id.Name, id.ResourceGroup, id.LoadBalancerName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", id.Name, id.ResourceGroup, id.LoadBalancerName, err)
	}
	return nil
}

func expandArmLoadBalancerBackendAddresses(input []interface{}) *[]network.LoadBalancerBackendAddress {
	result := make([]network.LoadBalancerBackendAddress, 0)

	for _, e := range input {
		if e == nil {
			continue
		}
		v := e.(map[string]interface{})

		address := network.LoadBalancerBackendAddress{
			Name: utils.String(v["name"].(string)),
		}

		if v["virtual_network_id"] != nil || v["ip_address"] != nil {
			address.LoadBalancerBackendAddressPropertiesFormat = &network.LoadBalancerBackendAddressPropertiesFormat{
				VirtualNetwork: &network.SubResource{ID: utils.String(v["virtual_network_id"].(string))},
				IPAddress:      utils.String(v["ip_address"].(string)),
			}
		}
		result = append(result, address)
	}

	return &result
}

func flattenArmLoadBalancerBackendAddresses(input *[]network.LoadBalancerBackendAddress) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make([]interface{}, 0)

	for _, e := range *input {
		var name string
		if e.Name != nil {
			name = *e.Name
		}

		var (
			ipAddress string
			vnetId    string
			ipConfig  string
		)
		if prop := e.LoadBalancerBackendAddressPropertiesFormat; prop != nil {
			if prop.IPAddress != nil {
				ipAddress = *prop.IPAddress
			}
			if prop.VirtualNetwork != nil && prop.VirtualNetwork.ID != nil {
				vnetId = *prop.VirtualNetwork.ID
			}
			if prop.NetworkInterfaceIPConfiguration != nil && prop.NetworkInterfaceIPConfiguration.ID != nil {
				ipConfig = *prop.NetworkInterfaceIPConfiguration.ID
			}
		}

		v := map[string]interface{}{
			"name":                               name,
			"virtual_network_id":                 vnetId,
			"ip_address":                         ipAddress,
			"network_interface_ip_configuration": ipConfig,
		}
		output = append(output, v)
	}

	return output
}
