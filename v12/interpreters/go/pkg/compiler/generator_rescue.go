package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) inferRescueSubjectTypeExpr(ctx *compileContext, monitored ast.Expression) ast.TypeExpression {
	return g.inferHandledFailureTypeExpr(ctx, monitored)
}

func (g *generator) compileRaiseStatement(ctx *compileContext, stmt *ast.RaiseStatement) ([]string, bool) {
	if stmt == nil || stmt.Expression == nil {
		ctx.setReason("missing raise expression")
		return nil, false
	}
	exprLines, expr, goType, ok := g.compileExprLines(ctx, stmt.Expression, "")
	if !ok {
		return nil, false
	}
	var convLines []string
	valueRuntime := ""
	if errorLines, errorExpr, ok := g.nativeErrorValueLines(ctx, goType, expr); ok {
		convLines = errorLines
		valueRuntime = errorExpr
	} else {
		convLines, valueRuntime, ok = g.lowerRuntimeValue(ctx, expr, goType)
		if !ok {
			ctx.setReason("raise value unsupported")
			return nil, false
		}
	}
	raiseNode := g.diagNodeName(stmt, "*ast.RaiseStatement", "raise")
	lines := append([]string{}, exprLines...)
	lines = append(lines, convLines...)
	transferLines, ok := g.lowerControlTransfer(ctx, g.raiseControlExpr(raiseNode, valueRuntime))
	if !ok {
		return nil, false
	}
	lines = append(lines, transferLines...)
	return lines, true
}

