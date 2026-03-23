package compiler

import (
	"strings"
	"testing"
)

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
