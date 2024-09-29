package fix

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/magodo/terrafix-sdk/tfxsdk"
	"github.com/zclconf/go-cty/cty"
)

func ReferenceVirtualNetwork(_ int, traversals []hcl.Traversal) ([]hcl.Traversal, error) {
	var updates []hcl.Traversal
	for _, tv := range traversals {
		// The traversal's length must be larger than 2 for resource references
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
			return nil, err
		}
		updates = append(updates, tv)
	}
	return updates, nil
}
