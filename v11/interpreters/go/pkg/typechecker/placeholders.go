package typechecker

import "able/interpreter10-go/pkg/ast"

type placeholderPlan struct {
	paramCount int
}

type placeholderAnalyzer struct {
	highestIndex   int
	hasPlaceholder bool
}

func placeholderFunctionPlan(expr ast.Expression) (placeholderPlan, bool) {
	if expr == nil {
		return placeholderPlan{}, false
	}
	if _, ok := expr.(*ast.AssignmentExpression); ok {
		return placeholderPlan{}, false
	}
	if bin, ok := expr.(*ast.BinaryExpression); ok {
		if bin.Operator == "|>" || bin.Operator == "|>>" {
			return placeholderPlan{}, false
		}
	}
	analyzer := &placeholderAnalyzer{}
	if err := analyzer.visitExpression(expr); err != nil {
		return placeholderPlan{}, false
	}
	if !analyzer.hasPlaceholder {
		return placeholderPlan{}, false
	}
	if call, ok := expr.(*ast.FunctionCall); ok {
		calleeHas := expressionContainsPlaceholder(call.Callee)
		argsHave := false
		for _, arg := range call.Arguments {
			if expressionContainsPlaceholder(arg) {
				argsHave = true
				break
			}
		}
		if calleeHas && !argsHave {
			return placeholderPlan{}, false
		}
	}
	count := analyzer.highestIndex
	if count == 0 {
		count = 1
	}
	return placeholderPlan{paramCount: count}, true
}

func expressionContainsPlaceholder(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	analyzer := &placeholderAnalyzer{}
	_ = analyzer.visitExpression(expr)
	return analyzer.hasPlaceholder
}

func (p *placeholderAnalyzer) visitExpression(expr ast.Expression) error {
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
			return nil
		}
		if idx > p.highestIndex {
			p.highestIndex = idx
		}
		return nil
	case *ast.BinaryExpression:
		if err := p.visitExpression(e.Left); err != nil {
			return err
		}
		return p.visitExpression(e.Right)
	case *ast.UnaryExpression:
		return p.visitExpression(e.Operand)
	case *ast.FunctionCall:
		if err := p.visitExpression(e.Callee); err != nil {
			return err
		}
		for _, arg := range e.Arguments {
			if err := p.visitExpression(arg); err != nil {
				return err
			}
		}
		return nil
	case *ast.MemberAccessExpression:
		if err := p.visitExpression(e.Object); err != nil {
			return err
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return p.visitExpression(memberExpr)
		}
		return nil
	case *ast.ImplicitMemberExpression:
		return nil
	case *ast.IndexExpression:
		if err := p.visitExpression(e.Object); err != nil {
			return err
		}
		return p.visitExpression(e.Index)
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if err := p.visitStatement(stmt); err != nil {
				return err
			}
		}
		return nil
	case *ast.AssignmentExpression:
		if err := p.visitExpression(e.Right); err != nil {
			return err
		}
		if targetExpr, ok := e.Left.(ast.Expression); ok {
			return p.visitExpression(targetExpr)
		}
		return nil
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if err := p.visitExpression(part); err != nil {
				return err
			}
		}
		return nil
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && field.Value != nil {
				if err := p.visitExpression(field.Value); err != nil {
					return err
				}
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if err := p.visitExpression(src); err != nil {
				return err
			}
		}
		return nil
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			if err := p.visitExpression(el); err != nil {
				return err
			}
		}
		return nil
	case *ast.RangeExpression:
		if err := p.visitExpression(e.Start); err != nil {
			return err
		}
		return p.visitExpression(e.End)
	case *ast.MatchExpression:
		if err := p.visitExpression(e.Subject); err != nil {
			return err
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				if err := p.visitExpression(clause.Guard); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		return nil
	case *ast.OrElseExpression:
		if err := p.visitExpression(e.Expression); err != nil {
			return err
		}
		return p.visitExpression(e.Handler)
	case *ast.RescueExpression:
		if err := p.visitExpression(e.MonitoredExpression); err != nil {
			return err
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				if err := p.visitExpression(clause.Guard); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		return nil
	case *ast.EnsureExpression:
		if err := p.visitExpression(e.TryExpression); err != nil {
			return err
		}
		return p.visitExpression(e.EnsureBlock)
	case *ast.IfExpression:
		if err := p.visitExpression(e.IfCondition); err != nil {
			return err
		}
		if err := p.visitExpression(e.IfBody); err != nil {
			return err
		}
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if err := p.visitExpression(clause.Condition); err != nil {
				return err
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		if e.ElseBody != nil {
			if err := p.visitExpression(e.ElseBody); err != nil {
				return err
			}
		}
		return nil
	case *ast.PropagationExpression:
		return p.visitExpression(e.Expression)
	case *ast.AwaitExpression:
		return p.visitExpression(e.Expression)
	case *ast.LoopExpression:
		return p.visitExpression(e.Body)
	case *ast.IteratorLiteral,
		*ast.LambdaExpression,
		*ast.ProcExpression,
		*ast.SpawnExpression,
		*ast.Identifier,
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
		return p.visitExpression(expr)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return p.visitExpression(s.Argument)
		}
	case *ast.RaiseStatement:
		if s.Expression != nil {
			return p.visitExpression(s.Expression)
		}
	case *ast.ForLoop:
		if err := p.visitExpression(s.Iterable); err != nil {
			return err
		}
		return p.visitExpression(s.Body)
	case *ast.WhileLoop:
		if err := p.visitExpression(s.Condition); err != nil {
			return err
		}
		return p.visitExpression(s.Body)
	case *ast.BreakStatement:
		if s.Value != nil {
			return p.visitExpression(s.Value)
		}
	case *ast.ContinueStatement:
		return nil
	case *ast.YieldStatement:
		if s.Expression != nil {
			return p.visitExpression(s.Expression)
		}
	case *ast.PreludeStatement, *ast.ExternFunctionBody, *ast.ImportStatement, *ast.DynImportStatement, *ast.PackageStatement:
		return nil
	default:
		return nil
	}
	return nil
}
