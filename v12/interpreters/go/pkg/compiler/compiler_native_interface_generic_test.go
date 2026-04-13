package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func canonicalStdlibSourcePath(t *testing.T) string {
	t.Helper()
	repoRoot := repositoryRoot()
	for _, candidate := range []string{
		filepath.Join(repoRoot, "able-stdlib", "src"),
		filepath.Join(filepath.Dir(repoRoot), "able-stdlib", "src"),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	t.Skip("canonical able-stdlib src directory not available")
	return ""
}

func compileSourceWithStdlibPaths(t *testing.T, source string) *Result {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	searchPaths := []driver.SearchPath{
		{Path: root, Kind: driver.RootUser},
		{Path: filepath.Join(repositoryRoot(), "v12", "stdlib-deprecated-do-not-use", "src"), Kind: driver.RootStdlib},
		{Path: filepath.Join(repositoryRoot(), "v12", "kernel", "src"), Kind: driver.RootStdlib},
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
		PackageName: "main",
		EntryPath:   entryPath,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return result
}

func compileNoFallbackSourceWithCanonicalStdlibPaths(t *testing.T, source string) *Result {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	searchPaths := []driver.SearchPath{
		{Path: root, Kind: driver.RootUser},
		{Path: canonicalStdlibSourcePath(t), Kind: driver.RootStdlib},
		{Path: filepath.Join(repositoryRoot(), "v12", "kernel", "src"), Kind: driver.RootStdlib},
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
		EntryPath:          entryPath,
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with canonical stdlib under no-fallbacks: %v", err)
	}
	return result
}

func compileAndBuildCanonicalStdlibSource(t *testing.T, tempPrefix string, source string) *Result {
	t.Helper()
	result := compileNoFallbackSourceWithCanonicalStdlibPaths(t, source)
	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)
	outputDir := filepath.Join(workDir, "out")
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}
	build := exec.Command("go", "test", "-run", "^$", ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go test build failed: %v\n%s", err, string(output))
	}
	return result
}

func compileSourceWithCanonicalStdlibPaths(t *testing.T, source string) *Result {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	searchPaths := []driver.SearchPath{
		{Path: root, Kind: driver.RootUser},
		{Path: canonicalStdlibSourcePath(t), Kind: driver.RootStdlib},
		{Path: filepath.Join(repositoryRoot(), "v12", "kernel", "src"), Kind: driver.RootStdlib},
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
		PackageName: "main",
		EntryPath:   entryPath,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with canonical stdlib: %v", err)
	}
	return result
}

func TestCompilerPureGenericInterfaceAssignmentUsesNativeCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
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
		"fn main() -> String {",
		"  value: Echo = Box {}",
		"  value.pass<String>(\"ok\")",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type __able_iface_Echo interface {") {
		t.Fatalf("expected a native carrier for the pure-generic Echo interface:\n%s", compiledSrc)
	}
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var value __able_iface_Echo = __able_iface_Echo_wrap_ptr_Box(") {
		t.Fatalf("expected the pure-generic interface local to stay on the native carrier:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_iface_Echo_pass_dispatch(") {
		t.Fatalf("expected generic interface call to dispatch through the compiled native helper:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_iface_Echo_to_runtime_value(__able_runtime,",
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected pure-generic interface dispatch to avoid %q:\n%s", fragment, body)
		}
	}
	helperBody, ok := findCompiledFunction(result, "__able_compiled_iface_Echo_pass_dispatch")
	if !ok {
		t.Fatalf("could not find compiled generic interface dispatch helper")
	}
	if !strings.Contains(helperBody, "__able_compiled_impl_Echo_pass_") || !strings.Contains(helperBody, "_spec(") {
		t.Fatalf("expected generic interface dispatch helper to call the specialized compiled impl directly:\n%s", helperBody)
	}
	for _, fragment := range []string{
		"__able_iface_Echo_to_runtime_value(__able_runtime,",
	} {
		if strings.Contains(helperBody, fragment) {
			t.Fatalf("expected generic interface dispatch helper to avoid %q:\n%s", fragment, helperBody)
		}
	}
}

