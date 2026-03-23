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
	if m != nil && m.gen != nil {
		expr = normalizeTypeExprForPackage(m.gen, m.packageName, expr)
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
			}
			if goType, ok := m.gen.nativeStructCarrierTypeForExpr(m.packageName, t); ok {
				return goType, true
			}
			if info, ok := m.gen.ensureNativeInterfaceInfo(m.packageName, t); ok && info != nil {
				return info.GoType, true
			}
			if unionPkg, members, ok := m.gen.expandedUnionMembersInPackage(m.packageName, t); ok {
				return m.mapExpandedUnionMembers(unionPkg, t, members)
			}
		}
		if m != nil && m.gen != nil && m.gen.typeExprIsConcreteInPackage(m.packageName, t) {
			return "", false
		}
		return "runtime.Value", true
	case *ast.FunctionTypeExpression:
		if info, ok := m.gen.ensureNativeCallableInfo(m.packageName, t); ok && info != nil {
			return info.GoType, true
		}
		if m != nil && m.gen != nil && m.gen.typeExprIsConcreteInPackage(m.packageName, t) {
			return "", false
		}
		return "any", true
	case *ast.NullableTypeExpression:
		return m.mapNullableType(t)
	case *ast.ResultTypeExpression:
		return m.mapResultType(t)
	case *ast.UnionTypeExpression:
		return m.mapUnionType(t)
	case *ast.WildcardTypeExpression:
		return "any", true
	default:
		return "any", false
	}
}

func (m *TypeMapper) mapResultType(t *ast.ResultTypeExpression) (string, bool) {
	if t == nil || m == nil || m.gen == nil {
		return "", false
	}
	if info, ok := m.gen.ensureNativeResultUnionInfo(m.packageName, t); ok && info != nil {
		return info.GoType, true
	}
	if m.gen.typeExprIsConcreteInPackage(m.packageName, t) {
		return "", false
	}
	return "any", true
}

// mapArrayType maps Array<T>. Currently returns the existing Array struct
// pointer. TODO: monomorphize to []ElemGoType once slice intrinsics are ready.
func (m *TypeMapper) mapArrayType(t *ast.GenericTypeExpression) (string, bool) {
	if m != nil && m.gen != nil {
		if spec, ok := m.gen.monoArraySpecForArrayTypeExpr(m.packageName, t); ok && spec != nil {
			return spec.GoType, true
		}
	}
	if m != nil && m.gen != nil {
		if goType, ok := m.gen.nativeStructCarrierType(m.packageName, "Array"); ok {
			return goType, true
		}
	}
	if m != nil && m.gen != nil && m.gen.typeExprIsConcreteInPackage(m.packageName, t) {
		return "", false
	}
	return "any", true
}

// mapNullableType maps ?T. Pointer and slice types already have a nil carrier.
// Native scalar nullable values use typed Go pointers instead of any.
func (m *TypeMapper) mapNullableType(t *ast.NullableTypeExpression) (string, bool) {
	if t == nil || t.InnerType == nil {
		return "", false
	}
	innerType, ok := m.Map(t.InnerType)
	if !ok {
		return "", false
	}
	// Struct pointers already have a nil zero value.
	if strings.HasPrefix(innerType, "*") {
		return innerType, true
	}
	// Slices also have a nil zero value.
	if strings.HasPrefix(innerType, "[]") {
		return innerType, true
	}
	if m != nil && m.gen != nil && m.gen.goTypeHasNilZeroValue(innerType) {
		return innerType, true
	}
	if spec, ok := nativeNullableSpecForInnerType(innerType); ok {
		return spec.PtrType, true
	}
	if m != nil && m.gen != nil && m.gen.typeExprIsConcreteInPackage(m.packageName, t) {
		return "", false
	}
	return "any", true
}

func (m *TypeMapper) mapUnionType(t *ast.UnionTypeExpression) (string, bool) {
	if t == nil || m == nil || m.gen == nil {
		return "", false
	}
	return m.mapExpandedUnionMembers(m.packageName, t, t.Members)
}

func (m *TypeMapper) mapExpandedUnionMembers(pkgName string, expr ast.TypeExpression, members []ast.TypeExpression) (string, bool) {
	if m == nil || m.gen == nil || expr == nil {
		return "", false
	}
	if inner, ok := nativeUnionNullableInnerTypeExpr(members); ok {
		return (&TypeMapper{gen: m.gen, packageName: pkgName}).mapNullableType(ast.NewNullableTypeExpression(inner))
	}
	if info, ok := m.gen.nativeUnionTypeExprInPackage(pkgName, expr); ok && info != nil {
		return info.GoType, true
	}
	if m.gen.typeExprIsConcreteInPackage(pkgName, expr) {
		return "", false
	}
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
	case "Error":
		return "runtime.ErrorValue", true
	}
	if m != nil && m.gen != nil {
		if info, ok := m.gen.ensureNativeInterfaceInfo(m.packageName, ast.Ty(name)); ok && info != nil {
			return info.GoType, true
		}
	}
	if m != nil && m.gen != nil {
		if unionPkg, members, ok := m.gen.expandedUnionMembersInPackage(m.packageName, ast.Ty(name)); ok {
			if inner, ok := nativeUnionNullableInnerTypeExpr(members); ok {
				return (&TypeMapper{gen: m.gen, packageName: unionPkg}).mapNullableType(ast.NewNullableTypeExpression(inner))
			}
			if info, ok := m.gen.ensureNativeUnionInfo(unionPkg, members); ok && info != nil {
				return info.GoType, true
			}
		}
	}
	if m != nil && m.gen != nil {
		if goType, ok := m.gen.nativeStructCarrierType(m.packageName, name); ok {
			return goType, true
		}
	}
	return "runtime.Value", true
}
