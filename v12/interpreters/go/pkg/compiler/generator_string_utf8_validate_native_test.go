package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringUTF8ValidateFunctionUsesNativeValidator(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		Name:       "utf8_validate",
		GoName:     "fn_utf8_validate",
		ReturnType: "*StringEncodingError",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "bytes", GoName: "bytes", GoType: "*__able_array_u8"},
		},
	}
	if !isCanonicalPrimitiveStringUTF8ValidateFunction(info) {
		t.Fatal("expected canonical primitive utf8_validate to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringUTF8ValidateFunctionBody(&buf, info, nil, "nil")
	source := buf.String()
	for _, fragment := range []string{
		"if bytes != nil && utf8.Valid(bytes.Elements)",
		"return nil, nil",
		"func __able_compiled_fn_utf8_validate(",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated utf8_validate body missing %q:\n%s", fragment, source)
		}
	}

	info.Name = "validate_bytes"
	if isCanonicalPrimitiveStringUTF8ValidateFunction(info) {
		t.Fatal("only canonical utf8_validate may receive the String primitive lowering")
	}
}
