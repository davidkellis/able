package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileIteratorLiteral(ctx *compileContext, expr *ast.IteratorLiteral, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing iterator literal")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" && !g.canCoerceStaticExpr(expected, "runtime.Value") {
		ctx.setReason("iterator literal type mismatch")
		return "", "", false
	}
	binding := "gen"
	if expr.Binding != nil && expr.Binding.Name != "" {
		binding = expr.Binding.Name
	}
	genParam := "__able_gen"
	bodyCtx := ctx.child()
	bodyCtx.controlMode = compileControlModeErrorOnly
	bodyCtx.locals[binding] = paramInfo{Name: binding, GoName: genParam, GoType: "*__able_generator"}
	if binding != "gen" {
		bodyCtx.locals["gen"] = paramInfo{Name: "gen", GoName: genParam, GoType: "*__able_generator"}
	}
	lines := make([]string, 0, len(expr.Body)+1)
	for _, stmt := range expr.Body {
		stmtLines, ok := g.compileStatement(bodyCtx, stmt)
		if !ok {
			return "", "", false
		}
		lines = append(lines, stmtLines...)
	}
	lines = append(lines, "return nil")
	run := fmt.Sprintf("func(%s *__able_generator) error { %s }", genParam, strings.Join(lines, "; "))
	g.needsIterator = true
	return fmt.Sprintf("__able_new_iterator(%s)", run), "runtime.Value", true
}

func (g *generator) compileYieldStatement(ctx *compileContext, stmt *ast.YieldStatement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing yield statement")
		return nil, false
	}
	genParam, ok := ctx.lookup("gen")
	if !ok {
		ctx.setReason("yield may only appear inside iterator literal")
		return nil, false
	}
	genExpr := genParam.GoName
	if genExpr == "" {
		ctx.setReason("yield generator missing")
		return nil, false
	}
	genValue := genExpr
	var genConvLines []string
	if genParam.GoType == "*__able_generator" {
		lines := []string{}
		valueExpr := "runtime.NilValue{}"
		if stmt.Expression != nil {
			expr, exprType, ok := g.compileExpr(ctx, stmt.Expression, "")
			if !ok {
				return nil, false
			}
			if exprType != "runtime.Value" {
				convLines, converted, ok := g.runtimeValueLines(ctx, expr, exprType)
				if !ok {
					ctx.setReason("yield argument unsupported")
					return nil, false
				}
				lines = append(lines, convLines...)
				valueExpr = converted
			} else {
				valueExpr = expr
			}
		}
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := %s.emit(%s)", errTemp, genExpr, valueExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		)
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, false
		}
		lines = append(lines, controlLines...)
		return lines, true
	}
	if genParam.GoType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, genExpr, genParam.GoType)
		if !ok {
			ctx.setReason("yield generator unsupported")
			return nil, false
		}
		genConvLines = convLines
		genValue = converted
	}
	lines := append([]string{}, genConvLines...)
	args := []string{}
	if stmt.Expression != nil {
		expr, _, ok := g.compileExpr(ctx, stmt.Expression, "runtime.Value")
		if !ok {
			return nil, false
		}
		argTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", argTemp, expr))
		args = append(args, argTemp)
	}
	argList := "nil"
	if len(args) > 0 {
		argList = "[]runtime.Value{" + strings.Join(args, ", ") + "}"
	}
	var callOK bool
	lines, _, callOK = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_method_call(%s, %q, %s)", genValue, "yield", argList))
	if !callOK {
		return nil, false
	}
	return lines, true
}
