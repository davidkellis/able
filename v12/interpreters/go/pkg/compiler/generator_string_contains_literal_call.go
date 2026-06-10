package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// canonicalPrimitiveStringContainsLiteralCallExpr uses the language guarantee
// that String literals are valid UTF-8. Invalid or unchecked receivers still
// enter the canonical method so its StringEncodingError behavior is unchanged.
func (g *generator) canonicalPrimitiveStringContainsLiteralCallExpr(
	ctx *compileContext,
	call *ast.FunctionCall,
	method *methodInfo,
	args []string,
	callTarget string,
) (string, bool) {
	if g == nil || ctx == nil || call == nil || !isCanonicalPrimitiveStringContainsMethod(method) ||
		len(args) != 2 || callTarget == "" || len(call.Arguments) != 1 {
		return "", false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil {
		return "", false
	}
	if _, ok := member.Object.(*ast.Identifier); !ok {
		return "", false
	}
	if _, ok := call.Arguments[0].(*ast.StringLiteral); !ok {
		return "", false
	}
	fallback := fmt.Sprintf("%s(%s)", callTarget, g.compiledCallArgs(ctx, args))
	return fmt.Sprintf(
		"func() (bool, *__ableControl) { if utf8.ValidString(%s) { return strings.Contains(%s, %s), nil }; return %s }()",
		args[0], args[0], args[1], fallback,
	), true
}
