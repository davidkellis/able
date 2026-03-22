package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type monoArraySpec struct {
	Kind              monoArrayElemKind
	Suffix            string
	GoName            string
	GoType            string
	ElemGoType        string
	FromRuntimeHelper string
	ToRuntimeHelper   string
	SyncHelper        string
}

var stagedMonoArraySpecs = []monoArraySpec{
	{
		Kind:              monoArrayElemKindI8,
		Suffix:            "i8",
		GoName:            "__able_array_i8",
		GoType:            "*__able_array_i8",
		ElemGoType:        "int8",
		FromRuntimeHelper: "__able_array_i8_from",
		ToRuntimeHelper:   "__able_array_i8_to",
		SyncHelper:        "__able_array_i8_sync",
	},
	{
		Kind:              monoArrayElemKindI16,
		Suffix:            "i16",
		GoName:            "__able_array_i16",
		GoType:            "*__able_array_i16",
		ElemGoType:        "int16",
		FromRuntimeHelper: "__able_array_i16_from",
		ToRuntimeHelper:   "__able_array_i16_to",
		SyncHelper:        "__able_array_i16_sync",
	},
	{
		Kind:              monoArrayElemKindI32,
		Suffix:            "i32",
		GoName:            "__able_array_i32",
		GoType:            "*__able_array_i32",
		ElemGoType:        "int32",
		FromRuntimeHelper: "__able_array_i32_from",
		ToRuntimeHelper:   "__able_array_i32_to",
		SyncHelper:        "__able_array_i32_sync",
	},
	{
		Kind:              monoArrayElemKindI64,
		Suffix:            "i64",
		GoName:            "__able_array_i64",
		GoType:            "*__able_array_i64",
		ElemGoType:        "int64",
		FromRuntimeHelper: "__able_array_i64_from",
		ToRuntimeHelper:   "__able_array_i64_to",
		SyncHelper:        "__able_array_i64_sync",
	},
	{
		Kind:              monoArrayElemKindU16,
		Suffix:            "u16",
		GoName:            "__able_array_u16",
		GoType:            "*__able_array_u16",
		ElemGoType:        "uint16",
		FromRuntimeHelper: "__able_array_u16_from",
		ToRuntimeHelper:   "__able_array_u16_to",
		SyncHelper:        "__able_array_u16_sync",
	},
	{
		Kind:              monoArrayElemKindU32,
		Suffix:            "u32",
		GoName:            "__able_array_u32",
		GoType:            "*__able_array_u32",
		ElemGoType:        "uint32",
		FromRuntimeHelper: "__able_array_u32_from",
		ToRuntimeHelper:   "__able_array_u32_to",
		SyncHelper:        "__able_array_u32_sync",
	},
	{
		Kind:              monoArrayElemKindU64,
		Suffix:            "u64",
		GoName:            "__able_array_u64",
		GoType:            "*__able_array_u64",
		ElemGoType:        "uint64",
		FromRuntimeHelper: "__able_array_u64_from",
		ToRuntimeHelper:   "__able_array_u64_to",
		SyncHelper:        "__able_array_u64_sync",
	},
	{
		Kind:              monoArrayElemKindISize,
		Suffix:            "isize",
		GoName:            "__able_array_isize",
		GoType:            "*__able_array_isize",
		ElemGoType:        "int",
		FromRuntimeHelper: "__able_array_isize_from",
		ToRuntimeHelper:   "__able_array_isize_to",
		SyncHelper:        "__able_array_isize_sync",
	},
	{
		Kind:              monoArrayElemKindUSize,
		Suffix:            "usize",
		GoName:            "__able_array_usize",
		GoType:            "*__able_array_usize",
		ElemGoType:        "uint",
		FromRuntimeHelper: "__able_array_usize_from",
		ToRuntimeHelper:   "__able_array_usize_to",
		SyncHelper:        "__able_array_usize_sync",
	},
	{
		Kind:              monoArrayElemKindF32,
		Suffix:            "f32",
		GoName:            "__able_array_f32",
		GoType:            "*__able_array_f32",
		ElemGoType:        "float32",
		FromRuntimeHelper: "__able_array_f32_from",
		ToRuntimeHelper:   "__able_array_f32_to",
		SyncHelper:        "__able_array_f32_sync",
	},
	{
		Kind:              monoArrayElemKindF64,
		Suffix:            "f64",
		GoName:            "__able_array_f64",
		GoType:            "*__able_array_f64",
		ElemGoType:        "float64",
		FromRuntimeHelper: "__able_array_f64_from",
		ToRuntimeHelper:   "__able_array_f64_to",
		SyncHelper:        "__able_array_f64_sync",
	},
	{
		Kind:              monoArrayElemKindBool,
		Suffix:            "bool",
		GoName:            "__able_array_bool",
		GoType:            "*__able_array_bool",
		ElemGoType:        "bool",
		FromRuntimeHelper: "__able_array_bool_from",
		ToRuntimeHelper:   "__able_array_bool_to",
		SyncHelper:        "__able_array_bool_sync",
	},
	{
		Kind:              monoArrayElemKindU8,
		Suffix:            "u8",
		GoName:            "__able_array_u8",
		GoType:            "*__able_array_u8",
		ElemGoType:        "uint8",
		FromRuntimeHelper: "__able_array_u8_from",
		ToRuntimeHelper:   "__able_array_u8_to",
		SyncHelper:        "__able_array_u8_sync",
	},
	{
		Kind:              monoArrayElemKindChar,
		Suffix:            "char",
		GoName:            "__able_array_char",
		GoType:            "*__able_array_char",
		ElemGoType:        "rune",
		FromRuntimeHelper: "__able_array_char_from",
		ToRuntimeHelper:   "__able_array_char_to",
		SyncHelper:        "__able_array_char_sync",
	},
	{
		Kind:              monoArrayElemKindString,
		Suffix:            "String",
		GoName:            "__able_array_String",
		GoType:            "*__able_array_String",
		ElemGoType:        "string",
		FromRuntimeHelper: "__able_array_String_from",
		ToRuntimeHelper:   "__able_array_String_to",
		SyncHelper:        "__able_array_String_sync",
	},
}

