package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRescueTypedPatternBindingStaysNativeInterface(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct First {}",
		"",
		"impl Reader i32 for First { fn read(self: Self) -> i32 { 1 } }",
		"",
		"fn fail() -> First {",
		"  raise(First {})",
		"  First {}",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case reader: Reader i32 => reader.read()",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var reader __able_iface_Reader_i32 =") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var reader runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_iface_Reader_i32_try_from_value(__able_runtime,") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through the native interface matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternBindingStaysNativeError(t *testing.T) {
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
		"  raise(MyError { message: \"boom\" })",
		"  MyError { message: \"ok\" }",
		"}",
		"",
		"fn main() -> String {",
		"  fail() rescue {",
		"    case err: Error => err.message()",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var err runtime.ErrorValue =") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native error carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var err runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_is_error(") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through direct native error detection:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternBindingStaysNativeI32(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> i32 {",
		"  raise(41)",
		"  0",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case n: i32 => n + 1",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var n int32 =") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native i32 carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var n runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "runtime.IntegerType(\"i32\")") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through the native integer runtime type check:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternBindingStaysNativeNullableI32(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> i32 {",
		"  raise(41)",
		"  0",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case maybe: ?i32 => maybe match {",
		"      case n: i32 => n + 1,",
		"      case nil => 0",
		"    }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var maybe *int32 =") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native nullable i32 carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var maybe runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_nullable_i32_from_value(") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through the native nullable matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternBindingStaysNativeCallable(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> (i32 -> i32) {",
		"  raise(fn(value: i32) -> i32 { value + 1 })",
		"  fn(value: i32) -> i32 { value }",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case cb: (i32 -> i32) => cb(41)",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var cb __able_fn_int32_to_int32 =") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native callable carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var cb runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_fn_int32_to_int32_try_from_runtime_value(__able_runtime,") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through the native callable matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternBindingStaysNativeStruct(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> DivMod i32 {",
		"  raise(7 /% 3)",
		"  0 /% 1",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case pair: DivMod i32 => pair.remainder",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var pair *DivMod_i32 =") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native struct carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var pair runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_struct_DivMod_i32_try_from(") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through the native struct matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternBindingStaysNativeUnion(t *testing.T) {
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
		"type ResultText = Error | String",
		"",
		"fn fail() -> String {",
		"  raise(MyError { message: \"boom\" })",
		"  \"ok\"",
		"}",
		"",
		"fn main() -> String {",
		"  fail() rescue {",
		"    case result: ResultText =>",
		"      result match {",
		"        case err: Error => err.message(),",
		"        case text: String => text",
		"      }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var result __able_union_") {
		t.Fatalf("expected rescue typed-pattern binding to stay on the native union carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var result runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern binding to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "_try_from_value(__able_runtime,") {
		t.Fatalf("expected rescue typed-pattern binding to narrow through the native union matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternConditionStaysNativeString(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> String {",
		"  raise(\"boom\")",
		"  \"ok\"",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case _: String => 41",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern condition to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "case runtime.StringValue:") {
		t.Fatalf("expected rescue typed-pattern condition to narrow through the native string runtime type check:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternConditionStaysNativeNullableI32(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> i32 {",
		"  raise(41)",
		"  0",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case _: ?i32 => 41",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern condition to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_nullable_i32_from_value(") {
		t.Fatalf("expected rescue typed-pattern condition to narrow through the native nullable matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternConditionStaysNativeSingleton(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Ready {}",
		"",
		"fn fail() -> Ready {",
		"  raise(Ready)",
		"  Ready",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case _: Ready => 41",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern condition to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_struct_Ready_try_from(") {
		t.Fatalf("expected rescue typed-pattern condition to narrow through the native singleton matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternConditionStaysNativeUnion(t *testing.T) {
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
		"type ResultText = Error | String",
		"",
		"fn fail() -> String {",
		"  raise(MyError { message: \"boom\" })",
		"  \"ok\"",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case _: ResultText => 41",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern condition to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "_try_from_value(__able_runtime,") {
		t.Fatalf("expected rescue typed-pattern condition to narrow through the native union matcher helper:\n%s", body)
	}
}

func TestCompilerRescueTypedPatternConditionStaysNativeCallable(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> (i32 -> i32) {",
		"  raise(fn(value: i32) -> i32 { value + 1 })",
		"  fn(value: i32) -> i32 { value }",
		"}",
		"",
		"fn main() -> i32 {",
		"  fail() rescue {",
		"    case _: (i32 -> i32) => 41",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected rescue typed-pattern condition to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_fn_int32_to_int32_try_from_runtime_value(__able_runtime,") {
		t.Fatalf("expected rescue typed-pattern condition to narrow through the native callable matcher helper:\n%s", body)
	}
}

func TestCompilerNativeUnionTypedPatternWholeValueBindingUsesNativeInterfaceCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct First {}",
		"",
		"impl Reader i32 for First { fn read(self: Self) -> i32 { 1 } }",
		"",
		"union Item = First | String",
		"",
		"fn main() -> i32 {",
		"  item: Item = First {}",
		"  item match {",
		"    case reader: Reader i32 => reader.read(),",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var reader *First =") {
		t.Fatalf("expected native-union typed whole-value binding to stay on a native carrier without runtime boxing:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_impl_Reader_read_0_spec(reader)") {
		t.Fatalf("expected native-union typed whole-value binding to dispatch through the compiled Reader impl:\n%s", body)
	}
	if strings.Contains(body, "var reader runtime.Value") {
		t.Fatalf("expected native-union typed whole-value binding to avoid runtime.Value local:\n%s", body)
	}
}