func TestCompilerDefaultGenericInterfaceMethodUsesNativeReceiverBoundary(t *testing.T) {
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

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var labeler __able_iface_Tagger = __able_iface_Tagger_wrap_ptr_Labeler(") {
		t.Fatalf("expected interface-typed receiver to stay on the native carrier:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_iface_Tagger_tagged_default(") {
		t.Fatalf("expected default generic method call to use the compiled native default body directly:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_iface_Tagger_to_runtime_value(__able_runtime,",
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_member_get_method(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected default generic method call to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerGenericInterfaceExistentialExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct Tagged T { tag: String, value: T }",
		"",
		"interface Tokenizer for Self {",
		"  fn token<T>(self: Self, value: T) -> Tagged T",
		"}",
		"",
		"interface Tagger for Self {",
		"  fn tag(self: Self) -> String",
		"  fn tagged<T>(self: Self, value: T) -> Tagged T {",
		"    Tagged T { tag: self.tag(), value: value }",
		"  }",
		"}",
		"",
		"struct Prefixer { prefix: String }",
		"struct Labeler { label: String }",
		"",
		"impl Tokenizer for Prefixer {",
		"  fn token<T>(self: Self, value: T) -> Tagged T {",
		"    Tagged T { tag: self.prefix, value: value }",
		"  }",
		"}",
		"",
		"impl Tagger for Labeler {",
		"  fn tag(self: Self) -> String { self.label }",
		"}",
		"",
		"fn main() {",
		"  tokenizer: Tokenizer = Prefixer { prefix: \"tok\" }",
		"  labeler: Tagger = Labeler { label: \"L\" }",
		"  first := tokenizer.token(\"alpha\")",
		"  second := tokenizer.token<i32>(42)",
		"  third := labeler.tagged(\"beta\")",
		"  fourth := labeler.tagged<i32>(7)",
		"  if first.tag == \"tok\" && first.value == \"alpha\" &&",
		"     second.tag == \"tok\" && second.value == 42 &&",
		"     third.tag == \"L\" && third.value == \"beta\" &&",
		"     fourth.tag == \"L\" && fourth.value == 7 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-native-generic-iface-", source)
}

func TestCompilerInterfaceLookupGenericMethodFixturesRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generic interface fixture regression in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	for _, rel := range []string{
		"10_04_interface_dispatch_defaults_generics",
		"10_15_interface_default_generic_method",
	} {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			runCompilerInterfaceLookupAuditFixture(t, root, rel)
		})
	}
}

func TestCompilerImportedGenericInterfaceAdapterRendersConcreteHelper(t *testing.T) {
	result := compileExecFixtureResult(t, "10_04_interface_dispatch_defaults_generics")
	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_iface_Tokenizer_adapter_ptr_Prefixer struct {",
		"func __able_iface_Tokenizer_wrap_ptr_Prefixer(value *Prefixer) __able_iface_Tokenizer {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported Tokenizer<-Prefixer native adapter helper to be rendered; missing %q", fragment)
		}
	}
}

func TestCompilerGenericInterfaceBoundaryHelperRendersLateSpecializedConcreteAdapter(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EmptyArrayMatcher T {}",
		"",
		"impl Matcher (Array T) for EmptyArrayMatcher T {",
		"  fn matches(self: Self, value: Array T) -> bool {",
		"    value.len() == 0",
		"  }",
		"}",
		"",
		"fn apply(matcher: Matcher (Array i32)) -> bool {",
		"  matcher.matches([])",
		"}",
		"",
		"fn main() -> bool {",
		"  apply(EmptyArrayMatcher i32 {})",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_iface_Matcher_Array_i32__from_value(",
		"__able_struct_EmptyArrayMatcher_i32_from(",
		"__able_iface_Matcher_Array_i32__wrap_ptr_EmptyArrayMatcher_i32(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected Matcher<Array i32> boundary helper to render the specialized concrete adapter; missing %q", fragment)
		}
	}
}

