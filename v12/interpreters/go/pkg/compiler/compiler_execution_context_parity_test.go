package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerExperimentalExecutionContextFixtureParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping experimental execution-context fixture parity in short mode")
	}
	root := compilerExecutionContextFixtureRoot(t)
	for _, rel := range []string{
		"06_01_compiler_map_literal",
		"07_01_function_definition_generics_inference",
		"10_04_interface_dispatch_defaults_generics",
		"11_03_rescue_ensure",
		"13_03_package_config_prelude",
		"16_01_host_interop_inline_extern",
		"06_12_22_stdlib_io_temp",
		"06_01_compiler_spawn_await",
		"12_02_future_fairness_cancellation",
		"12_05_concurrency_channel_ping_pong",
		"12_09_nested_spawn_native_context",
		"15_04_background_work_flush",
	} {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			dir := filepath.Join(root, filepath.FromSlash(rel))
			manifest, err := interpreter.LoadFixtureManifest(dir)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if shouldSkipTarget(manifest.SkipTargets, "go") || fixtureHasTypecheckDiagnostics(manifest) {
				return
			}
			tree := runTreewalkerFixtureOutcome(t, dir, manifest)
			compiled := runCompiledFixtureOutcomeWithOptions(t, dir, manifest, Options{
				PackageName:                  "main",
				ExperimentalExecutionContext: true,
			})
			assertCompilerFixtureOutcomeParity(t, tree, compiled)
		})
	}
}

func TestCompilerExperimentalExecutionContextDynamicBoundaryParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping experimental execution-context dynamic-boundary parity in short mode")
	}
	root := compilerExecutionContextFixtureRoot(t)
	for _, rel := range []string{
		"06_10_dynamic_metaprogramming_package_object",
		"13_04_import_alias_selective_dynimport",
		"13_05_dynimport_interface_dispatch",
		"13_07_search_path_env_override",
	} {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			dir := filepath.Join(root, filepath.FromSlash(rel))
			manifest, err := interpreter.LoadFixtureManifest(dir)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if shouldSkipTarget(manifest.SkipTargets, "go") || fixtureHasTypecheckDiagnostics(manifest) {
				return
			}
			tree := runTreewalkerFixtureOutcome(t, dir, manifest)
			compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
				PackageName:                  "main",
				ExperimentalExecutionContext: true,
			})
			assertCompilerFixtureOutcomeParity(t, tree, compiled)
			if markers.FallbackCount != 0 || strings.TrimSpace(markers.FallbackNames) != "" {
				t.Fatalf("expected no fallback boundary calls, got count=%d names=%q", markers.FallbackCount, markers.FallbackNames)
			}
			if markers.ExplicitCount <= 0 || strings.TrimSpace(markers.ExplicitNames) == "" {
				t.Fatalf("expected explicit dynamic boundary calls, got count=%d names=%q", markers.ExplicitCount, markers.ExplicitNames)
			}
		})
	}
}

func TestCompilerExperimentalExecutionContextDynamicNamedAndValueBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping experimental execution-context dynamic boundary probes in short mode")
	}
	withTypecheckFixturesOff(t)
	for _, tc := range []struct {
		name           string
		source         string
		wantExit       int
		wantStdout     []string
		wantMarkerName string
	}{
		{
			name: "value",
			source: strings.Join([]string{
				"package exec_dynamic_context_value",
				"",
				"dynimport exec.dynamic_context_value.{apply_twice}",
				`pkg := dyn.def_package("exec.dynamic_context_value")!`,
				`pkg.def("fn apply_twice(f: i32 -> i32, value: i32) -> i32 { f(f(value)) }")!`,
				"",
				"fn main() -> void {",
				"  delta := 1",
				"  inc := fn(x: i32) -> i32 { x + delta }",
				"  print(`value ${apply_twice(inc, 40)}`)",
				"}",
				"",
			}, "\n"),
			wantStdout:     []string{"value 42"},
			wantMarkerName: "call_value",
		},
		{
			name: "named",
			source: strings.Join([]string{
				"package exec_dynamic_context_named",
				"",
				"dynimport exec.dynamic_context_named::dyn_pkg",
				`dyn_pkg_obj := dyn.def_package("exec.dynamic_context_named")!`,
				"",
				"fn main() -> void {",
				`  dyn.def_package("exec.dynamic_context_named")!`,
				"  missing_runtime_fn()",
				"}",
				"",
			}, "\n"),
			wantExit:       1,
			wantMarkerName: "call_named:missing_runtime_fn",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: "+tc.name+"\n"), 0o600); err != nil {
				t.Fatalf("write package.yml: %v", err)
			}
			if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(tc.source), 0o600); err != nil {
				t.Fatalf("write source: %v", err)
			}
			manifest := interpreter.FixtureManifest{Entry: "main.able"}
			tree := runTreewalkerFixtureOutcome(t, dir, manifest)
			compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
				PackageName:                  "main",
				ExperimentalExecutionContext: true,
			})
			assertCompilerFixtureOutcomeParity(t, tree, compiled)
			if tc.wantExit != 0 && compiled.Exit != tc.wantExit {
				t.Fatalf("exit = %d, want %d", compiled.Exit, tc.wantExit)
			}
			if tc.wantStdout != nil && !reflect.DeepEqual(compiled.Stdout, tc.wantStdout) {
				t.Fatalf("stdout = %v, want %v", compiled.Stdout, tc.wantStdout)
			}
			if markers.FallbackCount != 0 {
				t.Fatalf("expected no fallback boundary calls, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
			}
			explicit := parseBoundaryMarkerSnapshot(markers.ExplicitNames)
			if explicit[tc.wantMarkerName] <= 0 {
				t.Fatalf("expected %q dynamic marker, got %q", tc.wantMarkerName, markers.ExplicitNames)
			}
		})
	}
}

