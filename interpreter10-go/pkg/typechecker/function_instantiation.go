package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) instantiateFunctionCall(fnType FunctionType, call *ast.FunctionCall, argTypes []Type) (FunctionType, []Diagnostic) {
	subst := make(map[string]Type)
	var diags []Diagnostic

	if fnType.TypeParams != nil && call != nil && len(call.TypeArguments) > 0 {
		limit := len(fnType.TypeParams)
		if len(call.TypeArguments) < limit {
			limit = len(call.TypeArguments)
		}
		for i := 0; i < limit; i++ {
			param := fnType.TypeParams[i]
			if param.Name == "" {
				continue
			}
			argExpr := call.TypeArguments[i]
			typ := c.resolveTypeReference(argExpr)
			if typ == nil {
				typ = UnknownType{}
			}
			if existing, ok := subst[param.Name]; ok {
				if !typesEquivalentForSignature(existing, typ) {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: type argument '%s' provided multiple times with incompatible types (%s vs %s)", param.Name, typeName(existing), typeName(typ)),
						Node:    call,
					})
				}
				continue
			}
			subst[param.Name] = typ
		}
	}

	paramLen := len(fnType.Params)
	if len(argTypes) < paramLen {
		paramLen = len(argTypes)
	}
	for i := 0; i < paramLen; i++ {
		var argNode ast.Node
		if call != nil && i < len(call.Arguments) {
			if node, ok := call.Arguments[i].(ast.Node); ok {
				argNode = node
			}
		}
		diags = append(diags, c.inferTypeArguments(fnType.Params[i], argTypes[i], subst, argNode, i)...)
	}

	inst := substituteFunctionType(fnType, subst)
	return inst, diags
}

func (c *Checker) inferTypeArguments(expected Type, actual Type, subst map[string]Type, node ast.Node, index int) []Diagnostic {
	if expected == nil || isUnknownType(expected) {
		return nil
	}
	if actual == nil || isUnknownType(actual) {
		return nil
	}
	switch exp := expected.(type) {
	case TypeParameterType:
		return bindTypeParameter(exp.ParameterName, actual, subst, node, index)
	case ArrayType:
		if elem, ok := arrayElementType(actual); ok && elem != nil {
			return c.inferTypeArguments(exp.Element, elem, subst, node, index)
		}
	case NullableType:
		switch act := actual.(type) {
		case NullableType:
			return c.inferTypeArguments(exp.Inner, act.Inner, subst, node, index)
		default:
			return c.inferTypeArguments(exp.Inner, actual, subst, node, index)
		}
	case AppliedType:
		switch act := actual.(type) {
		case AppliedType:
			var result []Diagnostic
			result = append(result, c.inferTypeArguments(exp.Base, act.Base, subst, node, index)...)
			limit := len(exp.Arguments)
			if len(act.Arguments) < limit {
				limit = len(act.Arguments)
			}
			for i := 0; i < limit; i++ {
				result = append(result, c.inferTypeArguments(exp.Arguments[i], act.Arguments[i], subst, node, index)...)
			}
			return result
		case StructInstanceType:
			if name, ok := structName(exp.Base); ok && (act.StructName == "" || act.StructName == name) {
				if info, ok := c.structInfoFromType(exp.Base, name); ok {
					return c.inferStructArguments(info, exp.Arguments, act, subst, node, index)
				}
			}
		case StructType:
			if name, ok := structName(exp.Base); ok && (act.StructName == "" || act.StructName == name) {
				instance := StructInstanceType{
					StructName: act.StructName,
					Fields:     act.Fields,
					Positional: act.Positional,
				}
				if info, ok := c.structInfoFromType(exp.Base, name); ok {
					return c.inferStructArguments(info, exp.Arguments, instance, subst, node, index)
				}
			}
		case ArrayType:
			if name, ok := structName(exp.Base); ok && name == "Array" && len(exp.Arguments) > 0 {
				element := act.Element
				if element == nil {
					element = UnknownType{}
				}
				return c.inferTypeArguments(exp.Arguments[0], element, subst, node, index)
			}
		}
	case FunctionType:
		if fn, ok := actual.(FunctionType); ok {
			var result []Diagnostic
			limit := len(exp.Params)
			if len(fn.Params) < limit {
				limit = len(fn.Params)
			}
			for i := 0; i < limit; i++ {
				result = append(result, c.inferTypeArguments(exp.Params[i], fn.Params[i], subst, node, index)...)
			}
			result = append(result, c.inferTypeArguments(exp.Return, fn.Return, subst, node, index)...)
			return result
		}
	case ProcType:
		if proc, ok := actual.(ProcType); ok {
			return c.inferTypeArguments(exp.Result, proc.Result, subst, node, index)
		}
	case FutureType:
		if future, ok := actual.(FutureType); ok {
			return c.inferTypeArguments(exp.Result, future.Result, subst, node, index)
		}
	case RangeType:
		if rng, ok := actual.(RangeType); ok {
			return c.inferTypeArguments(exp.Element, rng.Element, subst, node, index)
		}
	}
	return nil
}

