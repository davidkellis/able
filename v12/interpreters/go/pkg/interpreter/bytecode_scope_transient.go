package interpreter

import "able/interpreter-go/pkg/ast"

func blockAllowsTransientRuntimeScope(block *ast.BlockExpression) bool {
	if block == nil {
		return true
	}
	for _, stmt := range block.Body {
		if !statementAllowsTransientRuntimeScope(stmt) {
			return false
		}
	}
	return true
}

func (i *Interpreter) blockAllowsTransientRuntimeScopeCached(block *ast.BlockExpression) bool {
	if block == nil {
		return true
	}
	if i == nil {
		return blockAllowsTransientRuntimeScope(block)
	}
	if i.envSingleThread {
		if allowed, ok := i.blockTransientRuntimeScopeCache[block]; ok {
			return allowed
		}
		allowed := blockAllowsTransientRuntimeScope(block)
		i.blockTransientRuntimeScopeCache[block] = allowed
		return allowed
	}
	i.blockTransientRuntimeScopeCacheMu.RLock()
	allowed, ok := i.blockTransientRuntimeScopeCache[block]
	i.blockTransientRuntimeScopeCacheMu.RUnlock()
	if ok {
		return allowed
	}
	allowed = blockAllowsTransientRuntimeScope(block)
	i.blockTransientRuntimeScopeCacheMu.Lock()
	if existing, ok := i.blockTransientRuntimeScopeCache[block]; ok {
		i.blockTransientRuntimeScopeCacheMu.Unlock()
		return existing
	}
	if i.blockTransientRuntimeScopeCache == nil {
		i.blockTransientRuntimeScopeCache = make(map[*ast.BlockExpression]bool)
	}
	i.blockTransientRuntimeScopeCache[block] = allowed
	i.blockTransientRuntimeScopeCacheMu.Unlock()
	return allowed
}

func statementAllowsTransientRuntimeScope(stmt ast.Statement) bool {
	if stmt == nil {
		return true
	}
	switch s := stmt.(type) {
	case *ast.FunctionDefinition, *ast.MethodsDefinition, *ast.ImplementationDefinition:
		return false
	case *ast.ForLoop:
		if s == nil {
			return true
		}
		return expressionAllowsTransientRuntimeScope(s.Iterable) &&
			blockAllowsTransientRuntimeScope(s.Body)
	case *ast.WhileLoop:
		if s == nil {
			return true
		}
		return expressionAllowsTransientRuntimeScope(s.Condition) &&
			blockAllowsTransientRuntimeScope(s.Body)
	case *ast.ReturnStatement:
		if s == nil {
			return true
		}
		return expressionAllowsTransientRuntimeScope(s.Argument)
	case *ast.RaiseStatement:
		if s == nil {
			return true
		}
		return expressionAllowsTransientRuntimeScope(s.Expression)
	case *ast.BreakStatement:
		if s == nil {
			return true
		}
		return expressionAllowsTransientRuntimeScope(s.Value)
	case *ast.YieldStatement:
		if s == nil {
			return true
		}
		return expressionAllowsTransientRuntimeScope(s.Expression)
	case *ast.ContinueStatement,
		*ast.RethrowStatement,
		*ast.StructDefinition,
		*ast.UnionDefinition,
		*ast.TypeAliasDefinition,
		*ast.InterfaceDefinition,
		*ast.ExternFunctionBody,
		*ast.ImportStatement,
		*ast.DynImportStatement,
		*ast.PackageStatement,
		*ast.PreludeStatement:
		return true
	case ast.Expression:
		return expressionAllowsTransientRuntimeScope(s)
	default:
		return false
	}
}

