package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticArrayForLoopInternal(
	ctx *compileContext,
	loop *ast.ForLoop,
	withResult bool,
	iterLines []string,
	iterExpr string,
	iterType string,
) ([]string, string, bool) {
	elementType := g.staticArrayElemGoType(iterType)
	if elementType == "" {
		ctx.setReason("for loop iterable unsupported")
		return nil, "", false
	}
	elementTemp := ctx.newTemp()
	loopLabelName := ctx.newTemp()
	resultTemp := ""
	if withResult {
		resultTemp = ctx.newTemp()
	}
	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	if withResult {
		bodyCtx.loopBreakValueTemp = resultTemp
	}
	newNames := map[string]struct{}{}
	collectPatternBindingNames(loop.Pattern, newNames)
	mode := patternBindingMode{declare: true, newNames: newNames}
	condLines, cond, ok := g.compileMatchPatternCondition(bodyCtx, loop.Pattern, elementTemp, elementType)
	if !ok {
		return nil, "", false
	}
	bindLines, ok := g.compileAssignmentPatternBindings(bodyCtx, loop.Pattern, elementTemp, elementType, mode)
	if !ok {
		return nil, "", false
	}
	if cond != "true" || len(condLines) > 0 {
		mismatchLine := fmt.Sprintf("break %s", loopLabelName)
		if withResult {
			mismatchLine = fmt.Sprintf("%s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"}; %s", resultTemp, mismatchLine)
		}
		var condPrefix []string
		condPrefix = append(condPrefix, condLines...)
		condPrefix = append(condPrefix, fmt.Sprintf("if !(%s) { %s }", cond, mismatchLine))
		bindLines = append(condPrefix, bindLines...)
	}
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, "", false
	}
	innerLines := append(bindLines, bodyLines...)
	iterTemp := ctx.newTemp()
	valuesTemp := ctx.newTemp()
	idxTemp := ctx.newTemp()
	lines := append([]string{}, iterLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", iterTemp, iterExpr))
	valuesExpr, ok := g.nativeArrayValuesExpr(iterTemp, iterType)
	if !ok {
		ctx.setReason("for loop iterable unsupported")
		return nil, "", false
	}
	if withResult {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.VoidValue{}", resultTemp))
	}
	if strings.HasPrefix(iterType, "*") {
		lines = append(lines, fmt.Sprintf("var %s []%s", valuesTemp, elementType))
		lines = append(lines, fmt.Sprintf("if %s != nil { %s = %s }", iterTemp, valuesTemp, valuesExpr))
	} else {
		lines = append(lines, fmt.Sprintf("%s := %s", valuesTemp, valuesExpr))
	}
	lines = append(lines,
		fmt.Sprintf("var %s %s", elementTemp, elementType),
		fmt.Sprintf("%s := 0", idxTemp),
	)
	forPrefix := "for {"
	if linesReferenceLabel(innerLines, loopLabelName) {
		forPrefix = fmt.Sprintf("%s: for {", loopLabelName)
	}
	lines = append(lines,
		forPrefix,
		fmt.Sprintf("if %s >= %s { break }", idxTemp, g.staticSliceLenExpr(valuesTemp)),
		fmt.Sprintf("%s = %s[%s]", elementTemp, valuesTemp, idxTemp),
		fmt.Sprintf("%s++", idxTemp),
	)
	lines = append(lines, indentLines(innerLines, 1)...)
	lines = append(lines, "}")
	return lines, resultTemp, true
}
