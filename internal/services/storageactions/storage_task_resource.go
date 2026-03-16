package storageactions

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/go-azure-sdk/resource-manager/storageactions/2023-01-01/storageactions"
	"github.com/hashicorp/go-azure-sdk/resource-manager/storageactions/2023-01-01/storagetasks"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type StorageActionsStorageTaskResource struct{}

var _ sdk.ResourceWithUpdate = StorageActionsStorageTaskResource{}

type StorageActionsStorageTaskModel struct {
	Name          string                                     `tfschema:"name"`
	ResourceGroup string                                     `tfschema:"resource_group_name"`
	Location      string                                     `tfschema:"location"`
	Identity      []identity.ModelSystemAssignedUserAssigned `tfschema:"identity"`
	Action        []StorageActionsStorageTaskActionModel     `tfschema:"action"`
	Enabled       bool                                       `tfschema:"enabled"`
	Description   string                                     `tfschema:"description"`
	Tags          map[string]string                          `tfschema:"tags"`
}

type StorageActionsStorageTaskActionModel struct {
	If   []StorageActionsStorageTaskActionIfModel   `tfschema:"if"`
	Else []StorageActionsStorageTaskActionElseModel `tfschema:"else"`
}

type StorageActionsStorageTaskActionIfModel struct {
	Condition  string                                    `tfschema:"condition"`
	Operations []StorageActionsStorageTaskOperationModel `tfschema:"operations"`
}

type StorageActionsStorageTaskActionElseModel struct {
	Operations []StorageActionsStorageTaskOperationModel `tfschema:"operations"`
}

type StorageActionsStorageTaskOperationModel struct {
	Name       string            `tfschema:"name"`
	Parameters map[string]string `tfschema:"parameters"`
}

func (r StorageActionsStorageTaskResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile(`^[a-z0-9]{3,18}$`),
				"Storage task name must be between 3 and 18 characters in length and use numbers and lower-case letters only.",
			),
		},
		"resource_group_name": commonschema.ResourceGroupName(),
		"location":            commonschema.Location(),
		"identity":            commonschema.SystemAssignedUserAssignedIdentityRequired(),
		"action": {
			Type:     pluginsdk.TypeList,
			Required: true,
			MinItems: 1,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"if": {
						Type:     pluginsdk.TypeList,
						Required: true,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"condition": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: validation.StringIsNotEmpty,
								},
								"operations": {
									Type:     pluginsdk.TypeList,
									Required: true,
									MinItems: 1,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringInSlice(storagetasks.PossibleValuesForStorageTaskOperationName(), false),
											},
											"parameters": {
												Type:     pluginsdk.TypeMap,
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
					"else": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"operations": {
									Type:     pluginsdk.TypeList,
									Required: true,
									MinItems: 1,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringInSlice(storagetasks.PossibleValuesForStorageTaskOperationName(), false),
											},
											"parameters": {
												Type:     pluginsdk.TypeMap,
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},
		"description": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"tags": commonschema.Tags(),
	}
}

func (r StorageActionsStorageTaskResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r StorageActionsStorageTaskResource) ResourceType() string {
	return "azurerm_storage_task"
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

			var model StorageActionsStorageTaskModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			id := storagetasks.NewStorageTaskID(subscriptionId, model.ResourceGroup, model.Name)
			existing, err := client.Get(ctx, id)
			if err != nil {
				if !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
				}
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			identityModel, err := identity.ExpandLegacySystemAndUserAssignedMapFromModel(model.Identity)
			if err != nil {
				return fmt.Errorf("expanding the system and user assigned identity: %v", err)
			}

			params := storagetasks.StorageTask{
				Location: location.Normalize(model.Location),
				Properties: storagetasks.StorageTaskProperties{
					Action:      r.expandAction(model.Action),
					Description: model.Description,
					Enabled:     model.Enabled,
				},
				Identity: *identityModel,
				Tags:     &model.Tags,
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

			resp, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			model := StorageActionsStorageTaskModel{}

			if respModel := resp.Model; respModel != nil {
				if tags := respModel.Tags; tags != nil {
					model.Tags = *respModel.Tags
				}
			}

			return metadata.Encode(&model)
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

			var model StorageActionsStorageTaskModel
			if err := metadata.Decode(&model); err != nil {
				return err
			}

			client := metadata.Client.StorageActions.StorageTasks

			existing, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}
			if existing.Model == nil {
				return fmt.Errorf("retrieving %s: `model` was nil", *id)
			}

			params := *existing.Model

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

func (r StorageActionsStorageTaskResource) expandOperations(input []StorageActionsStorageTaskOperationModel) []storagetasks.StorageTaskOperation {
	var output []storagetasks.StorageTaskOperation

	for _, model := range input {
		output = append(output, storagetasks.StorageTaskOperation{
			Name:       storagetasks.StorageTaskOperationName(model.Name),
			Parameters: &model.Parameters,
			OnFailure:  pointer.To(storagetasks.OnFailureBreak),
			OnSuccess:  pointer.To(storagetasks.OnSuccessContinue),
		})
	}

	return output
}

func (r StorageActionsStorageTaskResource) flattenOperations(input []storagetasks.StorageTaskOperation) []StorageActionsStorageTaskOperationModel {
	var output []StorageActionsStorageTaskOperationModel

	for _, op := range input {
		var params map[string]string
		if op.Parameters != nil {
			params = *op.Parameters
		}

		output = append(output, StorageActionsStorageTaskOperationModel{
			Name:       string(op.Name),
			Parameters: params,
		})
	}

	return output
}

func (r StorageActionsStorageTaskResource) expandIf(input []StorageActionsStorageTaskActionIfModel) *storagetasks.IfCondition {
	if len(input) == 0 {
		return nil
	}

	model := input[0]

	return &storagetasks.IfCondition{
		Condition:  model.Condition,
		Operations: r.expandOperations(model.Operations),
	}
}

func (r StorageActionsStorageTaskResource) flattenIf(input storagetasks.IfCondition) []StorageActionsStorageTaskActionIfModel {
	return []StorageActionsStorageTaskActionIfModel{
		{
			Condition:  input.Condition,
			Operations: r.flattenOperations(input.Operations),
		},
	}
}

func (r StorageActionsStorageTaskResource) expandElse(input []StorageActionsStorageTaskActionElseModel) *storagetasks.ElseCondition {
	if len(input) == 0 {
		return nil
	}

	model := input[0]

	return &storagetasks.ElseCondition{
		Operations: r.expandOperations(model.Operations),
	}
}

func (r StorageActionsStorageTaskResource) flattenElse(input *storagetasks.ElseCondition) []StorageActionsStorageTaskActionElseModel {
	if input == nil {
		return nil
	}

	return []StorageActionsStorageTaskActionElseModel{
		{
			Operations: r.flattenOperations(input.Operations),
		},
	}
}

func (r StorageActionsStorageTaskResource) expandAction(input []StorageActionsStorageTaskActionModel) storagetasks.StorageTaskAction {
	if len(input) == 0 {
		// This shouldn't happen as is guaranteed by the schema.
		return storagetasks.StorageTaskAction{}
	}

	model := input[0]

	ifCondition := r.expandIf(model.If)

	// This is guaranteed by the schema definition, but added here anyway to avoid panic.
	if ifCondition == nil {
		return nil
	}

	return &storagetasks.StorageTaskAction{
		If:   *ifCondition,
		Else: r.expandElse(model.Else),
	}
}

func (r StorageActionsStorageTaskResource) flattenAction(input *storagetasks.StorageTaskAction) []StorageActionsStorageTaskActionModel {
	if input == nil {
		return nil
	}

	return []StorageActionsStorageTaskActionModel{
		{
			If:   r.flattenIf(input.If),
			Else: r.flattenElse(input.Else),
		},
	}
}
