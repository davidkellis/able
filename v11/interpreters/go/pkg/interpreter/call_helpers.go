package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func minArgsForOverloads(overloads []*runtime.FunctionValue) int {
	min := -1
	for _, fn := range overloads {
		if fn == nil {
			continue
		}
		req := minArgsForFunctionValue(fn)
		if req < 0 {
			continue
		}
		if min < 0 || req < min {
			min = req
		}
	}
	if min < 0 {
		return 0
	}
	return min
}

func minArgsForFunctionValue(fn *runtime.FunctionValue) int {
	if fn == nil || fn.Declaration == nil {
		return 0
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		paramCount := len(decl.Params)
		expected := paramCount
		if decl.IsMethodShorthand {
			expected++
		}
		if paramCount > 0 && isNullableParam(decl.Params[paramCount-1]) {
			expected--
		}
		if expected < 0 {
			return 0
		}
		return expected
	case *ast.LambdaExpression:
		paramCount := len(decl.Params)
		if paramCount > 0 && isNullableParam(decl.Params[paramCount-1]) {
			paramCount--
		}
		if paramCount < 0 {
			return 0
		}
		return paramCount
	default:
		return 0
	}
}

func overloadName(call *ast.FunctionCall) string {
	if call == nil {
		return "(function)"
	}
	switch cal := call.Callee.(type) {
	case *ast.Identifier:
		return cal.Name
	case *ast.MemberAccessExpression:
		if id, ok := cal.Member.(*ast.Identifier); ok {
			return id.Name
		}
	}
	return "(function)"
}

func parameterSpecificity(typeExpr ast.TypeExpression, generics map[string]struct{}) int {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return 0
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return 0
		}
		if _, ok := generics[t.Name.Name]; ok {
			return 1
		}
		return 3
	case *ast.NullableTypeExpression:
		return 1 + parameterSpecificity(t.InnerType, generics)
	case *ast.GenericTypeExpression:
		score := 2 + parameterSpecificity(t.Base, generics)
		for _, arg := range t.Arguments {
			score += parameterSpecificity(arg, generics)
		}
		return score
	case *ast.FunctionTypeExpression, *ast.UnionTypeExpression:
		return 2
	default:
		return 1
	}
}

func functionGenericNameSet(fn *runtime.FunctionValue, decl *ast.FunctionDefinition) map[string]struct{} {
	names := genericNameSet(nil)
	if decl != nil {
		names = genericNameSet(decl.GenericParams)
	}
	if fn == nil || fn.MethodSet == nil || len(fn.MethodSet.GenericParams) == 0 {
		return names
	}
	for _, gp := range fn.MethodSet.GenericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		names[gp.Name.Name] = struct{}{}
	}
	return names
}

func arityMatchesRuntime(expected, actual int, optionalLast bool) bool {
	return actual == expected || (optionalLast && actual == expected-1)
}

func isNullableParam(param *ast.FunctionParameter) bool {
	if param == nil {
		return false
	}
	if param.ParamType == nil {
		return false
	}
	_, ok := param.ParamType.(*ast.NullableTypeExpression)
	return ok
}

func paramUsesGeneric(typeExpr ast.TypeExpression, generics map[string]struct{}) bool {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false
	}
	_, ok = generics[simple.Name.Name]
	return ok
}

func describeRuntimeValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return "String"
	case runtime.BoolValue:
		return "bool"
	case runtime.CharValue:
		return "char"
	case runtime.IntegerValue:
		return string(v.TypeSuffix)
	case *runtime.IntegerValue:
		if v == nil {
			return "nil"
		}
		return string(v.TypeSuffix)
	case runtime.FloatValue:
		return string(v.TypeSuffix)
	case *runtime.StructInstanceValue:
		if v == nil {
			return "nil"
		}
		if name := structTypeName(v); name != "" {
			return name
		}
		return "struct_instance"
	case runtime.NilValue:
		return "nil"
	}
	if val == nil {
		return "<nil>"
	}
	return val.Kind().String()
}
