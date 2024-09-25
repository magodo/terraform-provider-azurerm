package tfxsdk

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// TraversalMatches tells whether the traversal "t1" matches the sub-traversal t2.
//
// Any index/splat steps in "addr" will be ignored, except the current step in "subaddr" being
// tested is also a index/slat step.
//
// current step in "t2" being tested is also a index/slat step.
// E.g. "a[0].b.c" matches "a.b.c", but not "a.b"
func TraversalMatches(t1 hcl.Traversal, t2str string) (bool, error) {
	t2, err := ParseTraversal(t2str)
	if err != nil {
		return false, fmt.Errorf("parsing traversal %s: %v", t2str, err)
	}

	i1 := 0
	i2 := 0
	for ; i1 != len(t1) && i2 != len(t2); i1++ {
		n1, isAttr1 := FormatTraverse(t1[i1])
		n2, isAttr2 := FormatTraverse(t2[i2])

		// Skip indx/splat in t1 if the current focused t2 is an attr
		if !isAttr1 && isAttr2 {
			continue
		}

		if n1 != n2 {
			return false, nil
		}
		i2 += 1
	}
	if !(len(t1) == i1 && len(t2) == i2) {
		return false, nil
	}
	return true, nil
}

// FindSubAddr finds the last step's traversal index in "t1", for the sub-address "t2".
//
// Any index/splat steps in "t1" will be ignored, except the current step in "t2" being
// tested is also a index/slat step.
//
// If "t2" is not found in the "t1", -1 is returned.
// E.g. Given "a[0].b.c", and as "a.b", 2 is returned.
func FindSubAddr(t1 hcl.Traversal, t2str string) (int, error) {
	addr2, err := ParseTraversal(t2str)
	if err != nil {
		return 0, fmt.Errorf("parsing traversal %s: %v", t2str, err)
	}

	i2 := 0
	var idx int
	for i1 := range t1 {
		n1, isAttr1 := FormatTraverse(t1[i1])
		n2, isAttr2 := FormatTraverse(addr2[i2])

		// Skip indx/splat in addr if the current focused subaddr is an attr
		if !isAttr1 && isAttr2 {
			continue
		}

		if n1 != n2 {
			return -1, nil
		}
		i2 += 1
		if len(addr2) == i2 {
			idx = i1
			break
		}
	}
	if len(addr2) != i2 {
		return -1, nil
	}
	return idx, nil
}

func FormatTraversal(ts hcl.Traversal) string {
	var out string
	if len(ts) == 0 {
		return ""
	}
	for i, t := range ts {
		v, isAttr := FormatTraverse(t)
		if i == 0 {
			out += v
			continue
		}
		if isAttr {
			out += "." + v
		} else {
			out += "[" + v + "]"
		}
	}
	return out
}

func FormatTraverse(t hcl.Traverser) (str string, isAttr bool) {
	switch t := t.(type) {
	case hcl.TraverseRoot:
		return t.Name, true
	case hcl.TraverseAttr:
		return t.Name, true
	case hcl.TraverseIndex:
		return IndexKeyString(t.Key), false
	case hcl.TraverseSplat:
		return "*", false
	default:
		panic("unreachable")
	}
}

func IndexKeyString(key cty.Value) string {
	switch key.Type() {
	case cty.Number:
		f := key.AsBigFloat()
		idx, _ := f.Int64()
		return fmt.Sprintf("%d", idx)
	case cty.String:
		return key.AsString()
	default:
		panic(fmt.Sprintf("unsupported index key type %v", key.Type()))
	}
}

func ParseTraversal(t string) (hcl.Traversal, error) {
	exp, diags := hclsyntax.ParseExpression([]byte(t), "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}
	var tv hcl.Traversal
	switch exp := exp.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		tv = exp.AsTraversal()
	case *hclsyntax.RelativeTraversalExpr:
		tv = exp.AsTraversal()
	default:
		return nil, fmt.Errorf("invalid type %T", exp)
	}
	return tv, nil
}
