package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

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
			converted, ok := g.runtimeValueExpr(valueExpr, valueType)
			if !ok {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
			return lines, converted, "runtime.Value", true
		}
		if expected != "" && valueType == "runtime.Value" && expected != "runtime.Value" {
			converted, ok := g.expectRuntimeValueExpr(valueExpr, expected)
			if !ok {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
			return lines, converted, expected, true
		}
		return lines, valueExpr, valueType, ok
	case *ast.IfExpression:
		return g.compileIfExpression(ctx, e, expected)
	case *ast.BlockExpression:
		return g.compileBlockExpression(ctx, e, expected)
	default:
		compileExpected := expected
		if runtimeExpected {
			compileExpected = ""
		}
		valueExpr, valueType, ok := g.compileExpr(ctx, expr, compileExpected)
		if !ok {
			return nil, "", "", false
		}
		if runtimeExpected && valueType != "runtime.Value" {
			converted, ok := g.runtimeValueExpr(valueExpr, valueType)
			if !ok {
				ctx.setReason("expression type mismatch")
				return nil, "", "", false
			}
			return nil, converted, "runtime.Value", true
		}
		return nil, valueExpr, valueType, true
	}
}

func (g *generator) compileCondition(ctx *compileContext, expr ast.Expression) (string, bool) {
	if expr == nil {
		ctx.setReason("missing condition")
		return "", false
	}
	condExpr, condType, ok := g.compileExpr(ctx, expr, "")
	if !ok {
		return "", false
	}
	if condType == "bool" {
		return condExpr, true
	}
	condRuntime := condExpr
	if condType != "runtime.Value" {
		converted, ok := g.runtimeValueExpr(condExpr, condType)
		if !ok {
			ctx.setReason("condition unsupported")
			return "", false
		}
		condRuntime = converted
	}
	return fmt.Sprintf("__able_truthy(%s)", condRuntime), true
}

