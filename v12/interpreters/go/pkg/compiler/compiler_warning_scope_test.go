package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestCompilerCrossPackageStructNamesDoNotWarnAsDuplicates(t *testing.T) {
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	entryModule := ast.Mod(
		[]ast.Statement{mainFn},
		nil,
		ast.Pkg([]interface{}{"app"}, false),
	)
	entry := annotatedModule("app", entryModule, "app.able", nil)

	errorsModule := ast.Mod(
		[]ast.Statement{
			ast.StructDef("Span", []*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("u64"), "start"),
				ast.FieldDef(ast.Ty("u64"), "end"),
			}, ast.StructKindNamed, nil, nil, false),
		},
		nil,
		ast.Pkg([]interface{}{"able", "core", "errors"}, false),
	)
	errors := annotatedModule("able.core.errors", errorsModule, "errors.able", nil)

	regexModule := ast.Mod(
		[]ast.Statement{
			ast.StructDef("Span", []*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("u64"), "start"),
				ast.FieldDef(ast.Ty("u64"), "end"),
			}, ast.StructKindNamed, nil, nil, false),
		},
		nil,
		ast.Pkg([]interface{}{"able", "text", "regex"}, false),
	)
	regex := annotatedModule("able.text.regex", regexModule, "regex.able", nil)

	program := &driver.Program{
		Entry:   entry,
		Modules: []*driver.Module{entry, errors, regex},
	}
	result, err := New(Options{PackageName: "compiled"}).Compile(program)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "duplicate struct Span") {
			t.Fatalf("unexpected duplicate struct warning: %q", warning)
		}
	}
}

func TestCompilerDynamicWarningsIgnoreUnreachableModules(t *testing.T) {
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{ast.Ret(nil)},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	entryModule := ast.Mod(
		[]ast.Statement{mainFn},
		nil,
		ast.Pkg([]interface{}{"app"}, false),
	)
	entry := annotatedModule("app", entryModule, "app.able", nil)

	dynFn := ast.Fn(
		"probe",
		nil,
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("dyn"), "def_package"), ast.Str("demo.dynamic")),
			ast.Ret(nil),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	dynModule := ast.Mod(
		[]ast.Statement{dynFn},
		nil,
		ast.Pkg([]interface{}{"tools"}, false),
	)
	tools := annotatedModule("tools", dynModule, "tools.able", nil)

	program := &driver.Program{
		Entry:   entry,
		Modules: []*driver.Module{entry, tools},
	}
	result, err := New(Options{PackageName: "compiled"}).Compile(program)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "module tools uses dynamic calls") {
			t.Fatalf("unexpected dynamic warning for unreachable module: %q", warning)
		}
	}
}

func TestCompilerStaticFallbackGuardIgnoresUnreachableDynamicModules(t *testing.T) {
	fallbackFn := ast.Fn(
		"complex",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Bin("/", ast.Int(1), ast.Int(2))),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Call("complex"),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	entryModule := ast.Mod(
		[]ast.Statement{fallbackFn, mainFn},
		nil,
		ast.Pkg([]interface{}{"app"}, false),
	)
	entry := annotatedModule("app", entryModule, "app.able", nil)

	dynFn := ast.Fn(
		"probe",
		nil,
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("dyn"), "def_package"), ast.Str("demo.dynamic")),
			ast.Ret(nil),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	dynModule := ast.Mod(
		[]ast.Statement{dynFn},
		nil,
		ast.Pkg([]interface{}{"tools"}, false),
	)
	tools := annotatedModule("tools", dynModule, "tools.able", nil)

	program := &driver.Program{
		Entry:   entry,
		Modules: []*driver.Module{entry, tools},
	}
	_, err := New(Options{PackageName: "compiled", RequireStaticNoFallbacks: true}).Compile(program)
	if err == nil {
		t.Fatalf("expected static fallback guard error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "static fallback not allowed") {
		t.Fatalf("expected static fallback guard error, got %q", msg)
	}
	if !strings.Contains(msg, "app.complex") {
		t.Fatalf("expected fallback name in error, got %q", msg)
	}
}