func bindTypeParameter(name string, actual Type, subst map[string]Type, node ast.Node, index int) []Diagnostic {
	if name == "" || actual == nil || isUnknownType(actual) {
		return nil
	}
	if existing, ok := subst[name]; ok {
		if typesEquivalentForSignature(existing, actual) {
			return nil
		}
		if typeAssignable(existing, actual) && typeAssignable(actual, existing) {
			return nil
		}
		argLabel := index + 1
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: type parameter %s inferred as %s but argument %d has type %s", name, typeName(existing), argLabel, typeName(actual)),
			Node:    node,
		}}
	}
	subst[name] = actual
	return nil
}

func (c *Checker) structInfoFromType(base Type, fallback string) (StructType, bool) {
	switch st := base.(type) {
	case StructType:
		return st, true
	case AppliedType:
		return c.structInfoFromType(st.Base, fallback)
	}
	if fallback == "" {
		if name, ok := structName(base); ok {
			fallback = name
		}
	}
	if fallback == "" {
		return StructType{}, false
	}
	return c.lookupStructType(fallback)
}

func (c *Checker) lookupStructType(name string) (StructType, bool) {
	if c.global == nil || name == "" {
		return StructType{}, false
	}
	if decl, ok := c.global.Lookup(name); ok {
		if st, ok := decl.(StructType); ok {
			return st, true
		}
	}
	return StructType{}, false
}

func (c *Checker) inferStructArguments(info StructType, args []Type, actual StructInstanceType, subst map[string]Type, node ast.Node, index int) []Diagnostic {
	if info.StructName != "" && actual.StructName != "" && info.StructName != actual.StructName {
		return nil
	}
	substitution := make(map[string]Type, len(info.TypeParams))
	if len(info.TypeParams) > 0 && len(args) == 0 {
		// Ensure keys exist so substitution returns Unknown for missing args.
		for _, param := range info.TypeParams {
			if param.Name == "" {
				continue
			}
			substitution[param.Name] = UnknownType{}
		}
	}
	for i, param := range info.TypeParams {
		if param.Name == "" {
			continue
		}
		if i < len(args) && args[i] != nil {
			substitution[param.Name] = args[i]
		} else if _, exists := substitution[param.Name]; !exists {
			substitution[param.Name] = UnknownType{}
		}
	}

	apply := func(t Type) Type {
		if len(substitution) == 0 || t == nil {
			return t
		}
		return substituteType(t, substitution)
	}

	var result []Diagnostic
	if len(info.Fields) > 0 && len(actual.Fields) > 0 {
		for name, fieldType := range info.Fields {
			expected := apply(fieldType)
			actualField, ok := actual.Fields[name]
			if !ok || actualField == nil {
				continue
			}
			result = append(result, c.inferTypeArguments(expected, actualField, subst, node, index)...)
		}
	}

	if len(info.Positional) > 0 && len(actual.Positional) > 0 {
		limit := len(info.Positional)
		if len(actual.Positional) < limit {
			limit = len(actual.Positional)
		}
		for i := 0; i < limit; i++ {
			expected := apply(info.Positional[i])
			actualField := actual.Positional[i]
			if actualField == nil {
				continue
			}
			result = append(result, c.inferTypeArguments(expected, actualField, subst, node, index)...)
		}
	}

	return result
}
