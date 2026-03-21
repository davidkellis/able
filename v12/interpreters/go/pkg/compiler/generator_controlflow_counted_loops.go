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

	resultType := expected
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
	g.seedCountedLoopIntegerFact(bodyCtx, binding, boundExpr)
	bodyLines, ok := g.compileBlockStatement(bodyCtx, coreBlock)
	if !ok {
		return nil, "", "", false
	}

	lines := []string{fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", valueTemp)}
	forLine := fmt.Sprintf("for %s < %s {", binding.GoName, boundGoExpr)
	if linesReferenceLabel(bodyLines, loopLabelName) {
		forLine = fmt.Sprintf("%s: for %s < %s {", loopLabelName, binding.GoName, boundGoExpr)
	}
	lines = append(lines, forLine)
	lines = append(lines, indentLines(bodyLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s++", binding.GoName))
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
