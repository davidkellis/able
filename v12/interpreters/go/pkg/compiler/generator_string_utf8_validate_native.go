package compiler

import "bytes"

// isCanonicalPrimitiveStringUTF8ValidateFunction identifies the internal
// validator used at String byte boundaries. It is a primitive String rule,
// not a lowering for a nominal container or application function.
func isCanonicalPrimitiveStringUTF8ValidateFunction(info *functionInfo) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	return info.Package == "able.text.string" && info.Name == "utf8_validate" &&
		info.ReturnType == "*StringEncodingError" &&
		len(info.Params) == 1 && info.Params[0].GoType == "*__able_array_u8"
}

func (g *generator) renderCanonicalPrimitiveStringUTF8ValidateFunctionBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	if info == nil || len(info.Params) != 1 {
		g.renderCompiledFunctionBody(buf, info, lines, retExpr)
		return
	}
	bytes := info.Params[0].GoName
	g.renderCompiledFunctionBodyWithPrefix(buf, info, []string{
		"if " + bytes + " != nil && utf8.Valid(" + bytes + ".Elements) {",
		"\treturn nil, nil",
		"}",
	}, lines, retExpr)
}
