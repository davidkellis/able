package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringCharsMethodUsesNativeByteSlice(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_chars",
		ReturnType: "__able_iface_Iterator_char",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "chars",
		ExpectsSelf: true,
		Info:        info,
	}
	if !isCanonicalPrimitiveStringCharsMethod(method) {
		t.Fatal("expected canonical primitive String.chars to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringCharsMethodBody(&buf, info, nil, "nil")
	source := buf.String()
	for _, fragment := range []string{
		"if utf8.ValidString(self) && len(self) <= 2147483647",
		"&__able_array_u8{Elements: []uint8(self)}",
		"__able_iface_Iterator_char_wrap_ptr_RawStringCharsIter",
		"return nil, nil",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated String.chars body missing %q:\n%s", fragment, source)
		}
	}

	method.TargetName = "StringBuilder"
	if isCanonicalPrimitiveStringCharsMethod(method) {
		t.Fatal("StringBuilder must not receive the String primitive lowering")
	}
}
