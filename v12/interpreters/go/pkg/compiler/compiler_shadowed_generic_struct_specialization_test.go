package compiler

import (
	"strings"
	"testing"
)

func TestCompilerSpecializedGenericImportedShadowedGenericStructResultAliasReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Outcome}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn id<T>(value: T) -> T { value }",
			"",
			"fn main() -> i32 {",
			"  outcome: Outcome RemoteThing = Box { value: RemoteThing { remote: 7 } }",
			"  picked := id(outcome)",
			"  picked match {",
			"    case box: Box RemoteThing => box.value.remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"struct Box T { value: T }",
			"type Outcome T = Result (Box T)",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed generic-struct result alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") || !strings.Contains(body, ".Value.Remote") {
		t.Fatalf("expected specialized generic imported shadowed generic-struct result alias return to stay native:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_union_") {
		t.Fatalf("expected specialized generic imported shadowed generic-struct result alias helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic imported shadowed generic-struct result alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericImportedShadowedGenericStructUnionAliasReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Choice}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn id<T>(value: T) -> T { value }",
			"",
			"fn main() -> i32 {",
			"  choice: Choice RemoteThing = Box { value: RemoteThing { remote: 7 } }",
			"  picked := id(choice)",
			"  picked match {",
			"    case box: Box RemoteThing => box.value.remote,",
			"    case _: String => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"struct Box T { value: T }",
			"type Choice T = (Box T) | String",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed generic-struct union alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") || !strings.Contains(body, ".Value.Remote") {
		t.Fatalf("expected specialized generic imported shadowed generic-struct union alias return to stay native:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_union_") {
		t.Fatalf("expected specialized generic imported shadowed generic-struct union alias helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic imported shadowed generic-struct union alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
