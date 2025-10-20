package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkExpression(env *Environment, expr ast.Expression) ([]Diagnostic, Type) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		suffix := "i32"
		if e.IntegerType != nil {
			suffix = string(*e.IntegerType)
		}
		typ := IntegerType{Suffix: suffix}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.FloatLiteral:
		suffix := "f64"
		if e.FloatType != nil {
			suffix = string(*e.FloatType)
		}
		typ := FloatType{Suffix: suffix}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.BooleanLiteral:
		typ := PrimitiveType{Kind: PrimitiveBool}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.NilLiteral:
		typ := PrimitiveType{Kind: PrimitiveNil}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.StringLiteral:
		typ := PrimitiveType{Kind: PrimitiveString}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.ArrayLiteral:
		return c.checkArrayLiteral(env, e)
	case *ast.BlockExpression:
		blockEnv := env.Extend()
		var (
			diags      []Diagnostic
			resultType Type = UnknownType{}
		)
		for idx, stmt := range e.Body {
			switch s := stmt.(type) {
			case *ast.AssignmentExpression:
				assignDiags := c.checkStatement(blockEnv, s)
				diags = append(diags, assignDiags...)
				if idx == len(e.Body)-1 {
					if inferred, ok := c.infer.get(s.Right); ok {
						resultType = inferred
					}
				}
			case ast.Expression:
				exprDiags, exprType := c.checkExpression(blockEnv, s)
				diags = append(diags, exprDiags...)
				if idx == len(e.Body)-1 {
					resultType = exprType
				}
			default:
				diags = append(diags, c.checkStatement(blockEnv, s)...)
			}
		}
		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.IfExpression:
		var diags []Diagnostic
		condDiags, condType := c.checkExpression(env, e.IfCondition)
		diags = append(diags, condDiags...)
		if !typeAssignable(condType, PrimitiveType{Kind: PrimitiveBool}) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: if condition must be bool",
				Node:    e.IfCondition,
			})
		}

		branchTypes := make([]Type, 0, 1+len(e.OrClauses))
		if e.IfBody != nil {
			bodyDiags, bodyType := c.checkExpression(env, e.IfBody)
			diags = append(diags, bodyDiags...)
			branchTypes = append(branchTypes, bodyType)
		} else {
			branchTypes = append(branchTypes, UnknownType{})
		}

		for _, clause := range e.OrClauses {
			if clause == nil {
				continue
			}
			if clause.Condition != nil {
				orCondDiags, orCondType := c.checkExpression(env, clause.Condition)
				diags = append(diags, orCondDiags...)
				if !typeAssignable(orCondType, PrimitiveType{Kind: PrimitiveBool}) {
					diags = append(diags, Diagnostic{
						Message: "typechecker: if-or condition must be bool",
						Node:    clause.Condition,
					})
				}
			}
			if clause.Body != nil {
				bodyDiags, bodyType := c.checkExpression(env, clause.Body)
				diags = append(diags, bodyDiags...)
				branchTypes = append(branchTypes, bodyType)
			} else {
				branchTypes = append(branchTypes, UnknownType{})
			}
		}

		resultType := mergeBranchTypes(branchTypes)
		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.UnaryExpression:
		return c.checkUnaryExpression(env, e)
	case *ast.BinaryExpression:
		return c.checkBinaryExpression(env, e)
	case *ast.LambdaExpression:
		return c.checkLambdaExpression(env, e)
	case *ast.FunctionCall:
		var diags []Diagnostic
		var builtinName string
		if ident, ok := e.Callee.(*ast.Identifier); ok && ident != nil {
			builtinName = ident.Name
		}
		calleeDiags, calleeType := c.checkExpression(env, e.Callee)
		diags = append(diags, calleeDiags...)

		argTypes := make([]Type, len(e.Arguments))
		for i, arg := range e.Arguments {
			argDiags, argType := c.checkExpression(env, arg)
			diags = append(diags, argDiags...)
			argTypes[i] = argType
		}

		resultType := Type(UnknownType{})
		if fnType, ok := calleeType.(FunctionType); ok {
			if len(e.TypeArguments) > 0 && len(fnType.TypeParams) != len(e.TypeArguments) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: function expects %d type arguments, got %d", len(fnType.TypeParams), len(e.TypeArguments)),
					Node:    e,
				})
			}
			instantiated, instDiags := c.instantiateFunctionCall(fnType, e, argTypes)
			diags = append(diags, instDiags...)
			if len(instantiated.Obligations) > 0 {
				c.obligations = append(c.obligations, instantiated.Obligations...)
			}
			if len(instantiated.Params) != len(argTypes) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: function expects %d arguments, got %d", len(instantiated.Params), len(argTypes)),
					Node:    e,
				})
			} else {
				for i, expected := range instantiated.Params {
					if !typeAssignable(argTypes[i], expected) {
						diags = append(diags, Diagnostic{
							Message: fmt.Sprintf("typechecker: argument %d has type %s, expected %s", i+1, typeName(argTypes[i]), typeName(expected)),
							Node:    e.Arguments[i],
						})
					}
				}
			}
			resultType = instantiated.Return
		} else if !isUnknownType(calleeType) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: cannot call non-function value",
				Node:    e.Callee,
			})
		}

		diags = append(diags, c.checkBuiltinCallContext(builtinName, e)...)

		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.MemberAccessExpression:
		return c.checkMemberAccess(env, e)
	case *ast.IndexExpression:
		return c.checkIndexExpression(env, e)
	case *ast.StructLiteral:
		return c.checkStructLiteral(env, e)
	case *ast.MatchExpression:
		return c.checkMatchExpression(env, e)
	case *ast.RangeExpression:
		return c.checkRangeExpression(env, e)
	case *ast.RescueExpression:
		return c.checkRescueExpression(env, e)
	case *ast.OrElseExpression:
		return c.checkOrElseExpression(env, e)
	case *ast.EnsureExpression:
		return c.checkEnsureExpression(env, e)
	case *ast.BreakpointExpression:
		return c.checkBreakpointExpression(env, e)
	case *ast.StringInterpolation:
		return c.checkStringInterpolation(env, e)
	case *ast.ProcExpression:
		return c.checkProcExpression(env, e)
	case *ast.SpawnExpression:
		return c.checkSpawnExpression(env, e)
	case *ast.PropagationExpression:
		return c.checkPropagationExpression(env, e)
	case *ast.Identifier:
		if typ, ok := env.Lookup(e.Name); ok {
			c.infer.set(e, typ)
			return nil, typ
		}
		if c.allowDynamicLookups {
			c.infer.set(e, UnknownType{})
			return nil, UnknownType{}
		}
		diag := Diagnostic{Message: fmt.Sprintf("typechecker: undefined identifier '%s'", e.Name), Node: expr}
		return []Diagnostic{diag}, UnknownType{}
	default:
		diag := Diagnostic{Message: fmt.Sprintf("typechecker: unsupported expression %T", expr), Node: expr}
		return []Diagnostic{diag}, UnknownType{}
	}
}

