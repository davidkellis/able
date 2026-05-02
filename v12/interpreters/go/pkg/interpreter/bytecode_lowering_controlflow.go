package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func placeholderPlanForExpression(expr ast.Expression) (placeholderPlan, bool, error) {
	if expr == nil {
		return placeholderPlan{}, false, nil
	}
	switch n := expr.(type) {
	case *ast.AssignmentExpression:
		return placeholderPlan{}, false, nil
	case *ast.BinaryExpression:
		if n.Operator == "|>" || n.Operator == "|>>" {
			return placeholderPlan{}, false, nil
		}
	}
	plan, ok, err := analyzePlaceholderExpression(expr)
	if err != nil || !ok {
		return plan, ok, err
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
			return placeholderPlan{}, false, nil
		}
	}
	return plan, true, nil
}

func emitBlock(ctx *bytecodeLoweringContext, i *Interpreter, block *ast.BlockExpression) error {
	if block == nil {
		return bytecodeUnsupported("nil block")
	}
	ctx.enterScope()
	if len(block.Body) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
		ctx.exitScope()
		return nil
	}
	for idx, stmt := range block.Body {
		if stmt == nil {
			return bytecodeUnsupported("nil statement in block")
		}
		if err := emitStatement(ctx, i, stmt, idx == len(block.Body)-1); err != nil {
			return err
		}
	}
	ctx.exitScope()
	return nil
}

func emitIf(ctx *bytecodeLoweringContext, i *Interpreter, expr *ast.IfExpression) error {
	if expr == nil {
		return bytecodeUnsupported("nil if expression")
	}
	jumpToElse := -1
	if instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, expr.IfCondition); ok {
		jumpToElse = ctx.emit(instr)
	} else if instr, ok := bytecodeJumpIfFalseBoolSlotInstruction(ctx, expr.IfCondition); ok {
		jumpToElse = ctx.emit(instr)
	} else {
		if err := emitExpression(ctx, i, expr.IfCondition); err != nil {
			return err
		}
		jumpToElse = ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
	}
	if err := emitBlock(ctx, i, expr.IfBody); err != nil {
		return err
	}
	jumpToEnd := []int{ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})}
	ctx.patchJump(jumpToElse, len(ctx.instructions))

	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			return bytecodeUnsupported("nil elsif clause")
		}
		jumpToNext := -1
		if instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, clause.Condition); ok {
			jumpToNext = ctx.emit(instr)
		} else if instr, ok := bytecodeJumpIfFalseBoolSlotInstruction(ctx, clause.Condition); ok {
			jumpToNext = ctx.emit(instr)
		} else {
			if err := emitExpression(ctx, i, clause.Condition); err != nil {
				return err
			}
			jumpToNext = ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
		}
		if err := emitBlock(ctx, i, clause.Body); err != nil {
			return err
		}
		jumpToEnd = append(jumpToEnd, ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1}))
		ctx.patchJump(jumpToNext, len(ctx.instructions))
	}

	if expr.ElseBody != nil {
		if err := emitBlock(ctx, i, expr.ElseBody); err != nil {
			return err
		}
	} else {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
	}

	end := len(ctx.instructions)
	for _, idx := range jumpToEnd {
		ctx.patchJump(idx, end)
	}
	return nil
}