func (g *generator) monoArraySpecForGoType(goType string) (*monoArraySpec, bool) {
	if g == nil || goType == "" {
		return nil, false
	}
	for i := range stagedMonoArraySpecs {
		spec := &stagedMonoArraySpecs[i]
		if goType == spec.GoType || goType == spec.GoName {
			return spec, true
		}
	}
	for _, spec := range g.monoArraySpecs {
		if spec != nil && (goType == spec.GoType || goType == spec.GoName) {
			return spec, true
		}
	}
	return nil, false
}

func (g *generator) monoArrayGenericArrayHelper(goType string) (string, bool) {
	spec, ok := g.monoArraySpecForGoType(goType)
	if !ok || spec == nil {
		return "", false
	}
	return spec.GoName + "_as_Array", true
}

func (g *generator) monoArraySpecForElementGoType(goType string) (*monoArraySpec, bool) {
	if g == nil {
		return nil, false
	}
	if g.monoArraysEnabled() {
		for i := range stagedMonoArraySpecs {
			spec := &stagedMonoArraySpecs[i]
			if spec.ElemGoType == goType {
				return spec, true
			}
		}
	}
	for _, spec := range g.monoArraySpecs {
		if spec != nil && spec.ElemGoType == goType {
			return spec, true
		}
	}
	return g.ensureCarrierMonoArraySpec(goType)
}

func (g *generator) sortedMonoArraySpecs() []monoArraySpec {
	if g == nil {
		return nil
	}
	specs := make([]monoArraySpec, 0, len(stagedMonoArraySpecs)+len(g.monoArraySpecs))
	if g.monoArraysEnabled() {
		specs = append(specs, stagedMonoArraySpecs...)
	}
	if len(g.monoArraySpecs) == 0 {
		return specs
	}
	keys := make([]string, 0, len(g.monoArraySpecs))
	for key := range g.monoArraySpecs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if spec := g.monoArraySpecs[key]; spec != nil {
			specs = append(specs, *spec)
		}
	}
	return specs
}

func (g *generator) canSynthesizeCarrierMonoArraySpec(elemGoType string) bool {
	if g == nil || elemGoType == "" || elemGoType == "runtime.Value" || elemGoType == "any" {
		return false
	}
	if g.isMonoArrayType(elemGoType) {
		return true
	}
	if g.isArrayStructType(elemGoType) {
		return true
	}
	if g.isNativeStructPointerType(elemGoType) {
		return true
	}
	if g.nativeInterfaceInfoForGoType(elemGoType) != nil {
		return true
	}
	if g.nativeCallableInfoForGoType(elemGoType) != nil {
		return true
	}
	if g.nativeUnionInfoForGoType(elemGoType) != nil {
		return true
	}
	if g.isNativeNullableValueType(elemGoType) {
		return true
	}
	return false
}

func (g *generator) ensureCarrierMonoArraySpec(elemGoType string) (*monoArraySpec, bool) {
	if g == nil || !g.canSynthesizeCarrierMonoArraySpec(elemGoType) {
		return nil, false
	}
	for _, spec := range g.monoArraySpecs {
		if spec != nil && spec.ElemGoType == elemGoType {
			return spec, true
		}
	}
	token := sanitizeIdent(strings.TrimPrefix(elemGoType, "*"))
	token = strings.TrimPrefix(token, "__able_")
	token = strings.Trim(token, "_")
	if token == "" {
		token = "carrier"
	}
	goName := "__able_array_" + token
	spec := &monoArraySpec{
		Kind:              monoArrayElemKindUnknown,
		Suffix:            token,
		GoName:            goName,
		GoType:            "*" + goName,
		ElemGoType:        elemGoType,
		FromRuntimeHelper: goName + "_from",
		ToRuntimeHelper:   goName + "_to",
		SyncHelper:        goName + "_sync",
	}
	g.monoArraySpecs[spec.GoType] = spec
	return spec, true
}

