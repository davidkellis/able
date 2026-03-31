package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func compileNoFallbackPackage(t *testing.T, pkgName string, files map[string]string) *Result {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: "+pkgName+"\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
		EmitMain:           true,
		EntryPath:          entryPath,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	return result
}

func TestCompilerImportedPackageSelectorCallStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.io.temp",
			"",
			"fn main() -> String {",
			"  temp.dir(\"probe\").path",
			"}",
			"",
		}, "\n"),
		"io/temp/helpers.able": strings.Join([]string{
			"struct TempDir { path: String }",
			"",
			"fn dir(prefix: String) -> TempDir {",
			"  TempDir { path: prefix }",
			"}",
			"",
		}, "\n"),
	})

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_fn_dir(") {
		t.Fatalf("expected imported package selector call to lower to a compiled direct call:\n%s", mainBody)
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value(", "__able_method_call_node(", "__able_member_get_method("} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected imported package selector call to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerImportedZeroArgPackageSelectorCallStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.io",
			"",
			"fn main() -> String {",
			"  io.stdout()",
			"}",
			"",
		}, "\n"),
		"io/helpers.able": strings.Join([]string{
			"fn stdout() -> String {",
			"  \"ok\"",
			"}",
			"",
		}, "\n"),
	})

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_fn_stdout(") {
		t.Fatalf("expected imported zero-arg package selector call to lower to a compiled direct call:\n%s", mainBody)
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value(", "__able_method_call_node(", "__able_member_get_method("} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected imported zero-arg package selector call to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerImportedNominalStaticMethodCallStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.io.temp",
			"",
			"fn main() -> String {",
			"  temp.TempDir.new(\"probe\").path",
			"}",
			"",
		}, "\n"),
		"io/temp/helpers.able": strings.Join([]string{
			"struct TempDir { path: String }",
			"",
			"methods TempDir {",
			"  fn new(path: String) -> TempDir {",
			"    TempDir { path }",
			"  }",
			"}",
			"",
		}, "\n"),
	})

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_method_TempDir_new(") {
		t.Fatalf("expected imported nominal static method call to lower to a compiled direct call:\n%s", mainBody)
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value(", "__able_method_call_node(", "__able_member_get_method(", "__able_member_get(__able_env_get(\"temp\""} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected imported nominal static method call to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerPackageInitVerboseLambdaNestedReturnCompiles(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.helper",
			"",
			"fn main() -> i32 {",
			"  helper.read_value()",
			"}",
			"",
		}, "\n"),
		"helper/module.able": strings.Join([]string{
			"fn choose(f: i32 -> i32) -> i32 {",
			"  f(1)",
			"}",
			"",
			"value := choose(fn(v: i32) -> i32 {",
			"  if v > 0 { return 7 }",
			"  9",
			"})",
			"",
			"fn read_value() -> i32 {",
			"  value",
			"}",
			"",
		}, "\n"),
	})

	packageInitSrc, ok := result.Files["compiled_package_inits.go"]
	if !ok {
		t.Fatalf("expected compiled package init output for imported helper package")
	}
	initText := string(packageInitSrc)
	if !strings.Contains(initText, "__able_run_compiled_package_init_demo_helper_") {
		t.Fatalf("expected helper package init to be rendered:\n%s", initText)
	}
}

func TestCompilerSamePackageConstantAccessStaysStatic(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"FNV_OFFSET: u64 := 14695981039346656037_u64",
		"",
		"struct KernelHasher { state: u64 }",
		"",
		"methods KernelHasher {",
		"  fn new() -> KernelHasher {",
		"    KernelHasher { state: FNV_OFFSET }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_method_KernelHasher_new")
	if !ok {
		t.Fatalf("could not find compiled KernelHasher.new body")
	}
	if strings.Contains(body, "__able_env_get(\"FNV_OFFSET\"") {
		t.Fatalf("expected same-package constant access to avoid env lookup fallback:\n%s", body)
	}
	if !strings.Contains(body, "uint64(14695981039346656037)") {
		t.Fatalf("expected same-package constant access to lower as a static native constant:\n%s", body)
	}
}