func expressionAllowsTransientRuntimeScope(expr ast.Expression) bool {
	if expr == nil {
		return true
	}
	switch n := expr.(type) {
	case *ast.LambdaExpression,
		*ast.PlaceholderExpression,
		*ast.IteratorLiteral,
		*ast.SpawnExpression:
		return false
	case *ast.StringLiteral,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral,
		*ast.Identifier,
		*ast.ImplicitMemberExpression:
		return true
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if !expressionAllowsTransientRuntimeScope(el) {
				return false
			}
		}
		return true
	case *ast.MapLiteral:
		for _, el := range n.Elements {
			switch item := el.(type) {
			case *ast.MapLiteralEntry:
				if item != nil &&
					(!expressionAllowsTransientRuntimeScope(item.Key) ||
						!expressionAllowsTransientRuntimeScope(item.Value)) {
					return false
				}
			case *ast.MapLiteralSpread:
				if item != nil && !expressionAllowsTransientRuntimeScope(item.Expression) {
					return false
				}
			default:
				return false
			}
		}
		return true
	case *ast.UnaryExpression:
		return expressionAllowsTransientRuntimeScope(n.Operand)
	case *ast.TypeCastExpression:
		return expressionAllowsTransientRuntimeScope(n.Expression)
	case *ast.BinaryExpression:
		return expressionAllowsTransientRuntimeScope(n.Left) &&
			expressionAllowsTransientRuntimeScope(n.Right)
	case *ast.FunctionCall:
		if !expressionAllowsTransientRuntimeScope(n.Callee) {
			return false
		}
		for _, arg := range n.Arguments {
			if !expressionAllowsTransientRuntimeScope(arg) {
				return false
			}
		}
		return true
	case *ast.BlockExpression:
		return blockAllowsTransientRuntimeScope(n)
	case *ast.AssignmentExpression:
		if targetExpr, ok := n.Left.(ast.Expression); ok && !expressionAllowsTransientRuntimeScope(targetExpr) {
			return false
		}
		return expressionAllowsTransientRuntimeScope(n.Right)
	case *ast.RangeExpression:
		return expressionAllowsTransientRuntimeScope(n.Start) &&
			expressionAllowsTransientRuntimeScope(n.End)
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if !expressionAllowsTransientRuntimeScope(part) {
				return false
			}
		}
		return true
	case *ast.MemberAccessExpression:
		if !expressionAllowsTransientRuntimeScope(n.Object) {
			return false
		}
		memberExpr, ok := n.Member.(ast.Expression)
		return !ok || expressionAllowsTransientRuntimeScope(memberExpr)
	case *ast.IndexExpression:
		return expressionAllowsTransientRuntimeScope(n.Object) &&
			expressionAllowsTransientRuntimeScope(n.Index)
	case *ast.StructLiteral:
		for _, field := range n.Fields {
			if field != nil && !expressionAllowsTransientRuntimeScope(field.Value) {
				return false
			}
		}
		for _, src := range n.FunctionalUpdateSources {
			if !expressionAllowsTransientRuntimeScope(src) {
				return false
			}
		}
		return true
	case *ast.IfExpression:
		if !expressionAllowsTransientRuntimeScope(n.IfCondition) ||
			!blockAllowsTransientRuntimeScope(n.IfBody) {
			return false
		}
		for _, clause := range n.ElseIfClauses {
			if clause == nil {
				continue
			}
			if !expressionAllowsTransientRuntimeScope(clause.Condition) ||
				!blockAllowsTransientRuntimeScope(clause.Body) {
				return false
			}
		}
		return blockAllowsTransientRuntimeScope(n.ElseBody)
	case *ast.MatchExpression:
		if !expressionAllowsTransientRuntimeScope(n.Subject) {
			return false
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if !expressionAllowsTransientRuntimeScope(clause.Guard) ||
				!expressionAllowsTransientRuntimeScope(clause.Body) {
				return false
			}
		}
		return true
	case *ast.RescueExpression:
		if !expressionAllowsTransientRuntimeScope(n.MonitoredExpression) {
			return false
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if !expressionAllowsTransientRuntimeScope(clause.Guard) ||
				!expressionAllowsTransientRuntimeScope(clause.Body) {
				return false
			}
		}
		return true
	case *ast.EnsureExpression:
		return expressionAllowsTransientRuntimeScope(n.TryExpression) &&
			blockAllowsTransientRuntimeScope(n.EnsureBlock)
	case *ast.OrElseExpression:
		return expressionAllowsTransientRuntimeScope(n.Expression) &&
			blockAllowsTransientRuntimeScope(n.Handler)
	case *ast.BreakpointExpression:
		return blockAllowsTransientRuntimeScope(n.Body)
	case *ast.PropagationExpression:
		return expressionAllowsTransientRuntimeScope(n.Expression)
	case *ast.AwaitExpression:
		return expressionAllowsTransientRuntimeScope(n.Expression)
	case *ast.LoopExpression:
		return blockAllowsTransientRuntimeScope(n.Body)
	default:
		return false
	}
}
