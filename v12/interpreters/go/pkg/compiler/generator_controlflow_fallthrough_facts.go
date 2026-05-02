package compiler

import (
	"math"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) seedFactsAfterTerminatingIf(ctx *compileContext, expr *ast.IfExpression) {
	if g == nil || ctx == nil || expr == nil || expr.IfCondition == nil || expr.IfBody == nil {
		return
	}
	if len(expr.ElseIfClauses) != 0 || expr.ElseBody != nil || !blockAlwaysExits(expr.IfBody) {
		return
	}
	g.seedFactsFromFalseCondition(ctx, expr.IfCondition)
}

func blockAlwaysExits(block *ast.BlockExpression) bool {
	if block == nil || len(block.Body) == 0 {
		return false
	}
	return statementAlwaysExits(block.Body[len(block.Body)-1])
}

func statementAlwaysExits(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement, *ast.RaiseStatement, *ast.RethrowStatement:
		return true
	case *ast.BlockExpression:
		return blockAlwaysExits(s)
	case *ast.IfExpression:
		if s == nil || s.IfBody == nil || s.ElseBody == nil {
			return false
		}
		if !blockAlwaysExits(s.IfBody) || !blockAlwaysExits(s.ElseBody) {
			return false
		}
		for _, clause := range s.ElseIfClauses {
			if clause == nil || clause.Body == nil || !blockAlwaysExits(clause.Body) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (g *generator) seedFactsFromFalseCondition(ctx *compileContext, cond ast.Expression) {
	binary, ok := cond.(*ast.BinaryExpression)
	if !ok || binary == nil {
		return
	}
	name, lowerBound, ok := falseConditionIntegerLowerBound(binary)
	if !ok || lowerBound < 0 {
		return
	}
	binding, ok := ctx.lookup(name)
	if !ok || binding.GoName == "" || !g.isIntegerType(binding.GoType) {
		return
	}
	fact, _ := ctx.integerFactForGoName(binding.GoName)
	fact.NonNegative = true
	if fact.hasUsefulFact() {
		ctx.setIntegerFact(binding.GoName, fact)
	}
}

func falseConditionIntegerLowerBound(binary *ast.BinaryExpression) (string, int64, bool) {
	if binary == nil {
		return "", 0, false
	}
	if ident, ok := binary.Left.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
		if value, ok := integerLiteralInt64(binary.Right); ok {
			switch binary.Operator {
			case "<":
				return ident.Name, value, true
			case "<=":
				if value == math.MaxInt64 {
					return "", 0, false
				}
				return ident.Name, value + 1, true
			}
		}
	}
	if ident, ok := binary.Right.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
		if value, ok := integerLiteralInt64(binary.Left); ok {
			switch binary.Operator {
			case ">":
				return ident.Name, value, true
			case ">=":
				if value == math.MaxInt64 {
					return "", 0, false
				}
				return ident.Name, value + 1, true
			}
		}
	}
	return "", 0, false
}

func integerLiteralInt64(expr ast.Expression) (int64, bool) {
	lit, ok := expr.(*ast.IntegerLiteral)
	if !ok || lit == nil || lit.Value == nil || !lit.Value.IsInt64() {
		return 0, false
	}
	return lit.Value.Int64(), true
}
