package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

type TypeMapper struct {
	gen         *generator
	packageName string
}

func NewTypeMapper(gen *generator, packageName string) *TypeMapper {
	return &TypeMapper{gen: gen, packageName: packageName}
}

func (m *TypeMapper) Map(expr ast.TypeExpression) (string, bool) {
	if expr == nil {
		return "any", true
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return "", false
		}
		return m.mapSimple(t.Name.Name)
	case *ast.GenericTypeExpression:
		if t == nil {
			return "", false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			switch base.Name.Name {
			case "Array":
				return m.mapArrayType(t)
			case "HashMap", "Map", "DivMod":
				return "any", true
			}
		}
		// Generic struct types (TreeMap<K,V>, etc.) keep runtime.Value
		// so that self-as-runtime.Value field access works correctly.
		return "runtime.Value", true
	case *ast.FunctionTypeExpression:
		return "any", true
	case *ast.NullableTypeExpression:
		return m.mapNullableType(t)
	case *ast.ResultTypeExpression:
		return "any", true
	case *ast.UnionTypeExpression:
		return "any", true
	case *ast.WildcardTypeExpression:
		return "any", true
	default:
		return "any", false
	}
}

// mapArrayType maps Array<T>. Currently returns the existing Array struct
// pointer. TODO: monomorphize to []ElemGoType once slice intrinsics are ready.
func (m *TypeMapper) mapArrayType(t *ast.GenericTypeExpression) (string, bool) {
	if m != nil && m.gen != nil {
		if info, ok := m.gen.structInfoForTypeName(m.packageName, "Array"); ok && info != nil {
			return "*" + info.GoName, true
		}
		if info, ok := m.gen.structInfoByNameUnique("Array"); ok && info != nil {
			return "*" + info.GoName, true
		}
	}
	return "any", true
}

// mapNullableType maps ?T. For pointer types (structs), nil is the absent
// value so the Go type is just the pointer. For value types and any, use any.
func (m *TypeMapper) mapNullableType(t *ast.NullableTypeExpression) (string, bool) {
	if t == nil || t.InnerType == nil {
		return "any", true
	}
	innerType, ok := m.Map(t.InnerType)
	if !ok {
		return "any", true
	}
	// Struct pointers already have a nil zero value.
	if strings.HasPrefix(innerType, "*") {
		return innerType, true
	}
	// Slices also have a nil zero value.
	if strings.HasPrefix(innerType, "[]") {
		return innerType, true
	}
	// For value types (int32, string, bool, etc.), nullable requires any
	// since the value type itself has no nil representation.
	return "any", true
}

func (m *TypeMapper) mapSimple(name string) (string, bool) {
	switch name {
	case "bool", "Bool":
		return "bool", true
	case "String":
		return "string", true
	case "string":
		return "string", true
	case "char", "Char":
		return "rune", true
	case "i8":
		return "int8", true
	case "i16":
		return "int16", true
	case "i32":
		return "int32", true
	case "i64":
		return "int64", true
	case "u8":
		return "uint8", true
	case "u16":
		return "uint16", true
	case "u32":
		return "uint32", true
	case "u64":
		return "uint64", true
	case "isize":
		return "int", true
	case "usize":
		return "uint", true
	case "f32":
		return "float32", true
	case "f64":
		return "float64", true
	case "void", "Void":
		return "struct{}", true
	}
	if m != nil && m.gen != nil {
		if info, ok := m.gen.structInfoForTypeName(m.packageName, name); ok && info != nil {
			return "*" + info.GoName, true
		}
	}
	if m != nil && m.gen != nil {
		if info, ok := m.gen.structInfoByNameUnique(name); ok && info != nil {
			return "*" + info.GoName, true
		}
	}
	return "runtime.Value", true
}