func TestCompilerOptionalLastNamedFunctionCallStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-optional-last-named-call", strings.Join([]string{
		"package demo",
		"",
		"fn pick(base: i32, extra: ?i32) -> i32 {",
		"  extra match {",
		"    case nil => base,",
		"    case value: i32 => base + value",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  pick(10)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	if !strings.Contains(body, "__able_compiled_fn_pick(") {
		t.Fatalf("expected omitted optional arg named call to stay on the direct compiled path:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value(", "__able_method_call_node("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected omitted optional arg named call to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerOptionalLastMethodCallStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-optional-last-method-call", strings.Join([]string{
		"package demo",
		"",
		"struct Box {}",
		"",
		"methods Box {",
		"  fn #take(base: i32, extra: ?i32) -> i32 {",
		"    extra match {",
		"      case nil => base,",
		"      case value: i32 => base + value",
		"    }",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  Box {}.take(10)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	if !strings.Contains(body, "__able_compiled_method_Box_take(") {
		t.Fatalf("expected omitted optional arg method call to stay on the direct compiled path:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value(", "__able_method_call_node(", "__able_member_get_method("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected omitted optional arg method call to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerStdlibStaticOverloadedMethodChainStaysNative(t *testing.T) {
	result := compileSourceWithStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.core.interfaces.{Error}",
		"import able.text.string.{String}",
		"import able.text.automata_dsl.{AutomataDSL}",
		"",
		"fn make_string(value: String) -> String {",
		"  String.from_builtin(value) match {",
		"    case err: Error => { raise err },",
		"    case created: String => created",
		"  }",
		"}",
		"",
		"fn main() -> bool {",
		"  expr := AutomataDSL.union(",
		"    AutomataDSL.literal_char('x'),",
		"    AutomataDSL.literal_char('y')",
		"  ).optional()",
		"  nfa := expr.to_nfa()",
		"  nfa.matches(make_string(\"x\"))",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main body")
	}
	for _, fragment := range []string{
		"__able_compiled_method_AutomataDSL_union_a(",
		"__able_compiled_method_AutomataExpr_optional(",
		"__able_compiled_method_AutomataExpr_to_nfa(",
		"__able_compiled_method_NFA_matches(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected static automata method chain to lower through %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected static automata method chain to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerStdlibStaticOverloadedMethodChainExecutes(t *testing.T) {
	compileAndRunExecSourceWithOptions(t, "ablec-automata-static-overload-chain-exec", strings.Join([]string{
		"package demo",
		"",
		"import able.core.interfaces.{Error}",
		"import able.text.string.{String}",
		"import able.text.automata_dsl.{AutomataDSL}",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn make_string(value: String) -> String {",
		"  String.from_builtin(value) match {",
		"    case err: Error => { raise err },",
		"    case created: String => created",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  expr := AutomataDSL.union(",
		"    AutomataDSL.literal_char('x'),",
		"    AutomataDSL.literal_char('y')",
		"  ).optional()",
		"  nfa := expr.to_nfa()",
		"  if nfa.matches(make_string(\"\")) && nfa.matches(make_string(\"x\")) && nfa.matches(make_string(\"y\")) && !nfa.matches(make_string(\"xy\")) {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
}

func TestCompilerLocalConcreteApplyBindingStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Apply Args Result for T {",
		"  fn apply(self: Self, args: Args) -> Result",
		"}",
		"",
		"struct Multiplier { factor: i32 }",
		"",
		"impl Apply i32 i32 for Multiplier {",
		"  fn apply(self: Self, value: i32) -> i32 { self.factor * value }",
		"}",
		"",
		"fn main() -> i32 {",
		"  callable := Multiplier { factor: 6 }",
		"  callable(7)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_compiled_impl_Apply_") {
		t.Fatalf("expected local concrete apply binding to use compiled Apply dispatch:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_method_call_node(", "__able_member_get_method("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected local concrete apply binding to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerLocalInterfaceApplyBindingStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Apply Args Result for T {",
		"  fn apply(self: Self, args: Args) -> Result",
		"}",
		"",
		"struct Multiplier { factor: i32 }",
		"",
		"impl Apply i32 i32 for Multiplier {",
		"  fn apply(self: Self, value: i32) -> i32 { self.factor * value }",
		"}",
		"",
		"fn main() -> i32 {",
		"  callable: Apply i32 i32 = Multiplier { factor: 6 }",
		"  callable(7)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var callable __able_iface_Apply_i32_i32 = ") {
		t.Fatalf("expected local Apply interface binding to stay on the native carrier:\n%s", body)
	}
	if !strings.Contains(body, "callable.apply(") {
		t.Fatalf("expected local Apply interface binding to dispatch through the native interface method:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_method_call_node(", "__able_member_get_method("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected local Apply interface binding to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerNativeCallableStructFieldCallStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct ExampleContext {}",
		"",
		"struct ExampleDefinition {",
		"  body: ExampleContext -> String",
		"}",
		"",
		"fn main() -> String {",
		"  example := ExampleDefinition {",
		"    body: { _ctx => \"ok\" }",
		"  }",
		"  example.body(ExampleContext {})",
		"}",
		"",
	}, "\n"))

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "example.Body") {
		t.Fatalf("expected callable struct field call to lower to native field access plus static apply:\n%s", mainBody)
	}
	for _, fragment := range []string{"__able_method_call_node(", "__able_call_value(", "__able_member_get_method("} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected callable struct field call to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerNativeCallableStructFieldCallExecutes(t *testing.T) {
	compileAndRunSource(t, "ablec-callable-struct-field-", strings.Join([]string{
		"package demo",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct ExampleContext {}",
		"",
		"struct ExampleDefinition {",
		"  body: ExampleContext -> void",
		"}",
		"",
		"fn main() -> void {",
		"  example := ExampleDefinition {",
		"    body: { _ctx => __able_os_exit(0) }",
		"  }",
		"  example.body(ExampleContext {})",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n"))
}

func TestCompilerDispatchTouchpointsStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Point { x: i32, y: i32 }",
		"",
		"methods Point {",
		"  fn #bump(dx: i32) -> i32 {",
		"    #x = #x + dx",
		"    #x",
		"  }",
		"}",
		"",
		"interface Apply Args Result for T {",
		"  fn apply(self: Self, args: Args) -> Result",
		"}",
		"",
		"struct Multiplier { factor: i32 }",
		"",
		"impl Apply i32 i32 for Multiplier {",
		"  fn apply(self: Self, value: i32) -> i32 { self.factor * value }",
		"}",
		"",
		"interface Echo for Self {",
		"  fn pass<T>(self: Self, value: T) -> T",
		"}",
		"",
		"struct Box {}",
		"",
		"impl Echo for Box {",
		"  fn pass<T>(self: Self, value: T) -> T { value }",
		"}",
		"",
		"interface Index Idx Output for T {",
		"  fn get(self: Self, idx: Idx) -> !Output",
		"}",
		"",
		"interface IndexMut Idx Output for T {",
		"  fn set(self: Self, idx: Idx, value: Output) -> !void",
		"}",
		"",
		"struct Bag { items: Array i32 }",
		"",
		"impl Index i32 i32 for Bag {",
		"  fn get(self: Self, idx: i32) -> !i32 { self.items[idx] }",
		"}",
		"",
		"impl IndexMut i32 i32 for Bag {",
		"  fn set(self: Self, idx: i32, value: i32) -> !void {",
		"    self.items[idx] = value",
		"    return",
		"  }",
		"}",
		"",
		"fn main() -> String {",
		"  point := Point { x: 1, y: 2 }",
		"  _ = point.bump(4)",
		"  point.x = point.x + 1",
		"  direct := Multiplier { factor: 3 }",
		"  iface: Apply i32 i32 = Multiplier { factor: 4 }",
		"  bag := Bag { items: [10, 20, 30] }",
		"  bag[1] = direct(3)",
		"  echo: Echo = Box {}",
		"  `${point.x} ${bag[1]! as i32} ${iface(5)} ${echo.pass<String>(\"ok\")}`",
		"}",
		"",
	}, "\n"))

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_impl_Apply_") {
		t.Fatalf("expected combined dispatch audit to use compiled Apply lowering:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_impl_IndexMut_") {
		t.Fatalf("expected combined dispatch audit to use compiled IndexMut lowering:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_impl_Index_") {
		t.Fatalf("expected combined dispatch audit to use compiled Index lowering:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_iface_Echo_pass_dispatch(") {
		t.Fatalf("expected combined dispatch audit to use compiled generic interface dispatch:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"__able_call_value(",
		"__able_method_call_node(",
		"__able_member_get(",
		"__able_member_set(",
		"__able_member_get_method(",
		"__able_index(",
		"__able_index_set(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected combined static dispatch body to avoid %q:\n%s", fragment, mainBody)
		}
	}

	helperBody := mustCompiledFunctionBody(t, result, "__able_compiled_iface_Echo_pass_dispatch")
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
		"__able_iface_Echo_to_runtime_value(__able_runtime,",
	} {
		if strings.Contains(helperBody, fragment) {
			t.Fatalf("expected compiled generic interface dispatch helper to avoid %q:\n%s", fragment, helperBody)
		}
	}

	indexMutBody := mustCompiledFunctionBody(t, result, "__able_compiled_impl_IndexMut_set_0")
	for _, fragment := range []string{
		"__able_index_set(",
		"__able_struct_Array_to(",
	} {
		if strings.Contains(indexMutBody, fragment) {
			t.Fatalf("expected array-backed field mutation to stay on the native array carrier and avoid %q:\n%s", fragment, indexMutBody)
		}
	}

	indexGetBody := mustCompiledFunctionBody(t, result, "__able_compiled_impl_Index_get_0")
	for _, fragment := range []string{
		"__able_index(",
	} {
		if strings.Contains(indexGetBody, fragment) {
			t.Fatalf("expected array-backed field reads to stay on the native array carrier and avoid %q:\n%s", fragment, indexGetBody)
		}
	}
}

func TestCompilerSpawnCapturedReceiverDispatchStaysNative(t *testing.T) {
	result := compileExecFixtureResult(t, "12_05_concurrency_channel_ping_pong")

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_member_get_method(",
		"__able_struct_Channel_to(__able_runtime, channel)",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected spawned captured receiver dispatch to avoid %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_method_Channel_send(",
		"__able_compiled_method_Channel_receive(",
		"__able_compiled_method_Channel_close(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected spawned captured receiver dispatch to stay on compiled methods %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerSpawnSiblingCapturedReceiverDispatchBuildsAndStaysNative(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.concurrency.{Channel}",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn main() -> void {",
		"  ready := Channel.new(1)",
		"  sender := spawn { ready.send(true) }",
		"  receiver := spawn { ready.receive() }",
		"  sender.value()!",
		"  receiver.value()!",
		"  __able_os_exit(0)",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackExecSource(t, "ablec-spawn-sibling-captured-receiver", source)
	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"__able_member_get_method(",
		"__able_struct_Channel_to(__able_runtime, ready)",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected sibling spawned captured receiver dispatch to avoid %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_method_Channel_send_spec(",
		"__able_compiled_method_Channel_receive(ready)",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected sibling spawned captured receiver dispatch to stay on compiled methods %q:\n%s", fragment, mainBody)
		}
	}

	compileAndRunExecSourceWithOptions(t, "ablec-spawn-sibling-captured-receiver-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
}

func TestCompilerStaticIndexedStructMemberAccessFixtureStaysNative(t *testing.T) {
	result := compileExecFixtureResult(t, "06_01_compiler_dynamic_member_access")
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"__able_member_get(",
		"__able_member_get_method(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected indexed struct member access to avoid %q:\n%s", fragment, mainBody)
		}
	}
	if !strings.Contains(mainBody, ".Value") {
		t.Fatalf("expected indexed struct member access to lower to direct field access:\n%s", mainBody)
	}
}

func TestCompilerStaticIndexedStructMemberCompoundFixtureStaysNative(t *testing.T) {
	result := compileExecFixtureResult(t, "06_01_compiler_dynamic_member_compound")
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}

	bumpBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_bump")
	for _, fragment := range []string{
		"__able_member_get(",
		"__able_member_get_method(",
	} {
		if strings.Contains(bumpBody, fragment) {
			t.Fatalf("expected indexed struct member compound assignment to avoid %q:\n%s", fragment, bumpBody)
		}
	}
	if !strings.Contains(bumpBody, ".Value") {
		t.Fatalf("expected indexed struct member compound assignment to lower to direct field access:\n%s", bumpBody)
	}
}

func TestCompilerChainedGenericReceiverDispatchStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn main() -> bool {",
		"  eq(12_u32).matches(12_u32)",
		"}",
		"",
	}, "\n"))

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_fn_eq") {
		t.Fatalf("expected chained generic receiver to stay on a compiled constructor call:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_impl_Matcher_matches_0_spec(") {
		t.Fatalf("expected chained generic receiver to dispatch through the compiled specialized impl:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected chained generic receiver dispatch to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerChainedGenericReceiverDispatchExecutes(t *testing.T) {
	compileAndRunSource(t, "ablec-chained-generic-receiver-", strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn main() {",
		"  if eq(12_u32).matches(12_u32) {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n"))
}

func TestCompilerChainedGenericReceiverResultArgumentStaysNative(t *testing.T) {
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
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct Expectation T { actual: T }",
		"",
		"fn expect<T>(value: T) -> Expectation T {",
		"  Expectation T { actual: value }",
		"}",
		"",
		"methods Expectation T {",
		"  fn to(self: Self, matcher: Matcher T) -> bool {",
		"    matcher.matches(self.actual)",
		"  }",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn success() -> !u32 { 12_u32 }",
		"",
		"fn main() -> bool {",
		"  expect(success()!).to(eq(12_u32))",
		"}",
		"",
	}, "\n"))

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"__able_compiled_method_Expectation_to_spec(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected chained generic receiver with result-producing argument to use specialized nominal dispatch %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected chained generic receiver with result-producing argument to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerChainedGenericReceiverResultArgumentExecutes(t *testing.T) {
	compileAndRunSource(t, "ablec-chained-generic-receiver-result-", strings.Join([]string{
		"package demo",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct Expectation T { actual: T }",
		"",
		"fn expect<T>(value: T) -> Expectation T {",
		"  Expectation T { actual: value }",
		"}",
		"",
		"methods Expectation T {",
		"  fn to(self: Self, matcher: Matcher T) -> bool {",
		"    matcher.matches(self.actual)",
		"  }",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn success() -> !u32 { 12_u32 }",
		"",
		"fn main() {",
		"  if expect(success()!).to(eq(12_u32)) {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n"))
}

func TestCompilerGenericExpectationResultCarrierStaysNative(t *testing.T) {
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
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct Expectation T { actual: T }",
		"",
		"fn expect<T>(value: T) -> Expectation T {",
		"  Expectation T { actual: value }",
		"}",
		"",
		"methods Expectation T {",
		"  fn to(self: Self, matcher: Matcher T) -> bool {",
		"    matcher.matches(self.actual)",
		"  }",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn success() -> !u32 { 12_u32 }",
		"",
		"fn main() -> bool {",
		"  expect(success()).to(eq(12_u32))",
		"}",
		"",
	}, "\n"))

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(mainBody, "__able_compiled_method_Expectation_to_spec") {
		t.Fatalf("expected result-backed expectation call to stay on compiled nominal dispatch:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"__able_iface_Matcher_Result_from_value(",
		"__able_iface_Matcher_Result_u32__from_value(",
		"__able_union_uint32_or_runtime_ErrorValue_to_value(",
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected result-backed expectation path to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerGenericExpectationResultCarrierExecutes(t *testing.T) {
	compileAndRunSource(t, "ablec-generic-expectation-result-", strings.Join([]string{
		"package demo",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct Expectation T { actual: T }",
		"",
		"fn expect<T>(value: T) -> Expectation T {",
		"  Expectation T { actual: value }",
		"}",
		"",
		"methods Expectation T {",
		"  fn to(self: Self, matcher: Matcher T) -> bool {",
		"    matcher.matches(self.actual)",
		"  }",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn success() -> !u32 { 12_u32 }",
		"",
		"fn main() {",
		"  if expect(success()).to(eq(12_u32)) {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n"))
}
