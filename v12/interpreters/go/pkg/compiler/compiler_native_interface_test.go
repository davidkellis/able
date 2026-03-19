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

func TestCompilerGenericInterfaceReturnFromStructLiteralStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Iterable T for Self {",
		"  fn each(self: Self, visit: T -> void) -> void {",
		"    visit(self.first())",
		"  }",
		"  fn first(self: Self) -> T",
		"}",
		"",
		"struct RangeLike { start: i32 }",
		"",
		"impl Iterable i32 for RangeLike {",
		"  fn first(self: Self) -> i32 { self.start }",
		"}",
		"",
		"fn make() -> Iterable i32 {",
		"  RangeLike { start: 7 }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_make")
	if !ok {
		t.Fatalf("could not find compiled make function")
	}
	if !strings.Contains(body, "__able_iface_Iterable_i32_wrap_ptr_RangeLike(") {
		t.Fatalf("expected struct literal return to wrap directly into the native generic interface carrier:\n%s", body)
	}
	if strings.Contains(body, "__able_any_to_value(") || strings.Contains(body, "bridge.MatchType(") {
		t.Fatalf("expected generic interface return to avoid runtime-value fallback:\n%s", body)
	}
}

func TestCompilerUnionTargetInterfaceAssignmentStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Tag for Self {",
		"  fn tag(self: Self) -> String",
		"}",
		"",
		"struct Alpha {}",
		"struct Beta {}",
		"",
		"impl Tag for Alpha | Beta {",
		"  fn tag(self: Self) -> String {",
		"    self match {",
		"      case Alpha => \"alpha\",",
		"      case Beta => \"beta\"",
		"    }",
		"  }",
		"}",
		"",
		"fn main() -> String {",
		"  first: Tag = Alpha {}",
		"  second: Tag = Beta {}",
		"  `${first.tag()} ${second.tag()}`",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_iface_Tag_wrap___able_union_") {
		t.Fatalf("expected union-target interface assignment to wrap through the native union carrier:\n%s", body)
	}
	if strings.Contains(body, "__able_try_cast(") || strings.Contains(body, "bridge.MatchType(") {
		t.Fatalf("expected union-target interface assignment to avoid dynamic fallback:\n%s", body)
	}
}

func TestCompilerStaticIndexInterfacesStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Bag { items: Array i32 }",
		"",
		"interface Index Idx Output for T {",
		"  fn get(self: Self, idx: Idx) -> !Output",
		"}",
		"",
		"interface IndexMut Idx Output for T {",
		"  fn set(self: Self, idx: Idx, value: Output) -> !void",
		"}",
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
		"fn main() -> !i32 {",
		"  bag := Bag { items: [10, 20, 30] }",
		"  bag[1] = 99",
		"  bag[1]",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_compiled_impl_IndexMut_") {
		t.Fatalf("expected index assignment to dispatch through the compiled IndexMut impl:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_impl_Index_") {
		t.Fatalf("expected index read to dispatch through the compiled Index impl:\n%s", body)
	}
	for _, fragment := range []string{"__able_index_set(", "__able_index(", "bridge.IndexAssign(", "bridge.Index("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected static index lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerConcreteReceiverInterfaceMethodStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Named for Self {",
		"  fn name(self: Self) -> String",
		"  fn describe(self: Self) -> String { `hi ${self.name()}` }",
		"}",
		"",
		"struct Person { raw: String }",
		"",
		"impl Named for Person {",
		"  fn name(self: Self) -> String { self.raw }",
		"}",
		"",
		"fn main() -> String {",
		"  person := Person { raw: \"Ada\" }",
		"  person.describe()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_compiled_impl_Named_describe_") {
		t.Fatalf("expected concrete receiver interface method call to dispatch through the compiled interface impl:\n%s", body)
	}
	for _, fragment := range []string{"__able_method_call_node(", "__able_call_value(", "bridge.MatchType("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete receiver interface method call to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerConcreteReceiverApplyStaysNative(t *testing.T) {
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
		"  multiplier := Multiplier { factor: 6 }",
		"  multiplier(7)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_compiled_impl_Apply_") {
		t.Fatalf("expected concrete receiver apply call to dispatch through the compiled Apply impl:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_method_call_node(", "bridge.MatchType("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete receiver apply call to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerInterfaceMethodWithLambdaArgStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Walker for Self {",
		"  fn each(self: Self, visit: i32 -> void) -> void",
		"}",
		"",
		"struct Box { value: i32 }",
		"",
		"impl Walker for Box {",
		"  fn each(self: Self, visit: i32 -> void) -> void {",
		"    visit(self.value)",
		"  }",
		"}",
		"",
		"fn check(walker: Walker) -> void {",
		"  walker.each({ n => {",
		"    n match {",
		"      case 1 => {},",
		"      case _ => {}",
		"    }",
		"  } })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_check")
	if !ok {
		t.Fatalf("could not find compiled check function")
	}
	if !strings.Contains(body, "walker.each(") {
		t.Fatalf("expected native interface method dispatch for callback-bearing method:\n%s", body)
	}
	for _, fragment := range []string{"__able_method_call_node(", "__able_call_value(", "bridge.MatchType("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected callback-bearing interface method call to avoid %q:\n%s", fragment, body)
		}
	}
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "bridge.AsInt(__able_lambda_arg_0_value, 32)") {
		t.Fatalf("expected lambda argument to use the native callback parameter type instead of runtime.Value:\n%s", compiledSrc)
	}
}

func TestCompilerNativeInterfaceRuntimeAdapterUsesStructZeroForVoidReturn(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reporter for Self {",
		"  fn emit(self: Self, line: String) -> void",
		"}",
		"",
		"struct Console {}",
		"",
		"impl Reporter for Console {",
		"  fn emit(self: Self, line: String) -> void {}",
		"}",
		"",
		"fn send(reporter: Reporter) -> void {",
		"  reporter.emit(\"ok\")",
		"}",
		"",
	}, "\n"))

	adapterBody, ok := findCompiledDeclByPrefix(result, "func (w __able_iface_Reporter_runtime_adapter) emit(")
	if !ok {
		t.Fatalf("could not find Reporter runtime adapter method")
	}
	for _, fragment := range []string{
		"result, control := __able_method_call(w.Value, \"emit\", args)",
		"_ = result",
		"converted := struct{}{}",
		"return converted, nil",
	} {
		if !strings.Contains(adapterBody, fragment) {
			t.Fatalf("expected Reporter runtime adapter to contain %q:\n%s", fragment, adapterBody)
		}
	}
}

