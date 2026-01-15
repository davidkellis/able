package interpreter

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) toHostValue(typeExpr ast.TypeExpression, value runtime.Value, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}
	for {
		switch v := value.(type) {
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.InterfaceValue:
			if v != nil {
				value = v.Underlying
				continue
			}
		}
		break
	}
	hostVal, err := i.coerceRuntimeToHost(typeExpr, value, targetType)
	if err != nil {
		return reflect.Value{}, err
	}
	rv := reflect.ValueOf(hostVal)
	if !rv.IsValid() {
		return reflect.Zero(targetType), nil
	}
	if rv.Type().AssignableTo(targetType) {
		return rv, nil
	}
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType), nil
	}
	return reflect.Value{}, fmt.Errorf("extern argument cannot convert %s to %s", rv.Type(), targetType)
}

func (i *Interpreter) coerceRuntimeToHost(typeExpr ast.TypeExpression, value runtime.Value, targetType reflect.Type) (any, error) {
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
			typeExpr = expanded
		}
	} else {
		return i.coerceRuntimeByKind(value, targetType)
	}
	switch t := typeExpr.(type) {
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if member == nil {
				continue
			}
			if i.matchesType(member, value) {
				return i.coerceRuntimeToHost(member, value, targetType)
			}
		}
		return i.coerceRuntimeByKind(value, targetType)
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return nil, nil
		}
		if targetType.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("extern nullable type expects pointer target")
		}
		elemVal, err := i.coerceRuntimeToHost(t.InnerType, value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(targetType.Elem())
		if elemVal != nil {
			ev := reflect.ValueOf(elemVal)
			if ev.IsValid() && ev.Type().ConvertibleTo(targetType.Elem()) {
				ptr.Elem().Set(ev.Convert(targetType.Elem()))
			}
		}
		return ptr.Interface(), nil
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && normalizeKernelAliasName(base.Name.Name) == "Array" {
			arr, err := i.toArrayValue(value)
			if err != nil {
				return nil, err
			}
			elemType := targetType.Elem()
			slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
			var elemExpr ast.TypeExpression
			if len(t.Arguments) > 0 {
				elemExpr = t.Arguments[0]
			}
			for idx, elem := range arr.Elements {
				hostElem, err := i.coerceRuntimeToHost(elemExpr, elem, elemType)
				if err != nil {
					return nil, err
				}
				ev := reflect.ValueOf(hostElem)
				if ev.IsValid() && ev.Type().ConvertibleTo(elemType) {
					slice.Index(idx).Set(ev.Convert(elemType))
				}
			}
			return slice.Interface(), nil
		}
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
		switch name {
		case "String":
			return i.coerceStringValue(value)
		case "bool":
			if v, ok := value.(runtime.BoolValue); ok {
				return v.Val, nil
			}
		case "char":
			if v, ok := value.(runtime.CharValue); ok {
				return v.Val, nil
			}
		case "IoHandle", "ProcHandle":
			if hv, ok := value.(*runtime.HostHandleValue); ok && hv != nil {
				return hv.Value, nil
			}
			return nil, nil
		case "f32", "f64":
			if f, ok := value.(runtime.FloatValue); ok {
				return f.Val, nil
			}
			if f, ok := value.(*runtime.FloatValue); ok && f != nil {
				return f.Val, nil
			}
			if iv, ok := value.(runtime.IntegerValue); ok {
				return bigIntToFloat(iv.Val), nil
			}
			if iv, ok := value.(*runtime.IntegerValue); ok && iv != nil {
				return bigIntToFloat(iv.Val), nil
			}
		case "i8", "i16", "i32", "i64":
			return coerceIntValue(value, name)
		case "u8", "u16", "u32", "u64":
			return coerceUintValue(value, name)
		case "i128", "u128":
			return coerceBigInt(value)
		}
		if def, ok := i.lookupStructDefinition(name); ok {
			return i.structToHostValue(def, value)
		}
	}
	if targetType.Kind() == reflect.Interface {
		return value, nil
	}
	return nil, fmt.Errorf("unsupported extern argument type")
}

