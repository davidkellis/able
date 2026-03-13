package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileMatchExpression(ctx *compileContext, match *ast.MatchExpression, expected string) ([]string, string, string, bool) {
	if match == nil || match.Subject == nil {
		ctx.setReason("missing match expression")
		return nil, "", "", false
	}
	subjectLines, subjectExpr, subjectType, ok := g.compileExprLines(ctx, match.Subject, "")
	if !ok {
		return nil, "", "", false
	}
	subjectTemp := ctx.newTemp()
	resultType := expected
	explicitExpected := expected != "" && expected != "runtime.Value"
	compileExpected := expected
	if !explicitExpected {
		compileExpected = ""
	}
	inferredType := ""
	mismatch := false
	type matchClause struct {
		condLines  []string
		cond       string
		bindLines  []string
		guardLines []string
		guardExpr  string
		bodyLines  []string
		bodyExpr   string
		bodyType   string
	}
	clauses := make([]matchClause, 0, len(match.Clauses))
	for _, clause := range match.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		condLines, cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, subjectTemp, subjectType)
		if !ok {
			return nil, "", "", false
		}
		guardExpr := ""
		var guardLines []string
		if clause.Guard != nil {
			gl, gv, ok := g.compileCondition(clauseCtx, clause.Guard)
			if !ok {
				return nil, "", "", false
			}
			guardLines = gl
			guardExpr = gv
		}
		bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(clauseCtx, compileExpected, clause.Body)
		if !ok {
			return nil, "", "", false
		}
		if explicitExpected {
			if !g.typeMatches(resultType, bodyType) {
				ctx.setReason("match clause type mismatch")
				return nil, "", "", false
			}
		} else {
			if inferredType == "" {
				inferredType = bodyType
			} else if bodyType != inferredType {
				mismatch = true
			}
		}
		clauses = append(clauses, matchClause{
			condLines:  condLines,
			cond:       cond,
			bindLines:  bindLines,
			guardLines: guardLines,
			guardExpr:  guardExpr,
			bodyLines:  bodyLines,
			bodyExpr:   bodyExpr,
			bodyType:   bodyType,
		})
	}
	if !explicitExpected {
		if inferredType == "" || mismatch {
			resultType = "runtime.Value"
		} else {
			resultType = inferredType
		}
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	for idx := range clauses {
		clause := &clauses[idx]
		if clause.bodyType == resultType {
			if wrapLines, wrapped, ok := g.nativeUnionWrapLines(ctx, resultType, clause.bodyType, clause.bodyExpr); ok {
				clause.bodyLines = append(clause.bodyLines, wrapLines...)
				clause.bodyExpr = wrapped
				clause.bodyType = resultType
			}
			continue
		}
		if g.typeMatches(resultType, clause.bodyType) {
			if wrapLines, wrapped, ok := g.nativeUnionWrapLines(ctx, resultType, clause.bodyType, clause.bodyExpr); ok {
				clause.bodyLines = append(clause.bodyLines, wrapLines...)
				clause.bodyExpr = wrapped
				clause.bodyType = resultType
			}
			continue
		}
		switch {
		case resultType == "runtime.Value" && clause.bodyType != "runtime.Value":
			convLines, converted, ok := g.runtimeValueLines(ctx, clause.bodyExpr, clause.bodyType)
			if !ok {
				ctx.setReason("match clause type mismatch")
				return nil, "", "", false
			}
			clause.bodyLines = append(clause.bodyLines, convLines...)
			clause.bodyExpr = converted
			clause.bodyType = resultType
		case resultType == "any":
			clause.bodyType = resultType
		case clause.bodyType == "runtime.Value" && resultType != "runtime.Value":
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, clause.bodyExpr, resultType)
			if !ok {
				ctx.setReason("match clause type mismatch")
				return nil, "", "", false
			}
			clause.bodyLines = append(clause.bodyLines, convLines...)
			clause.bodyExpr = converted
			clause.bodyType = resultType
		case clause.bodyType == "any":
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, clause.bodyExpr, resultType)
			if !ok {
				ctx.setReason("match clause type mismatch")
				return nil, "", "", false
			}
			clause.bodyLines = append(clause.bodyLines, convLines...)
			clause.bodyExpr = converted
			clause.bodyType = resultType
		default:
			ctx.setReason("match clause type mismatch")
			return nil, "", "", false
		}
	}
	matchedTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	matchNode := g.diagNodeName(match, "*ast.MatchExpression", "match")
	lines := append([]string{}, subjectLines...)
	lines = append(lines,
		fmt.Sprintf("%s := %s", subjectTemp, subjectExpr),
		fmt.Sprintf("%s := false", matchedTemp),
		fmt.Sprintf("var %s %s", resultTemp, resultType),
	)
	for _, clause := range clauses {
		branchLines := []string{}
		branchLines = append(branchLines, clause.bindLines...)
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
			lines = append(lines, fmt.Sprintf("if !%s {", matchedTemp))
			lines = append(lines, indentLines(clause.condLines, 1)...)
			lines = append(lines, fmt.Sprintf("\tif %s {", clause.cond))
			lines = append(lines, indentLines(branchLines, 2)...)
			lines = append(lines, "\t}")
			lines = append(lines, "}")
		} else {
			lines = append(lines, fmt.Sprintf("if !%s && %s {", matchedTemp, clause.cond))
			lines = append(lines, indentLines(branchLines, 1)...)
			lines = append(lines, "}")
		}
	}
	lines = append(lines, fmt.Sprintf("if !%s { bridge.RaiseRuntimeErrorWithContext(__able_runtime, %s, fmt.Errorf(\"Non-exhaustive match\")) }", matchedTemp, matchNode))
	return lines, resultTemp, resultType, true
}