func (g *generator) coerceIfBranch(ctx *compileContext, resultType string, expr string, exprType string) (string, bool) {
	if resultType == "" || exprType == "" {
		ctx.setReason("if branch type mismatch")
		return "", false
	}
	if resultType == exprType {
		return expr, true
	}
	if resultType == "runtime.Value" && exprType != "runtime.Value" {
		converted, ok := g.runtimeValueExpr(expr, exprType)
		if !ok {
			ctx.setReason("if branch type mismatch")
			return "", false
		}
		return converted, true
	}
	if resultType != "runtime.Value" && exprType == "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(expr, resultType)
		if !ok {
			ctx.setReason("if branch type mismatch")
			return "", false
		}
		return converted, true
	}
	ctx.setReason("if branch type mismatch")
	return "", false
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
	condExpr, ok := g.compileCondition(ctx, expr.IfCondition)
	if !ok {
		return nil, "", "", false
	}
	bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(ctx.child(), expected, expr.IfBody)
	if !ok {
		return nil, "", "", false
	}
	resultType := expected
	if resultType == "" {
		if expr.ElseBody == nil {
			resultType = "runtime.Value"
		} else {
			resultType = bodyType
		}
	}
	bodyExpr, ok = g.coerceIfBranch(ctx, resultType, bodyExpr, bodyType)
	if !ok {
		return nil, "", "", false
	}
	type ifBranch struct {
		cond  string
		lines []string
		expr  string
	}
	branches := []ifBranch{{cond: condExpr, lines: bodyLines, expr: bodyExpr}}
	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			continue
		}
		if clause.Condition == nil || clause.Body == nil {
			ctx.setReason("incomplete else-if clause")
			return nil, "", "", false
		}
		clauseCondExpr, ok := g.compileCondition(ctx, clause.Condition)
		if !ok {
			return nil, "", "", false
		}
		clauseLines, clauseExpr, clauseType, ok := g.compileTailExpression(ctx.child(), resultType, clause.Body)
		if !ok {
			return nil, "", "", false
		}
		clauseExpr, ok = g.coerceIfBranch(ctx, resultType, clauseExpr, clauseType)
		if !ok {
			return nil, "", "", false
		}
		branches = append(branches, ifBranch{cond: clauseCondExpr, lines: clauseLines, expr: clauseExpr})
	}
	var elseLines []string
	elseExpr := ""
	elseType := ""
	if expr.ElseBody != nil {
		elseLines, elseExpr, elseType, ok = g.compileTailExpression(ctx.child(), resultType, expr.ElseBody)
		if !ok {
			return nil, "", "", false
		}
		elseExpr, ok = g.coerceIfBranch(ctx, resultType, elseExpr, elseType)
		if !ok {
			return nil, "", "", false
		}
	} else {
		if resultType != "runtime.Value" && !g.isVoidType(resultType) {
			ctx.setReason("if expression requires else")
			return nil, "", "", false
		}
		elseExpr = safeNilReturnExpr(resultType)
		elseType = resultType
	}
	temp := ctx.newTemp()
	lines := []string{fmt.Sprintf("var %s %s", temp, resultType)}
	lines = append(lines, fmt.Sprintf("if %s {", branches[0].cond))
	lines = append(lines, indentLines(branches[0].lines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", temp, branches[0].expr))
	for idx := 1; idx < len(branches); idx++ {
		branch := branches[idx]
		lines = append(lines, fmt.Sprintf("} else if %s {", branch.cond))
		lines = append(lines, indentLines(branch.lines, 1)...)
		lines = append(lines, fmt.Sprintf("\t%s = %s", temp, branch.expr))
	}
	lines = append(lines, "} else {")
	lines = append(lines, indentLines(elseLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", temp, elseExpr))
	lines = append(lines, "}")
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
		ctx.setReason("empty block requires void return")
		return nil, "", "", false
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
				lines = append(lines, fmt.Sprintf("return %s", returnExpr))
				blockExpr := fmt.Sprintf("func() %s { %s }()", returnType, strings.Join(lines, "; "))
				return nil, blockExpr, returnType, true
			}
			if rethrowStmt, ok := stmt.(*ast.RethrowStatement); ok {
				stmtLines, ok := g.compileRethrowStatement(child, rethrowStmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
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
				lines = append(lines, fmt.Sprintf("return %s", returnExpr))
				blockExpr := fmt.Sprintf("func() %s { %s }()", returnType, strings.Join(lines, "; "))
				return nil, blockExpr, returnType, true
			}
			expr, ok := stmt.(ast.Expression)
			if !ok || expr == nil {
				if loop, ok := stmt.(*ast.ForLoop); ok && (expected == "" || expected == "runtime.Value") {
					loopLines, loopResult, ok := g.compileForLoopInternal(child, loop, true)
					if !ok {
						return nil, "", "", false
					}
					lines = append(lines, loopLines...)
					returnType := "runtime.Value"
					returnExpr := loopResult
					if returnExpr == "" {
						returnExpr = "runtime.VoidValue{}"
					}
					lines = append(lines, fmt.Sprintf("return %s", returnExpr))
					blockExpr := fmt.Sprintf("func() %s { %s }()", returnType, strings.Join(lines, "; "))
					return nil, blockExpr, returnType, true
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
				lines = append(lines, fmt.Sprintf("return %s", returnExpr))
				blockExpr := fmt.Sprintf("func() %s { %s }()", returnType, strings.Join(lines, "; "))
				return nil, blockExpr, returnType, true
			}
			returnLines, returnExpr, returnType, ok := g.compileTailExpression(child, expected, expr)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, returnLines...)
			lines = append(lines, fmt.Sprintf("return %s", returnExpr))
			blockExpr := fmt.Sprintf("func() %s { %s }()", returnType, strings.Join(lines, "; "))
			return nil, blockExpr, returnType, true
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
	condExpr, ok := g.compileCondition(ctx, expr.IfCondition)
	if !ok {
		return nil, false
	}
	bodyCtx := ctx.child()
	bodyLines, ok := g.compileBlockStatement(bodyCtx, expr.IfBody)
	if !ok {
		ctx.setReason(bodyCtx.reason)
		return nil, false
	}
	lines := []string{fmt.Sprintf("if %s {", condExpr)}
	lines = append(lines, indentLines(bodyLines, 1)...)
	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			continue
		}
		if clause.Condition == nil || clause.Body == nil {
			ctx.setReason("incomplete else-if clause")
			return nil, false
		}
		clauseCondExpr, ok := g.compileCondition(ctx, clause.Condition)
		if !ok {
			return nil, false
		}
		clauseCtx := ctx.child()
		clauseLines, ok := g.compileBlockStatement(clauseCtx, clause.Body)
		if !ok {
			ctx.setReason(clauseCtx.reason)
			return nil, false
		}
		lines = append(lines, fmt.Sprintf("} else if %s {", clauseCondExpr))
		lines = append(lines, indentLines(clauseLines, 1)...)
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
	return lines, true
}

func (g *generator) compileWhileLoop(ctx *compileContext, loop *ast.WhileLoop) ([]string, bool) {
	if loop == nil || loop.Condition == nil || loop.Body == nil {
		ctx.setReason("missing while loop")
		return nil, false
	}
	condExpr, ok := g.compileCondition(ctx, loop.Condition)
	if !ok {
		return nil, false
	}
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, false
	}
	brokeTemp := ctx.newTemp()
	contTemp := ctx.newTemp()
	innerLines := []string{
		fmt.Sprintf("defer func() { if recovered := recover(); recovered != nil { switch recovered.(type) { case __able_break: %s = true; case __able_continue_signal: %s = true; default: panic(recovered) } } }()", brokeTemp, contTemp),
	}
	innerLines = append(innerLines, bodyLines...)
	loopLines := []string{
		fmt.Sprintf("var %s bool", brokeTemp),
		fmt.Sprintf("var %s bool", contTemp),
		"for {",
		fmt.Sprintf("if !%s { break }", condExpr),
		fmt.Sprintf("%s = false", brokeTemp),
		fmt.Sprintf("%s = false", contTemp),
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
		fmt.Sprintf("if %s { break }", brokeTemp),
		fmt.Sprintf("if %s { continue }", contTemp),
		"}",
	}
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
	iterExpr, iterType, ok := g.compileExpr(ctx, loop.Iterable, "")
	if !ok {
		return nil, "", false
	}
	iterRuntime, ok := g.runtimeValueExpr(iterExpr, iterType)
	if !ok {
		ctx.setReason("for loop iterable unsupported")
		return nil, "", false
	}
	elementTemp := ctx.newTemp()
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	newNames := map[string]struct{}{}
	collectPatternBindingNames(loop.Pattern, newNames)
	mode := patternBindingMode{declare: true, newNames: newNames}
	cond, ok := g.compileMatchPatternCondition(bodyCtx, loop.Pattern, elementTemp, "runtime.Value")
	if !ok {
		return nil, "", false
	}
	bindLines, ok := g.compileAssignmentPatternBindings(bodyCtx, loop.Pattern, elementTemp, "runtime.Value", mode)
	if !ok {
		return nil, "", false
	}
	mismatchTemp := ctx.newTemp()
	resultTemp := ""
	if withResult {
		resultTemp = ctx.newTemp()
	}
	if cond != "true" {
		mismatchLine := fmt.Sprintf("%s = true; return", mismatchTemp)
		if withResult {
			mismatchLine = fmt.Sprintf("%s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"}; %s", resultTemp, mismatchLine)
		}
		bindLines = append([]string{fmt.Sprintf("if !(%s) { %s }", cond, mismatchLine)}, bindLines...)
	}
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, "", false
	}
	brokeTemp := ctx.newTemp()
	contTemp := ctx.newTemp()
	innerLines := []string{
		fmt.Sprintf("defer func() { if recovered := recover(); recovered != nil { switch recovered.(type) { case __able_break: %s = true; case __able_continue_signal: %s = true; default: panic(recovered) } } }()", brokeTemp, contTemp),
	}
	innerLines = append(innerLines, bindLines...)
	innerLines = append(innerLines, bodyLines...)
	iterTemp := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	iteratorTemp := ctx.newTemp()
	doneTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", iterTemp, iterRuntime),
	}
	if withResult {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.VoidValue{}", resultTemp))
	}
	lines = append(lines,
		fmt.Sprintf("var %s bool", mismatchTemp),
		fmt.Sprintf("var %s bool", brokeTemp),
		fmt.Sprintf("var %s bool", contTemp),
		fmt.Sprintf("%s, %s := __able_array_values(%s)", valuesTemp, okTemp, iterTemp),
		fmt.Sprintf("if %s {", okTemp),
		fmt.Sprintf("for _, %s := range %s {", elementTemp, valuesTemp),
		fmt.Sprintf("%s = false", brokeTemp),
		fmt.Sprintf("%s = false", contTemp),
		fmt.Sprintf("%s = false", mismatchTemp),
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
		fmt.Sprintf("if %s { break }", mismatchTemp),
		fmt.Sprintf("if %s { break }", brokeTemp),
		fmt.Sprintf("if %s { continue }", contTemp),
		"}",
		"} else {",
		fmt.Sprintf("%s := __able_resolve_iterator(%s)", iteratorTemp, iterTemp),
		fmt.Sprintf("defer %s.Close()", iteratorTemp),
		"for {",
		fmt.Sprintf("%s, %s, %s := %s.Next()", elementTemp, doneTemp, errTemp, iteratorTemp),
		fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		fmt.Sprintf("if %s { break }", doneTemp),
		fmt.Sprintf("%s = false", brokeTemp),
		fmt.Sprintf("%s = false", contTemp),
		fmt.Sprintf("%s = false", mismatchTemp),
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
		fmt.Sprintf("if %s { break }", mismatchTemp),
		fmt.Sprintf("if %s { break }", brokeTemp),
		fmt.Sprintf("if %s { continue }", contTemp),
		"}",
		"}",
	)
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
	valueExpr := "runtime.NilValue{}"
	if stmt.Value != nil {
		expr, goType, ok := g.compileExpr(ctx, stmt.Value, "")
		if !ok {
			return nil, false
		}
		valueRuntime, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("break value unsupported")
			return nil, false
		}
		valueExpr = valueRuntime
	}
	if label != "" {
		return []string{fmt.Sprintf("__able_break_label(%q, %s)", label, valueExpr)}, true
	}
	return []string{fmt.Sprintf("__able_break_value(%s)", valueExpr)}, true
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
	return []string{"__able_continue()"}, true
}

