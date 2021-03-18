package trafficmanager

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceTrafficManagerUserMetricsKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTrafficManagerUserMetricsKeyCreate,
		Read:   resourceTrafficManagerUserMetricsKeyRead,
		Delete: resourceTrafficManagerUserMetricsKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTrafficManagerUserMetricsKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).TrafficManager.UserMetricsKeysClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Network/trafficManagerUserMetricsKeys", subscriptionId)

	resp, err := client.Get(ctx)
	if err != nil {
		if !utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("checking for existing %s: %+v", id, err)
		}
	}

	if resp.ID != nil && *resp.ID != "" {
		return tf.ImportAsExistsError("azurerm_traffic_manager_user_metrics_key", id)
	}

	if _, err := client.CreateOrUpdate(ctx); err != nil {
		return fmt.Errorf("creating Traffic Manager User Metrics Keys: %v", err)
	}

	d.SetId(id)

	return resourceTrafficManagerUserMetricsKeyRead(d, meta)
}

func resourceTrafficManagerUserMetricsKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).TrafficManager.UserMetricsKeysClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resp, err := client.Get(ctx)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Traffic Manager User Metrics Key was not found - removing from state!")
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Traffic Manager User Metrics Key: %v", err)
	}

	if prop := resp.UserMetricsProperties; prop != nil {
		key := ""
		if prop.Key != nil {
			key = *prop.Key
		}
		d.Set("key", key)
	}

	return nil
}

func resourceTrafficManagerUserMetricsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).TrafficManager.UserMetricsKeysClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	_, err := client.Delete(ctx)
	if err != nil {
		return fmt.Errorf("deleting Traffic Manager User Metrics Key: %+v", err)
	}

	return nil
}
