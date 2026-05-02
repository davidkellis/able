package compiler

import (
	"strings"
	"testing"
)

func TestCompilerResultReturnUsesNativeCarrier(t *testing.T) {
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
		"fn value(ok: bool) -> !String {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_value(ok bool) (__able_union_") {
		t.Fatalf("expected result-returning function to use a native union carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_value")
	if !ok {
		t.Fatalf("could not find compiled value function")
	}
	if strings.Contains(body, "__able_any_to_value(") {
		t.Fatalf("expected result return path to avoid broad any conversion:\n%s", body)
	}
	if !strings.Contains(body, "runtime.ErrorValue{Message:") {
		t.Fatalf("expected concrete Error return to normalize into a runtime.ErrorValue member:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_impl_Error_message_") {
		t.Fatalf("expected concrete Error return to derive message through the compiled Error impl:\n%s", body)
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_value")
	if !ok {
		t.Fatalf("could not find value wrapper")
	}
	if !strings.Contains(wrapBody, "_to_value(rt, compiledResult)") {
		t.Fatalf("expected result wrapper to use explicit native union return conversion:\n%s", wrapBody)
	}
}

func TestCompilerResultPropagationUsesNativeCarrier(t *testing.T) {
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
		"fn value(ok: bool) -> !String {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  value(false)! or { err => err.message() }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(body, "__able_is_error(") {
		t.Fatalf("expected native result propagation to avoid runtime-value error probing:\n%s", body)
	}
	if !strings.Contains(body, "__able_raise_control(") {
		t.Fatalf("expected native result propagation to raise the error branch directly:\n%s", body)
	}
	if !strings.Contains(body, "_as_") {
		t.Fatalf("expected native result propagation/or-else to unwrap carrier helpers:\n%s", body)
	}
}

func TestCompilerResultPropagationErrorBindingStaysNativeErrorCarrier(t *testing.T) {
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
		"fn value(ok: bool) -> !String {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  value(false)! or { err => err.message() }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var err runtime.ErrorValue") {
		t.Fatalf("expected result propagation or-else binding to stay on the native error carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var err runtime.Value",
		"__able_method_call_node(",
		"__able_any_to_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected result propagation or-else binding to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerResultPropagationBlockErrorBindingStaysNativeErrorCarrier(t *testing.T) {
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
		"fn value(ok: bool) -> !String {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  do {",
		"    text := value(false)!",
		"    text",
		"  } or { err => err.message() }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var err runtime.ErrorValue") {
		t.Fatalf("expected block-wrapped result propagation or-else binding to stay on the native error carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var err runtime.Value",
		"__able_method_call_node(",
		"__able_any_to_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected block-wrapped result propagation or-else binding to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerResultPropagationExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct MyError { message: String }

impl Error for MyError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

fn value(ok: bool) -> !String {
  if ok { "ok" } else { MyError { message: "bad" } }
}

fn main() {
  handled := value(false)! or { err => err.message() }
  if handled == "bad" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-native-result-", source)
}

func TestCompilerNilPropagationReturnsNullableFromCurrentFunction(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn maybe_text(ok: bool) -> ?String {",
		"  if ok { \"value\" } else { nil }",
		"}",
		"",
		"fn marker(ok: bool) -> ?String {",
		"  maybe_text(ok)!",
		"  \"after\"",
		"}",
		"",
	}, "\n")
	result := compileNoFallbackSource(t, source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_marker")
	if !ok {
		t.Fatalf("could not find marker function")
	}
	if !strings.Contains(body, "return nil, nil") {
		t.Fatalf("expected nil propagation to return a normal nullable nil value:\n%s", body)
	}
	if strings.Contains(body, "__able_raise_control(nil, runtime.NilValue{})") {
		t.Fatalf("expected nil propagation not to lower as a raise control:\n%s", body)
	}

	compileAndRunSource(t, "ablec-nil-propagation-nullable-", `extern go fn __able_os_exit(code: i32) -> void {}

fn maybe_text(ok: bool) -> ?String {
  if ok { "value" } else { nil }
}

fn marker(ok: bool) -> ?String {
  maybe_text(ok)!
  "after"
}

fn main() {
  value := marker(false)
  handled := value or { "nil" }
  if handled == "nil" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`)
}

func TestCompilerNilPropagationReturnsNativeNullableFromCurrentFunction(t *testing.T) {
	compileAndRunSource(t, "ablec-nil-propagation-native-nullable-", `extern go fn __able_os_exit(code: i32) -> void {}

fn maybe_value(ok: bool) -> ?i32 {
  if ok { 7 } else { nil }
}

fn marker(ok: bool) -> ?i32 {
  maybe_value(ok)!
  99
}

fn main() {
  value := marker(false)
  handled := value or { -1 }
  if handled == -1 {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`)
}

func TestCompilerResultVoidReturnUsesNativeCarrier(t *testing.T) {
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
		"fn touch(ok: bool) -> !void {",
		"  if ok {",
		"    return",
		"  }",
		"  MyError { message: \"bad\" }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_touch(ok bool) (__able_union_") {
		t.Fatalf("expected !void return to use a native union carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_touch")
	if !ok {
		t.Fatalf("could not find compiled touch function")
	}
	for _, fragment := range []string{
		"runtime.VoidValue{}",
		"__able_any_to_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected !void return path to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerLocalErrorBindingDoesNotRaiseAsStatementResult(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"fn main() -> String {",
		"  err := MyError { message: \"cannot convert\" }",
		"  err.message",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(body, "__able_raise_control(nil, __able_tmp_") {
		t.Fatalf("expected local error binding statement to avoid implicit raise-on-discard:\n%s", body)
	}
	if !strings.Contains(body, "var err *MyError = &MyError{Message: \"cannot convert\"}") {
		t.Fatalf("expected local error binding to stay as a native struct binding:\n%s", body)
	}
}

func TestCompilerLocalErrorBindingExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"struct MyError { message: String }",
		"",
		"fn main() -> void {",
		"  err := MyError { message: \"cannot convert\" }",
		"  print(err.message)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-local-error-binding-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "cannot convert" {
		t.Fatalf("expected compiled local error binding program to print message, got %q", stdout)
	}
}

func TestCompilerVoidCallableCoercionPropagatesReturnedErrorCarrier(t *testing.T) {
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
		"fn fail() -> !i32 {",
		"  MyError { message: \"boom\" }",
		"}",
		"",
		"fn captures(cb: () -> void) -> bool {",
		"  handled := false",
		"  do { cb() } rescue {",
		"    case err: Error => {",
		"      handled = err.message() == \"boom\"",
		"      nil",
		"    }",
		"  }",
		"  handled",
		"}",
		"",
		"fn main() -> bool {",
		"  captures(fn() { fail() })",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_fn_void_to_struct___from_runtime_value(",
		"if errVal, ok, nilPtr := __able_runtime_error_value(result); ok || nilPtr {",
		"return struct{}{}, __able_raise_control(nil, errVal)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected void-callable coercion to re-raise returned Error carriers via the shared callable bridge (%q):\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerVoidCallableCoercionPropagatesReturnedErrorCarrierExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct MyError { message: String }

impl Error for MyError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

fn fail() -> !i32 {
  MyError { message: "boom" }
}

fn captures(cb: () -> void) -> bool {
  handled := false
  do { cb() } rescue {
    case err: Error => {
      handled = err.message() == "boom"
      nil
    }
  }
  handled
}

fn main() {
  if captures(fn() { fail() }) {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-void-callable-result-propagation-", source)
}

func TestCompilerDirectErrorReturnUsesNativeCarrier(t *testing.T) {
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
		"fn value() -> Error {",
		"  MyError { message: \"bad\" }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_value() (runtime.ErrorValue, *__ableControl)") {
		t.Fatalf("expected direct Error return to use runtime.ErrorValue carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_value")
	if !ok {
		t.Fatalf("could not find compiled value function")
	}
	if strings.Contains(body, "__able_any_to_value(") {
		t.Fatalf("expected direct Error return to avoid broad any conversion:\n%s", body)
	}
	if !strings.Contains(body, "runtime.ErrorValue{Message:") {
		t.Fatalf("expected direct Error return to normalize into runtime.ErrorValue:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_impl_Error_message_") {
		t.Fatalf("expected direct Error return to derive message through compiled Error impl:\n%s", body)
	}
}

func TestCompilerDirectErrorReturnExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct MyError { message: String }

impl Error for MyError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

fn value() -> Error {
  MyError { message: "bad" }
}

fn main() {
  if value().message() == "bad" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-native-error-return-", source)
}

func TestCompilerDirectErrorMessageAndCauseStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct RootError { message: String }",
		"struct OuterError { message: String, cause: ?Error }",
		"",
		"impl Error for RootError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"impl Error for OuterError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { self.cause }",
		"}",
		"",
		"fn value() -> Error {",
		"  OuterError { message: \"outer\", cause: RootError { message: \"root\" } }",
		"}",
		"",
		"fn describe() -> String {",
		"  parent := value()",
		"  inner := parent.cause()",
		"  if parent.message() == \"outer\" {",
		"    inner match {",
		"      case err: Error => err.message(),",
		"      case _ => \"none\"",
		"    }",
		"  } else {",
		"    \"bad\"",
		"  }",
		"}",
		"",
	}, "\n"))

	valueBody, ok := findCompiledFunction(result, "__able_compiled_fn_value")
	if !ok {
		t.Fatalf("could not find compiled value function")
	}
	if !strings.Contains(valueBody, "[\"cause\"] = __able_nullable_error_to_value(") {
		t.Fatalf("expected native error normalization to preserve cause payloads:\n%s", valueBody)
	}

	describeBody, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	if !strings.Contains(describeBody, ".Message") {
		t.Fatalf("expected direct Error.message() to lower to runtime.ErrorValue field access:\n%s", describeBody)
	}
	if !strings.Contains(describeBody, "Payload[\"cause\"]") {
		t.Fatalf("expected direct Error.cause() to lower to payload access:\n%s", describeBody)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_member_get_method(",
		"__able_call_value(",
	} {
		if strings.Contains(describeBody, fragment) {
			t.Fatalf("expected direct Error.message()/cause() to avoid %q:\n%s", fragment, describeBody)
		}
	}
}

func TestCompilerImplWrapperPrefersConcreteResultMethod(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface NumericConversions {",
		"  fn to_i64(self: Self) -> Result i64",
		"}",
		"",
		"struct BigBox {}",
		"",
		"methods BigBox {",
		"  fn to_i64(self: Self) -> Result i64 { 0 }",
		"}",
		"",
		"impl NumericConversions for BigBox {",
		"  fn to_i64(self: Self) -> Result i64 { self.to_i64() }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_method_BigBox_to_i64(") {
		t.Fatalf("expected concrete BigBox.to_i64 method to compile:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_impl_NumericConversions_to_i64_0")
	if !ok {
		t.Fatalf("could not find compiled impl wrapper")
	}
	if strings.Contains(body, ":= __able_compiled_impl_NumericConversions_to_i64_0(") {
		t.Fatalf("expected impl wrapper to avoid self-recursion:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_method_BigBox_to_i64(self)") {
		t.Fatalf("expected impl wrapper to forward to the concrete method:\n%s", body)
	}
}

func TestCompilerRaiseErrorImplementerUsesNativeErrorValue(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct RootError { code: i32 }",
		"",
		"impl Error for RootError {",
		"  fn message(self: Self) -> String { `root ${self.code}` }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn boom() -> String {",
		"  raise RootError { code: 7 }",
		"  \"ok\"",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_boom")
	if !ok {
		t.Fatalf("could not find compiled boom function")
	}
	if strings.Contains(body, "__able_any_to_value(&RootError") {
		t.Fatalf("expected raise to normalize Error implementers to runtime.ErrorValue:\n%s", body)
	}
	if !strings.Contains(body, "runtime.ErrorValue{Message:") {
		t.Fatalf("expected raise to materialize a native runtime.ErrorValue:\n%s", body)
	}
}

func TestCompilerDirectErrorCauseExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct RootError { message: String }
struct OuterError { message: String, cause: ?Error }

impl Error for RootError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

impl Error for OuterError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { self.cause }
}

fn value() -> Error {
  OuterError { message: "outer", cause: RootError { message: "root" } }
}

fn main() {
  parent := value()
  inner := parent.cause()
  outcome := if parent.message() == "outer" {
    inner match {
      case err: Error => err.message(),
      case _ => "none"
    }
  } else {
    "bad"
  }
  if outcome == "root" {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-native-error-cause-", source)
}

func TestCompilerStaticCallReturningConcreteErrorCoercesToNativeResult(t *testing.T) {
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
		"fn fail() -> MyError {",
		"  MyError { message: \"bad\" }",
		"}",
		"",
		"fn value(flag: bool) -> !i32 {",
		"  if flag { 7 } else { fail() }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_value")
	if !ok {
		t.Fatalf("could not find compiled value function")
	}
	if !strings.Contains(body, "__able_compiled_fn_fail()") {
		t.Fatalf("expected value to call the compiled fail helper:\n%s", body)
	}
	if !strings.Contains(body, "_wrap_runtime_ErrorValue(") {
		t.Fatalf("expected concrete error call result to be wrapped into the native result carrier:\n%s", body)
	}
	if strings.Contains(body, "__able_any_to_value(") {
		t.Fatalf("expected concrete error call result coercion to avoid any conversion:\n%s", body)
	}
}