func (g *generator) compileLoopExpression(ctx *compileContext, loop *ast.LoopExpression, expected string) (string, string, bool) {
	if loop == nil || loop.Body == nil {
		ctx.setReason("missing loop expression")
		return "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = "runtime.Value"
	}
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return "", "", false
	}
	brokeTemp := ctx.newTemp()
	contTemp := ctx.newTemp()
	valueTemp := ctx.newTemp()
	innerLines := []string{
		fmt.Sprintf("defer func() { if recovered := recover(); recovered != nil { switch sig := recovered.(type) { case __able_break: %s = true; %s = sig.value; case __able_continue_signal: %s = true; default: panic(recovered) } } }()", brokeTemp, valueTemp, contTemp),
	}
	innerLines = append(innerLines, bodyLines...)
	retExpr := valueTemp
	if resultType != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(valueTemp, resultType)
		if !ok {
			ctx.setReason("loop expression type mismatch")
			return "", "", false
		}
		retExpr = converted
	}
	loopLines := []string{
		fmt.Sprintf("var %s bool", brokeTemp),
		fmt.Sprintf("var %s bool", contTemp),
		fmt.Sprintf("var %s runtime.Value", valueTemp),
		"for {",
		fmt.Sprintf("%s = false", brokeTemp),
		fmt.Sprintf("%s = false", contTemp),
		fmt.Sprintf("%s = runtime.NilValue{}", valueTemp),
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
		fmt.Sprintf("if %s { return %s }", brokeTemp, retExpr),
		fmt.Sprintf("if %s { continue }", contTemp),
		"}",
	}
	expr := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(loopLines, "; "))
	return expr, resultType, true
}

