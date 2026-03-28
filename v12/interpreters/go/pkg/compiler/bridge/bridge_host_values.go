package bridge

import (
	"fmt"
	"math/big"
	"reflect"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var anyReflectType = reflect.TypeOf((*any)(nil)).Elem()

func RuntimeValueToHost[T any](typeExpr ast.TypeExpression, value runtime.Value) (T, error) {
	var zero T
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	hostVal, err := runtimeValueToHost(typeExpr, value, targetType)
	if err != nil {
		return zero, err
	}
	if hostVal == nil {
		return zero, nil
	}
	rv := reflect.ValueOf(hostVal)
	if !rv.IsValid() {
		return zero, nil
	}
	if rv.Type().AssignableTo(targetType) {
		return rv.Interface().(T), nil
	}
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType).Interface().(T), nil
	}
	return zero, fmt.Errorf("extern argument cannot convert %s to %s", rv.Type(), targetType)
}

func HostValueToRuntime(rt *Runtime, typeExpr ast.TypeExpression, value any) (runtime.Value, error) {
	return hostValueToRuntime(rt, typeExpr, reflect.ValueOf(value))
}

func HostResultToRuntime(rt *Runtime, innerType ast.TypeExpression, value any, err error) (runtime.Value, error) {
	if err != nil {
		errValue := ErrorValue(rt, ToString(err.Error()))
		return errValue, nil
	}
	return HostValueToRuntime(rt, innerType, value)
}

func runtimeValueToHost(typeExpr ast.TypeExpression, value runtime.Value, targetType reflect.Type) (any, error) {
	if targetType == nil {
		targetType = anyReflectType
	}
	if value == nil {
		return reflect.Zero(targetType).Interface(), nil
	}
	value = unwrapInterface(value)
	switch t := typeExpr.(type) {
	case nil, *ast.WildcardTypeExpression:
		return runtimeValueToHostByKind(value, targetType)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if member == nil {
				continue
			}
			if _, ok := matchTypeWithoutInterpreter(member, value); ok {
				return runtimeValueToHost(member, value, targetType)
			}
		}
		return runtimeValueToHostByKind(value, targetType)
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return nil, nil
		}
		if targetType.Kind() != reflect.Pointer {
			if targetType.Kind() == reflect.Interface {
				return runtimeValueToHost(t.InnerType, value, anyReflectType)
			}
			return runtimeValueToHost(t.InnerType, value, targetType)
		}
		elemVal, err := runtimeValueToHost(t.InnerType, value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(targetType.Elem())
		if elemVal != nil {
			ev := reflect.ValueOf(elemVal)
			if ev.IsValid() {
				if ev.Type().AssignableTo(targetType.Elem()) {
					ptr.Elem().Set(ev)
				} else if ev.Type().ConvertibleTo(targetType.Elem()) {
					ptr.Elem().Set(ev.Convert(targetType.Elem()))
				}
			}
		}
		return ptr.Interface(), nil
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && normalizeKernelTypeName(base.Name.Name) == "Array" {
			arr, err := arrayValueFromRuntime(value)
			if err != nil {
				return nil, err
			}
			var elemExpr ast.TypeExpression
			if len(t.Arguments) > 0 {
				elemExpr = t.Arguments[0]
			}
			if targetType.Kind() != reflect.Slice {
				out := make([]any, len(arr.Elements))
				for idx, elem := range arr.Elements {
					hostElem, err := runtimeValueToHost(elemExpr, elem, anyReflectType)
					if err != nil {
						return nil, err
					}
					out[idx] = hostElem
				}
				return out, nil
			}
			slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
			elemType := targetType.Elem()
			for idx, elem := range arr.Elements {
				hostElem, err := runtimeValueToHost(elemExpr, elem, elemType)
				if err != nil {
					return nil, err
				}
				ev := reflect.ValueOf(hostElem)
				if !ev.IsValid() {
					continue
				}
				if ev.Type().AssignableTo(elemType) {
					slice.Index(idx).Set(ev)
				} else if ev.Type().ConvertibleTo(elemType) {
					slice.Index(idx).Set(ev.Convert(elemType))
				}
			}
			return slice.Interface(), nil
		}
		return runtimeValueToHostByKind(value, targetType)
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return runtimeValueToHostByKind(value, targetType)
		}
		switch normalizeKernelTypeName(t.Name.Name) {
		case "String":
			return coerceStringValue(value)
		case "bool":
			return AsBool(value)
		case "char":
			return AsRune(value)
		case "IoHandle", "ProcHandle":
			if hv, ok := value.(*runtime.HostHandleValue); ok && hv != nil {
				return hv.Value, nil
			}
			return nil, nil
		case "f32":
			v, err := AsFloat(value)
			return float32(v), err
		case "f64":
			return AsFloat(value)
		case "i8":
			v, err := AsInt(value, 8)
			return int8(v), err
		case "i16":
			v, err := AsInt(value, 16)
			return int16(v), err
		case "i32":
			v, err := AsInt(value, 32)
			return int32(v), err
		case "i64":
			return AsInt(value, 64)
		case "u8":
			v, err := AsUint(value, 8)
			return uint8(v), err
		case "u16":
			v, err := AsUint(value, 16)
			return uint16(v), err
		case "u32":
			v, err := AsUint(value, 32)
			return uint32(v), err
		case "u64":
			return AsUint(value, 64)
		case "i128", "u128":
			val, err := extractInteger(value)
			if err != nil {
				return nil, err
			}
			return new(big.Int).Set(val), nil
		case "void":
			return nil, nil
		default:
			return structToHostValue(t.Name.Name, value)
		}
	default:
		return runtimeValueToHostByKind(value, targetType)
	}
}

