package compiler

import "fmt"

type nativeNullableSpec struct {
	CarrierType  string
	InnerType    string
	HelperStem   string
	ValueCarrier bool
}

var nativeNullableSpecs = []nativeNullableSpec{
	{CarrierType: "__able_nullable[bool]", InnerType: "bool", HelperStem: "bool", ValueCarrier: true},
	{CarrierType: "__able_nullable[string]", InnerType: "string", HelperStem: "string", ValueCarrier: true},
	{CarrierType: "*runtime.ErrorValue", InnerType: "runtime.ErrorValue", HelperStem: "error"},
	{CarrierType: "__able_nullable[rune]", InnerType: "rune", HelperStem: "char", ValueCarrier: true},
	{CarrierType: "__able_nullable[float32]", InnerType: "float32", HelperStem: "f32", ValueCarrier: true},
	{CarrierType: "__able_nullable[float64]", InnerType: "float64", HelperStem: "f64", ValueCarrier: true},
	{CarrierType: "__able_nullable[int]", InnerType: "int", HelperStem: "isize", ValueCarrier: true},
	{CarrierType: "__able_nullable[uint]", InnerType: "uint", HelperStem: "usize", ValueCarrier: true},
	{CarrierType: "__able_nullable[int8]", InnerType: "int8", HelperStem: "i8", ValueCarrier: true},
	{CarrierType: "__able_nullable[int16]", InnerType: "int16", HelperStem: "i16", ValueCarrier: true},
	{CarrierType: "__able_nullable[int32]", InnerType: "int32", HelperStem: "i32", ValueCarrier: true},
	{CarrierType: "__able_nullable[int64]", InnerType: "int64", HelperStem: "i64", ValueCarrier: true},
	{CarrierType: "__able_nullable[uint8]", InnerType: "uint8", HelperStem: "u8", ValueCarrier: true},
	{CarrierType: "__able_nullable[uint16]", InnerType: "uint16", HelperStem: "u16", ValueCarrier: true},
	{CarrierType: "__able_nullable[uint32]", InnerType: "uint32", HelperStem: "u32", ValueCarrier: true},
	{CarrierType: "__able_nullable[uint64]", InnerType: "uint64", HelperStem: "u64", ValueCarrier: true},
	{CarrierType: "__able_nullable[runtime.Int128]", InnerType: "runtime.Int128", HelperStem: "i128", ValueCarrier: true},
	{CarrierType: "__able_nullable[runtime.Uint128]", InnerType: "runtime.Uint128", HelperStem: "u128", ValueCarrier: true},
}

func nativeNullableSpecForCarrier(goType string) (nativeNullableSpec, bool) {
	for _, spec := range nativeNullableSpecs {
		if spec.CarrierType == goType {
			return spec, true
		}
	}
	return nativeNullableSpec{}, false
}

func nativeNullableSpecForInnerType(goType string) (nativeNullableSpec, bool) {
	for _, spec := range nativeNullableSpecs {
		if spec.InnerType == goType {
			return spec, true
		}
	}
	return nativeNullableSpec{}, false
}

func (g *generator) nativeNullableValueInnerType(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	return spec.InnerType, true
}

func (g *generator) nativeNullableCarrierType(innerType string) (string, bool) {
	spec, ok := nativeNullableSpecForInnerType(innerType)
	if !ok {
		return "", false
	}
	return spec.CarrierType, true
}

func (g *generator) isNativeNullableValueType(goType string) bool {
	_, ok := nativeNullableSpecForCarrier(goType)
	return ok
}

func (g *generator) nativeNullableWraps(expected, actual string) bool {
	inner, ok := g.nativeNullableValueInnerType(expected)
	return ok && inner == actual
}

func (g *generator) nativeNullableFromRuntimeHelper(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	return "__able_nullable_" + spec.HelperStem + "_from_value", true
}

func (g *generator) nativeNullableFromRuntimePanicHelper(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	return "__able_nullable_" + spec.HelperStem + "_from_value_or_panic", true
}

func (g *generator) nativeNullableToRuntimeHelper(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	return "__able_nullable_" + spec.HelperStem + "_to_value", true
}

func (g *generator) nativeNullableAbsentExpr(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	if spec.ValueCarrier {
		return spec.CarrierType + "{}", true
	}
	return fmt.Sprintf("(%s)(nil)", spec.CarrierType), true
}

func (g *generator) nativeNullablePresentExpr(goType, expr string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	if spec.ValueCarrier {
		return fmt.Sprintf("__able_some(%s)", expr), true
	}
	return fmt.Sprintf("__able_ptr(%s)", expr), true
}

func (g *generator) nativeNullableHasValueExpr(goType, expr string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	if spec.ValueCarrier {
		return fmt.Sprintf("(%s).valid", expr), true
	}
	return fmt.Sprintf("(%s != nil)", expr), true
}

func (g *generator) nativeNullableIsNilExpr(goType, expr string) (string, bool) {
	present, ok := g.nativeNullableHasValueExpr(goType, expr)
	if !ok {
		return "", false
	}
	return "(!" + present + ")", true
}

func (g *generator) nativeNullableValueExpr(goType, expr string) (string, bool) {
	spec, ok := nativeNullableSpecForCarrier(goType)
	if !ok {
		return "", false
	}
	if spec.ValueCarrier {
		return fmt.Sprintf("(%s).value", expr), true
	}
	return fmt.Sprintf("(*%s)", expr), true
}
