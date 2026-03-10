package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// arrayHandleLines generates the common preamble for array intrinsics:
// extract storage handle and convert to int64.
func (g *generator) arrayHandleLines(ctx *compileContext, objExpr string, callNode string) (lines []string, objTemp, handleTemp string) {
	objTemp = ctx.newTemp()
	handleRawTemp := ctx.newTemp()
	handleTemp = ctx.newTemp()
	handleErrTemp := ctx.newTemp()
	lines = []string{
		fmt.Sprintf("__able_push_call_frame(%s)", callNode),
		fmt.Sprintf("%s := %s", objTemp, objExpr),
		fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
		fmt.Sprintf("%s, %s := bridge.AsInt(%s, 64)", handleTemp, handleErrTemp, handleRawTemp),
		fmt.Sprintf("if %s != nil { panic(%s) }", handleErrTemp, handleErrTemp),
	}
	return lines, objTemp, handleTemp
}

// compileArrayMethodLenIntrinsic compiles arr.len() → int32 for all arrays.
func (g *generator) compileArrayMethodLenIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	lengthTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
	)
	resultTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := int32(%s)", resultTemp, lengthTemp))
	lines = append(lines, "__able_pop_call_frame()")
	if expected == "runtime.Value" {
		valueExpr, ok := g.runtimeValueExpr(resultTemp, "int32")
		if !ok {
			return nil, "", "", false
		}
		return lines, valueExpr, "runtime.Value", true
	}
	if expected == "" || expected == "int32" {
		return lines, resultTemp, "int32", true
	}
	valueExpr, ok := g.runtimeValueExpr(resultTemp, "int32")
	if !ok {
		return nil, "", "", false
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueExpr, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

// compileArrayMethodPushIntrinsic compiles arr.push(value) for all arrays.
// For mono arrays it uses typed writes; for non-mono it uses runtime.ArrayStoreWrite.
func (g *generator) compileArrayMethodPushIntrinsic(
	ctx *compileContext,
	objExpr string,
	monoKind monoArrayElemKind,
	monoEnabled bool,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 1 {
		return nil, "", "", false
	}
	// Mono path: use typed write.
	if monoEnabled {
		monoGoType := g.monoArrayElemGoType(monoKind)
		if monoGoType == "" {
			return nil, "", "", false
		}
		valArgLines, valueExpr, valueType, ok := g.compileExprLines(ctx, args[0], monoGoType)
		if !ok {
			return nil, "", "", false
		}
		coercedValueExpr, ok := g.coerceExprToGoType(valueExpr, valueType, monoGoType)
		if !ok {
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
		lengthTemp := ctx.newTemp()
		nextLenTemp := ctx.newTemp()
		capacityTemp := ctx.newTemp()
		writeExpr, ok := g.monoArrayWriteExpr(monoKind, handleTemp, lengthTemp, valueTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(valArgLines, append(lines, []string{
			fmt.Sprintf("%s := %s", valueTemp, coercedValueExpr),
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
			"__able_pop_call_frame()",
		}...)...)
		if expected == "runtime.Value" {
			return lines, "runtime.VoidValue{}", "runtime.Value", true
		}
		if expected == "" || expected == "struct{}" {
			return lines, "struct{}{}", "struct{}", true
		}
		return nil, "", "", false
	}
	// Non-mono path: use runtime.ArrayStoreWrite with runtime.Value.
	valArgLines, valueExpr, valueType, ok := g.compileExprLines(ctx, args[0], "")
	if !ok {
		return nil, "", "", false
	}
	valConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
	if !ok {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	lengthTemp := ctx.newTemp()
	nextLenTemp := ctx.newTemp()
	capacityTemp := ctx.newTemp()
	valueTemp := ctx.newTemp()
	lines = append(valArgLines, append(valConvLines, append(lines, []string{
		fmt.Sprintf("%s := %s", valueTemp, valueRuntime),
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("__able_panic_on_error(runtime.ArrayStoreWrite(%s, %s, %s))", handleTemp, lengthTemp, valueTemp),
		fmt.Sprintf("%s := %s + 1", nextLenTemp, lengthTemp),
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, nextLenTemp),
		fmt.Sprintf("if int(%s.Capacity) < %s {", objTemp, nextLenTemp),
		fmt.Sprintf("\t%s, err := runtime.ArrayStoreCapacity(%s)", capacityTemp, handleTemp),
		"\tif err != nil { panic(err) }",
		fmt.Sprintf("\t%s.Capacity = int32(%s)", objTemp, capacityTemp),
		"}",
		"__able_pop_call_frame()",
	}...)...)...)
	if expected == "runtime.Value" {
		return lines, "runtime.VoidValue{}", "runtime.Value", true
	}
	if expected == "" || expected == "struct{}" {
		return lines, "struct{}{}", "struct{}", true
	}
	return nil, "", "", false
}

// compileArrayMethodPopIntrinsic compiles arr.pop() → ?T for all arrays.
func (g *generator) compileArrayMethodPopIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	lengthTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	idxTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
		fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
		fmt.Sprintf("if %s > 0 {", lengthTemp),
		fmt.Sprintf("\t%s := %s - 1", idxTemp, lengthTemp),
		fmt.Sprintf("\t%s_read, err := runtime.ArrayStoreRead(%s, %s)", resultTemp, handleTemp, idxTemp),
		"\tif err != nil { panic(err) }",
		fmt.Sprintf("\tif %s_read != nil { %s = %s_read }", resultTemp, resultTemp, resultTemp),
		fmt.Sprintf("\t__able_panic_on_error(runtime.ArrayStoreSetLength(%s, %s))", handleTemp, idxTemp),
		fmt.Sprintf("\t%s.Length = int32(%s)", objTemp, idxTemp),
		"}",
		"__able_pop_call_frame()",
	)
	if expected == "" || expected == "runtime.Value" {
		return lines, resultTemp, "runtime.Value", true
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

// compileArrayMethodFirstLastIntrinsic compiles arr.first() or arr.last() → ?T.
func (g *generator) compileArrayMethodFirstLastIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
	isFirst bool,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	lengthTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
		fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
		fmt.Sprintf("if %s > 0 {", lengthTemp),
	)
	var idxExpr string
	if isFirst {
		idxExpr = "0"
	} else {
		idxTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("\t%s := %s - 1", idxTemp, lengthTemp))
		idxExpr = idxTemp
	}
	lines = append(lines,
		fmt.Sprintf("\t%s_read, err := runtime.ArrayStoreRead(%s, %s)", resultTemp, handleTemp, idxExpr),
		"\tif err != nil { panic(err) }",
		fmt.Sprintf("\tif %s_read != nil { %s = %s_read }", resultTemp, resultTemp, resultTemp),
		"}",
		"__able_pop_call_frame()",
	)
	if expected == "" || expected == "runtime.Value" {
		return lines, resultTemp, "runtime.Value", true
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

// compileArrayMethodIsEmptyIntrinsic compiles arr.is_empty() → bool.
func (g *generator) compileArrayMethodIsEmptyIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	lengthTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
		fmt.Sprintf("%s := %s == 0", resultTemp, lengthTemp),
		"__able_pop_call_frame()",
	)
	if expected == "" || expected == "bool" {
		return lines, resultTemp, "bool", true
	}
	if expected == "runtime.Value" {
		valueExpr, ok := g.runtimeValueExpr(resultTemp, "bool")
		if !ok {
			return nil, "", "", false
		}
		return lines, valueExpr, "runtime.Value", true
	}
	return nil, "", "", false
}

