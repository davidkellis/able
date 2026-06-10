package compiler

import "bytes"

// isCanonicalPrimitiveStringContainsMethod identifies the language primitive,
// not an application method with the same spelling. String is an allowed
// primitive lowering boundary; nominal user types remain on shared lowering.
func isCanonicalPrimitiveStringContainsMethod(method *methodInfo) bool {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return false
	}
	info := method.Info
	if info.Package != "able.text.string" || method.TargetName != "String" ||
		method.MethodName != "contains" || !method.ExpectsSelf ||
		info.ReturnType != "bool" || len(info.Params) != 2 {
		return false
	}
	return info.Params[0].GoType == "string" && info.Params[1].GoType == "string"
}

func (g *generator) renderCanonicalPrimitiveStringContainsMethodBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	if info == nil || len(info.Params) != 2 {
		g.renderCompiledMethodBody(buf, info, lines, retExpr)
		return
	}
	self := info.Params[0].GoName
	needle := info.Params[1].GoName
	g.renderCompiledMethodBodyWithPrefix(buf, info, []string{
		"if utf8.ValidString(" + self + ") && utf8.ValidString(" + needle + ") {",
		"\treturn strings.Contains(" + self + ", " + needle + "), nil",
		"}",
	}, lines, retExpr)
}
