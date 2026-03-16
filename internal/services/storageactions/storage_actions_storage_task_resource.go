package storageactions

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/go-azure-sdk/resource-manager/storageactions/2023-01-01/storagetasks"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type StorageActionsStorageTaskResource struct{}

var _ sdk.ResourceWithUpdate = StorageActionsStorageTaskResource{}

type StorageActionsStorageTaskModel struct {
	Name          string                                     `tfschema:"name"`
	ResourceGroup string                                     `tfschema:"resource_group_name"`
	Location      string                                     `tfschema:"location"`
	Identity      []identity.ModelSystemAssignedUserAssigned `tfschema:"identity"`
	Tags          map[string]string                          `tfschema:"tags"`
}

func (r StorageActionsStorageTaskResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},
		"resource_group_name": commonschema.ResourceGroupName(),
		"location":            commonschema.Location(),
		"identity":            commonschema.SystemAssignedUserAssignedIdentityOptional(),
		"tags":                commonschema.Tags(),
	}
}

func (r StorageActionsStorageTaskResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r StorageActionsStorageTaskResource) ResourceType() string {
	return "azurerm_storage_actions_storage_task"
}

func (r StorageActionsStorageTaskResource) ModelObject() interface{} {
	return &StorageActionsStorageTaskModel{}
}

func (r StorageActionsStorageTaskResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return storagetasks.ValidateStorageTaskID
}

func (r StorageActionsStorageTaskResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.StorageActions.StorageTasks
			subscriptionId := metadata.Client.Account.SubscriptionId

			var plan StorageActionsStorageTaskModel
			if err := metadata.Decode(&plan); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			id := storagetasks.NewStorageTaskID(subscriptionId, plan.ResourceGroup, plan.Name)
			existing, err := client.Get(ctx, id)
			if err != nil {
				if !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
				}
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			identityModel, err := identity.ExpandLegacySystemAndUserAssignedMapFromModel(plan.Identity)
			if err != nil {
				return fmt.Errorf("expanding the system and user assigned identity: %v", err)
			}
			params := storagetasks.StorageTask{
				Location:   location.Normalize(plan.Location),
				Properties: storagetasks.StorageTaskProperties{},
				Identity:   *identityModel,
				Tags:       &plan.Tags,
			}

			if err := client.CreateThenPoll(ctx, id, params); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r StorageActionsStorageTaskResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,

		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.StorageActions.StorageTasks
			id, err := storagetasks.ParseStorageTaskID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			existing, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(existing.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			state := StorageActionsStorageTaskModel{}

			if model := existing.Model; model != nil {
				if tags := model.Tags; tags != nil {
					state.Tags = *model.Tags
				}
			}

			return metadata.Encode(&state)
		},
	}
}

func (r StorageActionsStorageTaskResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := storagetasks.ParseStorageTaskID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var plan StorageActionsStorageTaskModel
			if err := metadata.Decode(&plan); err != nil {
				return err
			}

			client := metadata.Client.StorageActions.StorageTasks

			resp, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}
			if resp.Model == nil {
				return fmt.Errorf("unexpected nil model returned")
			}

			params := *resp.Model

			// TODO: update the params
			// if props := params.Properties; props != nil {
			// 	if metadata.ResourceData.HasChange("xxx") {
			// 		props.Xxx = plan.Xxx
			// 	}

			if err := client.CreateThenPoll(ctx, *id, params); err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}
			return nil
		},
	}
}

func (r StorageActionsStorageTaskResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.StorageActions.StorageTasks

			id, err := storagetasks.ParseStorageTaskID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if err := client.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			return nil
		},
	}
}
