package compiler

import (
	"strings"
	"testing"
)

func TestCompilerInferredGenericFunctionCallStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn id<T>(value: T) -> T { value }",
		"",
		"fn main() -> i32 {",
		"  id(1)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_compiled_fn_id_spec(int32(1))") {
		t.Fatalf("expected checked generic free-function call to lower through the specialized compiled helper:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_call_named(\"id\"",
		"runtime.Value",
		"bridge.AsInt(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected checked generic free-function call to avoid %q:\n%s", fragment, body)
		}
	}
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value int32) (int32, *__ableControl)") {
		t.Fatalf("expected specialized generic free-function signature to stay native:\n%s", compiledSrc)
	}
}

func TestCompilerGenericAliasFunctionCallsStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Pair T = Array T",
		"",
		"fn make_pair<T>(value: T) -> Pair T {",
		"  [value, value]",
		"}",
		"",
		"fn first<T>(values: Pair T) -> T {",
		"  values[0]",
		"}",
		"",
		"fn main() -> i32 {",
		"  first(make_pair(5))",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_call_named(\"make_pair\"",
		"__able_call_named(\"first\"",
		"runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected generic-alias free-function calls to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_compiled_fn_make_pair_spec(") || !strings.Contains(body, "__able_compiled_fn_first_spec(") {
		t.Fatalf("expected generic-alias free-function calls to use specialized compiled helpers:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_fn_make_pair_spec(value int32) (*__able_array_i32, *__ableControl)",
		"func __able_compiled_fn_first_spec(values *__able_array_i32) (int32, *__ableControl)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generic alias specialization to produce native signatures containing %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerInferredGenericNominalReturnStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-inferred-generic-nominal-return", strings.Join([]string{
		"package demo",
		"",
		"struct Pair A B { left: A, right: B }",
		"",
		"fn make_pair(left: A, right: B) {",
		"  Pair A B { left: left, right: right }",
		"}",
		"",
		"fn main() -> void {",
		"  pair := make_pair(true, 42)",
		"  print(`${pair.right}`)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var pair runtime.Value",
		"__able_any_to_value(__able_tmp_",
		"__able_struct_Pair_bool_i32_from(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected inferred generic nominal return to stay native and avoid %q:\n%s", fragment, body)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_make_pair_spec(left bool, right int32) (*Pair_bool_i32, *__ableControl)") {
		t.Fatalf("expected inferred generic nominal specialization to produce a concrete native return signature:\n%s", compiledSrc)
	}

	stdout := strings.TrimSpace(compileAndRunExecSourceWithOptions(t, "ablec-inferred-generic-nominal-return-exec", strings.Join([]string{
		"package demo",
		"",
		"struct Pair A B { left: A, right: B }",
		"",
		"fn make_pair(left: A, right: B) {",
		"  Pair A B { left: left, right: right }",
		"}",
		"",
		"fn main() -> void {",
		"  pair := make_pair(true, 42)",
		"  print(`${pair.right}`)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
		EmitMain:           true,
	}))
	if stdout != "42" {
		t.Fatalf("expected compiled inferred generic nominal return path to execute, got %q", stdout)
	}
}

func TestCompilerSpecializedGenericUnionReturnStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn choose<T>(flag: bool, value: T) {",
		"  if flag { value } else { \"ok\" }",
		"}",
		"",
		"fn main() -> i32 {",
		"  picked := choose(true, 7)",
		"  picked match {",
		"    case n: i32 => n,",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic union return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") {
		t.Fatalf("expected specialized generic union return local to stay on a native union carrier:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_choose_spec(flag bool, value int32) (__able_union_") {
		t.Fatalf("expected specialized generic union helper to use a native union return signature:\n%s", compiledSrc)
	}
	if strings.Contains(compiledSrc, "func __able_compiled_fn_choose_spec(flag bool, value int32) (runtime.Value, *__ableControl)") {
		t.Fatalf("expected specialized generic union helper to avoid runtime.Value return signature:\n%s", compiledSrc)
	}
}

func TestCompilerSpecializedDuplicateUnionReturnCollapsesToNativeCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Choice T = T | String",
		"",
		"fn choose<T>(flag: bool, value: T) -> Choice T {",
		"  if flag { value } else { \"ok\" }",
		"}",
		"",
		"fn main() -> String {",
		"  picked := choose(true, \"yes\")",
		"  picked",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"var picked __able_union_",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized duplicate-union local to collapse to a native string carrier and avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked string") {
		t.Fatalf("expected specialized duplicate-union local to stay on native string carrier:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_choose_spec(flag bool, value string) (string, *__ableControl)") {
		t.Fatalf("expected specialized duplicate-union helper to collapse to native string return signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_choose_spec(flag bool, value string) (__able_union_",
		"func __able_compiled_fn_choose_spec(flag bool, value string) (runtime.Value, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized duplicate-union helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedDuplicateResultReturnCollapsesToNativeErrorCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct RootError { message: String }",
		"",
		"impl Error for RootError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn choose<T>(flag: bool, value: T) -> !T {",
		"  if flag { value } else { RootError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  err: Error = RootError { message: \"boom\" }",
		"  picked := choose(true, err)",
		"  picked.message()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"var picked __able_union_",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized duplicate-result local to collapse to native Error carrier and avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked runtime.ErrorValue") {
		t.Fatalf("expected specialized duplicate-result local to stay on native Error carrier:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_choose_spec(flag bool, value runtime.ErrorValue) (runtime.ErrorValue, *__ableControl)") {
		t.Fatalf("expected specialized duplicate-result helper to collapse to native Error return signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_choose_spec(flag bool, value runtime.ErrorValue) (__able_union_",
		"func __able_compiled_fn_choose_spec(flag bool, value runtime.ErrorValue) (runtime.Value, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized duplicate-result helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericInterfaceReturnStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
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
		"fn id<T>(value: T) -> T { value }",
		"",
		"fn main() -> i32 {",
		"  reader: Reader i32 = First {}",
		"  picked := id(reader)",
		"  picked.read()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic interface return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_iface_Reader_i32") {
		t.Fatalf("expected specialized generic interface return local to stay on the native interface carrier:\n%s", body)
	}
	if !strings.Contains(body, "picked.read()") {
		t.Fatalf("expected specialized generic interface return to keep direct native interface dispatch:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_iface_Reader_i32) (__able_iface_Reader_i32, *__ableControl)") {
		t.Fatalf("expected specialized generic interface helper to use a native interface signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic interface helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericUnionInterfaceMemberStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
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
		"type Choice T = T | String",
		"",
		"fn choose<T>(flag: bool, value: T) -> Choice T {",
		"  if flag { value } else { \"ok\" }",
		"}",
		"",
		"fn main() -> i32 {",
		"  reader: Reader i32 = First {}",
		"  picked := choose(true, reader)",
		"  picked match {",
		"    case next: Reader i32 => next.read(),",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic union interface local to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") {
		t.Fatalf("expected specialized generic union interface local to stay on a native union carrier:\n%s", body)
	}
	if !strings.Contains(body, "var next __able_iface_Reader_i32") {
		t.Fatalf("expected specialized generic union interface branch binding to stay on the native interface carrier:\n%s", body)
	}
	if !strings.Contains(body, "next.read()") {
		t.Fatalf("expected specialized generic union interface match to keep direct native interface dispatch:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_choose_spec(flag bool, value __able_iface_Reader_i32) (__able_union_") {
		t.Fatalf("expected specialized generic union interface helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_choose_spec(flag bool, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_choose_spec(flag bool, value any) (any, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic union interface helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericImportedShadowedInterfaceAliasReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First, Choice}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn id<T>(value: T) -> T { value }",
			"",
			"fn main() -> i32 {",
			"  choice: Choice (RemoteReader i32) = First {}",
			"  picked := id(choice)",
			"  picked match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0",
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
			"type Choice T = T | String",
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
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed interface alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") || !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected specialized generic imported shadowed interface alias return to stay native:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_union_") {
		t.Fatalf("expected specialized generic imported shadowed interface alias helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic imported shadowed interface alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericImportedShadowedCallableAliasReturnStaysNative(t *testing.T) {
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
			"  outcome: Outcome (() -> RemoteThing) = fn() -> RemoteThing { RemoteThing { remote: 7 } }",
			"  picked := id(outcome)",
			"  picked match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
			"type Outcome T = Result T",
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
		"__able_call_value(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed callable alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") || !strings.Contains(body, ".Remote") {
		t.Fatalf("expected specialized generic imported shadowed callable alias return to stay native:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_union_") {
		t.Fatalf("expected specialized generic imported shadowed callable alias helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic imported shadowed callable alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericImportedShadowedInterfaceResultAliasReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First, Outcome}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn id<T>(value: T) -> T { value }",
			"",
			"fn main() -> i32 {",
			"  outcome: Outcome (RemoteReader i32) = First {}",
			"  picked := id(outcome)",
			"  picked match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: Error => 0",
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
			"type Outcome T = Result T",
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
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed interface result alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") || !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected specialized generic imported shadowed interface result alias return to stay native:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_union_") {
		t.Fatalf("expected specialized generic imported shadowed interface result alias helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic imported shadowed interface result alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerSpecializedGenericImportedShadowedCallableUnionAliasReturnStaysNative(t *testing.T) {
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
			"  choice: Choice (() -> RemoteThing) = fn() -> RemoteThing { RemoteThing { remote: 7 } }",
			"  picked := id(choice)",
			"  picked match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: String => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
			"type Choice T = T | String",
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
		"__able_call_value(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed callable union alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_union_") || !strings.Contains(body, ".Remote") {
		t.Fatalf("expected specialized generic imported shadowed callable union alias return to stay native:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_union_") {
		t.Fatalf("expected specialized generic imported shadowed callable union alias helper to use a native union signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized generic imported shadowed callable union alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
