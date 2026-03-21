package compiler

import (
	"strings"
	"testing"
)

func TestCompilerTypeAliasStructFieldStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-type-alias-struct-field", strings.Join([]string{
		"package demo",
		"",
		"type MyHandle = i64",
		"",
		"struct Box {",
		"  handle: MyHandle",
		"}",
		"",
		"fn build() -> Box {",
		"  Box { handle: 7 }",
		"}",
		"",
		"fn read(box: Box) -> i64 {",
		"  box.handle",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type Box struct {\n\tHandle int64\n}") {
		t.Fatalf("expected alias-backed struct field to lower to int64:\n%s", compiledSrc)
	}
	if strings.Contains(compiledSrc, "type Box struct {\n\tHandle runtime.Value\n}") {
		t.Fatalf("expected alias-backed struct field to avoid runtime.Value:\n%s", compiledSrc)
	}
}
