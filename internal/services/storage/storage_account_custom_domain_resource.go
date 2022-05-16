package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-04-01/storage"
	"github.com/hashicorp/terraform-provider-azurerm/internal/locks"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type StorageAccountCustomDomainResource struct{}

var _ sdk.ResourceWithUpdate = StorageAccountCustomDomainResource{}

type StorageAccountCustomDomainModel struct {
	Name             string `tfschema:"name"`
	StorageAccountId string `tfschema:"storage_account_id"`
	UseSubdomain     bool   `tfschema:"use_subdomain"`
}

func (r StorageAccountCustomDomainResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"storage_account_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.StorageAccountID,
		},
		"use_subdomain": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
		},
	}
}

func (r StorageAccountCustomDomainResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r StorageAccountCustomDomainResource) ResourceType() string {
	return "azurerm_storage_account_custom_domain"
}

func (r StorageAccountCustomDomainResource) ModelObject() interface{} {
	return &StorageAccountCustomDomainModel{}
}

func (r StorageAccountCustomDomainResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.StorageAccountCustomDomainID
}

func (r StorageAccountCustomDomainResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Storage.AccountsClient

			var plan StorageAccountCustomDomainModel
			if err := metadata.Decode(&plan); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			accountId, err := parse.StorageAccountID(plan.StorageAccountId)
			if err != nil {
				return err
			}

			id := parse.NewStorageAccountCustomDomainID(accountId.SubscriptionId, accountId.ResourceGroup, accountId.Name, plan.Name)

			storageAccount, err := client.GetProperties(ctx, id.ResourceGroup, id.StorageAccountName, "")
			if err != nil {
				return fmt.Errorf("retrieving %s: %v", accountId, err)
			}
			if storageAccount.AccountProperties == nil {
				return fmt.Errorf("unexpected nil properties for %s", accountId)
			}
			if storageAccount.AccountProperties.CustomDomain != nil {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			params := storage.AccountUpdateParameters{
				AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
					CustomDomain: &storage.CustomDomain{
						Name:             &plan.Name,
						UseSubDomainName: &plan.UseSubdomain,
					},
				},
			}
			locks.ByName(accountId.Name, storageAccountResourceName)
			defer locks.UnlockByName(accountId.Name, storageAccountResourceName)

			if _, err = client.Update(ctx, id.ResourceGroup, id.StorageAccountName, params); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r StorageAccountCustomDomainResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,

		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Storage.AccountsClient
			id, err := parse.StorageAccountCustomDomainID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}
			accountId := parse.NewStorageAccountID(id.SubscriptionId, id.ResourceGroup, id.StorageAccountName)

			existing, err := client.GetProperties(ctx, id.ResourceGroup, id.StorageAccountName, "")
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}
			if existing.AccountProperties == nil {
				return fmt.Errorf("unexpected nil properties for %s", accountId)
			}
			if existing.AccountProperties.CustomDomain == nil {
				if utils.ResponseWasNotFound(existing.Response) {
					return metadata.MarkAsGone(id)
				}
			}

			model := StorageAccountCustomDomainModel{
				StorageAccountId: accountId.ID(),
			}

			if v := existing.AccountProperties.CustomDomain.Name; v != nil {
				model.Name = *v
			}
			if v := existing.AccountProperties.CustomDomain.UseSubDomainName; v != nil {
				model.UseSubdomain = *v
			}

			return metadata.Encode(&model)
		},
	}
}

func (r StorageAccountCustomDomainResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := parse.StorageAccountCustomDomainID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}
			accountId := parse.NewStorageAccountID(id.SubscriptionId, id.ResourceGroup, id.StorageAccountName)

			var plan StorageAccountCustomDomainModel
			if err := metadata.Decode(&plan); err != nil {
				return err
			}

			client := metadata.Client.Storage.AccountsClient

			storageAccount, err := client.GetProperties(ctx, id.ResourceGroup, id.StorageAccountName, "")
			if err != nil {
				return fmt.Errorf("retrieving %s: %v", accountId, err)
			}
			if storageAccount.AccountProperties == nil {
				return fmt.Errorf("unexpected nil properties for %s", accountId)
			}
			if storageAccount.AccountProperties.CustomDomain == nil {
				return fmt.Errorf("%s not exists", id)
			}

			customDomainProp := storageAccount.AccountProperties.CustomDomain
			params := storage.AccountUpdateParameters{
				AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
					CustomDomain: customDomainProp,
				},
			}
			if metadata.ResourceData.HasChange("use_subdomain") {
				params.AccountPropertiesUpdateParameters.CustomDomain.UseSubDomainName = &plan.UseSubdomain
			}

			locks.ByName(accountId.Name, storageAccountResourceName)
			defer locks.UnlockByName(accountId.Name, storageAccountResourceName)

			if _, err = client.Update(ctx, id.ResourceGroup, id.StorageAccountName, params); err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}

			return nil
		},
	}
}

func (r StorageAccountCustomDomainResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Storage.AccountsClient

			id, err := parse.StorageAccountCustomDomainID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}
			accountId := parse.NewStorageAccountID(id.SubscriptionId, id.ResourceGroup, id.StorageAccountName)

			// storageAccount, err := client.GetProperties(ctx, id.ResourceGroup, id.StorageAccountName, "")
			// if err != nil {
			// 	return fmt.Errorf("retrieving %s: %v", accountId, err)
			// }
			// if storageAccount.AccountProperties == nil {
			// 	return fmt.Errorf("unexpected nil properties for %s", accountId)
			// }
			// if storageAccount.AccountProperties.CustomDomain == nil {
			// 	return nil
			// }

			// storageAccount.AccountProperties.CustomDomain = nil

			params := storage.AccountUpdateParameters{
				AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
					CustomDomain: &storage.CustomDomain{},
				},
			}

			locks.ByName(accountId.Name, storageAccountResourceName)
			defer locks.UnlockByName(accountId.Name, storageAccountResourceName)
			if _, err = client.Update(ctx, id.ResourceGroup, id.StorageAccountName, params); err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			return nil
		},
	}
}
