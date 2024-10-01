package tfxsdk

import (
	"context"
	"encoding/json"
	"maps"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

// DefinitionConfigUpgraders is a collection of DefinitionConfigUpgrader, from the schema version 0 to the latest supported version. Each DefinitionConfigUpgraders upgrades the definition config one version forward.
//
// The schema version expects to be continuouse within the major version of the provider.
type DefinitionConfigUpgraders map[int]DefinitionConfigUpgrader

// DefinitionConfigUpgrader upgrades the definition config from the current schema version to the next version
type DefinitionConfigUpgrader struct {
	DefinitionConfigUpgrader func(context.Context, UpgradeDefinitionConfigRequest, *UpgradeDefinitionConfigResponse)
}

type UpgradeDefinitionConfigRequest struct {
	// The syntax body that descirbes this definition
	SyntaxBody *hclsyntax.Body
	// The hclwrite.Body that the user is supposed to make change to
	WriteBody *hclwrite.Body
	// State can be nil for modules without state, or resource whose address contains the index
	State map[string]interface{}
}

type UpgradeDefinitionConfigResponse struct {
	// State represents the upgraded states to the next schema version.
	// The provider is expected to invoke its StateUpgrader to retrieve the result.
	State map[string]interface{}
	Error error
}

type DefinitionFixers map[BlockType]map[string]DefinitionConfigUpgraders

type FixConfigDefinitionFunction struct {
	Fixers DefinitionFixers
}

var _ function.Function = FixConfigDefinitionFunction{}

func NewFixConfigDefinitionFunction(fixers DefinitionFixers) function.Function {
	return &FixConfigDefinitionFunction{Fixers: fixers}
}

func (a FixConfigDefinitionFunction) Metadata(_ context.Context, _ function.MetadataRequest, response *function.MetadataResponse) {
	response.Name = "terrafix_config_definition"
}

func (a FixConfigDefinitionFunction) Definition(_ context.Context, _ function.DefinitionRequest, response *function.DefinitionResponse) {
	response.Definition = function.Definition{
		Summary:             "Fix a Terraform config definition",
		Description:         "Fix a Terraform config definition for a provider, resource or data source",
		MarkdownDescription: "Fix a Terraform config definition for a provider, resource or data source",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "block_type",
				Description:         "Block type: provider, resource, datasource",
				MarkdownDescription: "Block type: provider, resource, datasource",
			},
			function.StringParameter{
				Name:                "block_name",
				Description:         "The block name (e.g. provider name, resource type)",
				MarkdownDescription: "The block name (e.g. provider name, resource type)",
			},
			function.Int64Parameter{
				Name:                "version",
				Description:         "The version of the schema, inferred from the Terraform state",
				MarkdownDescription: "The version of the schema, inferred from the Terraform state",
			},
			function.StringParameter{
				Name:                "raw_content",
				Description:         "The content of the block definition",
				MarkdownDescription: "The content of the block definition",
			},
			function.StringParameter{
				Name:                "raw_state",
				Description:         "(Optional) The content of the block's terraform state. Only for resource or data source",
				MarkdownDescription: "(Optional) The content of the block's terraform state. Only for resource or data source",
			},
		},
		Return: function.StringReturn{},
	}
}

func (a FixConfigDefinitionFunction) Run(ctx context.Context, request function.RunRequest, response *function.RunResponse) {
	var blockType, blockName string
	var version int64
	var rawContent, rawState string

	response.Error = function.ConcatFuncErrors(request.Arguments.Get(ctx, &blockType, &blockName, &version, &rawContent, &rawState))
	if response.Error != nil {
		return
	}

	var state map[string]interface{}
	if rawState != "" {
		var tstate tfjson.StateResource
		if err := json.Unmarshal([]byte(rawState), &tstate); err != nil {
			response.Error = function.NewFuncError(err.Error())
			return
		}
		state = tstate.AttributeValues
	}

	if m, ok := a.Fixers[BlockType(blockType)]; ok {
		if us, ok := m[blockName]; ok {
			versions := slices.Sorted(maps.Keys(us))
			version := int(version)
			idx, err := SchemaVersioIndex(versions, version)
			if err != nil {
				response.Error = function.NewFuncError(err.Error())
				return
			}
			if idx != -1 {
				for _, v := range versions[idx:] {
					u := us[v]

					sf, diags := hclsyntax.ParseConfig([]byte(rawContent), "", hcl.InitialPos)
					if diags.HasErrors() {
						response.Error = function.NewFuncError(diags.Error())
						return
					}
					sbody := sf.Body.(*hclsyntax.Body).Blocks[0].Body

					wf, diags := hclwrite.ParseConfig([]byte(rawContent), "", hcl.InitialPos)
					if diags.HasErrors() {
						response.Error = function.NewFuncError(diags.Error())
						return
					}
					wbody := wf.Body().Blocks()[0].Body()

					var resp UpgradeDefinitionConfigResponse
					req := UpgradeDefinitionConfigRequest{
						SyntaxBody: sbody,
						WriteBody:  wbody,
						State:      state,
					}
					u.DefinitionConfigUpgrader(ctx, req, &resp)
					if resp.Error != nil {
						response.Error = function.NewFuncError(resp.Error.Error())
						return
					}

					// Update rawContent and state
					rawContent = string(wf.Bytes())
					state = resp.State
				}
			}
		}
	}

	response.Error = function.ConcatFuncErrors(response.Result.Set(ctx, rawContent))
}
