package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

type IRLowerer struct {
	types typechecker.InferenceMap

	current  *IRBlock
	function *IRFunction
	blocks   map[string]*IRBlock

	nextValueID int
	nextBlockID int

	scopes      []map[string]*IRSlot
	errorBlock  *IRBlock
	loopStack   []loopContext
	errorStack  []errorContext
	breakpoints []breakpointContext

	spawnCount int
	iterCount  int

	iteratorStack []string
}

type loopContext struct {
	breakLabel    string
	continueLabel string
	resultSlot    *IRSlot
}

type errorContext struct {
	label string
	slot  *IRSlot
}

type breakpointContext struct {
	label         string
	breakLabel    string
	resultSlot    *IRSlot
	continueLabel string
}

func LowerFunction(def *ast.FunctionDefinition, pkg string, types typechecker.InferenceMap) (*IRFunction, error) {
	if def == nil || def.ID == nil {
		return nil, fmt.Errorf("compiler: missing function definition")
	}
	l := &IRLowerer{
		types:  types,
		blocks: make(map[string]*IRBlock),
	}
	fn := &IRFunction{
		Name:       def.ID.Name,
		Package:    pkg,
		Params:     nil,
		ReturnType: l.typeOf(def),
		Locals:     nil,
		Blocks:     make(map[string]*IRBlock),
		EntryLabel: "entry",
		Source:     def,
	}
	l.function = fn
	entry := l.newNamedBlock("entry")
	l.current = entry

	l.pushScope()
	if err := l.lowerParams(def.Params); err != nil {
		return nil, err
	}
	if def.Body == nil {
		return nil, fmt.Errorf("compiler: function %s missing body", def.ID.Name)
	}
	value, err := l.lowerBlock(def.Body)
	if err != nil {
		return nil, err
	}
	if l.current.Terminator == nil {
		if err := l.current.SetTerminator(&IRReturn{Value: value, Source: def.Body}); err != nil {
			return nil, err
		}
	}
	l.popScope()

	fn.Blocks = l.blocks
	return fn, nil
}

func LowerFunctionWithAnalysis(def *ast.FunctionDefinition, pkg string, analysis *ProgramAnalysis) (*IRFunction, error) {
	if analysis == nil {
		return nil, fmt.Errorf("compiler: analysis is nil")
	}
	var types typechecker.InferenceMap
	if analysis.Typecheck.Inferred != nil {
		types = analysis.Typecheck.Inferred[pkg]
	}
	return LowerFunction(def, pkg, types)
}

func (l *IRLowerer) lowerParams(params []*ast.FunctionParameter) error {
	for _, param := range params {
		if param == nil {
			continue
		}
		name, ok := l.paramName(param.Name)
		if !ok {
			return fmt.Errorf("compiler: unsupported parameter pattern")
		}
		value := l.newValue(name, l.typeOf(param), param)
		l.function.Params = append(l.function.Params, IRParam{Value: value, Mutable: false})
		slot := l.declareSlot(name, l.typeOf(param), false, param)
		if err := l.current.Append(&IRStore{Slot: slot, Value: IRValueUse{Value: value}, Source: param}); err != nil {
			return err
		}
	}
	return nil
}

func (l *IRLowerer) lowerBlock(block *ast.BlockExpression) (IRValueRef, error) {
	if block == nil {
		return IRConst{Literal: ast.NewNilLiteral(), TypeRef: l.typeOf(block)}, nil
	}
	var last IRValueRef = IRConst{Literal: ast.NewNilLiteral(), TypeRef: l.typeOf(block)}
	for _, stmt := range block.Body {
		if l.current.Terminator != nil {
			return last, nil
		}
		if stmt == nil {
			continue
		}
		switch s := stmt.(type) {
		case *ast.ReturnStatement:
			val, err := l.lowerExpr(s.Argument)
			if err != nil {
				return nil, err
			}
			if err := l.current.SetTerminator(&IRReturn{Value: val, Source: s}); err != nil {
				return nil, err
			}
			return val, nil
		case *ast.WhileLoop:
			val, err := l.lowerWhileLoop(s)
			if err != nil {
				return nil, err
			}
			last = val
		case *ast.ForLoop:
			val, err := l.lowerForLoop(s)
			if err != nil {
				return nil, err
			}
			last = val
		case *ast.BreakStatement:
			if err := l.lowerBreak(s); err != nil {
				return nil, err
			}
		case *ast.ContinueStatement:
			if err := l.lowerContinue(s); err != nil {
				return nil, err
			}
		case *ast.RaiseStatement:
			if err := l.lowerRaise(s); err != nil {
				return nil, err
			}
		case *ast.YieldStatement:
			if err := l.lowerYield(s); err != nil {
				return nil, err
			}
		default:
			if expr, ok := stmt.(ast.Expression); ok {
				val, err := l.lowerExpr(expr)
				if err != nil {
					return nil, err
				}
				last = val
				continue
			}
			return nil, fmt.Errorf("compiler: unsupported statement in block: %T", stmt)
		}
	}
	return last, nil
}

