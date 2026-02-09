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

func (l *IRLowerer) lowerSpawn(expr *ast.SpawnExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: spawn is nil")
	}
	captures := l.captureSlots()
	body, err := l.lowerSpawnFunction(expr.Expression, captures)
	if err != nil {
		return nil, err
	}
	value := l.newValue("spawn", l.typeOf(expr), expr)
	errValue := l.newValue("spawn_err", l.typeOf(expr), expr)
	if err := l.current.Append(&IRSpawn{
		Value:    value,
		Error:    errValue,
		Body:     body,
		Captures: captures,
		Source:   expr,
	}); err != nil {
		return nil, err
	}
	okBlock := l.newNamedBlock("spawn_ok")
	errBlock := l.ensureErrorBlock()
	if handler, ok := l.currentErrorHandler(); ok {
		errBlock = l.blocks[handler.label]
		if errBlock == nil {
			return nil, fmt.Errorf("compiler: missing error handler block %s", handler.label)
		}
		if handler.slot != nil {
			if err := l.current.Append(&IRStore{Slot: handler.slot, Value: IRValueUse{Value: errValue}, Source: expr}); err != nil {
				return nil, err
			}
		}
	}
	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  okBlock.Label,
		ErrLabel: errBlock.Label,
	}); err != nil {
		return nil, err
	}
	l.current = okBlock
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) lowerAwait(expr *ast.AwaitExpression) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: await is nil")
	}
	value, err := l.lowerExpr(expr.Expression)
	if err != nil {
		return nil, err
	}
	return l.emitInvoke(IROpAwait, nil, []IRValueRef{value}, expr)
}

