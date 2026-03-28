package compiler

import (
	"strings"
	"testing"
)

func TestCompilerInlineClosedUnionParamAndMatchStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Cat { name: String }",
		"struct Dog { age: i32 }",
		"",
		"fn describe(value: Cat | Dog) -> String {",
		"  value match {",
		"    case cat: Cat => cat.name,",
		"    case dog: Dog => `dog ${dog.age}`",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type __able_union_") {
		t.Fatalf("expected compiled output to define a native union carrier")
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe(value __able_union_") {
		t.Fatalf("expected describe to keep a native union parameter:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	if strings.Contains(body, "__able_try_cast(") || strings.Contains(body, "bridge.MatchType(") {
		t.Fatalf("expected native union match lowering in compiled function body:\n%s", body)
	}
	if !strings.Contains(body, "_as_") {
		t.Fatalf("expected native union branch extraction helper in compiled function body:\n%s", body)
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_describe")
	if !ok {
		t.Fatalf("could not find describe wrapper")
	}
	if !strings.Contains(wrapBody, "_from_value(rt, arg0Value)") {
		t.Fatalf("expected wrapper to use explicit native union arg conversion:\n%s", wrapBody)
	}
}

func TestCompilerNamedClosedUnionReturnWrapsNativeVariants(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Circle { radius: i32 }",
		"struct Rect { width: i32 }",
		"union Shape = Circle | Rect",
		"",
		"fn choose(flag: bool) -> Shape {",
		"  if flag { Circle { radius: 7 } } else { Rect { width: 9 } }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_choose(flag bool) (__able_union_") {
		t.Fatalf("expected choose to return a native union carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "_wrap_ptr_") {
		t.Fatalf("expected union return lowering to wrap native struct variants:\n%s", compiledSrc)
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_choose")
	if !ok {
		t.Fatalf("could not find choose wrapper")
	}
	if !strings.Contains(wrapBody, "_to_value(rt, compiledResult)") {
		t.Fatalf("expected wrapper to use explicit native union return conversion:\n%s", wrapBody)
	}
}

func TestCompilerParameterizedStructUnionMembersUseConcreteStructHelpers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Box T { value: T }",
		"union Item = Box i32 | String",
		"",
		"fn id(value: Item) -> Item { value }",
		"",
	}, "\n"))

	fromBody, ok := findCompiledFunction(result, "__able_union__Box_i32_or_string_from_value")
	if !ok {
		t.Fatalf("could not find native union from-value helper")
	}
	if !strings.Contains(fromBody, "__able_struct_Box_i32_from(coerced)") {
		t.Fatalf("expected native union conversion to use the concrete struct helper for Box i32:\n%s", fromBody)
	}

	toBody, ok := findCompiledFunction(result, "__able_union__Box_i32_or_string_to_value")
	if !ok {
		t.Fatalf("could not find native union to-value helper")
	}
	if !strings.Contains(toBody, "__able_struct_Box_i32_to(rt, raw.Value)") {
		t.Fatalf("expected native union conversion to use the concrete struct runtime helper for Box i32:\n%s", toBody)
	}
	if strings.Contains(fromBody, "__able_struct_Box_from(coerced)") || strings.Contains(toBody, "__able_struct_Box_to(rt, raw.Value)") {
		t.Fatalf("expected native union conversion to avoid base generic struct helpers for Box i32:\nfrom:\n%s\n\nto:\n%s", fromBody, toBody)
	}
}

func TestCompilerOrElseOnErrorUnionUsesNativeCarrierDetection(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn value(ok: bool) -> String | MyError {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  value(false) or { err => err.message() }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "_as_") {
		t.Fatalf("expected or-else to unwrap native union carriers:\n%s", body)
	}
	if strings.Contains(body, "__able_is_error(") {
		t.Fatalf("expected native union or-else to avoid runtime-value error probing:\n%s", body)
	}
}

func TestCompilerMatchOnErrorUnionUsesNativeCarrierDetection(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn value(ok: bool) -> String | MyError {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  value(false) match {",
		"    case err: Error => err.message(),",
		"    case text: String => text",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "_as_") {
		t.Fatalf("expected match to unwrap native union carriers:\n%s", body)
	}
	if strings.Contains(body, "__able_try_cast(") || strings.Contains(body, "bridge.MatchType(") {
		t.Fatalf("expected Error-typed native union match to avoid dynamic type probing:\n%s", body)
	}
	for _, fragment := range []string{
		"runtime.ErrorValue{Message:",
		"var err runtime.ErrorValue =",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected Error binding conversion to stay on the native error carrier at the branch edge (%q):\n%s", fragment, body)
		}
	}
}

func TestCompilerMatchOnErrorUnionExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct MyError { message: String }

impl Error for MyError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

fn value(ok: bool) -> String | MyError {
  if ok { "ok" } else { MyError { message: "bad" } }
}

fn main() {
  outcome := value(false) match {
    case err: Error => err.message(),
    case text: String => text
  }
  if outcome == "bad" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-native-union-error-match-", source)
}

func TestCompilerExplicitErrorUnionParamAndMatchStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn describe(value: String | Error) -> String {",
		"  value match {",
		"    case err: Error => err.message(),",
		"    case text: String => text",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe(value __able_union_") {
		t.Fatalf("expected explicit Error union param to use a native union carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	if !strings.Contains(body, "_as_runtime_ErrorValue(") {
		t.Fatalf("expected explicit Error union match to unwrap the runtime.ErrorValue branch natively:\n%s", body)
	}
	if strings.Contains(body, "__able_try_cast(") || strings.Contains(body, "bridge.MatchType(") {
		t.Fatalf("expected explicit Error union match to avoid dynamic type probing:\n%s", body)
	}
}
