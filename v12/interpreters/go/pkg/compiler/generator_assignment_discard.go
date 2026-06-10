package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileDiscardAssignment(ctx *compileContext, assign *ast.AssignmentExpression) ([]string, string, string, bool) {
	if assign == nil || assign.Operator != ast.AssignmentAssign {
		ctx.setReason("discard assignment cannot declare")
		return nil, "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", assign.Right)
	if !ok {
		return nil, "", "", false
	}
	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	return lines, valueTemp, valueType, true
}

func isDiscardAssignmentPattern(pattern ast.Pattern) bool {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return p != nil
	case *ast.Identifier:
		return p != nil && p.Name == "_"
	default:
		return false
	}
}
