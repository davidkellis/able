package interpreter

import "able/interpreter-go/pkg/ast"

func canUseSelfCallSlot(def *ast.FunctionDefinition) bool {
	if def == nil || def.ID == nil || def.ID.Name == "" || def.Body == nil {
		return false
	}
	return !blockMutatesIdentifier(def.Body, def.ID.Name)
}

func blockMutatesIdentifier(block *ast.BlockExpression, name string) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		if statementMutatesIdentifier(stmt, name) {
			return true
		}
	}
	return false
}

func statementMutatesIdentifier(stmt ast.Statement, name string) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.ForLoop:
		if ident, ok := s.Pattern.(*ast.Identifier); ok && ident != nil && ident.Name == name {
			return true
		}
		return expressionMutatesIdentifier(s.Iterable, name) || blockMutatesIdentifier(s.Body, name)
	case *ast.WhileLoop:
		return expressionMutatesIdentifier(s.Condition, name) || blockMutatesIdentifier(s.Body, name)
	case *ast.ReturnStatement:
		return expressionMutatesIdentifier(s.Argument, name)
	case *ast.YieldStatement:
		return expressionMutatesIdentifier(s.Expression, name)
	case *ast.BreakStatement:
		return expressionMutatesIdentifier(s.Value, name)
	case ast.Expression:
		return expressionMutatesIdentifier(s, name)
	default:
		return false
	}
}

func expressionMutatesIdentifier(expr ast.Expression, name string) bool {
	if expr == nil {
		return false
	}
	switch n := expr.(type) {
	case *ast.AssignmentExpression:
		if assignmentTargetMutatesIdentifier(n.Left, name) {
			return true
		}
		return expressionMutatesIdentifier(n.Right, name)
	case *ast.BinaryExpression:
		return expressionMutatesIdentifier(n.Left, name) || expressionMutatesIdentifier(n.Right, name)
	case *ast.UnaryExpression:
		return expressionMutatesIdentifier(n.Operand, name)
	case *ast.FunctionCall:
		if expressionMutatesIdentifier(n.Callee, name) {
			return true
		}
		for _, arg := range n.Arguments {
			if expressionMutatesIdentifier(arg, name) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		return expressionMutatesIdentifier(n.Object, name)
	case *ast.IndexExpression:
		return expressionMutatesIdentifier(n.Object, name) || expressionMutatesIdentifier(n.Index, name)
	case *ast.BlockExpression:
		return blockMutatesIdentifier(n, name)
	case *ast.IfExpression:
		if expressionMutatesIdentifier(n.IfCondition, name) || blockMutatesIdentifier(n.IfBody, name) {
			return true
		}
		for _, clause := range n.ElseIfClauses {
			if clause != nil && (expressionMutatesIdentifier(clause.Condition, name) || blockMutatesIdentifier(clause.Body, name)) {
				return true
			}
		}
		return blockMutatesIdentifier(n.ElseBody, name)
	case *ast.LoopExpression:
		return blockMutatesIdentifier(n.Body, name)
	case *ast.ArrayLiteral:
		for _, element := range n.Elements {
			if expressionMutatesIdentifier(element, name) {
				return true
			}
		}
		return false
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if expressionMutatesIdentifier(part, name) {
				return true
			}
		}
		return false
	case *ast.TypeCastExpression:
		return expressionMutatesIdentifier(n.Expression, name)
	case *ast.RangeExpression:
		return expressionMutatesIdentifier(n.Start, name) || expressionMutatesIdentifier(n.End, name)
	case *ast.PropagationExpression:
		return expressionMutatesIdentifier(n.Expression, name)
	case *ast.AwaitExpression:
		return expressionMutatesIdentifier(n.Expression, name)
	case *ast.StructLiteral:
		for _, src := range n.FunctionalUpdateSources {
			if expressionMutatesIdentifier(src, name) {
				return true
			}
		}
		for _, field := range n.Fields {
			if field != nil && expressionMutatesIdentifier(field.Value, name) {
				return true
			}
		}
		return false
	case *ast.MapLiteral:
		for _, element := range n.Elements {
			switch e := element.(type) {
			case *ast.MapLiteralEntry:
				if expressionMutatesIdentifier(e.Key, name) || expressionMutatesIdentifier(e.Value, name) {
					return true
				}
			case *ast.MapLiteralSpread:
				if expressionMutatesIdentifier(e.Expression, name) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

func assignmentTargetMutatesIdentifier(target ast.AssignmentTarget, name string) bool {
	if target == nil {
		return false
	}
	if ident, ok := target.(*ast.Identifier); ok && ident != nil {
		return ident.Name == name
	}
	if pattern, ok := target.(ast.Pattern); ok {
		return patternContainsIdentifierName(pattern, name)
	}
	return false
}

func patternContainsIdentifierName(pattern ast.Pattern, name string) bool {
	if pattern == nil {
		return false
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		return p.Name == name
	case *ast.TypedPattern:
		return patternContainsIdentifierName(p.Pattern, name)
	case *ast.ArrayPattern:
		for _, element := range p.Elements {
			if patternContainsIdentifierName(element, name) {
				return true
			}
		}
		return patternContainsIdentifierName(p.RestPattern, name)
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name == name {
				return true
			}
			if patternContainsIdentifierName(field.Pattern, name) {
				return true
			}
		}
	}
	return false
}
