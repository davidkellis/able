package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeArrayIndexGetSlotInstruction(ctx *bytecodeLoweringContext, expr *ast.IndexExpression) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || expr == nil {
		return bytecodeInstruction{}, false
	}
	objIdent, ok := expr.Object.(*ast.Identifier)
	if !ok || objIdent == nil {
		return bytecodeInstruction{}, false
	}
	idxIdent, ok := expr.Index.(*ast.Identifier)
	if !ok || idxIdent == nil {
		return bytecodeInstruction{}, false
	}
	objSlot, ok := ctx.lookupSlot(objIdent.Name)
	if !ok {
		return bytecodeInstruction{}, false
	}
	idxSlot, ok := ctx.lookupSlot(idxIdent.Name)
	if !ok {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  objSlot,
		loopBreak: idxSlot,
		node:      expr,
	}, true
}

func bytecodeArrayIndexSetSlotInstruction(ctx *bytecodeLoweringContext, expr *ast.AssignmentExpression, indexExpr *ast.IndexExpression) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || expr == nil || indexExpr == nil {
		return bytecodeInstruction{}, false
	}
	if expr.Operator != ast.AssignmentAssign {
		return bytecodeInstruction{}, false
	}
	objIdent, ok := indexExpr.Object.(*ast.Identifier)
	if !ok || objIdent == nil {
		return bytecodeInstruction{}, false
	}
	idxIdent, ok := indexExpr.Index.(*ast.Identifier)
	if !ok || idxIdent == nil {
		return bytecodeInstruction{}, false
	}
	objSlot, ok := ctx.lookupSlot(objIdent.Name)
	if !ok {
		return bytecodeInstruction{}, false
	}
	idxSlot, ok := ctx.lookupSlot(idxIdent.Name)
	if !ok {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:        bytecodeOpArrayIndexSetSlot,
		argCount:  objSlot,
		loopBreak: idxSlot,
		node:      expr,
	}, true
}

func bytecodeIntegerTypeToken(suffix runtime.IntegerType) uint16 {
	switch suffix {
	case runtime.IntegerI8:
		return bytecodeIndexTypeI8
	case runtime.IntegerI16:
		return bytecodeIndexTypeI16
	case runtime.IntegerI32:
		return bytecodeIndexTypeI32
	case runtime.IntegerI64:
		return bytecodeIndexTypeI64
	case runtime.IntegerI128:
		return bytecodeIndexTypeI128
	case runtime.IntegerU8:
		return bytecodeIndexTypeU8
	case runtime.IntegerU16:
		return bytecodeIndexTypeU16
	case runtime.IntegerU32:
		return bytecodeIndexTypeU32
	case runtime.IntegerU64:
		return bytecodeIndexTypeU64
	case runtime.IntegerU128:
		return bytecodeIndexTypeU128
	case runtime.IntegerIsize:
		return bytecodeIndexTypeIsize
	case runtime.IntegerUsize:
		return bytecodeIndexTypeUsize
	default:
		return bytecodeIndexTypeUnknown
	}
}

func bytecodeFloatTypeToken(suffix runtime.FloatType) uint16 {
	switch suffix {
	case runtime.FloatF32:
		return bytecodeIndexTypeF32
	case runtime.FloatF64:
		return bytecodeIndexTypeF64
	default:
		return bytecodeIndexTypeUnknown
	}
}

func bytecodeIndexTypeTokenFromTypeName(name string) (uint16, bool) {
	switch name {
	case string(runtime.IntegerI8):
		return bytecodeIndexTypeI8, true
	case string(runtime.IntegerI16):
		return bytecodeIndexTypeI16, true
	case string(runtime.IntegerI32):
		return bytecodeIndexTypeI32, true
	case string(runtime.IntegerI64):
		return bytecodeIndexTypeI64, true
	case string(runtime.IntegerI128):
		return bytecodeIndexTypeI128, true
	case string(runtime.IntegerU8):
		return bytecodeIndexTypeU8, true
	case string(runtime.IntegerU16):
		return bytecodeIndexTypeU16, true
	case string(runtime.IntegerU32):
		return bytecodeIndexTypeU32, true
	case string(runtime.IntegerU64):
		return bytecodeIndexTypeU64, true
	case string(runtime.IntegerU128):
		return bytecodeIndexTypeU128, true
	case string(runtime.IntegerIsize):
		return bytecodeIndexTypeIsize, true
	case string(runtime.IntegerUsize):
		return bytecodeIndexTypeUsize, true
	case string(runtime.FloatF32):
		return bytecodeIndexTypeF32, true
	case string(runtime.FloatF64):
		return bytecodeIndexTypeF64, true
	case "String":
		return bytecodeIndexTypeString, true
	case "bool":
		return bytecodeIndexTypeBool, true
	case "char":
		return bytecodeIndexTypeChar, true
	case "nil":
		return bytecodeIndexTypeNil, true
	case "void":
		return bytecodeIndexTypeVoid, true
	default:
		return bytecodeIndexTypeUnknown, false
	}
}

func bytecodeIndexValueTypeToken(value runtime.Value) (uint16, bool) {
	normalized := unwrapInterfaceValue(value)
	if kind, _, ok := bytecodeRawIntegerValueInfo(normalized); ok {
		token := bytecodeIntegerTypeToken(kind)
		return token, token != bytecodeIndexTypeUnknown
	}
	if _, kind, ok := bytecodeDirectRawFloatValue(normalized); ok {
		token := bytecodeFloatTypeToken(kind)
		return token, token != bytecodeIndexTypeUnknown
	}
	switch v := normalized.(type) {
	case runtime.IntegerValue:
		token := bytecodeIntegerTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case *runtime.IntegerValue:
		if v == nil {
			return bytecodeIndexTypeUnknown, false
		}
		token := bytecodeIntegerTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case runtime.FloatValue:
		token := bytecodeFloatTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case *runtime.FloatValue:
		if v == nil {
			return bytecodeIndexTypeUnknown, false
		}
		token := bytecodeFloatTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case runtime.StringValue, *runtime.StringValue:
		return bytecodeIndexTypeString, true
	case runtime.BoolValue, *runtime.BoolValue:
		return bytecodeIndexTypeBool, true
	case runtime.CharValue, *runtime.CharValue:
		return bytecodeIndexTypeChar, true
	case runtime.NilValue:
		return bytecodeIndexTypeNil, true
	case runtime.VoidValue:
		return bytecodeIndexTypeVoid, true
	default:
		return bytecodeIndexTypeUnknown, false
	}
}

func bytecodeArrayElementTypeTokenFromValues(values []runtime.Value) (uint16, bool) {
	if len(values) == 0 {
		return bytecodeIndexTypeUnknown, true
	}
	return bytecodeIndexValueTypeToken(values[0])
}

func bytecodeArrayElementTypeToken(arr *runtime.ArrayValue) (uint16, bool) {
	if arr == nil {
		return bytecodeIndexTypeUnknown, false
	}
	if arr.State != nil && arr.State.ElementTypeTokenKnown {
		return arr.State.ElementTypeToken, true
	}
	for _, handle := range []int64{arr.Handle, arr.TrackedHandle} {
		if handle == 0 {
			continue
		}
		if typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle); err == nil && ok {
			return bytecodeIndexTypeTokenFromTypeName(typeName)
		}
	}
	return bytecodeArrayElementTypeTokenFromValues(arr.Elements)
}
