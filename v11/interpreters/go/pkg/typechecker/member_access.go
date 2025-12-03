package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter10-go/pkg/ast"
)

const nullableTypeLabel = "<nullable>"

func (c *Checker) checkMemberAccess(env *Environment, expr *ast.MemberAccessExpression) ([]Diagnostic, Type) {
	var diags []Diagnostic
	if expr == nil {
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
		if ty.Fields != nil {
			if fieldType, ok := ty.Fields[memberName]; ok {
				final := c.finalizeMemberAccessType(expr, wrapType, fieldType)
				return diags, final
			}
		}
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		}
		if isErrorStructType(ty) {
			if memberType, ok := c.errorMemberType(memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, memberType)
				return diags, final
			}
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
		if fieldType, ok := ty.Fields[memberName]; ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fieldType)
			return diags, final
		}
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		}
		if isErrorStructInstanceType(ty) {
			if memberType, ok := c.errorMemberType(memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, memberType)
				return diags, final
			}
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
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
			return diags, final
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: array has no member '%s' (import able.collections.array for stdlib helpers)", memberName),
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
			if fnType, ok := c.lookupMethod(objectType, memberName); ok {
				final := c.finalizeMemberAccessType(expr, wrapType, fnType)
				return diags, final
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: string has no member '%s' (import able.text.string for stdlib helpers)", memberName),
				Node:    expr,
			})
			break
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
		if ty.Methods != nil {
			if methodType, ok := ty.Methods[memberName]; ok {
				bound := bindMethodType(methodType)
				final := c.finalizeMemberAccessType(expr, wrapType, bound)
				return diags, final
			}
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
			if iface.Methods != nil {
				if methodType, ok := iface.Methods[memberName]; ok {
					subst := make(map[string]Type, len(iface.TypeParams))
					for i, spec := range iface.TypeParams {
						if i < len(ty.Arguments) && ty.Arguments[i] != nil {
							subst[spec.Name] = ty.Arguments[i]
						}
					}
					inst := substituteFunctionType(methodType, subst)
					inst = bindMethodType(inst)
					final := c.finalizeMemberAccessType(expr, wrapType, inst)
					return diags, final
				}
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: interface '%s' has no method '%s'", iface.InterfaceName, memberName),
				Node:    expr,
			})
			break
		}
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
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
		default:
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: iterator has no member '%s'", memberName),
				Node:    expr,
			})
		}
	case ProcType:
		if positionalAccess {
			diags = append(diags, Diagnostic{
				Message: "typechecker: positional member access not supported on proc handles",
				Node:    expr,
			})
			break
		}
		fnType, procDiags := procMemberFunction(memberName, ty, expr)
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
		fnType, futureDiags := futureMemberFunction(memberName, ty, expr)
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
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			final := c.finalizeMemberAccessType(expr, wrapType, fnType)
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

func (c *Checker) finalizeMemberAccessType(expr *ast.MemberAccessExpression, objectType Type, memberType Type) Type {
	final := memberType
	if expr != nil && expr.Safe && typeCanBeNil(objectType) && !typeCanBeNil(memberType) {
		if fn, ok := memberType.(FunctionType); ok {
			wrapped := fn
			if wrapped.Return == nil {
				wrapped.Return = UnknownType{}
			}
			wrapped.Return = NullableType{Inner: wrapped.Return}
			final = wrapped
		} else if memberType == nil {
			final = NullableType{Inner: UnknownType{}}
		} else {
			final = NullableType{Inner: memberType}
		}
	}
	if expr != nil {
		c.infer.set(expr, final)
	}
	return final
}

func makeValueUnion(success Type) Type {
	procErr := StructType{StructName: "ProcError"}
	members := []Type{success, procErr}
	return UnionLiteralType{Members: members}
}

func stripNilFromUnion(u UnionLiteralType) Type {
	nonNil := make([]Type, 0, len(u.Members))
	for _, member := range u.Members {
		if prim, ok := member.(PrimitiveType); ok && prim.Kind == PrimitiveNil {
			continue
		}
		nonNil = append(nonNil, member)
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return UnionLiteralType{Members: nonNil}
	}
}

func (c *Checker) lookupMethod(object Type, name string) (FunctionType, bool) {
	bestFn, bestScore, found := c.lookupMethodInMethodSets(object, name)
	if implFn, implScore, implFound := c.lookupMethodInImplementations(object, name); implFound {
		if !found || implScore > bestScore {
			return implFn, true
		}
		return bestFn, true
	}
	return bestFn, found
}

