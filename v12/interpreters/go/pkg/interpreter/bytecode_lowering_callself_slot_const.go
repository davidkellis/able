package interpreter

import "able/interpreter-go/pkg/ast"

func frameLayoutAllowsSelfCallIntSubSlotConst(layout *bytecodeFrameLayout) bool {
	if layout == nil || !layout.selfCallOneArgFast {
		return false
	}
	if layout.firstParamType == nil {
		return true
	}
	switch layout.firstParamSimple {
	case "Int",
		"i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128",
		"isize", "usize":
		return true
	default:
		return false
	}
}

func bytecodeSelfCallSlotConstInstruction(ctx *bytecodeLoweringContext, call *ast.FunctionCall) (bytecodeInstruction, bool) {
	if ctx == nil || call == nil || ctx.selfCallSlot < 0 || len(call.Arguments) != 1 || len(call.TypeArguments) > 0 {
		return bytecodeInstruction{}, false
	}
	if !frameLayoutAllowsSelfCallIntSubSlotConst(ctx.frameLayout) {
		return bytecodeInstruction{}, false
	}
	bin, ok := call.Arguments[0].(*ast.BinaryExpression)
	if !ok || bin == nil {
		return bytecodeInstruction{}, false
	}
	argInstr, ok := bytecodeBinarySlotConstInstruction(ctx, bin)
	if !ok || argInstr.op != bytecodeOpBinaryIntSubSlotConst {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:              bytecodeOpCallSelfIntSubSlotConst,
		target:          ctx.selfCallSlot,
		argCount:        argInstr.target,
		value:           argInstr.value,
		intImmediate:    argInstr.intImmediate,
		intImmediateRaw: argInstr.intImmediateRaw,
		hasIntImmediate: argInstr.hasIntImmediate,
		hasIntRaw:       argInstr.hasIntRaw,
		node:            call,
	}, true
}
