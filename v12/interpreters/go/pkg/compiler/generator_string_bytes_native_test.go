package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringBytesMethodUsesNativeByteSlice(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_bytes",
		ReturnType: "__able_iface_Iterator_u8",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "bytes",
		ExpectsSelf: true,
		Info:        info,
	}
	if !isCanonicalPrimitiveStringBytesMethod(method) {
		t.Fatal("expected canonical primitive String.bytes to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringBytesMethodBody(&buf, info, nil, "nil")
	source := buf.String()
	for _, fragment := range []string{
		"if utf8.ValidString(self) && len(self) <= 2147483647",
		"&__able_array_u8{Elements: []uint8(self)}",
		"__able_iface_Iterator_u8_wrap_ptr_RawStringBytesIter",
		"return nil, nil",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated String.bytes body missing %q:\n%s", fragment, source)
		}
	}

	method.TargetName = "StringBuilder"
	if isCanonicalPrimitiveStringBytesMethod(method) {
		t.Fatal("StringBuilder must not receive the String primitive lowering")
	}
}
