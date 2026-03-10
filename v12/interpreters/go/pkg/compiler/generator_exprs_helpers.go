package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) typeMatches(expected, actual string) bool {
	if expected == "" {
		return true
	}
	if expected == actual {
		return true
	}
	// any accepts all types (any is Go's interface{}, all types satisfy it).
	if expected == "any" {
		return true
	}
	return false
}

func (g *generator) wrapLinesAsExpression(ctx *compileContext, lines []string, expr string, exprType string) (string, string, bool) {
	if len(lines) == 0 {
		return expr, exprType, true
	}
	if expr == "" || exprType == "" {
		ctx.setReason("missing expression")
		return "", "", false
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", exprType, strings.Join(lines, "; "), expr), exprType, true
}

func (g *generator) zeroValueExpr(goType string) (string, bool) {
	switch goType {
	case "struct{}":
		return "struct{}{}", true
	case "runtime.Value":
		return "runtime.NilValue{}", true
	case "any":
		return "nil", true
	case "bool":
		return "false", true
	case "string":
		return "\"\"", true
	case "rune":
		return "rune(0)", true
	case "float32":
		return "float32(0)", true
	case "float64":
		return "float64(0)", true
	case "int":
		return "int(0)", true
	case "uint":
		return "uint(0)", true
	case "int8":
		return "int8(0)", true
	case "int16":
		return "int16(0)", true
	case "int32":
		return "int32(0)", true
	case "int64":
		return "int64(0)", true
	case "uint8":
		return "uint8(0)", true
	case "uint16":
		return "uint16(0)", true
	case "uint32":
		return "uint32(0)", true
	case "uint64":
		return "uint64(0)", true
	}
	if g.typeCategory(goType) == "struct" {
		base := strings.TrimPrefix(goType, "*")
		if strings.HasPrefix(goType, "*") {
			return fmt.Sprintf("&%s{}", base), true
		}
		return fmt.Sprintf("%s{}", base), true
	}
	return "", false
}

func (g *generator) typedStringifyExpr(expr string, goType string) (string, bool) {
	switch goType {
	case "bool":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatBool(%s)", expr), true
	case "rune":
		return fmt.Sprintf("string(%s)", expr), true
	case "float32":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatFloat(float64(%s), 'f', -1, 32)", expr), true
	case "float64":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatFloat(%s, 'f', -1, 64)", expr), true
	case "int":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatInt(int64(%s), 10)", expr), true
	case "int8", "int16", "int32", "int64":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatInt(int64(%s), 10)", expr), true
	case "uint":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatUint(uint64(%s), 10)", expr), true
	case "uint8", "uint16", "uint32", "uint64":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatUint(uint64(%s), 10)", expr), true
	}
	return "", false
}

func (g *generator) interfaceArgExprLines(ctx *compileContext, argExpr string, ifaceType ast.TypeExpression, context string, genericNames map[string]struct{}) ([]string, string, bool) {
	if argExpr == "" {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	if g.typeExprHasGeneric(ifaceType, genericNames) {
		lines := []string{
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", argExpr, resultTemp, resultTemp, argExpr),
		}
		return lines, resultTemp, true
	}
	rendered, ok := g.renderTypeExpression(ifaceType)
	if !ok {
		return nil, "", false
	}
	expected := typeExpressionToString(ifaceType)
	if context == "" {
		context = "<call>"
	}
	valTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }"),
		fmt.Sprintf("%s, %s, %s := bridge.MatchType(__able_runtime, %s, %s)", valTemp, okTemp, errTemp, rendered, argExpr),
		fmt.Sprintf("__able_panic_on_error(%s)", errTemp),
		fmt.Sprintf("if !%s { panic(fmt.Errorf(\"type mismatch calling %s: expected %s\")) }", okTemp, context, expected),
		fmt.Sprintf("var %s runtime.Value", resultTemp),
		fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", valTemp, resultTemp, resultTemp, valTemp),
	}
	return lines, resultTemp, true
}

func (g *generator) interfaceReturnExprLines(ctx *compileContext, valueExpr string, ifaceType ast.TypeExpression, genericNames map[string]struct{}) ([]string, string, bool) {
	if valueExpr == "" {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	if g.typeExprHasGeneric(ifaceType, genericNames) {
		lines := []string{
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", valueExpr, resultTemp, resultTemp, valueExpr),
		}
		return lines, resultTemp, true
	}
	rendered, ok := g.renderTypeExpression(ifaceType)
	if !ok {
		return nil, "", false
	}
	expected := typeExpressionToString(ifaceType)
	valTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }"),
		fmt.Sprintf("%s, %s, %s := bridge.MatchType(__able_runtime, %s, %s)", valTemp, okTemp, errTemp, rendered, valueExpr),
		fmt.Sprintf("__able_panic_on_error(%s)", errTemp),
		fmt.Sprintf("if !%s { panic(fmt.Errorf(\"return type mismatch: expected %s\")) }", okTemp, expected),
		fmt.Sprintf("var %s runtime.Value", resultTemp),
		fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", valTemp, resultTemp, resultTemp, valTemp),
	}
	return lines, resultTemp, true
}