func (l *IRLowerer) lowerExpr(expr ast.Expression) (IRValueRef, error) {
	if expr == nil {
		return IRConst{Literal: ast.NewNilLiteral(), TypeRef: l.typeOf(expr)}, nil
	}
	switch e := expr.(type) {
	case *ast.ArrayLiteral:
		return l.lowerArrayLiteral(e)
	case *ast.MapLiteral:
		return l.lowerMapLiteral(e)
	case ast.Literal:
		return IRConst{Literal: e, TypeRef: l.typeOf(e)}, nil
	case *ast.Identifier:
		return l.loadIdentifier(e)
	case *ast.BlockExpression:
		l.pushScope()
		val, err := l.lowerBlock(e)
		l.popScope()
		return val, err
	case *ast.UnaryExpression:
		operand, err := l.lowerExpr(e.Operand)
		if err != nil {
			return nil, err
		}
		return l.emitCompute(IROpUnary, string(e.Operator), []IRValueRef{operand}, e)
	case *ast.BinaryExpression:
		left, err := l.lowerExpr(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := l.lowerExpr(e.Right)
		if err != nil {
			return nil, err
		}
		return l.emitCompute(IROpBinary, e.Operator, []IRValueRef{left, right}, e)
	case *ast.FunctionCall:
		return l.lowerCall(e)
	case *ast.MemberAccessExpression:
		obj, err := l.lowerExpr(e.Object)
		if err != nil {
			return nil, err
		}
		member, ok := e.Member.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("compiler: unsupported member access expression")
		}
		memberVal, err := l.lowerExpr(member)
		if err != nil {
			return nil, err
		}
		return l.emitInvoke(IROpMember, nil, []IRValueRef{obj, memberVal}, e)
	case *ast.IndexExpression:
		obj, err := l.lowerExpr(e.Object)
		if err != nil {
			return nil, err
		}
		idx, err := l.lowerExpr(e.Index)
		if err != nil {
			return nil, err
		}
		return l.emitInvoke(IROpIndex, nil, []IRValueRef{obj, idx}, e)
	case *ast.TypeCastExpression:
		value, err := l.lowerExpr(e.Expression)
		if err != nil {
			return nil, err
		}
		return l.emitInvoke(IROpCast, nil, []IRValueRef{value}, e)
	case *ast.RangeExpression:
		start, err := l.lowerExpr(e.Start)
		if err != nil {
			return nil, err
		}
		end, err := l.lowerExpr(e.End)
		if err != nil {
			return nil, err
		}
		return l.emitInvoke(IROpRange, nil, []IRValueRef{start, end}, e)
	case *ast.StringInterpolation:
		return l.lowerStringInterpolation(e)
	case *ast.StructLiteral:
		return l.lowerStructLiteral(e)
	case *ast.IteratorLiteral:
		return l.lowerIteratorLiteral(e)
	case *ast.IfExpression:
		return l.lowerIf(e)
	case *ast.MatchExpression:
		return l.lowerMatch(e)
	case *ast.AssignmentExpression:
		return l.lowerAssignment(e)
	case *ast.LoopExpression:
		return l.lowerLoopExpression(e)
	case *ast.BreakpointExpression:
		return l.lowerBreakpointExpression(e)
	case *ast.SpawnExpression:
		return l.lowerSpawn(e)
	case *ast.AwaitExpression:
		return l.lowerAwait(e)
	case *ast.PropagationExpression:
		return l.lowerPropagation(e)
	case *ast.OrElseExpression:
		return l.lowerOrElse(e)
	case *ast.RescueExpression:
		return l.lowerRescue(e)
	case *ast.EnsureExpression:
		return l.lowerEnsure(e)
	default:
		return nil, fmt.Errorf("compiler: unsupported expression %T", expr)
	}
}

