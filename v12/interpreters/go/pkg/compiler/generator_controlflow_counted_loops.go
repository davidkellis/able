package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileCountedLoopExpression(ctx *compileContext, loop *ast.LoopExpression, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || loop == nil || loop.Body == nil {
		return nil, "", "", false
	}
	stmts := loop.Body.Body
	if len(stmts) < 2 {
		return nil, "", "", false
	}
	varName, boundExpr, ok := g.matchCountedLoopGuard(stmts[0])
	if !ok {
		return nil, "", "", false
	}
	if !g.matchCountedLoopIncrement(stmts[len(stmts)-1], varName) {
		return nil, "", "", false
	}
	if countedLoopBodyAssignsName(stmts[1:len(stmts)-1], varName) {
		return nil, "", "", false
	}
	binding, ok := ctx.lookup(varName)
	if !ok || binding.GoName == "" || !g.isIntegerType(binding.GoType) {
		return nil, "", "", false
	}
	boundLines, boundGoExpr, boundType, ok := g.compileExprLines(ctx, boundExpr, "")
	if !ok || len(boundLines) != 0 || boundGoExpr == "" || boundType != binding.GoType {
		return nil, "", "", false
	}

	resultType := g.inferLoopExpressionResultType(ctx, loop, expected)
	if resultType == "" {
		resultType = "runtime.Value"
	}
	loopLabelName := ctx.newTemp()
	valueTemp := ctx.newTemp()
	coreBlock := ast.NewBlockExpression(stmts[1 : len(stmts)-1])
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	bodyCtx.loopBreakValueTemp = valueTemp
	bodyCtx.loopBreakValueType = resultType
	bodyCtx.loopBreakProbe = nil
	g.seedCountedLoopIntegerFact(bodyCtx, binding, boundExpr)
	bodyLines, ok := g.compileBlockStatement(bodyCtx, coreBlock)
	if !ok {
		return nil, "", "", false
	}
	zeroExpr, ok := g.zeroValueExpr(resultType)
	if !ok {
		ctx.setReason("loop expression type mismatch")
		return nil, "", "", false
	}

	lines := []string{fmt.Sprintf("var %s %s = %s", valueTemp, resultType, zeroExpr)}
	forLine := fmt.Sprintf("for %s < %s {", binding.GoName, boundGoExpr)
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for %s < %s {", loopLabelName, binding.GoName, boundGoExpr)
	}
	lines = append(lines, forLine)
	lines = append(lines, indentLines(bodyLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s++", binding.GoName))
	lines = append(lines, "}")
	return lines, valueTemp, resultType, true
}

func (g *generator) compileCountedLoopStatement(ctx *compileContext, loop *ast.LoopExpression) ([]string, bool) {
	if g == nil || ctx == nil || loop == nil || loop.Body == nil {
		return nil, false
	}
	stmts := loop.Body.Body
	if len(stmts) < 2 {
		return nil, false
	}
	varName, boundExpr, ok := g.matchCountedLoopGuard(stmts[0])
	if !ok {
		return nil, false
	}
	if !g.matchCountedLoopIncrement(stmts[len(stmts)-1], varName) {
		return nil, false
	}
	coreStmts := stmts[1 : len(stmts)-1]
	if countedLoopBodyAssignsName(coreStmts, varName) || countedLoopBodyContainsValueBreak(coreStmts) {
		return nil, false
	}
	binding, ok := ctx.lookup(varName)
	if !ok || binding.GoName == "" || !g.isIntegerType(binding.GoType) {
		return nil, false
	}
	boundLines, boundGoExpr, boundType, ok := g.compileExprLines(ctx, boundExpr, "")
	if !ok || len(boundLines) != 0 || boundGoExpr == "" || boundType != binding.GoType {
		return nil, false
	}

	loopLabelName := ctx.newTemp()
	coreBlock := ast.NewBlockExpression(coreStmts)
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	bodyCtx.loopBreakValueTemp = ""
	bodyCtx.loopBreakValueType = "runtime.Value"
	bodyCtx.loopBreakProbe = nil
	g.seedCountedLoopIntegerFact(bodyCtx, binding, boundExpr)
	bodyLines, ok := g.compileBlockStatement(bodyCtx, coreBlock)
	if !ok {
		return nil, false
	}

	forLine := fmt.Sprintf("for %s < %s {", binding.GoName, boundGoExpr)
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for %s < %s {", loopLabelName, binding.GoName, boundGoExpr)
	}
	lines := []string{forLine}
	lines = append(lines, indentLines(bodyLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s++", binding.GoName))
	lines = append(lines, "}")
	return lines, true
}

func (g *generator) seedCountedLoopIntegerFact(ctx *compileContext, binding paramInfo, boundExpr ast.Expression) {
	if g == nil || ctx == nil || binding.GoName == "" || !g.isIntegerType(binding.GoType) {
		return
	}
	fact := integerFact{}
	if g.isUnsignedIntegerType(binding.GoType) {
		fact.NonNegative = true
	}
	boundFact, ok := g.exprIntegerFact(ctx, boundExpr)
	if ok && boundFact.NonNegative {
		fact.NonNegative = true
	}
	if ok && boundFact.HasMax {
		fact.HasMax = true
		if boundFact.MaxInclusive > 0 {
			fact.MaxInclusive = boundFact.MaxInclusive - 1
		} else {
			fact.MaxInclusive = 0
		}
	}
	if fact.hasUsefulFact() {
		ctx.setIntegerFact(binding.GoName, fact)
	}
}

func countedLoopBodyContainsValueBreak(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		if countedLoopStatementContainsValueBreak(stmt) {
			return true
		}
	}
	return false
}

func countedLoopStatementContainsValueBreak(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.FunctionDefinition:
		return false
	case *ast.BreakStatement:
		return s.Value != nil
	case *ast.AssignmentExpression:
		return countedLoopExpressionContainsValueBreak(s.Right)
	case *ast.BlockExpression:
		return countedLoopBodyContainsValueBreak(s.Body)
	case *ast.IfExpression:
		if countedLoopExpressionContainsValueBreak(s.IfCondition) || countedLoopBodyContainsValueBreak(s.IfBody.Body) {
			return true
		}
		for _, clause := range s.ElseIfClauses {
			if clause == nil {
				continue
			}
			if countedLoopExpressionContainsValueBreak(clause.Condition) || countedLoopBodyContainsValueBreak(clause.Body.Body) {
				return true
			}
		}
		return s.ElseBody != nil && countedLoopBodyContainsValueBreak(s.ElseBody.Body)
	case *ast.WhileLoop:
		return countedLoopExpressionContainsValueBreak(s.Condition) || countedLoopBodyContainsValueBreak(s.Body.Body)
	case *ast.ForLoop:
		return countedLoopExpressionContainsValueBreak(s.Iterable) || countedLoopBodyContainsValueBreak(s.Body.Body)
	case *ast.LoopExpression:
		return countedLoopBodyContainsValueBreak(s.Body.Body)
	case *ast.ReturnStatement:
		return countedLoopExpressionContainsValueBreak(s.Argument)
	case *ast.RaiseStatement:
		return countedLoopExpressionContainsValueBreak(s.Expression)
	case *ast.YieldStatement:
		return countedLoopExpressionContainsValueBreak(s.Expression)
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			return countedLoopExpressionContainsValueBreak(expr)
		}
		return false
	}
}

func countedLoopExpressionContainsValueBreak(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.AssignmentExpression:
		return countedLoopStatementContainsValueBreak(e)
	case *ast.BinaryExpression:
		return countedLoopExpressionContainsValueBreak(e.Left) || countedLoopExpressionContainsValueBreak(e.Right)
	case *ast.UnaryExpression:
		return countedLoopExpressionContainsValueBreak(e.Operand)
	case *ast.TypeCastExpression:
		return countedLoopExpressionContainsValueBreak(e.Expression)
	case *ast.FunctionCall:
		if countedLoopExpressionContainsValueBreak(e.Callee) {
			return true
		}
		for _, arg := range e.Arguments {
			if countedLoopExpressionContainsValueBreak(arg) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		return countedLoopExpressionContainsValueBreak(e.Object) || countedLoopExpressionContainsValueBreak(e.Member)
	case *ast.IndexExpression:
		return countedLoopExpressionContainsValueBreak(e.Object) || countedLoopExpressionContainsValueBreak(e.Index)
	case *ast.BlockExpression:
		return countedLoopBodyContainsValueBreak(e.Body)
	case *ast.IteratorLiteral:
		return countedLoopBodyContainsValueBreak(e.Body)
	case *ast.IfExpression:
		return countedLoopStatementContainsValueBreak(e)
	case *ast.MatchExpression:
		if countedLoopExpressionContainsValueBreak(e.Subject) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if countedLoopExpressionContainsValueBreak(clause.Guard) || countedLoopExpressionContainsValueBreak(clause.Body) {
				return true
			}
		}
		return false
	case *ast.RangeExpression:
		return countedLoopExpressionContainsValueBreak(e.Start) || countedLoopExpressionContainsValueBreak(e.End)
	case *ast.LambdaExpression:
		return countedLoopExpressionContainsValueBreak(e.Body)
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if countedLoopExpressionContainsValueBreak(part) {
				return true
			}
		}
		return false
	case *ast.SpawnExpression:
		return countedLoopExpressionContainsValueBreak(e.Expression)
	case *ast.AwaitExpression:
		return countedLoopExpressionContainsValueBreak(e.Expression)
	case *ast.PropagationExpression:
		return countedLoopExpressionContainsValueBreak(e.Expression)
	case *ast.OrElseExpression:
		if countedLoopExpressionContainsValueBreak(e.Expression) {
			return true
		}
		return e.Handler != nil && countedLoopBodyContainsValueBreak(e.Handler.Body)
	case *ast.BreakpointExpression:
		return e.Body != nil && countedLoopBodyContainsValueBreak(e.Body.Body)
	case *ast.EnsureExpression:
		if countedLoopExpressionContainsValueBreak(e.TryExpression) {
			return true
		}
		return e.EnsureBlock != nil && countedLoopBodyContainsValueBreak(e.EnsureBlock.Body)
	default:
		return false
	}
}

