package compiler

import (
	"path/filepath"
	"strings"
	"testing"
)

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

func TestCompilerIteratorInterfaceBoundaryAcceptsRuntimeIteratorDirectly(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_18_stdlib_collections_array_range")

	body, ok := findCompiledFunction(result, "__able_iface_Iterator_i32_from_value")
	if !ok {
		t.Fatalf("could not find Iterator<i32> boundary helper")
	}
	if !strings.Contains(body, "if iter, ok, nilPtr := __able_runtime_iterator_value(value); ok || nilPtr {") {
		t.Fatalf("expected Iterator<i32> boundary helper to fast-path raw runtime iterators:\n%s", body)
	}
	if !strings.Contains(body, "return __able_iface_Iterator_i32_wrap_runtime(iter), nil") {
		t.Fatalf("expected Iterator<i32> boundary helper to wrap raw runtime iterators directly:\n%s", body)
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
