package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// interfaceMethodDynamicSafety classifies a raw interface method signature
// before Self is substituted with an existential interface view.
func interfaceMethodDynamicSafety(method FunctionType) (bool, string) {
	for index, param := range method.Params {
		if index == 0 && isExactSelfType(param) {
			continue
		}
		if typeContainsSelf(param) {
			if index == 0 {
				return false, "the dispatch receiver contains Self but is not exactly Self"
			}
			return false, fmt.Sprintf("parameter %d contains Self", index+1)
		}
	}
	if typeContainsSelf(method.Return) && !isExactSelfType(method.Return) {
		return false, "the result contains Self but is not exactly Self"
	}
	for _, param := range method.TypeParams {
		for _, constraint := range param.Constraints {
			if typeContainsSelf(constraint) {
				return false, fmt.Sprintf("generic parameter %s has a constraint containing Self", param.Name)
			}
		}
	}
	for _, clause := range method.Where {
		if typeContainsSelf(clause.Subject) {
			return false, fmt.Sprintf("where clause for %s has a subject containing Self", clause.TypeParam)
		}
		for _, constraint := range clause.Constraints {
			if typeContainsSelf(constraint) {
				return false, fmt.Sprintf("where clause for %s has a constraint containing Self", clause.TypeParam)
			}
		}
	}
	for _, obligation := range method.Obligations {
		if typeContainsSelf(obligation.Subject) || typeContainsSelf(obligation.Constraint) {
			return false, "a method obligation contains Self"
		}
	}
	return true, ""
}

func isExactSelfType(typ Type) bool {
	param, ok := typ.(TypeParameterType)
	return ok && param.ParameterName == "Self"
}

func typeContainsSelf(typ Type) bool {
	switch value := typ.(type) {
	case nil:
		return false
	case TypeParameterType:
		return value.ParameterName == "Self"
	case FunctionType:
		for _, param := range value.Params {
			if typeContainsSelf(param) {
				return true
			}
		}
		if typeContainsSelf(value.Return) {
			return true
		}
		for _, param := range value.TypeParams {
			for _, constraint := range param.Constraints {
				if typeContainsSelf(constraint) {
					return true
				}
			}
		}
		for _, clause := range value.Where {
			if typeContainsSelf(clause.Subject) {
				return true
			}
			for _, constraint := range clause.Constraints {
				if typeContainsSelf(constraint) {
					return true
				}
			}
		}
		for _, obligation := range value.Obligations {
			if typeContainsSelf(obligation.Subject) || typeContainsSelf(obligation.Constraint) {
				return true
			}
		}
		return false
	case FunctionOverloadType:
		for _, overload := range value.Overloads {
			if typeContainsSelf(overload) {
				return true
			}
		}
		return false
	case FutureType:
		return typeContainsSelf(value.Result)
	case AppliedType:
		if typeContainsSelf(value.Base) {
			return true
		}
		return anyTypeContainsSelf(value.Arguments)
	case NullableType:
		return typeContainsSelf(value.Inner)
	case UnionLiteralType:
		return anyTypeContainsSelf(value.Members)
	case AliasType:
		return typeContainsSelf(value.Target)
	case StructInstanceType:
		return anyTypeContainsSelf(value.TypeArgs)
	case ArrayType:
		return typeContainsSelf(value.Element)
	case MapType:
		return typeContainsSelf(value.Key) || typeContainsSelf(value.Value)
	case RangeType:
		if typeContainsSelf(value.Element) {
			return true
		}
		return anyTypeContainsSelf(value.Bounds)
	case IteratorType:
		return typeContainsSelf(value.Element)
	default:
		return false
	}
}

func anyTypeContainsSelf(types []Type) bool {
	for _, typ := range types {
		if typeContainsSelf(typ) {
			return true
		}
	}
	return false
}

func (c *Checker) interfaceMemberFunctionForAccess(
	env *Environment,
	expr *ast.MemberAccessExpression,
	iface InterfaceType,
	args []Type,
	self Type,
	name string,
) (FunctionType, []Diagnostic, bool) {
	if iface.InterfaceName == "" || iface.Methods == nil {
		return FunctionType{}, nil, false
	}
	raw, ok := iface.Methods[name]
	if !ok {
		return FunctionType{}, nil, false
	}
	if !c.interfaceMemberObjectIsTypeNamespace(env, expr.Object) {
		if safe, reason := interfaceMethodDynamicSafety(raw); !safe {
			return FunctionType{}, []Diagnostic{{
				Code: DiagnosticCodeStaticOnlyInterfaceMethod,
				Message: fmt.Sprintf(
					"typechecker: interface method '%s.%s' is static-only on interface values: %s",
					iface.InterfaceName,
					name,
					reason,
				),
				Node: expr,
			}}, true
		}
	}
	method, ok := c.interfaceMemberFunction(iface, args, self, name)
	return method, nil, ok
}

func (c *Checker) interfaceMemberObjectIsTypeNamespace(env *Environment, object ast.Expression) bool {
	switch value := object.(type) {
	case *ast.Identifier:
		for scope := env; scope != nil; scope = scope.parent {
			typ, ok := scope.symbols[value.Name]
			if !ok {
				continue
			}
			if scope != c.global {
				return false
			}
			_, _, isInterface := interfaceFromType(typ)
			return isInterface
		}
	case *ast.MemberAccessExpression:
		pkgType, ok := c.infer.get(value.Object)
		if !ok {
			return false
		}
		pkg, ok := pkgType.(PackageType)
		if !ok {
			return false
		}
		member, ok := value.Member.(*ast.Identifier)
		if !ok || pkg.Symbols == nil {
			return false
		}
		typ, ok := pkg.Symbols[member.Name]
		if !ok {
			return false
		}
		_, _, isInterface := interfaceFromType(typ)
		return isInterface
	}
	return false
}