func (l *IRLowerer) lowerStringInterpolation(expr *ast.StringInterpolation) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: string interpolation is nil")
	}
	parts := make([]IRValueRef, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		val, err := l.lowerExpr(part)
		if err != nil {
			return nil, err
		}
		parts = append(parts, val)
	}
	value := l.newValue("interp", l.typeOf(expr), expr)
	if err := l.current.Append(&IRStringInterpolation{Dest: value, Parts: parts, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) lowerStructLiteral(expr *ast.StructLiteral) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: struct literal is nil")
	}
	fields := make([]IRStructField, 0, len(expr.Fields))
	for _, field := range expr.Fields {
		if field == nil {
			continue
		}
		val, err := l.lowerExpr(field.Value)
		if err != nil {
			return nil, err
		}
		name := ""
		if field.Name != nil {
			name = field.Name.Name
		}
		fields = append(fields, IRStructField{
			Name:        name,
			Value:       val,
			IsShorthand: field.IsShorthand,
		})
	}
	updates := make([]IRValueRef, 0, len(expr.FunctionalUpdateSources))
	for _, src := range expr.FunctionalUpdateSources {
		val, err := l.lowerExpr(src)
		if err != nil {
			return nil, err
		}
		updates = append(updates, val)
	}
	structName := ""
	if expr.StructType != nil {
		structName = expr.StructType.Name
	}
	value := l.newValue("struct", l.typeOf(expr), expr)
	if err := l.current.Append(&IRStructLiteral{
		Dest:          value,
		StructName:    structName,
		Positional:    expr.IsPositional,
		Fields:        fields,
		Updates:       updates,
		TypeArguments: expr.TypeArguments,
		Source:        expr,
	}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) lowerArrayLiteral(expr *ast.ArrayLiteral) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: array literal is nil")
	}
	args := make([]IRValueRef, 0, len(expr.Elements))
	for _, element := range expr.Elements {
		val, err := l.lowerExpr(element)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	value := l.newValue("array", l.typeOf(expr), expr)
	if err := l.current.Append(&IRArrayLiteral{Dest: value, Elements: args, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) lowerMapLiteral(expr *ast.MapLiteral) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: map literal is nil")
	}
	elements := make([]IRMapElement, 0, len(expr.Elements))
	for _, element := range expr.Elements {
		switch entry := element.(type) {
		case *ast.MapLiteralEntry:
			key, err := l.lowerExpr(entry.Key)
			if err != nil {
				return nil, err
			}
			val, err := l.lowerExpr(entry.Value)
			if err != nil {
				return nil, err
			}
			elements = append(elements, IRMapEntry{Key: key, Value: val})
		case *ast.MapLiteralSpread:
			spread, err := l.lowerExpr(entry.Expression)
			if err != nil {
				return nil, err
			}
			elements = append(elements, IRMapSpread{Value: spread})
		}
	}
	value := l.newValue("map", l.typeOf(expr), expr)
	if err := l.current.Append(&IRMapLiteral{Dest: value, Elements: elements, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) lowerIteratorLiteral(expr *ast.IteratorLiteral) (IRValueRef, error) {
	if expr == nil {
		return nil, fmt.Errorf("compiler: iterator literal is nil")
	}
	value := l.newValue("iter", l.typeOf(expr), expr)
	binding := "gen"
	if expr.Binding != nil && expr.Binding.Name != "" {
		binding = expr.Binding.Name
	}
	captures := l.captureSlots()
	body, err := l.lowerIteratorFunction(expr, binding, captures)
	if err != nil {
		return nil, err
	}
	if err := l.current.Append(&IRIteratorLiteral{
		Dest:        value,
		BindingName: binding,
		Body:        body,
		Captures:    captures,
		Source:      expr,
	}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) lowerForLoop(loop *ast.ForLoop) (IRValueRef, error) {
	if loop == nil {
		return nil, fmt.Errorf("compiler: for loop is nil")
	}
	iterable, err := l.lowerExpr(loop.Iterable)
	if err != nil {
		return nil, err
	}
	iterValue, err := l.emitInvoke(IROpIterator, nil, []IRValueRef{iterable}, loop)
	if err != nil {
		return nil, err
	}
	iterSlot := l.newSlot("for_iter", nil, false, loop)
	if err := l.current.Append(&IRStore{Slot: iterSlot, Value: iterValue, Source: loop}); err != nil {
		return nil, err
	}

	resultSlot := l.newSlot("for_result", l.typeOf(loop), false, loop)
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: IRVoid{}, Source: loop}); err != nil {
		return nil, err
	}

	nextBlock := l.newNamedBlock("for_next")
	bodyBlock := l.newNamedBlock("for_body")
	exitBlock := l.newNamedBlock("for_exit")
	checkBlock := l.newNamedBlock("for_check")

	if err := l.current.SetTerminator(&IRJump{Target: nextBlock.Label}); err != nil {
		return nil, err
	}

	l.current = nextBlock
	iterLoad := l.newValue("for_iter", nil, loop)
	if err := l.current.Append(&IRLoad{Dest: iterLoad, Slot: iterSlot, Source: loop}); err != nil {
		return nil, err
	}
	nextValue := l.newValue("for_value", nil, loop)
	doneValue := l.newValue("for_done", nil, loop)
	errValue := l.newValue("for_err", nil, loop)
	if err := l.current.Append(&IRIterNext{
		Iterator: IRValueUse{Value: iterLoad},
		Value:    nextValue,
		Done:     doneValue,
		Error:    errValue,
		Source:   loop,
	}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  checkBlock.Label,
		ErrLabel: l.ensureErrorBlock().Label,
	}); err != nil {
		return nil, err
	}

	l.current = checkBlock
	if err := l.current.SetTerminator(&IRBranch{
		Condition:  IRValueUse{Value: doneValue},
		TrueLabel:  exitBlock.Label,
		FalseLabel: bodyBlock.Label,
	}); err != nil {
		return nil, err
	}

	l.current = bodyBlock
	l.pushScope()
	l.declarePatternBindings(loop.Pattern, true)
	bindings, err := l.collectPatternBindings(loop.Pattern)
	if err != nil {
		return nil, err
	}
	assignErr := l.newValue("for_assign_err", nil, loop)
	if err := l.current.Append(&IRDestructure{
		Pattern:  loop.Pattern,
		Value:    IRValueUse{Value: nextValue},
		Error:    assignErr,
		Source:   loop,
		Bindings: bindings,
	}); err != nil {
		return nil, err
	}

	assignOk := l.newNamedBlock("for_assign_ok")
	assignFail := l.newNamedBlock("for_assign_fail")
	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: assignErr},
		OkLabel:  assignOk.Label,
		ErrLabel: assignFail.Label,
	}); err != nil {
		return nil, err
	}

	l.current = assignFail
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: IRValueUse{Value: assignErr}, Source: loop}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRJump{Target: exitBlock.Label}); err != nil {
		return nil, err
	}

	l.current = assignOk
	l.pushLoop(loopContext{breakLabel: exitBlock.Label, continueLabel: nextBlock.Label, resultSlot: resultSlot})
	if _, err := l.lowerBlock(loop.Body); err != nil {
		return nil, err
	}
	l.popLoop()
	l.popScope()
	if l.current.Terminator == nil {
		if err := l.current.SetTerminator(&IRJump{Target: nextBlock.Label}); err != nil {
			return nil, err
		}
	}

	l.current = exitBlock
	result := l.newValue("for_result", l.typeOf(loop), loop)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: loop}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) lowerBreak(stmt *ast.BreakStatement) error {
	if stmt == nil {
		return fmt.Errorf("compiler: break is nil")
	}
	if stmt.Label != nil && stmt.Label.Name != "" {
		ctx, ok := l.lookupBreakpoint(stmt.Label.Name)
		if !ok {
			return fmt.Errorf("compiler: unknown break label %s", stmt.Label.Name)
		}
		value := IRValueRef(IRVoid{})
		if stmt.Value != nil {
			val, err := l.lowerExpr(stmt.Value)
			if err != nil {
				return err
			}
			value = val
		}
		if ctx.resultSlot != nil {
			if err := l.current.Append(&IRStore{Slot: ctx.resultSlot, Value: value, Source: stmt}); err != nil {
				return err
			}
		}
		return l.current.SetTerminator(&IRJump{Target: ctx.breakLabel})
	}
	ctx, ok := l.currentLoop()
	if !ok {
		return fmt.Errorf("compiler: break outside loop")
	}
	if ctx.resultSlot != nil {
		value := IRValueRef(IRConst{Literal: ast.NewNilLiteral(), TypeRef: l.typeOf(stmt)})
		if stmt.Value != nil {
			val, err := l.lowerExpr(stmt.Value)
			if err != nil {
				return err
			}
			value = val
		}
		if err := l.current.Append(&IRStore{Slot: ctx.resultSlot, Value: value, Source: stmt}); err != nil {
			return err
		}
	}
	return l.current.SetTerminator(&IRJump{Target: ctx.breakLabel})
}