func runtimeValueToHostByKind(value runtime.Value, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.String:
		return coerceStringValue(value)
	case reflect.Bool:
		return AsBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num, err := toInt64(value)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(num).Convert(targetType).Interface(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num, err := toUint64(value)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(num).Convert(targetType).Interface(), nil
	case reflect.Float32, reflect.Float64:
		num, err := AsFloat(value)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(num).Convert(targetType).Interface(), nil
	case reflect.Slice:
		arr, err := arrayValueFromRuntime(value)
		if err != nil {
			return nil, err
		}
		slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
		for idx, elem := range arr.Elements {
			hostElem, err := runtimeValueToHost(nil, elem, targetType.Elem())
			if err != nil {
				return nil, err
			}
			ev := reflect.ValueOf(hostElem)
			if !ev.IsValid() {
				continue
			}
			if ev.Type().AssignableTo(targetType.Elem()) {
				slice.Index(idx).Set(ev)
			} else if ev.Type().ConvertibleTo(targetType.Elem()) {
				slice.Index(idx).Set(ev.Convert(targetType.Elem()))
			}
		}
		return slice.Interface(), nil
	case reflect.Interface:
		return value, nil
	case reflect.Pointer:
		if _, ok := value.(runtime.NilValue); ok {
			return nil, nil
		}
		elem, err := runtimeValueToHost(nil, value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(targetType.Elem())
		if elem != nil {
			ev := reflect.ValueOf(elem)
			if ev.IsValid() {
				if ev.Type().AssignableTo(targetType.Elem()) {
					ptr.Elem().Set(ev)
				} else if ev.Type().ConvertibleTo(targetType.Elem()) {
					ptr.Elem().Set(ev.Convert(targetType.Elem()))
				}
			}
		}
		return ptr.Interface(), nil
	default:
		return nil, fmt.Errorf("unsupported extern argument type")
	}
}

func HostValueToRuntimeExpr(rt *Runtime, typeExpr ast.TypeExpression, value reflect.Value) (runtime.Value, error) {
	return hostValueToRuntime(rt, typeExpr, value)
}

func hostValueToRuntime(rt *Runtime, typeExpr ast.TypeExpression, value reflect.Value) (runtime.Value, error) {
	if rt != nil && typeExpr != nil {
		expanded, err := ExpandTypeAliases(rt, typeExpr)
		if err != nil {
			return nil, err
		}
		if expanded != nil {
			typeExpr = expanded
		}
	}
	for value.IsValid() && value.Kind() == reflect.Interface {
		if value.IsNil() {
			value = reflect.Value{}
			break
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		if _, ok := typeExpr.(*ast.NullableTypeExpression); ok {
			return runtime.NilValue{}, nil
		}
		if simple, ok := typeExpr.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && normalizeKernelTypeName(simple.Name.Name) == "void" {
			return runtime.VoidValue{}, nil
		}
		if _, ok := typeExpr.(*ast.UnionTypeExpression); !ok {
			return nil, fmt.Errorf("extern value is nil")
		}
	}
	switch t := typeExpr.(type) {
	case nil, *ast.WildcardTypeExpression:
		return hostValueToRuntimeByKind(value)
	case *ast.UnionTypeExpression:
		var lastErr error
		for _, member := range t.Members {
			if member == nil {
				continue
			}
			converted, err := hostValueToRuntime(rt, member, value)
			if err == nil {
				return converted, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return runtime.NilValue{}, nil
	case *ast.NullableTypeExpression:
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return runtime.NilValue{}, nil
			}
			return hostValueToRuntime(rt, t.InnerType, value.Elem())
		}
		if value.Kind() == reflect.Invalid {
			return runtime.NilValue{}, nil
		}
		return hostValueToRuntime(rt, t.InnerType, value)
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && normalizeKernelTypeName(base.Name.Name) == "Array" {
			if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
				return nil, fmt.Errorf("extern expected slice result")
			}
			var elemExpr ast.TypeExpression
			if len(t.Arguments) > 0 {
				elemExpr = t.Arguments[0]
			}
			elements := make([]runtime.Value, value.Len())
			for idx := 0; idx < value.Len(); idx++ {
				elem, err := hostValueToRuntime(rt, elemExpr, value.Index(idx))
				if err != nil {
					return nil, err
				}
				elements[idx] = elem
			}
			return &runtime.ArrayValue{Elements: elements}, nil
		}
		return hostValueToRuntimeByKind(value)
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return hostValueToRuntimeByKind(value)
		}
		name := normalizeKernelTypeName(t.Name.Name)
		switch name {
		case "String":
			return runtime.StringValue{Val: fmt.Sprint(value.Interface())}, nil
		case "bool":
			if value.Kind() == reflect.Bool {
				return runtime.BoolValue{Val: value.Bool()}, nil
			}
		case "char":
			switch value.Kind() {
			case reflect.Int32, reflect.Int, reflect.Int64:
				return runtime.CharValue{Val: rune(value.Int())}, nil
			}
		case "IoHandle", "ProcHandle":
			return &runtime.HostHandleValue{HandleType: name, Value: value.Interface()}, nil
		case "f32", "f64":
			if value.Kind() == reflect.Float32 || value.Kind() == reflect.Float64 {
				return runtime.FloatValue{Val: value.Convert(reflect.TypeOf(float64(0))).Float(), TypeSuffix: runtime.FloatType(name)}, nil
			}
		case "i8", "i16", "i32", "i64":
			if value.Kind() >= reflect.Int && value.Kind() <= reflect.Int64 {
				return runtime.NewSmallInt(value.Int(), runtime.IntegerType(name)), nil
			}
		case "u8", "u16", "u32", "u64":
			if value.Kind() >= reflect.Uint && value.Kind() <= reflect.Uint64 {
				return runtime.NewBigIntValue(bigIntFromUint(value.Uint()), runtime.IntegerType(name)), nil
			}
		case "i128", "u128":
			if value.Kind() == reflect.Pointer {
				if bi, ok := value.Interface().(*big.Int); ok && bi != nil {
					return runtime.NewBigIntValue(new(big.Int).Set(bi), runtime.IntegerType(name)), nil
				}
			}
		case "void":
			return runtime.VoidValue{}, nil
		default:
			if rt == nil {
				return nil, fmt.Errorf("missing runtime bridge for struct conversion")
			}
			def, err := rt.StructDefinition(name)
			if err == nil && def != nil {
				return structFromHostValue(def, value, func(expr ast.TypeExpression, v reflect.Value) (runtime.Value, error) {
					return hostValueToRuntime(rt, expr, v)
				})
			}
			unionDef, err := rt.UnionDefinition(name)
			if err == nil && unionDef != nil && unionDef.Node != nil {
				unionExpr := &ast.UnionTypeExpression{Members: unionDef.Node.Variants}
				return hostValueToRuntime(rt, unionExpr, value)
			}
		}
	}
	return hostValueToRuntimeByKind(value)
}

