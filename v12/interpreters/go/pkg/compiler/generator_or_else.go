package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compilePropagationExpression(ctx *compileContext, expr *ast.PropagationExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing propagation expression")
		return "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = valueType
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	if valueType != "runtime.Value" {
		if !g.typeMatches(resultType, valueType) {
			ctx.setReason("propagation type mismatch")
			return "", "", false
		}
		if len(valueLines) == 0 {
			return valueExpr, valueType, true
		}
		exprValue := fmt.Sprintf("func() %s { %s; return %s }()", valueType, strings.Join(valueLines, "; "), valueExpr)
		return exprValue, valueType, true
	}
	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	lines = append(lines, fmt.Sprintf("if __able_is_error(%s) { bridge.Raise(__able_error_value(%s)) }", valueTemp, valueTemp))
	resultExpr := valueTemp
	if resultType != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(valueTemp, resultType)
		if !ok {
			ctx.setReason("propagation type mismatch")
			return "", "", false
		}
		resultExpr = converted
	}
	exprValue := fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr)
	return exprValue, resultType, true
}

func (g *generator) compileOrElseExpression(ctx *compileContext, expr *ast.OrElseExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.Handler == nil {
		ctx.setReason("missing or-else expression")
		return "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return "", "", false
	}
	handlerCtx := ctx.child()
	bindingName := ""
	if expr.ErrorBinding != nil && expr.ErrorBinding.Name != "" {
		bindingName = expr.ErrorBinding.Name
		handlerCtx.locals[bindingName] = paramInfo{Name: bindingName, GoName: sanitizeIdent(bindingName), GoType: "runtime.Value"}
	}

	preferredType := expected
	if preferredType == "" && valueType != "runtime.Value" {
		preferredType = valueType
	}
	handlerLines, handlerExpr, handlerType, ok := g.compileBlockExpression(handlerCtx, expr.Handler, preferredType)
	if !ok {
		return "", "", false
	}

	resultType := preferredType
	if resultType == "" {
		if handlerType != "" {
			resultType = handlerType
		} else {
			resultType = valueType
		}
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}

	handlerResultExpr := handlerExpr
	switch {
	case handlerType == resultType:
	case handlerType == "runtime.Value" && resultType != "runtime.Value":
		converted, ok := g.expectRuntimeValueExpr(handlerExpr, resultType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		handlerResultExpr = converted
	case resultType == "runtime.Value" && handlerType != "runtime.Value":
		converted, ok := g.runtimeValueExpr(handlerExpr, handlerType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		handlerResultExpr = converted
	default:
		ctx.setReason("or-else type mismatch")
		return "", "", false
	}

	resultTemp := ctx.newTemp()
	failedTemp := ctx.newTemp()
	valueTemp := ctx.newTemp()
	failureTemp := ""
	errorTemp := ""
	if bindingName != "" {
		failureTemp = ctx.newTemp()
		errorTemp = ctx.newTemp()
	}

	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s bool", failedTemp),
	}
	if bindingName != "" {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value", failureTemp))
		lines = append(lines, fmt.Sprintf("var %s bool", errorTemp))
	}
	lines = append(lines, "func() {")
	if bindingName != "" {
		lines = append(lines, fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { if val, ok := recovered.(runtime.Value); ok { %s = val; %s = true; %s = true } else { panic(recovered) } } }()", failureTemp, failedTemp, errorTemp))
	} else {
		lines = append(lines, fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { if _, ok := recovered.(runtime.Value); ok { %s = true } else { panic(recovered) } } }()", failedTemp))
	}
	lines = append(lines, indentLines(valueLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s := %s", valueTemp, valueExpr))
	successExpr := valueTemp
	switch {
	case valueType == resultType:
	case valueType == "runtime.Value" && resultType != "runtime.Value":
		converted, ok := g.expectRuntimeValueExpr(valueTemp, resultType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		successExpr = converted
	case resultType == "runtime.Value" && valueType != "runtime.Value":
		converted, ok := g.runtimeValueExpr(valueTemp, valueType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		successExpr = converted
	default:
		ctx.setReason("or-else type mismatch")
		return "", "", false
	}
	if valueType == "runtime.Value" {
		if bindingName != "" {
			lines = append(lines, fmt.Sprintf("\tif __able_is_nil(%s) { %s = runtime.NilValue{}; %s = true; return }", valueTemp, failureTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("\tif __able_is_error(%s) { %s = %s; %s = true; %s = true; return }", valueTemp, failureTemp, valueTemp, failedTemp, errorTemp))
		} else {
			lines = append(lines, fmt.Sprintf("\tif __able_is_nil(%s) { %s = true; return }", valueTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("\tif __able_is_error(%s) { %s = true; return }", valueTemp, failedTemp))
		}
	}
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, successExpr))
	lines = append(lines, "}()")

	lines = append(lines, fmt.Sprintf("if %s {", failedTemp))
	if bindingName != "" {
		goName := sanitizeIdent(bindingName)
		lines = append(lines, fmt.Sprintf("\tvar %s runtime.Value", goName))
		lines = append(lines, fmt.Sprintf("\tif %s { %s = %s } else { %s = runtime.NilValue{} }", errorTemp, goName, failureTemp, goName))
		lines = append(lines, fmt.Sprintf("\t_ = %s", goName))
	}
	lines = append(lines, indentLines(handlerLines, 1)...)
	lines = append(lines, fmt.Sprintf("\treturn %s", handlerResultExpr))
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("return %s", resultTemp))

	exprValue := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return exprValue, resultType, true
}
