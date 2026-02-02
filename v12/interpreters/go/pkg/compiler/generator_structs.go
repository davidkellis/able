package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) fieldInfo(info *structInfo, name string) *fieldInfo {
	for idx := range info.Fields {
		if info.Fields[idx].Name == name {
			return &info.Fields[idx]
		}
	}
	return nil
}

func (g *generator) structFieldForMember(info *structInfo, member ast.Expression) (*fieldInfo, bool) {
	if member == nil {
		return nil, false
	}
	switch m := member.(type) {
	case *ast.Identifier:
		if m == nil || m.Name == "" {
			return nil, false
		}
		if info == nil {
			return nil, true
		}
		return g.fieldInfo(info, m.Name), true
	case *ast.IntegerLiteral:
		idx, ok := memberIndexFromLiteral(m)
		if !ok {
			return nil, false
		}
		if info == nil || info.Kind != ast.StructKindPositional {
			return nil, true
		}
		if idx < 0 || idx >= len(info.Fields) {
			return nil, true
		}
		return &info.Fields[idx], true
	default:
		return nil, false
	}
}

func memberIndexFromLiteral(lit *ast.IntegerLiteral) (int, bool) {
	if lit == nil || lit.Value == nil || !lit.Value.IsInt64() {
		return 0, false
	}
	idx := lit.Value.Int64()
	if idx < 0 {
		return 0, false
	}
	maxInt := int64(^uint(0) >> 1)
	if idx > maxInt {
		return 0, false
	}
	return int(idx), true
}

func safeParamName(name string, idx int) string {
	candidate := sanitizeIdent(name)
	if candidate == "" || candidate == "err" || candidate == "args" || candidate == "rt" || candidate == "ctx" {
		return fmt.Sprintf("p%d", idx)
	}
	return candidate
}