func hostValueToRuntimeByKind(value reflect.Value) (runtime.Value, error) {
	if !value.IsValid() {
		return runtime.NilValue{}, nil
	}
	switch value.Kind() {
	case reflect.String:
		return runtime.StringValue{Val: value.String()}, nil
	case reflect.Bool:
		return runtime.BoolValue{Val: value.Bool()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return runtime.NewSmallInt(value.Int(), runtime.IntegerI64), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return runtime.NewBigIntValue(bigIntFromUint(value.Uint()), runtime.IntegerU64), nil
	case reflect.Float32:
		return runtime.FloatValue{Val: value.Convert(reflect.TypeOf(float64(0))).Float(), TypeSuffix: runtime.FloatF32}, nil
	case reflect.Float64:
		return runtime.FloatValue{Val: value.Float(), TypeSuffix: runtime.FloatF64}, nil
	case reflect.Pointer:
		if value.IsNil() {
			return runtime.NilValue{}, nil
		}
		if bi, ok := value.Interface().(*big.Int); ok && bi != nil {
			return runtime.NewBigIntValue(new(big.Int).Set(bi), runtime.IntegerI128), nil
		}
		return hostValueToRuntimeByKind(value.Elem())
	case reflect.Slice, reflect.Array:
		elements := make([]runtime.Value, value.Len())
		for idx := 0; idx < value.Len(); idx++ {
			elem, err := hostValueToRuntimeByKind(value.Index(idx))
			if err != nil {
				return nil, err
			}
			elements[idx] = elem
		}
		return &runtime.ArrayValue{Elements: elements}, nil
	default:
		return nil, fmt.Errorf("unsupported extern return type %s", value.Kind())
	}
}

func structToHostValue(name string, value runtime.Value) (any, error) {
	var def *runtime.StructDefinitionValue
	switch v := unwrapInterface(value).(type) {
	case *runtime.StructInstanceValue:
		def = v.Definition
	case *runtime.StructDefinitionValue:
		def = v
	}
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("expected %s struct instance", name)
	}
	if def.Node.Kind == ast.StructKindSingleton || len(def.Node.Fields) == 0 {
		return def.Node.ID.Name, nil
	}
	inst, ok := unwrapInterface(value).(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, fmt.Errorf("expected %s struct instance", name)
	}
	if def.Node.Kind == ast.StructKindPositional && len(inst.Positional) > 0 {
		out := make([]any, len(inst.Positional))
		for idx, elem := range inst.Positional {
			var fieldType ast.TypeExpression
			if idx < len(def.Node.Fields) && def.Node.Fields[idx] != nil {
				fieldType = def.Node.Fields[idx].FieldType
			}
			hostElem, err := runtimeValueToHost(fieldType, elem, anyReflectType)
			if err != nil {
				return nil, err
			}
			out[idx] = hostElem
		}
		return out, nil
	}
	out := make(map[string]any, len(def.Node.Fields))
	for _, field := range def.Node.Fields {
		if field == nil || field.Name == nil {
			continue
		}
		fieldVal, ok := inst.Fields[field.Name.Name]
		if !ok {
			if field.FieldType != nil {
				if _, nullable := field.FieldType.(*ast.NullableTypeExpression); nullable {
					out[field.Name.Name] = nil
					continue
				}
			}
			return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, field.Name.Name)
		}
		hostVal, err := runtimeValueToHost(field.FieldType, fieldVal, anyReflectType)
		if err != nil {
			return nil, err
		}
		out[field.Name.Name] = hostVal
	}
	return out, nil
}

