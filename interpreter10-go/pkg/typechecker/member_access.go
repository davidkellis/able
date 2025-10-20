package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkMemberAccess(env *Environment, expr *ast.MemberAccessExpression) ([]Diagnostic, Type) {
	var diags []Diagnostic
	if expr == nil {
		return nil, UnknownType{}
	}
	objectDiags, objectType := c.checkExpression(env, expr.Object)
	diags = append(diags, objectDiags...)

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

	switch ty := objectType.(type) {
	case StructType:
		if positionalAccess {
			if positionalIndex < len(ty.Positional) {
				result := ty.Positional[positionalIndex]
				c.infer.set(expr, result)
				return diags, result
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: struct '%s' has no positional member %d", ty.StructName, positionalIndex),
				Node:    expr,
			})
			break
		}
		if ty.Fields != nil {
			if fieldType, ok := ty.Fields[memberName]; ok {
				c.infer.set(expr, fieldType)
				return diags, fieldType
			}
		}
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			c.infer.set(expr, fnType)
			return diags, fnType
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: struct '%s' has no member '%s'", ty.StructName, memberName),
			Node:    expr,
		})
	case StructInstanceType:
		if positionalAccess {
			if positionalIndex < len(ty.Positional) {
				result := ty.Positional[positionalIndex]
				c.infer.set(expr, result)
				return diags, result
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: struct '%s' has no positional member %d", ty.StructName, positionalIndex),
				Node:    expr,
			})
			break
		}
		if fieldType, ok := ty.Fields[memberName]; ok {
			c.infer.set(expr, fieldType)
			return diags, fieldType
		}
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			c.infer.set(expr, fnType)
			return diags, fnType
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: struct '%s' has no member '%s'", ty.StructName, memberName),
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
				c.infer.set(expr, bound)
				return diags, bound
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
					c.infer.set(expr, inst)
					return diags, inst
				}
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: interface '%s' has no method '%s'", iface.InterfaceName, memberName),
				Node:    expr,
			})
			break
		}
		if fnType, ok := c.lookupMethod(objectType, memberName); ok {
			c.infer.set(expr, fnType)
			return diags, fnType
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
		c.infer.set(expr, fnType)
		return diags, fnType
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
		c.infer.set(expr, fnType)
		return diags, fnType
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
			c.infer.set(expr, fnType)
			return diags, fnType
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
			c.infer.set(expr, fnType)
			return diags, fnType
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot access member '%s' on type %s", memberName, typeName(objectType)),
			Node:    expr,
		})
	}

	c.infer.set(expr, UnknownType{})
	return diags, UnknownType{}
}

func procMemberFunction(name string, proc ProcType, node ast.Node) (Type, []Diagnostic) {
	var diags []Diagnostic
	switch name {
	case "status":
		return FunctionType{
			Params: nil,
			Return: StructType{StructName: "ProcStatus"},
		}, diags
	case "value":
		return FunctionType{
			Params: nil,
			Return: makeValueUnion(proc.Result),
		}, diags
	case "cancel":
		return FunctionType{
			Params: nil,
			Return: PrimitiveType{Kind: PrimitiveNil},
		}, diags
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: proc handle has no member '%s'", name),
			Node:    node,
		})
		return UnknownType{}, diags
	}
}

func futureMemberFunction(name string, future FutureType, node ast.Node) (Type, []Diagnostic) {
	var diags []Diagnostic
	switch name {
	case "status":
		return FunctionType{
			Params: nil,
			Return: StructType{StructName: "ProcStatus"},
		}, diags
	case "value":
		return FunctionType{
			Params: nil,
			Return: makeValueUnion(future.Result),
		}, diags
	case "cancel":
		diags = append(diags, Diagnostic{
			Message: "typechecker: future handles do not support cancel()",
			Node:    node,
		})
		return UnknownType{}, diags
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: future handle has no member '%s'", name),
			Node:    node,
		})
		return UnknownType{}, diags
	}
}

func makeValueUnion(success Type) Type {
	procErr := StructType{StructName: "ProcError"}
	members := []Type{success, procErr}
	return UnionLiteralType{Members: members}
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
		method, ok := spec.Methods[name]
		if !ok {
			continue
		}
		if len(subst) > 0 {
			method = substituteFunctionType(method, subst)
		}
		method = bindMethodType(method)
		if !found || score > bestScore {
			bestScore = score
			bestFn = method
			found = true
		}
	}
	return bestFn, bestScore, found
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
		method = bindMethodType(method)
		if !found || score > bestScore {
			bestScore = score
			bestFn = method
			found = true
		}
	}
	return bestFn, bestScore, found
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
	objInfo, ok := structInfoFromType(object)
	if !ok {
		return nil, 0, false
	}
	targetInfo, ok := structInfoFromType(target)
	if !ok {
		return nil, 0, false
	}
	if targetInfo.name == "" || objInfo.name == "" || targetInfo.name != objInfo.name {
		return nil, 0, false
	}
	subst := make(map[string]Type)
	score := 0
	for idx, targetArg := range targetInfo.args {
		var objArg Type = UnknownType{}
		if idx < len(objInfo.args) && objInfo.args[idx] != nil {
			objArg = objInfo.args[idx]
		}
		if objArg == nil {
			objArg = UnknownType{}
		}
		switch val := targetArg.(type) {
		case TypeParameterType:
			if existing, ok := subst[val.ParameterName]; ok {
				if !typesEquivalentForSignature(existing, objArg) {
					return nil, 0, false
				}
			} else {
				subst[val.ParameterName] = objArg
				if objArg != nil && !isUnknownType(objArg) {
					score++
				}
			}
		default:
			if !typesEquivalentForSignature(val, objArg) {
				return nil, 0, false
			}
		}
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
		return nil, score, true
	}
	return subst, score, true
}

type structInfo struct {
	name string
	args []Type
}

func structInfoFromType(t Type) (structInfo, bool) {
	switch v := t.(type) {
	case StructType:
		return structInfo{name: v.StructName}, v.StructName != ""
	case StructInstanceType:
		return structInfo{name: v.StructName}, v.StructName != ""
	case AppliedType:
		name, ok := structName(v.Base)
		if !ok {
			return structInfo{}, false
		}
		return structInfo{name: name, args: v.Arguments}, true
	default:
		return structInfo{}, false
	}
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
