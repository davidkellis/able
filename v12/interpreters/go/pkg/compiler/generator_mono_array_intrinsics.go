package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// arrayObjLines generates the common preamble for array intrinsics:
// capture the array object pointer.
func (g *generator) arrayObjLines(ctx *compileContext, objExpr string, callNode string) (lines []string, objTemp string) {
	objTemp = ctx.newTemp()
	lines = []string{
		fmt.Sprintf("__able_push_call_frame(%s)", callNode),
		fmt.Sprintf("%s := %s", objTemp, objExpr),
	}
	return lines, objTemp
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
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	resultTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := int32(len(%s.Elements))", resultTemp, objTemp),
		"__able_pop_call_frame()",
	)
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
func (g *generator) compileArrayMethodPushIntrinsic(
	ctx *compileContext,
	objExpr string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 1 {
		return nil, "", "", false
	}
	// Compile the value argument and convert to runtime.Value.
	valArgLines, valueExpr, valueType, ok := g.compileExprLines(ctx, args[0], "")
	if !ok {
		return nil, "", "", false
	}
	valConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
	if !ok {
		return nil, "", "", false
	}
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	valueTemp := ctx.newTemp()
	lines = append(valArgLines, append(valConvLines, append(lines, []string{
		fmt.Sprintf("%s := %s", valueTemp, valueRuntime),
		fmt.Sprintf("%s.Elements = append(%s.Elements, %s)", objTemp, objTemp, valueTemp),
		fmt.Sprintf("__able_struct_Array_sync(%s)", objTemp),
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
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	resultTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := len(%s.Elements)", lengthTemp, objTemp),
		fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
		fmt.Sprintf("if %s > 0 {", lengthTemp),
		fmt.Sprintf("\t%s = %s.Elements[%s-1]", resultTemp, objTemp, lengthTemp),
		fmt.Sprintf("\t%s.Elements = %s.Elements[:%s-1]", objTemp, objTemp, lengthTemp),
		"}",
		fmt.Sprintf("__able_struct_Array_sync(%s)", objTemp),
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
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	resultTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := len(%s.Elements)", lengthTemp, objTemp),
		fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
		fmt.Sprintf("if %s > 0 {", lengthTemp),
	)
	if isFirst {
		lines = append(lines, fmt.Sprintf("\t%s = %s.Elements[0]", resultTemp, objTemp))
	} else {
		lines = append(lines, fmt.Sprintf("\t%s = %s.Elements[%s-1]", resultTemp, objTemp, lengthTemp))
	}
	lines = append(lines,
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
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	resultTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := len(%s.Elements) == 0", resultTemp, objTemp),
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
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	lines = append(lines,
		fmt.Sprintf("%s.Elements = %s.Elements[:0]", objTemp, objTemp),
		fmt.Sprintf("__able_struct_Array_sync(%s)", objTemp),
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
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	resultTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := int32(cap(%s.Elements))", resultTemp, objTemp),
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
