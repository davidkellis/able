package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNoFallbacksForLocalTypeDefinitions(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  type LocalAlias = i32",
		"  struct LocalStruct {",
		"    value: i32",
		"  }",
		"  union LocalUnion = nil | LocalStruct",
		"  interface LocalIface {",
		"    fn value(self: Self) -> i32",
		"  }",
		"  1",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "DefineStruct(\"LocalStruct\"") {
		t.Fatalf("expected local struct definition to define struct bindings in current env")
	}
	if !strings.Contains(compiledSrc, "runtime.UnionDefinitionValue{Node: ast.NewUnionDefinition(ast.NewIdentifier(\"LocalUnion\")") {
		t.Fatalf("expected local union definition to compile into runtime union binding")
	}
	if !strings.Contains(compiledSrc, "runtime.InterfaceDefinitionValue{Node: ast.NewInterfaceDefinition(ast.NewIdentifier(\"LocalIface\")") {
		t.Fatalf("expected local interface definition to compile into runtime interface binding")
	}
	if strings.Contains(compiledSrc, "CallOriginal(\"demo.main\"") {
		t.Fatalf("expected local type definition path to stay compiled without call_original fallback")
	}
}

func TestCompilerNoFallbacksForLocalInterfaceDefinitionWithDefaultImpl(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  interface LocalDefault {",
		"    fn value(self: Self) -> i32 { 7 }",
		"  }",
		"  1",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "runtime.InterfaceDefinitionValue{Node: ast.NewInterfaceDefinition(ast.NewIdentifier(\"LocalDefault\")") {
		t.Fatalf("expected local interface with default impl signature to compile into runtime interface binding")
	}
	if strings.Contains(compiledSrc, "CallOriginal(\"demo.main\"") {
		t.Fatalf("expected local interface default-impl path to stay compiled without call_original fallback")
	}
}
