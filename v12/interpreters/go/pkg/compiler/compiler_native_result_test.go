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

func TestCompilerDirectErrorCauseExecutes(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct RootError { message: String }
struct OuterError { message: String, cause: ?Error }

impl Error for RootError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
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