func (l *IRLowerer) lowerContinue(stmt *ast.ContinueStatement) error {
	if stmt == nil {
		return fmt.Errorf("compiler: continue is nil")
	}
	if stmt.Label != nil && stmt.Label.Name != "" {
		return fmt.Errorf("compiler: labeled continue not supported in IR lowering")
	}
	ctx, ok := l.currentLoop()
	if !ok {
		return fmt.Errorf("compiler: continue outside loop")
	}
	return l.current.SetTerminator(&IRJump{Target: ctx.continueLabel})
}

func (l *IRLowerer) lowerPatternAssignment(expr *ast.AssignmentExpression, pattern ast.Pattern, right IRValueRef) (IRValueRef, error) {
	l.declarePatternBindingsForAssignment(pattern, expr.Operator == ast.AssignmentDeclare)
	bindings, err := l.collectPatternBindings(pattern)
	if err != nil {
		return nil, err
	}
	errValue := l.newValue("assign_err", l.typeOf(expr), expr)
	if err := l.current.Append(&IRDestructure{
		Pattern:  pattern,
		Value:    right,
		Error:    errValue,
		Source:   expr,
		Bindings: bindings,
	}); err != nil {
		return nil, err
	}

	okBlock := l.newNamedBlock("assign_ok")
	errBlock := l.newNamedBlock("assign_err")
	joinBlock := l.newNamedBlock("assign_join")
	resultSlot := l.newSlot("assign_result", l.typeOf(expr), false, expr)

	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  okBlock.Label,
		ErrLabel: errBlock.Label,
	}); err != nil {
		return nil, err
	}

	l.current = okBlock
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: right, Source: expr}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
		return nil, err
	}

	l.current = errBlock
	if err := l.current.Append(&IRStore{Slot: resultSlot, Value: IRValueUse{Value: errValue}, Source: expr}); err != nil {
		return nil, err
	}
	if err := l.current.SetTerminator(&IRJump{Target: joinBlock.Label}); err != nil {
		return nil, err
	}

	l.current = joinBlock
	result := l.newValue("assign_result", l.typeOf(expr), expr)
	if err := l.current.Append(&IRLoad{Dest: result, Slot: resultSlot, Source: expr}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: result}, nil
}

