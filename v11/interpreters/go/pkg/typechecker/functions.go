package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func isVoidType(t Type) bool {
	if t == nil {
		return false
	}
	if name, ok := structName(t); ok && name == "void" {
		return true
	}
	return false
}

func (c *Checker) checkFunctionDefinition(env *Environment, def *ast.FunctionDefinition) []Diagnostic {
	if def == nil {
		return nil
	}

	bodyEnv := env.Extend()
	var diags []Diagnostic

	var fnType FunctionType
	if def.ID != nil {
		if typ, ok := env.Lookup(def.ID.Name); ok {
			if sig, ok := typ.(FunctionType); ok {
				fnType = sig
			}
		}
		if fnType.Params == nil && fnType.Return == nil {
			if typ, ok := c.global.Lookup(def.ID.Name); ok {
				if sig, ok := typ.(FunctionType); ok {
					fnType = sig
				}
			}
		}
	}

	c.pushConstraintScope(fnType.TypeParams, fnType.Where)
	defer c.popConstraintScope()

	// Bind parameters using the collected signature when available.
	for idx, param := range def.Params {
		if param == nil {
			continue
		}
		paramType := Type(UnknownType{})
		if idx < len(fnType.Params) && fnType.Params[idx] != nil {
			paramType = fnType.Params[idx]
		} else if param.ParamType != nil {
			paramType = c.resolveTypeReference(param.ParamType)
		}
		if target, ok := param.Name.(ast.AssignmentTarget); ok {
			diags = append(diags, c.bindPattern(bodyEnv, target, paramType, true, nil)...)
		} else {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unsupported function parameter pattern %T", param.Name),
				Node:    param,
			})
		}
	}

	expectedReturn := fnType.Return
	c.pushReturnType(expectedReturn)
	defer c.popReturnType()

	if def.Body != nil {
		bodyDiags, bodyType := c.checkExpression(bodyEnv, def.Body)
		diags = append(diags, bodyDiags...)

		if isVoidType(expectedReturn) {
			bodyType = expectedReturn
		}

		if expectedReturn != nil && !isUnknownType(expectedReturn) && bodyType != nil && !isUnknownType(bodyType) {
			if coerced, ok := normalizeResultReturn(bodyType, expectedReturn); ok {
				bodyType = coerced
			} else if !typeAssignable(bodyType, expectedReturn) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: function '%s' body returns %s, expected %s", defName(def), typeName(bodyType), typeName(expectedReturn)),
					Node:    def.Body,
				})
			}
		}
	}

	return diags
}

func defName(def *ast.FunctionDefinition) string {
	if def != nil && def.ID != nil {
		return def.ID.Name
	}
	return "<anonymous>"
}

func (c *Checker) checkLambdaExpression(env *Environment, expr *ast.LambdaExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	lambdaEnv := env.Extend()
	var diags []Diagnostic

	paramTypes := make([]Type, len(expr.Params))
	for idx, param := range expr.Params {
		if param == nil {
			paramTypes[idx] = UnknownType{}
			continue
		}
		paramType := Type(UnknownType{})
		if param.ParamType != nil {
			paramType = c.resolveTypeReference(param.ParamType)
		}
		paramTypes[idx] = paramType
		if target, ok := param.Name.(ast.AssignmentTarget); ok {
			diags = append(diags, c.bindPattern(lambdaEnv, target, paramType, true, nil)...)
		} else {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unsupported lambda parameter pattern %T", param.Name),
				Node:    param,
			})
		}
	}

	var (
		expectedReturn Type = UnknownType{}
		bodyType       Type = UnknownType{}
	)
	if expr.ReturnType != nil {
		expectedReturn = c.resolveTypeReference(expr.ReturnType)
	}

	c.pushReturnType(expectedReturn)
	bodyDiags, inferredReturn := c.checkExpression(lambdaEnv, expr.Body)
	c.popReturnType()

	diags = append(diags, bodyDiags...)
	bodyType = inferredReturn

	if isVoidType(expectedReturn) {
		bodyType = expectedReturn
	}

	if expectedReturn != nil && !isUnknownType(expectedReturn) {
		if bodyType != nil && !isUnknownType(bodyType) {
			if coerced, ok := normalizeResultReturn(bodyType, expectedReturn); ok {
				bodyType = coerced
			} else if !typeAssignable(bodyType, expectedReturn) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: lambda body returns %s, expected %s", typeName(bodyType), typeName(expectedReturn)),
					Node:    expr.Body,
				})
			}
		}
		bodyType = expectedReturn
	}

	fnType := FunctionType{
		Params: paramTypes,
		Return: bodyType,
	}
	c.infer.set(expr, fnType)
	return diags, fnType
}