func (g *generator) matchCountedLoopGuardFromStatements(stmts []ast.Statement) (string, ast.Expression, bool) {
	if len(stmts) == 0 {
		return "", nil, false
	}
	return g.matchCountedLoopGuard(stmts[0])
}

func (g *generator) matchCountedLoopGuard(stmt ast.Statement) (string, ast.Expression, bool) {
	if g == nil || stmt == nil {
		return "", nil, false
	}
	ifExpr, ok := stmt.(*ast.IfExpression)
	if !ok || ifExpr == nil || ifExpr.IfCondition == nil || ifExpr.IfBody == nil {
		return "", nil, false
	}
	if len(ifExpr.ElseIfClauses) != 0 || ifExpr.ElseBody != nil || len(ifExpr.IfBody.Body) != 1 {
		return "", nil, false
	}
	breakStmt, ok := ifExpr.IfBody.Body[0].(*ast.BreakStatement)
	if !ok || breakStmt == nil || breakStmt.Label != nil || breakStmt.Value != nil {
		return "", nil, false
	}
	binary, ok := ifExpr.IfCondition.(*ast.BinaryExpression)
	if !ok || binary == nil || binary.Operator != ">=" {
		return "", nil, false
	}
	leftIdent, ok := binary.Left.(*ast.Identifier)
	if !ok || leftIdent == nil || leftIdent.Name == "" {
		return "", nil, false
	}
	return leftIdent.Name, binary.Right, true
}

