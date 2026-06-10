package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringLenBytesMethodUsesNativeValidUTF8Guard(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_len_bytes",
		ReturnType: "uint64",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "len_bytes",
		ExpectsSelf: true,
		Info:        info,
	}
	if !isCanonicalPrimitiveStringLenBytesMethod(method) {
		t.Fatal("expected canonical primitive String.len_bytes to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringLenBytesMethodBody(&buf, info, nil, "0")
	source := buf.String()
	for _, fragment := range []string{
		"if utf8.ValidString(self) && len(self) <= 2147483647",
		"return uint64(len(self)), nil",
		"return 0, nil",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated String.len_bytes body missing %q:\n%s", fragment, source)
		}
	}

	method.TargetName = "StringBuilder"
	if isCanonicalPrimitiveStringLenBytesMethod(method) {
		t.Fatal("StringBuilder.len_bytes must not receive the String primitive lowering")
	}
}

func TestCanonicalPrimitiveStringLenBytesCallUsesNativeValidUTF8Guard(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_len_bytes",
		ReturnType: "uint64",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "len_bytes",
		ExpectsSelf: true,
		Info:        info,
	}
	gen := &generator{}
	ctx := newCompileContext(gen, nil, nil, nil, "demo", nil)
	expr, ok := gen.canonicalPrimitiveStringLenBytesCallExpr(
		ctx,
		method,
		[]string{"next_value()"},
		"__able_compiled_entry_method_String_len_bytes",
	)
	if !ok {
		t.Fatal("expected canonical primitive String.len_bytes call lowering")
	}
	for _, fragment := range []string{
		"func(__able_string_len_bytes_value string)",
		"utf8.ValidString(__able_string_len_bytes_value)",
		"return uint64(len(__able_string_len_bytes_value)), nil",
		"return __able_compiled_entry_method_String_len_bytes(__able_string_len_bytes_value)",
		"}(next_value())",
	} {
		if !strings.Contains(expr, fragment) {
			t.Fatalf("generated String.len_bytes call missing %q:\n%s", fragment, expr)
		}
	}
	if strings.Count(expr, "next_value()") != 1 {
		t.Fatalf("generated String.len_bytes call evaluates receiver more than once:\n%s", expr)
	}

	method.TargetName = "StringBuilder"
	if _, ok := gen.canonicalPrimitiveStringLenBytesCallExpr(
		ctx,
		method,
		[]string{"builder"},
		"__able_compiled_entry_method_StringBuilder_len_bytes",
	); ok {
		t.Fatal("StringBuilder.len_bytes call must not receive primitive String lowering")
	}
}