func (i *Interpreter) coerceRuntimeByKind(value runtime.Value, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.String:
		return i.coerceStringValue(value)
	case reflect.Bool:
		if v, ok := value.(runtime.BoolValue); ok {
			return v.Val, nil
		}
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
		if f, ok := value.(runtime.FloatValue); ok {
			return reflect.ValueOf(f.Val).Convert(targetType).Interface(), nil
		}
		if f, ok := value.(*runtime.FloatValue); ok && f != nil {
			return reflect.ValueOf(f.Val).Convert(targetType).Interface(), nil
		}
	case reflect.Interface:
		return value, nil
	case reflect.Pointer:
		if _, ok := value.(runtime.NilValue); ok {
			return nil, nil
		}
		elemVal, err := i.coerceRuntimeByKind(value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(targetType.Elem())
		ev := reflect.ValueOf(elemVal)
		if ev.IsValid() && ev.Type().ConvertibleTo(targetType.Elem()) {
			ptr.Elem().Set(ev.Convert(targetType.Elem()))
		}
		return ptr.Interface(), nil
	case reflect.Slice:
		arr, err := i.toArrayValue(value)
		if err != nil {
			return nil, err
		}
		elemType := targetType.Elem()
		slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
		for idx, elem := range arr.Elements {
			hostElem, err := i.coerceRuntimeByKind(elem, elemType)
			if err != nil {
				return nil, err
			}
			ev := reflect.ValueOf(hostElem)
			if ev.IsValid() && ev.Type().ConvertibleTo(elemType) {
				slice.Index(idx).Set(ev.Convert(elemType))
			}
		}
		return slice.Interface(), nil
	}
	return nil, fmt.Errorf("unsupported extern argument type")
}

func (i *Interpreter) fromHostResults(def *ast.ExternFunctionBody, results []reflect.Value) (runtime.Value, error) {
	if def == nil || def.Signature == nil {
		return runtime.NilValue{}, nil
	}
	if ret, ok := def.Signature.ReturnType.(*ast.ResultTypeExpression); ok {
		if len(results) != 2 {
			return nil, fmt.Errorf("extern result expects two return values")
		}
		errVal := results[1]
		if !errVal.IsNil() {
			if err, ok := errVal.Interface().(error); ok {
				return nil, raiseSignal{value: runtime.ErrorValue{Message: err.Error()}}
			}
			return nil, raiseSignal{value: runtime.ErrorValue{Message: "extern host error"}}
		}
		return i.fromHostValue(ret.InnerType, results[0])
	}
	if len(results) == 0 || def.Signature.ReturnType == nil {
		return runtime.VoidValue{}, nil
	}
	return i.fromHostValue(def.Signature.ReturnType, results[0])
}

func (i *Interpreter) fromHostValue(typeExpr ast.TypeExpression, value reflect.Value) (runtime.Value, error) {
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
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
		if simple, ok := typeExpr.(*ast.SimpleTypeExpression); ok && simple != nil {
			if normalizeKernelAliasName(simple.Name.Name) == "void" {
				return runtime.VoidValue{}, nil
			}
		}
		if _, ok := typeExpr.(*ast.UnionTypeExpression); !ok {
			return nil, fmt.Errorf("extern value is nil")
		}
	}
	switch t := typeExpr.(type) {
	case *ast.UnionTypeExpression:
		memberErrors := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			if member == nil {
				continue
			}
			converted, err := i.fromHostValue(member, value)
			if err == nil {
				return converted, nil
			}
			memberErrors = append(memberErrors, fmt.Sprintf("%s: %v", typeKey(member), err))
		}
		if len(memberErrors) > 0 {
			return nil, fmt.Errorf("unsupported extern return type %s (%s)", typeKey(t), strings.Join(memberErrors, "; "))
		}
		return runtime.NilValue{}, nil
	case *ast.NullableTypeExpression:
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return runtime.NilValue{}, nil
			}
			return i.fromHostValue(t.InnerType, value.Elem())
		}
		if value.Kind() == reflect.Invalid {
			return runtime.NilValue{}, nil
		}
		return i.fromHostValue(t.InnerType, value)
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && normalizeKernelAliasName(base.Name.Name) == "Array" {
			if value.Kind() != reflect.Slice {
				return nil, fmt.Errorf("extern expected slice result")
			}
			var elemExpr ast.TypeExpression
			if len(t.Arguments) > 0 {
				elemExpr = t.Arguments[0]
			}
			elements := make([]runtime.Value, value.Len())
			for idx := 0; idx < value.Len(); idx++ {
				elemVal, err := i.fromHostValue(elemExpr, value.Index(idx))
				if err != nil {
					return nil, err
				}
				elements[idx] = elemVal
			}
			return i.newArrayValue(elements, len(elements)), nil
		}
		return nil, fmt.Errorf("unsupported extern return type %s", typeKey(t))
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
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
				return runtime.IntegerValue{Val: bigIntFromInt(value.Int()), TypeSuffix: runtime.IntegerType(name)}, nil
			}
		case "u8", "u16", "u32", "u64":
			if value.Kind() >= reflect.Uint && value.Kind() <= reflect.Uint64 {
				return runtime.IntegerValue{Val: bigIntFromUint(value.Uint()), TypeSuffix: runtime.IntegerType(name)}, nil
			}
		case "i128", "u128":
			if value.Kind() == reflect.Pointer {
				if bi, ok := value.Interface().(*big.Int); ok && bi != nil {
					return runtime.IntegerValue{Val: new(big.Int).Set(bi), TypeSuffix: runtime.IntegerType(name)}, nil
				}
			}
		case "void":
			return runtime.VoidValue{}, nil
		}
		if def, ok := i.lookupStructDefinition(name); ok {
			return i.structFromHostValue(def, value)
		}
		if unionDef, ok := i.unionDefinitions[name]; ok && unionDef != nil && unionDef.Node != nil {
			unionExpr := &ast.UnionTypeExpression{Members: unionDef.Node.Variants}
			return i.fromHostValue(unionExpr, value)
		}
	}
	return nil, fmt.Errorf("unsupported extern return type %s", typeKey(typeExpr))
}

