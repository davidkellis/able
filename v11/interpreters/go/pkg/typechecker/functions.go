package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
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
	if def != nil {
		if _, dup := c.duplicateFunctions[def]; dup {
			return nil
		}
	}
	if def.ID != nil {
		if typ, ok := env.Lookup(def.ID.Name); ok && isUnknownType(typ) {
			return nil
		}
	}
	if def.ID != nil && env != nil {
		if _, exists := env.Lookup(def.ID.Name); !exists {
			env.Define(def.ID.Name, c.localFunctionSignature(def))
		}
	}
	c.pushFunctionGenericContext(def)
	defer c.popFunctionGenericContext()

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
			if msg, ok := literalMismatchMessage(bodyType, expectedReturn); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    def.Body,
				})
				return diags
			}
			if msg, ok := literalOverflowMessage(bodyType, expectedReturn); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    def.Body,
				})
				return diags
			}
			if coerced, ok := normalizeResultReturn(bodyType, expectedReturn); ok {
				bodyType = coerced
			} else {
				assignable := typeAssignable(bodyType, expectedReturn)
				if !assignable {
					if isResultType(expectedReturn) {
						if ok, _ := c.typeImplementsInterface(bodyType, InterfaceType{InterfaceName: "Error"}, nil); ok {
							assignable = true
					} else if name, ok := structName(bodyType); ok && strings.HasSuffix(name, "Error") {
						assignable = true
					}
				}
				}
				if !assignable {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf(
							"typechecker: function '%s' body returns %s, expected %s",
							defName(def),
							formatTypeForReturnDiagnostic(bodyType),
							formatTypeForReturnDiagnostic(expectedReturn),
						),
						Node: def.Body,
					})
				}
			}
		}
	}

	return diags
}

func (c *Checker) localFunctionSignature(def *ast.FunctionDefinition) FunctionType {
	if def == nil {
		return FunctionType{}
	}
	paramTypes := make([]Type, len(def.Params))
	for idx, param := range def.Params {
		if param == nil {
			paramTypes[idx] = UnknownType{}
			continue
		}
		if param.ParamType != nil {
			paramTypes[idx] = c.resolveTypeReference(param.ParamType)
		} else {
			paramTypes[idx] = UnknownType{}
		}
	}
	if def.IsMethodShorthand {
		paramTypes = append([]Type{UnknownType{}}, paramTypes...)
	}
	returnType := Type(UnknownType{})
	if def.ReturnType != nil {
		returnType = c.resolveTypeReference(def.ReturnType)
	}
	return FunctionType{Params: paramTypes, Return: returnType}
}

func (c *Checker) pushFunctionGenericContext(def *ast.FunctionDefinition) {
	if def == nil {
		return
	}
	ctx := functionGenericContext{
		def:      def,
		label:    fmt.Sprintf("fn %s", defName(def)),
		inferred: collectInferredGenericParams(def),
	}
	c.functionGenericStack = append(c.functionGenericStack, ctx)
}

func (c *Checker) popFunctionGenericContext() {
	if len(c.functionGenericStack) == 0 {
		return
	}
	c.functionGenericStack = c.functionGenericStack[:len(c.functionGenericStack)-1]
}

func collectInferredGenericParams(def *ast.FunctionDefinition) map[string]*ast.GenericParameter {
	inferred := make(map[string]*ast.GenericParameter)
	if def == nil {
		return inferred
	}
	for _, param := range def.GenericParams {
		if param == nil || !param.IsInferred || param.Name == nil || param.Name.Name == "" {
			continue
		}
		inferred[param.Name.Name] = param
	}
	return inferred
}

func defName(def *ast.FunctionDefinition) string {
	if def != nil && def.ID != nil {
		return def.ID.Name
	}
	return "<anonymous>"
}

func formatTypeForReturnDiagnostic(t Type) string {
	if t == nil {
		return "Unknown"
	}
	switch val := t.(type) {
	case UnknownType:
		return "Unknown"
	case TypeParameterType:
		return "Unknown"
	case ArrayType:
		return strings.TrimSpace("Array " + formatTypeForReturnDiagnostic(val.Element))
	case NullableType:
		return formatTypeForReturnDiagnostic(val.Inner) + "?"
	case RangeType:
		return strings.TrimSpace("Range " + formatTypeForReturnDiagnostic(val.Element))
	case IteratorType:
		return strings.TrimSpace("Iterator " + formatTypeForReturnDiagnostic(val.Element))
	case ProcType:
		return strings.TrimSpace("Proc " + formatTypeForReturnDiagnostic(val.Result))
	case FutureType:
		return strings.TrimSpace("Future " + formatTypeForReturnDiagnostic(val.Result))
	case AppliedType:
		base := formatTypeForReturnDiagnostic(val.Base)
		if len(val.Arguments) == 0 {
			return base
		}
		args := make([]string, len(val.Arguments))
		for i, arg := range val.Arguments {
			args[i] = formatTypeForReturnDiagnostic(arg)
		}
		return strings.TrimSpace(base + " " + strings.Join(args, " "))
	case UnionLiteralType:
		if len(val.Members) == 0 {
			return "Union"
		}
		members := make([]string, len(val.Members))
		for i, member := range val.Members {
			members[i] = formatTypeForReturnDiagnostic(member)
		}
		return strings.Join(members, " | ")
	case FunctionType:
		params := make([]string, len(val.Params))
		for i, param := range val.Params {
			params[i] = formatTypeForReturnDiagnostic(param)
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(params, ", "), formatTypeForReturnDiagnostic(val.Return))
	default:
		return formatType(t)
	}
}