func structFromHostValue(def *runtime.StructDefinitionValue, value reflect.Value, recurse func(ast.TypeExpression, reflect.Value) (runtime.Value, error)) (runtime.Value, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing")
	}
	if !value.IsValid() {
		return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
	}
	if def.Node.Kind == ast.StructKindSingleton || len(def.Node.Fields) == 0 {
		if value.Kind() == reflect.String && value.String() == def.Node.ID.Name {
			return &runtime.StructInstanceValue{Definition: def, Fields: map[string]runtime.Value{}}, nil
		}
		return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
	}
	if def.Node.Kind == ast.StructKindPositional && (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) {
		positional := make([]runtime.Value, value.Len())
		for idx := 0; idx < value.Len(); idx++ {
			var fieldType ast.TypeExpression
			if idx < len(def.Node.Fields) && def.Node.Fields[idx] != nil {
				fieldType = def.Node.Fields[idx].FieldType
			}
			elem, err := recurse(fieldType, value.Index(idx))
			if err != nil {
				return nil, err
			}
			positional[idx] = elem
		}
		return &runtime.StructInstanceValue{Definition: def, Positional: positional}, nil
	}
	fields := make(map[string]runtime.Value, len(def.Node.Fields))
	if value.Kind() == reflect.Map {
		for _, field := range def.Node.Fields {
			if field == nil || field.Name == nil {
				continue
			}
			fieldName := field.Name.Name
			key := reflect.ValueOf(fieldName)
			if key.Type().AssignableTo(value.Type().Key()) {
				key = key.Convert(value.Type().Key())
			} else if key.Type().ConvertibleTo(value.Type().Key()) {
				key = key.Convert(value.Type().Key())
			} else {
				continue
			}
			entry := value.MapIndex(key)
			if !entry.IsValid() {
				if field.FieldType != nil {
					if _, nullable := field.FieldType.(*ast.NullableTypeExpression); nullable {
						fields[fieldName] = runtime.NilValue{}
						continue
					}
				}
				return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
			}
			converted, err := recurse(field.FieldType, entry)
			if err != nil {
				return nil, err
			}
			fields[fieldName] = converted
		}
		return &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil
	}
	if value.Kind() == reflect.Struct {
		for _, field := range def.Node.Fields {
			if field == nil || field.Name == nil {
				continue
			}
			fieldName := field.Name.Name
			entry := value.FieldByName(fieldName)
			if !entry.IsValid() {
				if field.FieldType != nil {
					if _, nullable := field.FieldType.(*ast.NullableTypeExpression); nullable {
						fields[fieldName] = runtime.NilValue{}
						continue
					}
				}
				return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
			}
			converted, err := recurse(field.FieldType, entry)
			if err != nil {
				return nil, err
			}
			fields[fieldName] = converted
		}
		return &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil
	}
	return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
}

func coerceStringValue(value runtime.Value) (string, error) {
	return AsString(value)
}

func toInt64(value runtime.Value) (int64, error) {
	return AsInt(value, 64)
}

func toUint64(value runtime.Value) (uint64, error) {
	return AsUint(value, 64)
}

func bigIntFromUint(value uint64) *big.Int {
	out := new(big.Int)
	out.SetUint64(value)
	return out
}
