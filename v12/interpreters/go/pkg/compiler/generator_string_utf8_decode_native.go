package compiler

import "bytes"

// isCanonicalPrimitiveStringUTF8DecodeFunction identifies the internal String
// decoder shared by the language character operations. It is a String
// primitive boundary, not a lowering for a nominal iterator or container.
func isCanonicalPrimitiveStringUTF8DecodeFunction(info *functionInfo) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	return info.Package == "able.text.string" && info.Name == "utf8_decode" &&
		info.ReturnType == "__able_union__StringEncodingError_or__Utf8DecodeResult" &&
		len(info.Params) == 3 && info.Params[0].GoType == "*__able_array_u8" &&
		info.Params[1].GoType == "int32" && info.Params[2].GoType == "int32"
}

func (g *generator) renderCanonicalPrimitiveStringUTF8DecodeFunctionBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	if info == nil || len(info.Params) != 3 {
		g.renderCompiledFunctionBody(buf, info, lines, retExpr)
		return
	}
	bytes := info.Params[0].GoName
	offset := info.Params[1].GoName
	length := info.Params[2].GoName
	g.renderCompiledFunctionBodyWithPrefix(buf, info, []string{
		"if " + bytes + " != nil && " + offset + " >= 0 && " + length + " > " + offset + " && int64(" + length + ") <= int64(len(" + bytes + ".Elements)) {",
		"\t__able_native_rune, __able_native_width := utf8.DecodeRune(" + bytes + ".Elements[" + offset + ":" + length + "])",
		"\tif !(__able_native_rune == utf8.RuneError && __able_native_width == 1) {",
		"\t\treturn __able_union__StringEncodingError_or__Utf8DecodeResult_wrap_ptr_Utf8DecodeResult(&Utf8DecodeResult{Codepoint: __able_native_rune, Next_offset: " + offset + " + int32(__able_native_width)}), nil",
		"\t}",
		"}",
	}, lines, retExpr)
}
