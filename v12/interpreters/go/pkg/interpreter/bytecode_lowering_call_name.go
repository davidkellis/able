package interpreter

import "able/interpreter-go/pkg/ast"

func bytecodeCallNameSlotArgsInstruction(ctx *bytecodeLoweringContext, name string, call *ast.FunctionCall) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || call == nil || name == "" {
		return bytecodeInstruction{}, false
	}
	if len(call.Arguments) == 0 || len(call.Arguments) > 3 {
		return bytecodeInstruction{}, false
	}
	slots := [3]int{-1, -1, -1}
	for idx, arg := range call.Arguments {
		ident, ok := arg.(*ast.Identifier)
		if !ok || ident == nil {
			return bytecodeInstruction{}, false
		}
		slot, found := ctx.lookupSlot(ident.Name)
		if !found {
			return bytecodeInstruction{}, false
		}
		slots[idx] = slot
	}
	return bytecodeInstruction{
		op:           bytecodeOpCallName,
		name:         name,
		nameSimple:   bytecodeSimpleLookupName(name),
		argCount:     len(call.Arguments),
		target:       slots[0],
		loopBreak:    slots[1],
		loopContinue: slots[2],
		node:         call,
		slotArgs:     true,
	}, true
}
