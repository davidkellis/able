package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

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
