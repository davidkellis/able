package typechecker

import "able/interpreter10-go/pkg/ast"

func (c *Checker) checkMatchExpression(env *Environment, expr *ast.MatchExpression) ([]Diagnostic, Type) {
	var diags []Diagnostic
	if expr == nil {
		return nil, UnknownType{}
	}

	subjectDiags, subjectType := c.checkExpression(env, expr.Subject)
	diags = append(diags, subjectDiags...)

	branchTypes := make([]Type, 0, len(expr.Clauses))
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		clauseEnv := env.Extend()
		if clause.Pattern != nil {
			if target, ok := clause.Pattern.(ast.AssignmentTarget); ok {
				diags = append(diags, c.bindPattern(clauseEnv, target, subjectType, true)...)
			}
		}
		if clause.Guard != nil {
			guardDiags, guardType := c.checkExpression(clauseEnv, clause.Guard)
			diags = append(diags, guardDiags...)
			if !typeAssignable(guardType, PrimitiveType{Kind: PrimitiveBool}) {
				diags = append(diags, Diagnostic{
					Message: "typechecker: match guard must be bool",
					Node:    clause.Guard,
				})
			}
		}
		bodyDiags, bodyType := c.checkExpression(clauseEnv, clause.Body)
		diags = append(diags, bodyDiags...)
		branchTypes = append(branchTypes, bodyType)
	}

	resultType := mergeBranchTypes(branchTypes)
	c.infer.set(expr, resultType)
	return diags, resultType
}
