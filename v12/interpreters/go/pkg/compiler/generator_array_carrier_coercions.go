package compiler

import (
	"fmt"
	"strings"
)

func (g *generator) staticArrayCarrierCoercible(expected, actual string) bool {
	if g == nil || expected == "" || actual == "" {
		return false
	}
	if expected == actual {
		return true
	}
	if !g.isStaticArrayType(expected) || !g.isStaticArrayType(actual) {
		return false
	}
	return expected == "*Array" || actual == "*Array"
}

func (g *generator) staticArrayToRuntimeLines(ctx *compileContext, expr string, actual string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" || actual == "" {
		return nil, "", false
	}
	valueTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	var lines []string
	switch {
	case actual == "*Array":
		lines = []string{fmt.Sprintf("%s, %s := __able_struct_Array_to(__able_runtime, %s)", valueTemp, errTemp, expr)}
	case g.isMonoArrayType(actual):
		spec, ok := g.monoArraySpecForGoType(actual)
		if !ok || spec == nil {
			return nil, "", false
		}
		lines = []string{fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", valueTemp, errTemp, spec.ToRuntimeHelper, expr)}
	default:
		return nil, "", false
	}
	lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, valueTemp, true
}

func (g *generator) staticArrayFromRuntimeLines(ctx *compileContext, valueExpr string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || valueExpr == "" || expected == "" {
		return nil, "", false
	}
	if expected == "*Array" {
		return g.directRuntimeArrayValueToGenericArrayLines(ctx, valueExpr)
	}
	resultTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	var lines []string
	switch {
	case g.isMonoArrayType(expected):
		spec, ok := g.monoArraySpecForGoType(expected)
		if !ok || spec == nil {
			return nil, "", false
		}
		lines = []string{fmt.Sprintf("%s, %s := %s(%s)", resultTemp, errTemp, spec.FromRuntimeHelper, valueExpr)}
	default:
		return nil, "", false
	}
	lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp, true
}

func (g *generator) directRuntimeArrayValueToGenericArrayLines(ctx *compileContext, valueExpr string) ([]string, string, bool) {
	if g == nil || ctx == nil || valueExpr == "" {
		return nil, "", false
	}
	currentTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	stateTemp := ctx.newTemp()
	stateErrTemp := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := __able_unwrap_interface(%s)", currentTemp, valueExpr),
		fmt.Sprintf("var %s *Array", resultTemp),
		fmt.Sprintf("var %s error", errTemp),
		fmt.Sprintf("if _, isNil := %s.(runtime.NilValue); isNil {", currentTemp),
		fmt.Sprintf("\t%s = nil", resultTemp),
		fmt.Sprintf("} else if raw, ok, nilPtr := __able_runtime_array_value(%s); ok || nilPtr {", currentTemp),
		"\tif !ok || nilPtr {",
		fmt.Sprintf("\t\t%s = fmt.Errorf(\"expected Array value\")", errTemp),
		"\t} else if raw.Handle != 0 {",
		fmt.Sprintf("\t\t%s, %s := runtime.ArrayStoreState(raw.Handle)", stateTemp, stateErrTemp),
		fmt.Sprintf("\t\tif %s != nil {", stateErrTemp),
		fmt.Sprintf("\t\t\t%s = %s", errTemp, stateErrTemp),
		"\t\t} else {",
		fmt.Sprintf("\t\t\t%s := make([]runtime.Value, len(%s.Values), %s.Capacity)", valuesTemp, stateTemp, stateTemp),
		fmt.Sprintf("\t\t\tcopy(%s, %s.Values)", valuesTemp, stateTemp),
		fmt.Sprintf("\t\t\t%s = &Array{Storage_handle: raw.Handle, Elements: %s}", resultTemp, valuesTemp),
		fmt.Sprintf("\t\t\t__able_struct_Array_sync(%s)", resultTemp),
		"\t\t}",
		"\t} else {",
		fmt.Sprintf("\t\t%s := make([]runtime.Value, len(raw.Elements), cap(raw.Elements))", valuesTemp),
		fmt.Sprintf("\t\tcopy(%s, raw.Elements)", valuesTemp),
		fmt.Sprintf("\t\t%s = &Array{Storage_handle: raw.Handle, Elements: %s}", resultTemp, valuesTemp),
		fmt.Sprintf("\t\t__able_struct_Array_sync(%s)", resultTemp),
		"\t}",
		"} else {",
		fmt.Sprintf("\t%s = fmt.Errorf(\"expected Array value\")", errTemp),
		"}",
		fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
	}
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp, true
}