func (g *generator) compileBreakpointExpression(ctx *compileContext, expr *ast.BreakpointExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Body == nil {
		ctx.setReason("missing breakpoint expression")
		return "", "", false
	}
	if expr.Label == nil || expr.Label.Name == "" {
		ctx.setReason("breakpoint requires label")
		return "", "", false
	}
	label := expr.Label.Name
	bodyCtx := ctx.child()
	bodyCtx.pushBreakpoint(label)
	bodyLines, bodyExpr, bodyType, ok := g.compileBlockExpression(bodyCtx, expr.Body, expected)
	bodyCtx.popBreakpoint(label)
	if !ok {
		return "", "", false
	}
	if len(bodyLines) > 0 {
		ctx.setReason("breakpoint body produced statements")
		return "", "", false
	}

	resultType := expected
	if resultType == "" {
		if bodyType != "" {
			resultType = bodyType
		} else {
			resultType = "runtime.Value"
		}
	}

	bodyResultExpr := bodyExpr
	switch {
	case bodyType == resultType:
	case bodyType == "runtime.Value" && resultType != "runtime.Value":
		converted, ok := g.expectRuntimeValueExpr(bodyExpr, resultType)
		if !ok {
			ctx.setReason("breakpoint type mismatch")
			return "", "", false
		}
		bodyResultExpr = converted
	case resultType == "runtime.Value" && bodyType != "runtime.Value":
		converted, ok := g.runtimeValueExpr(bodyExpr, bodyType)
		if !ok {
			ctx.setReason("breakpoint type mismatch")
			return "", "", false
		}
		bodyResultExpr = converted
	default:
		ctx.setReason("breakpoint type mismatch")
		return "", "", false
	}

	breakValueTemp := ctx.newTemp()
	brokeTemp := ctx.newTemp()
	contTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	breakResultExpr := breakValueTemp
	if resultType != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(breakValueTemp, resultType)
		if !ok {
			ctx.setReason("breakpoint type mismatch")
			return "", "", false
		}
		breakResultExpr = converted
	}

	innerLines := []string{
		fmt.Sprintf("defer func() { if recovered := recover(); recovered != nil { switch sig := recovered.(type) { case __able_break_label_signal: if sig.label == %q { %s = true; %s = sig.value } else { panic(recovered) }; case __able_continue_label_signal: if sig.label == %q { %s = true } else { panic(recovered) }; default: panic(recovered) } } }()", label, brokeTemp, breakValueTemp, label, contTemp),
		fmt.Sprintf("%s = %s", resultTemp, bodyResultExpr),
	}

	lines := []string{
		fmt.Sprintf("var %s bool", brokeTemp),
		fmt.Sprintf("var %s bool", contTemp),
		fmt.Sprintf("var %s runtime.Value", breakValueTemp),
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		"for {",
		fmt.Sprintf("%s = false", brokeTemp),
		fmt.Sprintf("%s = false", contTemp),
		fmt.Sprintf("%s = runtime.NilValue{}", breakValueTemp),
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
		fmt.Sprintf("if %s { return %s }", brokeTemp, breakResultExpr),
		fmt.Sprintf("if %s { continue }", contTemp),
		fmt.Sprintf("return %s", resultTemp),
		"}",
	}

	exprValue := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return exprValue, resultType, true
}

