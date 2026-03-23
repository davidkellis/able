package compiler

// lowerRuntimeValue is the canonical boundary-adapter entrypoint for converting
// a static native carrier into runtime.Value.
func (g *generator) lowerRuntimeValue(ctx *compileContext, expr string, goType string) ([]string, string, bool) {
	return g.runtimeValueLines(ctx, expr, goType)
}

// lowerExpectRuntimeValue is the canonical boundary-adapter entrypoint for
// converting runtime.Value into a static expected carrier.
func (g *generator) lowerExpectRuntimeValue(ctx *compileContext, valueExpr string, expected string) ([]string, string, bool) {
	return g.expectRuntimeValueExprLines(ctx, valueExpr, expected)
}

// lowerWrapUnion is the canonical boundary-adapter entrypoint for wrapping a
// static carrier into a native union carrier.
func (g *generator) lowerWrapUnion(ctx *compileContext, expected, actual, expr string) ([]string, string, bool) {
	return g.nativeUnionWrapLines(ctx, expected, actual, expr)
}

// lowerWrapInterface is the canonical boundary-adapter entrypoint for wrapping
// a static carrier into a native interface carrier.
func (g *generator) lowerWrapInterface(ctx *compileContext, expected string, actual string, expr string) ([]string, string, bool) {
	return g.nativeInterfaceWrapLines(ctx, expected, actual, expr)
}

// lowerWrapCallable is the canonical boundary-adapter entrypoint for wrapping
// a static carrier into a native callable carrier.
func (g *generator) lowerWrapCallable(ctx *compileContext, expected string, actual string, expr string) ([]string, string, bool) {
	return g.nativeCallableWrapLines(ctx, expected, actual, expr)
}

// lowerCoerceExpectedStaticExpr is the canonical shared boundary-coercion
// entrypoint for static expressions.
func (g *generator) lowerCoerceExpectedStaticExpr(ctx *compileContext, lines []string, expr string, actual string, expected string) ([]string, string, string, bool) {
	return g.coerceExpectedStaticExpr(ctx, lines, expr, actual, expected)
}
