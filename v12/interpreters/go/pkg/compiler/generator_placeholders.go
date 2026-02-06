package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type placeholderPlan struct {
	paramCount int
}

type placeholderAnalyzer struct {
	highestExplicit int
	hasPlaceholder  bool
}

func analyzePlaceholderExpression(expr ast.Expression) (placeholderPlan, bool, error) {
	analyzer := &placeholderAnalyzer{}
	if err := analyzer.visitExpression(expr, true); err != nil {
		return placeholderPlan{}, false, err
	}
	if !analyzer.hasPlaceholder {
		return placeholderPlan{}, false, nil
	}
	paramCount := analyzer.highestExplicit
	if paramCount == 0 {
		paramCount = 1
	}
	return placeholderPlan{paramCount: paramCount}, true, nil
}

func expressionContainsPlaceholder(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.PlaceholderExpression:
		return true
	case *ast.BinaryExpression:
		return expressionContainsPlaceholder(e.Left) || expressionContainsPlaceholder(e.Right)
	case *ast.UnaryExpression:
		return expressionContainsPlaceholder(e.Operand)
	case *ast.TypeCastExpression:
		return expressionContainsPlaceholder(e.Expression)
	case *ast.FunctionCall:
		if expressionContainsPlaceholder(e.Callee) {
			return true
		}
		for _, arg := range e.Arguments {
			if expressionContainsPlaceholder(arg) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		if expressionContainsPlaceholder(e.Object) {
			return true
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return expressionContainsPlaceholder(memberExpr)
		}
		return false
	case *ast.ImplicitMemberExpression:
		return false
	case *ast.IndexExpression:
		return expressionContainsPlaceholder(e.Object) || expressionContainsPlaceholder(e.Index)
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if statementContainsPlaceholder(stmt) {
				return true
			}
		}
		return false
	case *ast.LoopExpression:
		if e.Body == nil {
			return false
		}
		return expressionContainsPlaceholder(e.Body)
	case *ast.AssignmentExpression:
		if expressionContainsPlaceholder(e.Right) {
			return true
		}
		if targetExpr, ok := e.Left.(ast.Expression); ok {
			return expressionContainsPlaceholder(targetExpr)
		}
		return false
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if expressionContainsPlaceholder(part) {
				return true
			}
		}
		return false
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && expressionContainsPlaceholder(field.Value) {
				return true
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if expressionContainsPlaceholder(src) {
				return true
			}
		}
		return false
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			if expressionContainsPlaceholder(el) {
				return true
			}
		}
		return false
	case *ast.RangeExpression:
		return expressionContainsPlaceholder(e.Start) || expressionContainsPlaceholder(e.End)
	case *ast.MatchExpression:
		if expressionContainsPlaceholder(e.Subject) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && expressionContainsPlaceholder(clause.Guard) {
				return true
			}
			if expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.OrElseExpression:
		return expressionContainsPlaceholder(e.Expression) || expressionContainsPlaceholder(e.Handler)
	case *ast.RescueExpression:
		if expressionContainsPlaceholder(e.MonitoredExpression) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && expressionContainsPlaceholder(clause.Guard) {
				return true
			}
			if expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.EnsureExpression:
		return expressionContainsPlaceholder(e.TryExpression) || expressionContainsPlaceholder(e.EnsureBlock)
	case *ast.IfExpression:
		if expressionContainsPlaceholder(e.IfCondition) || expressionContainsPlaceholder(e.IfBody) {
			return true
		}
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if expressionContainsPlaceholder(clause.Condition) || expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		if e.ElseBody != nil && expressionContainsPlaceholder(e.ElseBody) {
			return true
		}
		return false
	case *ast.IteratorLiteral:
		return false
	case *ast.LambdaExpression:
		return false
	case *ast.SpawnExpression, *ast.AwaitExpression:
		return false
	case *ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return false
	default:
		return false
	}
}

func statementContainsPlaceholder(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return expressionContainsPlaceholder(expr)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return expressionContainsPlaceholder(s.Argument)
		}
	case *ast.RaiseStatement:
		if s.Expression != nil {
			return expressionContainsPlaceholder(s.Expression)
		}
	case *ast.ForLoop:
		if expressionContainsPlaceholder(s.Iterable) {
			return true
		}
		return expressionContainsPlaceholder(s.Body)
	case *ast.WhileLoop:
		if expressionContainsPlaceholder(s.Condition) {
			return true
		}
		return expressionContainsPlaceholder(s.Body)
	case *ast.BreakStatement:
		if s.Value != nil {
			return expressionContainsPlaceholder(s.Value)
		}
	case *ast.ContinueStatement:
		return false
	case *ast.YieldStatement:
		if s.Expression != nil {
			return expressionContainsPlaceholder(s.Expression)
		}
	case *ast.PreludeStatement, *ast.ExternFunctionBody, *ast.ImportStatement, *ast.DynImportStatement, *ast.PackageStatement:
		return false
	default:
		return false
	}
	return false
}

