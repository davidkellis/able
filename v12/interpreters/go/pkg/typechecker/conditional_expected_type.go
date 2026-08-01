package typechecker

import "able/interpreter-go/pkg/ast"

func (c *Checker) checkIfExpressionWithExpectedType(
	env *Environment,
	expr *ast.IfExpression,
	expected Type,
) ([]Diagnostic, Type) {
	var diagnostics []Diagnostic
	conditionDiagnostics, _ := c.checkExpression(env, expr.IfCondition)
	diagnostics = append(diagnostics, conditionDiagnostics...)

	branchTypes := make([]Type, 0, 2+len(expr.ElseIfClauses))
	checkBranch := func(branch ast.Expression) {
		if branch == nil {
			branchTypes = append(branchTypes, UnknownType{})
			return
		}
		var branchDiagnostics []Diagnostic
		var branchType Type
		if expected != nil && !isUnknownType(expected) {
			branchDiagnostics, branchType = c.checkExpressionWithExpectedType(env, branch, expected)
		} else {
			branchDiagnostics, branchType = c.checkExpression(env, branch)
		}
		diagnostics = append(diagnostics, branchDiagnostics...)
		branchTypes = append(branchTypes, branchType)
	}

	checkBranch(expr.IfBody)
	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			continue
		}
		clauseDiagnostics, _ := c.checkExpression(env, clause.Condition)
		diagnostics = append(diagnostics, clauseDiagnostics...)
		checkBranch(clause.Body)
	}
	if expr.ElseBody != nil {
		checkBranch(expr.ElseBody)
	} else {
		branchTypes = append(branchTypes, PrimitiveType{Kind: PrimitiveNil})
	}

	resultType := buildUnionType(branchTypes...)
	if expected != nil && !isUnknownType(expected) {
		allContextual := true
		for _, branchType := range branchTypes {
			if !c.typeAssignableInExpectedContext(branchType, expected) {
				allContextual = false
				break
			}
		}
		if allContextual {
			resultType = expected
		}
	}
	c.infer.set(expr, resultType)
	return diagnostics, resultType
}

func (c *Checker) typeAssignableInExpectedContext(actual, expected Type) bool {
	if actual == nil || isUnknownType(actual) {
		return true
	}
	if typeAssignable(actual, expected) {
		return true
	}
	if _, ok := normalizeResultReturn(actual, expected); ok {
		return true
	}
	if c.typeAssignableToExpectedInterfaceMember(actual, expected) {
		return true
	}
	nullable, ok := expected.(NullableType)
	if !ok {
		return false
	}
	if primitive, ok := actual.(PrimitiveType); ok && primitive.Kind == PrimitiveNil {
		return true
	}
	if _, alreadyNullable := actual.(NullableType); alreadyNullable {
		return false
	}
	if typeAssignable(actual, nullable.Inner) {
		return true
	}
	if _, ok := normalizeResultReturn(actual, nullable.Inner); ok {
		return true
	}
	return c.typeAssignableToExpectedInterfaceMember(actual, nullable.Inner)
}
