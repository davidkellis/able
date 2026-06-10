package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileDiscardedTailExpression(ctx *compileContext, expr ast.Expression) ([]string, string, string, bool) {
	if ctx == nil {
		return nil, "", "", false
	}
	switch discarded := expr.(type) {
	case *ast.IfExpression:
		lines, compiled := g.compileIfStatement(ctx, discarded)
		return lines, "", "struct{}", compiled
	case *ast.MatchExpression:
		lines, compiled := g.compileMatchStatement(ctx, discarded)
		return lines, "", "struct{}", compiled
	case *ast.BlockExpression:
		lines, compiled := g.compileBlockStatement(ctx, discarded)
		return lines, "", "struct{}", compiled
	}
	previous := ctx.discardResult
	ctx.discardResult = true
	defer func() {
		ctx.discardResult = previous
	}()
	return g.compileTailExpression(ctx, "", expr)
}

func (g *generator) discardStatementResult(ctx *compileContext, lines []string, expr string, exprType string) ([]string, bool) {
	if g == nil || ctx == nil || expr == "" {
		return lines, true
	}
	valueTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, expr))
	if exprType == "runtime.ErrorValue" {
		return g.lowerDiscardedErrorCarrier(ctx, lines, valueTemp)
	}
	if g.isNativeErrorCarrierType(exprType) {
		valueLines, runtimeExpr, ok := g.lowerRuntimeValue(ctx, valueTemp, exprType)
		if !ok {
			ctx.setReason("statement result type mismatch")
			return nil, false
		}
		lines = append(lines, valueLines...)
		return g.lowerDiscardedErrorCarrier(ctx, lines, runtimeExpr)
	}
	if union := g.nativeUnionInfoForGoType(exprType); union != nil {
		if member, ok := g.nativeUnionMember(union, "runtime.ErrorValue"); ok && member != nil {
			errTemp := ctx.newTemp()
			okTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("if %s, %s := %s(%s); %s {", errTemp, okTemp, member.UnwrapHelper, valueTemp, okTemp))
			transferLines, ok := g.lowerControlTransfer(ctx, fmt.Sprintf("__able_raise_control(nil, %s)", errTemp))
			if !ok {
				return nil, false
			}
			lines = append(lines, indentLines(transferLines, 1)...)
			lines = append(lines, "}")
		}
	}
	if exprType == "runtime.Value" || exprType == "any" {
		probeExpr := valueTemp
		if exprType == "any" {
			runtimeTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", runtimeTemp, valueTemp))
			probeExpr = runtimeTemp
		}
		errTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("if %s, %s, _ := __able_runtime_error_value(%s); %s {", errTemp, okTemp, probeExpr, okTemp))
		transferLines, ok := g.lowerControlTransfer(ctx, fmt.Sprintf("__able_raise_control(nil, %s)", errTemp))
		if !ok {
			return nil, false
		}
		lines = append(lines, indentLines(transferLines, 1)...)
		lines = append(lines, "}")
	}
	lines = append(lines, fmt.Sprintf("_ = %s", valueTemp))
	return lines, true
}

func (g *generator) lowerDiscardedErrorCarrier(ctx *compileContext, lines []string, expr string) ([]string, bool) {
	if g == nil || ctx == nil || expr == "" {
		return nil, false
	}
	transferLines, ok := g.lowerControlTransfer(ctx, fmt.Sprintf("__able_raise_control(nil, %s)", expr))
	if !ok {
		return nil, false
	}
	lines = append(lines, transferLines...)
	return lines, true
}