// compileArrayMethodClearIntrinsic compiles arr.clear() → void.
func (g *generator) compileArrayMethodClearIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	lines = append(lines,
		fmt.Sprintf("__able_panic_on_error(runtime.ArrayStoreSetLength(%s, 0))", handleTemp),
		fmt.Sprintf("%s.Length = int32(0)", objTemp),
		"__able_pop_call_frame()",
	)
	if expected == "runtime.Value" {
		return lines, "runtime.VoidValue{}", "runtime.Value", true
	}
	if expected == "" || expected == "struct{}" {
		return lines, "struct{}{}", "struct{}", true
	}
	return nil, "", "", false
}

// compileArrayMethodCapacityIntrinsic compiles arr.capacity() → int32.
func (g *generator) compileArrayMethodCapacityIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp, handleTemp := g.arrayHandleLines(ctx, objExpr, callNode)
	capacityTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s, err := runtime.ArrayStoreCapacity(%s)", capacityTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("%s.Capacity = int32(%s)", objTemp, capacityTemp),
		fmt.Sprintf("%s := int32(%s)", resultTemp, capacityTemp),
		"__able_pop_call_frame()",
	)
	if expected == "" || expected == "int32" {
		return lines, resultTemp, "int32", true
	}
	if expected == "runtime.Value" {
		valueExpr, ok := g.runtimeValueExpr(resultTemp, "int32")
		if !ok {
			return nil, "", "", false
		}
		return lines, valueExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}