func emitIfStatement(ctx *bytecodeLoweringContext, i *Interpreter, expr *ast.IfExpression) error {
	if expr == nil {
		return bytecodeUnsupported("nil if expression")
	}
	if len(expr.ElseIfClauses) == 0 && expr.ElseBody == nil {
		if instr, ok := bytecodeReturnIfBinarySlotConstInstruction(ctx, expr.IfCondition, expr.IfBody); ok {
			ctx.emit(instr)
			return nil
		}
	}
	jumpToElse := -1
	if instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, expr.IfCondition); ok {
		jumpToElse = ctx.emit(instr)
	} else if instr, ok := bytecodeJumpIfFalseBoolSlotInstruction(ctx, expr.IfCondition); ok {
		jumpToElse = ctx.emit(instr)
	} else {
		if err := emitExpression(ctx, i, expr.IfCondition); err != nil {
			return err
		}
		jumpToElse = ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
	}
	if err := emitBlock(ctx, i, expr.IfBody); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	jumpToEnd := []int{ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})}
	ctx.patchJump(jumpToElse, len(ctx.instructions))

	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			return bytecodeUnsupported("nil elsif clause")
		}
		jumpToNext := -1
		if instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, clause.Condition); ok {
			jumpToNext = ctx.emit(instr)
		} else if instr, ok := bytecodeJumpIfFalseBoolSlotInstruction(ctx, clause.Condition); ok {
			jumpToNext = ctx.emit(instr)
		} else {
			if err := emitExpression(ctx, i, clause.Condition); err != nil {
				return err
			}
			jumpToNext = ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
		}
		if err := emitBlock(ctx, i, clause.Body); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
		jumpToEnd = append(jumpToEnd, ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1}))
		ctx.patchJump(jumpToNext, len(ctx.instructions))
	}

	if expr.ElseBody != nil {
		if err := emitBlock(ctx, i, expr.ElseBody); err != nil {
			return err
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	}

	end := len(ctx.instructions)
	for _, idx := range jumpToEnd {
		ctx.patchJump(idx, end)
	}
	return nil
}

func emitLoopExpression(ctx *bytecodeLoweringContext, i *Interpreter, loop *ast.LoopExpression) error {
	if loop == nil {
		return bytecodeUnsupported("nil loop expression")
	}
	if loop.Body == nil {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
		return nil
	}
	loopEnter := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopEnter, loopBreak: -1, loopContinue: -1})
	loopStart := len(ctx.instructions)
	ctx.pushLoop(loopStart)
	if err := emitBlock(ctx, i, loop.Body); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: loopStart})
	loopExit := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopExit})
	ctx.popLoop(loopExit)
	ctx.patchLoopTargets(loopEnter, loopExit, loopStart)
	return nil
}

func emitWhileLoop(ctx *bytecodeLoweringContext, i *Interpreter, loop *ast.WhileLoop) error {
	if loop == nil {
		return bytecodeUnsupported("nil while loop")
	}
	if loop.Condition == nil || loop.Body == nil {
		return bytecodeUnsupported("while loop missing condition/body")
	}
	loopEnter := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopEnter, loopBreak: -1, loopContinue: -1})
	loopStart := len(ctx.instructions)
	ctx.pushLoop(loopStart)
	jumpToNoBreak := -1
	if instr, ok := bytecodeJumpIfFalseBoolSlotInstruction(ctx, loop.Condition); ok {
		jumpToNoBreak = ctx.emit(instr)
	} else {
		if err := emitExpression(ctx, i, loop.Condition); err != nil {
			return err
		}
		jumpToNoBreak = ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})
	}
	if err := emitBlock(ctx, i, loop.Body); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: loopStart})
	noBreak := len(ctx.instructions)
	ctx.patchJump(jumpToNoBreak, noBreak)
	ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
	loopExit := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopExit})
	ctx.popLoop(loopExit)
	ctx.patchLoopTargets(loopEnter, loopExit, loopStart)
	return nil
}

func bytecodeJumpIfFalseBoolSlotInstruction(ctx *bytecodeLoweringContext, condition ast.Expression) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, false
	}
	ident, ok := condition.(*ast.Identifier)
	if !ok || ident == nil {
		return bytecodeInstruction{}, false
	}
	slot, found := ctx.lookupSlot(ident.Name)
	if !found || ctx.slotKind(slot) != bytecodeCellKindBool {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:       bytecodeOpJumpIfBoolSlotFalse,
		target:   -1,
		argCount: slot,
		name:     ident.Name,
		node:     ident,
	}, true
}

