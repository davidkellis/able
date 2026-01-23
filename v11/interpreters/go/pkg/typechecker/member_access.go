package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

const nullableTypeLabel = "<nullable>"

func (c *Checker) checkMemberAccess(env *Environment, expr *ast.MemberAccessExpression) ([]Diagnostic, Type) {
	return c.checkMemberAccessWithOptions(env, expr, false)
}

func (c *Checker) checkMemberAccessWithOptions(env *Environment, expr *ast.MemberAccessExpression, preferMethods bool) ([]Diagnostic, Type) {
	var diags []Diagnostic
	if expr == nil {
		return nil, UnknownType{}
	}
	if expressionContainsPlaceholder(expr.Object) {
		c.infer.set(expr, UnknownType{})
		return nil, UnknownType{}
	}
	objectDiags, objectType := c.checkExpression(env, expr.Object)
	diags = append(diags, objectDiags...)
	wrapType := objectType

	var (
		memberName       string
		positionalIndex  int
		positionalAccess bool
	)
	switch mem := expr.Member.(type) {
	case *ast.Identifier:
		memberName = mem.Name
	case *ast.IntegerLiteral:
		if mem.Value == nil || !mem.Value.IsInt64() {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access requires integer literal",
				Node:    expr.Member,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		idx := mem.Value.Int64()
		if idx < 0 {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access requires non-negative index",
				Node:    expr.Member,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		positionalIndex = int(idx)
		positionalAccess = true
	default:
		diags = append(diags, Diagnostic{
			Message: "typechecker: member access requires identifier or positional index",
			Node:    expr.Member,
		})
		c.infer.set(expr, UnknownType{})
		return diags, UnknownType{}
	}

	if expr.Safe {
		switch ty := objectType.(type) {
		case NullableType:
			objectType = ty.Inner
		case UnionLiteralType:
			if stripped := stripNilFromUnion(ty); stripped != nil {
				objectType = stripped
			}
		}
	}

	receiverScopeNames := receiverNamesForType(objectType)
	if alias, ok := objectType.(AliasType); ok {
		if target, _ := instantiateAlias(alias, nil); target != nil {
			objectType = target
		}
	}

	switch ty := objectType.(type) {
	case StructType:
		if positionalAccess {
			if positionalIndex < len(ty.Positional) {
				result := ty.Positional[positionalIndex]
				final := c.finalizeMemberAccessType(expr, wrapType, result)
				return diags, final
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: struct '%s' has no positional member %d", ty.StructName, positionalIndex),
				Node:    expr,
			})
			break
		}
		var (
			fieldType Type
			hasField  bool
		)
		if ty.Fields != nil {
			if candidate, ok := ty.Fields[memberName]; ok {
				fieldType = candidate
				hasField = true
				if !preferMethods || isCallableType(candidate) {
					final := c.finalizeMemberAccessType(expr, wrapType, candidate)
					return diags, final
				}
			}
		}
		var (
			candidates  []FunctionType
			methodFound bool
		)
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, true, true); ok {
			candidates = append(candidates, fnType)
			methodFound = true
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if isErrorStructType(ty) {
			if memberType, ok := c.errorMemberType(memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, memberType)
				return diags, final
			}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
			var done bool
			var final Type
			candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
			if done {
				return diags, final
			}
		}
		if len(candidates) > 1 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if len(candidates) == 1 {
			final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
			return diags, final
		}
		if hasField {
			final := c.finalizeMemberAccessType(expr, wrapType, fieldType)
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: struct '%s' has no member '%s'", ty.StructName, memberName),
			Node:    expr,
		})
	case StructInstanceType:
		if positionalAccess {
			if positionalIndex < len(ty.Positional) {
				result := ty.Positional[positionalIndex]
				final := c.finalizeMemberAccessType(expr, wrapType, result)
				return diags, final
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: struct '%s' has no positional member %d", ty.StructName, positionalIndex),
				Node:    expr,
			})
			break
		}
		var (
			fieldType Type
			hasField  bool
		)
		if candidate, ok := ty.Fields[memberName]; ok {
			fieldType = candidate
			hasField = true
			if !preferMethods || isCallableType(candidate) {
				final := c.finalizeMemberAccessType(expr, wrapType, candidate)
				return diags, final
			}
		}
		var (
			candidates  []FunctionType
			methodFound bool
		)
		allowMethodSets := true
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
			candidates = append(candidates, fnType)
			methodFound = true
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if isErrorStructInstanceType(ty) {
			if memberType, ok := c.errorMemberType(memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, memberType)
				return diags, final
			}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
			var done bool
			var final Type
			candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
			if done {
				return diags, final
			}
		}
		if len(candidates) > 1 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if len(candidates) == 1 {
			final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
			return diags, final
		}
		if hasField {
			final := c.finalizeMemberAccessType(expr, wrapType, fieldType)
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: struct '%s' has no member '%s'", ty.StructName, memberName),
			Node:    expr,
		})
	case ArrayType:
		if positionalAccess {
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		var (
			candidates  []FunctionType
			methodFound bool
		)
		allowMethodSets := true
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
			candidates = append(candidates, fnType)
			methodFound = true
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
			var done bool
			var final Type
			candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
			if done {
				return diags, final
			}
		}
		if len(candidates) > 1 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if len(candidates) == 1 {
			final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: array has no member '%s' (import able.collections.array for stdlib helpers)", memberName),
			Node:    expr,
		})
	case IntegerType, FloatType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: positional member access not supported on type %s", typeName(objectType)),
				Node:    expr,
			})
			break
		}
		var (
			candidates  []FunctionType
			methodFound bool
		)
		allowMethodSets := true
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
			candidates = append(candidates, fnType)
			methodFound = true
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
			var done bool
			var final Type
			candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
			if done {
				return diags, final
			}
		}
		if len(candidates) > 1 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if len(candidates) == 1 {
			final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot access member '%s' on type %s", memberName, typeName(objectType)),
			Node:    expr,
		})
	case PrimitiveType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: positional member access not supported on type %s", typeName(objectType)),
				Node:    expr,
			})
			break
		}
		if ty.Kind == PrimitiveString {
			var (
				candidates  []FunctionType
				methodFound bool
			)
			if fnType, ok, detail := c.lookupMethod(objectType, memberName, true, false); ok {
				candidates = append(candidates, fnType)
				methodFound = true
			} else if detail != "" {
				diags = append(diags, Diagnostic{
					Message: "typechecker: " + detail,
					Node:    expr,
				})
				c.infer.set(expr, UnknownType{})
				return diags, UnknownType{}
			}
			if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
				var done bool
				var final Type
				candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
				if done {
					return diags, final
				}
			}
			if len(candidates) > 1 {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
					Node:    expr,
				})
				c.infer.set(expr, UnknownType{})
				return diags, UnknownType{}
			}
			if len(candidates) == 1 {
				final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
				return diags, final
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: string has no member '%s' (import able.text.string for stdlib helpers)", memberName),
				Node:    expr,
			})
			break
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, ufcsType)
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot access member '%s' on type %s", memberName, typeName(objectType)),
			Node:    expr,
		})
	case InterfaceType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on interfaces",
				Node:    expr,
			})
			break
		}
		if ty.InterfaceName == "Error" {
			if memberType, ok := c.errorMemberType(memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, memberType)
				return diags, final
			}
		}
		if ty.Methods != nil {
			if methodType, ok := c.interfaceMemberFunction(ty, nil, ty, memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, methodType)
				return diags, final
			}
		}
		if res := c.interfaceFromName(ty.InterfaceName); res.err == "" {
			if methodType, ok := c.interfaceMemberFunction(res.iface, res.args, ty, memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, methodType)
				return diags, final
			}
		}
		allowMethodSets := allowMethodSetsForMember(env, memberName, receiverScopeNames)
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: interface '%s' has no method '%s'", ty.InterfaceName, memberName),
			Node:    expr,
		})
	case AppliedType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on this type",
				Node:    expr,
			})
			break
		}
		if iface, ok := ty.Base.(InterfaceType); ok {
			if iface.InterfaceName == "Error" {
				if memberType, ok := c.errorMemberType(memberName); ok {
					final := c.finalizeMemberAccessType(expr, wrapType, memberType)
					return diags, final
				}
			}
			if methodType, ok := c.interfaceMemberFunction(iface, ty.Arguments, ty, memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, methodType)
				return diags, final
			}
			if res := c.interfaceFromName(iface.InterfaceName); res.err == "" {
				args := ty.Arguments
				if len(args) == 0 {
					args = res.args
				}
				if methodType, ok := c.interfaceMemberFunction(res.iface, args, ty, memberName); ok {
					final := c.finalizeMemberAccessType(expr, wrapType, methodType)
					return diags, final
				}
			}
			allowMethodSets := allowMethodSetsForMember(env, memberName, receiverScopeNames)
			if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, fnType)
				return diags, final
			} else if detail != "" {
				diags = append(diags, Diagnostic{
					Message: "typechecker: " + detail,
					Node:    expr,
				})
				c.infer.set(expr, UnknownType{})
				return diags, UnknownType{}
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: interface '%s' has no method '%s'", iface.InterfaceName, memberName),
				Node:    expr,
			})
			break
		}
		var (
			fieldType Type
			hasField  bool
		)
		if baseStruct, ok := ty.Base.(StructType); ok {
			subst := make(map[string]Type, len(baseStruct.TypeParams))
			for i, param := range baseStruct.TypeParams {
				if param.Name == "" {
					continue
				}
				if i < len(ty.Arguments) && ty.Arguments[i] != nil {
					subst[param.Name] = ty.Arguments[i]
				}
			}
			if positionalAccess {
				if positionalIndex < len(baseStruct.Positional) {
					fieldType := baseStruct.Positional[positionalIndex]
					if len(subst) > 0 {
						fieldType = substituteType(fieldType, subst)
					}
					final := c.finalizeMemberAccessType(expr, wrapType, fieldType)
					return diags, final
				}
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: struct '%s' has no positional member %d", baseStruct.StructName, positionalIndex),
					Node:    expr,
				})
				break
			}
			if baseStruct.Fields != nil {
				if candidate, ok := baseStruct.Fields[memberName]; ok {
					if len(subst) > 0 {
						candidate = substituteType(candidate, subst)
					}
					fieldType = candidate
					hasField = true
					if !preferMethods || isCallableType(candidate) {
						final := c.finalizeMemberAccessType(expr, wrapType, candidate)
						return diags, final
					}
				}
			}
		}
		var (
			candidates  []FunctionType
			methodFound bool
		)
		allowMethodSets := allowMethodSetsForMember(env, memberName, receiverScopeNames)
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
			candidates = append(candidates, fnType)
			methodFound = true
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
			var done bool
			var final Type
			candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
			if done {
				return diags, final
			}
		}
		if len(candidates) > 1 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if len(candidates) == 1 {
			final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
			return diags, final
		}
		if hasField {
			final := c.finalizeMemberAccessType(expr, wrapType, fieldType)
			return diags, final
		}
	case IteratorType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on iterators",
				Node:    expr,
			})
			break
		}
		switch memberName {
		case "next":
			elem := ty.Element
			if elem == nil {
				elem = UnknownType{}
			}
			result := UnionLiteralType{
				Members: []Type{
					elem,
					StructType{StructName: "IteratorEnd"},
				},
			}
			fn := FunctionType{
				Params: nil,
				Return: result,
			}
			final := c.finalizeMemberAccessType(expr, wrapType, fn)
			return diags, final
		case "close":
			fn := FunctionType{
				Params: nil,
				Return: PrimitiveType{Kind: PrimitiveNil},
			}
			final := c.finalizeMemberAccessType(expr, wrapType, fn)
			return diags, final
		}
		if res := c.interfaceFromName("Iterator"); res.err == "" {
			elem := ty.Element
			if elem == nil {
				elem = UnknownType{}
			}
			if methodType, ok := c.interfaceMemberFunction(res.iface, []Type{elem}, ty, memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, methodType)
				return diags, final
			}
		}
		allowMethodSets := allowMethodSetsForMember(env, memberName, receiverScopeNames)
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, true); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, ufcsType)
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: iterator has no member '%s'", memberName),
			Node:    expr,
		})
	case ProcType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on proc handles",
				Node:    expr,
			})
			break
		}
		fnType, procDiags := c.procMemberFunction(memberName, ty, expr)
		diags = append(diags, procDiags...)
		final := c.finalizeMemberAccessType(expr, wrapType, fnType)
		return diags, final
	case FutureType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on futures",
				Node:    expr,
			})
			break
		}
		fnType, futureDiags := c.futureMemberFunction(memberName, ty, expr)
		diags = append(diags, futureDiags...)
		final := c.finalizeMemberAccessType(expr, wrapType, fnType)
		return diags, final
	case PackageType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on packages",
				Node:    expr,
			})
			break
		}
		if ty.Symbols != nil {
			if symbolType, ok := ty.Symbols[memberName]; ok && symbolType != nil {
				final := c.finalizeMemberAccessType(expr, wrapType, symbolType)
				return diags, final
			}
		}
		if ty.PrivateSymbols != nil {
			if _, ok := ty.PrivateSymbols[memberName]; ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: package '%s' has no symbol '%s'", ty.Package, memberName),
					Node:    expr,
				})
				break
			}
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: package '%s' has no symbol '%s'", ty.Package, memberName),
			Node:    expr,
		})
	case ImplementationNamespaceType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on implementations",
				Node:    expr,
			})
			break
		}
		if fnType, detail, ok := c.lookupImplementationNamespaceMethod(ty, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: implementation has no member '%s'", memberName),
			Node:    expr,
		})
	case UnknownType:
		c.infer.set(expr, UnknownType{})
		return diags, UnknownType{}
	case TypeParameterType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on type parameters",
				Node:    expr,
			})
			break
		}
		if fnType, ok := c.lookupTypeParamMethod(ty.ParameterName, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot access member '%s' on type parameter %s", memberName, ty.ParameterName),
			Node:    expr,
		})
	default:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: cannot access positional member %d on type %s", positionalIndex, typeName(objectType)),
				Node:    expr,
			})
			break
		}
		var candidates []FunctionType
		methodFound := false
		allowMethodSets := allowMethodSetsForMember(env, memberName, receiverScopeNames)
		if fnType, ok, detail := c.lookupMethod(objectType, memberName, allowMethodSets, false); ok {
			candidates = append(candidates, fnType)
			methodFound = true
		} else if detail != "" {
			diags = append(diags, Diagnostic{
				Message: "typechecker: " + detail,
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if ufcsType, ok := c.lookupUfcsFreeFunction(env, objectType, memberName); ok && !methodFound {
			var done bool
			var final Type
			candidates, final, done = c.appendUfcsCandidate(candidates, ufcsType, expr, wrapType)
			if done {
				return diags, final
			}
		}
		if len(candidates) > 1 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: ambiguous method resolution for '%s'", memberName),
				Node:    expr,
			})
			c.infer.set(expr, UnknownType{})
			return diags, UnknownType{}
		}
		if len(candidates) == 1 {
			final := c.finalizeMemberAccessType(expr, wrapType, candidates[0])
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot access member '%s' on type %s", memberName, typeName(objectType)),
			Node:    expr,
		})
	}

	c.infer.set(expr, UnknownType{})
	return diags, UnknownType{}
}

