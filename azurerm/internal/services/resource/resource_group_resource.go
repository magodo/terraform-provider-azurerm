package resource

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resource/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceResourceGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceResourceGroupCreateUpdate,
		Read:   resourceResourceGroupRead,
		Update: resourceResourceGroupCreateUpdate,
		Delete: resourceResourceGroupDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ResourceGroupID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(90 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(90 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"tags": tags.Schema(),
		},
	}
}

func resourceResourceGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient2
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	if d.IsNewResource() {
		existing, err := client.Get(ctx, name, nil)
		if err != nil {
			v, ok := err.(azcore.HTTPResponse)
			if !ok || v.RawResponse().StatusCode != http.StatusNotFound {
				return fmt.Errorf("Error checking for presence of existing resource group: %+v", err)
			}
		} else {
			if existing.ResourceGroup != nil && existing.ResourceGroup.ID != nil && *existing.ResourceGroup.ID != "" {
				return tf.ImportAsExistsError("azurerm_resource_group", *existing.ResourceGroup.ID)
			}
		}
	}

	parameters := armresources.ResourceGroup{
		Location: utils.String(location),
		Tags:     tags.Expand(t),
	}

	if _, err := client.CreateOrUpdate(ctx, name, parameters, nil); err != nil {
		return fmt.Errorf("Error creating Resource Group %q: %+v", name, err)
	}

	resp, err := client.Get(ctx, name, nil)
	if err != nil {
		return fmt.Errorf("Error retrieving Resource Group %q: %+v", name, err)
	}

	if resp.ResourceGroup == nil || resp.ResourceGroup.ID == nil || *resp.ResourceGroup.ID == "" {
		return fmt.Errorf("unexpected nil ID for the Resource Group %q", name)
	}
	d.SetId(*resp.ResourceGroup.ID)

	return resourceResourceGroupRead(d, meta)
}

func resourceResourceGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading resource group %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading resource group: %+v", err)
	}

	d.Set("name", resp.Name)
	d.Set("location", location.NormalizeNilable(resp.Location))
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceResourceGroupDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGroupID(d.Id())
	if err != nil {
		return err
	}

	deleteFuture, err := client.Delete(ctx, id.ResourceGroup, "")
	if err != nil {
		if response.WasNotFound(deleteFuture.Response()) {
			return nil
		}

		return fmt.Errorf("Error deleting Resource Group %q: %+v", id.ResourceGroup, err)
	}

	err = deleteFuture.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		if response.WasNotFound(deleteFuture.Response()) {
			return nil
		}

		return fmt.Errorf("Error deleting Resource Group %q: %+v", id.ResourceGroup, err)
	}

	return nil
}
