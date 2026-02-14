package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerDynamicBoundaryCallbackInterfaceConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback interface conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_interface_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_interface_success",
		"",
		"interface Greeter for Self {",
		"  fn greet(self: Self) -> String",
		"}",
		"",
		"struct Person { name: String }",
		"",
		"impl Greeter for Person {",
		"  fn greet(self: Self) -> String { `hi ${self.name}` }",
		"}",
		"",
		"dynimport exec.dynamic_cb_iface_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_iface_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: Greeter) -> String { value.greet() }",
		"  print(invoke(callback, Person { name: \"Ada\" }))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"hi Ada"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryCallbackInterfaceConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback interface conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_interface_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_interface_failure",
		"",
		"interface Greeter for Self {",
		"  fn greet(self: Self) -> String",
		"}",
		"",
		"dynimport exec.dynamic_cb_iface_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_iface_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: Greeter) -> String { value.greet() }",
		"  invoke(callback, \"Ada\")",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryCallbackUnionConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback union conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_union_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_union_success",
		"",
		"dynimport exec.dynamic_cb_union_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_union_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: i32 | String) -> String {",
		"    value match {",
		"      case v: i32 => `i ${v}`,",
		"      case v: String => `s ${v}`",
		"    }",
		"  }",
		"  print(invoke(callback, 7))",
		"  print(invoke(callback, \"ok\"))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"i 7", "s ok"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryCallbackUnionConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback union conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_union_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_union_failure",
		"",
		"dynimport exec.dynamic_cb_union_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_union_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: i32 | String) -> String {",
		"    value match {",
		"      case v: i32 => `i ${v}`,",
		"      case v: String => `s ${v}`",
		"    }",
		"  }",
		"  invoke(callback, true)",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryCallbackNullableConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback nullable conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_nullable_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_nullable_success",
		"",
		"dynimport exec.dynamic_cb_nullable_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_nullable_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: ?String) -> String {",
		"    value match {",
		"      case nil => \"nil\",",
		"      case v: String => `s ${v}`",
		"    }",
		"  }",
		"  print(invoke(callback, nil))",
		"  print(invoke(callback, \"ok\"))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"nil", "s ok"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryCallbackNullableConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback nullable conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_nullable_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_nullable_failure",
		"",
		"dynimport exec.dynamic_cb_nullable_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_nullable_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: ?String) -> String {",
		"    value match {",
		"      case nil => \"nil\",",
		"      case v: String => `s ${v}`",
		"    }",
		"  }",
		"  invoke(callback, true)",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func assertBoundaryCallValueMarkers(t *testing.T, markers compilerBoundaryMarkers) {
	t.Helper()
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker, markers=%q", markers.ExplicitNames)
	}
}

func joinLines(lines ...string) string {
	out := ""
	for i, line := range lines {
		if i > 0 {
			out += "\n"
		}
		out += line
	}
	return out
}

