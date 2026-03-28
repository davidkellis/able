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