func (g *generator) compileMatchExpression(ctx *compileContext, match *ast.MatchExpression, expected string) (string, string, bool) {
	if match == nil || match.Subject == nil {
		ctx.setReason("missing match expression")
		return "", "", false
	}
	subjectExpr, subjectType, ok := g.compileExpr(ctx, match.Subject, "")
	if !ok {
		return "", "", false
	}
	subjectTemp := ctx.newTemp()
	resultType := expected
	explicitExpected := expected != ""
	inferredType := ""
	mismatch := false
	type matchClause struct {
		cond      string
		bindLines []string
		guardExpr string
		bodyLines []string
		bodyExpr  string
		bodyType  string
	}
	clauses := make([]matchClause, 0, len(match.Clauses))
	for _, clause := range match.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, subjectTemp, subjectType)
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
		clauseExpected := resultType
		if !explicitExpected {
			clauseExpected = ""
		}
		bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(clauseCtx, clauseExpected, clause.Body)
		if !ok {
			return "", "", false
		}
		if explicitExpected {
			if !g.typeMatches(resultType, bodyType) {
				ctx.setReason("match clause type mismatch")
				return "", "", false
			}
		} else {
			if inferredType == "" {
				inferredType = bodyType
			} else if bodyType != inferredType {
				mismatch = true
			}
		}
		clauses = append(clauses, matchClause{
			cond:      cond,
			bindLines: bindLines,
			guardExpr: guardExpr,
			bodyLines: bodyLines,
			bodyExpr:  bodyExpr,
			bodyType:  bodyType,
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
		switch {
		case resultType == "runtime.Value" && clause.bodyType != "runtime.Value":
			converted, ok := g.runtimeValueExpr(clause.bodyExpr, clause.bodyType)
			if !ok {
				ctx.setReason("match clause type mismatch")
				return "", "", false
			}
			clause.bodyExpr = converted
			clause.bodyType = resultType
		case clause.bodyType == "runtime.Value" && resultType != "runtime.Value":
			converted, ok := g.expectRuntimeValueExpr(clause.bodyExpr, resultType)
			if !ok {
				ctx.setReason("match clause type mismatch")
				return "", "", false
			}
			clause.bodyExpr = converted
			clause.bodyType = resultType
		default:
			ctx.setReason("match clause type mismatch")
			return "", "", false
		}
	}
	matchedTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	matchNode := g.diagNodeName(match, "*ast.MatchExpression", "match")
	lines := []string{
		fmt.Sprintf("%s := %s", subjectTemp, subjectExpr),
		fmt.Sprintf("%s := false", matchedTemp),
		fmt.Sprintf("var %s %s", resultTemp, resultType),
	}
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
	lines = append(lines, fmt.Sprintf("if !%s { bridge.RaiseRuntimeErrorWithContext(__able_runtime, %s, fmt.Errorf(\"Non-exhaustive match\")) }", matchedTemp, matchNode))
	lines = append(lines, fmt.Sprintf("return %s", resultTemp))
	expr := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return expr, resultType, true
}

