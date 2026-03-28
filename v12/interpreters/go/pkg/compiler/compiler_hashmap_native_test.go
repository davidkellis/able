package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerHashMapStaticCarrierStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashmap-native-static", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_map.*",
		"",
		"fn build() -> HashMap String i32 {",
		"  values: HashMap String i32 = HashMap.with_capacity(2)",
		"  values.raw_set(\"a\", 1)",
		"  values.raw_set(\"b\", 2)",
		"  values",
		"}",
		"",
		"fn size_of(values: HashMap String i32) -> i32 {",
		"  values.raw_size()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_build() (*HashMap") {
		t.Fatalf("expected HashMap return to stay on the native carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_size_of(values *HashMap") {
		t.Fatalf("expected HashMap param to stay on the native carrier:\n%s", compiledSrc)
	}

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var values *HashMap",
		"__able_compiled_method_HashMap_with_capacity",
		"__able_compiled_method_HashMap_raw_set",
	} {
		if !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native HashMap build lowering to contain %q:\n%s", fragment, buildBody)
		}
	}
	for _, fragment := range []string{
		"var values any =",
		"__able_any_to_value(values)",
		"__able_try_cast(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native HashMap build lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}

	sizeBody, ok := findCompiledFunction(result, "__able_compiled_fn_size_of")
	if !ok {
		t.Fatalf("could not find compiled size_of function")
	}
	if !strings.Contains(sizeBody, "__able_compiled_method_HashMap_raw_size") {
		t.Fatalf("expected native HashMap param dispatch in size_of:\n%s", sizeBody)
	}
}

func TestCompilerHashMapLiteralStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashmap-native-literal", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_map.*",
		"",
		"fn main() -> i32 {",
		"  values: HashMap String i32 := #{ \"a\": 1, \"b\": 2 }",
		"  values[\"a\"]! + values[\"b\"]!",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *HashMap_String_i32 = func() *HashMap_String_i32 {",
		"handleRaw, err := __able_hash_map_handle_from_value(handleVal)",
		"return &HashMap_String_i32{Handle: handleRaw}",
		"__able_hash_map_set_impl(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected native HashMap literal lowering to contain %q:\n%s", fragment, body)
		}
	}
	if strings.Contains(body, "runtime.StructInstanceValue{Definition: def, Fields: map[string]runtime.Value{\"handle\": handleVal}}") {
		t.Fatalf("expected HashMap literal lowering to avoid runtime struct-instance materialization in the compiled body:\n%s", body)
	}
}