func emitForLoop(ctx *bytecodeLoweringContext, i *Interpreter, loop *ast.ForLoop) error {
	if loop == nil {
		return bytecodeUnsupported("nil for loop")
	}
	if loop.Iterable == nil || loop.Body == nil {
		return bytecodeUnsupported("for loop missing iterable/body")
	}
	if err := emitExpression(ctx, i, loop.Iterable); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpIterInit})
	loopEnter := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopEnter, loopBreak: -1, loopContinue: -1})
	loopStart := len(ctx.instructions)
	ctx.pushLoop(loopStart)

	ctx.emit(bytecodeInstruction{op: bytecodeOpIterNext})
	jumpToBody := ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})

	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.emit(bytecodeInstruction{op: bytecodeOpIterClose})
	ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.VoidValue{}})
	jumpToEnd := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})

	bodyStart := len(ctx.instructions)
	ctx.patchJump(jumpToBody, bodyStart)
	ctx.enterScope()
	if ctx.frameLayout != nil {
		if ident, ok := loop.Pattern.(*ast.Identifier); ok {
			slot := ctx.declareSlot(ident.Name)
			ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: ident.Name})
			ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpBindPattern, node: loop})
		}
	} else {
		ctx.emit(bytecodeInstruction{op: bytecodeOpBindPattern, node: loop})
	}
	if err := emitBlock(ctx, i, loop.Body); err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
	ctx.exitScope()
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: loopStart})

	breakCleanup := len(ctx.instructions)
	ctx.emit(bytecodeInstruction{op: bytecodeOpIterClose})
	loopExit := ctx.emit(bytecodeInstruction{op: bytecodeOpLoopExit})

	ctx.patchJump(jumpToEnd, loopExit)
	ctx.popLoopWithBreakTarget(loopExit, breakCleanup)
	ctx.patchLoopTargets(loopEnter, breakCleanup, loopStart)
	return nil
}

func emitBreakStatement(ctx *bytecodeLoweringContext, i *Interpreter, stmt *ast.BreakStatement) error {
	if stmt == nil {
		return bytecodeUnsupported("nil break statement")
	}
	if stmt.Label != nil {
		if stmt.Value != nil {
			if err := emitExpression(ctx, i, stmt.Value); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpBreakLabel, name: stmt.Label.Name, node: stmt})
		return nil
	}
	if len(ctx.loopStack) == 0 {
		if stmt.Value != nil {
			if err := emitExpression(ctx, i, stmt.Value); err != nil {
				return err
			}
		} else {
			ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
		}
		ctx.emit(bytecodeInstruction{op: bytecodeOpBreakSignal, node: stmt})
		return nil
	}
	if stmt.Value != nil {
		if err := emitExpression(ctx, i, stmt.Value); err != nil {
			return err
		}
	} else {
		ctx.emit(bytecodeInstruction{op: bytecodeOpConst, value: runtime.NilValue{}})
	}
	exitCount, err := ctx.loopExitCount()
	if err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope, argCount: exitCount})
	jumpIdx := ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1})
	ctx.appendBreakJump(jumpIdx)
	return nil
}

func emitContinueStatement(ctx *bytecodeLoweringContext, _ *Interpreter, stmt *ast.ContinueStatement) error {
	if stmt == nil {
		return bytecodeUnsupported("nil continue statement")
	}
	if stmt.Label != nil {
		return bytecodeUnsupported("labeled continue not supported")
	}
	if len(ctx.loopStack) == 0 {
		ctx.emit(bytecodeInstruction{op: bytecodeOpContinueSignal, node: stmt})
		return nil
	}
	exitCount, err := ctx.loopExitCount()
	if err != nil {
		return err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope, argCount: exitCount})
	ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: ctx.currentLoopStart()})
	return nil
}

