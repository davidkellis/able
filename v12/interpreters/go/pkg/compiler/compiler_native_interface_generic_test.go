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
	if !strings.Contains(body, "__able_iface_Echo_to_runtime_value(__able_runtime,") {
		t.Fatalf("expected generic interface call to narrow the runtime boundary to receiver conversion:\n%s", body)
	}
	if !strings.Contains(body, "__able_method_call_node(") {
		t.Fatalf("expected generic interface call to dispatch through the narrow generic-method boundary:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_member_get_method(", "bridge.MatchType("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected pure-generic interface dispatch to avoid %q:\n%s", fragment, body)
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
	if !strings.Contains(body, "__able_iface_Tagger_to_runtime_value(__able_runtime,") {
		t.Fatalf("expected default generic method call to narrow the runtime boundary to receiver conversion:\n%s", body)
	}
	if !strings.Contains(body, "__able_method_call_node(") {
		t.Fatalf("expected default generic method call to dispatch through the generic-method boundary:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_member_get_method(", "bridge.MatchType("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected default generic method dispatch to avoid %q:\n%s", fragment, body)
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