func TestCompilerGenericInterfaceBoundaryHelperRendersPackageInitSpecializedConcreteAdapter(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EmptyArrayMatcher T {}",
		"",
		"impl Matcher (Array T) for EmptyArrayMatcher T {",
		"  fn matches(self: Self, value: Array T) -> bool {",
		"    value.len() == 0",
		"  }",
		"}",
		"",
		"ROOT_MATCHER: Matcher (Array i32) = EmptyArrayMatcher i32 {}",
		"",
		"fn main() -> bool {",
		"  ROOT_MATCHER.matches([])",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_iface_Matcher_Array_i32__from_value(",
		"__able_struct_EmptyArrayMatcher_i32_from(",
		"__able_iface_Matcher_Array_i32__wrap_ptr_EmptyArrayMatcher_i32(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected package-init-specialized Matcher<Array i32> boundary helper to render the concrete adapter; missing %q", fragment)
		}
	}
}

func TestCompilerSpecializedImplCanonicalKeyPreventsDuplicateContainAllBodies(t *testing.T) {
	result := compileSourceWithStdlibPaths(t, strings.Join([]string{
		"package demo",
		"import able.spec.*",
		"",
		"fn main() -> bool {",
		"  values: Array String := Array.new()",
		"  values.push(\"a\")",
		"  expected: Array String := Array.new()",
		"  expected.push(\"a\")",
		"  matcher := contain_all(expected)",
		"  matcher.matches(values).passed",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if count := strings.Count(compiledSrc, "func __able_compiled_impl_Matcher_matches_0_m_spec"); count != 1 {
		t.Fatalf("expected exactly one specialized ContainAllMatcher<String>.matches body, found %d", count)
	}
	if strings.Contains(compiledSrc, "__able_compiled_impl_Matcher_matches_0_m_spec_a(") {
		t.Fatalf("expected stale duplicate specialization body to be absent")
	}
}

func TestCompilerContainAllStringMatcherUsesDirectStringSpecialization(t *testing.T) {
	result := compileSourceWithStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.spec.*",
		"",
		"fn main() -> bool {",
		"  values: Array String := Array.new()",
		"  values.push(\"a\")",
		"  values.push(\"b\")",
		"  expected: Array String := Array.new()",
		"  expected.push(\"a\")",
		"  matcher := contain_all(expected)",
		"  matcher.matches(values).passed",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	if strings.Contains(mainBody, "__able_method_call_node(") {
		t.Fatalf("expected contain_all matcher dispatch to stay static:\n%s", mainBody)
	}

	compiledSrc := string(result.Files["compiled.go"])
	matchesBody, ok := findCompiledDeclByPrefix(result, "func __able_compiled_impl_Matcher_matches_0_m_spec")
	if !ok {
		t.Fatalf("expected a specialized ContainAllMatcher<String>.matches body:\n%s", compiledSrc)
	}
	arrayContainsName, ok := calledFunctionNameFromBody(matchesBody, "__able_compiled_fn_array_contains_spec")
	if !ok {
		t.Fatalf("expected ContainAllMatcher<String>.matches to call a specialized shared Array<String> helper:\n%s", matchesBody)
	}
	if !strings.Contains(matchesBody, arrayContainsName+"(actual, value)") {
		t.Fatalf("expected ContainAllMatcher<String>.matches to call %s directly:\n%s", arrayContainsName, matchesBody)
	}
	for _, fragment := range []string{
		"__able_array_Array_from(",
		"__able_array_array_String_from(",
	} {
		if strings.Contains(matchesBody, fragment) {
			t.Fatalf("expected ContainAllMatcher<String>.matches path to avoid %q:\n%s", fragment, matchesBody)
		}
	}
}

func TestCompilerIterableDefaultMethodSelfSiblingStaysStaticForStdlibImports(t *testing.T) {
	result := compileSourceWithStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.spec.*",
		"",
		"fn main() -> void {}",
		"",
	}, "\n"))

	for _, fallback := range result.Fallbacks {
		if fallback.Name == "iface Iterable.iterator" {
			t.Fatalf("expected Iterable.iterator default method to stay static, got fallback reason %q", fallback.Reason)
		}
	}
	if !strings.Contains(string(result.Files["compiled.go"]), "__able_compiled_impl_Iterable_iterator_") {
		t.Fatalf("expected compiled Iterable.iterator impl helper to be rendered")
	}
}

