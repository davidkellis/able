package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) recoverTypedPatternCarrier(ctx *compileContext, expr ast.TypeExpression) (string, bool) {
	if g == nil || expr == nil {
		return "", false
	}
	mapped, ok := g.lowerCarrierType(ctx, g.lowerNormalizedTypeExpr(ctx, expr))
	if !ok || mapped == "" || mapped == "struct{}" || mapped == "runtime.Value" || mapped == "any" {
		return "", false
	}
	return mapped, true
}

func (g *generator) compileDynamicTypedPatternCast(ctx *compileContext, subjectTemp string, subjectType string, expr ast.TypeExpression) ([]string, string, string, string, bool) {
	if g == nil || ctx == nil || subjectTemp == "" || expr == nil {
		return nil, "", "", "", false
	}
	typeExpr, ok := g.renderTypeExpression(g.lowerNormalizedTypeExpr(ctx, expr))
	if !ok {
		return nil, "", "", "", false
	}
	g.needsAst = true
	lines := []string{}
	castSubject := subjectTemp
	if subjectType == "any" {
		convTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, subjectTemp))
		castSubject = convTemp
	}

	runtimeTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s, %s := __able_try_cast(%s, %s)", runtimeTemp, okTemp, controlTemp, castSubject, typeExpr))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", "", false
	}
	lines = append(lines, controlLines...)

	narrowedType := "runtime.Value"
	if mapped, ok := g.recoverTypedPatternCarrier(ctx, expr); ok {
		narrowedType = mapped
	}
	narrowedTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s %s", narrowedTemp, narrowedType))
	if narrowedType == "runtime.Value" {
		lines = append(lines, fmt.Sprintf("if %s { %s = %s }", okTemp, narrowedTemp, runtimeTemp))
		return lines, narrowedTemp, narrowedType, okTemp, true
	}

	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, runtimeTemp, narrowedType)
	if !ok {
		return nil, "", "", "", false
	}
	lines = append(lines, fmt.Sprintf("if %s {", okTemp))
	lines = append(lines, indentLines(convLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", narrowedTemp, converted))
	lines = append(lines, "}")
	return lines, narrowedTemp, narrowedType, okTemp, true
}
