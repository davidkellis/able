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
	lines := make([]string, 0, len(block.Body))
	for idx, stmt := range block.Body {
		isLast := idx == len(block.Body)-1
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			ctx.setReason("return not allowed in block expression")
			return nil, "", "", false
		}
		expr, ok := stmt.(ast.Expression)
		if !ok || expr == nil {
			ctx.setReason("unsupported block statement")
			return nil, "", "", false
		}
		if isLast {
			returnLines, returnExpr, returnType, ok := g.compileTailExpression(child, expected, expr)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, returnLines...)
			return lines, returnExpr, returnType, true
		}
		stmtLines, ok := g.compileStatement(child, expr)
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