func (l *IRLowerer) lowerIf(expr *ast.IfExpression) (IRValueRef, error) {
	condition, err := l.lowerExpr(expr.IfCondition)
	if err != nil {
		return nil, err
	}
	thenBlock := l.newNamedBlock("if_then")
	elseBlock := l.newNamedBlock("if_else")
	joinBlock := l.newNamedBlock("if_join")
	resultSlot := l.newSlot("if_result", l.typeOf(expr), false, expr)

	if err := l.current.SetTerminator(&IRBranch{Condition: condition, TrueLabel: thenBlock.Label, FalseLabel: elseBlock.Label}); err != nil {
		return nil, err
	}

	l.current = thenBlock
	thenVal, err := l.lowerBlock(expr.IfBody)
	if err != nil {
		return nil, err
	}
	if l.current.Terminator == nil {
		if err := l.current.Append(&IRStore{Slot: resultSlot, Value: thenVal, Source: expr.IfBody}); err != nil {
			return nil, err
		}
		if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = elseBlock
	var elseVal IRValueRef = IRConst{Literal: ast.NewNilLiteral(), TypeRef: l.typeOf(expr)}
	if expr.ElseBody != nil {
		elseVal, err = l.lowerBlock(expr.ElseBody)
		if err != nil {
			return nil, err
		}
	}
	if l.current.Terminator == nil {
		if err := l.current.Append(&IRStore{Slot: resultSlot, Value: elseVal, Source: expr.ElseBody}); err != nil {
			return nil, err
		}
		if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = joinBlock
	result := l.newValue("if_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerMatch(expr *ast.MatchExpression) (IRValueRef, error) {
	if len(expr.Clauses) == 0 {
		return nil, fmt.Errorf("compiler: match expression missing clauses")
	}
	subject, err := l.lowerExpr(expr.Subject)
	if err != nil {
		return nil, err
	}
	resultSlot := l.newSlot("match_result", l.typeOf(expr), false, expr)
	caseBlocks := make([]*IRBlock, len(expr.Clauses))
	for i := range expr.Clauses {
		caseBlocks[i] = l.newNamedBlock(fmt.Sprintf("match_case_%d", i))
	}
	joinBlock := l.newNamedBlock("match_join")
	noMatch := l.newNamedBlock("match_nomatch")

	if err := l.current.SetTerminator(&IRJump{Target: caseBlocks[0].Label}); err != nil {
		return nil, err
	}

	for i, clause := range expr.Clauses {
		if clause == nil {
			return nil, fmt.Errorf("compiler: match clause missing")
		}
		nextLabel := noMatch.Label
		if i+1 < len(caseBlocks) {
			nextLabel = caseBlocks[i+1].Label
		}
		l.current = caseBlocks[i]
		l.pushScope()
		l.declarePatternBindings(clause.Pattern, true)
		bindings, err := l.collectPatternBindings(clause.Pattern)
		if err != nil {
			return nil, err
		}

		patternErr := l.newValue("match_err", l.typeOf(expr), clause)
		if err := l.current.Append(&IRDestructure{
			Pattern:  clause.Pattern,
			Value:    subject,
			Error:    patternErr,
			Source:   clause,
			Bindings: bindings,
		}); err != nil {
			return nil, err
		}
		bodyBlock := l.newNamedBlock(fmt.Sprintf("match_body_%d", i))
		if clause.Guard != nil {
			guardBlock := l.newNamedBlock(fmt.Sprintf("match_guard_%d", i))
			if err := l.current.SetTerminator(&IRCheck{
				Error:    IRValueUse{Value: patternErr},
				OkLabel:  guardBlock.Label,
				ErrLabel: nextLabel,
			}); err != nil {
				return nil, err
			}
			l.current = guardBlock
			guardVal, err := l.lowerExpr(clause.Guard)
			if err != nil {
				return nil, err
			}
			if err := l.current.SetTerminator(&IRBranch{
				Condition:  guardVal,
				TrueLabel:  bodyBlock.Label,
				FalseLabel: nextLabel,
			}); err != nil {
				return nil, err
			}
		} else {
			if err := l.current.SetTerminator(&IRCheck{
				Error:    IRValueUse{Value: patternErr},
				OkLabel:  bodyBlock.Label,
				ErrLabel: nextLabel,
			}); err != nil {
				return nil, err
			}
		}

		l.current = bodyBlock
		bodyVal, err := l.lowerExpr(clause.Body)
		if err != nil {
			return nil, err
		}
		if l.current.Terminator == nil {
			if err := l.current.Append(&IRStore{Slot: resultSlot, Value: bodyVal, Source: clause.Body}); err != nil {
				return nil, err
			}
			if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
				return nil, err
			}
		}
		l.popScope()
	}

	l.current = noMatch
	if err := l.current.SetTerminator(&IRUnreachable{}); err != nil {
		return nil, err
	}

	l.current = joinBlock
	result := l.newValue("match_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerAssignment(expr *ast.AssignmentExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: assignment is nil")
	}
	switch expr.Operator {
	case ast.AssignmentAssign, ast.AssignmentDeclare:
	default:
		return nil, fmt.Errorf("compiler: unsupported assignment operator %s", expr.Operator)
	}

	right, err := l.lowerExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	switch target := expr.Left.(type) {
	case *ast.Identifier:
		if expr.Operator == ast.AssignmentDeclare {
			l.declareSlot(target.Name, l.typeOf(target), true, target)
		} else if l.lookupSlot(target.Name) == nil {
			l.declareSlot(target.Name, l.typeOf(target), true, target)
		}
		slot := l.lookupSlot(target.Name)
		if err := l.current.Append(&IRStore{Slot: slot, Value: right, Source: expr}); err != nil {
			return nil, err
		}
		return right, nil
	case *ast.TypedPattern:
		return l.lowerPatternAssignment(expr, target.Pattern, right)
	case ast.Pattern:
		return l.lowerPatternAssignment(expr, target, right)
	default:
		return nil, fmt.Errorf("compiler: unsupported assignment target %T", expr.Left)
	}
}

func (l *IRLowerer) lowerRaise(stmt *ast.RaiseStatement) error {
	if stmt == nil {
		return fmt.Errorf("compiler: raise is nil")
	}
	val, err := l.lowerExpr(stmt.Expression)
	if err != nil {
		return err
	}
	return l.propagateError(val, stmt)
}

func (l *IRLowerer) lowerPropagation(expr *ast.PropagationExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: propagation is nil")
	}
	value, err := l.lowerExpr(expr.Expression)
	if err != nil {
		return nil, err
	}
	return l.emitInvoke(IROpPropagate, nil, []IRValueRef{value}, expr)
}

