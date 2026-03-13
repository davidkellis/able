package compiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

const compilerDynamicBoundaryFixtureEnv = "ABLE_COMPILER_DYNAMIC_BOUNDARY_FIXTURES"

func TestCompilerDynamicBoundaryParityFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic boundary parity in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveCompilerDynamicBoundaryFixtures()
	if len(fixtures) == 0 {
		t.Skip("no dynamic-boundary fixtures configured")
	}
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			dir := filepath.Join(root, filepath.FromSlash(rel))
			manifest, err := interpreter.LoadFixtureManifest(dir)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if shouldSkipTarget(manifest.SkipTargets, "go") {
				return
			}
			if manifest.Expect.TypecheckDiagnostics != nil && len(manifest.Expect.TypecheckDiagnostics) > 0 {
				return
			}

			tree := runTreewalkerFixtureOutcome(t, dir, manifest)
			compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

			if tree.Exit != compiled.Exit {
				t.Fatalf("exit mismatch: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
			}
			if !reflect.DeepEqual(tree.Stdout, compiled.Stdout) {
				t.Fatalf("stdout mismatch: treewalker=%v compiled=%v", tree.Stdout, compiled.Stdout)
			}
			if !reflect.DeepEqual(tree.Stderr, compiled.Stderr) {
				t.Fatalf("stderr mismatch: treewalker=%v compiled=%v", tree.Stderr, compiled.Stderr)
			}
			if markers.FallbackCount != 0 {
				t.Fatalf("expected explicit dynamic boundary path (no fallback calls), got %d", markers.FallbackCount)
			}
			if strings.TrimSpace(markers.FallbackNames) != "" {
				t.Fatalf("expected no fallback call names for explicit boundary path, got %q", markers.FallbackNames)
			}
			if markers.ExplicitCount <= 0 {
				t.Fatalf("expected explicit dynamic boundary calls, got %d", markers.ExplicitCount)
			}
			if strings.TrimSpace(markers.ExplicitNames) == "" {
				t.Fatalf("expected explicit dynamic boundary call names, got empty")
			}
		})
	}
}

