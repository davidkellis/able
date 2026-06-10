package compiler

import (
	"strings"
	"testing"
)

func TestCompilerImportedEnvironmentIndependentCallsUseRawBodies(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{pure_chain::aliased_pure, checked, identity, shadow_seed}",
			"",
			"fn call_pure(value: i32) -> i32 { aliased_pure(value) }",
			"fn call_checked(value: i32) -> i32 { checked(value) }",
			"fn call_generic(value: i32) -> i32 { identity(value) }",
			"fn call_shadowed(value: i32) -> i32 { shadow_seed(value) }",
			"fn main() -> i32 { call_pure(20) + call_checked(2) + call_generic(3) + call_shadowed(4) }",
			"",
		}, "\n"),
		"remote/helpers.able": strings.Join([]string{
			"fn add_one(value: i32) -> i32 { value + 1 }",
			"fn pure_chain(value: i32) -> i32 { add_one(value) }",
			"fn checked(value: i32) -> i32 { value + 2 }",
			"fn identity<T>(value: T) -> T { value }",
			"SEED: i32 := 7",
			"fn shadow_seed(SEED: i32) -> i32 { SEED + 1 }",
			"",
		}, "\n"),
	})

	for _, test := range []struct {
		caller string
		callee string
	}{
		{"__able_compiled_fn_call_pure", "fn_pure_chain"},
		{"__able_compiled_fn_call_checked", "fn_checked"},
		{"__able_compiled_fn_call_shadowed", "fn_shadow_seed"},
	} {
		body := mustCompiledFunctionBody(t, result, test.caller)
		if !strings.Contains(body, "__able_compiled_"+test.callee+"(") {
			t.Fatalf("%s should call imported raw body %s:\n%s", test.caller, test.callee, body)
		}
		if strings.Contains(body, "__able_compiled_entry_"+test.callee+"(") {
			t.Fatalf("%s should avoid imported entry wrapper %s:\n%s", test.caller, test.callee, body)
		}
	}

	genericBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_call_generic")
	if strings.Contains(genericBody, "__able_compiled_entry_") || !strings.Contains(genericBody, "__able_compiled_fn_identity") {
		t.Fatalf("generic imported identity should use its proven-independent specialization raw body:\n%s", genericBody)
	}
}

func TestCompilerImportedEnvironmentDependentCallsRetainEntryWrappers(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{read_seed, read_seed_chain, bump_seed, calls_host}",
			"",
			"fn call_global() -> i32 { read_seed() }",
			"fn call_global_chain() -> i32 { read_seed_chain() }",
			"fn call_mutable_global() -> i32 { bump_seed() }",
			"fn call_unknown_extern() -> i32 { calls_host() }",
			"fn main() -> i32 { call_global() + call_global_chain() + call_mutable_global() + call_unknown_extern() }",
			"",
		}, "\n"),
		"remote/helpers.able": strings.Join([]string{
			"SEED: i32 := 7",
			"",
			"fn read_seed() -> i32 { SEED }",
			"fn read_seed_chain() -> i32 { read_seed() }",
			"fn bump_seed() -> i32 { SEED = SEED + 1; SEED }",
			"extern go fn host_answer() -> i32 { return 42 }",
			"fn calls_host() -> i32 { host_answer() }",
			"",
		}, "\n"),
	})

	for _, test := range []struct {
		caller string
		callee string
	}{
		{"__able_compiled_fn_call_global", "fn_read_seed"},
		{"__able_compiled_fn_call_global_chain", "fn_read_seed_chain"},
		{"__able_compiled_fn_call_mutable_global", "fn_bump_seed"},
		{"__able_compiled_fn_call_unknown_extern", "fn_calls_host"},
	} {
		body := mustCompiledFunctionBody(t, result, test.caller)
		if !strings.Contains(body, "__able_compiled_entry_"+test.callee+"(") {
			t.Fatalf("%s should retain imported entry wrapper %s:\n%s", test.caller, test.callee, body)
		}
	}
}

func TestCompilerImportedEntryWrapperLocalizesEnvironmentDependency(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.middle.{read_remote}",
			"",
			"fn main() -> i32 { read_remote() }",
			"",
		}, "\n"),
		"middle/helpers.able": strings.Join([]string{
			"import demo.remote.{read_seed}",
			"",
			"export read_remote",
			"fn read_remote() -> i32 { read_seed() }",
			"",
		}, "\n"),
		"remote/helpers.able": strings.Join([]string{
			"SEED: i32 := 7",
			"fn read_seed() -> i32 { SEED }",
			"fn bump_seed() -> i32 { SEED = SEED + 1; SEED }",
			"",
		}, "\n"),
	})

	middleBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_read_remote")
	if !strings.Contains(middleBody, "__able_compiled_entry_fn_read_seed(") {
		t.Fatalf("cross-package dependent call must retain the callee entry wrapper:\n%s", middleBody)
	}
	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_fn_read_remote(") ||
		strings.Contains(mainBody, "__able_compiled_entry_fn_read_remote(") {
		t.Fatalf("the callee entry wrapper should localize its environment dependency:\n%s", mainBody)
	}
}

