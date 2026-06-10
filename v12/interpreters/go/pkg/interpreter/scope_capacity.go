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

func expressionNeedsCurrentScopeBinding(expr ast.Expression) bool {
	return expressionCurrentScopeBindingCapacity(expr) > 0
}

func expressionCurrentScopeBindingCapacity(expr ast.Expression) int {
	if expr == nil {
		return 0
	}
	switch n := expr.(type) {
	case *ast.AssignmentExpression:
		total := 0
		if n.Operator == ast.AssignmentDeclare {
			total += assignmentTargetBindingCapacity(n.Left)
		}
		if targetExpr, ok := n.Left.(ast.Expression); ok {
			total += expressionCurrentScopeBindingCapacity(targetExpr)
		}
		total += expressionCurrentScopeBindingCapacity(n.Right)
		return total
	case *ast.BinaryExpression:
		return expressionCurrentScopeBindingCapacity(n.Left) + expressionCurrentScopeBindingCapacity(n.Right)
	case *ast.UnaryExpression:
		return expressionCurrentScopeBindingCapacity(n.Operand)
	case *ast.FunctionCall:
		total := expressionCurrentScopeBindingCapacity(n.Callee)
		for _, arg := range n.Arguments {
			total += expressionCurrentScopeBindingCapacity(arg)
		}
		return total
	case *ast.MemberAccessExpression:
		total := expressionCurrentScopeBindingCapacity(n.Object)
		if memberExpr, ok := n.Member.(ast.Expression); ok {
			total += expressionCurrentScopeBindingCapacity(memberExpr)
		}
		return total
	case *ast.IndexExpression:
		return expressionCurrentScopeBindingCapacity(n.Object) + expressionCurrentScopeBindingCapacity(n.Index)
	case *ast.ArrayLiteral:
		total := 0
		for _, el := range n.Elements {
			total += expressionCurrentScopeBindingCapacity(el)
		}
		return total
	case *ast.MapLiteral:
		total := 0
		for _, el := range n.Elements {
			switch item := el.(type) {
			case *ast.MapLiteralEntry:
				if item != nil {
					total += expressionCurrentScopeBindingCapacity(item.Key)
					total += expressionCurrentScopeBindingCapacity(item.Value)
				}
			case *ast.MapLiteralSpread:
				if item != nil {
					total += expressionCurrentScopeBindingCapacity(item.Expression)
				}
			}
		}
		return total
	case *ast.StructLiteral:
		total := 0
		for _, field := range n.Fields {
			if field != nil {
				total += expressionCurrentScopeBindingCapacity(field.Value)
			}
		}
		for _, src := range n.FunctionalUpdateSources {
			total += expressionCurrentScopeBindingCapacity(src)
		}
		return total
	case *ast.StringInterpolation:
		total := 0
		for _, part := range n.Parts {
			total += expressionCurrentScopeBindingCapacity(part)
		}
		return total
	case *ast.TypeCastExpression:
		return expressionCurrentScopeBindingCapacity(n.Expression)
	case *ast.RangeExpression:
		return expressionCurrentScopeBindingCapacity(n.Start) + expressionCurrentScopeBindingCapacity(n.End)
	case *ast.PropagationExpression:
		return expressionCurrentScopeBindingCapacity(n.Expression)
	case *ast.AwaitExpression:
		return expressionCurrentScopeBindingCapacity(n.Expression)
	case *ast.OrElseExpression:
		return expressionCurrentScopeBindingCapacity(n.Expression)
	case *ast.EnsureExpression:
		return expressionCurrentScopeBindingCapacity(n.TryExpression)
	case *ast.IfExpression:
		total := expressionCurrentScopeBindingCapacity(n.IfCondition)
		for _, clause := range n.ElseIfClauses {
			if clause != nil {
				total += expressionCurrentScopeBindingCapacity(clause.Condition)
			}
		}
		return total
	case *ast.MatchExpression:
		return expressionCurrentScopeBindingCapacity(n.Subject)
	case *ast.RescueExpression:
		return expressionCurrentScopeBindingCapacity(n.MonitoredExpression)
	case *ast.BlockExpression,
		*ast.LoopExpression,
		*ast.IteratorLiteral,
		*ast.SpawnExpression,
		*ast.LambdaExpression,
		*ast.PlaceholderExpression,
		*ast.ImplicitMemberExpression,
		*ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return 0
	default:
		return 0
	}
}

func clauseNeedsLocalScope(pattern ast.Pattern, guard ast.Expression, body ast.Expression) bool {
	return clauseLocalScopePlan(pattern, guard, body).needsLocalScope
}