func (c *Checker) lookupMethodInMethodSets(object Type, name string) (FunctionType, int, bool) {
	if len(c.methodSets) == 0 {
		return FunctionType{}, -1, false
	}
	var (
		found     bool
		bestScore = -1
		bestFn    FunctionType
	)
	for _, spec := range c.methodSets {
		subst, score, ok := matchMethodTarget(object, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		substitution := cloneTypeMap(subst)
		if substitution == nil {
			substitution = make(map[string]Type)
		}
		if object != nil {
			substitution["Self"] = object
		}
		method, ok := spec.Methods[name]
		derivedFromConstraints := false
		if !ok {
			if inferred, inferredOK := c.methodFromMethodSetConstraints(spec.Obligations, substitution, name); inferredOK {
				method = inferred
				ok = true
				derivedFromConstraints = true
			}
		}
		if !ok {
			continue
		}
		if len(substitution) > 0 {
			method = substituteFunctionType(method, substitution)
		}
		if len(spec.Obligations) > 0 && !derivedFromConstraints {
			obligations := populateObligationSubjects(spec.Obligations, object)
			obligations = substituteObligations(obligations, substitution)
			if len(obligations) > 0 {
				ownerLabel := methodSetOwnerLabel(spec, substitution, name)
				for i := range obligations {
					if obligations[i].Owner == "" || !strings.Contains(obligations[i].Owner, "::") {
						obligations[i].Owner = ownerLabel
					}
				}
				method.Obligations = append(method.Obligations, obligations...)
			}
		}
		if shouldBindSelfParam(method, object) {
			method = bindMethodType(method)
		}
		if !found || score > bestScore {
			bestScore = score
			bestFn = method
			found = true
		}
	}
	return bestFn, bestScore, found
}

func (c *Checker) methodFromMethodSetConstraints(obligations []ConstraintObligation, substitution map[string]Type, methodName string) (FunctionType, bool) {
	if len(obligations) == 0 {
		return FunctionType{}, false
	}
	for _, ob := range obligations {
		if ob.Constraint == nil {
			continue
		}
		constraint := ob.Constraint
		if len(substitution) > 0 {
			constraint = substituteType(constraint, substitution)
		}
		res := c.resolveConstraintInterfaceType(constraint)
		if res.err != "" {
			continue
		}
		if res.iface.Methods == nil {
			continue
		}
		method, ok := res.iface.Methods[methodName]
		if !ok {
			continue
		}
		replacement := cloneTypeMap(substitution)
		if len(res.iface.TypeParams) > 0 {
			if replacement == nil {
				replacement = make(map[string]Type)
			}
			for idx, param := range res.iface.TypeParams {
				var arg Type = UnknownType{}
				if idx < len(res.args) && res.args[idx] != nil {
					arg = res.args[idx]
				}
				replacement[param.Name] = arg
			}
		}
		if len(replacement) > 0 {
			method = substituteFunctionType(method, replacement)
		}
		return method, true
	}
	return FunctionType{}, false
}

func methodSetOwnerLabel(spec MethodSetSpec, substitution map[string]Type, methodName string) string {
	subject, ok := substitution["Self"]
	if !ok || subject == nil || isUnknownType(subject) {
		subject = spec.Target
	}
	label := nonEmpty(typeName(subject))
	return fmt.Sprintf("methods for %s::%s", label, methodName)
}

func (c *Checker) lookupMethodInImplementations(object Type, name string) (FunctionType, int, bool) {
	if len(c.implementations) == 0 {
		return FunctionType{}, -1, false
	}
	var (
		found     bool
		bestScore = -1
		bestFn    FunctionType
	)
	for _, spec := range c.implementations {
		subst, score, ok := matchMethodTarget(object, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		method, ok := spec.Methods[name]
		if !ok {
			continue
		}
		substitution := cloneTypeMap(subst)
		if substitution == nil {
			substitution = make(map[string]Type)
		}
		if object != nil {
			substitution["Self"] = object
		}
		if spec.InterfaceName != "" {
			res := c.interfaceFromName(spec.InterfaceName)
			if res.err == "" && res.iface.InterfaceName != "" {
				extendImplementationSubstitution(substitution, res.iface, spec.InterfaceArgs)
			}
		}
		for _, param := range spec.TypeParams {
			if param.Name == "" {
				continue
			}
			if _, ok := substitution[param.Name]; !ok {
				substitution[param.Name] = UnknownType{}
			}
		}
		if len(substitution) > 0 {
			method = substituteFunctionType(method, substitution)
		}
		if shouldBindSelfParam(method, object) {
			method = bindMethodType(method)
		}
		if !found || score > bestScore {
			bestScore = score
			bestFn = method
			found = true
		}
	}
	return bestFn, bestScore, found
}

func (c *Checker) lookupUfcsInherentMethod(object Type, name string) (FunctionType, bool) {
	if len(c.methodSets) == 0 {
		return FunctionType{}, false
	}
	var (
		found     bool
		bestScore = -1
		bestFn    FunctionType
	)
	for _, spec := range c.methodSets {
		subst, score, ok := matchMethodTarget(object, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		substitution := cloneTypeMap(subst)
		if substitution == nil {
			substitution = make(map[string]Type)
		}
		if object != nil {
			substitution["Self"] = object
		}
		method, ok := spec.Methods[name]
		derivedFromConstraints := false
		if !ok {
			if inferred, inferredOK := c.methodFromMethodSetConstraints(spec.Obligations, substitution, name); inferredOK {
				method = inferred
				ok = true
				derivedFromConstraints = true
			}
		}
		if !ok {
			continue
		}
		if len(substitution) > 0 {
			method = substituteFunctionType(method, substitution)
		}
		if len(spec.Obligations) > 0 && !derivedFromConstraints {
			obligations := populateObligationSubjects(spec.Obligations, object)
			obligations = substituteObligations(obligations, substitution)
			if len(obligations) > 0 {
				ownerLabel := methodSetOwnerLabel(spec, substitution, name)
				for i := range obligations {
					if obligations[i].Owner == "" || !strings.Contains(obligations[i].Owner, "::") {
						obligations[i].Owner = ownerLabel
					}
				}
				method.Obligations = append(method.Obligations, obligations...)
			}
		}
		if !shouldBindSelfParam(method, object) {
			continue
		}
		method = bindMethodType(method)
		if !found || score > bestScore {
			bestScore = score
			bestFn = method
			found = true
		}
	}
	return bestFn, found
}

func isErrorStructType(ty StructType) bool {
	return ty.StructName == "Error"
}

func isErrorStructInstanceType(ty StructInstanceType) bool {
	return ty.StructName == "Error"
}

func (c *Checker) errorMemberType(memberName string) (Type, bool) {
	switch memberName {
	case "value":
		return UnknownType{}, true
	case "message":
		return FunctionType{
			Params: nil,
			Return: PrimitiveType{Kind: PrimitiveString},
		}, true
	case "cause":
		var inner Type = StructType{StructName: "Error"}
		if c != nil {
			inner = c.lookupErrorType()
		}
		return FunctionType{
			Params: nil,
			Return: NullableType{Inner: inner},
		}, true
	default:
		return nil, false
	}
}

func extendImplementationSubstitution(subst map[string]Type, iface InterfaceType, args []Type) {
	if subst == nil {
		return
	}
	if len(iface.TypeParams) == 0 {
		return
	}
	for idx, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		var arg Type = UnknownType{}
		if idx < len(args) && args[idx] != nil {
			arg = substituteType(args[idx], subst)
		}
		subst[param.Name] = arg
	}
}

func matchMethodTarget(object Type, target Type, params []GenericParamSpec) (map[string]Type, int, bool) {
	if objPrim, ok := object.(PrimitiveType); ok {
		if targetPrim, ok := target.(PrimitiveType); ok && targetPrim.Kind == objPrim.Kind {
			return nil, 0, true
		}
	}
	if objInt, ok := object.(IntegerType); ok {
		if targetInt, ok := target.(IntegerType); ok && targetInt.Suffix == objInt.Suffix {
			return nil, 0, true
		}
	}
	if objFloat, ok := object.(FloatType); ok {
		if targetFloat, ok := target.(FloatType); ok && targetFloat.Suffix == objFloat.Suffix {
			return nil, 0, true
		}
	}
	if targetUnion, ok := target.(UnionLiteralType); ok {
		return matchUnionLiteralTarget(object, targetUnion, params)
	}
	objInfo, ok := structInfoFromType(object)
	if !ok {
		return nil, 0, false
	}
	targetInfo, ok := structInfoFromType(target)
	if !ok {
		return nil, 0, false
	}
	if targetInfo.name == "" || objInfo.name == "" || targetInfo.name != objInfo.name || targetInfo.isUnion != objInfo.isUnion || targetInfo.isNullable != objInfo.isNullable {
		return nil, 0, false
	}
	subst := make(map[string]Type)
	score := 0
	for idx, targetArg := range targetInfo.args {
		var objArg Type
		if idx < len(objInfo.args) {
			objArg = objInfo.args[idx]
		}
		ok, argScore := matchTypeArgument(objArg, targetArg, subst)
		if !ok {
			return nil, 0, false
		}
		score += argScore
	}
	return finalizeMatchResult(subst, params, score)
}

func matchTypeArgument(actual Type, pattern Type, subst map[string]Type) (bool, int) {
	if pattern == nil || isUnknownType(pattern) {
		return true, 0
	}
	if actual == nil || isUnknownType(actual) {
		return true, 0
	}
	switch p := pattern.(type) {
	case TypeParameterType:
		if p.ParameterName == "" {
			return true, 0
		}
		if existing, ok := subst[p.ParameterName]; ok {
			if !typesEquivalentForSignature(existing, actual) {
				return false, 0
			}
			return true, 0
		}
		subst[p.ParameterName] = actual
		if actual == nil || isUnknownType(actual) {
			return true, 0
		}
		return true, 1
	case NullableType:
		av, ok := actual.(NullableType)
		if !ok {
			return false, 0
		}
		return matchTypeArgument(av.Inner, p.Inner, subst)
	case AppliedType:
		av, ok := actual.(AppliedType)
		if !ok {
			return false, 0
		}
		if !nominalBasesCompatible(av.Base, p.Base) {
			return false, 0
		}
		if len(p.Arguments) != len(av.Arguments) {
			return false, 0
		}
		score := 0
		for i := range p.Arguments {
			ok, s := matchTypeArgument(av.Arguments[i], p.Arguments[i], subst)
			if !ok {
				return false, 0
			}
			score += s
		}
		return true, score
	case UnionType:
		name, ok := unionName(actual)
		if !ok || name != p.UnionName {
			return false, 0
		}
		return true, 0
	case StructType:
		name, ok := structName(actual)
		if !ok || name != p.StructName {
			return false, 0
		}
		return true, 0
	case StructInstanceType:
		name, ok := structName(actual)
		if !ok || name != p.StructName {
			return false, 0
		}
		return true, 0
	case UnionLiteralType:
		if actualUnion, ok := actual.(UnionLiteralType); ok {
			if len(actualUnion.Members) != len(p.Members) {
				return false, 0
			}
			score := 0
			for i := range p.Members {
				ok, s := matchTypeArgument(actualUnion.Members[i], p.Members[i], subst)
				if !ok {
					return false, 0
				}
				score += s
			}
			return true, score
		}
	}
	if typesEquivalentForSignature(actual, pattern) {
		return true, 0
	}
	return false, 0
}

func matchUnionLiteralTarget(object Type, target UnionLiteralType, params []GenericParamSpec) (map[string]Type, int, bool) {
	if objUnion, ok := object.(UnionLiteralType); ok {
		if len(objUnion.Members) != len(target.Members) {
			return nil, 0, false
		}
		subst := make(map[string]Type)
		score := 0
		for i := range target.Members {
			ok, s := matchTypeArgument(objUnion.Members[i], target.Members[i], subst)
			if !ok {
				return nil, 0, false
			}
			score += s
		}
		return finalizeMatchResult(subst, params, score)
	}
	for _, member := range target.Members {
		if member == nil {
			continue
		}
		subst := make(map[string]Type)
		ok, s := matchTypeArgument(object, member, subst)
		if !ok {
			continue
		}
		return finalizeMatchResult(subst, params, s)
	}
	return nil, 0, false
}

func finalizeMatchResult(subst map[string]Type, params []GenericParamSpec, score int) (map[string]Type, int, bool) {
	if len(subst) == 0 {
		subst = nil
	}
	subst = ensureTypeParams(subst, params)
	if subst == nil || len(subst) == 0 {
		return nil, score, true
	}
	return subst, score, true
}

func ensureTypeParams(subst map[string]Type, params []GenericParamSpec) map[string]Type {
	if len(params) == 0 {
		return subst
	}
	if subst == nil {
		subst = make(map[string]Type)
	}
	for _, param := range params {
		if param.Name == "" {
			continue
		}
		if _, ok := subst[param.Name]; !ok {
			subst[param.Name] = UnknownType{}
		}
	}
	if len(subst) == 0 {
		return nil
	}
	return subst
}

func nominalBasesCompatible(actual Type, pattern Type) bool {
	if pattern == nil {
		return true
	}
	if actual == nil {
		return false
	}
	switch pb := pattern.(type) {
	case StructType:
		name, ok := structName(actual)
		return ok && name == pb.StructName
	case StructInstanceType:
		name, ok := structName(actual)
		return ok && name == pb.StructName
	case UnionType:
		name, ok := unionName(actual)
		return ok && name == pb.UnionName
	case InterfaceType:
		if iface, ok := actual.(InterfaceType); ok {
			return iface.InterfaceName == pb.InterfaceName
		}
	default:
		return typesEquivalentForSignature(actual, pattern)
	}
	return false
}

type structInfo struct {
	name       string
	args       []Type
	isUnion    bool
	isNullable bool
}

func structInfoFromType(t Type) (structInfo, bool) {
	switch v := t.(type) {
	case StructType:
		return structInfo{name: v.StructName}, v.StructName != ""
	case StructInstanceType:
		return structInfo{name: v.StructName, args: v.TypeArgs}, v.StructName != ""
	case UnionType:
		return structInfo{name: v.UnionName, isUnion: true}, v.UnionName != ""
	case NullableType:
		return structInfo{name: nullableTypeLabel, args: []Type{v.Inner}, isNullable: true}, true
	case AppliedType:
		if baseName, ok := structName(v.Base); ok {
			return structInfo{name: baseName, args: v.Arguments}, true
		}
		if unionName, ok := unionName(v.Base); ok {
			return structInfo{name: unionName, args: v.Arguments, isUnion: true}, true
		}
	case ArrayType:
		return structInfo{name: "Array", args: []Type{v.Element}}, true
	}
	return structInfo{}, false
}

func cloneTypeMap(src map[string]Type) map[string]Type {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]Type, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (c *Checker) lookupTypeParamMethod(paramName, methodName string) (FunctionType, bool) {
	constraints := c.typeParamConstraints(paramName)
	for _, constraint := range constraints {
		if fnType, ok := c.methodFromConstraint(paramName, constraint, methodName); ok {
			return bindMethodType(fnType), true
		}
	}
	return FunctionType{}, false
}

func (c *Checker) methodFromConstraint(paramName string, constraint Type, methodName string) (FunctionType, bool) {
	res := c.resolveConstraintInterfaceType(constraint)
	if res.err != "" {
		return FunctionType{}, false
	}
	if res.iface.Methods == nil {
		return FunctionType{}, false
	}
	method, ok := res.iface.Methods[methodName]
	if !ok {
		return FunctionType{}, false
	}
	subst := make(map[string]Type)
	subst["Self"] = TypeParameterType{ParameterName: paramName}
	for idx, spec := range res.iface.TypeParams {
		var arg Type = UnknownType{}
		if idx < len(res.args) && res.args[idx] != nil {
			arg = res.args[idx]
		}
		subst[spec.Name] = arg
	}
	if len(subst) > 0 {
		method = substituteFunctionType(method, subst)
	}
	return method, true
}

func bindMethodType(method FunctionType) FunctionType {
	if len(method.Params) == 0 {
		return method
	}
	params := make([]Type, len(method.Params)-1)
	copy(params, method.Params[1:])
	return FunctionType{
		Params:      params,
		Return:      method.Return,
		TypeParams:  method.TypeParams,
		Where:       method.Where,
		Obligations: method.Obligations,
	}
}

func shouldBindSelfParam(method FunctionType, subject Type) bool {
	if len(method.Params) == 0 {
		return false
	}
	first := method.Params[0]
	if typesEquivalentForSignature(first, subject) {
		return true
	}
	switch v := first.(type) {
	case TypeParameterType:
		return v.ParameterName == "Self"
	default:
		if infoParam, ok := structInfoFromType(first); ok {
			if infoSubject, ok := structInfoFromType(subject); ok && infoParam.name != "" && infoParam.name == infoSubject.name {
				return true
			}
		}
	}
	return false
}