func (i *Interpreter) lookupStructDefinition(name string) (*runtime.StructDefinitionValue, bool) {
	if i == nil || i.global == nil {
		return nil, false
	}
	if def, ok := i.global.StructDefinition(name); ok && def != nil {
		return def, true
	}
	if val, err := i.global.Get(name); err == nil {
		if def, conv := toStructDefinitionValue(val, name); conv == nil && def != nil {
			return def, true
		}
	}
	if def, ok := i.lookupStructDefinitionInPackage(i.currentPackage, name); ok {
		return def, true
	}
	for pkgName := range i.packageRegistry {
		if pkgName == i.currentPackage {
			continue
		}
		if def, ok := i.lookupStructDefinitionInPackage(pkgName, name); ok {
			return def, true
		}
	}
	return nil, false
}

func (i *Interpreter) lookupStructDefinitionInPackage(pkgName, name string) (*runtime.StructDefinitionValue, bool) {
	if i == nil || name == "" || pkgName == "" {
		return nil, false
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok || bucket == nil {
		return nil, false
	}
	val, ok := bucket[name]
	if !ok {
		return nil, false
	}
	def, conv := toStructDefinitionValue(val, name)
	if conv == nil && def != nil {
		return def, true
	}
	return nil, false
}

func (i *Interpreter) structToHostValue(def *runtime.StructDefinitionValue, value runtime.Value) (any, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing")
	}
	if def.Node.Kind == ast.StructKindSingleton || len(def.Node.Fields) == 0 {
		return def.Node.ID.Name, nil
	}
	var inst *runtime.StructInstanceValue
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		inst = v
	case *runtime.StructDefinitionValue:
		if v != nil && v.Node != nil && v.Node.ID != nil && v.Node.ID.Name == def.Node.ID.Name {
			return def.Node.ID.Name, nil
		}
	}
	if inst == nil {
		return nil, fmt.Errorf("expected %s struct instance", def.Node.ID.Name)
	}
	if def.Node.Kind == ast.StructKindPositional && len(inst.Positional) > 0 {
		out := make([]any, len(inst.Positional))
		for idx, elem := range inst.Positional {
			var fieldType ast.TypeExpression
			if idx < len(def.Node.Fields) && def.Node.Fields[idx] != nil {
				fieldType = def.Node.Fields[idx].FieldType
			}
			hostElem, err := i.coerceRuntimeToHost(fieldType, elem, reflect.TypeOf((*any)(nil)).Elem())
			if err != nil {
				return nil, err
			}
			out[idx] = hostElem
		}
		return out, nil
	}
	if inst.Fields == nil {
		return nil, fmt.Errorf("expected %s struct fields", def.Node.ID.Name)
	}
	out := make(map[string]any)
	for _, field := range def.Node.Fields {
		if field == nil || field.Name == nil {
			continue
		}
		fieldName := field.Name.Name
		fieldVal, ok := inst.Fields[fieldName]
		if !ok {
			if field.FieldType != nil {
				if _, ok := field.FieldType.(*ast.NullableTypeExpression); ok {
					out[fieldName] = nil
					continue
				}
			}
			return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
		}
		hostVal, err := i.coerceRuntimeToHost(field.FieldType, fieldVal, reflect.TypeOf((*any)(nil)).Elem())
		if err != nil {
			return nil, err
		}
		out[fieldName] = hostVal
	}
	return out, nil
}

