package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileRaiseStatement(ctx *compileContext, stmt *ast.RaiseStatement) ([]string, bool) {
	if stmt == nil || stmt.Expression == nil {
		ctx.setReason("missing raise expression")
		return nil, false
	}
	expr, goType, ok := g.compileExpr(ctx, stmt.Expression, "")
	if !ok {
		return nil, false
	}
	valueRuntime, ok := g.runtimeValueExpr(expr, goType)
	if !ok {
		ctx.setReason("raise value unsupported")
		return nil, false
	}
	raiseNode := g.diagNodeName(stmt, "*ast.RaiseStatement", "raise")
	return []string{fmt.Sprintf("bridge.RaiseWithContext(__able_runtime, %s, %s)", raiseNode, valueRuntime)}, true
}

func (g *generator) compileRescueExpression(ctx *compileContext, expr *ast.RescueExpression, expected string) (string, string, bool) {
	if expr == nil || expr.MonitoredExpression == nil {
		ctx.setReason("missing rescue expression")
		return "", "", false
	}
	monitoredLines, monitoredExpr, monitoredType, ok := g.compileTailExpression(ctx, expected, expr.MonitoredExpression)
	if !ok {
		return "", "", false
	}
	if len(expr.Clauses) == 0 {
		ctx.setReason("rescue requires clauses")
		return "", "", false
	}
	type rescueClause struct {
		cond      string
		bindLines []string
		guardExpr string
		bodyLines []string
		bodyExpr  string
		bodyType  string
	}
	clauses := make([]rescueClause, 0, len(expr.Clauses))
	subjectTemp := ctx.newTemp()
	recoveredErrTemp := ctx.newTemp()
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		clauseCtx.rethrowVar = subjectTemp
		clauseCtx.rethrowErrVar = recoveredErrTemp
		cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, subjectTemp, "runtime.Value")
		if !ok {
			return "", "", false
		}
		guardExpr := ""
		if clause.Guard != nil {
			guardValue, ok := g.compileCondition(clauseCtx, clause.Guard)
			if !ok {
				return "", "", false
			}
			guardExpr = guardValue
		}
		bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(clauseCtx, expected, clause.Body)
		if !ok {
			return "", "", false
		}
		clauses = append(clauses, rescueClause{
			cond:      cond,
			bindLines: bindLines,
			guardExpr: guardExpr,
			bodyLines: bodyLines,
			bodyExpr:  bodyExpr,
			bodyType:  bodyType,
		})
	}
	resultType := expected
	if resultType == "" {
		resultType = monitoredType
		if resultType == "" && len(clauses) > 0 {
			resultType = clauses[0].bodyType
		}
		for _, clause := range clauses {
			if resultType == "" {
				resultType = clause.bodyType
				continue
			}
			if !g.typeMatches(resultType, clause.bodyType) {
				resultType = "runtime.Value"
				break
			}
		}
		if resultType == "" {
			resultType = "runtime.Value"
		} else if !g.typeMatches(resultType, monitoredType) {
			resultType = "runtime.Value"
		}
	}
	monitoredExpr, ok = g.coerceRescueBranch(ctx, resultType, monitoredExpr, monitoredType)
	if !ok {
		return "", "", false
	}
	for i := range clauses {
		coerced, ok := g.coerceRescueBranch(ctx, resultType, clauses[i].bodyExpr, clauses[i].bodyType)
		if !ok {
			return "", "", false
		}
		clauses[i].bodyExpr = coerced
		clauses[i].bodyType = resultType
	}
	resultTemp := ctx.newTemp()
	recoveredTemp := ctx.newTemp()
	recoveredOkTemp := ctx.newTemp()
	matchedTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s runtime.Value", recoveredTemp),
		fmt.Sprintf("var %s bool", recoveredOkTemp),
		fmt.Sprintf("var %s error", recoveredErrTemp),
		fmt.Sprintf("_ = %s", recoveredErrTemp),
		"func() {",
		fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { switch v := recovered.(type) { case runtime.Value: %s = v; %s = true; case error: if val, ok := interpreter.RaisedValue(v); ok { %s = val; %s = true; %s = v } else { panic(recovered) }; default: panic(recovered) } } }()", recoveredTemp, recoveredOkTemp, recoveredTemp, recoveredOkTemp, recoveredErrTemp),
	}
	lines = append(lines, indentLines(monitoredLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, monitoredExpr))
	lines = append(lines, "}()")
	lines = append(lines, fmt.Sprintf("if !%s { return %s }", recoveredOkTemp, resultTemp))
	lines = append(lines, fmt.Sprintf("%s := %s", subjectTemp, recoveredTemp))
	lines = append(lines, fmt.Sprintf("_ = %s", subjectTemp))
	lines = append(lines, fmt.Sprintf("%s := false", matchedTemp))
	for _, clause := range clauses {
		branchLines := []string{}
		branchLines = append(branchLines, clause.bindLines...)
		if clause.guardExpr != "" {
			branchLines = append(branchLines, fmt.Sprintf("if %s {", clause.guardExpr))
			branchLines = append(branchLines, indentLines(clause.bodyLines, 1)...)
			branchLines = append(branchLines, fmt.Sprintf("\t%s = %s", resultTemp, clause.bodyExpr))
			branchLines = append(branchLines, fmt.Sprintf("\t%s = true", matchedTemp))
			branchLines = append(branchLines, "}")
		} else {
			branchLines = append(branchLines, clause.bodyLines...)
			branchLines = append(branchLines, fmt.Sprintf("%s = %s", resultTemp, clause.bodyExpr))
			branchLines = append(branchLines, fmt.Sprintf("%s = true", matchedTemp))
		}
		lines = append(lines, fmt.Sprintf("if !%s && %s {", matchedTemp, clause.cond))
		lines = append(lines, indentLines(branchLines, 1)...)
		lines = append(lines, "}")
	}
	lines = append(lines, fmt.Sprintf("if !%s { panic(%s) }", matchedTemp, recoveredTemp))
	lines = append(lines, fmt.Sprintf("return %s", resultTemp))
	exprValue := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return exprValue, resultType, true
}

func (g *generator) coerceRescueBranch(ctx *compileContext, resultType string, expr string, exprType string) (string, bool) {
	if resultType == "" || exprType == "" || expr == "" {
		ctx.setReason("rescue clause type mismatch")
		return "", false
	}
	if resultType == exprType {
		return expr, true
	}
	if resultType == "runtime.Value" && exprType != "runtime.Value" {
		converted, ok := g.runtimeValueExpr(expr, exprType)
		if !ok {
			ctx.setReason("rescue clause type mismatch")
			return "", false
		}
		return converted, true
	}
	if resultType != "runtime.Value" && exprType == "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(expr, resultType)
		if !ok {
			ctx.setReason("rescue clause type mismatch")
			return "", false
		}
		return converted, true
	}
	ctx.setReason("rescue clause type mismatch")
	return "", false
}

func (g *generator) compileRethrowStatement(ctx *compileContext, stmt *ast.RethrowStatement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing rethrow")
		return nil, false
	}
	if ctx != nil && ctx.rethrowVar != "" {
		if ctx.rethrowErrVar != "" {
			return []string{
				fmt.Sprintf("if %s != nil { panic(%s) }", ctx.rethrowErrVar, ctx.rethrowErrVar),
				fmt.Sprintf("bridge.Raise(%s)", ctx.rethrowVar),
			}, true
		}
		return []string{fmt.Sprintf("bridge.Raise(%s)", ctx.rethrowVar)}, true
	}
	return []string{`bridge.Raise(runtime.ErrorValue{Message: "Unknown rethrow"})`}, true
}