func receiverNamesForType(t Type) []string {
	var names []string
	if alias, ok := t.(AliasType); ok && alias.AliasName != "" {
		names = append(names, alias.AliasName)
	}
	if name, ok := structName(t); ok && name != "" {
		for _, existing := range names {
			if existing == name {
				return names
			}
		}
		names = append(names, name)
	}
	return names
}

func allowMethodSetsForMember(env *Environment, memberName string, receiverNames []string) bool {
	if env == nil {
		return false
	}
	if _, ok := env.Lookup(memberName); ok {
		return true
	}
	for _, name := range receiverNames {
		if name == "" {
			continue
		}
		if _, ok := env.Lookup(name); ok {
			return true
		}
	}
	return false
}

func (c *Checker) interfaceMemberFunction(iface InterfaceType, args []Type, self Type, name string) (FunctionType, bool) {
	if iface.InterfaceName == "" || iface.Methods == nil {
		return FunctionType{}, false
	}
	methodType, ok := iface.Methods[name]
	if !ok {
		return FunctionType{}, false
	}
	subst := map[string]Type{"Self": self}
	if len(iface.TypeParams) > 0 {
		for i, spec := range iface.TypeParams {
			if spec.Name == "" {
				continue
			}
			arg := Type(UnknownType{})
			if i < len(args) && args[i] != nil {
				arg = args[i]
			}
			subst[spec.Name] = arg
		}
	}
	c.applySelfPatternConstructorSubstitution(subst, iface, self)
	methodType = substituteFunctionType(methodType, subst)
	methodType = bindMethodType(methodType)
	return methodType, true
}
