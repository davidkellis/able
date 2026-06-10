package interpreter

import "able/interpreter-go/pkg/ast"

func slotEligibleNonCapturingLambda(expr *ast.LambdaExpression) bool {
	if expr == nil || expr.Body == nil {
		return true
	}
	bound := make(map[string]int, len(expr.Params))
	for _, param := range expr.Params {
		if param == nil {
			return false
		}
		if _, ok := param.Name.(*ast.Identifier); !ok {
			return false
		}
		collectScopedPatternBindings(param.Name, bound)
	}
	return !exprHasFreeIdentifier(expr.Body, bound)
}

func exprHasFreeIdentifier(expr ast.Expression, bound map[string]int) bool {
	if expr == nil {
		return false
	}
	switch n := expr.(type) {
	case *ast.Identifier:
		return !isScopedNameBound(n.Name, bound)
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.CharLiteral,
		*ast.NilLiteral, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.PlaceholderExpression, *ast.ImplicitMemberExpression:
		return false
	case *ast.UnaryExpression:
		return exprHasFreeIdentifier(n.Operand, bound)
	case *ast.BinaryExpression:
		return exprHasFreeIdentifier(n.Left, bound) || exprHasFreeIdentifier(n.Right, bound)
	case *ast.AssignmentExpression:
		return assignmentHasFreeIdentifier(n, bound)
	case *ast.FunctionCall:
		if exprHasFreeIdentifier(n.Callee, bound) {
			return true
		}
		for _, arg := range n.Arguments {
			if exprHasFreeIdentifier(arg, bound) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		return exprHasFreeIdentifier(n.Object, bound)
	case *ast.IndexExpression:
		return exprHasFreeIdentifier(n.Object, bound) || exprHasFreeIdentifier(n.Index, bound)
	case *ast.BlockExpression:
		return blockHasFreeIdentifier(n, bound)
	case *ast.IfExpression:
		if exprHasFreeIdentifier(n.IfCondition, bound) || blockHasFreeIdentifier(n.IfBody, bound) || blockHasFreeIdentifier(n.ElseBody, bound) {
			return true
		}
		for _, clause := range n.ElseIfClauses {
			if clause == nil {
				continue
			}
			if exprHasFreeIdentifier(clause.Condition, bound) || blockHasFreeIdentifier(clause.Body, bound) {
				return true
			}
		}
		return false
	case *ast.MatchExpression:
		return matchHasFreeIdentifier(n, bound)
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if exprHasFreeIdentifier(el, bound) {
				return true
			}
		}
		return false
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if exprHasFreeIdentifier(part, bound) {
				return true
			}
		}
		return false
	case *ast.TypeCastExpression:
		return exprHasFreeIdentifier(n.Expression, bound)
	case *ast.RangeExpression:
		return exprHasFreeIdentifier(n.Start, bound) || exprHasFreeIdentifier(n.End, bound)
	case *ast.PropagationExpression:
		return exprHasFreeIdentifier(n.Expression, bound)
	case *ast.AwaitExpression:
		return exprHasFreeIdentifier(n.Expression, bound)
	case *ast.LoopExpression:
		return blockHasFreeIdentifier(n.Body, bound)
	case *ast.LambdaExpression:
		return !slotEligibleNonCapturingLambda(n)
	case *ast.StructLiteral, *ast.MapLiteral, *ast.SpawnExpression,
		*ast.IteratorLiteral, *ast.RescueExpression, *ast.EnsureExpression,
		*ast.BreakpointExpression, *ast.OrElseExpression:
		return true
	default:
		return true
	}
}