func (g *generator) runtimeValueToGenericArrayBoundaryLines(targetVar string, errVar string, valueExpr string, allowStructFallback bool) []string {
	if g == nil || targetVar == "" || errVar == "" || valueExpr == "" {
		return nil
	}
	base := sanitizeIdent(targetVar)
	if base == "" {
		base = "array"
	}
	currentTemp := base + "_current"
	stateTemp := base + "_state"
	stateErrTemp := base + "_state_err"
	valuesTemp := base + "_values"
	structValuesTemp := base + "_struct_values"
	structHandleTemp := base + "_struct_handle"
	structLengthTemp := base + "_struct_length"
	structCapacityTemp := base + "_struct_capacity"
	structErrTemp := base + "_struct_err"
	lines := []string{
		fmt.Sprintf("%s := __able_unwrap_interface(%s)", currentTemp, valueExpr),
		fmt.Sprintf("if _, isNil := %s.(runtime.NilValue); isNil {", currentTemp),
		fmt.Sprintf("\t%s = nil", targetVar),
		fmt.Sprintf("} else if raw, ok, nilPtr := __able_runtime_array_value(%s); ok || nilPtr {", currentTemp),
		"\tif !ok || nilPtr {",
		fmt.Sprintf("\t\t%s = fmt.Errorf(\"expected Array value\")", errVar),
		"\t} else if raw.Handle != 0 {",
		fmt.Sprintf("\t\t%s, %s := runtime.ArrayStoreState(raw.Handle)", stateTemp, stateErrTemp),
		fmt.Sprintf("\t\tif %s != nil {", stateErrTemp),
		fmt.Sprintf("\t\t\t%s = %s", errVar, stateErrTemp),
		"\t\t} else {",
		fmt.Sprintf("\t\t\t%s := make([]runtime.Value, len(%s.Values), %s.Capacity)", valuesTemp, stateTemp, stateTemp),
		fmt.Sprintf("\t\t\tcopy(%s, %s.Values)", valuesTemp, stateTemp),
		fmt.Sprintf("\t\t\t%s = &Array{Storage_handle: raw.Handle, Elements: %s}", targetVar, valuesTemp),
		fmt.Sprintf("\t\t\t__able_struct_Array_sync(%s)", targetVar),
		"\t\t}",
		"\t} else {",
		fmt.Sprintf("\t\t%s := make([]runtime.Value, len(raw.Elements), cap(raw.Elements))", valuesTemp),
		fmt.Sprintf("\t\tcopy(%s, raw.Elements)", valuesTemp),
		fmt.Sprintf("\t\t%s = &Array{Storage_handle: raw.Handle, Elements: %s}", targetVar, valuesTemp),
		fmt.Sprintf("\t\t__able_struct_Array_sync(%s)", targetVar),
		"\t}",
	}
	if allowStructFallback {
		lines = append(lines,
			fmt.Sprintf("} else if inst, ok := %s.(*runtime.StructInstanceValue); ok && inst != nil {", currentTemp),
			fmt.Sprintf("\t%s, %s, %s, %s, %s := __able_array_struct_instance_state(inst)", structValuesTemp, structHandleTemp, structLengthTemp, structCapacityTemp, structErrTemp),
			fmt.Sprintf("\tif %s != nil {", structErrTemp),
			fmt.Sprintf("\t\t%s = %s", errVar, structErrTemp),
			"\t} else {",
			fmt.Sprintf("\t\t%s = &Array{Length: %s, Capacity: %s, Storage_handle: %s, Elements: %s}", targetVar, structLengthTemp, structCapacityTemp, structHandleTemp, structValuesTemp),
			fmt.Sprintf("\t\t__able_struct_Array_sync(%s)", targetVar),
			"\t}",
			"} else {",
			fmt.Sprintf("\t%s = fmt.Errorf(\"expected Array value\")", errVar),
			"}",
		)
		return lines
	}
	lines = append(lines,
		"} else {",
		fmt.Sprintf("\t%s = fmt.Errorf(\"expected Array value\")", errVar),
		"}",
	)
	return lines
}

func (g *generator) runtimeValueToGenericArrayPanicExpr(valueExpr string, allowStructFallback bool) string {
	if g == nil || valueExpr == "" {
		return ""
	}
	lines := []string{
		fmt.Sprintf("value := %s", valueExpr),
		"var result *Array",
		"var err error",
	}
	lines = append(lines, g.runtimeValueToGenericArrayBoundaryLines("result", "err", "value", allowStructFallback)...)
	lines = append(lines,
		"if err != nil { panic(err) }",
		"return result",
	)
	return fmt.Sprintf("func() *Array {\n%s\n}()", strings.Join(indentLines(lines, 1), "\n"))
}