func (g *generator) compileMatchStatement(ctx *compileContext, match *ast.MatchExpression) ([]string, bool) {
	if match == nil || match.Subject == nil {
		ctx.setReason("missing match expression")
		return nil, false
	}
	subjectLines, subjectExpr, subjectType, ok := g.compileExprLines(ctx, match.Subject, "")
	if !ok {
		return nil, false
	}
	subjectTemp := ctx.newTemp()
	matchedTemp := ctx.newTemp()
	matchNode := g.diagNodeName(match, "*ast.MatchExpression", "match")
	lines := append([]string{}, subjectLines...)
	lines = append(lines,
		fmt.Sprintf("%s := %s", subjectTemp, subjectExpr),
		fmt.Sprintf("%s := false", matchedTemp),
	)
	for _, clause := range match.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		condLines, cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, subjectTemp, subjectType)
		if !ok {
			ctx.setReason(clauseCtx.reason)
			return nil, false
		}
		guardExpr := ""
		var guardLines []string
		if clause.Guard != nil {
			guardLines, guardExpr, ok = g.compileCondition(clauseCtx, clause.Guard)
			if !ok {
				ctx.setReason(clauseCtx.reason)
				return nil, false
			}
		}
		bodyLines := []string{}
		if block, ok := clause.Body.(*ast.BlockExpression); ok && block != nil {
			bodyLines, ok = g.compileBlockStatement(clauseCtx, block)
			if !ok {
				ctx.setReason(clauseCtx.reason)
				return nil, false
			}
		} else {
			exprLines, expr, _, ok := g.compileTailExpression(clauseCtx, "", clause.Body)
			if !ok {
				ctx.setReason(clauseCtx.reason)
				return nil, false
			}
			bodyLines = append(bodyLines, exprLines...)
			if expr != "" {
				bodyLines = append(bodyLines, fmt.Sprintf("_ = %s", expr))
			}
		}
		branchLines := append([]string{}, bindLines...)
		if guardExpr != "" {
			branchLines = append(branchLines, guardLines...)
			branchLines = append(branchLines, fmt.Sprintf("if %s {", guardExpr))
			branchLines = append(branchLines, indentLines(bodyLines, 1)...)
			branchLines = append(branchLines, fmt.Sprintf("\t%s = true", matchedTemp))
			branchLines = append(branchLines, "}")
		} else {
			branchLines = append(branchLines, bodyLines...)
			branchLines = append(branchLines, fmt.Sprintf("%s = true", matchedTemp))
		}
		if len(condLines) > 0 {
			lines = append(lines, fmt.Sprintf("if !%s {", matchedTemp))
			lines = append(lines, indentLines(condLines, 1)...)
			lines = append(lines, fmt.Sprintf("\tif %s {", cond))
			lines = append(lines, indentLines(branchLines, 2)...)
			lines = append(lines, "\t}")
			lines = append(lines, "}")
		} else {
			lines = append(lines, fmt.Sprintf("if !%s && %s {", matchedTemp, cond))
			lines = append(lines, indentLines(branchLines, 1)...)
			lines = append(lines, "}")
		}
	}
	lines = append(lines, fmt.Sprintf("if !%s { bridge.RaiseRuntimeErrorWithContext(__able_runtime, %s, fmt.Errorf(\"Non-exhaustive match\")) }", matchedTemp, matchNode))
	return lines, true
}
