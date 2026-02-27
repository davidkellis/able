package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileMonoArrayMethodLenIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) (string, string, bool) {
	if len(args) != 0 {
		return "", "", false
	}
	objTemp := ctx.newTemp()
	handleRawTemp := ctx.newTemp()
	handleTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("__able_push_call_frame(%s)", callNode),
		"defer __able_pop_call_frame()",
		fmt.Sprintf("%s := %s", objTemp, objExpr),
		fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
		fmt.Sprintf("%s := func() int64 { raw, err := bridge.AsInt(%s, 64); if err != nil { panic(err) }; return raw }()", handleTemp, handleRawTemp),
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
	}
	if expected == "runtime.Value" {
		valueExpr, ok := g.runtimeValueExpr(lengthTemp, "int32")
		if !ok {
			return "", "", false
		}
		return g.wrapLinesAsExpression(ctx, lines, valueExpr, "runtime.Value")
	}
	lengthExpr, lengthType, ok := g.wrapLinesAsExpression(ctx, lines, fmt.Sprintf("int32(%s)", lengthTemp), "int32")
	if !ok {
		return "", "", false
	}
	if expected == "" || expected == "int32" {
		return lengthExpr, lengthType, true
	}
	valueExpr, ok := g.runtimeValueExpr(lengthExpr, "int32")
	if !ok {
		return "", "", false
	}
	converted, ok := g.expectRuntimeValueExpr(valueExpr, expected)
	if !ok {
		return "", "", false
	}
	return converted, expected, true
}

func (g *generator) compileMonoArrayMethodPushIntrinsic(
	ctx *compileContext,
	objExpr string,
	monoKind monoArrayElemKind,
	args []ast.Expression,
	expected string,
	callNode string,
) (string, string, bool) {
	if len(args) != 1 {
		return "", "", false
	}
	monoGoType := g.monoArrayElemGoType(monoKind)
	if monoGoType == "" {
		return "", "", false
	}
	valueExpr, valueType, ok := g.compileExpr(ctx, args[0], monoGoType)
	if !ok {
		return "", "", false
	}
	coercedValueExpr, ok := g.coerceExprToGoType(valueExpr, valueType, monoGoType)
	if !ok {
		return "", "", false
	}
	objTemp := ctx.newTemp()
	valueTemp := ctx.newTemp()
	handleRawTemp := ctx.newTemp()
	handleTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	nextLenTemp := ctx.newTemp()
	capacityTemp := ctx.newTemp()
	writeExpr, ok := g.monoArrayWriteExpr(monoKind, handleTemp, lengthTemp, valueTemp)
	if !ok {
		return "", "", false
	}
	lines := []string{
		fmt.Sprintf("__able_push_call_frame(%s)", callNode),
		"defer __able_pop_call_frame()",
		fmt.Sprintf("%s := %s", objTemp, objExpr),
		fmt.Sprintf("%s := %s", valueTemp, coercedValueExpr),
		fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
		fmt.Sprintf("%s := func() int64 { raw, err := bridge.AsInt(%s, 64); if err != nil { panic(err) }; return raw }()", handleTemp, handleRawTemp),
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("__able_panic_on_error(%s)", writeExpr),
		fmt.Sprintf("%s := %s + 1", nextLenTemp, lengthTemp),
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, nextLenTemp),
		fmt.Sprintf("if int(%s.Capacity) < %s {", objTemp, nextLenTemp),
		fmt.Sprintf("\t%s, err := runtime.ArrayStoreCapacity(%s)", capacityTemp, handleTemp),
		"\tif err != nil { panic(err) }",
		fmt.Sprintf("\t%s.Capacity = int32(%s)", objTemp, capacityTemp),
		"}",
	}
	if expected == "runtime.Value" {
		return g.wrapLinesAsExpression(ctx, lines, "runtime.VoidValue{}", "runtime.Value")
	}
	if expected == "" || expected == "struct{}" {
		return g.wrapLinesAsExpression(ctx, lines, "struct{}{}", "struct{}")
	}
	return "", "", false
}