func TestCompilerDynamicBoundaryCallbackRoundtrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback parity in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_roundtrip\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_roundtrip",
		"",
		"dynimport exec.dynamic_cb_roundtrip.{apply_twice}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_roundtrip")!`,
		`pkg.def("fn apply_twice(f: i32 -> i32, value: i32) -> i32 { f(f(value)) }")!`,
		"",
		"fn main() -> void {",
		"  delta := 1",
		"  inc := fn(x: i32) -> i32 { x + delta }",
		"  print(`value ${apply_twice(inc, 40)}`)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"value 42"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for callback roundtrip boundary call, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_conversion_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_conversion_failure",
		"",
		"dynimport exec.dynamic_cb_typefail.{invoke_bad}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_typefail")!`,
		`pkg.def("fn invoke_bad(f) { f(\"oops\") }")!`,
		"",
		"fn main() -> void {",
		"  inc := fn(x: i32) -> i32 { x + 1 }",
		"  invoke_bad(inc)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for conversion-failure roundtrip, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackOverflowConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback overflow conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_overflow_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_overflow_failure",
		"",
		"dynimport exec.dynamic_cb_overflow.{invoke_overflow}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_overflow")!`,
		`pkg.def("fn invoke_overflow(f) { f(2147483648) }")!`,
		"",
		"fn main() -> void {",
		"  inc := fn(x: i32) -> i32 { x + 1 }",
		"  invoke_overflow(inc)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for overflow conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackUnsignedConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback unsigned conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_unsigned_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_unsigned_failure",
		"",
		"dynimport exec.dynamic_cb_unsigned.{invoke_negative}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_unsigned")!`,
		`pkg.def("fn invoke_negative(f) { f(-1) }")!`,
		"",
		"fn main() -> void {",
		"  clamp := fn(x: u8) -> u8 { x }",
		"  invoke_negative(clamp)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for unsigned conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackBoolConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback bool conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_bool_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_bool_failure",
		"",
		"dynimport exec.dynamic_cb_bool.{invoke_bad}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_bool")!`,
		`pkg.def("fn invoke_bad(f) { f(1) }")!`,
		"",
		"fn main() -> void {",
		"  pred := fn(flag: bool) -> bool { flag }",
		"  invoke_bad(pred)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for bool conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackStringConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback string conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_string_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_string_failure",
		"",
		"dynimport exec.dynamic_cb_string.{invoke_bad}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_string")!`,
		`pkg.def("fn invoke_bad(f) { f(true) }")!`,
		"",
		"fn main() -> void {",
		"  echo := fn(value: String) -> String { value }",
		"  invoke_bad(echo)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for string conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackStringConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback string conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_string_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_string_success",
		"",
		"dynimport exec.dynamic_cb_string_success.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_string_success")!`,
		`pkg.def("fn invoke(f) { f(\"able\") }")!`,
		"",
		"fn main() -> void {",
		"  echo := fn(value: String) -> String { `${value}!` }",
		"  print(invoke(echo))",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"able!"}
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
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for string conversion success, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackNilStringConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback nil->string conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_nil_string_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_nil_string_failure",
		"",
		"dynimport exec.dynamic_cb_nil_string.{invoke_bad}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_nil_string")!`,
		`pkg.def("fn invoke_bad(f) { f(nil) }")!`,
		"",
		"fn main() -> void {",
		"  echo := fn(value: String) -> String { value }",
		"  invoke_bad(echo)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for nil->string conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackCharConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback char conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_char_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_char_failure",
		"",
		"dynimport exec.dynamic_cb_char.{invoke_bad}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_char")!`,
		`pkg.def("fn invoke_bad(f) { f(\"ab\") }")!`,
		"",
		"fn main() -> void {",
		"  to_codepoint := fn(value: char) -> i32 { value as i32 }",
		"  invoke_bad(to_codepoint)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for char conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallbackStructConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback struct conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_struct_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_callback_struct_failure",
		"",
		"struct Payload { value: i32 }",
		"",
		"dynimport exec.dynamic_cb_struct.{invoke_bad}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_struct")!`,
		`pkg.def("fn invoke_bad(f) { f(nil) }")!`,
		"",
		"fn main() -> void {",
		"  use_payload := fn(item: Payload) -> i32 { item.value }",
		"  invoke_bad(use_payload)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_value"] <= 0 {
		t.Fatalf("expected call_value marker for struct conversion failure, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallNamedMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic call_named boundary in short mode")
	}
	withTypecheckFixturesOff(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_call_named\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_call_named",
		"",
		"dynimport exec.dynamic_boundary_call_named::dyn_pkg",
		"dyn_pkg_obj := dyn.def_package(\"exec.dynamic_boundary_call_named\")!",
		"",
		"fn main() -> void {",
		"  dyn.def_package(\"exec.dynamic_boundary_call_named\")!",
		"  missing_runtime_fn()",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure for unresolved named call: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if explicit["call_named:missing_runtime_fn"] <= 0 {
		t.Fatalf("expected call_named marker for missing_runtime_fn, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerDynamicBoundaryCallOriginalMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic call_original boundary in short mode")
	}
	withTypecheckFixturesOff(t)
	// This boundary test intentionally exercises an uncompileable function body so
	// wrapper emission can route through call_original markers.
	t.Setenv(compilerFixtureRequireNoFallbacksEnv, "0")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_call_original\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package exec_dynamic_boundary_call_original",
		"",
		"fn complex() -> i64 {",
		"  1 / 2",
		"}",
		"",
		"fn main() -> void {",
		"  complex()",
		"  print(\"ok\")",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{}
	manifest.Entry = "main.able"

	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure while exercising call_original path: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
	if markers.ExplicitCount <= 0 {
		t.Fatalf("expected explicit boundary calls, got %d", markers.ExplicitCount)
	}
	explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
	if !hasBoundaryMarkerPrefix(explicit, "call_original:") {
		t.Fatalf("expected call_original marker, markers=%q", markers.ExplicitNames)
	}
}

func TestCompilerTypedAssignReusesModuleBindingParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping typed assignment module binding parity in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: typed_assign_module_binding\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package typed_assign_module_binding",
		"",
		"value := 1",
		"",
		"fn update() -> i32 {",
		"  value: i32 = 2",
		"  value",
		"}",
		"",
		"fn main() -> void {",
		"  print(update())",
		"  print(value)",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"2", "2"}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if !reflect.DeepEqual(tree.Stderr, compiled.Stderr) {
		t.Fatalf("stderr mismatch: treewalker=%v compiled=%v", tree.Stderr, compiled.Stderr)
	}
	if markers.FallbackCount != 0 {
		t.Fatalf("expected no fallback calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
	}
}

func resolveCompilerDynamicBoundaryFixtures() []string {
	raw := strings.TrimSpace(os.Getenv(compilerDynamicBoundaryFixtureEnv))
	if raw == "" {
		return []string{
			"06_10_dynamic_metaprogramming_package_object",
			"13_04_import_alias_selective_dynimport",
			"13_05_dynimport_interface_dispatch",
			"13_07_search_path_env_override",
		}
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
	})
	fixtures := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		fixtures = append(fixtures, name)
	}
	return fixtures
}

type compilerBoundaryMarkers struct {
	FallbackCount int64
	FallbackNames string
	ExplicitCount int64
	ExplicitNames string
}