func TestCompilerEnvironmentIndependentTypedCallablesKeepOnlyRequiredEntryWrappers(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{apply}",
			"",
			"IMMUTABLE_SEED: i32 := 7",
			"MUTABLE_SEED: i32 := 9",
			"fn call_pure(value: i32) -> i32 { apply(fn(item: i32) -> i32 { item + 1 }, value) }",
			"fn call_immutable(value: i32) -> i32 { apply(fn(item: i32) -> i32 { item + IMMUTABLE_SEED }, value) }",
			"fn call_mutable(value: i32) -> i32 { apply(fn(item: i32) -> i32 { item + MUTABLE_SEED }, value) }",
			"fn bump_mutable() -> i32 { MUTABLE_SEED = MUTABLE_SEED + 1; MUTABLE_SEED }",
			"fn main() -> i32 { call_pure(4) + call_immutable(4) + call_mutable(4) + bump_mutable() }",
			"",
		}, "\n"),
		"remote/helpers.able": "fn apply(callback: (i32 -> i32), value: i32) -> i32 { callback(value) }\n",
	})

	for _, caller := range []string{"__able_compiled_fn_call_pure", "__able_compiled_fn_call_immutable", "__able_compiled_fn_call_mutable"} {
		body := mustCompiledFunctionBody(t, result, caller)
		if !strings.Contains(body, "__able_compiled_fn_apply(") || strings.Contains(body, "__able_compiled_entry_fn_apply(") {
			t.Fatalf("%s should call the environment-independent typed-callable body directly:\n%s", caller, body)
		}
	}

	pureBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_call_pure")
	if strings.Contains(pureBody, "bridge.SwapEnvIfNeeded") {
		t.Fatalf("parameter-only lambda should not recover the package environment:\n%s", pureBody)
	}
	immutableBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_call_immutable")
	if strings.Contains(immutableBody, "bridge.SwapEnvIfNeeded") {
		t.Fatalf("statically lowered immutable binding should not recover the package environment:\n%s", immutableBody)
	}
	mutableBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_call_mutable")
	if !strings.Contains(mutableBody, "bridge.SwapEnvIfNeeded") {
		t.Fatalf("mutable package-binding lambda must retain its environment guard:\n%s", mutableBody)
	}
}

func TestCompilerEnvironmentIndependentUnknownCallerUsesRawBody(t *testing.T) {
	g := &generator{
		environmentIndependentGoNames: map[string]bool{"impl_Map_get": true},
	}
	if got := g.compiledCallTargetNameForPackage("", "remote", "impl_Map_get"); got != "__able_compiled_impl_Map_get" {
		t.Fatalf("environment-independent adapter target = %q, want raw body", got)
	}
	if got := g.compiledCallTargetNameForPackage("", "remote", "impl_Map_set"); got != "__able_compiled_entry_impl_Map_set" {
		t.Fatalf("unproven adapter target = %q, want entry wrapper", got)
	}
}

func TestCompilerEnvironmentIndependenceKeepsRuntimeEntryWrapper(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able":           "package demo\n\nimport demo.remote.{answer}\n\nfn main() -> i32 { answer() }\n",
		"remote/helpers.able": "fn answer() -> i32 { 42 }\n",
	})
	entry := mustCompiledFunctionBody(t, result, "__able_compiled_entry_fn_answer")
	if !strings.Contains(entry, "bridge.SwapEnvIfNeeded(__able_runtime,") || !strings.Contains(entry, "return __able_compiled_fn_answer()") {
		t.Fatalf("runtime entry wrapper must remain available and preserve package environment semantics:\n%s", entry)
	}
}

func TestCompilerImportedMethodEnvironmentIndependenceUsesRawBodies(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Counter}",
			"",
			"fn pure(value: i32) -> i32 { Counter.add_one(value) }",
			"fn dependent() -> i32 { Counter.read_seed() }",
			"fn main() -> i32 { pure(4) + dependent() }",
			"",
		}, "\n"),
		"remote/helpers.able": strings.Join([]string{
			"export Counter",
			"",
			"SEED: i32 := 7",
			"",
			"struct Counter {}",
			"",
			"methods Counter {",
			"  fn add_one(value: i32) -> i32 { value + 1 }",
			"  fn read_seed() -> i32 { SEED }",
			"}",
			"",
		}, "\n"),
	})

	pureBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_pure")
	if !strings.Contains(pureBody, "__able_compiled_method_Counter_add_one(") ||
		strings.Contains(pureBody, "__able_compiled_entry_method_Counter_add_one(") {
		t.Fatalf("proven-independent imported method should use its raw body:\n%s", pureBody)
	}

	immutableBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_dependent")
	if !strings.Contains(immutableBody, "__able_compiled_method_Counter_read_seed(") ||
		strings.Contains(immutableBody, "__able_compiled_entry_method_Counter_read_seed(") {
		t.Fatalf("method reading a statically lowered immutable binding should use its raw body:\n%s", immutableBody)
	}
}