func (ctx *bytecodeLoweringContext) enterScope() {
	if ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes {
		ctx.emit(bytecodeInstruction{op: bytecodeOpEnterScope})
	}
	ctx.scopeDepth++
	if ctx.frameLayout != nil {
		ctx.slotScopes = append(ctx.slotScopes, make(map[string]int))
	}
}

func (ctx *bytecodeLoweringContext) exitScope() {
	if ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes {
		ctx.emit(bytecodeInstruction{op: bytecodeOpExitScope})
	}
	if ctx.scopeDepth > 0 {
		ctx.scopeDepth--
	}
	if ctx.frameLayout != nil && len(ctx.slotScopes) > 1 {
		ctx.slotScopes = ctx.slotScopes[:len(ctx.slotScopes)-1]
	}
}

func (ctx *bytecodeLoweringContext) lookupSlot(name string) (int, bool) {
	for i := len(ctx.slotScopes) - 1; i >= 0; i-- {
		if slot, ok := ctx.slotScopes[i][name]; ok {
			return slot, true
		}
	}
	return 0, false
}

func (ctx *bytecodeLoweringContext) declareSlot(name string) int {
	return ctx.declareSlotWithKind(name, bytecodeCellKindValue)
}

func (ctx *bytecodeLoweringContext) declareSlotWithKind(name string, kind bytecodeCellKind) int {
	slot := ctx.nextSlot
	ctx.nextSlot++
	ctx.setSlotKind(slot, kind)
	if len(ctx.slotScopes) > 0 {
		ctx.slotScopes[len(ctx.slotScopes)-1][name] = slot
	}
	return slot
}

func (ctx *bytecodeLoweringContext) setSlotKind(slot int, kind bytecodeCellKind) {
	if slot < 0 {
		return
	}
	for len(ctx.slotKinds) <= slot {
		ctx.slotKinds = append(ctx.slotKinds, bytecodeCellKindValue)
	}
	ctx.slotKinds[slot] = kind
}

func (ctx *bytecodeLoweringContext) slotKind(slot int) bytecodeCellKind {
	if slot < 0 || slot >= len(ctx.slotKinds) {
		return bytecodeCellKindValue
	}
	return ctx.slotKinds[slot]
}

func (ctx *bytecodeLoweringContext) pushLoop(start int) {
	ctx.loopStack = append(ctx.loopStack, loopContext{
		start:      start,
		scopeDepth: ctx.scopeDepth,
	})
}

func (ctx *bytecodeLoweringContext) popLoop(loopEnd int) {
	ctx.popLoopWithBreakTarget(loopEnd, loopEnd)
}

func (ctx *bytecodeLoweringContext) popLoopWithBreakTarget(_ int, breakTarget int) {
	if len(ctx.loopStack) == 0 {
		return
	}
	loop := ctx.loopStack[len(ctx.loopStack)-1]
	ctx.loopStack = ctx.loopStack[:len(ctx.loopStack)-1]
	for _, idx := range loop.breakJumps {
		ctx.patchJump(idx, breakTarget)
	}
}

func (ctx *bytecodeLoweringContext) appendBreakJump(index int) {
	if len(ctx.loopStack) == 0 {
		return
	}
	last := len(ctx.loopStack) - 1
	ctx.loopStack[last].breakJumps = append(ctx.loopStack[last].breakJumps, index)
}

func (ctx *bytecodeLoweringContext) currentLoopStart() int {
	if len(ctx.loopStack) == 0 {
		return -1
	}
	return ctx.loopStack[len(ctx.loopStack)-1].start
}

func (ctx *bytecodeLoweringContext) loopExitCount() (int, error) {
	if len(ctx.loopStack) == 0 {
		return 0, bytecodeUnsupported("break/continue outside loop")
	}
	loop := ctx.loopStack[len(ctx.loopStack)-1]
	exitCount := ctx.scopeDepth - loop.scopeDepth
	if exitCount <= 0 {
		return 0, bytecodeUnsupported("loop scope mismatch")
	}
	return exitCount, nil
}
