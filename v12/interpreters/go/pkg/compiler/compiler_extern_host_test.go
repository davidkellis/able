package compiler

import (
	"strings"
	"testing"
)

func TestCompilerGoExternCallStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"prelude go {",
		`import "os"`,
		"}",
		"",
		"extern go fn temp_dir() -> String {",
		"  return os.TempDir()",
		"}",
		"",
		"fn main() -> String {",
		"  temp_dir()",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("main body not found")
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value(", "__able_method_call_node(", "__able_member_get_method("} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected ordinary go extern static call to avoid %q:\n%s", fragment, mainBody)
		}
	}
	if !strings.Contains(mainBody, "__able_compiled_fn_temp_dir(") {
		t.Fatalf("expected main to call compiled extern wrapper directly:\n%s", mainBody)
	}
	if !strings.Contains(string(result.Files["compiled.go"]), "func __able_host_fn_temp_dir()") {
		t.Fatalf("expected compiled host extern function to be rendered")
	}
}

func TestCompilerGoExternStructBoundaryExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-go-extern-struct-boundary-", strings.Join([]string{
		"package demo",
		"",
		"struct Box {",
		"  value: i32",
		"}",
		"",
		"extern go fn increment(box: Box) -> Box {",
		"  raw, ok := box.(map[string]any)",
		"  if !ok {",
		`    return map[string]any{"value": int32(-1)}`,
		"  }",
		"  current, _ := raw[\"value\"].(int32)",
		`  return map[string]any{"value": current + 1}`,
		"}",
		"",
		"fn main() -> void {",
		"  box := increment(Box { value: 41 })",
		"  print(box.value)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "42" {
		t.Fatalf("expected extern struct boundary output 42, got %q", stdout)
	}
}

func TestCompilerGoExternUnionMonoArrayToRuntimeHelperUsesSpecializedArrayBridge(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-go-extern-array-union-", strings.Join([]string{
		"package demo",
		"",
		"struct BlobError {",
		"  message: String,",
		"}",
		"",
		"extern go fn read_bytes() -> BlobError | ?Array u8 {",
		"  return []uint8{65, 66}",
		"}",
		"",
		"fn main() -> void {",
		"  read_bytes() match {",
		"    case bytes: Array u8 => print(bytes.length()),",
		"    case _ => print(-1)",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName:              "main",
		ExperimentalMonoArrays:   true,
		RequireStaticNoFallbacks: true,
	})

	compiled := string(result.Files["compiled.go"])
	if !strings.Contains(compiled, "return __able_array_u8_to(rt, raw.Value)") {
		t.Fatalf("expected mono-array union runtime bridge to use __able_array_u8_to:\n%s", compiled)
	}
	if strings.Contains(compiled, "unsupported union member type *__able_array_u8") {
		t.Fatalf("expected mono-array union runtime bridge to avoid unsupported member fallback")
	}
}

func TestCompilerGoExternStructBoundaryWithNamedUnionFieldExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-go-extern-union-field-boundary-", strings.Join([]string{
		"package demo",
		"",
		"struct NotFound {}",
		"struct Other {}",
		"",
		"union IOErrorKind = NotFound | Other",
		"",
		"struct IOError {",
		"  kind: IOErrorKind,",
		"  message: String,",
		"  path: ?String,",
		"}",
		"",
		"extern go fn fail() -> IOError | void {",
		`  return map[string]any{"kind": "NotFound", "message": "missing", "path": nil}`,
		"}",
		"",
		"fn main() -> void {",
		"  fail() match {",
		"    case err: IOError => print(err.message),",
		"    case _ => print(\"ok\")",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "missing" {
		t.Fatalf("expected extern named-union field output missing, got %q", stdout)
	}
}

func TestCompilerGoExternGenericUnwrapExecutesForNativeSuccessMembers(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-go-extern-generic-unwrap-", strings.Join([]string{
		"package demo",
		"",
		"prelude go {",
		`import "errors"`,
		"}",
		"",
		"struct IOError {",
		"  message: String,",
		"}",
		"",
		"impl Error for IOError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"extern go fn write_count() -> IOError | i32 {",
		"  return int32(5)",
		"}",
		"",
		"extern go fn read_none() -> IOError | ?Array u8 {",
		"  _ = errors.New(\"unused\")",
		"  return nil",
		"}",
		"",
		"fn unwrap<T>(value: IOError | T) -> T {",
		"  value match {",
		"    case err: IOError => { raise err },",
		"    case ok: T => ok",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  count := unwrap(write_count())",
		"  maybe := unwrap(read_none())",
		"  print(count)",
		"  print(maybe == nil)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "5\ntrue" {
		t.Fatalf("expected extern generic unwrap output 5/true, got %q", stdout)
	}
}

func TestCompilerGoExternGenericUnwrapPreservesNullableReturnTypeExprs(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-go-extern-generic-unwrap-nullable-return-", strings.Join([]string{
		"package demo",
		"",
		"struct IOError {",
		"  message: String,",
		"}",
		"",
		"impl Error for IOError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"extern go fn read_none() -> IOError | ?Array u8 {",
		"  return nil",
		"}",
		"",
		"fn unwrap<T>(value: IOError | T) -> T {",
		"  value match {",
		"    case err: IOError => { raise err },",
		"    case ok: T => ok",
		"  }",
		"}",
		"",
		"fn read_implicit() -> ?Array u8 {",
		"  unwrap(read_none())",
		"}",
		"",
		"fn read_explicit() -> ?Array u8 {",
		"  return unwrap(read_none())",
		"}",
		"",
		"fn main() -> void {",
		"  print(read_implicit() == nil)",
		"  print(read_explicit() == nil)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "true\ntrue" {
		t.Fatalf("expected nullable unwrap return output true/true, got %q", stdout)
	}
}

func TestCompilerGoExternUnionReturnExecutesThroughAnyCarrier(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-go-extern-union-any-return-", strings.Join([]string{
		"package demo",
		"",
		"struct IOError {",
		"  message: String,",
		"}",
		"",
		"impl Error for IOError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"extern go fn maybe_name(ok: bool) -> IOError | String {",
		"  if ok {",
		"    return \"open\"",
		"  }",
		"  return map[string]any{ \"message\": \"closed\" }",
		"}",
		"",
		"fn main() -> void {",
		"  maybe_name(true) match {",
		"    case name: String => print(name),",
		"    case err: IOError => print(err.message)",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "open" {
		t.Fatalf("expected extern union return output open, got %q", stdout)
	}
}
