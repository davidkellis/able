package bridge

import (
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func matchTypeWithoutInterpreter(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool) {
	if value == nil {
		value = runtime.NilValue{}
	}
	switch t := typeExpr.(type) {
	case nil:
		return value, true
	case *ast.WildcardTypeExpression:
		return value, true
	case *ast.NullableTypeExpression:
		if isRuntimeNilValue(value) {
			return runtime.NilValue{}, true
		}
		return matchTypeWithoutInterpreter(t.InnerType, value)
	case *ast.ResultTypeExpression:
		if isRuntimeErrorValue(value) {
			return value, true
		}
		return matchTypeWithoutInterpreter(t.InnerType, value)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if coerced, ok := matchTypeWithoutInterpreter(member, value); ok {
				return coerced, true
			}
		}
		return nil, false
	case *ast.FunctionTypeExpression:
		if isRuntimeCallableValue(value) {
			return value, true
		}
		return nil, false
	case *ast.GenericTypeExpression:
		return matchGenericTypeWithoutInterpreter(t, value)
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return nil, false
		}
		return matchSimpleTypeWithoutInterpreter(t.Name.Name, value)
	default:
		return nil, false
	}
}

func matchGenericTypeWithoutInterpreter(expr *ast.GenericTypeExpression, value runtime.Value) (runtime.Value, bool) {
	if expr == nil {
		return nil, false
	}
	base, ok := expr.Base.(*ast.SimpleTypeExpression)
	if !ok || base == nil || base.Name == nil {
		return nil, false
	}
	baseName := normalizeKernelTypeName(base.Name.Name)
	if baseName == "Option" && len(expr.Arguments) > 0 {
		if isRuntimeNilValue(value) {
			return runtime.NilValue{}, true
		}
		return matchTypeWithoutInterpreter(expr.Arguments[0], value)
	}
	if baseName == "Result" && len(expr.Arguments) > 0 {
		if isRuntimeErrorValue(value) {
			return value, true
		}
		return matchTypeWithoutInterpreter(expr.Arguments[0], value)
	}
	coerced, ok := matchSimpleTypeWithoutInterpreter(baseName, value)
	if !ok {
		return nil, false
	}
	if len(expr.Arguments) == 0 || baseName != "Array" {
		return coerced, true
	}
	elemType := expr.Arguments[0]
	arr, err := arrayValueFromRuntime(value)
	if err != nil || arr == nil {
		// Allow matching struct-backed arrays where element storage is opaque.
		return coerced, true
	}
	for _, elem := range arr.Elements {
		if _, ok := matchTypeWithoutInterpreter(elemType, elem); !ok {
			return nil, false
		}
	}
	return coerced, true
}

func matchSimpleTypeWithoutInterpreter(name string, value runtime.Value) (runtime.Value, bool) {
	name = normalizeKernelTypeName(name)
	switch name {
	case "", "_", "Self":
		if value == nil {
			return runtime.NilValue{}, true
		}
		return value, true
	case "bool":
		v, err := AsBool(value)
		if err != nil {
			return nil, false
		}
		return ToBool(v), true
	case "String":
		v, err := AsString(value)
		if err == nil {
			return ToString(v), true
		}
		if structNameFromValue(value) == "String" {
			return value, true
		}
		return nil, false
	case "char":
		v, err := AsRune(value)
		if err != nil {
			return nil, false
		}
		return ToRune(v), true
	case "nil":
		if isRuntimeNilValue(value) {
			return runtime.NilValue{}, true
		}
		return nil, false
	case "void":
		if isRuntimeVoidValue(value) {
			return runtime.VoidValue{}, true
		}
		return nil, false
	case "f32":
		v, err := AsFloat(value)
		if err != nil {
			return nil, false
		}
		return ToFloat32(float32(v)), true
	case "f64":
		v, err := AsFloat(value)
		if err != nil {
			return nil, false
		}
		return ToFloat64(v), true
	case "Error":
		if isRuntimeErrorValue(value) {
			return value, true
		}
		return nil, false
	}

	if suffix, bits, signed, ok := staticIntegerType(name); ok {
		if signed {
			v, err := AsInt(value, bits)
			if err != nil {
				return nil, false
			}
			return ToInt(v, suffix), true
		}
		v, err := AsUint(value, bits)
		if err != nil {
			return nil, false
		}
		return ToUint(v, suffix), true
	}

	if iface := interfaceNameFromValue(value); iface != "" && iface == name {
		return value, true
	}
	if structName := structNameFromValue(value); structName != "" && structName == name {
		return value, true
	}
	if typeRef, ok := value.(runtime.TypeRefValue); ok {
		return value, normalizeKernelTypeName(typeRef.TypeName) == name
	}
	if typeRef, ok := value.(*runtime.TypeRefValue); ok {
		if typeRef == nil {
			return nil, false
		}
		return value, normalizeKernelTypeName(typeRef.TypeName) == name
	}
	return nil, false
}

