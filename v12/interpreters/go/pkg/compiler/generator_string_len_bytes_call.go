package compiler

import "fmt"

// canonicalPrimitiveStringLenBytesCallExpr keeps the valid native String path
// entirely in generated Go. Invalid or oversized values still enter the
// canonical method so its StringEncodingError and range semantics are
// preserved under the owning package environment.
func (g *generator) canonicalPrimitiveStringLenBytesCallExpr(
	ctx *compileContext,
	method *methodInfo,
	args []string,
	callTarget string,
) (string, bool) {
	if g == nil || ctx == nil || !isCanonicalPrimitiveStringLenBytesMethod(method) ||
		len(args) != 1 || callTarget == "" {
		return "", false
	}
	const valueName = "__able_string_len_bytes_value"
	fallback := fmt.Sprintf("%s(%s)", callTarget, g.compiledCallArgs(ctx, []string{valueName}))
	return fmt.Sprintf(
		"func(%s string) (uint64, *__ableControl) { if utf8.ValidString(%s) && len(%s) <= 2147483647 { return uint64(len(%s)), nil }; return %s }(%s)",
		valueName,
		valueName,
		valueName,
		valueName,
		fallback,
		args[0],
	), true
}
