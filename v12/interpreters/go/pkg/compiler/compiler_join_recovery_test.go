package compiler

import (
	"strings"
	"testing"
)

func TestCompilerIfJoinRecoversTypeExprBackedInterfaceCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct First {}",
		"struct Second {}",
		"",
		"impl Reader i32 for First { fn read(self: Self) -> i32 { 1 } }",
		"impl Reader i32 for Second { fn read(self: Self) -> i32 { 2 } }",
		"",
		"fn fail() -> First {",
		"  raise(First {})",
		"  First {}",
		"}",
		"",
		"fn main() -> i32 {",
		"  recovered := fail() rescue {",
		"    case reader: Reader i32 =>",
		"      if true { reader } else { Second {} }",
		"  }",
		"  recovered.read()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var recovered __able_iface_Reader_i32") {
		t.Fatalf("expected typed-pattern if-join to recover the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected typed-pattern if-join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerIfJoinRecoversTypeExprBackedErrorCarrier(t *testing.T) {
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
		"  recovered := fail() rescue {",
		"    case err: Error =>",
		"      if true { err } else { MyError { message: \"fallback\" } }",
		"  }",
		"  recovered.message()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var recovered runtime.ErrorValue") {
		t.Fatalf("expected typed-pattern if-join to recover the native error carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected typed-pattern error join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRescueIdentifierJoinRecoversPropagatedErrorCarrier(t *testing.T) {
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
		"  if ok { \"ok\" } else { MyError { message: \"boom\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  do {",
		"    value(false)!",
		"  } rescue {",
		"    case err => {",
		"      recovered_err := if true { err } else { MyError { message: \"fallback\" } }",
		"      recovered_err.message()",
		"    }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{"var err runtime.ErrorValue", "var recovered_err runtime.ErrorValue"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue identifier join to recover the native error carrier (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"var err runtime.Value",
		"var recovered_err runtime.Value",
		"var recovered_err __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue identifier join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRescueIdentifierJoinRecoversNonTailPropagatedErrorCarrier(t *testing.T) {
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
		"  if ok { \"ok\" } else { MyError { message: \"boom\" } }",
		"}",
		"",
		"fn main() -> String {",
		"  do {",
		"    text := value(false)!",
		"    text",
		"  } rescue {",
		"    case err => {",
		"      recovered_err := if true { err } else { MyError { message: \"fallback\" } }",
		"      recovered_err.message()",
		"    }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{"var err runtime.ErrorValue", "var recovered_err runtime.ErrorValue"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected non-tail propagated rescue identifier join to recover the native error carrier (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"var err runtime.Value",
		"var recovered_err runtime.Value",
		"var recovered_err __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected non-tail propagated rescue identifier join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRescueIdentifierJoinRecoversPropagatedCallableCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> (() -> i32) {",
		"  raise(fn() -> i32 { 1 })",
		"  fn() -> i32 { 0 }",
		"}",
		"",
		"fn main() -> i32 {",
		"  do {",
		"    fail()",
		"    0",
		"  } rescue {",
		"    case cb => {",
		"      recovered := if true { cb } else { fn() -> i32 { 2 } }",
		"      recovered()",
		"    }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{"var cb __able_fn_void_to_int32", "var recovered __able_fn_void_to_int32"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue identifier join to recover the native callable carrier (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"var cb runtime.Value",
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue callable join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRescueIdentifierJoinRecoversPropagatedImportedShadowedNominalCarrier(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn fail() -> RemoteThing {",
			"  raise(RemoteThing { remote: 1 })",
			"  RemoteThing { remote: 0 }",
			"}",
			"",
			"fn main() -> i32 {",
			"  do {",
			"    fail()",
			"    0",
			"  } rescue {",
			"    case thing => {",
			"      recovered := if true { thing } else { RemoteThing { remote: 2 } }",
			"      recovered.remote",
			"    }",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{"var thing *Thing_a", "var recovered *Thing_a", "recovered.Remote"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue identifier join to recover the imported native struct carrier (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"var thing runtime.Value",
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_member_get(",
		"__able_struct_Thing_from(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue imported shadowed nominal join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRescueIdentifierJoinRecoversPropagatedImportedShadowedInterfaceCarrier(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First, Second}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn fail() -> RemoteReader i32 {",
			"  raise(First {})",
			"  First {}",
			"}",
			"",
			"fn main() -> i32 {",
			"  do {",
			"    fail()",
			"    0",
			"  } rescue {",
			"    case reader => {",
			"      recovered := if true { reader } else { Second {} }",
			"      recovered.read()",
			"    }",
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
			"struct Second {}",
			"",
			"impl Reader i32 for First { fn read(self: Self) -> i32 { 1 } }",
			"impl Reader i32 for Second { fn read(self: Self) -> i32 { 2 } }",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var reader __able_iface_") || !strings.Contains(body, "var recovered __able_iface_") {
		t.Fatalf("expected propagated rescue identifier join to recover the imported native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var reader runtime.Value",
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected propagated rescue imported shadowed interface join to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "recovered.read()") {
		t.Fatalf("expected propagated rescue imported shadowed interface join to keep direct native interface dispatch:\n%s", body)
	}
}
