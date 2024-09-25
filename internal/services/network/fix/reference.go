package fix

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/magodo/terrafix-sdk/tfxsdk"
	"github.com/zclconf/go-cty/cty"
)

func ReferenceVirtualNetwork(_ int, traversals []hcl.Traversal) ([]hcl.Traversal, error) {
	var updates []hcl.Traversal
	for _, tv := range traversals {
		if len(tv) > 2 {
			idx, err := tfxsdk.FindSubAddr(tv,
				fmt.Sprintf("%s.location",
					tfxsdk.FormatTraversal(tv[:2]),
				),
			)
			if err != nil {
				return nil, err
			}
			if idx == -1 {
				updates = append(updates, tv)
			} else {
				naddr := append(hcl.Traversal{}, tv[:idx]...)
				naddr = append(naddr,
					hcl.TraverseAttr{Name: "locations"},
					hcl.TraverseIndex{Key: cty.NumberIntVal(0)},
				)
				naddr = append(naddr, tv[idx+1:]...)
				updates = append(updates, naddr)
			}
		} else {
			updates = append(updates, tv)
		}
	}
	return updates, nil
}