func TestCompilerCanonicalStdlibStringImportKeepsIterableIteratorStatic(t *testing.T) {
	result := compileNoFallbackSourceWithCanonicalStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.text.string",
		"",
		"fn main() -> void {}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "__able_compiled_impl_Iterable_iterator_") {
		t.Fatalf("expected canonical stdlib string import to keep Iterable-for-Iterator specialization static:\n%s", compiledSrc)
	}
}

func TestCompilerCanonicalStdlibStringIterableEachUsesBuiltinReceiver(t *testing.T) {
	result := compileNoFallbackSourceWithCanonicalStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.text.string",
		"",
		"fn main() -> void {",
		"  \"abc\".each(fn(ch: char) -> void { _ = ch })",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	builtinReceiverPattern := regexp.MustCompile(`func __able_compiled_impl_Iterable_each_default_[A-Za-z0-9_]+\(self string, visit __able_fn_rune_to_struct__\)`)
	if !builtinReceiverPattern.MatchString(compiledSrc) {
		t.Fatalf("expected String Iterable.each default helper to use builtin string receiver:\n%s", compiledSrc)
	}
	structReceiverPattern := regexp.MustCompile(`func __able_compiled_impl_Iterable_each_default_[A-Za-z0-9_]+\(self \*String, visit __able_fn_rune_to_struct__\)`)
	if structReceiverPattern.MatchString(compiledSrc) {
		t.Fatalf("expected String Iterable.each default helper to avoid *String receiver:\n%s", compiledSrc)
	}
}

func TestCompilerCanonicalStdlibSpecImportKeepsIterableIteratorStatic(t *testing.T) {
	result := compileNoFallbackSourceWithCanonicalStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.spec.*",
		"",
		"fn main() -> void {}",
		"",
	}, "\n"))

	if !strings.Contains(string(result.Files["compiled.go"]), "__able_compiled_impl_Iterable_iterator_") {
		t.Fatalf("expected canonical stdlib spec import to keep Iterable-for-Iterator specialization static")
	}
}