func (g *generator) compileRescueExpression(ctx *compileContext, expr *ast.RescueExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.MonitoredExpression == nil {
		ctx.setReason("missing rescue expression")
		return nil, "", "", false
	}
	if len(expr.Clauses) == 0 {
		ctx.setReason("rescue requires clauses")
		return nil, "", "", false
	}

	type rescueClause struct {
		condLines    []string
		cond         string
		bindLines    []string
		guardLines   []string
		guardExpr    string
		bodyLines    []string
		bodyExpr     string
		bodyType     string
		bodyNode     ast.Expression
		bodyTypeExpr ast.TypeExpression
	}

	controlTemp := ctx.newTemp()
	monitoredDoneLabel := ctx.newTemp()
	subjectTemp := ctx.newTemp()
	monitoredCtx := ctx.child()
	monitoredCtx.controlCaptureVar = controlTemp
	monitoredCtx.controlCaptureLabel = monitoredDoneLabel
	monitoredLines, monitoredExpr, monitoredType, ok := g.compileTailExpression(monitoredCtx, expected, expr.MonitoredExpression)
	if !ok {
		return nil, "", "", false
	}
	monitoredTypeExpr, _ := g.inferExpressionTypeExpr(monitoredCtx, expr.MonitoredExpression, monitoredType)
	rescueSubjectTypeExpr := g.inferRescueSubjectTypeExpr(ctx, expr.MonitoredExpression)

	clauses := make([]rescueClause, 0, len(expr.Clauses))
	clauseTypes := make([]string, 0, len(expr.Clauses))
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		clauseCtx.rethrowVar = subjectTemp
		clauseCtx.rethrowControlVar = controlTemp
		if rescueSubjectTypeExpr != nil {
			clauseCtx.expectedTypeExpr = rescueSubjectTypeExpr
		}
		condLines, cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, subjectTemp, "runtime.Value")
		if !ok {
			return nil, "", "", false
		}
		guardExpr := ""
		var guardLines []string
		if clause.Guard != nil {
			guardLines, guardExpr, ok = g.compileCondition(clauseCtx, clause.Guard)
			if !ok {
				return nil, "", "", false
			}
		}
		bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(clauseCtx, expected, clause.Body)
		if !ok {
			return nil, "", "", false
		}
		bodyTypeExpr, _ := g.inferExpressionTypeExpr(clauseCtx, clause.Body, bodyType)
		clauses = append(clauses, rescueClause{
			condLines:    condLines,
			cond:         cond,
			bindLines:    bindLines,
			guardLines:   guardLines,
			guardExpr:    guardExpr,
			bodyLines:    bodyLines,
			bodyExpr:     bodyExpr,
			bodyType:     bodyType,
			bodyNode:     clause.Body,
			bodyTypeExpr: bodyTypeExpr,
		})
		clauseTypes = append(clauseTypes, bodyType)
	}

	resultType := expected
	if resultType == "" {
		resultType = monitoredType
		if resultType == "" && len(clauses) > 0 {
			resultType = clauses[0].bodyType
		}
		mismatch := false
		for _, clause := range clauses {
			if resultType == "" {
				resultType = clause.bodyType
				continue
			}
			if !g.typeMatches(resultType, clause.bodyType) {
				mismatch = true
			}
		}
		if resultType == "" {
			resultType = "runtime.Value"
		} else if mismatch || !g.typeMatches(resultType, monitoredType) {
			joinBranches := make([]joinBranchInfo, 0, len(clauses)+1)
			joinBranches = append(joinBranches, joinBranchInfo{
				GoType:   monitoredType,
				Expr:     expr.MonitoredExpression,
				TypeExpr: monitoredTypeExpr,
				SawNil:   g.joinBranchIsNilExpr(monitoredExpr, monitoredType),
			})
			for _, clause := range clauses {
				joinBranches = append(joinBranches, joinBranchInfo{
					GoType:   clause.bodyType,
					Expr:     clause.bodyNode,
					TypeExpr: clause.bodyTypeExpr,
					SawNil:   g.joinBranchIsNilExpr(clause.bodyExpr, clause.bodyType),
				})
			}
			if joinedType, ok := g.lowerJoinCarrierFromBranches(ctx, joinBranches); ok {
				resultType = joinedType
			} else {
				resultType = "runtime.Value"
			}
		}
	}

	monitoredCoerceLines, monitoredExpr, ok := g.coerceRescueBranch(ctx, resultType, monitoredExpr, monitoredType)
	if !ok {
		return nil, "", "", false
	}
	monitoredLines = append(monitoredLines, monitoredCoerceLines...)
	for i := range clauses {
		coerceLines, coerced, ok := g.coerceRescueBranch(ctx, resultType, clauses[i].bodyExpr, clauses[i].bodyType)
		if !ok {
			return nil, "", "", false
		}
		clauses[i].bodyLines = append(clauses[i].bodyLines, coerceLines...)
		clauses[i].bodyExpr = coerced
		clauses[i].bodyType = resultType
	}

	resultTemp := ctx.newTemp()
	matchedTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s *__ableControl", controlTemp),
		"{",
	}
	lines = append(lines, indentLines(monitoredLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, monitoredExpr))
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("if false { goto %s }", monitoredDoneLabel))
	lines = append(lines, fmt.Sprintf("%s:", monitoredDoneLabel))
	lines = append(lines, fmt.Sprintf("if %s != nil {", controlTemp))
	lines = append(lines, fmt.Sprintf("\t%s := __able_control_value(%s)", subjectTemp, controlTemp))
	lines = append(lines, fmt.Sprintf("\t_ = %s", subjectTemp))
	lines = append(lines, fmt.Sprintf("\t%s := false", matchedTemp))
	for _, clause := range clauses {
		branchLines := append([]string{}, clause.bindLines...)
		if clause.guardExpr != "" {
			branchLines = append(branchLines, clause.guardLines...)
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
		if len(clause.condLines) > 0 {
			lines = append(lines, fmt.Sprintf("\tif !%s {", matchedTemp))
			lines = append(lines, indentLines(clause.condLines, 2)...)
			lines = append(lines, fmt.Sprintf("\t\tif %s {", clause.cond))
			lines = append(lines, indentLines(branchLines, 3)...)
			lines = append(lines, "\t\t}")
			lines = append(lines, "\t}")
		} else {
			lines = append(lines, fmt.Sprintf("\tif !%s && %s {", matchedTemp, clause.cond))
			lines = append(lines, indentLines(branchLines, 2)...)
			lines = append(lines, "\t}")
		}
	}
	propagateLines, ok := g.lowerControlTransfer(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, fmt.Sprintf("\tif !%s {", matchedTemp))
	lines = append(lines, indentLines(propagateLines, 2)...)
	lines = append(lines, "\t}")
	lines = append(lines, "}")
	return lines, resultTemp, resultType, true
}

func (g *generator) coerceRescueBranch(ctx *compileContext, resultType string, expr string, exprType string) ([]string, string, bool) {
	if resultType == "" || exprType == "" || expr == "" {
		ctx.setReason("rescue clause type mismatch")
		return nil, "", false
	}
	lines, converted, ok := g.coerceJoinBranch(ctx, resultType, expr, exprType)
	if !ok {
		ctx.setReason("rescue clause type mismatch")
		return nil, "", false
	}
	return lines, converted, true
}

func (g *generator) compileRethrowStatement(ctx *compileContext, stmt *ast.RethrowStatement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing rethrow")
		return nil, false
	}
	if ctx != nil && ctx.rethrowControlVar != "" {
		lines, ok := g.lowerControlTransfer(ctx, ctx.rethrowControlVar)
		if !ok {
			return nil, false
		}
		return lines, true
	}
	if ctx != nil && ctx.rethrowVar != "" {
		lines, ok := g.lowerControlTransfer(ctx, g.raiseControlExpr("nil", ctx.rethrowVar))
		if !ok {
			return nil, false
		}
		return lines, true
	}
	lines, ok := g.lowerControlTransfer(ctx, `__able_raise_control(nil, runtime.ErrorValue{Message: "Unknown rethrow"})`)
	if !ok {
		return nil, false
	}
	return lines, true
}
