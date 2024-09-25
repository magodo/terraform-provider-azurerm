package tfxsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ReferenceFixFunction func(version int, traversals []hcl.Traversal) ([]hcl.Traversal, error)

type ReferenceFixers map[BlockType]map[string]ReferenceFixFunction

type FixConfigReferenceFunction struct {
	Fixers ReferenceFixers
}

var _ function.Function = FixConfigReferenceFunction{}

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

	if m, ok := a.Fixers[BlockType(blockType)]; ok {
		if u, ok := m[blockName]; ok {
			var err error
			traversals, err = u(int(version), traversals)
			if err != nil {
				response.Error = function.NewFuncError(err.Error())
				return
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
