package compiler

type nativeNullableSpec struct {
	PtrType    string
	InnerType  string
	HelperStem string
}

var nativeNullableSpecs = []nativeNullableSpec{
	{PtrType: "*bool", InnerType: "bool", HelperStem: "bool"},
	{PtrType: "*string", InnerType: "string", HelperStem: "string"},
	{PtrType: "*runtime.ErrorValue", InnerType: "runtime.ErrorValue", HelperStem: "error"},
	{PtrType: "*rune", InnerType: "rune", HelperStem: "char"},
	{PtrType: "*float32", InnerType: "float32", HelperStem: "f32"},
	{PtrType: "*float64", InnerType: "float64", HelperStem: "f64"},
	{PtrType: "*int", InnerType: "int", HelperStem: "isize"},
	{PtrType: "*uint", InnerType: "uint", HelperStem: "usize"},
	{PtrType: "*int8", InnerType: "int8", HelperStem: "i8"},
	{PtrType: "*int16", InnerType: "int16", HelperStem: "i16"},
	{PtrType: "*int32", InnerType: "int32", HelperStem: "i32"},
	{PtrType: "*int64", InnerType: "int64", HelperStem: "i64"},
	{PtrType: "*uint8", InnerType: "uint8", HelperStem: "u8"},
	{PtrType: "*uint16", InnerType: "uint16", HelperStem: "u16"},
	{PtrType: "*uint32", InnerType: "uint32", HelperStem: "u32"},
	{PtrType: "*uint64", InnerType: "uint64", HelperStem: "u64"},
}

func nativeNullableSpecForPointer(goType string) (nativeNullableSpec, bool) {
	for _, spec := range nativeNullableSpecs {
		if spec.PtrType == goType {
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
	spec, ok := nativeNullableSpecForPointer(goType)
	if !ok {
		return "", false
	}
	return spec.InnerType, true
}

func (g *generator) nativeNullablePointerType(innerType string) (string, bool) {
	spec, ok := nativeNullableSpecForInnerType(innerType)
	if !ok {
		return "", false
	}
	return spec.PtrType, true
}

func (g *generator) isNativeNullableValueType(goType string) bool {
	_, ok := nativeNullableSpecForPointer(goType)
	return ok
}

func (g *generator) nativeNullableWraps(expected, actual string) bool {
	inner, ok := g.nativeNullableValueInnerType(expected)
	return ok && inner == actual
}

func (g *generator) nativeNullableFromRuntimeHelper(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForPointer(goType)
	if !ok {
		return "", false
	}
	return "__able_nullable_" + spec.HelperStem + "_from_value", true
}

func (g *generator) nativeNullableFromRuntimePanicHelper(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForPointer(goType)
	if !ok {
		return "", false
	}
	return "__able_nullable_" + spec.HelperStem + "_from_value_or_panic", true
}

func (g *generator) nativeNullableToRuntimeHelper(goType string) (string, bool) {
	spec, ok := nativeNullableSpecForPointer(goType)
	if !ok {
		return "", false
	}
	return "__able_nullable_" + spec.HelperStem + "_to_value", true
}
