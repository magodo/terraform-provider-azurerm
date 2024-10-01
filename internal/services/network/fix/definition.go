package fix

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/network/migration"
	"github.com/magodo/terrafix-sdk/tfxsdk"
	"github.com/zclconf/go-cty/cty"
)

var DefinitionVirtualNetwork = tfxsdk.DefinitionConfigUpgraders{
	0: {
		DefinitionConfigUpgrader: func(ctx context.Context, req tfxsdk.UpgradeDefinitionConfigRequest, resp *tfxsdk.UpgradeDefinitionConfigResponse) {
			sbody, wbody, state := req.SyntaxBody, req.WriteBody, req.State
			// Location: Changed from location -> locations
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
								resp.Error = fmt.Errorf(`traversal matching for "locations.0": %v`, err)
								return
							}
							if ok {
								tks = hclwrite.TokensForTraversal(append(tv[:2], hcl.TraverseAttr{Name: "locations"}))
							}
						}
					}
				}

				wbody.SetAttributeRaw("locations", tks)
				wbody.RemoveAttribute("location")

				// Upgrade state
				if state != nil {
					state, err := migration.VirtualNetworkV0ToV1{}.UpgradeFunc()(ctx, state, nil)
					if err != nil {
						resp.Error = fmt.Errorf("migrate state: %v", err.Error())
					}
					resp.State = state
				}
			}

		},
	},
	1: {
		DefinitionConfigUpgrader: func(ctx context.Context, req tfxsdk.UpgradeDefinitionConfigRequest, resp *tfxsdk.UpgradeDefinitionConfigResponse) {
			wbody, state := req.WriteBody, req.State
			// UUID: Changed from computed to required
			var uuidVal string
			if state != nil {
				if v, ok := state["uuid"]; ok {
					uuidVal = v.(string)
				}
			}
			if uuidVal == "" {
				uuidVal = "TERRAFIX TODO: Find out the uuid from state"
			}
			wbody.SetAttributeValue("uuid", cty.StringVal(uuidVal))

			// Upgrade state
			if state != nil {
				state, err := migration.VirtualNetworkV1ToV2{}.UpgradeFunc()(ctx, state, nil)
				if err != nil {
					resp.Error = fmt.Errorf("migrate state: %v", err.Error())
				}
				resp.State = state
			}
		},
	},
}