func (g *generator) compileMatchStatement(ctx *compileContext, match *ast.MatchExpression) ([]string, bool) {
	if match == nil || match.Subject == nil {
		ctx.setReason("missing match expression")
		return nil, false
	}
	subjectExpr, subjectType, ok := g.compileExpr(ctx, match.Subject, "")
	if !ok {
		return nil, false
	}
	subjectTemp := ctx.newTemp()
	matchedTemp := ctx.newTemp()
	matchNode := g.diagNodeName(match, "*ast.MatchExpression", "match")
	lines := []string{
		fmt.Sprintf("%s := %s", subjectTemp, subjectExpr),
		fmt.Sprintf("%s := false", matchedTemp),
	}
	for _, clause := range match.Clauses {
		if clause == nil {
			continue
		}
		clauseCtx := ctx.child()
		cond, bindLines, ok := g.compileMatchPattern(clauseCtx, clause.Pattern, subjectTemp, subjectType)
		if !ok {
			ctx.setReason(clauseCtx.reason)
			return nil, false
		}
		guardExpr := ""
		if clause.Guard != nil {
			guardExpr, ok = g.compileCondition(clauseCtx, clause.Guard)
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
			branchLines = append(branchLines, fmt.Sprintf("if %s {", guardExpr))
			branchLines = append(branchLines, indentLines(bodyLines, 1)...)
			branchLines = append(branchLines, fmt.Sprintf("\t%s = true", matchedTemp))
			branchLines = append(branchLines, "}")
		} else {
			branchLines = append(branchLines, bodyLines...)
			branchLines = append(branchLines, fmt.Sprintf("%s = true", matchedTemp))
		}
		lines = append(lines, fmt.Sprintf("if !%s && %s {", matchedTemp, cond))
		lines = append(lines, indentLines(branchLines, 1)...)
		lines = append(lines, "}")
	}
	lines = append(lines, fmt.Sprintf("if !%s { bridge.RaiseRuntimeErrorWithContext(__able_runtime, %s, fmt.Errorf(\"Non-exhaustive match\")) }", matchedTemp, matchNode))
	return lines, true
}
