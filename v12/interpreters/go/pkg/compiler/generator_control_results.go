package compiler

import "fmt"

const compileControlModeNativeCall = "nativecall"
const compileControlModeErrorOnly = "erroronly"
const compileControlModeRuntimeValueError = "runtimevalueerror"

func (g *generator) controlTransferLines(ctx *compileContext, controlExpr string) ([]string, bool) {
	if g == nil || ctx == nil || controlExpr == "" {
		return nil, false
	}
	if ctx.controlCaptureVar != "" && ctx.controlCaptureLabel != "" {
		transfer := fmt.Sprintf("goto %s", ctx.controlCaptureLabel)
		if ctx.controlCaptureBreak {
			transfer = fmt.Sprintf("break %s", ctx.controlCaptureLabel)
		}
		return []string{
			fmt.Sprintf("%s = %s", ctx.controlCaptureVar, controlExpr),
			transfer,
		}, true
	}
	if ctx.controlMode == compileControlModeNativeCall {
		return []string{
			fmt.Sprintf("return nil, __able_control_to_error(__able_runtime, callCtx, %s)", controlExpr),
		}, true
	}
	if ctx.controlMode == compileControlModeErrorOnly {
		return []string{
			fmt.Sprintf("return __able_control_to_error(__able_runtime, nil, %s)", controlExpr),
		}, true
	}
	if ctx.controlMode == compileControlModeRuntimeValueError {
		return []string{
			fmt.Sprintf("return nil, __able_control_to_error(__able_runtime, nil, %s)", controlExpr),
		}, true
	}
	if ctx.returnType == "" {
		ctx.setReason("missing control return type")
		return nil, false
	}
	zeroExpr, ok := g.zeroValueExpr(ctx.returnType)
	if !ok {
		ctx.setReason("missing control return zero value")
		return nil, false
	}
	return []string{fmt.Sprintf("return %s, %s", zeroExpr, controlExpr)}, true
}

func (g *generator) controlCheckLines(ctx *compileContext, controlExpr string) ([]string, bool) {
	if g == nil || ctx == nil || controlExpr == "" {
		return nil, false
	}
	transferLines, ok := g.controlTransferLines(ctx, controlExpr)
	if !ok {
		return nil, false
	}
	lines := []string{fmt.Sprintf("if %s != nil {", controlExpr)}
	lines = append(lines, indentLines(transferLines, 1)...)
	lines = append(lines, "}")
	return lines, true
}

func (g *generator) runtimeErrorControlExpr(nodeExpr string, errExpr string) string {
	if nodeExpr == "" {
		nodeExpr = "nil"
	}
	return fmt.Sprintf("__able_runtime_error_control(%s, %s)", nodeExpr, errExpr)
}

func (g *generator) raiseControlExpr(nodeExpr string, valueExpr string) string {
	if nodeExpr == "" {
		nodeExpr = "nil"
	}
	return fmt.Sprintf("__able_raise_control(%s, %s)", nodeExpr, valueExpr)
}

func (g *generator) nilPropagationReturnLines(ctx *compileContext) ([]string, bool) {
	expr, ok := g.nilPropagationReturnExpr(ctx)
	if !ok {
		if ctx != nil {
			ctx.setReason("nil propagation requires nil-compatible return type")
		}
		return nil, false
	}
	return []string{fmt.Sprintf("return %s, nil", expr)}, true
}

func (g *generator) nilPropagationReturnExpr(ctx *compileContext) (string, bool) {
	if g == nil || ctx == nil || ctx.returnType == "" {
		return "", false
	}
	switch ctx.returnType {
	case "runtime.Value":
		return "runtime.NilValue{}", true
	case "any":
		return "nil", true
	}
	if _, nullable := g.nativeNullableValueInnerType(ctx.returnType); nullable {
		return "nil", true
	}
	if expr, ok := g.nativeUnionNilExpr(ctx.returnType); ok {
		return expr, true
	}
	if ctx.returnTypeExpr != nil && g.typeExprIncludesNilInPackage(ctx.packageName, ctx.returnTypeExpr) {
		if typedNil, ok := g.typedNilExpr(ctx.returnType); ok {
			return typedNil, true
		}
	}
	return "", false
}
