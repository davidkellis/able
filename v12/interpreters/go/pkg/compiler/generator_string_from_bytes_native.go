package compiler

import "bytes"

// isCanonicalPrimitiveStringFromBytesUncheckedMethod identifies the internal
// primitive String construction step. It intentionally excludes methods on
// nominal types, including StringBuilder, and preserves String.from_bytes'
// preceding UTF-8 validation.
func isCanonicalPrimitiveStringFromBytesUncheckedMethod(method *methodInfo) bool {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return false
	}
	info := method.Info
	return info.Package == "able.text.string" && method.TargetName == "String" &&
		method.MethodName == "from_bytes_unchecked" && !method.ExpectsSelf &&
		info.ReturnType == "string" && len(info.Params) == 1 &&
		info.Params[0].GoType == "*__able_array_u8"
}

func (g *generator) renderCanonicalPrimitiveStringFromBytesUncheckedMethodBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	if info == nil || len(info.Params) != 1 {
		g.renderCompiledMethodBody(buf, info, lines, retExpr)
		return
	}
	bytes := info.Params[0].GoName
	g.renderCompiledMethodBodyWithPrefix(buf, info, []string{
		"if " + bytes + " != nil {",
		"\treturn string(" + bytes + ".Elements), nil",
		"}",
	}, lines, retExpr)
}