func TestCompilerTypedArrayDefaultMethodsKeepConcreteReceivers(t *testing.T) {
	result := compileAndBuildCanonicalStdlibSource(t, "ablec-canonical-array-defaults-", strings.Join([]string{
		"package demo",
		"",
		"import able.spec.*",
		"",
		"fn main() -> void {",
		"  labels: Array String := Array.new()",
		"  labels.push(\"a\")",
		"  _ = labels.drop(0)",
		"  _ = labels.lazy()",
		"  numbers: Array i32 := Array.new()",
		"  numbers.push(1)",
		"  _ = numbers.drop(0)",
		"  _ = numbers.lazy()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main body")
	}
	dropCalls := regexp.MustCompile(`__able_compiled_impl_Enumerable_drop_default_[A-Za-z0-9_]+_spec(?:_[A-Za-z0-9_]+)?\(`).FindAllString(mainBody, -1)
	if len(dropCalls) < 2 {
		t.Fatalf("expected typed Array drop calls to use specialized impl helpers:\n%s", mainBody)
	}
	lazyCalls := regexp.MustCompile(`__able_compiled_impl_Enumerable_lazy_default_[A-Za-z0-9_]+_spec(?:_[A-Za-z0-9_]+)?\(`).FindAllString(mainBody, -1)
	if len(lazyCalls) < 2 {
		t.Fatalf("expected typed Array lazy calls to use specialized impl helpers:\n%s", mainBody)
	}
	if !regexp.MustCompile(`func __able_compiled_impl_Enumerable_drop_default_[A-Za-z0-9_]+_spec(?:_[A-Za-z0-9_]+)?\(self \*[A-Za-z0-9_]+, count int32\) \(\*[A-Za-z0-9_]+, \*__ableControl\)`).MatchString(compiledSrc) {
		t.Fatalf("expected typed Array drop default helper to stay on the shared impl specialization path:\n%s", compiledSrc)
	}
	if !regexp.MustCompile(`func __able_compiled_impl_Enumerable_lazy_default_[A-Za-z0-9_]+_spec\(self \*[A-Za-z0-9_]+\) \(__able_iface_Iterator_String, \*__ableControl\)`).MatchString(compiledSrc) {
		t.Fatalf("expected String Array lazy default helper to keep the concrete iterator result:\n%s", compiledSrc)
	}
	if !regexp.MustCompile(`func __able_compiled_impl_Enumerable_lazy_default_[A-Za-z0-9_]+_spec_[A-Za-z0-9_]+\(self \*[A-Za-z0-9_]+\) \(__able_iface_Iterator_i32, \*__ableControl\)`).MatchString(compiledSrc) {
		t.Fatalf("expected i32 Array lazy default helper to keep the concrete iterator result:\n%s", compiledSrc)
	}
}

func TestCompilerCanonicalStdlibExpectationResultArgumentStaysConcrete(t *testing.T) {
	result := compileSourceWithCanonicalStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.numbers.bigint.{BigInt}",
		"import able.spec.*",
		"",
		"fn main() -> bool {",
		"  expect(BigInt.from_u64(42_u64).to_i32()).to(eq(42))",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"Expectation_Result_i32",
		"__able_iface_Matcher_Result_i32_",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected canonical stdlib expectation result path to stay concrete; missing %q:\n%s", fragment, compiledSrc)
		}
	}
	for _, fragment := range []string{
		"__able_iface_Matcher_Result_from_value(",
		"type mismatch: expected Matcher<Result>",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected canonical stdlib expectation result path to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerGenericInterfaceBoundaryHelperSynthesizesImplGenericConcreteAdapter(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct Contains T { expected: T }",
		"",
		"impl Matcher (Array T) for Contains T {",
		"  fn matches(self: Self, value: Array T) -> bool {",
		"    value.len() > 0",
		"  }",
		"}",
		"",
		"fn make() -> Contains String {",
		"  Contains String { expected: \"x\" }",
		"}",
		"",
		"fn accept(matcher: Matcher (Array String)) -> bool {",
		"  matcher.matches([\"x\"])",
		"}",
		"",
		"fn main() -> bool {",
		"  value := make()",
		"  accept(value)",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_iface_Matcher_Array_String__from_value(",
		"__able_struct_Contains_String_from(",
		"__able_iface_Matcher_Array_String__wrap_ptr_Contains_String(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected impl-generic Matcher<Array String> boundary helper to synthesize the concrete adapter; missing %q", fragment)
		}
	}
}

func TestCompilerGenericInterfaceBoundaryHelperRejectsMismatchedSpecializedReceiver(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EqMatcher T {}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool {",
		"    true",
		"  }",
		"}",
		"",
		"fn keep_string(value: Matcher (Array String)) -> bool {",
		"  value.matches([\"x\"])",
		"}",
		"",
		"fn keep_i32(value: Matcher (Array i32)) -> bool {",
		"  value.matches([1])",
		"}",
		"",
		"fn main() -> bool {",
		"  keep_string(EqMatcher (Array String) {}) && keep_i32(EqMatcher (Array i32) {})",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "__able_iface_Matcher_Array_i32__wrap_ptr_EqMatcher_Array_i32(") {
		t.Fatalf("expected Matcher<Array i32> boundary helper to render the matching EqMatcher<Array i32> adapter")
	}
	if strings.Contains(compiledSrc, "__able_iface_Matcher_Array_i32__wrap_ptr_EqMatcher_Array_String(") {
		t.Fatalf("expected Matcher<Array i32> boundary helper to reject the mismatched EqMatcher<Array String> specialization:\n%s", compiledSrc)
	}
}

