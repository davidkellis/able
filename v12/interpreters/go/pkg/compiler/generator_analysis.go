package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) detectAstNeeds() {
	if g == nil {
		return
	}
	for _, info := range g.allFunctionInfos() {
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
	case *ast.TypeCastExpression:
		return true
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
	case *ast.MatchExpression:
		if exprUsesMapLiteral(e.Subject) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && exprUsesMapLiteral(clause.Guard) {
				return true
			}
			if clause.Body != nil && exprUsesMapLiteral(clause.Body) {
				return true
			}
		}
		return false
	case *ast.RescueExpression:
		if exprUsesMapLiteral(e.MonitoredExpression) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && exprUsesMapLiteral(clause.Guard) {
				return true
			}
			if clause.Body != nil && exprUsesMapLiteral(clause.Body) {
				return true
			}
		}
		return false
	case *ast.LambdaExpression:
		if e.Body == nil {
			return false
		}
		return exprUsesMapLiteral(e.Body)
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

func blockHasBreakContinueRescue(body *ast.BlockExpression) bool {
	if body == nil || body.Body == nil {
		return false
	}
	for _, stmt := range body.Body {
		if statementHasBreakContinueRescue(stmt) {
			return true
		}
	}
	return false
}

func statementHasBreakContinueRescue(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.BreakStatement, *ast.ContinueStatement:
		return true
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return exprHasBreakContinueRescue(s.Argument)
		}
		return false
	case *ast.WhileLoop:
		if exprHasBreakContinueRescue(s.Condition) {
			return true
		}
		return blockHasBreakContinueRescue(s.Body)
	case *ast.ForLoop:
		if exprHasBreakContinueRescue(s.Iterable) {
			return true
		}
		return blockHasBreakContinueRescue(s.Body)
	case *ast.AssignmentExpression:
		return exprHasBreakContinueRescue(s.Right)
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			return exprHasBreakContinueRescue(expr)
		}
		return false
	}
}

func exprHasBreakContinueRescue(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.RescueExpression:
		return true
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if statementHasBreakContinueRescue(stmt) {
				return true
			}
		}
		return false
	case *ast.IfExpression:
		if exprHasBreakContinueRescue(e.IfCondition) || exprHasBreakContinueRescue(e.IfBody) {
			return true
		}
		for _, clause := range e.ElseIfClauses {
			if clause != nil && (exprHasBreakContinueRescue(clause.Condition) || exprHasBreakContinueRescue(clause.Body)) {
				return true
			}
		}
		return e.ElseBody != nil && exprHasBreakContinueRescue(e.ElseBody)
	case *ast.MatchExpression:
		if exprHasBreakContinueRescue(e.Subject) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && exprHasBreakContinueRescue(clause.Guard) {
				return true
			}
			if clause.Body != nil && exprHasBreakContinueRescue(clause.Body) {
				return true
			}
		}
		return false
	case *ast.FunctionCall:
		if exprHasBreakContinueRescue(e.Callee) {
			return true
		}
		for _, arg := range e.Arguments {
			if exprHasBreakContinueRescue(arg) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		if exprHasBreakContinueRescue(e.Object) {
			return true
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return exprHasBreakContinueRescue(memberExpr)
		}
		return false
	case *ast.IndexExpression:
		return exprHasBreakContinueRescue(e.Object) || exprHasBreakContinueRescue(e.Index)
	case *ast.UnaryExpression:
		return exprHasBreakContinueRescue(e.Operand)
	case *ast.BinaryExpression:
		return exprHasBreakContinueRescue(e.Left) || exprHasBreakContinueRescue(e.Right)
	case *ast.LambdaExpression:
		if e.Body != nil {
			return exprHasBreakContinueRescue(e.Body)
		}
		return false
	case *ast.LoopExpression:
		if e.Body != nil {
			return exprHasBreakContinueRescue(e.Body)
		}
		return false
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && field.Value != nil && exprHasBreakContinueRescue(field.Value) {
				return true
			}
		}
		for _, source := range e.FunctionalUpdateSources {
			if exprHasBreakContinueRescue(source) {
				return true
			}
		}
		return false
	case *ast.ArrayLiteral:
		for _, elem := range e.Elements {
			if exprHasBreakContinueRescue(elem) {
				return true
			}
		}
		return false
	case *ast.MapLiteral:
		for _, elem := range e.Elements {
			switch entry := elem.(type) {
			case *ast.MapLiteralEntry:
				if exprHasBreakContinueRescue(entry.Key) || exprHasBreakContinueRescue(entry.Value) {
					return true
				}
			case *ast.MapLiteralSpread:
				if exprHasBreakContinueRescue(entry.Expression) {
					return true
				}
			}
		}
		return false
	case *ast.AssignmentExpression:
		return exprHasBreakContinueRescue(e.Right)
	case *ast.AwaitExpression:
		return exprHasBreakContinueRescue(e.Expression)
	default:
		return false
	}
}