func (i *Interpreter) structFromHostValue(def *runtime.StructDefinitionValue, value reflect.Value) (runtime.Value, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing")
	}
	if value.Kind() == reflect.Invalid {
		return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
	}
	if def.Node.Kind == ast.StructKindSingleton || len(def.Node.Fields) == 0 {
		if value.Kind() == reflect.String && value.String() == def.Node.ID.Name {
			return &runtime.StructInstanceValue{Definition: def, Fields: map[string]runtime.Value{}}, nil
		}
		return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
	}
	if def.Node.Kind == ast.StructKindPositional && value.Kind() == reflect.Slice {
		positional := make([]runtime.Value, value.Len())
		for idx := 0; idx < value.Len(); idx++ {
			var fieldType ast.TypeExpression
			if idx < len(def.Node.Fields) && def.Node.Fields[idx] != nil {
				fieldType = def.Node.Fields[idx].FieldType
			}
			elem, err := i.fromHostValue(fieldType, value.Index(idx))
			if err != nil {
				return nil, err
			}
			positional[idx] = elem
		}
		return &runtime.StructInstanceValue{Definition: def, Positional: positional}, nil
	}
	fields := make(map[string]runtime.Value)
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
					if _, ok := field.FieldType.(*ast.NullableTypeExpression); ok {
						fields[fieldName] = runtime.NilValue{}
						continue
					}
				}
				return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
			}
			converted, err := i.fromHostValue(field.FieldType, entry)
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
					if _, ok := field.FieldType.(*ast.NullableTypeExpression); ok {
						fields[fieldName] = runtime.NilValue{}
						continue
					}
				}
				return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
			}
			converted, err := i.fromHostValue(field.FieldType, entry)
			if err != nil {
				return nil, err
			}
			fields[fieldName] = converted
		}
		return &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil
	}
	return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
}

func (i *Interpreter) coerceStringValue(value runtime.Value) (string, error) {
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val, nil
	case *runtime.StringValue:
		if v == nil {
			return "", fmt.Errorf("string value is nil")
		}
		return v.Val, nil
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return "", fmt.Errorf("string value is invalid")
		}
		if v.Definition.Node.ID.Name != "String" {
			return "", fmt.Errorf("expected String struct")
		}
		var bytesVal runtime.Value
		if v.Fields != nil {
			bytesVal = v.Fields["bytes"]
		}
		if bytesVal == nil && len(v.Positional) > 0 {
			bytesVal = v.Positional[0]
		}
		arr, err := i.toArrayValue(bytesVal)
		if err != nil {
			return "", err
		}
		if _, err := i.ensureArrayState(arr, 0); err != nil {
			return "", err
		}
		buf := make([]byte, len(arr.Elements))
		for idx, elem := range arr.Elements {
			num, err := toInt64(elem)
			if err != nil {
				return "", err
			}
			if num < 0 || num > 0xff {
				return "", fmt.Errorf("string byte out of range")
			}
			buf[idx] = byte(num)
		}
		return string(buf), nil
	}
	return "", fmt.Errorf("expected String value")
}

func coerceIntValue(value runtime.Value, kind string) (any, error) {
	num, err := toInt64(value)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "i8":
		return int8(num), nil
	case "i16":
		return int16(num), nil
	case "i32":
		return int32(num), nil
	case "i64":
		return int64(num), nil
	}
	return num, nil
}

func coerceUintValue(value runtime.Value, kind string) (any, error) {
	num, err := toUint64(value)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "u8":
		return uint8(num), nil
	case "u16":
		return uint16(num), nil
	case "u32":
		return uint32(num), nil
	case "u64":
		return uint64(num), nil
	}
	return num, nil
}

func coerceBigInt(value runtime.Value) (any, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val == nil {
			return nil, fmt.Errorf("integer is nil")
		}
		return new(big.Int).Set(v.Val), nil
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil {
			return nil, fmt.Errorf("integer is nil")
		}
		return new(big.Int).Set(v.Val), nil
	}
	return nil, fmt.Errorf("expected integer")
}

func toInt64(value runtime.Value) (int64, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val != nil && v.Val.IsInt64() {
			return v.Val.Int64(), nil
		}
	case *runtime.IntegerValue:
		if v != nil && v.Val != nil && v.Val.IsInt64() {
			return v.Val.Int64(), nil
		}
	}
	return 0, fmt.Errorf("expected integer")
}

func toUint64(value runtime.Value) (uint64, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val != nil && v.Val.IsUint64() {
			return v.Val.Uint64(), nil
		}
	case *runtime.IntegerValue:
		if v != nil && v.Val != nil && v.Val.IsUint64() {
			return v.Val.Uint64(), nil
		}
	}
	return 0, fmt.Errorf("expected unsigned integer")
}

func bigIntFromInt(val int64) *big.Int {
	return big.NewInt(val)
}

func bigIntFromUint(val uint64) *big.Int {
	b := new(big.Int)
	b.SetUint64(val)
	return b
}
