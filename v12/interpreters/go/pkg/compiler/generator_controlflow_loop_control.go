package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileBreakStatement(ctx *compileContext, stmt *ast.BreakStatement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing break")
		return nil, false
	}
	label := ""
	if stmt.Label != nil {
		label = stmt.Label.Name
		if label == "" {
			ctx.setReason("missing break label")
			return nil, false
		}
		if !ctx.hasBreakpoint(label) {
			ctx.setReason("unknown break label")
			return nil, false
		}
	} else if ctx.loopDepth <= 0 {
		ctx.setReason("break used outside loop")
		return nil, false
	}
	// Labeled break (breakpoint expression) — use Go's native break with labeled switch
	if label != "" {
		goLabel := ctx.breakpointGoLabels[label]
		resultTemp := ctx.breakpointResultTemps[label]
		if goLabel == "" || resultTemp == "" {
			ctx.setReason("break label not mapped to Go label")
			return nil, false
		}
		resultType := "runtime.Value"
		if ctx.breakpointResultTypes != nil && ctx.breakpointResultTypes[label] != "" {
			resultType = ctx.breakpointResultTypes[label]
		}
		if ctx.breakpointResultProbes != nil {
			if probe := ctx.breakpointResultProbes[label]; probe != nil && stmt.Value == nil {
				probe.sawNil = true
			}
		}
		valueExpr := ""
		if stmt.Value != nil {
			valLines, expr, goType, ok := g.compileExprLines(ctx, stmt.Value, "")
			if !ok {
				return nil, false
			}
			if ctx.breakpointResultProbes != nil {
				if probe := ctx.breakpointResultProbes[label]; probe != nil {
					probe.branchTypes = append(probe.branchTypes, goType)
					inferred, _ := g.inferExpressionTypeExpr(ctx, stmt.Value, goType)
					probe.branchTypeExprs = append(probe.branchTypeExprs, inferred)
				}
			}
			convLines, coercedExpr, ok := g.controlFlowResultExpr(ctx, resultType, expr, goType)
			if !ok {
				ctx.setReason(fmt.Sprintf("break value unsupported (%s -> %s, label=%s)", goType, resultType, label))
				return nil, false
			}
			valueExpr = coercedExpr
			if len(valLines) > 0 || len(convLines) > 0 {
				result := append([]string{}, valLines...)
				result = append(result, convLines...)
				result = append(result,
					fmt.Sprintf("%s = %s", resultTemp, valueExpr),
					fmt.Sprintf("break %s", goLabel),
				)
				return result, true
			}
		} else {
			nilExpr, ok := g.controlFlowNilResultExpr(resultType)
			if !ok {
				ctx.setReason("break value unsupported")
				return nil, false
			}
			valueExpr = nilExpr
		}
		return []string{
			fmt.Sprintf("%s = %s", resultTemp, valueExpr),
			fmt.Sprintf("break %s", goLabel),
		}, true
	}
	// Loop break — use Go's native break with label
	var lines []string
	resultType := ctx.loopBreakValueType
	if ctx.loopBreakValueTemp == "" && resultType == "" {
		if stmt.Value != nil {
			valueLines, valueExpr, valueType, ok := g.compileExprLines(ctx, stmt.Value, "")
			if !ok {
				return nil, false
			}
			lines = append(lines, valueLines...)
			lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
			if !ok {
				return nil, false
			}
		}
		if ctx.loopLabel != "" {
			lines = append(lines, fmt.Sprintf("break %s", ctx.loopLabel))
		} else {
			lines = append(lines, "break")
		}
		return lines, true
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	if stmt.Value != nil {
		valLines, expr, goType, ok := g.compileExprLines(ctx, stmt.Value, "")
		if !ok {
			return nil, false
		}
		lines = append(lines, valLines...)
		if ctx.loopBreakProbe != nil {
			ctx.loopBreakProbe.branchTypes = append(ctx.loopBreakProbe.branchTypes, goType)
			inferred, _ := g.inferExpressionTypeExpr(ctx, stmt.Value, goType)
			ctx.loopBreakProbe.branchTypeExprs = append(ctx.loopBreakProbe.branchTypeExprs, inferred)
		}
		convLines, coercedExpr, ok := g.controlFlowResultExpr(ctx, resultType, expr, goType)
		if !ok {
			ctx.setReason(fmt.Sprintf("break value unsupported (%s -> %s)", goType, resultType))
			return nil, false
		}
		lines = append(lines, convLines...)
		if ctx.loopBreakValueTemp != "" {
			lines = append(lines, fmt.Sprintf("%s = %s", ctx.loopBreakValueTemp, coercedExpr))
		}
	} else if ctx.loopBreakValueTemp != "" {
		nilExpr, ok := g.controlFlowNilResultExpr(resultType)
		if !ok {
			ctx.setReason("break value unsupported")
			return nil, false
		}
		if ctx.loopBreakProbe != nil {
			ctx.loopBreakProbe.sawNil = true
		}
		lines = append(lines, fmt.Sprintf("%s = %s", ctx.loopBreakValueTemp, nilExpr))
	}
	if ctx.loopLabel != "" {
		lines = append(lines, fmt.Sprintf("break %s", ctx.loopLabel))
	} else {
		lines = append(lines, "break")
	}
	return lines, true
}