func (l *IRLowerer) declarePatternBindingsForAssignment(pattern ast.Pattern, forceDeclare bool) {
	if pattern == nil {
		return
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		l.declareBindingNameForAssignment(p.Name, l.typeOf(p), forceDeclare, p)
	case *ast.TypedPattern:
		l.declarePatternBindingsForAssignment(p.Pattern, forceDeclare)
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil {
				l.declareBindingNameForAssignment(field.Binding.Name, l.typeOf(field.Binding), forceDeclare, field)
			}
			l.declarePatternBindingsForAssignment(field.Pattern, forceDeclare)
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			l.declarePatternBindingsForAssignment(elem, forceDeclare)
		}
		if p.RestPattern != nil {
			l.declarePatternBindingsForAssignment(p.RestPattern, forceDeclare)
		}
	}
}

func (l *IRLowerer) declareBindingNameForAssignment(name string, typ typechecker.Type, forceDeclare bool, source ast.Node) {
	if name == "" {
		return
	}
	if forceDeclare {
		l.declareSlot(name, typ, true, source)
		return
	}
	if l.lookupSlot(name) == nil {
		l.declareSlot(name, typ, true, source)
	}
}

func (l *IRLowerer) collectPatternBindings(pattern ast.Pattern) (map[ast.Node]*IRSlot, error) {
	if pattern == nil {
		return nil, nil
	}
	bindings := make(map[ast.Node]*IRSlot)
	var walk func(pat ast.Pattern) error
	walk = func(pat ast.Pattern) error {
		if pat == nil {
			return nil
		}
		switch p := pat.(type) {
		case *ast.Identifier:
			if p.Name == "" || p.Name == "_" {
				return nil
			}
			slot := l.lookupSlot(p.Name)
			if slot == nil {
				return fmt.Errorf("compiler: missing slot for pattern %s", p.Name)
			}
			bindings[p] = slot
		case *ast.TypedPattern:
			return walk(p.Pattern)
		case *ast.StructPattern:
			for _, field := range p.Fields {
				if field == nil {
					continue
				}
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					slot := l.lookupSlot(field.Binding.Name)
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for pattern %s", field.Binding.Name)
					}
					bindings[field] = slot
				}
				if err := walk(field.Pattern); err != nil {
					return err
				}
			}
		case *ast.ArrayPattern:
			for _, elem := range p.Elements {
				if err := walk(elem); err != nil {
					return err
				}
			}
			if p.RestPattern != nil {
				switch rest := p.RestPattern.(type) {
				case *ast.Identifier:
					if rest.Name == "" || rest.Name == "_" {
						return nil
					}
					slot := l.lookupSlot(rest.Name)
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for pattern %s", rest.Name)
					}
					bindings[rest] = slot
				case *ast.WildcardPattern:
					return nil
				default:
					if err := walk(rest); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := walk(pattern); err != nil {
		return nil, err
	}
	return bindings, nil
}

func (l *IRLowerer) lowerCall(call *ast.FunctionCall) (IRValueRef, error) {
	if call == nil {
		return nil, fmt.Errorf("compiler: call is nil")
	}
	callee, err := l.lowerExpr(call.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]IRValueRef, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		val, err := l.lowerExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	return l.emitInvoke(IROpCall, callee, args, call)
}

func (l *IRLowerer) emitCompute(op IROp, operator string, args []IRValueRef, source ast.Node) (IRValueRef, error) {
	value := l.newValue("tmp", l.typeOf(source), source)
	if err := l.current.Append(&IRCompute{Dest: value, Op: op, Operator: operator, Args: args, Source: source}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) emitInvoke(op IROp, callee IRValueRef, args []IRValueRef, source ast.Node) (IRValueRef, error) {
	value := l.newValue("tmp", l.typeOf(source), source)
	errValue := l.newValue("err", l.typeOf(source), source)
	if err := l.current.Append(&IRInvoke{
		Value:  value,
		Error:  errValue,
		Op:     op,
		Callee: callee,
		Args:   args,
		Source: source,
	}); err != nil {
		return nil, err
	}
	okBlock := l.newNamedBlock("invoke_ok")
	errBlock := l.ensureErrorBlock()
	if handler, ok := l.currentErrorHandler(); ok {
		errBlock = l.blocks[handler.label]
		if errBlock == nil {
			return nil, fmt.Errorf("compiler: missing error handler block %s", handler.label)
		}
		if handler.slot != nil {
			if err := l.current.Append(&IRStore{Slot: handler.slot, Value: IRValueUse{Value: errValue}, Source: source}); err != nil {
				return nil, err
			}
		}
	}
	if err := l.current.SetTerminator(&IRCheck{
		Error:    IRValueUse{Value: errValue},
		OkLabel:  okBlock.Label,
		ErrLabel: errBlock.Label,
	}); err != nil {
		return nil, err
	}
	l.current = okBlock
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) ensureErrorBlock() *IRBlock {
	if l.errorBlock != nil {
		return l.errorBlock
	}
	block := l.newNamedBlock("error")
	_ = block.SetTerminator(&IRUnreachable{})
	l.errorBlock = block
	return block
}

func (l *IRLowerer) loadIdentifier(id *ast.Identifier) (IRValueRef, error) {
	if id == nil {
		return nil, fmt.Errorf("compiler: identifier is nil")
	}
	slot := l.lookupSlot(id.Name)
	if slot == nil {
		return IRGlobal{Name: id.Name, TypeRef: l.typeOf(id)}, nil
	}
	value := l.newValue(id.Name, l.typeOf(id), id)
	if err := l.current.Append(&IRLoad{Dest: value, Slot: slot, Source: id}); err != nil {
		return nil, err
	}
	return IRValueUse{Value: value}, nil
}

func (l *IRLowerer) paramName(pattern ast.Pattern) (string, bool) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name == "" {
			return "", false
		}
		return p.Name, true
	case *ast.TypedPattern:
		return l.paramName(p.Pattern)
	default:
		return "", false
	}
}

