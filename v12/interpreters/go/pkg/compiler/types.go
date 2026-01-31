package compiler

import "able/interpreter-go/pkg/ast"

type TypeMapper struct {
	structs map[string]*structInfo
}

func NewTypeMapper(structs map[string]*structInfo) *TypeMapper {
	return &TypeMapper{structs: structs}
}

func (m *TypeMapper) Map(expr ast.TypeExpression) (string, bool) {
	if expr == nil {
		return "", false
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
			case "Array", "HashMap", "Map", "DivMod":
				return "runtime.Value", true
			}
		}
		return "", false
	case *ast.FunctionTypeExpression:
		return "", false
	case *ast.NullableTypeExpression:
		return "", false
	case *ast.ResultTypeExpression:
		return "", false
	case *ast.UnionTypeExpression:
		return "", false
	case *ast.WildcardTypeExpression:
		return "", false
	default:
		return "", false
	}
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
	if info, ok := m.structs[name]; ok {
		return info.GoName, info.Supported
	}
	return "runtime.Value", false
}
