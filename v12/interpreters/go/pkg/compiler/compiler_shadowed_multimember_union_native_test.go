package compiler

import (
	"strings"
	"testing"
)

func TestCompilerImportedShadowedInterfaceMultiMemberUnionAliasStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First, Choice3}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn main() -> i32 {",
			"  local: Choice3 (RemoteReader i32) = if true {",
			"    First {}",
			"  } else {",
			"    if false { \"ok\" } else { 9 }",
			"  }",
			"  local match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0,",
			"    case n: i32 => n",
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
	if !strings.Contains(body, "var local __able_union_") {
		t.Fatalf("expected imported shadowed interface multi-member alias local to use a native union carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var local runtime.Value",
		"var local any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported shadowed interface multi-member alias local to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected imported shadowed interface multi-member alias local to keep direct native interface dispatch:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported shadowed interface multi-member alias local to avoid residual runtime.Value union members %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerImportedShadowedCallableMultiMemberResultAliasStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Outcome3}",
			"",
			"struct Thing { local: i32 }",
			"",
			"struct MyError { message: String }",
			"",
			"impl Error for MyError {",
			"  fn message(self: Self) -> String { self.message }",
			"  fn cause(self: Self) -> ?Error { nil }",
			"}",
			"",
			"fn main() -> i32 {",
			"  local: Outcome3 (() -> RemoteThing) = if true {",
			"    fn() -> RemoteThing { RemoteThing { remote: 7 } }",
			"  } else {",
			"    if false { \"bad\" } else { MyError { message: \"oops\" } }",
			"  }",
			"  local match {",
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
	if !strings.Contains(body, "var local __able_union_") {
		t.Fatalf("expected imported shadowed callable multi-member alias local to use a native union carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var local runtime.Value",
		"var local any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_call_value(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported shadowed callable multi-member alias local to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, ".Remote") {
		t.Fatalf("expected imported shadowed callable multi-member alias local to keep direct native field access on the callable result:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported shadowed callable multi-member alias local to avoid residual runtime.Value union members %q:\n%s", fragment, compiledSrc)
		}
	}
}
