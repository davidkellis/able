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
	runtimeExpected := expected == "runtime.Value"
	switch e := expr.(type) {
	case *ast.AssignmentExpression:
		lines, valueExpr, valueType, ok := g.compileAssignment(ctx, e)
		if !ok {
			return nil, "", "", false
		}
		if runtimeExpected && valueType != "runtime.Value" {
			convLines, converted, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
			if !ok {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, "runtime.Value", true
		}
		if expected != "" && valueType == "runtime.Value" && expected != "runtime.Value" {
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueExpr, expected)
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
			convLines, converted, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
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
			convLines, converted, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
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
			convLines, converted, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
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
	condLines, condExpr, condType, ok := g.compileExprLines(ctx, expr, "")
	if !ok {
		return nil, "", false
	}
	if condType == "bool" {
		return condLines, condExpr, true
	}
	condRuntime := condExpr
	if condType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, condExpr, condType)
		if !ok {
			ctx.setReason("condition unsupported")
			return nil, "", false
		}
		condLines = append(condLines, convLines...)
		condRuntime = converted
	}
	return condLines, fmt.Sprintf("__able_truthy(%s)", condRuntime), true
}

func (g *generator) coerceIfBranch(ctx *compileContext, resultType string, expr string, exprType string) ([]string, string, bool) {
	if resultType == "" || exprType == "" {
		ctx.setReason("if branch type mismatch")
		return nil, "", false
	}
	if g.typeMatches(resultType, exprType) {
		return nil, expr, true
	}
	if resultType == "runtime.Value" && exprType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, expr, exprType)
		if !ok {
			ctx.setReason("if branch type mismatch")
			return nil, "", false
		}
		return convLines, converted, true
	}
	if resultType == "any" {
		// any accepts all types without conversion
		return nil, expr, true
	}
	if resultType != "runtime.Value" && exprType == "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, expr, resultType)
		if !ok {
			ctx.setReason("if branch type mismatch")
			return nil, "", false
		}
		return convLines, converted, true
	}
	if exprType == "any" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, expr, resultType)
		if !ok {
			ctx.setReason("if branch type mismatch")
			return nil, "", false
		}
		return convLines, converted, true
	}
	ctx.setReason("if branch type mismatch")
	return nil, "", false
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
	}
	branches := []ifBranch{{condLines: condLines, cond: condExpr, lines: bodyLines, expr: bodyExpr, exprType: bodyType}}
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
		clauseLines, clauseExpr, clauseType, ok := g.compileTailExpression(ctx.child(), compileExpected, clause.Body)
		if !ok {
			return nil, "", "", false
		}
		branches = append(branches, ifBranch{condLines: clauseCondLines, cond: clauseCondExpr, lines: clauseLines, expr: clauseExpr, exprType: clauseType})
	}
	var elseLines []string
	elseExpr := ""
	elseType := ""
	if expr.ElseBody != nil {
		elseLines, elseExpr, elseType, ok = g.compileTailExpression(ctx.child(), compileExpected, expr.ElseBody)
		if !ok {
			return nil, "", "", false
		}
	}
	// Determine result type
	resultType := expected
	if !explicitExpected {
		if expr.ElseBody == nil {
			resultType = "runtime.Value"
		} else {
			// Infer from branch type agreement (like match expressions)
			inferredType := bodyType
			mismatch := false
			for _, b := range branches[1:] {
				if b.exprType != inferredType {
					mismatch = true
					break
				}
			}
			if !mismatch && elseType != inferredType {
				mismatch = true
			}
			if inferredType == "" || mismatch {
				resultType = "runtime.Value"
			} else {
				resultType = inferredType
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
	if len(block.Body) == 0 {
		if expected == "" || g.isVoidType(expected) {
			return nil, "struct{}{}", "struct{}", true
		}
		if expected == "runtime.Value" {
			return nil, "runtime.VoidValue{}", "runtime.Value", true
		}
		if expected == "any" {
			return nil, "runtime.VoidValue{}", "any", true
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
				if loop, ok := stmt.(*ast.ForLoop); ok && (expected == "" || expected == "runtime.Value") {
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
				if expected != "" && !g.isVoidType(expected) && expected != "runtime.Value" {
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
				if expected == "" || expected == "runtime.Value" {
					returnType = "runtime.Value"
					returnExpr = "runtime.VoidValue{}"
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
	lines := make([]string, 0, len(block.Body))
	for _, stmt := range block.Body {
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
	iterConvLines, iterRuntime, ok := g.runtimeValueLines(ctx, iterExpr, iterType)
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
	}
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
		fmt.Sprintf("\tif %s >= len(%s) { break }", idxTemp, valuesTemp),
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
		valueExpr := "runtime.NilValue{}"
		if stmt.Value != nil {
			valLines, expr, goType, ok := g.compileExprLines(ctx, stmt.Value, "")
			if !ok {
				return nil, false
			}
			convLines, valueRuntime, ok := g.runtimeValueLines(ctx, expr, goType)
			if !ok {
				ctx.setReason("break value unsupported")
				return nil, false
			}
			valueExpr = valueRuntime
			if len(valLines) > 0 || len(convLines) > 0 {
				result := append([]string{}, valLines...)
				result = append(result, convLines...)
				result = append(result,
					fmt.Sprintf("%s = %s", resultTemp, valueExpr),
					fmt.Sprintf("break %s", goLabel),
				)
				return result, true
			}
		}
		return []string{
			fmt.Sprintf("%s = %s", resultTemp, valueExpr),
			fmt.Sprintf("break %s", goLabel),
		}, true
	}
	// Loop break — use Go's native break with label
	var lines []string
	if stmt.Value != nil {
		valLines, expr, goType, ok := g.compileExprLines(ctx, stmt.Value, "")
		if !ok {
			return nil, false
		}
		lines = append(lines, valLines...)
		convLines, valueRuntime, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("break value unsupported")
			return nil, false
		}
		lines = append(lines, convLines...)
		if ctx.loopBreakValueTemp != "" {
			lines = append(lines, fmt.Sprintf("%s = %s", ctx.loopBreakValueTemp, valueRuntime))
		}
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
	resultType := expected
	if resultType == "" {
		resultType = "runtime.Value"
	}
	loopLabelName := ctx.newTemp()
	valueTemp := ctx.newTemp()
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	bodyCtx.loopBreakValueTemp = valueTemp
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, "", "", false
	}
	lines := []string{
		fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", valueTemp),
	}
	forLine := "for {"
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for {", loopLabelName)
	}
	lines = append(lines, forLine)
	lines = append(lines, indentLines(bodyLines, 1)...)
	lines = append(lines, "}")
	retExpr := valueTemp
	if resultType != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueTemp, resultType)
		if !ok {
			ctx.setReason("loop expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		retExpr = converted
	}
	return lines, retExpr, resultType, true
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
	resultType := expected
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
	bodyCtx.breakpointGoLabels[label] = goLabel
	bodyCtx.breakpointResultTemps[label] = resultTemp

	// Compile the body block as statements + tail expression.
	// The result temp is always runtime.Value since break values are runtime.Value.
	stmts := expr.Body.Body
	var bodyLines []string
	for idx, stmt := range stmts {
		isLast := idx == len(stmts)-1
		if isLast {
			// Try to compile last statement as a value expression
			if tailExpr, ok := stmt.(ast.Expression); ok {
				tailLines, tailValue, tailType, ok := g.compileExprLines(bodyCtx, tailExpr, "")
				if ok {
					bodyLines = append(bodyLines, tailLines...)
					runtimeValue := tailValue
					if tailType != "runtime.Value" {
						convLines, converted, ok := g.runtimeValueLines(ctx, tailValue, tailType)
						if !ok {
							ctx.setReason("breakpoint type mismatch")
							return nil, "", "", false
						}
						bodyLines = append(bodyLines, convLines...)
						runtimeValue = converted
					}
					bodyLines = append(bodyLines, fmt.Sprintf("%s = %s", resultTemp, runtimeValue))
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
	lines := []string{
		fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
	}
	lines = append(lines, fmt.Sprintf("%s: switch { default: %s }", goLabel, strings.Join(bodyLines, "; ")))

	// Convert to expected type
	retExpr := resultTemp
	if resultType != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, resultType)
		if !ok {
			ctx.setReason("breakpoint type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		retExpr = converted
	}

	return lines, retExpr, resultType, true
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
	// Match patterns operate on runtime.Value; if the subject compiled to a
	// struct type (e.g. *Array from an array literal), convert it so pattern
	// matching can proceed normally.
	if subjectType != "runtime.Value" && g.typeCategory(subjectType) == "struct" {
		convLines, converted, ok := g.runtimeValueLines(ctx, subjectExpr, subjectType)
		if ok {
			subjectLines = append(subjectLines, convLines...)
			subjectExpr = converted
			subjectType = "runtime.Value"
		}
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
			continue
		}
		if g.typeMatches(resultType, clause.bodyType) {
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
			// any accepts all types without conversion
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
