package compiler

import "fmt"

func (g *generator) compiledAppendControlCallFrameLine(controlExpr string, callExpr string) string {
	return fmt.Sprintf("%s = __able_append_control_call_frame(%s, %s)", controlExpr, controlExpr, callExpr)
}

func (g *generator) compiledControlCheckWithCallFrameLines(ctx *compileContext, controlExpr string, callExpr string) ([]string, bool) {
	if g == nil || ctx == nil || controlExpr == "" {
		return nil, false
	}
	transferLines, ok := g.controlTransferLines(ctx, controlExpr)
	if !ok {
		return nil, false
	}
	lines := []string{fmt.Sprintf("if %s != nil {", controlExpr)}
	lines = append(lines, fmt.Sprintf("\t%s", g.compiledAppendControlCallFrameLine(controlExpr, callExpr)))
	lines = append(lines, indentLines(transferLines, 1)...)
	lines = append(lines, "}")
	return lines, true
}