func TestCompilerHashMapCarrierArrayStaysSpecialized(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashmap-native-array", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_map.*",
		"",
		"fn build() -> Array (HashMap String i32) {",
		"  maps: Array (HashMap String i32) = []",
		"  values := HashMap.new()",
		"  values.raw_set(\"a\", 1)",
		"  maps.push(values)",
		"  maps",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_HashMap_String_i32 struct {",
		"Elements       []*HashMap_String_i32",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected Array(HashMap) carrier lowering to contain %q", fragment)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var maps *__able_array_HashMap_String_i32 =",
		"var values *HashMap =",
		"__able_nominal_coerce_HashMap_to_HashMap_String_i32",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected Array(HashMap) build lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"[]runtime.Value{",
		"runtime.ArrayValue",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected Array(HashMap) build lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerGenericMapInterfaceSignatureStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashmap-native-map-iface", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_map.*",
		"import able.collections.map.*",
		"",
		"fn use(values: Map String i32) -> i32 {",
		"  values.get(\"a\") match {",
		"    case nil => 0",
		"    case value: i32 => value",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  values := HashMap.new()",
		"  values.set(\"a\", 7)",
		"  view: Map String i32 = values",
		"  use(view)",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type __able_iface_Map_String_i32 interface {") {
		t.Fatalf("expected Map<String,i32> native interface carrier to be emitted:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_use(values __able_iface_Map_String_i32) (int32, *__ableControl)") {
		t.Fatalf("expected Map<String,i32> param to stay on the native interface carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_use")
	if !ok {
		t.Fatalf("could not find compiled use function")
	}
	if !strings.Contains(body, "values.get(\"a\")") {
		t.Fatalf("expected native Map interface dispatch in use:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_call_value(",
		"__able_method_call_node(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected Map interface use to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerHashMapNativeCarrierExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_map.*",
		"",
		"fn total() -> i32 {",
		"  first: HashMap String i32 = #{ \"a\": 1, \"b\": 2 }",
		"  second := HashMap.new()",
		"  second.raw_set(\"a\", 3)",
		"  second.raw_set(\"b\", 4)",
		"  maps: Array (HashMap String i32) = []",
		"  maps.push(first)",
		"  maps.push(second)",
		"  total := 0",
		"  for values in maps {",
		"    total = total + values[\"a\"]! + values[\"b\"]!",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> void {",
		"  print(total())",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-hashmap-native-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "10" {
		t.Fatalf("expected native HashMap carrier program to print 10, got %q", stdout)
	}
}

func TestCompilerHashSetStaticCarrierStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashset-native-static", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_set.*",
		"",
		"fn build() -> HashSet i32 {",
		"  values: HashSet i32 = HashSet.with_capacity(2)",
		"  values.add(1)",
		"  values.add(2)",
		"  values",
		"}",
		"",
		"fn size_of(values: HashSet i32) -> i32 {",
		"  values.size()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_build() (*HashSet_i32, *__ableControl)") {
		t.Fatalf("expected HashSet return to stay on the native carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_size_of(values *HashSet_i32) (int32, *__ableControl)") {
		t.Fatalf("expected HashSet param to stay on the native carrier:\n%s", compiledSrc)
	}

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var values *HashSet_i32 =",
		"__able_compiled_method_HashSet_with_capacity_spec(",
	} {
		if !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native HashSet build lowering to contain %q:\n%s", fragment, buildBody)
		}
	}
	if addName, ok := calledFunctionNameFromBody(buildBody, "__able_compiled_method_HashSet_add_spec"); !ok {
		t.Fatalf("expected native HashSet build lowering to call a specialized add helper:\n%s", buildBody)
	} else if !strings.Contains(buildBody, addName+"(values, int32(1))") || !strings.Contains(buildBody, addName+"(values, int32(2))") {
		t.Fatalf("expected native HashSet build lowering to reuse the specialized add helper:\n%s", buildBody)
	}
	for _, fragment := range []string{
		"var values any =",
		"__able_any_to_value(values)",
		"__able_try_cast(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native HashSet build lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}

	sizeBody, ok := findCompiledFunction(result, "__able_compiled_fn_size_of")
	if !ok {
		t.Fatalf("could not find compiled size_of function")
	}
	if !strings.Contains(sizeBody, "__able_compiled_method_HashSet_size_spec(values)") {
		t.Fatalf("expected native HashSet param dispatch in size_of:\n%s", sizeBody)
	}
}

func TestCompilerHashSetCarrierArrayStaysSpecialized(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashset-native-array", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_set.*",
		"",
		"fn build() -> Array (HashSet i32) {",
		"  sets: Array (HashSet i32) = []",
		"  values := HashSet.new()",
		"  values.add(1)",
		"  sets.push(values)",
		"  sets",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_HashSet_i32 struct {",
		"Elements       []*HashSet_i32",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected Array(HashSet) carrier lowering to contain %q", fragment)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var sets *__able_array_HashSet_i32 =",
		"var values *HashSet =",
		"__able_nominal_coerce_HashSet_to_HashSet_i32",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected Array(HashSet) build lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"[]runtime.Value{",
		"runtime.ArrayValue",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected Array(HashSet) build lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerHashSetIteratorWrapsConcreteNativeIterator(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-hashset-native-iterator", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_set.*",
		"import able.core.iteration.{Iterator, IteratorEnd}",
		"",
		"fn build() -> Iterator i32 {",
		"  values := HashSet.with_capacity(3)",
		"  values.add(1)",
		"  values.add(2)",
		"  values.iterator()",
		"}",
		"",
		"fn count(iter: Iterator i32) -> i32 {",
		"  total := 0",
		"  loop {",
		"    iter.next() match {",
		"      case IteratorEnd {} => break,",
		"      case _value: i32 => total = total + 1",
		"    }",
		"  }",
		"  total",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_iface_Iterator_i32_wrap_ptr_HashSetIterator_i32(") {
		t.Fatalf("expected Iterator<i32> concrete adapter for HashSetIterator_i32 to be emitted:\n%s", compiledSrc)
	}

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_build() (__able_iface_Iterator_i32, *__ableControl)",
		"__able_compiled_impl_Enumerable_iterator_0_spec(",
		"__able_nominal_coerce_HashSet_to_HashSet_i32(",
		"__able_iface_Iterator_A_to_runtime_value(__able_runtime,",
		"__able_iface_Iterator_i32_from_value(__able_runtime,",
	} {
		if !strings.Contains(compiledSrc, fragment) && !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected HashSet iterator lowering to contain %q:\n%s", fragment, buildBody)
		}
	}
	for _, fragment := range []string{
		"__able_call_named(",
		"__able_try_cast(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected HashSet iterator lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}
}

func TestCompilerHashSetIteratorNativeCarrierExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_set.*",
		"import able.core.iteration.{Iterator, IteratorEnd}",
		"",
		"fn build() -> Iterator i32 {",
		"  values := HashSet.with_capacity(3)",
		"  values.add(1)",
		"  values.add(2)",
		"  values.iterator()",
		"}",
		"",
		"fn count(iter: Iterator i32) -> i32 {",
		"  total := 0",
		"  loop {",
		"    iter.next() match {",
		"      case IteratorEnd {} => { break },",
		"      case _value: i32 => { total = total + 1 }",
		"    }",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> void {",
		"  print(count(build()))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-hashset-native-iterator-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "2" {
		t.Fatalf("expected native HashSet iterator program to print 2, got %q", stdout)
	}
}

func compileNoFallbackExecSource(t *testing.T, tempPrefix string, source string) *Result {
	t.Helper()
	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)

	entryPath := filepath.Join(workDir, "main.able")
	if err := os.WriteFile(filepath.Join(workDir, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	searchPaths, err := buildExecSearchPaths(entryPath, workDir, interpreter.FixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
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
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	_ = moduleRoot
	return result
}

func compileNoFallbackExecSourceWithOptions(t *testing.T, tempPrefix string, source string, opts Options) *Result {
	t.Helper()
	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)

	entryPath := filepath.Join(workDir, "main.able")
	if err := os.WriteFile(filepath.Join(workDir, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	searchPaths, err := buildExecSearchPaths(entryPath, workDir, interpreter.FixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	opts.RequireNoFallbacks = true
	result, err := New(opts).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	_ = moduleRoot
	return result
}

func compileAndRunExecSourceWithOptions(t *testing.T, tempPrefix string, source string, opts Options) string {
	t.Helper()
	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)

	entryPath := filepath.Join(workDir, "app.able")
	if err := os.WriteFile(filepath.Join(workDir, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	searchPaths, err := buildExecSearchPaths(entryPath, workDir, interpreter.FixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	outputDir := filepath.Join(workDir, "out")
	opts.EntryPath = entryPath
	result, err := New(opts).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-bin")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	run := exec.Command(binPath)
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(output))
	}
	return string(output)
}