func (c *Checker) checkReturnStatement(env *Environment, stmt *ast.ReturnStatement) []Diagnostic {
	if stmt == nil {
		return nil
	}

	expected, ok := c.currentReturnType()
	var diags []Diagnostic
	if !ok {
		diags = append(diags, Diagnostic{
			Message: "typechecker: return statement outside function",
			Node:    stmt,
		})
		if stmt.Argument != nil {
			exprDiags, _ := c.checkExpression(env, stmt.Argument)
			diags = append(diags, exprDiags...)
		}
		return diags
	}

	var returnType Type = PrimitiveType{Kind: PrimitiveNil}
	if stmt.Argument != nil {
		argDiags, argType := c.checkExpression(env, stmt.Argument)
		diags = append(diags, argDiags...)
		returnType = argType
	}

	if isVoidType(expected) {
		returnType = expected
	}

	if expected != nil && !isUnknownType(expected) {
		if stmt.Argument == nil {
			nilType := PrimitiveType{Kind: PrimitiveNil}
			if !typeAssignable(nilType, expected) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: return expects %s, got Nil", typeName(expected)),
					Node:    stmt,
				})
			}
			returnType = expected
		} else if returnType != nil && !isUnknownType(returnType) {
			if coerced, ok := normalizeResultReturn(returnType, expected); ok {
				returnType = coerced
			} else if !typeAssignable(returnType, expected) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: return expects %s, got %s", typeName(expected), typeName(returnType)),
					Node:    stmt,
				})
			} else {
				returnType = expected
			}
		} else {
			returnType = expected
		}
	}

	c.infer.set(stmt, returnType)
	return diags
}

func (c *Checker) checkImplementationDefinition(env *Environment, def *ast.ImplementationDefinition) []Diagnostic {
	if def == nil {
		return nil
	}
	var diags []Diagnostic
	localEnv := env.Extend()
	var spec *ImplementationSpec
	for i := range c.implementations {
		if c.implementations[i].Definition == def {
			spec = &c.implementations[i]
			break
		}
	}
	if spec != nil {
		c.pushConstraintScope(spec.TypeParams, spec.Where)
		defer c.popConstraintScope()
		if len(spec.Obligations) > 0 {
			subject := spec.Target
			obligations := populateObligationSubjects(spec.Obligations, subject)
			subst := map[string]Type{"Self": subject}
			obligations = substituteObligations(obligations, subst)
			obligationDiags := c.evaluateObligations(obligations)
			diags = append(diags, obligationDiags...)
			if len(obligationDiags) > 0 {
				return diags
			}
		}
	}
	for _, fn := range def.Definitions {
		if fn == nil {
			continue
		}
		if spec != nil && fn.ID != nil {
			if sig, ok := spec.Methods[fn.ID.Name]; ok {
				localEnv.Define(fn.ID.Name, sig)
			}
		}
		diags = append(diags, c.checkFunctionDefinition(localEnv, fn)...)
	}
	return diags
}

func (c *Checker) checkMethodsDefinition(env *Environment, def *ast.MethodsDefinition) []Diagnostic {
	if def == nil {
		return nil
	}
	var diags []Diagnostic
	localEnv := env.Extend()
	var spec *MethodSetSpec
	for i := range c.methodSets {
		if c.methodSets[i].Definition == def {
			spec = &c.methodSets[i]
			break
		}
	}
	if spec != nil {
		c.pushConstraintScope(spec.TypeParams, spec.Where)
		defer c.popConstraintScope()
	}
	for _, fn := range def.Definitions {
		if fn == nil {
			continue
		}
		if spec != nil && fn.ID != nil {
			if sig, ok := spec.Methods[fn.ID.Name]; ok {
				localEnv.Define(fn.ID.Name, sig)
			}
		}
		diags = append(diags, c.checkFunctionDefinition(localEnv, fn)...)
	}
	return diags
}
