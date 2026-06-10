package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringFromBytesUncheckedMethodUsesNativeByteSlice(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_from_bytes_unchecked",
		ReturnType: "string",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "bytes", GoName: "bytes", GoType: "*__able_array_u8"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "from_bytes_unchecked",
		ExpectsSelf: false,
		Info:        info,
	}
	if !isCanonicalPrimitiveStringFromBytesUncheckedMethod(method) {
		t.Fatal("expected canonical primitive String.from_bytes_unchecked to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringFromBytesUncheckedMethodBody(&buf, info, nil, "\"\"")
	source := buf.String()
	for _, fragment := range []string{
		"if bytes != nil",
		"return string(bytes.Elements), nil",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated String.from_bytes_unchecked body missing %q:\n%s", fragment, source)
		}
	}

	method.TargetName = "StringBuilder"
	if isCanonicalPrimitiveStringFromBytesUncheckedMethod(method) {
		t.Fatal("StringBuilder must not receive the String primitive lowering")
	}
}
