package interpreter

import "able/interpreter-go/pkg/ast"

func blockLocalBindingCapacity(block *ast.BlockExpression) int {
	if block == nil {
		return 0
	}
	total := 0
	for _, stmt := range block.Body {
		total += statementLocalBindingCapacity(stmt)
	}
	return total
}

func statementLocalBindingCapacity(stmt ast.Statement) int {
	switch n := stmt.(type) {
	case *ast.AssignmentExpression:
		if n.Operator != ast.AssignmentDeclare {
			return 0
		}
		return assignmentTargetBindingCapacity(n.Left)
	case *ast.FunctionDefinition:
		return namedBindingCapacity(n.ID)
	case *ast.StructDefinition:
		return namedBindingCapacity(n.ID)
	case *ast.UnionDefinition:
		return namedBindingCapacity(n.ID)
	case *ast.TypeAliasDefinition:
		return namedBindingCapacity(n.ID)
	case *ast.InterfaceDefinition:
		return namedBindingCapacity(n.ID)
	default:
		return 0
	}
}

func assignmentTargetBindingCapacity(target ast.AssignmentTarget) int {
	switch t := target.(type) {
	case *ast.Identifier:
		return namedBindingCapacity(t)
	case *ast.TypedPattern:
		return patternBindingCapacity(t.Pattern)
	case ast.Pattern:
		return patternBindingCapacity(t)
	default:
		return 0
	}
}

func patternBindingCapacity(pattern ast.Pattern) int {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return namedBindingCapacity(p)
	case *ast.WildcardPattern, *ast.LiteralPattern:
		return 0
	case *ast.TypedPattern:
		return patternBindingCapacity(p.Pattern)
	case *ast.ArrayPattern:
		total := 0
		for _, el := range p.Elements {
			total += patternBindingCapacity(el)
		}
		if rest, ok := p.RestPattern.(ast.Pattern); ok {
			total += patternBindingCapacity(rest)
		}
		return total
	case *ast.StructPattern:
		total := 0
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil {
				total += namedBindingCapacity(field.Binding)
				continue
			}
			total += patternBindingCapacity(field.Pattern)
		}
		return total
	default:
		return 0
	}
}

func functionLocalBindingCapacity(decl *ast.FunctionDefinition, call *ast.FunctionCall) int {
	if decl == nil {
		return 0
	}
	total := blockLocalBindingCapacity(decl.Body)
	for _, param := range decl.Params {
		if param == nil {
			continue
		}
		total += patternBindingCapacity(param.Name)
	}
	if call != nil && len(decl.GenericParams) > 0 {
		total += len(decl.GenericParams) * 2
	}
	return total
}

func lambdaLocalBindingCapacity(expr *ast.LambdaExpression, call *ast.FunctionCall) int {
	if expr == nil {
		return 0
	}
	total := 0
	if body, ok := expr.Body.(*ast.BlockExpression); ok {
		total += blockLocalBindingCapacity(body)
	}
	for _, param := range expr.Params {
		if param == nil {
			continue
		}
		total += patternBindingCapacity(param.Name)
	}
	if call != nil && len(expr.GenericParams) > 0 {
		total += len(expr.GenericParams) * 2
	}
	return total
}

func namedBindingCapacity(id *ast.Identifier) int {
	if id == nil || id.Name == "" || id.Name == "_" {
		return 0
	}
	return 1
}
