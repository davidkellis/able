package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticArrayFactoryCall(
	ctx *compileContext,
	typeName string,
	methodName string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if g == nil || ctx == nil || typeName != "Array" {
		return nil, "", "", false
	}
	arrayType := expected
	if arrayType == "" {
		arrayType = "*Array"
	}
	if !g.isStaticArrayType(arrayType) {
		return nil, "", "", false
	}
	lines := []string{fmt.Sprintf("__able_push_call_frame(%s)", callNode)}
	switch methodName {
	case "new":
		if len(args) != 0 {
			return nil, "", "", false
		}
		arrayTemp := ctx.newTemp()
		if spec, ok := g.monoArraySpecForGoType(arrayType); ok && spec != nil {
			lines = append(lines, fmt.Sprintf("%s := &%s{}", arrayTemp, spec.GoName))
		} else {
			lines = append(lines, fmt.Sprintf("%s := &Array{}", arrayTemp))
		}
		lines = append(lines, g.staticArraySyncCall(arrayType, arrayTemp), "__able_pop_call_frame()")
		if expected == "" || g.typeMatches(expected, arrayType) {
			return lines, arrayTemp, arrayType, true
		}
		return g.coerceStaticArrayFactoryResult(ctx, lines, arrayTemp, arrayType, expected)
	case "with_capacity":
		if len(args) != 1 {
			return nil, "", "", false
		}
		capLines, capExpr, capType, ok := g.compileExprLines(ctx, args[0], "int32")
		if !ok {
			return nil, "", "", false
		}
		capTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		capacityTemp := ctx.newTemp()
		arrayTemp := ctx.newTemp()
		lines = append(lines, capLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", capTemp, capExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, capTemp, capType, idxTemp, capacityTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("if %s < 0 {", capacityTemp))
		transferLines, ok := g.controlTransferLines(ctx, g.runtimeErrorControlExpr(callNode, "fmt.Errorf(\"capacity must be a non-negative integer\")"))
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, indentLines(transferLines, 1)...)
		lines = append(lines, "}")
		if spec, ok := g.monoArraySpecForGoType(arrayType); ok && spec != nil {
			lines = append(lines, fmt.Sprintf("%s := &%s{Elements: make([]%s, 0, %s)}", arrayTemp, spec.GoName, spec.ElemGoType, capacityTemp))
		} else {
			lines = append(lines, fmt.Sprintf("%s := &Array{Elements: make([]runtime.Value, 0, %s)}", arrayTemp, capacityTemp))
		}
		lines = append(lines, g.staticArraySyncCall(arrayType, arrayTemp), "__able_pop_call_frame()")
		if expected == "" || g.typeMatches(expected, arrayType) {
			return lines, arrayTemp, arrayType, true
		}
		return g.coerceStaticArrayFactoryResult(ctx, lines, arrayTemp, arrayType, expected)
	default:
		return nil, "", "", false
	}
}

func (g *generator) coerceStaticArrayFactoryResult(
	ctx *compileContext,
	lines []string,
	arrayExpr string,
	arrayType string,
	expected string,
) ([]string, string, string, bool) {
	valueLines, valueExpr, ok := g.runtimeValueLines(ctx, arrayExpr, arrayType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, valueLines...)
	if expected == "runtime.Value" {
		return lines, valueExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueExpr, expected)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}
