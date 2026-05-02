package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func linesReferenceLabel(lines []string, label string) bool {
	// Search for "break label" or "continue label" references in the body lines.
	breakRef := "break " + label
	continueRef := "continue " + label
	for _, line := range lines {
		if strings.Contains(line, breakRef) || strings.Contains(line, continueRef) {
			return true
		}
	}
	return false
}

func (g *generator) compileTailExpression(ctx *compileContext, expected string, expr ast.Expression) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing expression")
		return nil, "", "", false
	}
	if expected != "" && g.isVoidType(expected) {
		stmtLines, ok := g.compileStatement(ctx, expr)
		if !ok {
			return nil, "", "", false
		}
		return stmtLines, "struct{}{}", "struct{}", true
	}
	runtimeExpected := expected == "runtime.Value"
	switch e := expr.(type) {
	case *ast.AssignmentExpression:
		lines, valueExpr, valueType, ok := g.compileAssignment(ctx, e)
		if !ok {
			return nil, "", "", false
		}
		if runtimeExpected && valueType != "runtime.Value" {
			convLines, converted, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
			if !ok {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, "runtime.Value", true
		}
		if expected != "" && valueType == "runtime.Value" && expected != "runtime.Value" {
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, valueExpr, expected)
			if !ok {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		return lines, valueExpr, valueType, ok
	case *ast.IfExpression:
		lines, valueExpr, valueType, ok := g.compileIfExpression(ctx, e, expected)
		if !ok {
			return nil, "", "", false
		}
		if runtimeExpected && valueType != "runtime.Value" {
			convLines, converted, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
			if !ok {
				ctx.setReason("if expression type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, "runtime.Value", true
		}
		return lines, valueExpr, valueType, true
	case *ast.BlockExpression:
		lines, valueExpr, valueType, ok := g.compileBlockExpression(ctx, e, expected)
		if !ok {
			return nil, "", "", false
		}
		if runtimeExpected && valueType != "runtime.Value" {
			convLines, converted, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
			if !ok {
				ctx.setReason("block expression type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, "runtime.Value", true
		}
		return lines, valueExpr, valueType, true
	default:
		compileExpected := expected
		if runtimeExpected {
			compileExpected = ""
		}
		lines, valueExpr, valueType, ok := g.compileExprLines(ctx, expr, compileExpected)
		if !ok {
			return nil, "", "", false
		}
		if runtimeExpected && valueType != "runtime.Value" {
			convLines, converted, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
			if !ok {
				ctx.setReason("expression type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, "runtime.Value", true
		}
		return lines, valueExpr, valueType, true
	}
}

func (g *generator) compileCondition(ctx *compileContext, expr ast.Expression) ([]string, string, bool) {
	if expr == nil {
		ctx.setReason("missing condition")
		return nil, "", false
	}
	if assign, ok := expr.(*ast.AssignmentExpression); ok {
		if pattern, ok := assign.Left.(*ast.TypedPattern); ok {
			condLines, condExpr, condType, ok := g.compilePatternAssignment(ctx, assign, pattern)
			if !ok {
				return nil, "", false
			}
			if condBoolExpr, ok := g.staticTruthinessExpr(condExpr, condType); ok {
				return condLines, condBoolExpr, true
			}
			condRuntime := condExpr
			if condType != "runtime.Value" {
				convLines, converted, ok := g.lowerRuntimeValue(ctx, condExpr, condType)
				if !ok {
					ctx.setReason("condition unsupported")
					return nil, "", false
				}
				condLines = append(condLines, convLines...)
				condRuntime = converted
			}
			return condLines, fmt.Sprintf("__able_truthy(%s)", condRuntime), true
		}
	}
	condLines, condExpr, condType, ok := g.compileExprLines(ctx, expr, "")
	if !ok {
		return nil, "", false
	}
	if condBoolExpr, ok := g.staticTruthinessExpr(condExpr, condType); ok {
		return condLines, condBoolExpr, true
	}
	condRuntime := condExpr
	if condType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, condExpr, condType)
		if !ok {
			ctx.setReason("condition unsupported")
			return nil, "", false
		}
		condLines = append(condLines, convLines...)
		condRuntime = converted
	}
	return condLines, fmt.Sprintf("__able_truthy(%s)", condRuntime), true
}

func (g *generator) staticTruthinessExpr(expr string, goType string) (string, bool) {
	if g == nil || goType == "" {
		return "", false
	}
	if goType == "bool" {
		return expr, true
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return "", false
	}
	if g.staticFalsyErrorCarrierType(goType) {
		return "false", true
	}
	if g.isNilableStaticCarrierType(goType) {
		return fmt.Sprintf("%s != nil", expr), true
	}
	return "", false
}

func (g *generator) staticFalsyErrorCarrierType(goType string) bool {
	if g == nil || goType == "" || goType == "runtime.Value" || goType == "any" {
		return false
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return false
	}
	if innerType, nullable := g.nativeNullableValueInnerType(goType); nullable {
		return g.staticFalsyErrorCarrierType(innerType)
	}
	if g.isNativeErrorCarrierType(goType) {
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		return g.interfaceTypeExprSatisfies(iface.TypeExpr, "Error")
	}
	return false
}

func (g *generator) interfaceTypeExprSatisfies(expr ast.TypeExpression, interfaceName string) bool {
	if g == nil || expr == nil || interfaceName == "" {
		return false
	}
	pkgName := g.resolvedTypeExprPackage("", expr)
	baseName, ok := typeExprBaseName(expr)
	if !ok || baseName == "" {
		return false
	}
	if ifacePkg, _, _, _, ok := interfaceExprInfo(g, pkgName, expr); ok && ifacePkg != "" {
		pkgName = ifacePkg
	}
	if baseName == interfaceName {
		return true
	}
	if !g.isInterfaceName(baseName) {
		return false
	}
	for _, candidate := range g.interfaceSearchNamesForPackage(pkgName, baseName, make(map[string]struct{})) {
		if candidate == interfaceName {
			return true
		}
	}
	return false
}

func (g *generator) coerceIfBranch(ctx *compileContext, resultType string, expr string, exprType string) ([]string, string, bool) {
	if resultType == "" || exprType == "" {
		ctx.setReason("if branch type mismatch")
		return nil, "", false
	}
	lines, converted, ok := g.coerceJoinBranch(ctx, resultType, expr, exprType)
	if !ok {
		ctx.setReason("if branch type mismatch")
		return nil, "", false
	}
	return lines, converted, true
}

func (g *generator) compileIfExpression(ctx *compileContext, expr *ast.IfExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing if expression")
		return nil, "", "", false
	}
	if expr.IfCondition == nil || expr.IfBody == nil {
		ctx.setReason("incomplete if expression")
		return nil, "", "", false
	}
	condLines, condExpr, ok := g.compileCondition(ctx, expr.IfCondition)
	if !ok {
		return nil, "", "", false
	}
	// When expected is "" or "runtime.Value", compile branches without a type hint
	// so they infer their natural types. The result type is determined from agreement.
	explicitExpected := expected != "" && expected != "runtime.Value"
	compileExpected := expected
	if !explicitExpected {
		compileExpected = ""
	}
	bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(ctx.child(), compileExpected, expr.IfBody)
	if !ok {
		return nil, "", "", false
	}
	type ifBranch struct {
		condLines []string
		cond      string
		lines     []string
		expr      string
		exprType  string
		source    ast.Expression
		typeExpr  ast.TypeExpression
	}
	bodyTypeExpr, _ := g.inferExpressionTypeExpr(ctx.child(), expr.IfBody, bodyType)
	branches := []ifBranch{{condLines: condLines, cond: condExpr, lines: bodyLines, expr: bodyExpr, exprType: bodyType, source: expr.IfBody, typeExpr: bodyTypeExpr}}
	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			continue
		}
		if clause.Condition == nil || clause.Body == nil {
			ctx.setReason("incomplete else-if clause")
			return nil, "", "", false
		}
		clauseCondLines, clauseCondExpr, ok := g.compileCondition(ctx, clause.Condition)
		if !ok {
			return nil, "", "", false
		}
		clauseCtx := ctx.child()
		clauseLines, clauseExpr, clauseType, ok := g.compileTailExpression(clauseCtx, compileExpected, clause.Body)
		if !ok {
			return nil, "", "", false
		}
		clauseTypeExpr, _ := g.inferExpressionTypeExpr(clauseCtx, clause.Body, clauseType)
		branches = append(branches, ifBranch{condLines: clauseCondLines, cond: clauseCondExpr, lines: clauseLines, expr: clauseExpr, exprType: clauseType, source: clause.Body, typeExpr: clauseTypeExpr})
	}
	var elseLines []string
	elseExpr := ""
	elseType := ""
	var elseTypeExpr ast.TypeExpression
	if expr.ElseBody != nil {
		elseCtx := ctx.child()
		elseLines, elseExpr, elseType, ok = g.compileTailExpression(elseCtx, compileExpected, expr.ElseBody)
		if !ok {
			return nil, "", "", false
		}
		elseTypeExpr, _ = g.inferExpressionTypeExpr(elseCtx, expr.ElseBody, elseType)
	}
	// Determine result type
	resultType := expected
	if !explicitExpected {
		if expr.ElseBody == nil {
			resultType = "runtime.Value"
		} else {
			joinBranches := make([]joinBranchInfo, 0, len(branches)+1)
			for _, branch := range branches {
				joinBranches = append(joinBranches, joinBranchInfo{
					GoType:   branch.exprType,
					Expr:     branch.source,
					TypeExpr: branch.typeExpr,
					SawNil:   g.joinBranchIsNilExpr(branch.expr, branch.exprType),
				})
			}
			joinBranches = append(joinBranches, joinBranchInfo{
				GoType:   elseType,
				Expr:     expr.ElseBody,
				TypeExpr: elseTypeExpr,
				SawNil:   g.joinBranchIsNilExpr(elseExpr, elseType),
			})
			if joinedType, ok := g.lowerJoinCarrierFromBranches(ctx, joinBranches); ok {
				resultType = joinedType
			} else {
				resultType = "runtime.Value"
			}
		}
	}
	// Coerce all branches to result type
	for idx := range branches {
		b := &branches[idx]
		var coerceLines []string
		coerceLines, b.expr, ok = g.coerceIfBranch(ctx, resultType, b.expr, b.exprType)
		if !ok {
			return nil, "", "", false
		}
		b.lines = append(b.lines, coerceLines...)
	}
	if expr.ElseBody != nil {
		var coerceLines []string
		coerceLines, elseExpr, ok = g.coerceIfBranch(ctx, resultType, elseExpr, elseType)
		if !ok {
			return nil, "", "", false
		}
		elseLines = append(elseLines, coerceLines...)
	} else {
		if resultType != "runtime.Value" && resultType != "any" && !g.isVoidType(resultType) {
			ctx.setReason("if expression requires else")
			return nil, "", "", false
		}
		elseExpr = safeNilReturnExpr(resultType)
		if wrapped, ok := g.nativeUnionNilExpr(resultType); ok {
			elseExpr = wrapped
		}
	}
	temp := ctx.newTemp()
	lines := []string{fmt.Sprintf("var %s %s", temp, resultType)}
	lines = append(lines, branches[0].condLines...)
	lines = append(lines, fmt.Sprintf("if %s {", branches[0].cond))
	lines = append(lines, indentLines(branches[0].lines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", temp, branches[0].expr))
	for idx := 1; idx < len(branches); idx++ {
		branch := branches[idx]
		if len(branch.condLines) > 0 {
			lines = append(lines, "} else {")
			lines = append(lines, indentLines(branch.condLines, 1)...)
			lines = append(lines, fmt.Sprintf("\tif %s {", branch.cond))
			lines = append(lines, indentLines(branch.lines, 2)...)
			lines = append(lines, fmt.Sprintf("\t\t%s = %s", temp, branch.expr))
		} else {
			lines = append(lines, fmt.Sprintf("} else if %s {", branch.cond))
			lines = append(lines, indentLines(branch.lines, 1)...)
			lines = append(lines, fmt.Sprintf("\t%s = %s", temp, branch.expr))
		}
	}
	// Close any nested else blocks from branches with condLines
	closingBraces := 0
	for idx := 1; idx < len(branches); idx++ {
		if len(branches[idx].condLines) > 0 {
			closingBraces++
		}
	}
	lines = append(lines, "} else {")
	lines = append(lines, indentLines(elseLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", temp, elseExpr))
	lines = append(lines, "}")
	for i := 0; i < closingBraces; i++ {
		lines = append(lines, "}")
	}
	return lines, temp, resultType, true
}

func (g *generator) compileBlockExpression(ctx *compileContext, block *ast.BlockExpression, expected string) ([]string, string, string, bool) {
	if block == nil {
		ctx.setReason("missing block expression")
		return nil, "", "", false
	}
	child := ctx.child()
	if child != nil {
		child.blockStatements = block.Body
		switch {
		case expected == "", expected == "runtime.Value", expected == "any":
			child.expectedTypeExpr = nil
		default:
			expectedTypeExpr := ctx.expectedTypeExpr
			if expectedTypeExpr == nil {
				expectedTypeExpr = ctx.returnTypeExpr
			}
			child.expectedTypeExpr = g.concretizedExpectedTypeExpr(child, expected, expectedTypeExpr)
		}
	}
	if len(block.Body) == 0 {
		if expected == "" || g.isVoidType(expected) {
			return nil, "struct{}{}", "struct{}", true
		}
		if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, expected); ok {
			return nil, successExpr, expected, true
		}
		ctx.setReason("empty block requires void return")
		return nil, "", "", false
	}
	needsScope := len(block.Body) > 1
	wrapScope := func(lines []string, expr string, goType string) ([]string, string, string, bool) {
		if !needsScope {
			return lines, expr, goType, true
		}
		resultTemp := ctx.newTemp()
		wrapped := []string{fmt.Sprintf("var %s %s", resultTemp, goType)}
		wrapped = append(wrapped, "{")
		wrapped = append(wrapped, indentLines(lines, 1)...)
		wrapped = append(wrapped, fmt.Sprintf("\t%s = %s", resultTemp, expr))
		wrapped = append(wrapped, "}")
		return wrapped, resultTemp, goType, true
	}
	lines := make([]string, 0, len(block.Body)+1)
	for idx, stmt := range block.Body {
		child.statementIndex = idx
		isLast := idx == len(block.Body)-1
		if isLast {
			if raiseStmt, ok := stmt.(*ast.RaiseStatement); ok {
				stmtLines, ok := g.compileRaiseStatement(child, raiseStmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				// Raise panics — expression is unreachable but needed for type
				returnType := "runtime.Value"
				returnExpr := "runtime.NilValue{}"
				switch {
				case expected != "" && g.isVoidType(expected):
					returnType = "struct{}"
					returnExpr = "struct{}{}"
				case expected != "" && expected != "runtime.Value":
					zeroExpr, ok := g.zeroValueExpr(expected)
					if !ok {
						ctx.setReason("missing block return expression")
						return nil, "", "", false
					}
					returnType = expected
					returnExpr = zeroExpr
				}
				return wrapScope(lines, returnExpr, returnType)
			}
			if rethrowStmt, ok := stmt.(*ast.RethrowStatement); ok {
				stmtLines, ok := g.compileRethrowStatement(child, rethrowStmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				// Rethrow panics — expression is unreachable but needed for type
				returnType := "runtime.Value"
				returnExpr := "runtime.NilValue{}"
				switch {
				case expected != "" && g.isVoidType(expected):
					returnType = "struct{}"
					returnExpr = "struct{}{}"
				case expected != "" && expected != "runtime.Value":
					zeroExpr, ok := g.zeroValueExpr(expected)
					if !ok {
						ctx.setReason("missing block return expression")
						return nil, "", "", false
					}
					returnType = expected
					returnExpr = zeroExpr
				}
				return wrapScope(lines, returnExpr, returnType)
			}
			expr, ok := stmt.(ast.Expression)
			if !ok || expr == nil {
				// Handle return statements as block-ending statements.
				// The return exits the enclosing function; the block value is unreachable.
				if ret, ok := stmt.(*ast.ReturnStatement); ok && ret != nil {
					retLines, ok := g.compileStatement(child, ret)
					if !ok {
						return nil, "", "", false
					}
					lines = append(lines, retLines...)
					returnType := expected
					if returnType == "" {
						returnType = "runtime.Value"
					}
					returnExpr, ok := g.zeroValueExpr(returnType)
					if !ok {
						returnExpr = "nil"
						returnType = "any"
					}
					return wrapScope(lines, returnExpr, returnType)
				}
				if loop, ok := stmt.(*ast.ForLoop); ok && (expected == "" || expected == "runtime.Value" || expected == "any") {
					loopLines, loopResult, ok := g.compileForLoopInternal(child, loop, true)
					if !ok {
						return nil, "", "", false
					}
					lines = append(lines, loopLines...)
					returnExpr := loopResult
					if returnExpr == "" {
						returnExpr = "runtime.VoidValue{}"
					}
					return wrapScope(lines, returnExpr, "runtime.Value")
				}
				if expected != "" && !g.isVoidType(expected) && expected != "runtime.Value" && expected != "any" {
					ctx.setReason("missing block return expression")
					return nil, "", "", false
				}
				stmtLines, ok := g.compileStatement(child, stmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				returnType := "struct{}"
				returnExpr := "struct{}{}"
				switch {
				case expected == "" || expected == "runtime.Value":
					returnType = "runtime.Value"
					returnExpr = "runtime.VoidValue{}"
				case expected == "any":
					returnType = "any"
					returnExpr = "nil"
				}
				return wrapScope(lines, returnExpr, returnType)
			}
			returnLines, returnExpr, returnType, ok := g.compileTailExpression(child, expected, expr)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, returnLines...)
			return wrapScope(lines, returnExpr, returnType)
		}
		stmtLines, ok := g.compileStatement(child, stmt)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, stmtLines...)
	}
	ctx.setReason("missing block return expression")
	return nil, "", "", false
}

func indentLines(lines []string, tabs int) []string {
	if len(lines) == 0 || tabs <= 0 {
		return lines
	}
	prefix := strings.Repeat("\t", tabs)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, prefix+line)
	}
	return out
}

func (g *generator) compileBlockStatement(ctx *compileContext, block *ast.BlockExpression) ([]string, bool) {
	if block == nil {
		ctx.setReason("missing block")
		return nil, false
	}
	child := ctx.child()
	child.blockStatements = block.Body
	lines := make([]string, 0, len(block.Body))
	for idx, stmt := range block.Body {
		child.statementIndex = idx
		stmtLines, ok := g.compileStatement(child, stmt)
		if !ok {
			ctx.setReason(child.reason)
			return nil, false
		}
		lines = append(lines, stmtLines...)
	}
	return lines, true
}

func (g *generator) compileIfStatement(ctx *compileContext, expr *ast.IfExpression) ([]string, bool) {
	if expr == nil {
		ctx.setReason("missing if statement")
		return nil, false
	}
	if expr.IfCondition == nil || expr.IfBody == nil {
		ctx.setReason("incomplete if statement")
		return nil, false
	}
	condLines, condExpr, ok := g.compileCondition(ctx, expr.IfCondition)
	if !ok {
		return nil, false
	}
	bodyCtx := ctx.child()
	bodyLines, ok := g.compileBlockStatement(bodyCtx, expr.IfBody)
	if !ok {
		ctx.setReason(bodyCtx.reason)
		return nil, false
	}
	var lines []string
	lines = append(lines, condLines...)
	lines = append(lines, fmt.Sprintf("if %s {", condExpr))
	lines = append(lines, indentLines(bodyLines, 1)...)
	closingBraces := 0
	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			continue
		}
		if clause.Condition == nil || clause.Body == nil {
			ctx.setReason("incomplete else-if clause")
			return nil, false
		}
		clauseCondLines, clauseCondExpr, ok := g.compileCondition(ctx, clause.Condition)
		if !ok {
			return nil, false
		}
		clauseCtx := ctx.child()
		clauseLines, ok := g.compileBlockStatement(clauseCtx, clause.Body)
		if !ok {
			ctx.setReason(clauseCtx.reason)
			return nil, false
		}
		if len(clauseCondLines) > 0 {
			lines = append(lines, "} else {")
			lines = append(lines, indentLines(clauseCondLines, 1)...)
			lines = append(lines, fmt.Sprintf("\tif %s {", clauseCondExpr))
			lines = append(lines, indentLines(clauseLines, 2)...)
			closingBraces++
		} else {
			lines = append(lines, fmt.Sprintf("} else if %s {", clauseCondExpr))
			lines = append(lines, indentLines(clauseLines, 1)...)
		}
	}
	if expr.ElseBody != nil {
		elseCtx := ctx.child()
		elseLines, ok := g.compileBlockStatement(elseCtx, expr.ElseBody)
		if !ok {
			ctx.setReason(elseCtx.reason)
			return nil, false
		}
		lines = append(lines, "} else {")
		lines = append(lines, indentLines(elseLines, 1)...)
	}
	lines = append(lines, "}")
	for i := 0; i < closingBraces; i++ {
		lines = append(lines, "}")
	}
	g.seedFactsAfterTerminatingIf(ctx, expr)
	return lines, true
}

func (g *generator) compileWhileLoop(ctx *compileContext, loop *ast.WhileLoop) ([]string, bool) {
	if loop == nil || loop.Condition == nil || loop.Body == nil {
		ctx.setReason("missing while loop")
		return nil, false
	}
	condLines, condExpr, ok := g.compileCondition(ctx, loop.Condition)
	if !ok {
		return nil, false
	}
	loopLabelName := ctx.newTemp()
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	bodyCtx.loopBreakProbe = nil
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, false
	}
	forLine := "for {"
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for {", loopLabelName)
	}
	loopLines := []string{forLine}
	loopLines = append(loopLines, indentLines(condLines, 1)...)
	loopLines = append(loopLines, fmt.Sprintf("if !%s { break }", condExpr))
	loopLines = append(loopLines, indentLines(bodyLines, 1)...)
	loopLines = append(loopLines, "}")
	return loopLines, true
}

func (g *generator) compileForLoop(ctx *compileContext, loop *ast.ForLoop) ([]string, bool) {
	lines, _, ok := g.compileForLoopInternal(ctx, loop, false)
	return lines, ok
}

func (g *generator) compileForLoopInternal(ctx *compileContext, loop *ast.ForLoop, withResult bool) ([]string, string, bool) {
	if loop == nil || loop.Iterable == nil || loop.Body == nil {
		ctx.setReason("missing for loop")
		return nil, "", false
	}
	iterLines, iterExpr, iterType, ok := g.compileExprLines(ctx, loop.Iterable, "")
	if !ok {
		return nil, "", false
	}
	if g.isStaticArrayType(iterType) {
		return g.compileStaticArrayForLoopInternal(ctx, loop, withResult, iterLines, iterExpr, iterType)
	}
	if lines, result, ok := g.compileStaticIterableForLoopInternal(ctx, loop, withResult, loop.Iterable, iterLines, iterExpr, iterType); ok {
		return lines, result, true
	}
	iterConvLines, iterRuntime, ok := g.lowerRuntimeValue(ctx, iterExpr, iterType)
	if !ok {
		ctx.setReason("for loop iterable unsupported")
		return nil, "", false
	}
	elementTemp := ctx.newTemp()
	loopLabelName := ctx.newTemp()
	resultTemp := ""
	if withResult {
		resultTemp = ctx.newTemp()
	}
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	if withResult {
		bodyCtx.loopBreakValueTemp = resultTemp
		bodyCtx.loopBreakValueType = "runtime.Value"
	}
	bodyCtx.loopBreakProbe = nil
	newNames := map[string]struct{}{}
	collectPatternBindingNames(loop.Pattern, newNames)
	mode := patternBindingMode{declare: true, newNames: newNames}
	condLines, cond, ok := g.compileMatchPatternCondition(bodyCtx, loop.Pattern, elementTemp, "runtime.Value")
	if !ok {
		return nil, "", false
	}
	bindLines, ok := g.compileAssignmentPatternBindings(bodyCtx, loop.Pattern, elementTemp, "runtime.Value", mode)
	if !ok {
		return nil, "", false
	}
	if cond != "true" || len(condLines) > 0 {
		mismatchLine := fmt.Sprintf("break %s", loopLabelName)
		if withResult {
			mismatchLine = fmt.Sprintf("%s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"}; %s", resultTemp, mismatchLine)
		}
		var condPrefix []string
		condPrefix = append(condPrefix, condLines...)
		condPrefix = append(condPrefix, fmt.Sprintf("if !(%s) { %s }", cond, mismatchLine))
		bindLines = append(condPrefix, bindLines...)
	}
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, "", false
	}
	// Build inner body: bind + body
	innerLines := append(bindLines, bodyLines...)
	iterTemp := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	isArrayTemp := ctx.newTemp()
	iteratorTemp := ctx.newTemp()
	idxTemp := ctx.newTemp()
	doneTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := append([]string{}, iterLines...)
	lines = append(lines, iterConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", iterTemp, iterRuntime))
	if withResult {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.VoidValue{}", resultTemp))
	}
	lines = append(lines,
		fmt.Sprintf("%s, %s := __able_array_values(%s)", valuesTemp, isArrayTemp, iterTemp),
		fmt.Sprintf("var %s *runtime.IteratorValue", iteratorTemp),
		fmt.Sprintf("if !%s { %s = __able_resolve_iterator(%s); defer %s.Close() }", isArrayTemp, iteratorTemp, iterTemp, iteratorTemp),
		fmt.Sprintf("var %s runtime.Value", elementTemp),
		fmt.Sprintf("%s := 0", idxTemp),
	)
	forPrefix := "for {"
	if linesReferenceLabel(innerLines, loopLabelName) {
		forPrefix = fmt.Sprintf("%s: for {", loopLabelName)
	}
	lines = append(lines,
		forPrefix,
		fmt.Sprintf("if %s {", isArrayTemp),
		fmt.Sprintf("\tif %s >= %s { break }", idxTemp, g.staticSliceLenExpr(valuesTemp)),
		fmt.Sprintf("\t%s = %s[%s]", elementTemp, valuesTemp, idxTemp),
		fmt.Sprintf("\t%s++", idxTemp),
		"} else {",
		fmt.Sprintf("\tvar %s bool", doneTemp),
		fmt.Sprintf("\tvar %s error", errTemp),
		fmt.Sprintf("\t%s, %s, %s = %s.Next()", elementTemp, doneTemp, errTemp, iteratorTemp),
		fmt.Sprintf("\tif %s != nil { panic(%s) }", errTemp, errTemp),
		fmt.Sprintf("\tif %s { break }", doneTemp),
		"}",
	)
	lines = append(lines, indentLines(innerLines, 1)...)
	lines = append(lines, "}")
	return lines, resultTemp, true
}

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
