package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type fixedIntegerCarrier struct {
	bits   int
	signed bool
	wide   bool
}

func fixedIntegerCarrierForGoType(goType string) (fixedIntegerCarrier, bool) {
	switch goType {
	case "int8":
		return fixedIntegerCarrier{bits: 8, signed: true}, true
	case "int16":
		return fixedIntegerCarrier{bits: 16, signed: true}, true
	case "int32":
		return fixedIntegerCarrier{bits: 32, signed: true}, true
	case "int64":
		return fixedIntegerCarrier{bits: 64, signed: true}, true
	case "runtime.Int128":
		return fixedIntegerCarrier{bits: 128, signed: true, wide: true}, true
	case "uint8":
		return fixedIntegerCarrier{bits: 8}, true
	case "uint16":
		return fixedIntegerCarrier{bits: 16}, true
	case "uint32":
		return fixedIntegerCarrier{bits: 32}, true
	case "uint64":
		return fixedIntegerCarrier{bits: 64}, true
	case "runtime.Uint128":
		return fixedIntegerCarrier{bits: 128, wide: true}, true
	default:
		return fixedIntegerCarrier{}, false
	}
}

func splitFixedIntegerMethodName(name string) (mode string, operation string, ok bool) {
	for _, candidate := range []string{"wrapping", "saturating", "checked"} {
		if !strings.HasPrefix(name, candidate+"_") {
			continue
		}
		operation = strings.TrimPrefix(name, candidate+"_")
		switch operation {
		case "add", "sub", "mul":
			return candidate, operation, true
		default:
			return "", "", false
		}
	}
	return "", "", false
}

func (g *generator) compileFixedIntegerArithmeticMethodCall(
	ctx *compileContext,
	objExpr string,
	objType string,
	methodName string,
	args []ast.Expression,
	expected string,
) ([]string, string, string, bool) {
	carrier, carrierOK := fixedIntegerCarrierForGoType(objType)
	mode, operation, methodOK := splitFixedIntegerMethodName(methodName)
	if !carrierOK || !methodOK || len(args) != 1 {
		return nil, "", "", false
	}

	leftTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("%s := %s", leftTemp, objExpr)}
	argLines, argExpr, argType, ok := g.compileExprLines(ctx, args[0], objType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, argLines...)
	if argType != objType {
		if argType != "runtime.Value" && argType != "any" {
			return nil, "", "", false
		}
		convLines, converted, convertedOK := g.lowerExpectRuntimeValue(ctx, argExpr, objType)
		if !convertedOK {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		argExpr = converted
	}
	rightTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", rightTemp, argExpr))

	if carrier.wide {
		return g.compileWideIntegerArithmeticMethodResult(
			ctx,
			lines,
			leftTemp,
			rightTemp,
			objType,
			mode,
			operation,
			expected,
		)
	}
	return g.compileNativeIntegerArithmeticMethodResult(
		ctx,
		lines,
		leftTemp,
		rightTemp,
		objType,
		carrier,
		mode,
		operation,
		expected,
	)
}

func (g *generator) compileWideIntegerArithmeticMethodResult(
	ctx *compileContext,
	lines []string,
	left string,
	right string,
	goType string,
	mode string,
	operation string,
	expected string,
) ([]string, string, string, bool) {
	method := strings.ToUpper(operation[:1]) + operation[1:]
	switch mode {
	case "wrapping":
		method = "Wrapping" + method
	case "saturating":
		method = "Saturating" + method
	case "checked":
		method += "Checked"
	}

	resultTemp := ctx.newTemp()
	resultType := goType
	if mode != "checked" {
		lines = append(lines, fmt.Sprintf("%s := %s.%s(%s)", resultTemp, left, method, right))
		return g.finishFixedIntegerArithmeticResult(ctx, lines, resultTemp, resultType, expected)
	}

	presentTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := %s.%s(%s)", resultTemp, presentTemp, left, method, right))
	nullableType, ok := g.nativeNullableCarrierType(goType)
	if !ok {
		return nil, "", "", false
	}
	nullableTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := %s{}", nullableTemp, nullableType),
		fmt.Sprintf("if %s { %s = __able_some(%s) }", presentTemp, nullableTemp, resultTemp),
	)
	return g.finishFixedIntegerArithmeticResult(ctx, lines, nullableTemp, nullableType, expected)
}

func (g *generator) compileNativeIntegerArithmeticMethodResult(
	ctx *compileContext,
	lines []string,
	left string,
	right string,
	goType string,
	carrier fixedIntegerCarrier,
	mode string,
	operation string,
	expected string,
) ([]string, string, string, bool) {
	opSymbol := map[string]string{"add": "+", "sub": "-", "mul": "*"}[operation]
	resultTemp := ctx.newTemp()
	if mode == "wrapping" {
		lines = append(lines, fmt.Sprintf("%s := %s(%s %s %s)", resultTemp, goType, left, opSymbol, right))
		return g.finishFixedIntegerArithmeticResult(ctx, lines, resultTemp, goType, expected)
	}

	signedness := "unsigned"
	baseType := "uint64"
	if carrier.signed {
		signedness = "signed"
		baseType = "int64"
	}
	helperMode := mode
	if mode == "checked" {
		helperMode = "try"
	}
	helper := fmt.Sprintf("__able_%s_%s_%s", helperMode, operation, signedness)
	if mode == "saturating" {
		lines = append(lines,
			fmt.Sprintf("%s := %s(%s(%s), %s(%s), %d)", resultTemp, helper, baseType, left, baseType, right, carrier.bits),
		)
		castTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s(%s)", castTemp, goType, resultTemp))
		return g.finishFixedIntegerArithmeticResult(ctx, lines, castTemp, goType, expected)
	}

	presentTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s, %s := %s(%s(%s), %s(%s), %d)", resultTemp, presentTemp, helper, baseType, left, baseType, right, carrier.bits),
	)
	nullableType, ok := g.nativeNullableCarrierType(goType)
	if !ok {
		return nil, "", "", false
	}
	nullableTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := %s{}", nullableTemp, nullableType),
		fmt.Sprintf("if %s { %s = __able_some(%s(%s)) }", presentTemp, nullableTemp, goType, resultTemp),
	)
	return g.finishFixedIntegerArithmeticResult(ctx, lines, nullableTemp, nullableType, expected)
}

func (g *generator) finishFixedIntegerArithmeticResult(
	ctx *compileContext,
	lines []string,
	result string,
	resultType string,
	expected string,
) ([]string, string, string, bool) {
	if expected == "" || g.typeMatches(expected, resultType) {
		return lines, result, resultType, true
	}
	valueLines, valueExpr, ok := g.lowerRuntimeValue(ctx, result, resultType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, valueLines...)
	if expected == "runtime.Value" {
		return lines, valueExpr, expected, true
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, valueExpr, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}
