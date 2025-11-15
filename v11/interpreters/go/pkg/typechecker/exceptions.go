package typechecker

import "able/interpreter10-go/pkg/ast"

func (c *Checker) checkPropagationExpression(env *Environment, expr *ast.PropagationExpression) ([]Diagnostic, Type) {
	bodyDiags, bodyType := c.checkExpression(env, expr.Expression)
	var diags []Diagnostic
	diags = append(diags, bodyDiags...)

	if union, ok := bodyType.(UnionLiteralType); ok {
		if len(union.Members) == 2 && unionContainsProcError(union) {
			// Treat propagation as extracting the success branch.
			success := unionSuccessBranch(union)
			c.infer.set(expr, success)
			return diags, success
		}
	}
	c.infer.set(expr, bodyType)
	return diags, bodyType
}

func unionContainsProcError(union UnionLiteralType) bool {
	for _, member := range union.Members {
		if member != nil && member.Name() == "Struct:ProcError" {
			return true
		}
	}
	return false
}

func unionSuccessBranch(union UnionLiteralType) Type {
	for _, member := range union.Members {
		if member == nil {
			continue
		}
		if member.Name() != "Struct:ProcError" {
			return member
		}
	}
	return UnknownType{}
}

func (c *Checker) lookupErrorType() Type {
	if c == nil || c.global == nil {
		return StructType{StructName: "Error"}
	}
	if typ, ok := c.global.Lookup("Error"); ok && typ != nil {
		return typ
	}
	return StructType{StructName: "Error"}
}

func (c *Checker) checkRaiseStatement(env *Environment, stmt *ast.RaiseStatement) []Diagnostic {
	if stmt == nil {
		return nil
	}
	if stmt.Expression == nil {
		return []Diagnostic{{
			Message: "typechecker: raise requires an expression",
			Node:    stmt,
		}}
	}
	diags, _ := c.checkExpression(env, stmt.Expression)
	c.infer.set(stmt, PrimitiveType{Kind: PrimitiveNil})
	return diags
}

func (c *Checker) checkRescueExpression(env *Environment, expr *ast.RescueExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var diags []Diagnostic
	monitoredDiags, monitoredType := c.checkExpression(env, expr.MonitoredExpression)
	diags = append(diags, monitoredDiags...)

	resultType := monitoredType
	errorType := c.lookupErrorType()
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		clauseEnv := env.Extend()
		if clause.Pattern != nil {
			if target, ok := clause.Pattern.(ast.AssignmentTarget); ok {
				diags = append(diags, c.bindPattern(clauseEnv, target, errorType, true, nil)...)
			}
		}
		c.pushRescueContext()
		if clause.Guard != nil {
			guardDiags, guardType := c.checkExpression(clauseEnv, clause.Guard)
			diags = append(diags, guardDiags...)
			if !typeAssignable(guardType, PrimitiveType{Kind: PrimitiveBool}) && !isUnknownType(guardType) {
				diags = append(diags, Diagnostic{
					Message: "typechecker: rescue guard must evaluate to bool",
					Node:    clause.Guard,
				})
			}
		}
		bodyDiags, bodyType := c.checkExpression(clauseEnv, clause.Body)
		diags = append(diags, bodyDiags...)
		resultType = mergeTypesAllowUnion(resultType, bodyType)
		c.popRescueContext()
	}

	c.infer.set(expr, resultType)
	return diags, resultType
}

func (c *Checker) checkOrElseExpression(env *Environment, expr *ast.OrElseExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var diags []Diagnostic
	exprDiags, exprType := c.checkExpression(env, expr.Expression)
	diags = append(diags, exprDiags...)

	handlerType := Type(PrimitiveType{Kind: PrimitiveNil})
	if expr.Handler != nil {
		handlerEnv := env.Extend()
		if expr.ErrorBinding != nil {
			handlerEnv.Define(expr.ErrorBinding.Name, c.lookupErrorType())
		}
		handlerDiags, inferredHandlerType := c.checkExpression(handlerEnv, expr.Handler)
		diags = append(diags, handlerDiags...)
		handlerType = inferredHandlerType
	}

	resultType := mergeTypesAllowUnion(exprType, handlerType)
	c.infer.set(expr, resultType)
	return diags, resultType
}

func (c *Checker) checkEnsureExpression(env *Environment, expr *ast.EnsureExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}
	var diags []Diagnostic
	tryDiags, tryType := c.checkExpression(env, expr.TryExpression)
	diags = append(diags, tryDiags...)

	if expr.EnsureBlock != nil {
		blockDiags, _ := c.checkExpression(env, expr.EnsureBlock)
		diags = append(diags, blockDiags...)
	}

	if tryType == nil {
		tryType = UnknownType{}
	}
	c.infer.set(expr, tryType)
	return diags, tryType
}

func (c *Checker) checkRethrowStatement(stmt *ast.RethrowStatement) []Diagnostic {
	if stmt == nil {
		return nil
	}
	// Without rescue-context tracking we simply allow rethrow.
	if !c.inRescueContext() {
		return []Diagnostic{{
			Message: "typechecker: rethrow is only valid inside rescue handlers",
			Node:    stmt,
		}}
	}
	c.infer.set(stmt, PrimitiveType{Kind: PrimitiveNil})
	return nil
}
