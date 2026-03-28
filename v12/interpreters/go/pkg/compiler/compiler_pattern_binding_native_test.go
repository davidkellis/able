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
	if strings.Contains(body, "var reader runtime.Value") {
		t.Fatalf("expected rescue typed-pattern binding to avoid runtime.Value local:\n%s", body)
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
	if strings.Contains(body, "var err runtime.Value") {
		t.Fatalf("expected rescue typed-pattern binding to avoid runtime.Value local:\n%s", body)
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
