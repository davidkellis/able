package compiler

import (
	"strings"
	"testing"
)

func TestCompilerTreatsSelfTypedFirstMethodParamAsInstanceReceiver(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"struct Counter {",
			"  n: i32",
			"}",
			"",
			"methods Counter {",
			"  fn bump(this: Self) -> i32 {",
			"    this.n + 1",
			"  }",
			"}",
			"",
			"fn main() -> void {",
			"  _ = Counter { n: 1 }.bump()",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Counter\", \"bump\", true") {
		t.Fatalf("expected Counter.bump to be registered as a compiled instance method when first param type is Self")
	}
	if strings.Contains(compiledSrc, "__able_register_compiled_method(\"Counter\", \"bump\", false") {
		t.Fatalf("expected Counter.bump not to be registered as a static method")
	}
	for _, fragment := range []string{
		"__able_register_compiled_method_direct(\"Counter\", \"bump\", __able_wrap_method_Counter_bump_direct)",
		"func __able_wrap_method_Counter_bump_direct(rt *bridge.Runtime, __able_direct_env *runtime.Environment, receiver runtime.Value, args []runtime.Value)",
		"return __able_wrap_method_Counter_bump_direct(rt, __able_compiled_direct_env_from_native(ctx), args[0], args[1:])",
		"arg0Value := receiver",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected split-receiver compiled method ABI fragment %q", fragment)
		}
	}
}

func TestCompilerReservesSplitReceiverWrapperParameterNames(t *testing.T) {
	for idx, name := range []string{"receiver", "__able_direct_env"} {
		if got := safeParamName(name, idx); got == name {
			t.Fatalf("safeParamName(%q) must not collide with the split-receiver wrapper", name)
		}
	}
}
