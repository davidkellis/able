package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeIntegerCastTargetKind(target ast.TypeExpression) (runtime.IntegerType, bool) {
	simple, ok := target.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return "", false
	}
	kind := runtime.IntegerType(simple.Name.Name)
	_, ok = lookupIntegerInfo(kind)
	return kind, ok
}

func bytecodeEmitIntegerDivCast(ctx *bytecodeLoweringContext, i *Interpreter, cast *ast.TypeCastExpression) (bool, error) {
	if cast == nil {
		return false, nil
	}
	targetKind, ok := bytecodeIntegerCastTargetKind(cast.TargetType)
	if !ok {
		return false, nil
	}
	bin, ok := cast.Expression.(*ast.BinaryExpression)
	if !ok || bin == nil || bin.Operator != "/" {
		return false, nil
	}
	if err := emitExpression(ctx, i, bin.Left); err != nil {
		return false, err
	}
	if err := emitExpression(ctx, i, bin.Right); err != nil {
		return false, err
	}
	ctx.emit(bytecodeInstruction{
		op:       bytecodeOpBinaryIntDivCast,
		operator: string(targetKind),
		node:     cast,
	})
	return true, nil
}
