package typechecker

import "able/interpreter-go/pkg/ast"

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

	var stripOptionOrResult func(t Type) Type
	stripOptionOrResult = func(t Type) Type {
		switch v := t.(type) {
		case NullableType:
			return stripOptionOrResult(v.Inner)
		case AppliedType:
			if name, ok := structName(v.Base); ok && name == "Result" && len(v.Arguments) > 0 {
				return stripOptionOrResult(v.Arguments[0])
			}
		case UnionLiteralType:
			var members []Type
			for _, member := range v.Members {
				if member == nil || isUnknownType(member) {
					continue
				}
				if isFailureType(c, member) {
					continue
				}
				members = append(members, stripOptionOrResult(member))
			}
			return buildUnionType(members...)
		case UnionType:
			var members []Type
			for _, member := range v.Variants {
				if member == nil || isUnknownType(member) {
					continue
				}
				if isFailureType(c, member) {
					continue
				}
				members = append(members, stripOptionOrResult(member))
			}
			return buildUnionType(members...)
		}
		return t
	}

	successType := stripOptionOrResult(exprType)

	handlerType := Type(PrimitiveType{Kind: PrimitiveNil})
	handlerReturns := false
	if expr.Handler != nil {
		handlerEnv := env.Extend()
		if expr.ErrorBinding != nil {
			handlerEnv.Define(expr.ErrorBinding.Name, c.lookupErrorType())
		}
		handlerDiags, inferredHandlerType := c.checkExpression(handlerEnv, expr.Handler)
		diags = append(diags, handlerDiags...)
		handlerType = inferredHandlerType
		for _, stmt := range expr.Handler.Body {
			if stmt == nil {
				continue
			}
			if _, ok := stmt.(*ast.ReturnStatement); ok {
				handlerReturns = true
				break
			}
		}
	}

	resultType := successType
	if !handlerReturns {
		resultType = mergeTypesAllowUnion(successType, handlerType)
	}
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

func isFailureType(c *Checker, t Type) bool {
	if t == nil || isUnknownType(t) {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok && prim.Kind == PrimitiveNil {
		return true
	}
	if iface, ok := t.(InterfaceType); ok && iface.InterfaceName == "Error" {
		return true
	}
	if name, ok := structName(t); ok && name == "Error" {
		return true
	}
	if ok, _ := c.typeImplementsInterface(t, InterfaceType{InterfaceName: "Error"}, nil); ok {
		return true
	}
	if union, ok := t.(UnionLiteralType); ok {
		for _, member := range union.Members {
			if isFailureType(c, member) {
				return true
			}
		}
	}
	return false
}