func (g *generator) coerceStaticArrayCarrierLines(ctx *compileContext, expr string, actual string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" || !g.staticArrayCarrierCoercible(expected, actual) {
		return nil, "", false
	}
	if actual == expected {
		return nil, expr, true
	}
	if directLines, converted, ok := g.directStaticArrayCarrierCoercionLines(ctx, expr, actual, expected); ok {
		return directLines, converted, true
	}
	toRuntimeLines, runtimeExpr, ok := g.staticArrayToRuntimeLines(ctx, expr, actual)
	if !ok {
		return nil, "", false
	}
	fromRuntimeLines, converted, ok := g.staticArrayFromRuntimeLines(ctx, runtimeExpr, expected)
	if !ok {
		return nil, "", false
	}
	lines := append(toRuntimeLines, fromRuntimeLines...)
	return lines, converted, true
}

func (g *generator) directStaticArrayCarrierCoercionLines(ctx *compileContext, expr string, actual string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" {
		return nil, "", false
	}
	switch {
	case expected == "*Array" && g.isMonoArrayType(actual):
		return g.directMonoArrayToGenericArrayLines(ctx, expr, actual)
	case actual == "*Array" && g.isMonoArrayType(expected):
		return g.directGenericArrayToMonoArrayLines(ctx, expr, expected)
	default:
		return nil, "", false
	}
}

func (g *generator) directMonoArrayToGenericArrayLines(ctx *compileContext, expr string, actual string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" {
		return nil, "", false
	}
	spec, ok := g.monoArraySpecForGoType(actual)
	if !ok || spec == nil {
		return nil, "", false
	}
	sourceTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	indexTemp := ctx.newTemp()
	rawTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", sourceTemp, expr),
		fmt.Sprintf("var %s *Array", resultTemp),
		fmt.Sprintf("if %s != nil {", sourceTemp),
		fmt.Sprintf("\t%s := make([]runtime.Value, %s, %s)", valuesTemp, g.staticArrayLengthExpr(sourceTemp), g.staticArrayCapacityExpr(sourceTemp)),
		fmt.Sprintf("\tfor %s, %s := range %s.Elements {", indexTemp, rawTemp, sourceTemp),
	}
	valueLines, valueExpr, ok := g.lowerRuntimeValue(ctx, rawTemp, spec.ElemGoType)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, indentLines(valueLines, 2)...)
	lines = append(lines, fmt.Sprintf("\t\t%s[%s] = %s", valuesTemp, indexTemp, valueExpr))
	lines = append(lines,
		"\t}",
		fmt.Sprintf("\t%s = &Array{Elements: %s}", resultTemp, valuesTemp),
		fmt.Sprintf("\t__able_struct_Array_sync(%s)", resultTemp),
		"}",
	)
	return lines, resultTemp, true
}

func (g *generator) directGenericArrayToMonoArrayLines(ctx *compileContext, expr string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" {
		return nil, "", false
	}
	spec, ok := g.monoArraySpecForGoType(expected)
	if !ok || spec == nil {
		return nil, "", false
	}
	sourceTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	indexTemp := ctx.newTemp()
	rawTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", sourceTemp, expr),
		fmt.Sprintf("var %s %s", resultTemp, expected),
		fmt.Sprintf("if %s != nil {", sourceTemp),
		fmt.Sprintf("\t%s := make([]%s, %s, %s)", valuesTemp, spec.ElemGoType, g.staticArrayLengthExpr(sourceTemp), g.staticArrayCapacityExpr(sourceTemp)),
		fmt.Sprintf("\tfor %s, %s := range %s.Elements {", indexTemp, rawTemp, sourceTemp),
	}
	valueLines, valueExpr, ok := g.lowerExpectRuntimeValue(ctx, rawTemp, spec.ElemGoType)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, indentLines(valueLines, 2)...)
	lines = append(lines, fmt.Sprintf("\t\t%s[%s] = %s", valuesTemp, indexTemp, valueExpr))
	lines = append(lines,
		"\t}",
		fmt.Sprintf("\t%s = &%s{Elements: %s}", resultTemp, spec.GoName, valuesTemp),
		"}",
	)
	return lines, resultTemp, true
}
