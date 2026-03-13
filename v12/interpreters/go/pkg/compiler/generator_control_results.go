package compiler

import "fmt"

const compileControlModeNativeCall = "nativecall"
const compileControlModeErrorOnly = "erroronly"

func (g *generator) controlTransferLines(ctx *compileContext, controlExpr string) ([]string, bool) {
	if g == nil || ctx == nil || controlExpr == "" {
		return nil, false
	}
	if ctx.controlCaptureVar != "" && ctx.controlCaptureLabel != "" {
		return []string{
			fmt.Sprintf("%s = %s", ctx.controlCaptureVar, controlExpr),
			fmt.Sprintf("goto %s", ctx.controlCaptureLabel),
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
