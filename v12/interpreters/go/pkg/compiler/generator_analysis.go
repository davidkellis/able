package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) detectAstNeeds() {
	if g == nil {
		return
	}
	for _, info := range g.functions {
		if info == nil || !info.Compileable || info.Definition == nil || info.Definition.Body == nil {
			continue
		}
		if functionUsesMapLiteral(info.Definition) {
			g.needsAst = true
			return
		}
	}
}

func functionUsesMapLiteral(def *ast.FunctionDefinition) bool {
	if def == nil || def.Body == nil {
		return false
	}
	for _, stmt := range def.Body.Body {
		if statementUsesMapLiteral(stmt) {
			return true
		}
	}
	return false
}

func statementUsesMapLiteral(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	if ret, ok := stmt.(*ast.ReturnStatement); ok {
		if ret.Argument == nil {
			return false
		}
		return exprUsesMapLiteral(ret.Argument)
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return exprUsesMapLiteral(expr)
	}
	return false
}

func exprUsesMapLiteral(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.MapLiteral:
		return true
	case *ast.ArrayLiteral:
		for _, element := range e.Elements {
			if exprUsesMapLiteral(element) {
				return true
			}
		}
		return false
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && field.Value != nil && exprUsesMapLiteral(field.Value) {
				return true
			}
		}
		for _, source := range e.FunctionalUpdateSources {
			if exprUsesMapLiteral(source) {
				return true
			}
		}
		return false
	case *ast.IndexExpression:
		return exprUsesMapLiteral(e.Object) || exprUsesMapLiteral(e.Index)
	case *ast.MemberAccessExpression:
		if exprUsesMapLiteral(e.Object) {
			return true
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return exprUsesMapLiteral(memberExpr)
		}
		return false
	case *ast.UnaryExpression:
		return exprUsesMapLiteral(e.Operand)
	case *ast.BinaryExpression:
		return exprUsesMapLiteral(e.Left) || exprUsesMapLiteral(e.Right)
	case *ast.FunctionCall:
		if exprUsesMapLiteral(e.Callee) {
			return true
		}
		for _, arg := range e.Arguments {
			if exprUsesMapLiteral(arg) {
				return true
			}
		}
		return false
	case *ast.IfExpression:
		if exprUsesMapLiteral(e.IfCondition) || exprUsesMapLiteral(e.IfBody) {
			return true
		}
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if exprUsesMapLiteral(clause.Condition) || exprUsesMapLiteral(clause.Body) {
				return true
			}
		}
		if e.ElseBody != nil && exprUsesMapLiteral(e.ElseBody) {
			return true
		}
		return false
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if statementUsesMapLiteral(stmt) {
				return true
			}
		}
		return false
	case *ast.AssignmentExpression:
		if exprUsesMapLiteral(e.Right) {
			return true
		}
		switch target := e.Left.(type) {
		case *ast.IndexExpression:
			return exprUsesMapLiteral(target.Object) || exprUsesMapLiteral(target.Index)
		case *ast.MemberAccessExpression:
			return exprUsesMapLiteral(target.Object)
		case *ast.TypedPattern:
			if ident, ok := target.Pattern.(ast.Expression); ok {
				return exprUsesMapLiteral(ident)
			}
			return false
		default:
			if exprTarget, ok := target.(ast.Expression); ok {
				return exprUsesMapLiteral(exprTarget)
			}
			return false
		}
	default:
		return false
	}
}
