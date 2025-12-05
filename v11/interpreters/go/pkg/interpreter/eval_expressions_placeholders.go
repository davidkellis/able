package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type placeholderPlan struct {
	explicitIndices map[int]struct{}
	paramCount      int
}

type placeholderContext int

const (
	contextRoot placeholderContext = iota
	contextCallCallee
	contextOther
)

type placeholderAnalyzer struct {
	explicit        map[int]struct{}
	implicitCount   int
	highestExplicit int
	hasPlaceholder  bool
	relevant        bool
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
		for _, clause := range e.OrClauses {
			if clause == nil {
				continue
			}
			if clause.Condition != nil && expressionContainsPlaceholder(clause.Condition) {
				return true
			}
			if expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.IteratorLiteral:
		return false
	case *ast.LambdaExpression:
		return false
	case *ast.ProcExpression, *ast.SpawnExpression, *ast.AwaitExpression:
		return false
	case *ast.TopicReferenceExpression,
		*ast.Identifier,
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

func analyzePlaceholderExpression(expr ast.Expression) (placeholderPlan, bool, error) {
	analyzer := &placeholderAnalyzer{
		explicit: make(map[int]struct{}),
	}
	if err := analyzer.visitExpression(expr); err != nil {
		return placeholderPlan{}, false, err
	}
	if !analyzer.hasPlaceholder {
		return placeholderPlan{}, false, nil
	}
	paramCount := analyzer.highestExplicit
	implicitTotal := len(analyzer.explicit) + analyzer.implicitCount
	if implicitTotal > paramCount {
		paramCount = implicitTotal
	}
	return placeholderPlan{
		explicitIndices: analyzer.explicit,
		paramCount:      paramCount,
	}, true, nil
}

type placeholderClosure struct {
	interpreter *Interpreter
	expression  ast.Expression
	env         *runtime.Environment
	plan        placeholderPlan
}

func extractFunctionGenerics(funcNode ast.Node) ([]*ast.GenericParameter, []*ast.WhereClauseConstraint) {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		return fn.GenericParams, fn.WhereClause
	case *ast.LambdaExpression:
		return fn.GenericParams, fn.WhereClause
	default:
		return nil, nil
	}
}

func extractFunctionParams(funcNode ast.Node) []*ast.FunctionParameter {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		return fn.Params
	case *ast.LambdaExpression:
		return fn.Params
	default:
		return nil
	}
}

func functionNameForErrors(funcNode ast.Node) string {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		if fn.ID != nil && fn.ID.Name != "" {
			return fn.ID.Name
		}
	case *ast.LambdaExpression:
		return "(lambda)"
	}
	return "(lambda)"
}

func (i *Interpreter) evaluatePlaceholderExpression(expr *ast.PlaceholderExpression, env *runtime.Environment) (runtime.Value, error) {
	state := i.stateFromEnv(env)
	frame, ok := state.currentPlaceholderFrame()
	if !ok {
		return nil, fmt.Errorf("Expression placeholder used outside of placeholder lambda")
	}
	if expr.Index != nil {
		idx := *expr.Index
		val, err := frame.valueAt(idx)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}
	idx, err := frame.nextImplicitIndex()
	if err != nil {
		return nil, err
	}
	val, err := frame.valueAt(idx)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return runtime.NilValue{}, nil
	}
	return val, nil
}

func (i *Interpreter) tryBuildPlaceholderFunction(node ast.Expression, env *runtime.Environment) (runtime.Value, bool, error) {
	state := i.stateFromEnv(env)
	if state.hasPlaceholderFrame() {
		return nil, false, nil
	}
	switch expr := node.(type) {
	case *ast.AssignmentExpression:
		return nil, false, nil
	case *ast.BinaryExpression:
		if expr.Operator == "|>" || expr.Operator == "|>>" {
			return nil, false, nil
		}
	}
	plan, ok, err := analyzePlaceholderExpression(node)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	if call, isCall := node.(*ast.FunctionCall); isCall {
		calleeHas := expressionContainsPlaceholder(call.Callee)
		argsHave := false
		for _, arg := range call.Arguments {
			if expressionContainsPlaceholder(arg) {
				argsHave = true
				break
			}
		}
		if calleeHas && !argsHave {
			return nil, false, nil
		}
	}
	closure := &placeholderClosure{
		interpreter: i,
		expression:  node,
		env:         env,
		plan:        plan,
	}
	fn := runtime.NativeFunctionValue{
		Name:  "<placeholder>",
		Arity: plan.paramCount,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return closure.invoke(args)
		},
	}
	return fn, true, nil
}
