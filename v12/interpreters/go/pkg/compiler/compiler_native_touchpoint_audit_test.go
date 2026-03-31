package compiler

import (
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func mustCompiledFunctionBody(t *testing.T, result *Result, name string) string {
	t.Helper()
	body, ok := findCompiledFunction(result, name)
	if !ok {
		t.Fatalf("could not find compiled function %s", name)
	}
	return body
}

func assertBodyAvoidsFragments(t *testing.T, funcName string, body string, fragments []string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(body, fragment) {
			t.Fatalf("%s should avoid %q:\n%s", funcName, fragment, body)
		}
	}
}

func TestCompilerBroadStaticNativeTouchpointsStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #inc() -> i32 { #value + 1 }",
		"}",
		"",
		"interface Reader for Self {",
		"  fn read(self: Self) -> i32",
		"}",
		"",
		"impl Reader for Counter {",
		"  fn read(self: Self) -> i32 { self.value }",
		"}",
		"",
		"union MaybeCounter = Counter | nil",
		"",
		"fn apply_twice(f: i32 -> i32, value: i32) -> i32 {",
		"  f(f(value))",
		"}",
		"",
		"fn describe(value: MaybeCounter) -> i32 {",
		"  value match {",
		"    case counter: Counter => counter.inc(),",
		"    case nil => 0",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3]",
		"  arr[1] = 9",
		"  counter := Counter { value: arr[1]! as i32 }",
		"  reader: Reader = counter",
		"  add1 := fn(value: i32) -> i32 { value + 1 }",
		"  describe(counter) + reader.read() + apply_twice(add1, 39)",
		"}",
		"",
	}, "\n"))

	for _, item := range []struct {
		name string
		body string
	}{
		{name: "__able_compiled_fn_main", body: mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")},
		{name: "__able_compiled_fn_apply_twice", body: mustCompiledFunctionBody(t, result, "__able_compiled_fn_apply_twice")},
		{name: "__able_compiled_fn_describe", body: mustCompiledFunctionBody(t, result, "__able_compiled_fn_describe")},
	} {
		assertBodyAvoidsFragments(t, item.name, item.body, []string{
			"__able_index(",
			"__able_index_set(",
			"__able_member_get(",
			"__able_member_set(",
			"__able_member_get_method(",
			"__able_call_value(",
			"__able_call_value_fast(",
			"__able_method_call_node(",
			"bridge.MatchType(",
			"__able_try_cast(",
			"__able_any_to_value(",
			"panic(",
			"recover(",
			"func() ",
		})
	}

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "var reader __able_iface_Reader = __able_iface_Reader_wrap_ptr_Counter(") {
		t.Fatalf("expected object-safe interface assignment to stay on the native carrier:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "&__able_array_i32{Length: int32(3), Capacity: int32(3), Storage_handle: int64(0), Elements: []int32{") {
		t.Fatalf("expected array literal to stay on the compiler-native i32 array carrier:\n%s", mainBody)
	}

	applyBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_apply_twice")
	if !strings.Contains(applyBody, "__able_fn_int32_to_int32") {
		t.Fatalf("expected function-typed param calls to stay on the native callable carrier:\n%s", applyBody)
	}
}

func TestCompilerGenericInterfaceTouchpointsStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Tagged T { tag: String, value: T }",
		"",
		"interface Tagger for Self {",
		"  fn tag(self: Self) -> String",
		"  fn tagged<T>(self: Self, value: T) -> Tagged T {",
		"    Tagged T { tag: self.tag(), value: value }",
		"  }",
		"}",
		"",
		"struct Labeler { label: String }",
		"",
		"impl Tagger for Labeler {",
		"  fn tag(self: Self) -> String { self.label }",
		"}",
		"",
		"fn main() -> String {",
		"  labeler: Tagger = Labeler { label: \"L\" }",
		"  first := labeler.tagged(\"alpha\")",
		"  first.tag",
		"}",
		"",
	}, "\n"))

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "var labeler __able_iface_Tagger = __able_iface_Tagger_wrap_ptr_Labeler(") {
		t.Fatalf("expected interface-typed local to stay on the native carrier:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_iface_Tagger_tagged_default(") {
		t.Fatalf("expected generic interface default-method dispatch to stay on the compiled native helper path:\n%s", mainBody)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", mainBody, []string{
		"__able_iface_Tagger_to_runtime_value(__able_runtime,",
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_call_value_fast(",
		"__able_member_get(",
		"__able_member_set(",
		"__able_member_get_method(",
		"bridge.MatchType(",
		"__able_try_cast(",
		"__able_any_to_value(",
		"panic(",
		"recover(",
		"func() ",
	})
}

func TestCompilerPatternControlTouchpointsStayNative(t *testing.T) {
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
		"fn value(ok: bool) -> !i32 {",
		"  if ok { 1 } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> i32 {",
		"  mixed := if true { 1 } else { \"bad\" }",
		"  from_match := mixed match {",
		"    case n: i32 => n,",
		"    case _ => 0",
		"  }",
		"  from_or := value(false) or { 7 }",
		"  from_rescue := do {",
		"    raise(\"boom\")",
		"    0",
		"  } rescue {",
		"    case _ => 9",
		"  }",
		"  from_loop := loop {",
		"    if true { break from_match + from_or + from_rescue }",
		"    break 0",
		"  }",
		"  breakpoint 'done {",
		"    if false { break 'done 0 }",
		"    from_loop",
		"  }",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"bridge.MatchType(",
		"__able_try_cast(",
		"panic(",
		"recover(",
		"func() ",
	})
}

func TestCompilerStaticNativeFixturesExecuteWithoutExplicitBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler native touchpoint fixture audit in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	for _, rel := range []string{
		"06_01_compiler_placeholder_lambda",
		"06_01_compiler_bound_method_value",
		"06_08_array_ops_mutability",
		"10_03_interface_type_dynamic_dispatch",
	} {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			dir := filepath.Join(root, filepath.FromSlash(rel))
			manifest, err := interpreter.LoadFixtureManifest(dir)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if shouldSkipTarget(manifest.SkipTargets, "go") {
				t.Skip("fixture skipped for go target")
			}
			if manifest.Expect.TypecheckDiagnostics != nil && len(manifest.Expect.TypecheckDiagnostics) > 0 {
				t.Skip("fixture expects typecheck diagnostics")
			}
			outcome, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)
			expectedExit := 0
			if manifest.Expect.Exit != nil {
				expectedExit = *manifest.Expect.Exit
			}
			if outcome.Exit != expectedExit {
				t.Fatalf("expected exit=%d, got exit=%d stderr=%v stdout=%v", expectedExit, outcome.Exit, outcome.Stderr, outcome.Stdout)
			}
			if markers.FallbackCount != 0 {
				t.Fatalf("expected no fallback boundary crossings, got %d (%q)", markers.FallbackCount, markers.FallbackNames)
			}
			if markers.ExplicitCount != 0 {
				t.Fatalf("expected no explicit dynamic boundary crossings for fully native static fixture, got %d (%q)", markers.ExplicitCount, markers.ExplicitNames)
			}
		})
	}
}