func TestCompilerNativeInterfaceRuntimeAdapterWritesBackPointerArgs(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reporter for Self {",
		"  fn emit(self: Self, line: String) -> void",
		"}",
		"",
		"interface Framework for Self {",
		"  fn run(self: Self, reporter: Reporter) -> void",
		"}",
		"",
		"struct Progress { lines: Array String }",
		"struct Host {}",
		"",
		"impl Reporter for Progress {",
		"  fn emit(self: Self, line: String) -> void {",
		"    self.lines.push(line)",
		"  }",
		"}",
		"",
		"impl Framework for Host {",
		"  fn run(self: Self, reporter: Reporter) -> void {",
		"    reporter.emit(\"ok\")",
		"  }",
		"}",
		"",
		"fn invoke(framework: Framework, reporter: Reporter) -> void {",
		"  framework.run(reporter)",
		"}",
		"",
	}, "\n"))

	applyBody, ok := findCompiledFunction(result, "__able_iface_Reporter_apply_runtime_value")
	if !ok {
		t.Fatalf("could not find Reporter runtime writeback helper")
	}
	for _, fragment := range []string{
		"switch typed := value.(type) {",
		"case __able_iface_Reporter_adapter_ptr_Progress:",
		"converted, err := __able_struct_Progress_from(runtimeValue)",
		"*typed.Value = *converted",
	} {
		if !strings.Contains(applyBody, fragment) {
			t.Fatalf("expected Reporter writeback helper to contain %q:\n%s", fragment, applyBody)
		}
	}

	adapterBody, ok := findCompiledDeclByPrefix(result, "func (w __able_iface_Framework_runtime_adapter) run(")
	if !ok {
		t.Fatalf("could not find Framework runtime adapter method")
	}
	if !strings.Contains(adapterBody, "if err := __able_iface_Reporter_apply_runtime_value(arg0, arg0Value); err != nil {") {
		t.Fatalf("expected Framework runtime adapter to write back Reporter args after runtime dispatch:\n%s", adapterBody)
	}
	if !strings.Contains(adapterBody, "result, control := __able_method_call(w.Value, \"run\", args)") {
		t.Fatalf("expected Framework runtime adapter to dispatch through the runtime method call shim:\n%s", adapterBody)
	}
}
