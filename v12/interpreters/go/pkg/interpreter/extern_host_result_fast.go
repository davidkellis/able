package interpreter

import (
	"reflect"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func externUnwrapHostReflectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Interface {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func externReflectStringResult(value reflect.Value) (runtime.Value, bool) {
	value = externUnwrapHostReflectValue(value)
	if !value.IsValid() || value.Kind() != reflect.String {
		return nil, false
	}
	return runtime.StringValue{Val: value.String()}, true
}

func externReflectStringSliceResult(value reflect.Value) (runtime.Value, bool) {
	value = externUnwrapHostReflectValue(value)
	if !value.IsValid() || value.Kind() != reflect.Slice || value.Type().Elem().Kind() != reflect.String {
		return nil, false
	}
	if stringsValue, ok := value.Interface().([]string); ok {
		return externStringSliceResult(stringsValue), true
	}
	length := value.Len()
	elements := make([]runtime.Value, length)
	for idx := 0; idx < length; idx++ {
		elements[idx] = runtime.StringValue{Val: value.Index(idx).String()}
	}
	return &runtime.ArrayValue{Elements: elements}, true
}

func externIsArrayStringType(expr ast.TypeExpression) bool {
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil {
		return false
	}
	return externSimpleTypeName(generic.Base) == "Array" &&
		len(generic.Arguments) == 1 &&
		externSimpleTypeName(generic.Arguments[0]) == "String"
}

func externUnionHasArrayStringMember(expr ast.TypeExpression) bool {
	union, ok := expr.(*ast.UnionTypeExpression)
	if !ok || union == nil {
		return false
	}
	for _, member := range union.Members {
		if externIsArrayStringType(member) {
			return true
		}
	}
	return false
}

func externUnionPreferredMemberForHostValue(union *ast.UnionTypeExpression, value reflect.Value) ast.TypeExpression {
	if union == nil {
		return nil
	}
	value = externUnwrapHostReflectValue(value)
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.String:
		for _, member := range union.Members {
			if externSimpleTypeName(member) == "String" {
				return member
			}
		}
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.String {
			for _, member := range union.Members {
				if externIsArrayStringType(member) {
					return member
				}
			}
		}
	}
	return nil
}
