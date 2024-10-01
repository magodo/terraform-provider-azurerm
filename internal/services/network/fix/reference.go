package fix

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/magodo/terrafix-sdk/tfxsdk"
	"github.com/zclconf/go-cty/cty"
)

var ReferenceVirtualNetwork = tfxsdk.ReferenceConfigUpgraders{
	0: {
		ReferenceConfigUpgrader: func(ctx context.Context, req tfxsdk.UpgradeReferenceConfigRequest, resp *tfxsdk.UpgradeReferenceConfigResponse) {
			tvs := req.Traversals
			var utvs []hcl.Traversal
			for _, tv := range tvs {
				// The traversal's length must be larger than 2 for resource references
				// Location -> Locations
				tv, err := tfxsdk.TraversalReplace(
					tv,
					append(
						append(hcl.Traversal{}, tv[:2]...),
						hcl.TraverseAttr{Name: "location"},
					),
					append(
						hcl.Traversal{},
						hcl.TraverseAttr{Name: "locations"},
						hcl.TraverseIndex{Key: cty.NumberIntVal(0)},
					),
				)
				if err != nil {
					resp.Error = err
					return
				}

				// guid -> uuid
				tv, err = tfxsdk.TraversalReplace(
					tv,
					append(
						append(hcl.Traversal{}, tv[:2]...),
						hcl.TraverseAttr{Name: "guid"},
					),
					append(
						hcl.Traversal{},
						hcl.TraverseAttr{Name: "uuid"},
					),
				)
				if err != nil {
					resp.Error = err
					return
				}

				utvs = append(utvs, tv)
			}
			resp.Traversals = utvs
			return
		},
	},
}
