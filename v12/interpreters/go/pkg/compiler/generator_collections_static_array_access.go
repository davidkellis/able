package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileIndexExpression(ctx *compileContext, expr *ast.IndexExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing index expression")
		return nil, "", "", false
	}
	objLines, objExpr, objType, ok := g.compileExprLines(ctx, expr.Object, "")
	if !ok {
		return nil, "", "", false
	}
	if recoverLines, recoveredExpr, recoveredType, recovered := g.recoverDispatchExpr(ctx, expr.Object, objExpr, objType); recovered {
		objLines = append(objLines, recoverLines...)
		objExpr = recoveredExpr
		objType = recoveredType
	}
	if g.isStaticArrayType(objType) {
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, expr.Index, "")
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		lengthTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := append([]string{}, objLines...)
		lines = append(lines, idxLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			ctx.setReason("index expression unsupported")
			return nil, "", "", false
		}
		resultType := "runtime.Value"
		if expected != "" && expected != "runtime.Value" && expected != "any" {
			resultType = expected
		}
		elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexTemp), resultType)
		if !ok {
			ctx.setReason("index expression unsupported")
			return nil, "", "", false
		}
		if resultType == "runtime.Value" {
			lines = append(lines,
				fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)),
				fmt.Sprintf("var %s runtime.Value", resultTemp),
				fmt.Sprintf("if %s < 0 || %s >= %s { %s = __able_index_error(%s, %s) } else {", indexTemp, indexTemp, lengthTemp, resultTemp, indexTemp, lengthTemp),
			)
			lines = append(lines, indentLines(elemLines, 1)...)
			lines = append(lines,
				fmt.Sprintf("\t%s = %s", resultTemp, elemExpr),
				"}",
			)
			return lines, resultTemp, "runtime.Value", true
		}
		transferLines, ok := g.lowerControlTransfer(ctx, g.raiseControlExpr("nil", fmt.Sprintf("__able_error_value(__able_index_error(%s, %s))", indexTemp, lengthTemp)))
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)),
			fmt.Sprintf("var %s %s", resultTemp, resultType),
			fmt.Sprintf("if %s < 0 || %s >= %s {", indexTemp, indexTemp, lengthTemp),
		)
		lines = append(lines, indentLines(transferLines, 1)...)
		lines = append(lines, "} else {")
		lines = append(lines, indentLines(elemLines, 1)...)
		lines = append(lines,
			fmt.Sprintf("\t%s = %s", resultTemp, elemExpr),
			"}",
		)
		return lines, resultTemp, resultType, true
	}
	if staticLines, staticExpr, staticType, ok := g.compileStaticIndexGet(ctx, expr, expected, objExpr, objType); ok {
		lines := append([]string{}, objLines...)
		lines = append(lines, staticLines...)
		return lines, staticExpr, staticType, true
	}
	objConvLines, objValue, ok := g.lowerRuntimeValue(ctx, objExpr, objType)
	if !ok {
		ctx.setReason("index object unsupported")
		return nil, "", "", false
	}
	idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, expr.Index, "")
	if !ok {
		return nil, "", "", false
	}
	idxConvLines, idxValue, ok := g.lowerRuntimeValue(ctx, idxExpr, idxType)
	if !ok {
		ctx.setReason("index expression unsupported")
		return nil, "", "", false
	}
	lines := append([]string{}, objLines...)
	lines = append(lines, objConvLines...)
	lines = append(lines, idxLines...)
	lines = append(lines, idxConvLines...)
	baseTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := __able_index(%s, %s)", baseTemp, controlTemp, objValue, idxValue))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	baseExpr := baseTemp
	if expected == "" || expected == "runtime.Value" {
		return lines, baseExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, baseExpr, expected)
	if !ok {
		ctx.setReason("index expression type mismatch")
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

func (g *generator) compileArrayMethodIntrinsicCall(
	ctx *compileContext,
	objNode ast.Expression,
	objExpr string,
	objType string,
	methodName string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if g == nil || ctx == nil || !g.isStaticArrayType(objType) {
		return nil, "", "", false
	}
	effectiveExpected := expected
	if effectiveExpected == "" {
		switch methodName {
		case "get", "pop", "first", "last", "read_slot":
			if inferred, ok := g.staticArrayDefaultNullableResultType(ctx, objNode, objType); ok {
				effectiveExpected = inferred
			}
		}
	}
	switch methodName {
	case "len":
		return g.compileArrayMethodLenIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "push":
		return g.compileArrayMethodPushIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "pop":
		return g.compileArrayMethodPopIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "first":
		return g.compileArrayMethodFirstLastIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode, true)
	case "last":
		return g.compileArrayMethodFirstLastIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode, false)
	case "is_empty":
		return g.compileArrayMethodIsEmptyIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "clear":
		return g.compileArrayMethodClearIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "capacity":
		return g.compileArrayMethodCapacityIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "reserve":
		return g.compileArrayMethodReserveIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "clone_shallow":
		return g.compileArrayMethodCloneShallowIntrinsic(ctx, objExpr, objType, args, effectiveExpected, callNode)
	case "get":
		if len(args) != 1 {
			return nil, "", "", false
		}
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, args[0], "")
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		lengthTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		resultType := effectiveExpected
		if resultType == "" {
			resultType = "runtime.Value"
		}
		lines := append(idxLines, fmt.Sprintf("%s := %s", objTemp, objExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexTemp), resultType)
		if !ok {
			return nil, "", "", false
		}
		if resultType == "runtime.Value" {
			lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp))
		} else {
			lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, resultType))
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)),
			fmt.Sprintf("if %s >= 0 && %s < %s {", indexTemp, indexTemp, lengthTemp),
		)
		lines = append(lines, indentLines(elemLines, 1)...)
		lines = append(lines,
			fmt.Sprintf("\t%s = %s", resultTemp, elemExpr),
			"}",
		)
		return lines, resultTemp, resultType, true
	case "set":
		if len(args) != 2 {
			return nil, "", "", false
		}
		setIdxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, args[0], "")
		if !ok {
			return nil, "", "", false
		}
		setValLines, valueExpr, ok := g.compileStaticArrayValueArg(ctx, objType, args[1])
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		valueTemp := ctx.newTemp()
		lengthTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := append(setIdxLines, setValLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s", valueTemp, valueExpr),
			fmt.Sprintf("%s := %s", lengthTemp, g.staticArrayLengthExpr(objTemp)),
			fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
			fmt.Sprintf("if %s < 0 || %s >= %s {", indexTemp, indexTemp, lengthTemp),
			fmt.Sprintf("\t%s = __able_index_error(%s, %s)", resultTemp, indexTemp, lengthTemp),
			"} else {",
			fmt.Sprintf("\t%s.Elements[%s] = %s", objTemp, indexTemp, valueTemp),
			"}",
			g.staticArraySyncCall(objType, objTemp),
		)
		if effectiveExpected == "" || effectiveExpected == "runtime.Value" {
			return lines, resultTemp, "runtime.Value", true
		}
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, effectiveExpected)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, effectiveExpected, true
	case "read_slot":
		if len(args) != 1 {
			return nil, "", "", false
		}
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, args[0], "")
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		resultType := effectiveExpected
		if resultType == "" {
			resultType = "runtime.Value"
		}
		lines := append(idxLines, fmt.Sprintf("%s := %s", objTemp, objExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		elemLines, elemExpr, _, ok := g.staticArrayResultExprLines(ctx, objType, fmt.Sprintf("%s.Elements[%s]", objTemp, indexTemp), resultType)
		if !ok {
			return nil, "", "", false
		}
		if resultType == "runtime.Value" {
			lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp))
		} else {
			lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, resultType))
		}
		lines = append(lines, fmt.Sprintf("if %s >= 0 && %s < %s {", indexTemp, indexTemp, g.staticArrayLengthExpr(objTemp)))
		lines = append(lines, indentLines(elemLines, 1)...)
		lines = append(lines,
			fmt.Sprintf("\t%s = %s", resultTemp, elemExpr),
			"}",
		)
		return lines, resultTemp, resultType, true
	case "write_slot":
		if len(args) != 2 {
			return nil, "", "", false
		}
		setIdxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, args[0], "")
		if !ok {
			return nil, "", "", false
		}
		setValLines, valueExpr, ok := g.compileStaticArrayValueArg(ctx, objType, args[1])
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		valueTemp := ctx.newTemp()
		lines := append(setIdxLines, setValLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s", valueTemp, valueExpr),
			fmt.Sprintf("if %s >= 0 && %s < %s {", indexTemp, indexTemp, g.staticArrayLengthExpr(objTemp)),
			fmt.Sprintf("\t%s.Elements[%s] = %s", objTemp, indexTemp, valueTemp),
			fmt.Sprintf("} else if %s >= 0 {", indexTemp),
			fmt.Sprintf("\tfor %s <= %s { %s.Elements = append(%s.Elements, %s) }", g.staticArrayLengthExpr(objTemp), indexTemp, objTemp, objTemp, g.staticArrayZeroValueExpr(objType)),
			fmt.Sprintf("\t%s.Elements[%s] = %s", objTemp, indexTemp, valueTemp),
			"}",
			g.staticArraySyncCall(objType, objTemp),
		)
		if effectiveExpected == "" || effectiveExpected == "runtime.Value" {
			return lines, "runtime.VoidValue{}", "runtime.Value", true
		}
		if effectiveExpected == "struct{}" {
			return lines, "struct{}{}", "struct{}", true
		}
		return nil, "", "", false
	}
	return nil, "", "", false
}