func TestCompilerExperimentalExecutionContextBoundMethodValueCarriesContext(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-exec-context-bound-method", strings.Join([]string{
		"package demo",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #add(delta: i32) -> i32 { #value + delta }",
		"}",
		"",
		"fn main() -> i32 {",
		"  counter := Counter { value: 10 }",
		"  add_fn := counter.add",
		"  handle := spawn { add_fn(5) }",
		"  await [handle] as i32",
		"}",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main_ctx")
	if !strings.Contains(mainBody, "__able_compiled_method_Counter_add_ctx(") {
		t.Fatalf("bound method value should carry the caller execution context:\n%s", mainBody)
	}
	if strings.Contains(mainBody, "__able_compiled_method_Counter_add(") {
		t.Fatalf("context-aware bound method value must not re-enter the compatibility body:\n%s", mainBody)
	}
	methodWrapper, ok := findCompiledDeclByPrefix(result, "func __able_wrap_method_Counter_add")
	if !ok {
		t.Fatalf("could not find Counter.add native wrapper")
	}
	if !strings.Contains(methodWrapper, "__able_wrap_method_Counter_add_ctx_direct(rt, ctx, args[0], args[1:])") {
		t.Fatalf("native bound-method wrapper must preserve its native call context:\n%s", methodWrapper)
	}
	contextDirectWrapper, ok := findCompiledDeclByPrefix(result, "func __able_wrap_method_Counter_add_ctx_direct")
	if !ok {
		t.Fatalf("could not find Counter.add context-aware split-receiver wrapper")
	}
	if !strings.Contains(contextDirectWrapper, "__able_compiled_method_Counter_add_ctx(") {
		t.Fatalf("context-aware split-receiver wrapper must call the context body:\n%s", contextDirectWrapper)
	}
	compatibilityWrapper, ok := findCompiledDeclByPrefix(result, "func __able_wrap_method_Counter_add_direct")
	if !ok {
		t.Fatalf("could not find Counter.add compatibility split-receiver wrapper")
	}
	if !strings.Contains(compatibilityWrapper, "__able_wrap_method_Counter_add_ctx_direct(rt, &runtime.NativeCallContext{Env: __able_direct_env}, receiver, args)") {
		t.Fatalf("legacy split-receiver entry must adapt once into the context-aware wrapper:\n%s", compatibilityWrapper)
	}
}

func compilerExecutionContextFixtureRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	return root
}

func fixtureHasTypecheckDiagnostics(manifest interpreter.FixtureManifest) bool {
	return manifest.Expect.TypecheckDiagnostics != nil && len(manifest.Expect.TypecheckDiagnostics) > 0
}

func assertCompilerFixtureOutcomeParity(t *testing.T, tree compilerFixtureOutcome, compiled compilerFixtureOutcome) {
	t.Helper()
	if tree.Exit != compiled.Exit {
		t.Fatalf("exit mismatch: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	if !reflect.DeepEqual(tree.Stdout, compiled.Stdout) {
		t.Fatalf("stdout mismatch: treewalker=%v compiled=%v", tree.Stdout, compiled.Stdout)
	}
	if !reflect.DeepEqual(tree.Stderr, compiled.Stderr) {
		t.Fatalf("stderr mismatch: treewalker=%v compiled=%v", tree.Stderr, compiled.Stderr)
	}
}
