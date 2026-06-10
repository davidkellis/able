package compiler

import (
	"bytes"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCanonicalPrimitiveStringContainsMethodUsesNativeValidUTF8Guard(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_contains",
		ReturnType: "bool",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
			{Name: "needle", GoName: "needle", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "contains",
		ExpectsSelf: true,
		Info:        info,
	}
	if !isCanonicalPrimitiveStringContainsMethod(method) {
		t.Fatal("expected canonical primitive String.contains to be recognized")
	}

	var buf bytes.Buffer
	(&generator{}).renderCanonicalPrimitiveStringContainsMethodBody(&buf, info, nil, "false")
	source := buf.String()
	for _, fragment := range []string{
		"if utf8.ValidString(self) && utf8.ValidString(needle)",
		"return strings.Contains(self, needle), nil",
		"return false, nil",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated String.contains body missing %q:\n%s", fragment, source)
		}
	}

	method.Info.Package = "demo"
	if isCanonicalPrimitiveStringContainsMethod(method) {
		t.Fatal("user-defined String.contains must not receive the primitive lowering")
	}
}

func TestCanonicalPrimitiveStringContainsLiteralCallUsesKnownValidNeedle(t *testing.T) {
	info := &functionInfo{
		Package:    "able.text.string",
		GoName:     "method_String_contains",
		ReturnType: "bool",
		Definition: &ast.FunctionDefinition{},
		Params: []paramInfo{
			{Name: "self", GoName: "self", GoType: "string"},
			{Name: "needle", GoName: "needle", GoType: "string"},
		},
	}
	method := &methodInfo{
		TargetName:  "String",
		MethodName:  "contains",
		ExpectsSelf: true,
		Info:        info,
	}
	call := &ast.FunctionCall{
		Callee: &ast.MemberAccessExpression{
			Object: ast.NewIdentifier("value"),
			Member: ast.NewIdentifier("contains"),
		},
		Arguments: []ast.Expression{&ast.StringLiteral{Value: "ei"}},
	}
	gen := &generator{}
	expr, ok := gen.canonicalPrimitiveStringContainsLiteralCallExpr(
		newCompileContext(gen, nil, nil, nil, "demo", nil),
		call,
		method,
		[]string{"value", `"ei"`},
		"__able_compiled_entry_method_String_contains",
	)
	if !ok {
		t.Fatal("expected literal String.contains call to use the known-valid path")
	}
	for _, fragment := range []string{
		"utf8.ValidString(value)",
		`strings.Contains(value, "ei")`,
		`return __able_compiled_entry_method_String_contains(value, "ei")`,
	} {
		if !strings.Contains(expr, fragment) {
			t.Fatalf("literal String.contains call missing %q:\n%s", fragment, expr)
		}
	}

	call.Arguments[0] = ast.NewIdentifier("needle")
	if _, ok := gen.canonicalPrimitiveStringContainsLiteralCallExpr(
		newCompileContext(gen, nil, nil, nil, "demo", nil),
		call,
		method,
		[]string{"value", "needle"},
		"__able_compiled_entry_method_String_contains",
	); ok {
		t.Fatal("dynamic String.contains needle must retain the canonical method")
	}
}
