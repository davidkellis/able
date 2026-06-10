package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringSplitCallUsesNativeValidUTF8Path(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_split",
		ReturnType: "*__able_array_String",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
			{Name: "delimiter", GoName: "delimiter", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "split",
		ExpectsSelf: true,
		Info:        info,
	}
	if !isCanonicalPrimitiveStringSplitMethod(method) {
		t.Fatal("expected canonical primitive String.split to be recognized")
	}

	gen := &generator{}
	expr, ok := gen.canonicalPrimitiveStringSplitCallExpr(
		newCompileContext(gen, nil, nil, nil, "demo", nil),
		method,
		[]string{"value()", "delimiter()"},
		"__able_compiled_entry_method_String_split",
	)
	if !ok {
		t.Fatal("expected static String.split call to use the native path")
	}
	for _, fragment := range []string{
		"func(__able_string_split_value string, __able_string_split_delimiter string)",
		"utf8.ValidString(__able_string_split_value)",
		"utf8.ValidString(__able_string_split_delimiter)",
		"Elements: strings.Split(__able_string_split_value, __able_string_split_delimiter)",
		"return __able_compiled_entry_method_String_split(__able_string_split_value, __able_string_split_delimiter)",
		"}(value(), delimiter())",
	} {
		if !strings.Contains(expr, fragment) {
			t.Fatalf("native String.split call missing %q:\n%s", fragment, expr)
		}
	}

	method.Info.Package = "demo"
	if isCanonicalPrimitiveStringSplitMethod(method) {
		t.Fatal("user-defined String.split must not receive the primitive lowering")
	}
}