func (l *IRLowerer) lowerOrElse(expr *ast.OrElseExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: or-else is nil")
	}
	resultSlot := l.newSlot("or_result", l.typeOf(expr), false, expr)
	errSlot := l.newSlot("or_error", nil, false, expr)

	failErrBlock := l.newNamedBlock("or_fail_error")
	failNilBlock := l.newNamedBlock("or_fail_nil")
	joinBlock := l.newNamedBlock("or_join")

	l.pushErrorHandler(failErrBlock.Label, errSlot)
	value, err := l.lowerExpr(expr.Expression)
	if err != nil {
		l.popErrorHandler()
		return nil, err
	}
	l.popErrorHandler()

	if l.current.Terminator == nil {
		if err := l.current.Append(&IRStore{Slot: resultSlot, Value: value, Source: expr.Expression}); err != nil {
			return nil, err
		}
		errValue, err := l.emitCompute(IROpAsError, "", []IRValueRef{value}, expr)
		if err != nil {
			return nil, err
		}
		if err := l.current.Append(&IRStore{Slot: errSlot, Value: errValue, Source: expr}); err != nil {
			return nil, err
		}
		isNil, err := l.emitCompute(IROpIsNil, "", []IRValueRef{value}, expr)
		if err != nil {
			return nil, err
		}
		checkErrBlock := l.newNamedBlock("or_check_error")
		if err := l.current.SetTerminator(&IRBranch{
			Condition:  isNil,
			TrueLabel:  failNilBlock.Label,
			FalseLabel: checkErrBlock.Label,
		}); err != nil {
			return nil, err
		}
		l.current = checkErrBlock
		errIsNil, err := l.emitCompute(IROpIsNil, "", []IRValueRef{errValue}, expr)
		if err != nil {
			return nil, err
		}
		isErr, err := l.emitCompute(IROpUnary, "!", []IRValueRef{errIsNil}, expr)
		if err != nil {
			return nil, err
		}
		if err := l.current.SetTerminator(&IRBranch{
			Condition:  isErr,
			TrueLabel:  failErrBlock.Label,
			FalseLabel: joinBlock.Label,
		}); err != nil {
			return nil, err
		}
	}

	handlerBlock := func(block *IRBlock, bindError bool) error {
		l.current = block
		l.pushScope()
		if bindError && expr.ErrorBinding != nil && expr.ErrorBinding.Name != "" {
			slot := l.declareSlot(expr.ErrorBinding.Name, l.typeOf(expr.ErrorBinding), false, expr.ErrorBinding)
			errValue := l.newValue("or_err", nil, expr)
			if err := l.current.Append(&IRLoad{Dest: errValue, Slot: errSlot, Source: expr}); err != nil {
				return err
			}
			if err := l.current.Append(&IRStore{Slot: slot, Value: IRValueUse{Value: errValue}, Source: expr}); err != nil {
				return err
			}
		}
		val, err := l.lowerBlock(expr.Handler)
		if err != nil {
			l.popScope()
			return err
		}
		l.popScope()
		if l.current.Terminator == nil {
			if err := l.current.Append(&IRStore{Slot: resultSlot, Value: val, Source: expr.Handler}); err != nil {
				return err
			}
			return l.current.SetTerminator(&IRJump{Target: joinBlock.Label})
		}
		return nil
	}

	if err := handlerBlock(failErrBlock, true); err != nil {
		return nil, err
	}
	if err := handlerBlock(failNilBlock, false); err != nil {
		return nil, err
	}

	l.current = joinBlock
	result := l.newValue("or_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerRescue(expr *ast.RescueExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: rescue is nil")
	}
	resultSlot := l.newSlot("rescue_result", l.typeOf(expr), false, expr)
	errSlot := l.newSlot("rescue_error", nil, false, expr)
	handlerBlock := l.newNamedBlock("rescue_handler")
	joinBlock := l.newNamedBlock("rescue_join")

	l.pushErrorHandler(handlerBlock.Label, errSlot)
	value, err := l.lowerExpr(expr.MonitoredExpression)
	if err != nil {
		l.popErrorHandler()
		return nil, err
	}
	l.popErrorHandler()

	if l.current.Terminator == nil {
		if err := l.current.Append(&IRStore{Slot: resultSlot, Value: value, Source: expr.MonitoredExpression}); err != nil {
			return nil, err
		}
		if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = handlerBlock
	errValue := l.newValue("rescue_error", nil, expr)
	if err := l.current.Append(&IRLoad{Dest: errValue, Slot: errSlot, Source: expr}); err != nil {
		return nil, err
	}

	_, hasOuter := l.currentErrorHandler()

	if len(expr.Clauses) == 0 {
		if hasOuter {
			return nil, l.propagateError(IRValueUse{Value: errValue}, expr)
		}
		return nil, l.current.SetTerminator(&IRJump{Target: l.ensureErrorBlock().Label})
	}

	caseBlocks := make([]*IRBlock, len(expr.Clauses))
	for i := range expr.Clauses {
		caseBlocks[i] = l.newNamedBlock(fmt.Sprintf("rescue_case_%d", i))
	}
	noMatch := l.newNamedBlock("rescue_nomatch")
	if err := l.current.SetTerminator(&IRJump{Target: caseBlocks[0].Label}); err != nil {
		return nil, err
	}

	for i, clause := range expr.Clauses {
		if clause == nil {
			return nil, fmt.Errorf("compiler: rescue clause missing")
		}
		nextLabel := noMatch.Label
		if i+1 < len(caseBlocks) {
			nextLabel = caseBlocks[i+1].Label
		}
		l.current = caseBlocks[i]
		l.pushScope()
		l.declarePatternBindings(clause.Pattern, true)
		bindings, err := l.collectPatternBindings(clause.Pattern)
		if err != nil {
			return nil, err
		}
		patternErr := l.newValue("rescue_match_err", nil, clause)
		if err := l.current.Append(&IRDestructure{
			Pattern:  clause.Pattern,
			Value:    IRValueUse{Value: errValue},
			Error:    patternErr,
			Source:   clause,
			Bindings: bindings,
		}); err != nil {
			return nil, err
		}
		bodyBlock := l.newNamedBlock(fmt.Sprintf("rescue_body_%d", i))
		if clause.Guard != nil {
			guardBlock := l.newNamedBlock(fmt.Sprintf("rescue_guard_%d", i))
			if err := l.current.SetTerminator(&IRCheck{
				Error:    IRValueUse{Value: patternErr},
				OkLabel:  guardBlock.Label,
				ErrLabel: nextLabel,
			}); err != nil {
				return nil, err
			}
			l.current = guardBlock
			guardVal, err := l.lowerExpr(clause.Guard)
			if err != nil {
				return nil, err
			}
			if err := l.current.SetTerminator(&IRBranch{
				Condition:  guardVal,
				TrueLabel:  bodyBlock.Label,
				FalseLabel: nextLabel,
			}); err != nil {
				return nil, err
			}
		} else {
			if err := l.current.SetTerminator(&IRCheck{
				Error:    IRValueUse{Value: patternErr},
				OkLabel:  bodyBlock.Label,
				ErrLabel: nextLabel,
			}); err != nil {
				return nil, err
			}
		}

		l.current = bodyBlock
		bodyVal, err := l.lowerExpr(clause.Body)
		if err != nil {
			return nil, err
		}
		if l.current.Terminator == nil {
			if err := l.current.Append(&IRStore{Slot: resultSlot, Value: bodyVal, Source: clause.Body}); err != nil {
				return nil, err
			}
			if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
				return nil, err
			}
		}
		l.popScope()
	}

	l.current = noMatch
	if hasOuter {
		if err := l.propagateError(IRValueUse{Value: errValue}, expr); err != nil {
			return nil, err
		}
	} else if err := l.current.SetTerminator(&IRJump{Target: l.ensureErrorBlock().Label}); err != nil {
		return nil, err
	}

	l.current = joinBlock
	result := l.newValue("rescue_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerEnsure(expr *ast.EnsureExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: ensure is nil")
	}
	resultSlot := l.newSlot("ensure_result", l.typeOf(expr), false, expr)
	errSlot := l.newSlot("ensure_error", nil, false, expr)
	ensureBlock := l.newNamedBlock("ensure_block")
	okBlock := l.newNamedBlock("ensure_ok")
	errBlock := l.newNamedBlock("ensure_err")
	joinBlock := l.newNamedBlock("ensure_join")

	l.pushErrorHandler(ensureBlock.Label, errSlot)
	value, err := l.lowerExpr(expr.TryExpression)
	if err != nil {
		l.popErrorHandler()
		return nil, err
	}
	l.popErrorHandler()
	if l.current.Terminator == nil {
		if err := l.current.Append(&IRStore{Slot: resultSlot, Value: value, Source: expr.TryExpression}); err != nil {
			return nil, err
		}
		if err := l.current.Append(&IRStore{Slot: errSlot, Value: IRConst{Literal: ast.NewNilLiteral(), TypeRef: nil}, Source: expr}); err != nil {
			return nil, err
		}
		if err := l.current.SetTerminator(&IRJump{Target: ensureBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = ensureBlock
	if expr.EnsureBlock != nil {
		if _, err := l.lowerBlock(expr.EnsureBlock); err != nil {
			return nil, err
		}
	}
	errValue := l.newValue("ensure_err", nil, expr)
	if err := l.current.Append(&IRLoad{Dest: errValue, Slot: errSlot, Source: expr}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  okBlock.Label,
		ErrLabel: errBlock.Label,
	}); err != nil {
		return nil, err
	}

	l.current = okBlock
	if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
		return nil, err
	}

	l.current = errBlock
	if err := l.propagateError(IRValueUse{Value: errValue}, expr); err != nil {
		return nil, err
	}

	l.current = joinBlock
	result := l.newValue("ensure_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerWhileLoop(loop *ast.WhileLoop) (IRValueRef, error) {
	if loop == nil {
		return nil, fmt.Errorf("compiler: while loop is nil")
	}
	resultSlot := l.newSlot("while_result", l.typeOf(loop), false, loop)
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: IRVoid{}, Source: loop}); err != nil {
		return nil, err
	}
	condBlock := l.newNamedBlock("while_cond")
	bodyBlock := l.newNamedBlock("while_body")
	exitBlock := l.newNamedBlock("while_exit")

	if err := l.current.SetTerminator(&IRJump{Target: condBlock.Label}); err != nil {
		return nil, err
	}

	l.current = condBlock
	condValue, err := l.lowerExpr(loop.Condition)
	if err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRBranch{
		Condition:  condValue,
		TrueLabel:  bodyBlock.Label,
		FalseLabel: exitBlock.Label,
	}); err != nil {
		return nil, err
	}

	l.current = bodyBlock
	l.pushLoop(loopContext{breakLabel: exitBlock.Label, continueLabel: condBlock.Label, resultSlot: resultSlot})
	if _, err := l.lowerBlock(loop.Body); err != nil {
		return nil, err
	}
	l.popLoop()
	if l.current.Terminator == nil {
		if err := l.current.SetTerminator(&IRJump{Target: condBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = exitBlock
	result := l.newValue("while_result", l.typeOf(loop), loop)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: loop}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerLoopExpression(loop *ast.LoopExpression) (IRValueRef, error) {
	if loop == nil {
		return nil, fmt.Errorf("compiler: loop expression is nil")
	}
	bodyBlock := l.newNamedBlock("loop_body")
	exitBlock := l.newNamedBlock("loop_exit")
	resultSlot := l.newSlot("loop_result", l.typeOf(loop), false, loop)
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: IRVoid{}, Source: loop}); err != nil {
		return nil, err
	}

	if err := l.current.SetTerminator(&IRJump{Target: bodyBlock.Label}); err != nil {
		return nil, err
	}

	l.current = bodyBlock
	l.pushLoop(loopContext{breakLabel: exitBlock.Label, continueLabel: bodyBlock.Label, resultSlot: resultSlot})
	if _, err := l.lowerBlock(loop.Body); err != nil {
		return nil, err
	}
	l.popLoop()
	if l.current.Terminator == nil {
		if err := l.current.SetTerminator(&IRJump{Target: bodyBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = exitBlock
	result := l.newValue("loop_result", l.typeOf(loop), loop)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: loop}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerBreakpointExpression(expr *ast.BreakpointExpression) (IRValueRef, error) {
	if expr == nil || expr.Label == nil || expr.Label.Name == "" {
		return nil, fmt.Errorf("compiler: breakpoint expression missing label")
	}
	bodyBlock := l.newNamedBlock("breakpoint_body")
	exitBlock := l.newNamedBlock("breakpoint_exit")
	resultSlot := l.newSlot("breakpoint_result", l.typeOf(expr), false, expr)

	if err := l.current.SetTerminator(&IRJump{Target: bodyBlock.Label}); err != nil {
		return nil, err
	}

	l.current = bodyBlock
	l.pushBreakpoint(breakpointContext{
		label:      expr.Label.Name,
		breakLabel: exitBlock.Label,
		resultSlot: resultSlot,
	})
	l.pushScope()
	val, err := l.lowerBlock(expr.Body)
	l.popScope()
	l.popBreakpoint()
	if err != nil {
		return nil, err
	}
	if l.current.Terminator == nil {
		if err := l.current.Append(&IRStore{Slot: resultSlot, Value: val, Source: expr.Body}); err != nil {
			return nil, err
		}
		if err := l.current.SetTerminator(&IRJump{Target: exitBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = exitBlock
	result := l.newValue("breakpoint_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}