func (g *generator) compileContinueStatement(ctx *compileContext, stmt *ast.ContinueStatement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing continue")
		return nil, false
	}
	if ctx.loopDepth <= 0 {
		ctx.setReason("continue used outside loop")
		return nil, false
	}
	if stmt.Label != nil {
		ctx.setReason("labeled continue unsupported")
		return nil, false
	}
	if ctx.loopLabel != "" {
		return []string{fmt.Sprintf("continue %s", ctx.loopLabel)}, true
	}
	return []string{"continue"}, true
}

func (g *generator) compileLoopExpression(ctx *compileContext, loop *ast.LoopExpression, expected string) ([]string, string, string, bool) {
	if loop == nil || loop.Body == nil {
		ctx.setReason("missing loop expression")
		return nil, "", "", false
	}
	if lines, expr, goType, ok := g.compileCountedLoopExpression(ctx, loop, expected); ok {
		return lines, expr, goType, true
	}
	resultType := g.inferLoopExpressionResultType(ctx, loop, expected)
	if resultType == "" {
		resultType = "runtime.Value"
	}
	loopLabelName := ctx.newTemp()
	valueTemp := ctx.newTemp()
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	bodyCtx.loopBreakValueTemp = valueTemp
	bodyCtx.loopBreakValueType = resultType
	bodyCtx.loopBreakProbe = nil
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, "", "", false
	}
	zeroExpr, ok := g.zeroValueExpr(resultType)
	if !ok {
		ctx.setReason("loop expression type mismatch")
		return nil, "", "", false
	}
	lines := []string{
		fmt.Sprintf("var %s %s = %s", valueTemp, resultType, zeroExpr),
	}
	forLine := "for {"
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for {", loopLabelName)
	}
	lines = append(lines, forLine)
	lines = append(lines, indentLines(bodyLines, 1)...)
	lines = append(lines, "}")
	return lines, valueTemp, resultType, true
}

// compileLoopStatement lowers a loop whose expression result is discarded.
// Break values still run for their effects and error propagation, but no result
// carrier is needed after the loop.
func (g *generator) compileLoopStatement(ctx *compileContext, loop *ast.LoopExpression) ([]string, bool) {
	if loop == nil || loop.Body == nil {
		return nil, false
	}
	loopLabelName := ctx.newTemp()
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	bodyCtx.loopBreakValueTemp = ""
	bodyCtx.loopBreakValueType = ""
	bodyCtx.loopBreakProbe = nil
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, false
	}
	forLine := "for {"
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for {", loopLabelName)
	}
	lines := []string{forLine}
	lines = append(lines, indentLines(bodyLines, 1)...)
	lines = append(lines, "}")
	return lines, true
}

func (g *generator) compileBreakpointExpression(ctx *compileContext, expr *ast.BreakpointExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Body == nil {
		ctx.setReason("missing breakpoint expression")
		return nil, "", "", false
	}
	if expr.Label == nil || expr.Label.Name == "" {
		ctx.setReason("breakpoint requires label")
		return nil, "", "", false
	}
	label := expr.Label.Name
	resultType := g.inferBreakpointExpressionResultType(ctx, expr, expected)
	if resultType == "" {
		resultType = "runtime.Value"
	}

	goLabel := ctx.newTemp()
	resultTemp := ctx.newTemp()

	bodyCtx := ctx.child()
	bodyCtx.pushBreakpoint(label)
	if bodyCtx.breakpointGoLabels == nil {
		bodyCtx.breakpointGoLabels = make(map[string]string)
	}
	if bodyCtx.breakpointResultTemps == nil {
		bodyCtx.breakpointResultTemps = make(map[string]string)
	}
	if bodyCtx.breakpointResultTypes == nil {
		bodyCtx.breakpointResultTypes = make(map[string]string)
	}
	bodyCtx.breakpointGoLabels[label] = goLabel
	bodyCtx.breakpointResultTemps[label] = resultTemp
	bodyCtx.breakpointResultTypes[label] = resultType

	// Compile the body block as statements + tail expression.
	stmts := expr.Body.Body
	var bodyLines []string
	for idx, stmt := range stmts {
		isLast := idx == len(stmts)-1
		if isLast {
			// Try to compile last statement as a value expression
			if tailExpr, ok := stmt.(ast.Expression); ok {
				tailLines, tailValue, tailType, ok := g.compileTailExpression(bodyCtx, resultType, tailExpr)
				if ok {
					bodyLines = append(bodyLines, tailLines...)
					coerceLines, coercedExpr, ok := g.controlFlowResultExpr(ctx, resultType, tailValue, tailType)
					if !ok {
						ctx.setReason("breakpoint type mismatch")
						return nil, "", "", false
					}
					bodyLines = append(bodyLines, coerceLines...)
					bodyLines = append(bodyLines, fmt.Sprintf("%s = %s", resultTemp, coercedExpr))
					break
				}
			}
		}
		stmtLines, ok := g.compileStatement(bodyCtx, stmt)
		if !ok {
			return nil, "", "", false
		}
		bodyLines = append(bodyLines, stmtLines...)
	}
	bodyCtx.popBreakpoint(label)

	// Build labeled switch
	zeroExpr, ok := g.zeroValueExpr(resultType)
	if !ok {
		ctx.setReason("breakpoint type mismatch")
		return nil, "", "", false
	}
	lines := []string{fmt.Sprintf("var %s %s = %s", resultTemp, resultType, zeroExpr)}
	lines = append(lines, fmt.Sprintf("%s: switch { default: %s }", goLabel, strings.Join(bodyLines, "; ")))
	return lines, resultTemp, resultType, true
}