func staticTypeExpressionFromValue(value runtime.Value) ast.TypeExpression {
	value = unwrapInterface(value)
	switch typed := value.(type) {
	case runtime.StringValue, *runtime.StringValue:
		return ast.Ty("String")
	case runtime.BoolValue, *runtime.BoolValue:
		return ast.Ty("bool")
	case runtime.CharValue, *runtime.CharValue:
		return ast.Ty("char")
	case runtime.NilValue, *runtime.NilValue:
		return ast.Ty("nil")
	case runtime.VoidValue, *runtime.VoidValue:
		return ast.Ty("void")
	case runtime.IntegerValue:
		if typed.TypeSuffix != "" {
			return ast.Ty(string(typed.TypeSuffix))
		}
		return ast.Ty("i32")
	case *runtime.IntegerValue:
		if typed == nil {
			return ast.Ty("nil")
		}
		if typed.TypeSuffix != "" {
			return ast.Ty(string(typed.TypeSuffix))
		}
		return ast.Ty("i32")
	case runtime.FloatValue:
		if typed.TypeSuffix != "" {
			return ast.Ty(string(typed.TypeSuffix))
		}
		return ast.Ty("f64")
	case *runtime.FloatValue:
		if typed == nil {
			return ast.Ty("nil")
		}
		if typed.TypeSuffix != "" {
			return ast.Ty(string(typed.TypeSuffix))
		}
		return ast.Ty("f64")
	case *runtime.ArrayValue:
		return ast.Gen(ast.Ty("Array"), ast.NewWildcardTypeExpression())
	case *runtime.HashMapValue:
		return ast.Gen(ast.Ty("HashMap"), ast.NewWildcardTypeExpression(), ast.NewWildcardTypeExpression())
	case *runtime.FutureValue:
		return ast.Gen(ast.Ty("Future"), ast.NewWildcardTypeExpression())
	case *runtime.IteratorValue:
		return ast.Gen(ast.Ty("Iterator"), ast.NewWildcardTypeExpression())
	case runtime.ErrorValue, *runtime.ErrorValue:
		return ast.Ty("Error")
	case runtime.TypeRefValue:
		return ast.Ty(normalizeKernelTypeName(typed.TypeName))
	case *runtime.TypeRefValue:
		if typed == nil {
			return ast.Ty("nil")
		}
		return ast.Ty(normalizeKernelTypeName(typed.TypeName))
	}
	if iface := interfaceNameFromValue(value); iface != "" {
		return ast.Ty(iface)
	}
	if name := structNameFromValue(value); name != "" {
		return ast.Ty(name)
	}
	return nil
}

func staticIntegerType(name string) (suffix runtime.IntegerType, bits int, signed bool, ok bool) {
	switch normalizeKernelTypeName(name) {
	case "i8":
		return runtime.IntegerI8, 8, true, true
	case "i16":
		return runtime.IntegerI16, 16, true, true
	case "i32":
		return runtime.IntegerI32, 32, true, true
	case "i64":
		return runtime.IntegerI64, 64, true, true
	case "i128":
		return runtime.IntegerI128, 128, true, true
	case "isize":
		return runtime.IntegerIsize, NativeIntBits, true, true
	case "u8":
		return runtime.IntegerU8, 8, false, true
	case "u16":
		return runtime.IntegerU16, 16, false, true
	case "u32":
		return runtime.IntegerU32, 32, false, true
	case "u64":
		return runtime.IntegerU64, 64, false, true
	case "u128":
		return runtime.IntegerU128, 128, false, true
	case "usize":
		return runtime.IntegerUsize, NativeIntBits, false, true
	default:
		return "", 0, false, false
	}
}