func (l *IRLowerer) declarePatternBindings(pattern ast.Pattern, allowDeclare bool) {
	if pattern == nil {
		return
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name == "" {
			return
		}
		if allowDeclare {
			l.declareSlot(p.Name, l.typeOf(p), true, p)
		}
	case *ast.TypedPattern:
		l.declarePatternBindings(p.Pattern, allowDeclare)
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" && allowDeclare {
				l.declareSlot(field.Binding.Name, l.typeOf(field.Binding), true, field)
			}
			l.declarePatternBindings(field.Pattern, allowDeclare)
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			l.declarePatternBindings(elem, allowDeclare)
		}
		if p.RestPattern != nil {
			l.declarePatternBindings(p.RestPattern, allowDeclare)
		}
	}
}

func (l *IRLowerer) pushScope() {
	l.scopes = append(l.scopes, make(map[string]*IRSlot))
}

func (l *IRLowerer) popScope() {
	if len(l.scopes) == 0 {
		return
	}
	l.scopes = l.scopes[:len(l.scopes)-1]
}

func (l *IRLowerer) lookupSlot(name string) *IRSlot {
	for i := len(l.scopes) - 1; i >= 0; i-- {
		if slot, ok := l.scopes[i][name]; ok {
			return slot
		}
	}
	return nil
}

