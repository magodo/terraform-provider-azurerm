package maps

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/maps/mgmt/2020-02-01-preview/maps"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/maps/validate"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/maps/parse"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceMapsAccountCreator() *schema.Resource {
	return &schema.Resource{
		Create: resourceMapsAccountCreatorCreateUpdate,
		Read:   resourceMapsAccountCreatorRead,
		Update: resourceMapsAccountCreatorCreateUpdate,
		Delete: resourceMapsAccountCreatorDelete,

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.CreatorID(id)
			return err
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
				ValidateFunc: validate.CreatorName(),
			},

			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AccountID,
			},

			"location": location.Schema(),

			"tags": tags.Schema(),
		},
	}
}

func resourceMapsAccountCreatorCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Maps.CreatorsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	accountId, err := parse.AccountID(d.Get("account_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewCreatorID(accountId.SubscriptionId, accountId.ResourceGroup, accountId.Name, name)

	location := azure.NormalizeLocation(d.Get("location").(string))

	if d.IsNewResource() {
		resp, err := client.Get(ctx, id.ResourceGroup, id.AccountName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("checking for existing Maps Account Creator %q: %+v", id, err)
			}
		}

		if resp.ID != nil && *resp.ID != "" {
			id, err := parse.CreatorID(*resp.ID)
			if err != nil {
				return err
			}
			return tf.ImportAsExistsError("azurerm_maps_account_creator", id.ID())
		}
	}

	param := maps.CreatorCreateParameters{
		Location: &location,
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.AccountName, id.Name, param); err != nil {
		return fmt.Errorf("creating/updating Maps Account Creator %q: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceMapsAccountCreatorRead(d, meta)
}

func resourceMapsAccountCreatorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Maps.CreatorsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CreatorID(d.Id())
	if err != nil {
		return err
	}
	accountId := parse.NewAccountID(id.SubscriptionId, id.ResourceGroup, id.AccountName)

	resp, err := client.Get(ctx, id.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Maps Account Creator %q was not found - removing from state!", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Maps Account Creator %q: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("account_id", accountId.ID())
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMapsAccountCreatorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Maps.CreatorsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CreatorID(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(ctx, id.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Maps Account Creator %q: %+v", id, err)
	}

	return nil
}
