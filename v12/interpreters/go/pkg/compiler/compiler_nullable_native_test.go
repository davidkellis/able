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

func TestCompilerNullableArrayNilComparisonStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Box {",
		"  value: i32",
		"}",
		"",
		"fn maybe(flag: bool) -> ?Box {",
		"  if flag { Box { value: 1 } } else { nil }",
		"}",
		"",
		"fn left() -> bool {",
		"  maybe(false) == nil",
		"}",
		"",
		"fn right() -> bool {",
		"  nil == maybe(false)",
		"}",
		"",
	}, "\n"))

	for _, name := range []string{"__able_compiled_fn_left", "__able_compiled_fn_right"} {
		body, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find compiled function %s", name)
		}
		if !strings.Contains(body, "== nil") && !strings.Contains(body, "nil ==") && !strings.Contains(body, "== (*Box)(nil)") && !strings.Contains(body, "(*Box)(nil) ==") {
			t.Fatalf("expected native nil comparison in %s:\n%s", name, body)
		}
		for _, fragment := range []string{
			"__able_struct_Array_to(",
			"__able_binary_op(",
		} {
			if strings.Contains(body, fragment) {
				t.Fatalf("expected native nil comparison in %s to avoid %q:\n%s", name, fragment, body)
			}
		}
	}
}

func TestCompilerNullableStringEqualityStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn same(left: ?String, right: ?String) -> bool {",
		"  left == right",
		"}",
		"",
		"fn different(left: ?String, right: ?String) -> bool {",
		"  left != right",
		"}",
		"",
	}, "\n"))

	for _, name := range []string{"__able_compiled_fn_same", "__able_compiled_fn_different"} {
		body, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find compiled function %s", name)
		}
		for _, fragment := range []string{
			"== nil",
			"!= nil",
			"(*",
		} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("expected native nullable string equality lowering in %s to contain %q:\n%s", name, fragment, body)
			}
		}
		for _, fragment := range []string{
			"__able_binary_op(",
			"__able_nullable_string_to_value(",
			"bridge.AsBool(",
		} {
			if strings.Contains(body, fragment) {
				t.Fatalf("expected native nullable string equality lowering in %s to avoid %q:\n%s", name, fragment, body)
			}
		}
	}
}

func TestCompilerNullableStringEqualityExecutes(t *testing.T) {
	compileAndRunSource(t, "ablec-nullable-string-equality-", strings.Join([]string{
		"package demo",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn same(left: ?String, right: ?String) -> bool {",
		"  left == right",
		"}",
		"",
		"fn different(left: ?String, right: ?String) -> bool {",
		"  left != right",
		"}",
		"",
		"fn main() {",
		"  if same(\"a\", \"a\") && same(nil, nil) && different(\"a\", nil) && different(nil, \"b\") {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n"))
}

func TestCompilerAnyNilComparisonUsesRuntimeNilHelper(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn is_nil(value: any) -> bool {",
		"  value == nil",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_is_nil")
	if !ok {
		t.Fatalf("could not find compiled is_nil function")
	}
	if !strings.Contains(body, "__able_is_nil(value)") && !strings.Contains(body, "__able_is_nil(__able_any_to_value(value))") {
		t.Fatalf("expected runtime nil comparison to lower through __able_is_nil:\n%s", body)
	}
	if strings.Contains(body, "__able_binary_op(\"==\"") {
		t.Fatalf("expected runtime nil comparison to avoid __able_binary_op equality fallback:\n%s", body)
	}
}

func TestCompilerSafeNavigationMethodAndFieldStayNativeNullable(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Box { value: i32 }",
		"",
		"methods Box {",
		"  fn add(self: Self, delta: i32) -> i32 { self.value + delta }",
		"}",
		"",
		"fn maybe_box(make_nil: bool) -> ?Box {",
		"  if make_nil { nil } else { Box { value: 10 } }",
		"}",
		"",
		"fn main() -> String {",
		"  b := maybe_box(true)?.add(1)",
		"  c := maybe_box(true)?.value",
		"  `b ${b} c ${c}`",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"var b *int32",
		"var c *int32",
		"__able_ptr(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected safe navigation to keep native nullable carriers and contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_member_get_method(",
		"__able_call_value(",
		"bridge.AsInt(__able_tmp_",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected safe navigation to avoid %q in compiled main:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerSafeNavigationMethodAndFieldExecute(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"struct Box { value: i32 }",
		"",
		"methods Box {",
		"  fn add(self: Self, delta: i32) -> i32 { self.value + delta }",
		"}",
		"",
		"ticks := 0",
		"",
		"fn tick() -> i32 {",
		"  ticks = ticks + 1",
		"  ticks",
		"}",
		"",
		"fn maybe_box(make_nil: bool) -> ?Box {",
		"  if make_nil { nil } else { Box { value: 10 } }",
		"}",
		"",
		"fn main() -> void {",
		"  _ = maybe_box(false)?.add(tick())",
		"  if ticks != 1 { raise \"bad-a\" }",
		"  _ = maybe_box(true)?.add(tick())",
		"  if ticks != 1 { raise \"bad-b\" }",
		"  _ = maybe_box(true)?.value",
		"  if ticks != 1 { raise \"bad-c\" }",
		"}",
		"",
	}, "\n")

	_ = compileAndRunSourceWithOptions(t, "ablec-safe-navigation-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
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

func TestCompilerNullableStructTypedMatchRequiresNonNil(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"fn maybe(flag: bool) -> ?MyError {",
		"  if flag { MyError { message: \"bad\" } } else { nil }",
		"}",
		"",
		"fn describe(flag: bool) -> String {",
		"  maybe(flag) match {",
		"    case err: MyError => err.message,",
		"    case nil => \"none\"",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	if !strings.Contains(body, "!= nil") {
		t.Fatalf("expected nullable struct typed match to guard against nil before narrowing:\n%s", body)
	}
	if strings.Contains(body, "if true {") {
		t.Fatalf("expected nullable struct typed match to avoid unconditional typed-pattern success:\n%s", body)
	}
}

func TestCompilerNullableStructTypedMatchExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct MyError { message: String }

fn maybe(flag: bool) -> ?MyError {
  if flag { MyError { message: "bad" } } else { nil }
}

fn describe(flag: bool) -> String {
  maybe(flag) match {
    case err: MyError => err.message,
    case nil => "none"
  }
}

fn main() {
  if describe(true) == "bad" && describe(false) == "none" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-nullable-struct-match-", source)
}

func TestCompilerTypedMatchOnNullableArrayUnionExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct IOError {}

fn unwrap(value: IOError | ?Array i32) -> ?Array i32 {
  value match {
    case _: IOError => nil
    case ok: ?Array i32 => ok
  }
}

fn main() {
  value: IOError | ?Array i32 = nil
  out := unwrap(value)
  if out == nil {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-nullable-array-union-match-", source)
}

func TestCompilerGenericTypedMatchOnNullableArrayUnionExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct IOError {}

fn unwrap<T>(value: IOError | T) -> T {
  value match {
    case _: IOError => nil
    case ok: T => ok
  }
}

fn main() {
  value: IOError | ?Array i32 = nil
  out := unwrap(value)
  if out == nil {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-generic-nullable-array-union-match-", source)
}