func (g *generator) matchCountedLoopIncrement(stmt ast.Statement, name string) bool {
	if g == nil || stmt == nil || name == "" {
		return false
	}
	assign, ok := stmt.(*ast.AssignmentExpression)
	if !ok || assign == nil {
		return false
	}
	switch assign.Operator {
	case ast.AssignmentAdd:
		targetName, _, ok := g.assignmentTargetName(assign.Left)
		if !ok || targetName != name {
			return false
		}
		return isPositiveOneIntegerLiteral(assign.Right)
	case ast.AssignmentAssign:
		targetName, _, ok := g.assignmentTargetName(assign.Left)
		if !ok || targetName != name {
			return false
		}
		binary, ok := assign.Right.(*ast.BinaryExpression)
		if !ok || binary == nil || binary.Operator != "+" {
			return false
		}
		leftIdent, ok := binary.Left.(*ast.Identifier)
		if !ok || leftIdent == nil || leftIdent.Name != name {
			return false
		}
		return isPositiveOneIntegerLiteral(binary.Right)
	default:
		return false
	}
}

func isPositiveOneIntegerLiteral(expr ast.Expression) bool {
	lit, ok := expr.(*ast.IntegerLiteral)
	if !ok || lit == nil || lit.Value == nil {
		return false
	}
	return lit.Value.Sign() == 1 && lit.Value.BitLen() == 1
}

func countedLoopBodyAssignsName(stmts []ast.Statement, name string) bool {
	for _, stmt := range stmts {
		if countedLoopStatementAssignsName(stmt, name) {
			return true
		}
	}
	return false
}

