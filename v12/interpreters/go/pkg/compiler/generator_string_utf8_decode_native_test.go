package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringUTF8DecodeFunctionUsesNativeDecoder(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		Name:       "utf8_decode",
		GoName:     "fn_utf8_decode",
		ReturnType: "__able_union__StringEncodingError_or__Utf8DecodeResult",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "bytes", GoName: "bytes", GoType: "*__able_array_u8"},
			{Name: "offset", GoName: "offset", GoType: "int32"},
			{Name: "length", GoName: "length", GoType: "int32"},
		},
	}
	if !isCanonicalPrimitiveStringUTF8DecodeFunction(info) {
		t.Fatal("expected canonical primitive utf8_decode to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringUTF8DecodeFunctionBody(&buf, info, nil, "nil")
	source := buf.String()
	for _, fragment := range []string{
		"offset >= 0",
		"length > offset",
		"utf8.DecodeRune(bytes.Elements[offset:length])",
		"__able_native_rune == utf8.RuneError && __able_native_width == 1",
		"__able_union__StringEncodingError_or__Utf8DecodeResult_wrap_ptr_Utf8DecodeResult",
		"return nil, nil",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated utf8_decode body missing %q:\n%s", fragment, source)
		}
	}

	info.Name = "decode_multibyte"
	if isCanonicalPrimitiveStringUTF8DecodeFunction(info) {
		t.Fatal("only canonical utf8_decode may receive the String primitive lowering")
	}
}
