package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) recoverStaticArrayPatternSubject(ctx *compileContext, subjectTemp string, subjectType string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || subjectTemp == "" {
		return nil, "", "", false
	}
	if subjectType != "runtime.Value" && subjectType != "any" {
		return nil, "", "", false
	}
	for _, typeExpr := range []ast.TypeExpression{ctx.matchSubjectTypeExpr, ctx.expectedTypeExpr} {
		if typeExpr == nil {
			continue
		}
		typeExpr = g.lowerNormalizedTypeExpr(ctx, typeExpr)
		recoveredType, ok := g.lowerCarrierType(ctx, typeExpr)
		if !ok || recoveredType == "" || recoveredType == "runtime.Value" || recoveredType == "any" || !g.isStaticArrayType(recoveredType) {
			continue
		}
		runtimeSubject := subjectTemp
		lines := []string{}
		if subjectType == "any" {
			runtimeTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", runtimeTemp, subjectTemp))
			runtimeSubject = runtimeTemp
		}
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, runtimeSubject, recoveredType)
		if !ok {
			continue
		}
		lines = append(lines, convLines...)
		convertedTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := %s", convertedTemp, converted),
			fmt.Sprintf("_ = %s", convertedTemp),
		)
		return lines, convertedTemp, recoveredType, true
	}
	return nil, "", "", false
}
