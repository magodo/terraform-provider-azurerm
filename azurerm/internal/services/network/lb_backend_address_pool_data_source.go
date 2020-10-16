package network

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
)

func dataSourceArmLoadBalancerBackendAddressPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmLoadBalancerBackendAddressPoolRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"loadbalancer_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"backend_ip_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

func dataSourceArmLoadBalancerBackendAddressPoolRead(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Network.LoadBalancerBackendAddressPoolsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return fmt.Errorf("parsing Load Balancer Name and Group: %+v", err)
	}

	resp, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("nil or empty id of Load Balancer Backend Address Pool %q (Resource Group %q / Load Balancer: %q): %+v", name, loadBalancerId.ResourceGroup, loadBalancerId.Name, err)
	}

	id, err := parse.LoadBalancerBackendAddressPoolID(*resp.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID(subscriptionId))

	d.Set("name", id.Name)
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