func clauseLocalScopePlan(pattern ast.Pattern, guard ast.Expression, body ast.Expression) clauseScopePlan {
	captureBindings := clauseReferencesPatternBindings(pattern, guard, body)
	patternCapacity := 0
	if captureBindings {
		patternCapacity = patternBindingCapacity(pattern)
	}
	currentScopeCapacity := expressionCurrentScopeBindingCapacity(guard) + expressionCurrentScopeBindingCapacity(body)
	needsLocalScope := captureBindings || currentScopeCapacity > 0
	transientScopeEnvOK := needsLocalScope && clauseAllowsTransientBindingEnv(guard, body)
	transientBindingSetOK := captureBindings && transientScopeEnvOK
	return clauseScopePlan{
		needsLocalScope:       needsLocalScope,
		capturePatternBinding: captureBindings,
		patternBindingCount:   patternCapacity,
		localBindingCapacity:  patternCapacity + currentScopeCapacity,
		transientScopeEnvOK:   transientScopeEnvOK,
		transientBindingSetOK: transientBindingSetOK,
		transientSingleBindOK: transientBindingSetOK &&
			simplePatternBoundIdentifierName(pattern) != "",
	}
}

func clauseAllowsTransientBindingEnv(guard ast.Expression, body ast.Expression) bool {
	return expressionAllowsTransientRuntimeScope(guard) &&
		expressionAllowsTransientRuntimeScope(body)
}

func clauseReferencesPatternBindings(pattern ast.Pattern, guard ast.Expression, body ast.Expression) bool {
	names := patternBoundIdentifierNameSet(pattern)
	if len(names) == 0 {
		return false
	}
	return bytecodeExpressionReferencesIdentifierSet(guard, names) ||
		bytecodeExpressionReferencesIdentifierSet(body, names)
}

func patternBoundIdentifierNameSet(pattern ast.Pattern) map[string]struct{} {
	count := patternBindingCapacity(pattern)
	if count == 0 {
		return nil
	}
	names := make(map[string]struct{}, count)
	collectPatternBoundIdentifierNames(pattern, names)
	if len(names) == 0 {
		return nil
	}
	return names
}

func collectPatternBoundIdentifierNames(pattern ast.Pattern, names map[string]struct{}) {
	if pattern == nil || names == nil {
		return
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p != nil && p.Name != "" && p.Name != "_" {
			names[p.Name] = struct{}{}
		}
	case *ast.TypedPattern:
		if p != nil {
			collectPatternBoundIdentifierNames(p.Pattern, names)
		}
	case *ast.ArrayPattern:
		if p == nil {
			return
		}
		for _, element := range p.Elements {
			collectPatternBoundIdentifierNames(element, names)
		}
		collectPatternBoundIdentifierNames(p.RestPattern, names)
	case *ast.StructPattern:
		if p == nil {
			return
		}
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				names[field.Binding.Name] = struct{}{}
			}
			collectPatternBoundIdentifierNames(field.Pattern, names)
		}
	}
}

func simplePatternBoundIdentifierName(pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p != nil && p.Name != "" && p.Name != "_" {
			return p.Name
		}
	case *ast.TypedPattern:
		if p != nil {
			return simplePatternBoundIdentifierName(p.Pattern)
		}
	}
	return ""
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
	return functionLocalBindingCapacityForLayout(decl, call, nil)
}

func functionLocalBindingCapacityForLayout(decl *ast.FunctionDefinition, call *ast.FunctionCall, layout *bytecodeFrameLayout) int {
	if decl == nil {
		return 0
	}
	total := blockLocalBindingCapacity(decl.Body)
	if layout == nil {
		for _, param := range decl.Params {
			if param == nil {
				continue
			}
			total += patternBindingCapacity(param.Name)
		}
	}
	if call != nil && len(decl.GenericParams) > 0 {
		total += len(decl.GenericParams) * 2
	}
	return total
}

func lambdaLocalBindingCapacity(expr *ast.LambdaExpression, call *ast.FunctionCall) int {
	return lambdaLocalBindingCapacityForLayout(expr, call, nil)
}

func lambdaLocalBindingCapacityForLayout(expr *ast.LambdaExpression, call *ast.FunctionCall, layout *bytecodeFrameLayout) int {
	if expr == nil {
		return 0
	}
	total := 0
	if body, ok := expr.Body.(*ast.BlockExpression); ok {
		total += blockLocalBindingCapacity(body)
	}
	if layout == nil {
		for _, param := range expr.Params {
			if param == nil {
				continue
			}
			total += patternBindingCapacity(param.Name)
		}
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