func (g *generator) monoArraySpecForArrayTypeExpr(pkgName string, expr ast.TypeExpression) (*monoArraySpec, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	gen, ok := expr.(*ast.GenericTypeExpression)
	if !ok || gen == nil {
		return nil, false
	}
	base, ok := gen.Base.(*ast.SimpleTypeExpression)
	if !ok || base == nil || base.Name == nil || base.Name.Name != "Array" || len(gen.Arguments) != 1 {
		return nil, false
	}
	elemGoType, ok := g.lowerCarrierTypeInPackage(pkgName, gen.Arguments[0])
	if !ok {
		return nil, false
	}
	return g.monoArraySpecForElementGoType(elemGoType)
}

func (g *generator) isMonoArrayType(goType string) bool {
	_, ok := g.monoArraySpecForGoType(goType)
	return ok
}

func (g *generator) isStaticArrayType(goType string) bool {
	return g.isArrayStructType(goType) || g.isMonoArrayType(goType)
}

func (g *generator) staticArrayElemGoType(goType string) string {
	if spec, ok := g.monoArraySpecForGoType(goType); ok {
		return spec.ElemGoType
	}
	if g.isArrayStructType(goType) {
		return "runtime.Value"
	}
	return ""
}

func (g *generator) staticArraySyncCall(goType string, valueExpr string) string {
	if spec, ok := g.monoArraySpecForGoType(goType); ok {
		return fmt.Sprintf("%s(%s)", spec.SyncHelper, valueExpr)
	}
	return fmt.Sprintf("__able_struct_Array_sync(%s)", valueExpr)
}

func (g *generator) staticArrayArgRequiresValue(pkgName string, typeExpr ast.TypeExpression, goType string) bool {
	if g == nil || !g.isMonoArrayType(goType) {
		return false
	}
	if typeExpr == nil {
		return true
	}
	if _, ok := typeExpr.(*ast.NullableTypeExpression); ok {
		return false
	}
	if _, members, ok := g.expandedUnionMembersInPackage(pkgName, typeExpr); ok {
		if _, nullable := nativeUnionNullableInnerTypeExpr(members); nullable {
			return false
		}
	}
	return true
}

func (g *generator) staticArrayElementRuntimeExpr(arrayType string, elemExpr string) (string, bool) {
	if spec, ok := g.monoArraySpecForGoType(arrayType); ok {
		if runtimeExpr, ok := g.runtimeValueExpr(elemExpr, spec.ElemGoType); ok {
			return runtimeExpr, true
		}
	}
	if g.isArrayStructType(arrayType) {
		return elemExpr, true
	}
	return "", false
}

func (g *generator) staticArrayResultValueExpr(arrayType string, elemExpr string) string {
	if g.isArrayStructType(arrayType) {
		return fmt.Sprintf("func() runtime.Value { if __v := %s; __v != nil { return __v }; return runtime.NilValue{} }()", elemExpr)
	}
	expr, ok := g.staticArrayElementRuntimeExpr(arrayType, elemExpr)
	if !ok {
		return elemExpr
	}
	return expr
}

func (g *generator) staticArrayZeroValueExpr(arrayType string) string {
	if spec, ok := g.monoArraySpecForGoType(arrayType); ok {
		if g.isNilableStaticCarrierType(spec.ElemGoType) {
			return "nil"
		}
		switch spec.Kind {
		case monoArrayElemKindBool:
			return "false"
		case monoArrayElemKindF32:
			return "float32(0)"
		case monoArrayElemKindF64:
			return "float64(0)"
		case monoArrayElemKindI8,
			monoArrayElemKindI16,
			monoArrayElemKindI32,
			monoArrayElemKindI64,
			monoArrayElemKindU8,
			monoArrayElemKindU16,
			monoArrayElemKindU32,
			monoArrayElemKindU64,
			monoArrayElemKindISize,
			monoArrayElemKindUSize,
			monoArrayElemKindChar:
			return spec.ElemGoType + "(0)"
		case monoArrayElemKindString:
			return `""`
		}
	}
	return "runtime.NilValue{}"
}

func (g *generator) staticArrayCoerceValueExprLines(ctx *compileContext, arrayType string, expr string, actualType string) ([]string, string, bool) {
	elemType := g.staticArrayElemGoType(arrayType)
	if elemType == "" {
		return nil, "", false
	}
	if elemType == "runtime.Value" {
		lines, converted, ok := g.lowerRuntimeValue(ctx, expr, actualType)
		return lines, converted, ok
	}
	lines, converted, _, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, expr, actualType, elemType)
	return lines, converted, ok
}
