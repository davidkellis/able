package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *Checker) validateImplementations() []Diagnostic {
	if len(c.implementations) == 0 || c.global == nil {
		return nil
	}

	var diags []Diagnostic
	for _, spec := range c.implementations {
		if spec.InterfaceName == "" {
			continue
		}

		decl, ok := c.global.Lookup(spec.InterfaceName)
		if !ok {
			continue
		}
		iface, ok := decl.(InterfaceType)
		if !ok {
			continue
		}
		if len(iface.Methods) == 0 {
			continue
		}

		subst := buildImplementationSubstitution(spec, iface)
		label := fmt.Sprintf("impl %s for %s", spec.InterfaceName, describeImplTarget(spec.Target))

		for name, ifaceMethod := range iface.Methods {
			expected := substituteFunctionType(ifaceMethod, subst)
			actual, ok := spec.Methods[name]
			if !ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s missing method '%s'", label, name),
					Node:    implementationMethodNode(spec.Definition, name),
				})
				continue
			}
			diags = append(diags, compareImplementationMethodSignature(label, spec, name, expected, actual)...)
		}
	}
	return diags
}

func buildImplementationSubstitution(spec ImplementationSpec, iface InterfaceType) map[string]Type {
	subst := make(map[string]Type, len(iface.TypeParams)+1)
	if spec.Target != nil && !isUnknownType(spec.Target) {
		subst["Self"] = spec.Target
	} else {
		subst["Self"] = UnknownType{}
	}
	for idx, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		var replacement Type = TypeParameterType{ParameterName: param.Name}
		if idx < len(spec.InterfaceArgs) && spec.InterfaceArgs[idx] != nil && !isUnknownType(spec.InterfaceArgs[idx]) {
			replacement = spec.InterfaceArgs[idx]
		}
		subst[param.Name] = replacement
	}
	return subst
}

func compareImplementationMethodSignature(label string, spec ImplementationSpec, methodName string, expected, actual FunctionType) []Diagnostic {
	var diags []Diagnostic
	node := implementationMethodNode(spec.Definition, methodName)

	expectedParams := expected.TypeParams
	actualParams := actual.TypeParams
	if len(spec.TypeParams) > 0 && len(actualParams) > 0 {
		implParams := collectGenericParamNameSet(spec.TypeParams)
		filtered := make([]GenericParamSpec, 0, len(actualParams))
		for _, param := range actualParams {
			if param.Name == "" {
				filtered = append(filtered, param)
				continue
			}
			if _, ok := implParams[param.Name]; ok {
				continue
			}
			filtered = append(filtered, param)
		}
		actualParams = filtered
	}

	if len(expectedParams) != len(actualParams) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' expects %d generic parameter(s), got %d",
				label, methodName, len(expectedParams), len(actualParams),
			),
			Node: node,
		})
	}

	if len(expected.Params) != len(actual.Params) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' expects %d parameter(s), got %d",
				label, methodName, len(expected.Params), len(actual.Params),
			),
			Node: node,
		})
	} else {
		for idx := range expected.Params {
			if !typesEquivalentForSignature(expected.Params[idx], actual.Params[idx]) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf(
						"typechecker: %s method '%s' parameter %d expected %s, got %s",
						label, methodName, idx+1,
						formatTypeForMessage(expected.Params[idx]),
						formatTypeForMessage(actual.Params[idx]),
					),
					Node: node,
				})
			}
		}
	}

	if !typesEquivalentForSignature(expected.Return, actual.Return) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' return type expected %s, got %s",
				label, methodName,
				formatTypeForMessage(expected.Return),
				formatTypeForMessage(actual.Return),
			),
			Node: node,
		})
	}

	return diags
}

func implementationMethodNode(def *ast.ImplementationDefinition, name string) ast.Node {
	if def == nil {
		return nil
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			continue
		}
		if fn.ID.Name == name {
			return fn
		}
	}
	return def
}

func describeImplTarget(t Type) string {
	if t == nil || isUnknownType(t) {
		return "<unknown>"
	}
	return formatType(t)
}

func formatTypeForMessage(t Type) string {
	return formatType(t)
}

func typesEquivalentForSignature(a, b Type) bool {
	if a == nil || b == nil {
		return isUnknownType(a) || isUnknownType(b)
	}
	if isUnknownType(a) || isUnknownType(b) {
		return true
	}

	switch av := a.(type) {
	case TypeParameterType:
		_, ok := b.(TypeParameterType)
		return ok
	case StructType:
		switch bv := b.(type) {
		case StructType:
			return av.StructName == bv.StructName
		case StructInstanceType:
			return av.StructName == bv.StructName
		}
	case StructInstanceType:
		switch bv := b.(type) {
		case StructType:
			return av.StructName == bv.StructName
		case StructInstanceType:
			return av.StructName == bv.StructName
		case AppliedType:
			return typesEquivalentForSignature(av, bv.Base)
		}
	case AppliedType:
		switch bv := b.(type) {
		case AppliedType:
			if !typesEquivalentForSignature(av.Base, bv.Base) {
				return false
			}
			if len(av.Arguments) != len(bv.Arguments) {
				return false
			}
			for i := range av.Arguments {
				if !typesEquivalentForSignature(av.Arguments[i], bv.Arguments[i]) {
					return false
				}
			}
			return true
		case StructType, StructInstanceType:
			return typesEquivalentForSignature(av.Base, bv)
		}
		return false
	case ArrayType:
		if bv, ok := b.(ArrayType); ok {
			return typesEquivalentForSignature(av.Element, bv.Element)
		}
	case NullableType:
		if bv, ok := b.(NullableType); ok {
			return typesEquivalentForSignature(av.Inner, bv.Inner)
		}
	case RangeType:
		if bv, ok := b.(RangeType); ok {
			return typesEquivalentForSignature(av.Element, bv.Element)
		}
	case UnionLiteralType:
		if bv, ok := b.(UnionLiteralType); ok {
			if len(av.Members) != len(bv.Members) {
				return false
			}
			for i := range av.Members {
				if !typesEquivalentForSignature(av.Members[i], bv.Members[i]) {
					return false
				}
			}
			return true
		}
	case FunctionType:
		bv, ok := b.(FunctionType)
		if !ok {
			return false
		}
		if len(av.Params) != len(bv.Params) {
			return false
		}
		for i := range av.Params {
			if !typesEquivalentForSignature(av.Params[i], bv.Params[i]) {
				return false
			}
		}
		return typesEquivalentForSignature(av.Return, bv.Return)
	case ProcType:
		if bv, ok := b.(ProcType); ok {
			return typesEquivalentForSignature(av.Result, bv.Result)
		}
	case FutureType:
		if bv, ok := b.(FutureType); ok {
			return typesEquivalentForSignature(av.Result, bv.Result)
		}
	}

	return a.Name() == b.Name()
}
