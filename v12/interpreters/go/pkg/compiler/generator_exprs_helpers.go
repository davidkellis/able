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
	return expected == actual
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

func (g *generator) interfaceArgExpr(argExpr string, ifaceType ast.TypeExpression, context string, genericNames map[string]struct{}) (string, bool) {
	if argExpr == "" {
		return "", false
	}
	if g.typeExprHasGeneric(ifaceType, genericNames) {
		return fmt.Sprintf("func() runtime.Value { if %s == nil { return runtime.NilValue{} }; return %s }()", argExpr, argExpr), true
	}
	rendered, ok := g.renderTypeExpression(ifaceType)
	if !ok {
		return "", false
	}
	expected := typeExpressionToString(ifaceType)
	if context == "" {
		context = "<call>"
	}
	return fmt.Sprintf("func() runtime.Value { if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }; val, ok, err := bridge.MatchType(__able_runtime, %s, %s); __able_panic_on_error(err); if !ok { panic(fmt.Errorf(\"type mismatch calling %s: expected %s\")) }; if val == nil { return runtime.NilValue{} }; return val }()", rendered, argExpr, context, expected), true
}

func (g *generator) interfaceReturnExpr(valueExpr string, ifaceType ast.TypeExpression, genericNames map[string]struct{}) (string, bool) {
	if valueExpr == "" {
		return "", false
	}
	if g.typeExprHasGeneric(ifaceType, genericNames) {
		return fmt.Sprintf("func() runtime.Value { if %s == nil { return runtime.NilValue{} }; return %s }()", valueExpr, valueExpr), true
	}
	rendered, ok := g.renderTypeExpression(ifaceType)
	if !ok {
		return "", false
	}
	expected := typeExpressionToString(ifaceType)
	return fmt.Sprintf("func() runtime.Value { if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }; val, ok, err := bridge.MatchType(__able_runtime, %s, %s); __able_panic_on_error(err); if !ok { panic(fmt.Errorf(\"return type mismatch: expected %s\")) }; if val == nil { return runtime.NilValue{} }; return val }()", rendered, valueExpr, expected), true
}
