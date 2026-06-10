//go:build !(js && wasm)

package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/typechecker"
)

// staticReceiverTypeForCall exposes a checked receiver type to generic union
// methods. It is intentionally call-local: dynamically typed and unchecked
// evaluation keeps the existing runtime-only dispatch behaviour.
func (i *Interpreter) staticReceiverTypeForCall(call *ast.FunctionCall, env *runtime.Environment) ast.TypeExpression {
	if i == nil || call == nil {
		return nil
	}
	if registered := i.registeredStaticCallReceiverType(call); registered != nil {
		return resolveStaticReceiverTypeBindings(registered, env, make(map[string]struct{}))
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Object == nil {
		return nil
	}
	facts := i.runtimeInferenceFactsSnapshot()
	typ, ok := facts[member.Object]
	if !ok || typ == nil {
		return nil
	}
	return resolveStaticReceiverTypeBindings(inferredTypeExpression(typ), env, make(map[string]struct{}))
}

func inferredTypeExpression(typ typechecker.Type) ast.TypeExpression {
	switch value := typ.(type) {
	case typechecker.PrimitiveType:
		switch value.Kind {
		case typechecker.PrimitiveBool:
			return ast.Ty("bool")
		case typechecker.PrimitiveChar:
			return ast.Ty("char")
		case typechecker.PrimitiveString:
			return ast.Ty("String")
		case typechecker.PrimitiveIoHandle:
			return ast.Ty("IoHandle")
		case typechecker.PrimitiveProcHandle:
			return ast.Ty("ProcHandle")
		case typechecker.PrimitiveNil:
			return ast.Ty("nil")
		}
	case typechecker.IntegerType:
		if value.Suffix != "" {
			return ast.Ty(value.Suffix)
		}
	case typechecker.FloatType:
		if value.Suffix != "" {
			return ast.Ty(value.Suffix)
		}
	case typechecker.TypeParameterType:
		if value.ParameterName != "" {
			return ast.Ty(value.ParameterName)
		}
	case typechecker.StructType:
		if value.StructName != "" {
			return ast.Ty(value.StructName)
		}
	case typechecker.StructInstanceType:
		return inferredAppliedTypeExpression(value.StructName, value.TypeArgs)
	case typechecker.InterfaceType:
		if value.InterfaceName != "" {
			return ast.Ty(value.InterfaceName)
		}
	case typechecker.ArrayType:
		return ast.Gen(ast.Ty("Array"), inferredTypeExpression(value.Element))
	case typechecker.RangeType:
		return ast.Gen(ast.Ty("Range"), inferredTypeExpression(value.Element))
	case typechecker.IteratorType:
		return ast.Gen(ast.Ty("Iterator"), inferredTypeExpression(value.Element))
	case typechecker.FutureType:
		return ast.Gen(ast.Ty("Future"), inferredTypeExpression(value.Result))
	case typechecker.NullableType:
		return ast.NewNullableTypeExpression(inferredTypeExpression(value.Inner))
	case typechecker.UnionType:
		return inferredUnionTypeExpression(value.Variants)
	case typechecker.UnionLiteralType:
		return inferredUnionTypeExpression(value.Members)
	case typechecker.FunctionType:
		params := make([]ast.TypeExpression, len(value.Params))
		for index, param := range value.Params {
			params[index] = inferredTypeExpression(param)
		}
		return ast.NewFunctionTypeExpression(params, inferredTypeExpression(value.Return))
	case typechecker.AppliedType:
		base := inferredTypeExpression(value.Base)
		if base == nil || len(value.Arguments) == 0 {
			return base
		}
		args := make([]ast.TypeExpression, len(value.Arguments))
		for index, arg := range value.Arguments {
			args[index] = inferredTypeExpression(arg)
		}
		return ast.NewGenericTypeExpression(base, args)
	case typechecker.AliasType:
		if value.Target != nil {
			return inferredTypeExpression(value.Target)
		}
		if value.AliasName != "" {
			return ast.Ty(value.AliasName)
		}
	}
	return ast.NewWildcardTypeExpression()
}

func inferredAppliedTypeExpression(name string, args []typechecker.Type) ast.TypeExpression {
	if name == "" {
		return ast.NewWildcardTypeExpression()
	}
	if len(args) == 0 {
		return ast.Ty(name)
	}
	expressions := make([]ast.TypeExpression, len(args))
	for index, arg := range args {
		expressions[index] = inferredTypeExpression(arg)
	}
	return ast.Gen(ast.Ty(name), expressions...)
}

func inferredUnionTypeExpression(members []typechecker.Type) ast.TypeExpression {
	if len(members) == 0 {
		return ast.NewWildcardTypeExpression()
	}
	expressions := make([]ast.TypeExpression, len(members))
	for index, member := range members {
		expressions[index] = inferredTypeExpression(member)
	}
	return ast.NewUnionTypeExpression(expressions)
}

func resolveStaticReceiverTypeBindings(expr ast.TypeExpression, env *runtime.Environment, seen map[string]struct{}) ast.TypeExpression {
	if expr == nil || env == nil {
		return expr
	}
	switch value := expr.(type) {
	case *ast.SimpleTypeExpression:
		if value == nil || value.Name == nil || value.Name.Name == "" {
			return expr
		}
		name := value.Name.Name
		if _, resolving := seen[name]; resolving {
			return expr
		}
		raw, ok := env.Lookup(name)
		if !ok {
			return expr
		}
		ref, ok := runtimeTypeRefExpression(raw)
		if !ok {
			return expr
		}
		seen[name] = struct{}{}
		resolved := resolveStaticReceiverTypeBindings(ref, env, seen)
		delete(seen, name)
		return resolved
	case *ast.GenericTypeExpression:
		if value == nil {
			return expr
		}
		base := resolveStaticReceiverTypeBindings(value.Base, env, seen)
		args := make([]ast.TypeExpression, len(value.Arguments))
		for index, arg := range value.Arguments {
			args[index] = resolveStaticReceiverTypeBindings(arg, env, seen)
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		inner := resolveStaticReceiverTypeBindings(value.InnerType, env, seen)
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		inner := resolveStaticReceiverTypeBindings(value.InnerType, env, seen)
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(value.Members))
		for index, member := range value.Members {
			members[index] = resolveStaticReceiverTypeBindings(member, env, seen)
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(value.ParamTypes))
		for index, param := range value.ParamTypes {
			params[index] = resolveStaticReceiverTypeBindings(param, env, seen)
		}
		result := resolveStaticReceiverTypeBindings(value.ReturnType, env, seen)
		return ast.NewFunctionTypeExpression(params, result)
	default:
		return expr
	}
}

func runtimeTypeRefExpression(value runtime.Value) (ast.TypeExpression, bool) {
	var ref runtime.TypeRefValue
	switch typed := value.(type) {
	case runtime.TypeRefValue:
		ref = typed
	case *runtime.TypeRefValue:
		if typed == nil {
			return nil, false
		}
		ref = *typed
	default:
		return nil, false
	}
	if ref.TypeName == "" {
		return nil, false
	}
	if len(ref.TypeArgs) == 0 {
		return ast.Ty(ref.TypeName), true
	}
	return ast.Gen(ast.Ty(ref.TypeName), ref.TypeArgs...), true
}