func runCompiledFixtureBoundaryOutcome(t *testing.T, dir string, manifest interpreter.FixtureManifest) (compilerFixtureOutcome, compilerBoundaryMarkers) {
	t.Helper()
	return runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{})
}

func runCompiledFixtureBoundaryOutcomeWithOptions(t *testing.T, dir string, manifest interpreter.FixtureManifest, opts Options) (compilerFixtureOutcome, compilerBoundaryMarkers) {
	t.Helper()
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)
	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("exec search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-dynamic-boundary")

	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	if !opts.RequireNoFallbacks {
		opts.RequireNoFallbacks = requireNoFallbacksForFixtureGates(t)
	}
	comp := New(opts)
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}
	harness := compilerHarnessSource(entryPath, searchPaths, manifest.Executor)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-fixture")
	if goruntime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath)
	env := withEnv(os.Environ(), "ABLE_COMPILER_BOUNDARY_MARKER", "1")
	env = withEnv(env, "ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE", "1")
	cmd.Env = applyFixtureEnv(env, manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("compiled fixture timed out after 60s")
	}
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run error: %v", runErr)
		}
	}

	stderrLines := splitLines(stderr.String())
	filtered, markers, err := extractBoundaryMarkers(stderrLines)
	if err != nil {
		t.Fatalf("boundary markers: %v; stderr=%q", err, stderr.String())
	}
	return compilerFixtureOutcome{
		Stdout: splitLines(stdout.String()),
		Stderr: filtered,
		Exit:   exitCode,
	}, markers
}

func extractBoundaryMarkers(lines []string) ([]string, compilerBoundaryMarkers, error) {
	filtered := make([]string, 0, len(lines))
	var markers compilerBoundaryMarkers
	var rawFallback string
	var rawExplicit string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "__ABLE_BOUNDARY_FALLBACK_CALLS="):
			rawFallback = strings.TrimPrefix(trimmed, "__ABLE_BOUNDARY_FALLBACK_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_BOUNDARY_FALLBACK_NAMES="):
			markers.FallbackNames = strings.TrimPrefix(trimmed, "__ABLE_BOUNDARY_FALLBACK_NAMES=")
		case strings.HasPrefix(trimmed, "__ABLE_BOUNDARY_EXPLICIT_CALLS="):
			rawExplicit = strings.TrimPrefix(trimmed, "__ABLE_BOUNDARY_EXPLICIT_CALLS=")
		case strings.HasPrefix(trimmed, "__ABLE_BOUNDARY_EXPLICIT_NAMES="):
			markers.ExplicitNames = strings.TrimPrefix(trimmed, "__ABLE_BOUNDARY_EXPLICIT_NAMES=")
		default:
			filtered = append(filtered, line)
		}
	}
	if rawFallback == "" {
		return nil, compilerBoundaryMarkers{}, fmt.Errorf("missing __ABLE_BOUNDARY_FALLBACK_CALLS marker")
	}
	if rawExplicit == "" {
		return nil, compilerBoundaryMarkers{}, fmt.Errorf("missing __ABLE_BOUNDARY_EXPLICIT_CALLS marker")
	}
	fallbackCount, err := strconv.ParseInt(rawFallback, 10, 64)
	if err != nil {
		return nil, compilerBoundaryMarkers{}, fmt.Errorf("invalid fallback count %q: %w", rawFallback, err)
	}
	explicitCount, err := strconv.ParseInt(rawExplicit, 10, 64)
	if err != nil {
		return nil, compilerBoundaryMarkers{}, fmt.Errorf("invalid explicit count %q: %w", rawExplicit, err)
	}
	markers.FallbackCount = fallbackCount
	markers.ExplicitCount = explicitCount
	return filtered, markers, nil
}

func parseBoundaryMarkerSnapshot(raw string) map[string]int64 {
	out := map[string]int64{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		out[key] = parsed
	}
	return out
}

func hasBoundaryMarkerPrefix(markers map[string]int64, prefix string) bool {
	if strings.TrimSpace(prefix) == "" {
		return false
	}
	for key, value := range markers {
		if value <= 0 {
			continue
		}
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func withTypecheckFixturesOff(t *testing.T) {
	t.Helper()
	prev, hadPrev := os.LookupEnv("ABLE_TYPECHECK_FIXTURES")
	if err := os.Setenv("ABLE_TYPECHECK_FIXTURES", "off"); err != nil {
		t.Fatalf("set ABLE_TYPECHECK_FIXTURES: %v", err)
	}
	t.Cleanup(func() {
		if hadPrev {
			_ = os.Setenv("ABLE_TYPECHECK_FIXTURES", prev)
			return
		}
		_ = os.Unsetenv("ABLE_TYPECHECK_FIXTURES")
	})
}
