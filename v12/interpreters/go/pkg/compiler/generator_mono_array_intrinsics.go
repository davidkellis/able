package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// arrayObjLines generates the common preamble for array intrinsics:
// capture the array object pointer.
func (g *generator) arrayObjLines(ctx *compileContext, objExpr string, callNode string) (lines []string, objTemp string) {
	_ = callNode
	objTemp = ctx.newTemp()
	lines = []string{
		fmt.Sprintf("%s := %s", objTemp, objExpr),
	}
	return lines, objTemp
}

func (g *generator) compileStaticArrayValueArg(
	ctx *compileContext,
	arrayType string,
	arg ast.Expression,
) ([]string, string, bool) {
	elemType := g.staticArrayElemGoType(arrayType)
	if elemType == "" {
		return nil, "", false
	}
	argLines, argExpr, argGoType, ok := g.compileExprLines(ctx, arg, elemType)
	if !ok {
		return nil, "", false
	}
	if elemType == "runtime.Value" {
		valLines, valueExpr, ok := g.lowerRuntimeValue(ctx, argExpr, argGoType)
		if !ok {
			return nil, "", false
		}
		return append(argLines, valLines...), valueExpr, true
	}
	return argLines, argExpr, true
}

// compileArrayMethodLenIntrinsic compiles arr.len() → int32 for all arrays.
func (g *generator) compileArrayMethodLenIntrinsic(
	ctx *compileContext,
	objExpr string,
	objType string,
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
		fmt.Sprintf("%s := int32(%s)", resultTemp, g.staticArrayLengthExpr(objTemp)),
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
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, valueExpr, expected)
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
	objType string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 1 {
		return nil, "", "", false
	}
	valArgLines, valueExpr, ok := g.compileStaticArrayValueArg(ctx, objType, args[0])
	if !ok {
		return nil, "", "", false
	}
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	valueTemp := ctx.newTemp()
	lines = append(valArgLines, append(lines, []string{
		fmt.Sprintf("%s := %s", valueTemp, valueExpr),
		fmt.Sprintf("%s.Elements = append(%s.Elements, %s)", objTemp, objTemp, valueTemp),
		g.staticArraySyncCall(objType, objTemp),
	}...)...)
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
	objType string,
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
	resultType := expected
	if resultType == "" {
		resultType = "runtime.Value"
	}
	elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s-1]", objTemp, lengthTemp), resultType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)))
	if resultType == "runtime.Value" {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp))
	} else {
		lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, resultType))
	}
	lines = append(lines, fmt.Sprintf("if %s > 0 {", lengthTemp))
	lines = append(lines, indentLines(elemLines, 1)...)
	lines = append(lines,
		fmt.Sprintf("\t%s = %s", resultTemp, elemExpr),
		fmt.Sprintf("\t%s.Elements = %s.Elements[:%s-1]", objTemp, objTemp, lengthTemp),
		"}",
		g.staticArraySyncCall(objType, objTemp),
	)
	return lines, resultTemp, resultType, true
}

// compileArrayMethodFirstLastIntrinsic compiles arr.first() or arr.last() → ?T.
func (g *generator) compileArrayMethodFirstLastIntrinsic(
	ctx *compileContext,
	objExpr string,
	objType string,
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
	resultType := expected
	if resultType == "" {
		resultType = "runtime.Value"
	}
	elemExprSource := fmt.Sprintf("%s.Elements[0]", objTemp)
	if !isFirst {
		elemExprSource = fmt.Sprintf("%s.Elements[%s-1]", objTemp, lengthTemp)
	}
	elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, elemExprSource, resultType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)))
	if resultType == "runtime.Value" {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp))
	} else {
		lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, resultType))
	}
	lines = append(lines, fmt.Sprintf("if %s > 0 {", lengthTemp))
	lines = append(lines, indentLines(elemLines, 1)...)
	lines = append(lines,
		fmt.Sprintf("\t%s = %s", resultTemp, elemExpr),
		"}",
	)
	return lines, resultTemp, resultType, true
}

// compileArrayMethodIsEmptyIntrinsic compiles arr.is_empty() → bool.
func (g *generator) compileArrayMethodIsEmptyIntrinsic(
	ctx *compileContext,
	objExpr string,
	objType string,
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
		fmt.Sprintf("%s := %s == 0", resultTemp, g.staticArrayLengthExpr(objTemp)),
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
	objType string,
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
		g.staticArraySyncCall(objType, objTemp),
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
	objType string,
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
		fmt.Sprintf("%s := int32(%s)", resultTemp, g.staticArrayCapacityExpr(objTemp)),
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
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

func (g *generator) compileArrayMethodReserveIntrinsic(
	ctx *compileContext,
	objExpr string,
	objType string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 1 {
		return nil, "", "", false
	}
	capLines, capExpr, capType, ok := g.compileExprLines(ctx, args[0], "")
	if !ok {
		return nil, "", "", false
	}
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	capTemp := ctx.newTemp()
	idxTemp := ctx.newTemp()
	capacityTemp := ctx.newTemp()
	lines = append(capLines, lines...)
	lines = append(lines, fmt.Sprintf("%s := %s", capTemp, capExpr))
	lines, ok = g.appendIndexIntLines(ctx, lines, capTemp, capType, idxTemp, capacityTemp)
	if !ok {
		return nil, "", "", false
	}
	elemType := g.staticArrayElemGoType(objType)
	if elemType == "" {
		return nil, "", "", false
	}
	lines = append(lines,
		fmt.Sprintf("if %s > %s {", capacityTemp, g.staticArrayCapacityExpr(objTemp)),
		fmt.Sprintf("\t__able_reserved := make([]%s, %s, %s)", elemType, g.staticArrayLengthExpr(objTemp), capacityTemp),
		fmt.Sprintf("\tcopy(__able_reserved, %s.Elements)", objTemp),
		fmt.Sprintf("\t%s.Elements = __able_reserved", objTemp),
		"}",
		g.staticArraySyncCall(objType, objTemp),
	)
	if expected == "runtime.Value" {
		return lines, "runtime.VoidValue{}", "runtime.Value", true
	}
	if expected == "" || expected == "struct{}" {
		return lines, "struct{}{}", "struct{}", true
	}
	return nil, "", "", false
}

func (g *generator) compileArrayMethodCloneShallowIntrinsic(
	ctx *compileContext,
	objExpr string,
	objType string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if len(args) != 0 {
		return nil, "", "", false
	}
	lines, objTemp := g.arrayObjLines(ctx, objExpr, callNode)
	cloneLines, cloneExpr, ok := g.staticArrayCloneLines(ctx, objType, fmt.Sprintf("%s.Elements", objTemp), g.staticArrayCapacityExpr(objTemp))
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, cloneLines...)
	if expected == "" || g.typeMatches(expected, objType) {
		return lines, cloneExpr, objType, true
	}
	valueLines, valueExpr, ok := g.lowerRuntimeValue(ctx, cloneExpr, objType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, valueLines...)
	if expected == "runtime.Value" {
		return lines, valueExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, valueExpr, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}