func interfaceNameFromValue(value runtime.Value) string {
	switch typed := value.(type) {
	case runtime.InterfaceValue:
		if typed.Interface != nil && typed.Interface.Node != nil && typed.Interface.Node.ID != nil {
			return normalizeKernelTypeName(typed.Interface.Node.ID.Name)
		}
	case *runtime.InterfaceValue:
		if typed != nil && typed.Interface != nil && typed.Interface.Node != nil && typed.Interface.Node.ID != nil {
			return normalizeKernelTypeName(typed.Interface.Node.ID.Name)
		}
	}
	return ""
}

func structNameFromValue(value runtime.Value) string {
	switch typed := value.(type) {
	case runtime.StructDefinitionValue:
		if typed.Node != nil && typed.Node.ID != nil {
			return normalizeKernelTypeName(typed.Node.ID.Name)
		}
	case *runtime.StructDefinitionValue:
		if typed != nil && typed.Node != nil && typed.Node.ID != nil {
			return normalizeKernelTypeName(typed.Node.ID.Name)
		}
	case *runtime.StructInstanceValue:
		if typed != nil && typed.Definition != nil && typed.Definition.Node != nil && typed.Definition.Node.ID != nil {
			return normalizeKernelTypeName(typed.Definition.Node.ID.Name)
		}
	}
	return ""
}

func isRuntimeNilValue(value runtime.Value) bool {
	value = unwrapInterface(value)
	switch value.(type) {
	case nil:
		return true
	case runtime.NilValue:
		return true
	case *runtime.NilValue:
		return true
	default:
		return false
	}
}

func isRuntimeVoidValue(value runtime.Value) bool {
	value = unwrapInterface(value)
	switch typed := value.(type) {
	case runtime.VoidValue:
		return true
	case *runtime.VoidValue:
		return typed != nil
	default:
		return false
	}
}

func isRuntimeErrorValue(value runtime.Value) bool {
	value = unwrapInterface(value)
	switch typed := value.(type) {
	case runtime.ErrorValue:
		return true
	case *runtime.ErrorValue:
		return typed != nil
	default:
		return interfaceNameFromValue(value) == "Error"
	}
}

func isRuntimeCallableValue(value runtime.Value) bool {
	value = unwrapInterface(value)
	switch value.(type) {
	case *runtime.FunctionValue,
		*runtime.FunctionOverloadValue,
		runtime.NativeFunctionValue,
		*runtime.NativeFunctionValue,
		runtime.BoundMethodValue,
		*runtime.BoundMethodValue,
		runtime.NativeBoundMethodValue,
		*runtime.NativeBoundMethodValue,
		runtime.PartialFunctionValue,
		*runtime.PartialFunctionValue:
		return true
	default:
		return false
	}
}

func normalizeKernelTypeName(name string) string {
	switch strings.TrimSpace(name) {
	case "KernelArray":
		return "Array"
	case "KernelChannel":
		return "Channel"
	case "KernelHashMap":
		return "HashMap"
	case "KernelMutex":
		return "Mutex"
	case "KernelRange":
		return "Range"
	case "KernelRangeFactory":
		return "RangeFactory"
	case "KernelRatio":
		return "Ratio"
	case "KernelAwaitable":
		return "Awaitable"
	case "KernelAwaitWaker":
		return "AwaitWaker"
	case "KernelAwaitRegistration":
		return "AwaitRegistration"
	case "KernelLess":
		return "Less"
	case "KernelGreater":
		return "Greater"
	case "KernelEqual":
		return "Equal"
	case "KernelOrdering":
		return "Ordering"
	default:
		return strings.TrimSpace(name)
	}
}