func countedLoopStatementAssignsName(stmt ast.Statement, name string) bool {
	if stmt == nil || name == "" {
		return false
	}
	switch s := stmt.(type) {
	case *ast.FunctionDefinition:
		if s != nil && s.Body != nil {
			return countedLoopBodyAssignsName(s.Body.Body, name)
		}
		return false
	case *ast.AssignmentExpression:
		if target, ok := countedLoopTargetName(s.Left); ok && target == name {
			return true
		}
		return countedLoopExpressionAssignsName(s.Right, name)
	case *ast.BlockExpression:
		return countedLoopBodyAssignsName(s.Body, name)
	case *ast.IfExpression:
		if countedLoopExpressionAssignsName(s.IfCondition, name) || countedLoopBodyAssignsName(s.IfBody.Body, name) {
			return true
		}
		for _, clause := range s.ElseIfClauses {
			if clause == nil {
				continue
			}
			if countedLoopExpressionAssignsName(clause.Condition, name) || countedLoopBodyAssignsName(clause.Body.Body, name) {
				return true
			}
		}
		if s.ElseBody != nil && countedLoopBodyAssignsName(s.ElseBody.Body, name) {
			return true
		}
		return false
	case *ast.WhileLoop:
		return countedLoopExpressionAssignsName(s.Condition, name) || countedLoopBodyAssignsName(s.Body.Body, name)
	case *ast.ForLoop:
		return countedLoopExpressionAssignsName(s.Iterable, name) || countedLoopBodyAssignsName(s.Body.Body, name)
	case *ast.LoopExpression:
		return countedLoopBodyAssignsName(s.Body.Body, name)
	case *ast.ReturnStatement:
		return countedLoopExpressionAssignsName(s.Argument, name)
	case *ast.RaiseStatement:
		return countedLoopExpressionAssignsName(s.Expression, name)
	case *ast.YieldStatement:
		return countedLoopExpressionAssignsName(s.Expression, name)
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			return countedLoopExpressionAssignsName(expr, name)
		}
		return false
	}
}

func countedLoopExpressionAssignsName(expr ast.Expression, name string) bool {
	if expr == nil || name == "" {
		return false
	}
	switch e := expr.(type) {
	case *ast.AssignmentExpression:
		return countedLoopStatementAssignsName(e, name)
	case *ast.BinaryExpression:
		return countedLoopExpressionAssignsName(e.Left, name) || countedLoopExpressionAssignsName(e.Right, name)
	case *ast.UnaryExpression:
		return countedLoopExpressionAssignsName(e.Operand, name)
	case *ast.TypeCastExpression:
		return countedLoopExpressionAssignsName(e.Expression, name)
	case *ast.FunctionCall:
		if countedLoopExpressionAssignsName(e.Callee, name) {
			return true
		}
		for _, arg := range e.Arguments {
			if countedLoopExpressionAssignsName(arg, name) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		return countedLoopExpressionAssignsName(e.Object, name) || countedLoopExpressionAssignsName(e.Member, name)
	case *ast.IndexExpression:
		return countedLoopExpressionAssignsName(e.Object, name) || countedLoopExpressionAssignsName(e.Index, name)
	case *ast.BlockExpression:
		return countedLoopBodyAssignsName(e.Body, name)
	case *ast.IteratorLiteral:
		return countedLoopBodyAssignsName(e.Body, name)
	case *ast.IfExpression:
		return countedLoopStatementAssignsName(e, name)
	case *ast.MatchExpression:
		if countedLoopExpressionAssignsName(e.Subject, name) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if countedLoopExpressionAssignsName(clause.Guard, name) || countedLoopExpressionAssignsName(clause.Body, name) {
				return true
			}
		}
		return false
	case *ast.RangeExpression:
		return countedLoopExpressionAssignsName(e.Start, name) || countedLoopExpressionAssignsName(e.End, name)
	case *ast.LambdaExpression:
		return countedLoopExpressionAssignsName(e.Body, name)
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if countedLoopExpressionAssignsName(part, name) {
				return true
			}
		}
		return false
	case *ast.SpawnExpression:
		return countedLoopExpressionAssignsName(e.Expression, name)
	case *ast.AwaitExpression:
		return countedLoopExpressionAssignsName(e.Expression, name)
	case *ast.PropagationExpression:
		return countedLoopExpressionAssignsName(e.Expression, name)
	case *ast.OrElseExpression:
		if countedLoopExpressionAssignsName(e.Expression, name) {
			return true
		}
		if e.Handler != nil {
			return countedLoopBodyAssignsName(e.Handler.Body, name)
		}
		return false
	case *ast.BreakpointExpression:
		if e.Body != nil {
			return countedLoopBodyAssignsName(e.Body.Body, name)
		}
		return false
	case *ast.EnsureExpression:
		if countedLoopExpressionAssignsName(e.TryExpression, name) {
			return true
		}
		if e.EnsureBlock != nil {
			return countedLoopBodyAssignsName(e.EnsureBlock.Body, name)
		}
		return false
	default:
		return false
	}
}

func countedLoopTargetName(target ast.AssignmentTarget) (string, bool) {
	switch t := target.(type) {
	case *ast.Identifier:
		if t != nil && t.Name != "" {
			return t.Name, true
		}
	case *ast.TypedPattern:
		if t != nil {
			if ident, ok := t.Pattern.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
				return ident.Name, true
			}
		}
	}
	return "", false
}
