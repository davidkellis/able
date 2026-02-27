package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerDynamicBoundaryMonoArrayCallbackConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_callback_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_callback_success",
		"",
		"dynimport exec.dynamic_cb_array_mono_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_mono_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: Array i32) -> i32 { values[0]! + values[1]! }",
		"  nums: Array i32 := [4, 5]",
		"  print(invoke(callback, nums))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"9"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayIntoDynamicInterpreterSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array interpreter boundary success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_dynamic_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_dynamic_success",
		"",
		"dynimport exec.dynamic_array_consumer_mono.{sum_ends,id}",
		"",
		`pkg := dyn.def_package("exec.dynamic_array_consumer_mono")!`,
		`pkg.def("fn sum_ends(values) { values[0]! + values[values.len() - 1]! }")!`,
		`pkg.def("fn id(values) { values }")!`,
		"",
		"fn main() -> void {",
		"  nums: Array i32 := [4, 5, 6]",
		"  print(sum_ends(nums))",
		"  roundtrip: Array i32 := id(nums)",
		"  print(roundtrip[0]! + roundtrip[2]!)",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"10", "10"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] < 2 {
		t.Fatalf("expected call_value markers for mono array dynamic calls, markers=%q", markers.ExplicitNames)
	}
}
func TestCompilerDynamicBoundaryMonoArrayCallbackNullableConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback nullable conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_nullable_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_nullable_success",
		"",
		"dynimport exec.dynamic_cb_array_nullable_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_nullable_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: nil | Array i32) -> i32 {",
		"    values match {",
		"      case nil => 0,",
		"      case xs: Array i32 => xs[0]! + xs[1]!",
		"    }",
		"  }",
		"  print(invoke(callback, nil))",
		"  nums: Array i32 := [4, 5]",
		"  print(invoke(callback, nums))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"0", "9"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayCallbackNullableConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback nullable conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_nullable_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_nullable_failure",
		"",
		"dynimport exec.dynamic_cb_array_nullable_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_nullable_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: nil | Array i32) -> i32 {",
		"    values match {",
		"      case nil => 0,",
		"      case xs: Array i32 => xs[0]! + xs[1]!",
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
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayCallbackUnionConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback union conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_union_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_union_success",
		"",
		"dynimport exec.dynamic_cb_array_union_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_union_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: Array i32 | String) -> i32 {",
		"    value match {",
		"      case xs: Array i32 => xs[0]! + xs[1]!",
		"      case s: String => 2",
		"    }",
		"  }",
		"  nums: Array i32 := [7, 8]",
		"  print(invoke(callback, nums))",
		"  print(invoke(callback, \"ok\"))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"15", "2"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayCallbackUnionConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback union conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_union_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_union_failure",
		"",
		"dynimport exec.dynamic_cb_array_union_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_union_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: Array i32 | String) -> i32 {",
		"    value match {",
		"      case xs: Array i32 => xs[0]! + xs[1]!",
		"      case s: String => 2",
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
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayCallbackInterfaceConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback interface conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_interface_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_interface_success",
		"",
		"interface Summable for Self {",
		"  fn sum(self: Self) -> i32",
		"}",
		"",
		"struct Numbers { values: Array i32 }",
		"",
		"impl Summable for Numbers {",
		"  fn sum(self: Self) -> i32 { self.values[0]! + self.values[1]! }",
		"}",
		"",
		"dynimport exec.dynamic_cb_array_interface_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_interface_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: Summable) -> i32 { value.sum() }",
		"  print(invoke(callback, Numbers { values: [10, 11] }))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"21"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayCallbackInterfaceConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array callback interface conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_interface_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_interface_failure",
		"",
		"interface Summable for Self {",
		"  fn sum(self: Self) -> i32",
		"}",
		"",
		"dynimport exec.dynamic_cb_array_interface_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_interface_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(value: Summable) -> i32 { value.sum() }",
		"  invoke(callback, \"Ada\")",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}
