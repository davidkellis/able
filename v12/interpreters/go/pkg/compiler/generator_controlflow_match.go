package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) narrowedNativeUnionSubjectType(ctx *compileContext, subjectType string, pattern ast.Pattern) string {
	union := g.nativeUnionInfoForGoType(subjectType)
	if g == nil || union == nil || pattern == nil {
		return subjectType
	}
	removeType := ""
	switch p := pattern.(type) {
	case *ast.StructPattern:
		if p != nil && p.StructType != nil && p.StructType.Name != "" {
			if mapped, ok := g.mapTypeExpressionInContext(ctx, ast.Ty(p.StructType.Name)); ok {
				removeType = mapped
			}
		}
	case *ast.TypedPattern:
		if p != nil && p.TypeAnnotation != nil {
			if target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, p.TypeAnnotation); ok && target.Member != nil {
				removeType = target.Member.GoType
			}
		}
	case *ast.Identifier:
		if p != nil && p.Name != "" && g.isSingletonPattern(ctx, p.Name) {
			if mapped, ok := g.mapTypeExpressionInContext(ctx, ast.Ty(p.Name)); ok {
				removeType = mapped
			}
		}
	}
	if removeType == "" {
		return subjectType
	}
	remaining := make([]string, 0, len(union.Members))
	for _, member := range union.Members {
		if member == nil || member.GoType == removeType {
			continue
		}
		remaining = append(remaining, member.GoType)
	}
	if len(remaining) == 1 {
		return remaining[0]
	}
	return subjectType
}

func (g *generator) narrowedNativeUnionSubjectExpr(ctx *compileContext, originalSubjectExpr string, originalSubjectType string, narrowedType string) ([]string, string) {
	if g == nil || ctx == nil || originalSubjectExpr == "" || originalSubjectType == "" || narrowedType == "" || narrowedType == originalSubjectType {
		return nil, originalSubjectExpr
	}
	union := g.nativeUnionInfoForGoType(originalSubjectType)
	if union == nil {
		return nil, originalSubjectExpr
	}
	member, ok := g.nativeUnionMember(union, narrowedType)
	if !ok || member == nil {
		return nil, originalSubjectExpr
	}
	narrowedTemp := ctx.newTemp()
	return []string{fmt.Sprintf("%s, _ := %s(%s)", narrowedTemp, member.UnwrapHelper, originalSubjectExpr)}, narrowedTemp
}

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
	clauseTypes := make([]string, 0, len(match.Clauses))
	type matchClause struct {
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
	clauses := make([]matchClause, 0, len(match.Clauses))
	clauseSubjectType := subjectType
	for _, clause := range match.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		clauseSubjectLines, clauseSubjectExpr := g.narrowedNativeUnionSubjectExpr(clauseCtx, subjectTemp, subjectType, clauseSubjectType)
		condLines, cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, clauseSubjectExpr, clauseSubjectType)
		if !ok {
			return nil, "", "", false
		}
		if cond == "true" && len(condLines) == 0 && len(bindLines) == 0 {
			clauseSubjectLines = nil
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
		bodyTypeExpr, _ := g.inferExpressionTypeExpr(clauseCtx, clause.Body, bodyType)
		if explicitExpected {
			if !g.typeMatches(resultType, bodyType) {
				ctx.setReason(fmt.Sprintf("match clause type mismatch (%s != %s, subject=%s, pattern=%T)", bodyType, resultType, clauseSubjectType, clause.Pattern))
				return nil, "", "", false
			}
		} else {
			if inferredType == "" {
				inferredType = bodyType
			} else if bodyType != inferredType {
				mismatch = true
			}
		}
		clauseTypes = append(clauseTypes, bodyType)
		clauses = append(clauses, matchClause{
			condLines:    append(clauseSubjectLines, condLines...),
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
		clauseSubjectType = g.narrowedNativeUnionSubjectType(clauseCtx, clauseSubjectType, clause.Pattern)
	}
	if !explicitExpected {
		joinBranches := make([]joinBranchInfo, 0, len(clauses))
		for _, clause := range clauses {
			joinBranches = append(joinBranches, joinBranchInfo{
				GoType:   clause.bodyType,
				Expr:     clause.bodyNode,
				TypeExpr: clause.bodyTypeExpr,
				SawNil:   g.joinBranchIsNilExpr(clause.bodyExpr, clause.bodyType),
			})
		}
		if joinedType, ok := g.joinResultTypeFromBranches(ctx, joinBranches); ok {
			resultType = joinedType
		} else if inferredType != "" && !mismatch {
			resultType = inferredType
		} else {
			resultType = "runtime.Value"
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
		coerceLines, converted, ok := g.coerceJoinBranch(ctx, resultType, clause.bodyExpr, clause.bodyType)
		if !ok {
			ctx.setReason(fmt.Sprintf("match clause type mismatch (%s != %s)", clause.bodyType, resultType))
			return nil, "", "", false
		}
		clause.bodyLines = append(clause.bodyLines, coerceLines...)
		clause.bodyExpr = converted
		clause.bodyType = resultType
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
	clauseSubjectType := subjectType
	for _, clause := range match.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		clauseSubjectLines, clauseSubjectExpr := g.narrowedNativeUnionSubjectExpr(clauseCtx, subjectTemp, subjectType, clauseSubjectType)
		condLines, cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, clauseSubjectExpr, clauseSubjectType)
		if !ok {
			ctx.setReason(clauseCtx.reason)
			return nil, false
		}
		if cond == "true" && len(condLines) == 0 && len(bindLines) == 0 {
			clauseSubjectLines = nil
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
			lines = append(lines, indentLines(append(clauseSubjectLines, condLines...), 1)...)
			lines = append(lines, fmt.Sprintf("\tif %s {", cond))
			lines = append(lines, indentLines(branchLines, 2)...)
			lines = append(lines, "\t}")
			lines = append(lines, "}")
		} else {
			lines = append(lines, fmt.Sprintf("if !%s && %s {", matchedTemp, cond))
			if len(clauseSubjectLines) > 0 {
				lines[len(lines)-1] = fmt.Sprintf("if !%s {", matchedTemp)
				lines = append(lines, indentLines(clauseSubjectLines, 1)...)
				lines = append(lines, fmt.Sprintf("\tif %s {", cond))
				lines = append(lines, indentLines(branchLines, 2)...)
				lines = append(lines, "\t}")
			} else {
				lines = append(lines, indentLines(branchLines, 1)...)
			}
			lines = append(lines, "}")
		}
		clauseSubjectType = g.narrowedNativeUnionSubjectType(clauseCtx, clauseSubjectType, clause.Pattern)
	}
	lines = append(lines, fmt.Sprintf("if !%s { bridge.RaiseRuntimeErrorWithContext(__able_runtime, %s, fmt.Errorf(\"Non-exhaustive match\")) }", matchedTemp, matchNode))
	return lines, true
}
