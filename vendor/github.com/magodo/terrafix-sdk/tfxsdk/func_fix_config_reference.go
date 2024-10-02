package tfxsdk

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ReferenceConfigUpgraders is a collection of ReferenceConfigUpgrader, from the schema version 0 to the latest supported version. Each ReferenceConfigUpgraders upgrades the reference config one version forward.
//
// The schema version expects to be continuouse within the major version of the provider.
type ReferenceConfigUpgraders map[int]ReferenceConfigUpgrader

// ReferenceConfigUpgrader upgrades the reference config from the current schema version to the next version
type ReferenceConfigUpgrader struct {
	ReferenceConfigUpgrader func(context.Context, UpgradeReferenceConfigRequest, *UpgradeReferenceConfigResponse)
}

type UpgradeReferenceConfigRequest struct {
	Traversals []hcl.Traversal
}

type UpgradeReferenceConfigResponse struct {
	Traversals []hcl.Traversal
	Error      error
}

type ReferenceFixers map[BlockType]map[string]ReferenceConfigUpgraders

type FixConfigReferenceFunction struct {
	Fixers ReferenceFixers
}

var _ function.Function = FixConfigReferenceFunction{}

// NewFixConfigReferenceFunction returns the provider function for fixing the config reference.
func NewFixConfigReferenceFunction(fixers ReferenceFixers) function.Function {
	return &FixConfigReferenceFunction{Fixers: fixers}
}

func (a FixConfigReferenceFunction) Metadata(_ context.Context, _ function.MetadataRequest, response *function.MetadataResponse) {
	response.Name = "terrafix_config_references"
}

func (a FixConfigReferenceFunction) Definition(_ context.Context, _ function.DefinitionRequest, response *function.DefinitionResponse) {
	response.Definition = function.Definition{
		Summary:             "Fix Terraform config reference origins",
		Description:         "Fix Terraform config reference origins targeting to a provider, resource or data source",
		MarkdownDescription: "Fix Terraform config reference origins targeting to a provider, resource or data source",
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
			function.ListParameter{
				Name:                "raw_contents",
				Description:         "The list of reference origin contents",
				MarkdownDescription: "The list of reference origin contents",
				ElementType:         basetypes.StringType{},
			},
		},
		Return: function.ListReturn{
			ElementType: basetypes.StringType{},
		},
	}
}

func (a FixConfigReferenceFunction) Run(ctx context.Context, request function.RunRequest, response *function.RunResponse) {
	var blockType, blockName string
	var version int
	var rawContents []string

	response.Error = function.ConcatFuncErrors(request.Arguments.Get(ctx, &blockType, &blockName, &version, &rawContents))
	if response.Error != nil {
		return
	}

	// Parsing traversals
	var traversals []hcl.Traversal
	for _, content := range rawContents {
		expr, diags := hclsyntax.ParseExpression([]byte(content), "", hcl.InitialPos)
		if diags.HasErrors() {
			response.Error = function.NewFuncError(diags.Error())
			return
		}
		var tv hcl.Traversal
		switch expr := expr.(type) {
		case *hclsyntax.ScopeTraversalExpr:
			tv = expr.AsTraversal()
		case *hclsyntax.RelativeTraversalExpr:
			tv = expr.AsTraversal()
		default:
			response.Error = function.NewFuncError(fmt.Sprintf("unexpected non-traversal expression: %s", content))
			return
		}
		traversals = append(traversals, tv)
	}

	// Fix traversals
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
					var resp UpgradeReferenceConfigResponse
					req := UpgradeReferenceConfigRequest{
						Traversals: traversals,
					}
					u.ReferenceConfigUpgrader(ctx, req, &resp)
					if resp.Error != nil {
						response.Error = function.NewFuncError(resp.Error.Error())
						return
					}
					// Update traversals
					traversals = resp.Traversals
				}
			}

		}
	}

	var updateContents []string
	for _, tv := range traversals {
		updateContents = append(updateContents, FormatTraversal(tv))
	}
	response.Error = function.ConcatFuncErrors(response.Result.Set(ctx, updateContents))
	return
}
