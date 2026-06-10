package compiler

import (
	"strings"
	"testing"
)

func TestCompilerLowersNamedSourceReexportWithoutBootstrap(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.wrapper.{value}",
			"",
			"fn main() -> i32 { value() }",
			"",
		}, "\n"),
		"source.able": strings.Join([]string{
			"package source",
			"",
			"fn value() -> i32 { 42 }",
			"",
		}, "\n"),
		"wrapper.able": strings.Join([]string{
			"package wrapper",
			"",
			"import demo.source.{value}",
			"export value",
			"",
		}, "\n"),
	})

	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("named source re-export should keep the compiled import path static")
	}
	if !strings.Contains(compiledSrc, "42") {
		t.Fatalf("expected the original re-exported function to be compiled")
	}
}

func TestCompilerLowersWildcardSourceReexportWithoutBootstrap(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.wrapper.*",
			"",
			"fn main() -> i32 { value() }",
			"",
		}, "\n"),
		"source.able": strings.Join([]string{
			"package source",
			"",
			"fn value() -> i32 { 42 }",
			"",
		}, "\n"),
		"wrapper.able": strings.Join([]string{
			"package wrapper",
			"",
			"export * from demo.source",
			"",
		}, "\n"),
	})

	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("wildcard source re-export should keep the compiled import path static")
	}
	if !strings.Contains(compiledSrc, "42") {
		t.Fatalf("expected the original wildcard re-exported function to be compiled")
	}
}

func TestCompilerPreservesStructCarrierThroughSourceReexport(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.wrapper.{Box}",
			"",
			"fn main() -> i32 {",
			"  box: Box = Box { value: 42 }",
			"  box.value",
			"}",
			"",
		}, "\n"),
		"source.able": strings.Join([]string{
			"package source",
			"",
			"struct Box { value: i32 }",
			"",
		}, "\n"),
		"wrapper.able": strings.Join([]string{
			"package wrapper",
			"",
			"import demo.source.{Box}",
			"export Box",
			"",
		}, "\n"),
	})

	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("re-exported struct should retain its native carrier")
	}
	if !strings.Contains(compiledSrc, "type Box struct") {
		t.Fatalf("expected the original struct carrier to be compiled")
	}
}
