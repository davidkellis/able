package compiler

import "bytes"

// isCanonicalPrimitiveStringLenBytesMethod identifies the primitive byte
// length operation. It intentionally excludes similarly named methods on
// StringBuilder, Grapheme, and user-defined nominal types.
func isCanonicalPrimitiveStringLenBytesMethod(method *methodInfo) bool {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return false
	}
	info := method.Info
	return info.Package == "able.text.string" && method.TargetName == "String" &&
		method.MethodName == "len_bytes" && method.ExpectsSelf &&
		info.ReturnType == "uint64" && len(info.Params) == 1 &&
		info.Params[0].GoType == "string"
}

func (g *generator) renderCanonicalPrimitiveStringLenBytesMethodBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	if info == nil || len(info.Params) != 1 {
		g.renderCompiledMethodBody(buf, info, lines, retExpr)
		return
	}
	self := info.Params[0].GoName
	g.renderCompiledMethodBodyWithPrefix(buf, info, []string{
		"if utf8.ValidString(" + self + ") && len(" + self + ") <= 2147483647 {",
		"\treturn uint64(len(" + self + ")), nil",
		"}",
	}, lines, retExpr)
}
