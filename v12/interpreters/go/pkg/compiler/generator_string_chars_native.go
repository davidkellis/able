package compiler

import "bytes"

// isCanonicalPrimitiveStringCharsMethod identifies the language String
// character iterator. It deliberately excludes methods with the same name on
// nominal types and retains the original body for invalid or oversized input.
func isCanonicalPrimitiveStringCharsMethod(method *methodInfo) bool {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return false
	}
	info := method.Info
	return info.Package == "able.text.string" && method.TargetName == "String" &&
		method.MethodName == "chars" && method.ExpectsSelf &&
		info.ReturnType == "__able_iface_Iterator_char" && len(info.Params) == 1 &&
		info.Params[0].GoType == "string"
}

func (g *generator) renderCanonicalPrimitiveStringCharsMethodBody(buf *bytes.Buffer, info *functionInfo, lines []string, retExpr string) {
	if info == nil || len(info.Params) != 1 {
		g.renderCompiledMethodBody(buf, info, lines, retExpr)
		return
	}
	self := info.Params[0].GoName
	g.renderCompiledMethodBodyWithPrefix(buf, info, []string{
		"if utf8.ValidString(" + self + ") && len(" + self + ") <= 2147483647 {",
		"\treturn __able_iface_Iterator_char_wrap_ptr_RawStringCharsIter(&RawStringCharsIter{Bytes: &__able_array_u8{Elements: []uint8(" + self + ")}, Offset: int32(0), Len_bytes: int32(len(" + self + "))}), nil",
		"}",
	}, lines, retExpr)
}