func TestCompilerCallableReturnCoercionInterfaceAdapterStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct RaiseErrorMatcher {}",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"impl Matcher (() -> void) for RaiseErrorMatcher {",
		"  fn matches(self: Self, invocation: () -> void) -> bool {",
		"    handled := false",
		"    do { invocation() } rescue {",
		"      case err: Error => {",
		"        handled = err.message() == \"boom\"",
		"        nil",
		"      }",
		"    }",
		"    handled",
		"  }",
		"}",
		"",
		"fn fail() -> !i32 {",
		"  MyError { message: \"boom\" }",
		"}",
		"",
		"fn main() -> bool {",
		"  matcher: Matcher (() -> !i32) = RaiseErrorMatcher {}",
		"  matcher.matches(fn() { fail() })",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "_adapter_ptr_RaiseErrorMatcher struct {") {
		t.Fatalf("expected callable-return interface coercion to render a concrete native adapter for RaiseErrorMatcher:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "wrap_ptr_RaiseErrorMatcher(") {
		t.Fatalf("expected callable-return matcher local to stay on a concrete native adapter:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_method_call(",
		"__able_method_call_node(",
		"bridge.MatchType(",
		"_wrap_runtime(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected callable-return interface adapter path to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerCallableReturnCoercionInterfaceAdapterExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct RaiseErrorMatcher {}",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"impl Matcher (() -> void) for RaiseErrorMatcher {",
		"  fn matches(self: Self, invocation: () -> void) -> bool {",
		"    handled := false",
		"    do { invocation() } rescue {",
		"      case err: Error => {",
		"        handled = err.message() == \"boom\"",
		"        nil",
		"      }",
		"    }",
		"    handled",
		"  }",
		"}",
		"",
		"fn fail() -> !i32 {",
		"  MyError { message: \"boom\" }",
		"}",
		"",
		"fn main() {",
		"  matcher: Matcher (() -> !i32) = RaiseErrorMatcher {}",
		"  if matcher.matches(fn() { fail() }) {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-native-callable-return-adapter-", source)
}

func TestCompilerIteratorInterfaceBoundaryAcceptsRuntimeIteratorDirectly(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_18_stdlib_collections_array_range")

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(mainBody, "__able_compiled_iface_Iterable_iterator_default(") {
		t.Fatalf("expected concrete Iterable receiver calls to dispatch through the carrier adapter instead of the generic default helper:\n%s", mainBody)
	}
	if !regexp.MustCompile(`__able_iface_Iterable_i32_wrap_ptr_Range\([^)]*\)\.iterator\(\)`).MatchString(mainBody) {
		t.Fatalf("expected concrete Range iterable calls to dispatch through the wrapped interface carrier method:\n%s", mainBody)
	}

	tryBody, ok := findCompiledFunction(result, "__able_iface_Iterator_i32_try_from_value")
	if !ok {
		t.Fatalf("could not find Iterator<i32> matcher helper")
	}
	if !strings.Contains(tryBody, "if iter, ok, nilPtr := __able_runtime_iterator_value(value); ok || nilPtr {") {
		t.Fatalf("expected Iterator<i32> matcher helper to fast-path raw runtime iterators:\n%s", tryBody)
	}
	if !strings.Contains(tryBody, "return __able_iface_Iterator_i32_wrap_runtime(iter), true, nil") {
		t.Fatalf("expected Iterator<i32> matcher helper to wrap raw runtime iterators directly:\n%s", tryBody)
	}
	body, ok := findCompiledFunction(result, "__able_iface_Iterator_i32_from_value")
	if !ok {
		t.Fatalf("could not find Iterator<i32> boundary helper")
	}
	if !strings.Contains(body, "converted, ok, err := __able_iface_Iterator_i32_try_from_value(rt, value)") {
		t.Fatalf("expected Iterator<i32> boundary helper to delegate to the shared matcher helper:\n%s", body)
	}

	nextBody, ok := findCompiledFunction(result, "(w __able_iface_Iterator_i32_runtime_adapter) next")
	if !ok {
		t.Fatalf("could not find Iterator<i32> runtime adapter next method")
	}
	if !strings.Contains(nextBody, "result, done, err := iter.Next()") {
		t.Fatalf("expected Iterator<i32> runtime adapter to fast-path raw iterator next calls:\n%s", nextBody)
	}

	controlBody, ok := findCompiledFunction(result, "__able_control_from_error_with_node")
	if !ok {
		t.Fatalf("could not find control error-normalization helper")
	}
	if !strings.Contains(controlBody, "case __able_generator_stop:") || !strings.Contains(controlBody, "return &__ableControl{Err: v}") {
		t.Fatalf("expected control error-normalization helper to preserve generator stop sentinels:\n%s", controlBody)
	}
}
