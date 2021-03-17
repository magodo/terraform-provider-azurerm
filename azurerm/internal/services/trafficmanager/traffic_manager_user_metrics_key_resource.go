package trafficmanager

import (
	"fmt"
	"log"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/trafficmanager/parse"

	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceTrafficManagerUserMetricsKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTrafficManagerUserMetricsKeyCreateUpdate,
		Read:   resourceTrafficManagerUserMetricsKeyRead,
		Update: resourceTrafficManagerUserMetricsKeyCreateUpdate,
		Delete: resourceTrafficManagerUserMetricsKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{},
	}
}

func resourceTrafficManagerUserMetricsKeyCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).TrafficManager.UserMetricsKeysClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Network/trafficManagerUserMetricsKeys", subscriptionId)

	if d.IsNewResource() {
		resp, err := client.Get(ctx)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("checking for existing %s: %+v", id, err)
			}
		}

		if resp.ID != nil && *resp.ID != "" {
			return tf.ImportAsExistsError("azurerm_traffic_manager_user_metrics_key", id)
		}
	}

	// TODO: d.Get() && Create

	d.SetId(id)

	return resourceTrafficManagerUserMetricsKeyRead(d, meta)
}

func resourceTrafficManagerUserMetricsKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).TrafficManager.UserMetricsKeysClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.TrafficManagerUserMetricsKeyID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.TrafficManagerUserMetricsKeyName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state!", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	// TODO: d.Set()

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceTrafficManagerUserMetricsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).TrafficManager.UserMetricsKeyClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.TrafficManagerUserMetricsKeyID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.TrafficManagerUserMetricsKeyName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
	}

	return nil
}
