package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNullableI32ParamAndMatchStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn describe(value: ?i32) -> String {",
		"  value match {",
		"    case n: i32 => `n ${n}`,",
		"    case _ => \"none\"",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe(value *int32) (string, *__ableControl)") {
		t.Fatalf("expected describe to keep a native nullable *int32 parameter")
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	for _, fragment := range []string{
		"!= nil",
		"(*__able_tmp_0)",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected nullable match lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nullable match lowering to avoid %q:\n%s", fragment, body)
		}
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_describe")
	if !ok {
		t.Fatalf("could not find describe wrapper")
	}
	if !strings.Contains(wrapBody, "__able_nullable_i32_from_value(arg0Value)") {
		t.Fatalf("expected describe wrapper to use the native nullable arg helper:\n%s", wrapBody)
	}
}

func TestCompilerNullableI32ReturnAndOrElseStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn maybe(flag: bool) -> ?i32 {",
		"  if flag { 7 } else { nil }",
		"}",
		"",
		"fn main() -> i32 {",
		"  maybe(true) or { 0 }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_maybe(flag bool) (*int32, *__ableControl)") {
		t.Fatalf("expected maybe to keep a native nullable *int32 return")
	}
	if !strings.Contains(compiledSrc, "__able_ptr(int32(7))") {
		t.Fatalf("expected nullable integer literal lowering to use __able_ptr(int32(7))")
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_maybe")
	if !ok {
		t.Fatalf("could not find maybe wrapper")
	}
	if !strings.Contains(wrapBody, "return __able_nullable_i32_to_value(compiledResult), nil") {
		t.Fatalf("expected maybe wrapper to use the native nullable return helper:\n%s", wrapBody)
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"== nil",
		"(*",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native nullable or-else lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_any_to_value(",
		"__able_try_cast(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native nullable or-else lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerNullableI64ReturnAndOrElseStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn maybe(flag: bool) -> ?i64 {",
		"  if flag { 7 } else { nil }",
		"}",
		"",
		"fn main() -> i64 {",
		"  maybe(true) or { 0 }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_maybe(flag bool) (*int64, *__ableControl)") {
		t.Fatalf("expected maybe to keep a native nullable *int64 return")
	}
	if !strings.Contains(compiledSrc, "__able_ptr(int64(7))") {
		t.Fatalf("expected nullable integer literal lowering to use __able_ptr(int64(7))")
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_maybe")
	if !ok {
		t.Fatalf("could not find maybe wrapper")
	}
	if !strings.Contains(wrapBody, "return __able_nullable_i64_to_value(compiledResult), nil") {
		t.Fatalf("expected maybe wrapper to use the native nullable i64 return helper:\n%s", wrapBody)
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"== nil",
		"(*",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native nullable or-else lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerNullableF64ReturnStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn maybe(flag: bool) -> ?f64 {",
		"  if flag { 1.25 } else { nil }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_maybe(flag bool) (*float64, *__ableControl)") {
		t.Fatalf("expected maybe to keep a native nullable *float64 return")
	}
	if !strings.Contains(compiledSrc, "__able_ptr(float64(1.25))") {
		t.Fatalf("expected nullable float literal lowering to use __able_ptr(float64(1.25))")
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_maybe")
	if !ok {
		t.Fatalf("could not find maybe wrapper")
	}
	if !strings.Contains(wrapBody, "return __able_nullable_f64_to_value(compiledResult), nil") {
		t.Fatalf("expected maybe wrapper to use the native nullable f64 return helper:\n%s", wrapBody)
	}
}

func TestCompilerNullableCharParamAndMatchStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn describe(value: ?char) -> String {",
		"  value match {",
		"    case ch: char => `c ${ch}`,",
		"    case _ => \"none\"",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe(value *rune) (string, *__ableControl)") {
		t.Fatalf("expected describe to keep a native nullable *rune parameter")
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	for _, fragment := range []string{
		"!= nil",
		"(*__able_tmp_0)",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected nullable char match lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nullable char match lowering to avoid %q:\n%s", fragment, body)
		}
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_describe")
	if !ok {
		t.Fatalf("could not find describe wrapper")
	}
	if !strings.Contains(wrapBody, "__able_nullable_char_from_value(arg0Value)") {
		t.Fatalf("expected describe wrapper to use the native nullable char arg helper:\n%s", wrapBody)
	}
}

func TestCompilerNullableErrorReturnAndMatchStayNative(t *testing.T) {
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
		"fn maybe(flag: bool) -> ?Error {",
		"  if flag { MyError { message: \"bad\" } } else { nil }",
		"}",
		"",
		"fn main() -> String {",
		"  maybe(true) match {",
		"    case err: Error => err.message(),",
		"    case _ => \"none\"",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_maybe(flag bool) (*runtime.ErrorValue, *__ableControl)") {
		t.Fatalf("expected maybe to keep a native nullable *runtime.ErrorValue return:\n%s", compiledSrc)
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_maybe")
	if !ok {
		t.Fatalf("could not find maybe wrapper")
	}
	if !strings.Contains(wrapBody, "return __able_nullable_error_to_value(compiledResult), nil") {
		t.Fatalf("expected maybe wrapper to use the native nullable error return helper:\n%s", wrapBody)
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"!= nil",
		"(*",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native nullable Error match lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native nullable Error match lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerNullableErrorReturnExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct MyError { message: String }

impl Error for MyError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

fn maybe(flag: bool) -> ?Error {
  if flag { MyError { message: "bad" } } else { nil }
}

fn main() {
  outcome := maybe(true) match {
    case err: Error => err.message(),
    case _ => "none"
  }
  if outcome == "bad" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-nullable-error-", source)
}