func (p *placeholderAnalyzer) visitExpression(expr ast.Expression, root bool) error {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.PlaceholderExpression:
		p.hasPlaceholder = true
		idx := 1
		if e.Index != nil {
			idx = *e.Index
		}
		if idx <= 0 {
			return fmt.Errorf("placeholder index must be positive, found @%d", idx)
		}
		if idx > p.highestExplicit {
			p.highestExplicit = idx
		}
	case *ast.BinaryExpression:
		if e.Operator == "|>" || e.Operator == "|>>" {
			return nil
		}
		if err := p.visitExpression(e.Left, false); err != nil {
			return err
		}
		return p.visitExpression(e.Right, false)
	case *ast.UnaryExpression:
		return p.visitExpression(e.Operand, false)
	case *ast.FunctionCall:
		if !root {
			return nil
		}
		if err := p.visitExpression(e.Callee, false); err != nil {
			return err
		}
		for _, arg := range e.Arguments {
			if err := p.visitExpression(arg, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.MemberAccessExpression:
		if err := p.visitExpression(e.Object, false); err != nil {
			return err
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return p.visitExpression(memberExpr, false)
		}
		return nil
	case *ast.ImplicitMemberExpression:
		return nil
	case *ast.IndexExpression:
		if err := p.visitExpression(e.Object, false); err != nil {
			return err
		}
		return p.visitExpression(e.Index, false)
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if err := p.visitStatement(stmt); err != nil {
				return err
			}
		}
		return nil
	case *ast.AssignmentExpression:
		if err := p.visitExpression(e.Right, false); err != nil {
			return err
		}
		if targetExpr, ok := e.Left.(ast.Expression); ok {
			return p.visitExpression(targetExpr, false)
		}
		return nil
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if err := p.visitExpression(part, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil {
				if err := p.visitExpression(field.Value, false); err != nil {
					return err
				}
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if err := p.visitExpression(src, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			if err := p.visitExpression(el, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.RangeExpression:
		if err := p.visitExpression(e.Start, false); err != nil {
			return err
		}
		return p.visitExpression(e.End, false)
	case *ast.MatchExpression:
		if err := p.visitExpression(e.Subject, false); err != nil {
			return err
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				if err := p.visitExpression(clause.Guard, false); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.OrElseExpression:
		if err := p.visitExpression(e.Expression, false); err != nil {
			return err
		}
		return p.visitExpression(e.Handler, false)
	case *ast.RescueExpression:
		if err := p.visitExpression(e.MonitoredExpression, false); err != nil {
			return err
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				if err := p.visitExpression(clause.Guard, false); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.EnsureExpression:
		if err := p.visitExpression(e.TryExpression, false); err != nil {
			return err
		}
		return p.visitExpression(e.EnsureBlock, false)
	case *ast.IfExpression:
		if err := p.visitExpression(e.IfCondition, false); err != nil {
			return err
		}
		if err := p.visitExpression(e.IfBody, false); err != nil {
			return err
		}
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if err := p.visitExpression(clause.Condition, false); err != nil {
				return err
			}
			if err := p.visitExpression(clause.Body, false); err != nil {
				return err
			}
		}
		if e.ElseBody != nil {
			if err := p.visitExpression(e.ElseBody, false); err != nil {
				return err
			}
		}
		return nil
	case *ast.IteratorLiteral:
		return nil
	case *ast.LambdaExpression:
		return nil
	case *ast.SpawnExpression, *ast.AwaitExpression:
		return nil
	case *ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return nil
	default:
		return nil
	}
	return nil
}

func (p *placeholderAnalyzer) visitStatement(stmt ast.Statement) error {
	if stmt == nil {
		return nil
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return p.visitExpression(expr, false)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return p.visitExpression(s.Argument, false)
		}
	case *ast.RaiseStatement:
		if s.Expression != nil {
			return p.visitExpression(s.Expression, false)
		}
	case *ast.ForLoop:
		if err := p.visitExpression(s.Iterable, false); err != nil {
			return err
		}
		return p.visitExpression(s.Body, false)
	case *ast.WhileLoop:
		if err := p.visitExpression(s.Condition, false); err != nil {
			return err
		}
		return p.visitExpression(s.Body, false)
	case *ast.BreakStatement:
		if s.Value != nil {
			return p.visitExpression(s.Value, false)
		}
	case *ast.ContinueStatement:
		return nil
	case *ast.YieldStatement:
		if s.Expression != nil {
			return p.visitExpression(s.Expression, false)
		}
	case *ast.PreludeStatement, *ast.ExternFunctionBody, *ast.ImportStatement, *ast.DynImportStatement, *ast.PackageStatement:
		return nil
	default:
		return nil
	}
	return nil
}

func (g *generator) compilePlaceholderLambda(ctx *compileContext, expr ast.Expression) (string, string, bool) {
	if ctx == nil || ctx.inPlaceholder {
		return "", "", false
	}
	if expr == nil {
		ctx.setReason("missing placeholder expression")
		return "", "", false
	}
	switch e := expr.(type) {
	case *ast.AssignmentExpression:
		return "", "", false
	case *ast.BinaryExpression:
		if e.Operator == "|>" || e.Operator == "|>>" {
			return "", "", false
		}
	}
	plan, ok, err := analyzePlaceholderExpression(expr)
	if err != nil {
		ctx.setReason(err.Error())
		return "", "", false
	}
	if !ok {
		return "", "", false
	}
	if call, isCall := expr.(*ast.FunctionCall); isCall {
		calleeHas := expressionContainsPlaceholder(call.Callee)
		argsHave := false
		for _, arg := range call.Arguments {
			if expressionContainsPlaceholder(arg) {
				argsHave = true
				break
			}
		}
		if calleeHas && !argsHave {
			return "", "", false
		}
	}

	lambdaCtx := ctx.child()
	lambdaCtx.inPlaceholder = true
	lambdaCtx.placeholderParams = make(map[int]paramInfo, plan.paramCount)
	paramLines := make([]string, 0, plan.paramCount)
	var firstParam paramInfo
	for i := 1; i <= plan.paramCount; i++ {
		name := fmt.Sprintf("__able_ph_%d", i)
		param := paramInfo{Name: name, GoName: name, GoType: "runtime.Value"}
		lambdaCtx.placeholderParams[i] = param
		lambdaCtx.locals[name] = param
		paramLines = append(paramLines, fmt.Sprintf("%s := args[%d]", name, i-1))
		if i == 1 {
			firstParam = param
		}
	}
	if plan.paramCount > 0 {
		lambdaCtx.implicitReceiver = firstParam
		lambdaCtx.hasImplicitReceiver = true
	}

	exprValue, exprType, ok := g.compileExpr(lambdaCtx, expr, "")
	if !ok {
		return "", "", false
	}
	implLines := make([]string, 0, len(paramLines)+3)
	implLines = append(implLines, "if __able_runtime != nil && callCtx != nil && callCtx.Env != nil { prevEnv := __able_runtime.SwapEnv(callCtx.Env); defer __able_runtime.SwapEnv(prevEnv) }")
	implLines = append(implLines, paramLines...)
	if g.isVoidType(exprType) {
		if exprValue != "" {
			implLines = append(implLines, fmt.Sprintf("_ = %s", exprValue))
		}
		implLines = append(implLines, "return runtime.VoidValue{}, nil")
	} else {
		resultExpr := exprValue
		if exprType != "runtime.Value" {
			converted, ok := g.runtimeValueExpr(exprValue, exprType)
			if !ok {
				ctx.setReason("placeholder result unsupported")
				return "", "", false
			}
			resultExpr = converted
		}
		implLines = append(implLines, fmt.Sprintf("return %s, nil", resultExpr))
	}
	implBody := strings.Join(implLines, "; ")
	lambdaExpr := fmt.Sprintf("runtime.Value(runtime.NativeFunctionValue{Name: %q, Arity: %d, Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) { %s }})", "<placeholder>", plan.paramCount, implBody)
	return lambdaExpr, "runtime.Value", true
}

func (g *generator) compilePlaceholderExpression(ctx *compileContext, expr *ast.PlaceholderExpression, expected string) (string, string, bool) {
	if ctx == nil || ctx.placeholderParams == nil {
		ctx.setReason("placeholder used outside lambda")
		return "", "", false
	}
	if expr == nil {
		ctx.setReason("missing placeholder")
		return "", "", false
	}
	idx := 1
	if expr.Index != nil {
		idx = *expr.Index
	}
	param, ok := ctx.placeholderParams[idx]
	if !ok {
		ctx.setReason("placeholder index out of range")
		return "", "", false
	}
	if expected == "" || expected == param.GoType {
		return param.GoName, param.GoType, true
	}
	if param.GoType == "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(param.GoName, expected)
		if !ok {
			ctx.setReason("placeholder type mismatch")
			return "", "", false
		}
		return converted, expected, true
	}
	ctx.setReason("placeholder type mismatch")
	return "", "", false
}
