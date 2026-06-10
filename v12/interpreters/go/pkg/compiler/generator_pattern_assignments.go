package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compilePatternAssignment(ctx *compileContext, assign *ast.AssignmentExpression, pattern ast.Pattern) ([]string, string, string, bool) {
	return g.compilePatternAssignmentMode(ctx, assign, pattern, false)
}

func (g *generator) compilePatternAssignmentMode(
	ctx *compileContext,
	assign *ast.AssignmentExpression,
	pattern ast.Pattern,
	discardResult bool,
) ([]string, string, string, bool) {
	if assign == nil {
		ctx.setReason("missing assignment")
		return nil, "", "", false
	}
	if assign.Operator != ast.AssignmentDeclare && assign.Operator != ast.AssignmentAssign {
		ctx.setReason("compound assignment not supported with patterns")
		return nil, "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", assign.Right)
	if !ok {
		return nil, "", "", false
	}
	mode := patternBindingMode{declare: assign.Operator == ast.AssignmentDeclare}
	if mode.declare {
		newNames := map[string]struct{}{}
		collectPatternBindingNames(pattern, newNames)
		if len(newNames) == 0 {
			ctx.setReason(":= requires new binding")
			return nil, "", "", false
		}
		filtered := map[string]struct{}{}
		for name := range newNames {
			if _, ok := ctx.lookupCurrent(name); !ok {
				filtered[name] = struct{}{}
			}
		}
		if len(filtered) == 0 {
			ctx.setReason(":= requires new binding")
			return nil, "", "", false
		}
		mode.newNames = filtered
	}

	if valueType == "runtime.Value" || valueType == "any" {
		valueTypeExpr, _ := g.inferLocalTypeExpr(ctx, assign.Right, valueType)
		previousExpectedTypeExpr := ctx.expectedTypeExpr
		if valueTypeExpr != nil {
			ctx.expectedTypeExpr = g.lowerNormalizedTypeExpr(ctx, valueTypeExpr)
		}
		defer func() {
			ctx.expectedTypeExpr = previousExpectedTypeExpr
		}()
		valConvLines, valueRuntime, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("pattern assignment value unsupported")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, valConvLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
		condLines, cond, ok := g.compileMatchPatternCondition(ctx, pattern, valueTemp, "runtime.Value")
		if !ok {
			return nil, "", "", false
		}
		bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern, valueTemp, "runtime.Value", mode)
		if !ok {
			return nil, "", "", false
		}
		declLines, assignLines := splitPatternBindingLines(bindLines)
		lines = append(lines, declLines...)
		lines = append(lines, condLines...)
		if discardResult {
			lines, ok = g.appendDiscardedPatternAssignment(ctx, lines, cond, assignLines)
			return lines, "", "", ok
		}
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s runtime.Value", resultTemp))
		if cond != "true" {
			lines = append(lines, fmt.Sprintf("if !(%s) { %s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"} } else {", cond, resultTemp))
			lines = append(lines, assignLines...)
			lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, valueTemp))
			lines = append(lines, "}")
		} else {
			lines = append(lines, assignLines...)
			lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, valueTemp))
		}
		return lines, resultTemp, "runtime.Value", true
	}

	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	condLines, cond, ok := g.compileMatchPatternCondition(ctx, pattern, valueTemp, valueType)
	if !ok {
		return nil, "", "", false
	}
	bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern, valueTemp, valueType, mode)
	if !ok {
		return nil, "", "", false
	}
	declLines, assignLines := splitPatternBindingLines(bindLines)
	lines = append(lines, declLines...)
	lines = append(lines, condLines...)
	if discardResult {
		lines, ok = g.appendDiscardedPatternAssignment(ctx, lines, cond, assignLines)
		return lines, "", "", ok
	}
	resultTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s runtime.Value", resultTemp))
	resultLines, resultExpr, ok := g.lowerRuntimeValue(ctx, valueTemp, valueType)
	if !ok {
		ctx.setReason("pattern assignment value unsupported")
		return nil, "", "", false
	}
	if cond != "true" {
		lines = append(lines, fmt.Sprintf("if !(%s) { %s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"} } else {", cond, resultTemp))
		lines = append(lines, assignLines...)
		lines = append(lines, resultLines...)
		lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, resultExpr))
		lines = append(lines, "}")
	} else {
		lines = append(lines, assignLines...)
		lines = append(lines, resultLines...)
		lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, resultExpr))
	}
	return lines, resultTemp, "runtime.Value", true
}

func (g *generator) appendDiscardedPatternAssignment(
	ctx *compileContext,
	lines []string,
	cond string,
	assignLines []string,
) ([]string, bool) {
	if cond == "true" {
		return append(lines, assignLines...), true
	}
	transferLines, ok := g.lowerControlTransfer(
		ctx,
		g.raiseControlExpr("nil", `runtime.ErrorValue{Message: "pattern assignment mismatch"}`),
	)
	if !ok {
		return nil, false
	}
	lines = append(lines, fmt.Sprintf("if !(%s) {", cond))
	lines = append(lines, indentLines(transferLines, 1)...)
	lines = append(lines, "} else {")
	lines = append(lines, indentLines(assignLines, 1)...)
	lines = append(lines, "}")
	return lines, true
}

func splitPatternBindingLines(lines []string) ([]string, []string) {
	if len(lines) == 0 {
		return nil, nil
	}
	decls := make([]string, 0, len(lines))
	assigns := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "var ") {
			if idx := strings.Index(trimmed, " = "); idx != -1 {
				decl := strings.TrimSpace(trimmed[:idx])
				expr := strings.TrimSpace(trimmed[idx+3:])
				fields := strings.Fields(decl)
				if len(fields) >= 2 {
					name := fields[1]
					decls = append(decls, decl)
					assigns = append(assigns, fmt.Sprintf("%s = %s", name, expr))
					continue
				}
			}
			decls = append(decls, line)
			continue
		}
		if strings.HasPrefix(trimmed, "_ = ") || strings.HasPrefix(trimmed, "_=") {
			decls = append(decls, line)
			continue
		}
		assigns = append(assigns, line)
	}
	return decls, assigns
}