func (c *Checker) checkStatement(env *Environment, stmt ast.Statement) []Diagnostic {
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
		var diags []Diagnostic
		if s.Operator == ast.AssignmentDeclare {
			diags = append(diags, c.bindPattern(env, s.Left, UnknownType{}, true)...)
		}
		rhsDiags, typ := c.checkExpression(env, s.Right)
		diags = append(diags, rhsDiags...)
		if typ == nil {
			typ = UnknownType{}
		}
		if s.Operator == ast.AssignmentDeclare {
			return append(diags, c.bindPattern(env, s.Left, typ, true)...)
		}
		return diags
	case *ast.WhileLoop:
		diags, _ := c.checkWhileLoop(env, s)
		return diags
	case *ast.ForLoop:
		diags, _ := c.checkForLoop(env, s)
		return diags
	case *ast.RaiseStatement:
		return c.checkRaiseStatement(env, s)
	case *ast.RethrowStatement:
		return c.checkRethrowStatement(s)
	case *ast.BreakStatement:
		return c.checkBreakStatement(env, s)
	case *ast.ContinueStatement:
		return c.checkContinueStatement(s)
	case *ast.StructDefinition, *ast.UnionDefinition, *ast.InterfaceDefinition:
		return nil
	case *ast.DynImportStatement:
		placeholder := Type(UnknownType{})
		if s.IsWildcard {
			c.allowDynamicLookups = true
		}
		if s.Alias != nil && s.Alias.Name != "" {
			env.Define(s.Alias.Name, placeholder)
		}
		for _, sel := range s.Selectors {
			if sel == nil {
				continue
			}
			if sel.Alias != nil && sel.Alias.Name != "" {
				env.Define(sel.Alias.Name, placeholder)
				continue
			}
			if sel.Name != nil && sel.Name.Name != "" {
				env.Define(sel.Name.Name, placeholder)
			}
		}
		return nil
	case *ast.FunctionDefinition:
		return c.checkFunctionDefinition(env, s)
	case *ast.ImplementationDefinition:
		return c.checkImplementationDefinition(env, s)
	case *ast.MethodsDefinition:
		return c.checkMethodsDefinition(env, s)
	case *ast.ReturnStatement:
		return c.checkReturnStatement(env, s)
	case ast.Expression:
		diags, _ := c.checkExpression(env, s)
		return diags
	default:
		return []Diagnostic{{Message: fmt.Sprintf("typechecker: unsupported statement %T", stmt), Node: stmt}}
	}
}
