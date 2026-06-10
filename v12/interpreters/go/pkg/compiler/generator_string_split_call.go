package compiler

import "fmt"

// isCanonicalPrimitiveStringSplitMethod identifies the language String
// primitive only. User-defined methods and non-primitive nominal types remain
// on the shared method lowering path.
func isCanonicalPrimitiveStringSplitMethod(method *methodInfo) bool {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return false
	}
	info := method.Info
	return info.Package == "able.text.string" && method.TargetName == "String" &&
		method.MethodName == "split" && method.ExpectsSelf &&
		info.ReturnType == "*__able_array_String" && len(info.Params) == 2 &&
		info.Params[0].GoType == "string" && info.Params[1].GoType == "string"
}

// canonicalPrimitiveStringSplitCallExpr keeps proven static String calls on
// native carriers. The canonical method remains the fallback for invalid UTF-8
// and for dynamic calls, preserving its diagnostic and package semantics.
func (g *generator) canonicalPrimitiveStringSplitCallExpr(
	ctx *compileContext,
	method *methodInfo,
	args []string,
	callTarget string,
) (string, bool) {
	if g == nil || ctx == nil || !isCanonicalPrimitiveStringSplitMethod(method) ||
		len(args) != 2 || callTarget == "" {
		return "", false
	}
	const valueName = "__able_string_split_value"
	const delimiterName = "__able_string_split_delimiter"
	fallback := fmt.Sprintf(
		"%s(%s)",
		callTarget,
		g.compiledCallArgs(ctx, []string{valueName, delimiterName}),
	)
	return fmt.Sprintf(
		"func(%s string, %s string) (*__able_array_String, *__ableControl) { if utf8.ValidString(%s) && utf8.ValidString(%s) { return &__able_array_String{Elements: strings.Split(%s, %s)}, nil }; return %s }(%s, %s)",
		valueName,
		delimiterName,
		valueName,
		delimiterName,
		valueName,
		delimiterName,
		fallback,
		args[0],
		args[1],
	), true
}
