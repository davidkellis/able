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
	switch e := expr.(type) {
	case *ast.AssignmentExpression:
		lines, valueExpr, valueType, ok := g.compileAssignment(ctx, e)
		if !ok {
			return nil, "", "", false
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
		valueExpr, valueType, ok := g.compileExpr(ctx, expr, expected)
		return nil, valueExpr, valueType, ok
	}
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
	if expr.ElseBody == nil {
		ctx.setReason("if expression requires else")
		return nil, "", "", false
	}
	condExpr, condType, ok := g.compileExpr(ctx, expr.IfCondition, "bool")
	if !ok {
		return nil, "", "", false
	}
	if condType != "bool" {
		ctx.setReason("if condition must be bool")
		return nil, "", "", false
	}
	bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(ctx.child(), expected, expr.IfBody)
	if !ok {
		return nil, "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = bodyType
	}
	if !g.typeMatches(resultType, bodyType) {
		ctx.setReason("if branch type mismatch")
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
		clauseCondExpr, clauseCondType, ok := g.compileExpr(ctx, clause.Condition, "bool")
		if !ok {
			return nil, "", "", false
		}
		if clauseCondType != "bool" {
			ctx.setReason("if condition must be bool")
			return nil, "", "", false
		}
		clauseLines, clauseExpr, clauseType, ok := g.compileTailExpression(ctx.child(), resultType, clause.Body)
		if !ok {
			return nil, "", "", false
		}
		if !g.typeMatches(resultType, clauseType) {
			ctx.setReason("if branch type mismatch")
			return nil, "", "", false
		}
		branches = append(branches, ifBranch{cond: clauseCondExpr, lines: clauseLines, expr: clauseExpr})
	}
	elseLines, elseExpr, elseType, ok := g.compileTailExpression(ctx.child(), resultType, expr.ElseBody)
	if !ok {
		return nil, "", "", false
	}
	if !g.typeMatches(resultType, elseType) {
		ctx.setReason("if branch type mismatch")
		return nil, "", "", false
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
		ctx.setReason("empty block requires void return")
		return nil, "", "", false
	}
	lines := make([]string, 0, len(block.Body)+1)
	for idx, stmt := range block.Body {
		isLast := idx == len(block.Body)-1
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			ctx.setReason("return not allowed in block expression")
			return nil, "", "", false
		}
		if isLast {
			expr, ok := stmt.(ast.Expression)
			if !ok || expr == nil {
				ctx.setReason("missing block return expression")
				return nil, "", "", false
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
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			ctx.setReason("return not allowed in statement block")
			return nil, false
		}
		stmtLines, ok := g.compileStatement(child, stmt)
		if !ok {
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
	condExpr, condType, ok := g.compileExpr(ctx, expr.IfCondition, "bool")
	if !ok {
		return nil, false
	}
	if condType != "bool" {
		ctx.setReason("if condition must be bool")
		return nil, false
	}
	bodyLines, ok := g.compileBlockStatement(ctx.child(), expr.IfBody)
	if !ok {
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
		clauseCondExpr, clauseCondType, ok := g.compileExpr(ctx, clause.Condition, "bool")
		if !ok {
			return nil, false
		}
		if clauseCondType != "bool" {
			ctx.setReason("if condition must be bool")
			return nil, false
		}
		clauseLines, ok := g.compileBlockStatement(ctx.child(), clause.Body)
		if !ok {
			return nil, false
		}
		lines = append(lines, fmt.Sprintf("} else if %s {", clauseCondExpr))
		lines = append(lines, indentLines(clauseLines, 1)...)
	}
	if expr.ElseBody != nil {
		elseLines, ok := g.compileBlockStatement(ctx.child(), expr.ElseBody)
		if !ok {
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
	condExpr, condType, ok := g.compileExpr(ctx, loop.Condition, "bool")
	if !ok {
		return nil, false
	}
	if condType != "bool" {
		ctx.setReason("while condition must be bool")
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
	if loop == nil || loop.Iterable == nil || loop.Body == nil {
		ctx.setReason("missing for loop")
		return nil, false
	}
	iterExpr, iterType, ok := g.compileExpr(ctx, loop.Iterable, "")
	if !ok {
		return nil, false
	}
	iterRuntime, ok := g.runtimeValueExpr(iterExpr, iterType)
	if !ok {
		ctx.setReason("for loop iterable unsupported")
		return nil, false
	}
	binding, ok := g.forLoopBinding(ctx, loop.Pattern)
	if !ok {
		return nil, false
	}
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bindLines := make([]string, 0, 1)
	if binding != nil && binding.name != "" {
		bodyCtx.locals[binding.name] = paramInfo{Name: binding.name, GoName: binding.goName, GoType: binding.goType}
	}
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, false
	}
	elementTemp := ctx.newTemp()
	if binding != nil && binding.name != "" {
		bindExpr := elementTemp
		if binding.goType != "runtime.Value" {
			converted, ok := g.expectRuntimeValueExpr(elementTemp, binding.goType)
			if !ok {
				ctx.setReason("for loop binding type mismatch")
				return nil, false
			}
			bindExpr = converted
		}
		bindLines = append(bindLines, fmt.Sprintf("%s := %s", binding.goName, bindExpr))
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
		fmt.Sprintf("var %s bool", brokeTemp),
		fmt.Sprintf("var %s bool", contTemp),
		fmt.Sprintf("%s, %s := __able_array_values(%s)", valuesTemp, okTemp, iterTemp),
		fmt.Sprintf("if %s {", okTemp),
		fmt.Sprintf("for _, %s := range %s {", elementTemp, valuesTemp),
		fmt.Sprintf("%s = false", brokeTemp),
		fmt.Sprintf("%s = false", contTemp),
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
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
		fmt.Sprintf("func() { %s }()", strings.Join(innerLines, "; ")),
		fmt.Sprintf("if %s { break }", brokeTemp),
		fmt.Sprintf("if %s { continue }", contTemp),
		"}",
		"}",
	}
	return lines, true
}

type forLoopBinding struct {
	name   string
	goName string
	goType string
}

func (g *generator) forLoopBinding(ctx *compileContext, pattern ast.Pattern) (*forLoopBinding, bool) {
	if pattern == nil {
		ctx.setReason("missing for loop pattern")
		return nil, false
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return nil, true
	case *ast.Identifier:
		if p == nil || p.Name == "" {
			ctx.setReason("missing for loop identifier")
			return nil, false
		}
		return &forLoopBinding{name: p.Name, goName: sanitizeIdent(p.Name), goType: "runtime.Value"}, true
	case *ast.TypedPattern:
		if p == nil || p.Pattern == nil || p.TypeAnnotation == nil {
			ctx.setReason("unsupported for loop pattern")
			return nil, false
		}
		ident, ok := p.Pattern.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			ctx.setReason("unsupported for loop pattern")
			return nil, false
		}
		mapped, ok := g.mapTypeExpression(p.TypeAnnotation)
		if !ok || mapped == "" || mapped == "struct{}" {
			ctx.setReason("unsupported for loop pattern type")
			return nil, false
		}
		return &forLoopBinding{name: ident.Name, goName: sanitizeIdent(ident.Name), goType: mapped}, true
	default:
		ctx.setReason("unsupported for loop pattern")
		return nil, false
	}
}

func (g *generator) compileBreakStatement(ctx *compileContext, stmt *ast.BreakStatement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing break")
		return nil, false
	}
	if ctx.loopDepth <= 0 {
		ctx.setReason("break used outside loop")
		return nil, false
	}
	if stmt.Label != nil {
		ctx.setReason("labeled break unsupported")
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
			guardValue, guardType, ok := g.compileExpr(clauseCtx, clause.Guard, "bool")
			if !ok {
				return "", "", false
			}
			if guardType != "bool" {
				ctx.setReason("match guard must be bool")
				return "", "", false
			}
			guardExpr = guardValue
		}
		bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(clauseCtx, resultType, clause.Body)
		if !ok {
			return "", "", false
		}
		if resultType == "" {
			resultType = bodyType
		} else if !g.typeMatches(resultType, bodyType) {
			ctx.setReason("match clause type mismatch")
			return "", "", false
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
	if resultType == "" {
		resultType = "runtime.Value"
	}
	matchedTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
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
	lines = append(lines, fmt.Sprintf("if !%s { panic(fmt.Errorf(\"Non-exhaustive match\")) }", matchedTemp))
	lines = append(lines, fmt.Sprintf("return %s", resultTemp))
	expr := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return expr, resultType, true
}