func (l *IRLowerer) declareSlot(name string, typ typechecker.Type, mutable bool, source ast.Node) *IRSlot {
	if name == "" {
		return nil
	}
	scope := l.scopes[len(l.scopes)-1]
	if slot, ok := scope[name]; ok {
		return slot
	}
	slot := &IRSlot{
		Name:    name,
		Type:    typ,
		Mutable: mutable,
		Source:  source,
	}
	scope[name] = slot
	l.function.Locals = append(l.function.Locals, slot)
	return slot
}

func (l *IRLowerer) newValue(name string, typ typechecker.Type, source ast.Node) *IRValue {
	l.nextValueID++
	return &IRValue{
		ID:     l.nextValueID,
		Name:   name,
		Type:   typ,
		Source: source,
	}
}

func (l *IRLowerer) newSlot(name string, typ typechecker.Type, mutable bool, source ast.Node) *IRSlot {
	return l.declareSlot(fmt.Sprintf("%s_%d", name, len(l.function.Locals)), typ, mutable, source)
}

func (l *IRLowerer) newNamedBlock(label string) *IRBlock {
	if _, ok := l.blocks[label]; ok {
		l.nextBlockID++
		label = fmt.Sprintf("%s_%d", label, l.nextBlockID)
	}
	block := &IRBlock{Label: label}
	l.blocks[label] = block
	return block
}

func (l *IRLowerer) pushLoop(ctx loopContext) {
	l.loopStack = append(l.loopStack, ctx)
}

func (l *IRLowerer) popLoop() {
	if len(l.loopStack) == 0 {
		return
	}
	l.loopStack = l.loopStack[:len(l.loopStack)-1]
}

func (l *IRLowerer) currentLoop() (loopContext, bool) {
	if len(l.loopStack) == 0 {
		return loopContext{}, false
	}
	return l.loopStack[len(l.loopStack)-1], true
}

func (l *IRLowerer) pushBreakpoint(ctx breakpointContext) {
	l.breakpoints = append(l.breakpoints, ctx)
}

func (l *IRLowerer) popBreakpoint() {
	if len(l.breakpoints) == 0 {
		return
	}
	l.breakpoints = l.breakpoints[:len(l.breakpoints)-1]
}

func (l *IRLowerer) lookupBreakpoint(label string) (breakpointContext, bool) {
	if label == "" {
		return breakpointContext{}, false
	}
	for i := len(l.breakpoints) - 1; i >= 0; i-- {
		if l.breakpoints[i].label == label {
			return l.breakpoints[i], true
		}
	}
	return breakpointContext{}, false
}

func (l *IRLowerer) pushErrorHandler(label string, slot *IRSlot) {
	l.errorStack = append(l.errorStack, errorContext{label: label, slot: slot})
}

func (l *IRLowerer) popErrorHandler() {
	if len(l.errorStack) == 0 {
		return
	}
	l.errorStack = l.errorStack[:len(l.errorStack)-1]
}

func (l *IRLowerer) currentErrorHandler() (errorContext, bool) {
	if len(l.errorStack) == 0 {
		return errorContext{}, false
	}
	return l.errorStack[len(l.errorStack)-1], true
}

func (l *IRLowerer) propagateError(errValue IRValueRef, source ast.Node) error {
	if handler, ok := l.currentErrorHandler(); ok {
		if handler.slot != nil {
			if err := l.current.Append(&IRStore{Slot: handler.slot, Value: errValue, Source: source}); err != nil {
				return err
			}
		}
		return l.current.SetTerminator(&IRJump{Target: handler.label})
	}
	return l.current.SetTerminator(&IRJump{Target: l.ensureErrorBlock().Label})
}

func (l *IRLowerer) typeOf(node ast.Node) typechecker.Type {
	if l.types == nil || node == nil {
		return nil
	}
	if typ, ok := l.types[node]; ok {
		return typ
	}
	return nil
}
