package compiler

import (
	"strings"
	"testing"
)

func TestCompilerInterfaceParamAndReturnStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
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
		"fn identity(value: Greeter) -> Greeter {",
		"  value",
		"}",
		"",
		"fn greet(value: Greeter) -> String {",
		"  value.greet()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type __able_iface_Greeter interface {") {
		t.Fatalf("expected a native Greeter interface carrier to be emitted:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_identity(value __able_iface_Greeter) (__able_iface_Greeter, *__ableControl)") {
		t.Fatalf("expected Greeter param/return to stay on the native interface carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_greet(value __able_iface_Greeter) (string, *__ableControl)") {
		t.Fatalf("expected Greeter param to stay on the native interface carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_greet")
	if !ok {
		t.Fatalf("could not find compiled greet function")
	}
	if !strings.Contains(body, "value.greet()") {
		t.Fatalf("expected native interface dispatch in greet body:\n%s", body)
	}
	for _, fragment := range []string{"bridge.MatchType(", "__able_method_call_node(", "__able_try_cast("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected native interface dispatch to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerNativeInterfaceExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"interface Greeter for Self {",
		"  fn greet(self: Self) -> String",
		"}",
		"",
		"struct Person { name: String }",
		"struct Loud { label: String }",
		"",
		"impl Greeter for Person {",
		"  fn greet(self: Self) -> String { `hi ${self.name}` }",
		"}",
		"",
		"impl Greeter for Loud {",
		"  fn greet(self: Self) -> String { `HEY ${self.label}` }",
		"}",
		"",
		"fn identity(value: Greeter) -> Greeter {",
		"  value",
		"}",
		"",
		"fn greet_twice(value: Greeter) -> String {",
		"  first := value.greet()",
		"  second := identity(value).greet()",
		"  `${first} / ${second}`",
		"}",
		"",
		"fn main() {",
		"  if greet_twice(Person { name: \"Ada\" }) == \"hi Ada / hi Ada\" &&",
		"     greet_twice(Loud { label: \"ABLE\" }) == \"HEY ABLE / HEY ABLE\" {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-native-interface-", source)
}

func TestCompilerTypedInterfaceAssignmentStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Speaker for Self {",
		"  fn label(self: Self) -> String",
		"}",
		"",
		"struct Robot { id: i32 }",
		"",
		"impl Speaker for Robot {",
		"  fn label(self: Self) -> String { `robot-${self.id}` }",
		"}",
		"",
		"fn main() -> String {",
		"  value: Speaker = Robot { id: 7 }",
		"  value.label()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var value __able_iface_Speaker = __able_iface_Speaker_wrap_ptr_Robot(") {
		t.Fatalf("expected typed interface assignment to wrap the native Robot carrier directly:\n%s", body)
	}
	if !strings.Contains(body, "value.label()") {
		t.Fatalf("expected typed interface local to dispatch through the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{"bridge.MatchType(", "__able_method_call_node(", "__able_try_cast("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected typed interface assignment to avoid %q:\n%s", fragment, body)
		}
	}
}