func stmtHasFreeIdentifier(stmt ast.Statement, bound map[string]int) bool {
	if stmt == nil {
		return false
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return exprHasFreeIdentifier(expr, bound)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		return exprHasFreeIdentifier(s.Argument, bound)
	case *ast.YieldStatement:
		return exprHasFreeIdentifier(s.Expression, bound)
	case *ast.RaiseStatement:
		return exprHasFreeIdentifier(s.Expression, bound)
	case *ast.BreakStatement:
		return exprHasFreeIdentifier(s.Value, bound)
	case *ast.ContinueStatement:
		return false
	case *ast.ForLoop:
		if exprHasFreeIdentifier(s.Iterable, bound) {
			return true
		}
		pushScopedPatternBindings(s.Pattern, bound)
		free := blockHasFreeIdentifier(s.Body, bound)
		popScopedPatternBindings(s.Pattern, bound)
		return free
	case *ast.WhileLoop:
		return exprHasFreeIdentifier(s.Condition, bound) || blockHasFreeIdentifier(s.Body, bound)
	case *ast.PackageStatement, *ast.PreludeStatement, *ast.ImportStatement,
		*ast.DynImportStatement, *ast.StructDefinition, *ast.UnionDefinition,
		*ast.TypeAliasDefinition, *ast.InterfaceDefinition, *ast.ExternFunctionBody:
		return false
	default:
		return true
	}
}

func blockHasFreeIdentifier(block *ast.BlockExpression, bound map[string]int) bool {
	if block == nil {
		return false
	}
	snapshot := snapshotScopedNameCounts(bound)
	defer restoreScopedNameCounts(bound, snapshot)
	for _, stmt := range block.Body {
		if stmtHasFreeIdentifier(stmt, bound) {
			return true
		}
	}
	return false
}

func assignmentHasFreeIdentifier(expr *ast.AssignmentExpression, bound map[string]int) bool {
	if expr == nil {
		return false
	}
	if exprHasFreeIdentifier(expr.Right, bound) {
		return true
	}
	name, ok := resolveAssignmentTargetName(expr.Left)
	if ok && name != "" && name != "_" {
		bound[name]++
		return false
	}
	switch target := expr.Left.(type) {
	case *ast.IndexExpression:
		return exprHasFreeIdentifier(target.Object, bound) || exprHasFreeIdentifier(target.Index, bound)
	case *ast.MemberAccessExpression:
		return exprHasFreeIdentifier(target.Object, bound)
	case *ast.ImplicitMemberExpression:
		return false
	default:
		return true
	}
}

func matchHasFreeIdentifier(expr *ast.MatchExpression, bound map[string]int) bool {
	if expr == nil {
		return false
	}
	if exprHasFreeIdentifier(expr.Subject, bound) {
		return true
	}
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		snapshot := snapshotScopedNameCounts(bound)
		pushScopedPatternBindings(clause.Pattern, bound)
		free := exprHasFreeIdentifier(clause.Guard, bound) || exprHasFreeIdentifier(clause.Body, bound)
		restoreScopedNameCounts(bound, snapshot)
		if free {
			return true
		}
	}
	return false
}

func isScopedNameBound(name string, bound map[string]int) bool {
	if name == "" {
		return true
	}
	return bound[name] > 0
}

func pushScopedPatternBindings(pattern ast.Pattern, bound map[string]int) {
	if bound == nil {
		return
	}
	collectScopedPatternBindings(pattern, bound)
}

func popScopedPatternBindings(pattern ast.Pattern, bound map[string]int) {
	if bound == nil {
		return
	}
	names := patternBoundIdentifierNameSet(pattern)
	for name := range names {
		if bound[name] <= 1 {
			delete(bound, name)
		} else {
			bound[name]--
		}
	}
}

func collectScopedPatternBindings(pattern ast.Pattern, bound map[string]int) {
	names := patternBoundIdentifierNameSet(pattern)
	for name := range names {
		bound[name]++
	}
}

func snapshotScopedNameCounts(bound map[string]int) map[string]int {
	if len(bound) == 0 {
		return nil
	}
	snapshot := make(map[string]int, len(bound))
	for name, count := range bound {
		snapshot[name] = count
	}
	return snapshot
}

func restoreScopedNameCounts(bound map[string]int, snapshot map[string]int) {
	for name := range bound {
		if _, ok := snapshot[name]; !ok {
			delete(bound, name)
		}
	}
	for name, count := range snapshot {
		bound[name] = count
	}
}
