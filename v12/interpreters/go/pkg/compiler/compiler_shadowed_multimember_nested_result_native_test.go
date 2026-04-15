package compiler

import (
	"strings"
	"testing"
)

func TestCompilerImportedNestedResultMultiMemberInterfaceAliasStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Choice3}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn main(value: !(Choice3 (RemoteReader i32))) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0,",
			"    case n: i32 => n,",
			"    case _: Error => -1",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type Choice3 T = T | String | i32",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"value __able_union___able_union_",
		"_as___able_union_",
		"func __able_compiled_fn_main(value runtime.Value)",
		"func __able_compiled_fn_main(value any)",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported nested result multi-member interface alias to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected imported nested result multi-member interface alias to keep direct native interface dispatch:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported nested result multi-member interface alias to avoid residual runtime.Value union members %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerImportedNestedResultMultiMemberCallableAliasStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Outcome3}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn main(value: !(Outcome3 (() -> RemoteThing))) -> i32 {",
			"  value match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: String => 0,",
			"    case _: Error => -1",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
			"type Outcome3 T = Error | T | String",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"value __able_union___able_union_",
		"_as___able_union_",
		"func __able_compiled_fn_main(value runtime.Value)",
		"func __able_compiled_fn_main(value any)",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_call_value(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported nested result multi-member callable alias to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, ".Remote") {
		t.Fatalf("expected imported nested result multi-member callable alias to keep direct native field access:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported nested result multi-member callable alias to avoid residual runtime.Value union members %q:\n%s", fragment, compiledSrc)
		}
	}
}