func interfaceFromType(t Type) (InterfaceType, []Type, bool) {
	switch v := t.(type) {
	case InterfaceType:
		return v, nil, true
	case AppliedType:
		if iface, ok := v.Base.(InterfaceType); ok {
			return iface, v.Arguments, true
		}
	}
	return InterfaceType{}, nil, false
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
			if msg, ok := literalMismatchMessage(bodyType, expectedReturn); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    expr.Body,
				})
				bodyType = expectedReturn
				fnType := FunctionType{Params: paramTypes, Return: bodyType}
				c.infer.set(expr, fnType)
				return diags, fnType
			}
			if msg, ok := literalOverflowMessage(bodyType, expectedReturn); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    expr.Body,
				})
				bodyType = expectedReturn
				fnType := FunctionType{Params: paramTypes, Return: bodyType}
				c.infer.set(expr, fnType)
				return diags, fnType
			}
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

	voidType := StructType{StructName: "void"}
	var returnType Type = voidType
	if stmt.Argument != nil {
		argDiags, argType := c.checkExpressionWithExpectedType(env, stmt.Argument, expected)
		diags = append(diags, argDiags...)
		returnType = argType
	}

	if isVoidType(expected) {
		returnType = expected
	}

	if expected != nil && !isUnknownType(expected) {
		if stmt.Argument == nil {
			if !isVoidType(expected) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: return expects %s, got void", typeName(expected)),
					Node:    stmt,
				})
			}
			returnType = expected
		} else if returnType != nil && !isUnknownType(returnType) {
			if msg, ok := literalMismatchMessage(returnType, expected); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    stmt,
				})
				return diags
			}
			if msg, ok := literalOverflowMessage(returnType, expected); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    stmt,
				})
				return diags
			}
			if coerced, ok := normalizeResultReturn(returnType, expected); ok {
				returnType = coerced
			} else {
				assignable := typeAssignable(returnType, expected)
				if !assignable {
					if iface, args, ok := interfaceFromType(expected); ok {
						if ok, _ := c.typeImplementsInterface(returnType, iface, args); ok {
							assignable = true
						}
					}
				}
				if !assignable {
				if isResultType(expected) {
					if ok, _ := c.typeImplementsInterface(returnType, InterfaceType{InterfaceName: "Error"}, nil); ok {
						assignable = true
					} else if name, ok := structName(returnType); ok && strings.HasSuffix(name, "Error") {
						assignable = true
					}
					}
				}
				if !assignable {
					message := fmt.Sprintf("typechecker: return expects %s, got %s", typeName(expected), typeName(returnType))
					if msg, ok := literalMismatchMessage(returnType, expected); ok {
						message = fmt.Sprintf("typechecker: %s", msg)
					}
					diags = append(diags, Diagnostic{
						Message: message,
						Node:    stmt,
					})
				} else {
					returnType = expected
				}
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
	return diags
}

func (c *Checker) checkMethodsDefinition(env *Environment, def *ast.MethodsDefinition) []Diagnostic {
	if def == nil {
		return nil
	}
	var diags []Diagnostic
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
	return diags
}

func literalOverflowMessage(actual Type, expected Type) (string, bool) {
	intVal, ok := actual.(IntegerType)
	if !ok || intVal.Literal == nil || intVal.Explicit {
		return "", false
	}
	var suffix string
	switch v := expected.(type) {
	case IntegerType:
		suffix = v.Suffix
	case StructType:
		suffix = v.StructName
	case StructInstanceType:
		suffix = v.StructName
	}
	if suffix == "" {
		return "", false
	}
	bounds, ok := integerBounds[suffix]
	if !ok {
		return "", false
	}
	if intVal.Literal.Cmp(bounds.min) < 0 || intVal.Literal.Cmp(bounds.max) > 0 {
		return fmt.Sprintf("literal %s does not fit in %s", intVal.Literal.String(), suffix), true
	}
	return "", false
}
