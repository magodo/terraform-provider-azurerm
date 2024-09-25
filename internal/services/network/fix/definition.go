package fix

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/magodo/terrafix-sdk/tfxsdk"
)

func DefinitionVirtualNetwork(_ int, sbody *hclsyntax.Body, wbody *hclwrite.Body) error {
	if vloc, ok := wbody.Attributes()["location"]; ok {
		tks := hclwrite.TokensForTuple([]hclwrite.Tokens{vloc.Expr().BuildTokens(nil)})

		// Restore the reference modification if this is a same-origin reference
		locAttr := sbody.Attributes["location"]
		var tv hcl.Traversal
		switch expr := locAttr.Expr.(type) {
		case *hclsyntax.ScopeTraversalExpr:
			tv = expr.AsTraversal()
		case *hclsyntax.RelativeTraversalExpr:
			tv = expr.AsTraversal()
		default:
		}
		if tv != nil {
			if len(tv) > 2 {
				if root, ok := tv[0].(hcl.TraverseRoot); ok && root.Name == "azurerm_virtual_network" {
					ok, err := tfxsdk.TraversalMatches(tv[2:], "locations.0")
					if err != nil {
						return fmt.Errorf(`traversal matching for "locations.0": %v`, err)
					}
					if ok {
						tks = hclwrite.TokensForTraversal(append(tv[:2], hcl.TraverseAttr{Name: "locations"}))
					}
				}
			}
		}

		wbody.SetAttributeRaw("locations", tks)
		wbody.RemoveAttribute("location")
	}
	return nil
}
