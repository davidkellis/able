package compiler

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompilerStaticGenericUnionKnownMethodAvoidsBoundMethodBox(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main", EmitMain: true})

	var awaitHelpers bytes.Buffer
	gen.renderRuntimeAwaitHelpers(&awaitHelpers)
	awaitSource := awaitHelpers.String()
	helperStart := strings.Index(awaitSource, "func __able_call_known_native_method_fast(")
	if helperStart < 0 {
		t.Fatal("known native method helper was not rendered")
	}
	knownMethodHelper := awaitSource[helperStart:]
	for _, fragment := range []string{
		"func __able_call_known_native_method_fast(receiver runtime.Value, entry *__able_compiled_method_entry, args []runtime.Value) (runtime.Value, error)",
		"if entry.direct != nil",
		"return entry.direct(__able_runtime, env, receiver, args)",
		"ctx := &runtime.NativeCallContext{Env: env, State: state}",
		"injected := append([]runtime.Value{receiver}, args...)",
		"return entry.fn.Impl(ctx, injected)",
	} {
		if !strings.Contains(knownMethodHelper, fragment) {
			t.Fatalf("known native method helper missing %q", fragment)
		}
	}
	directIndex := strings.Index(knownMethodHelper, "if entry.direct != nil")
	contextIndex := strings.Index(knownMethodHelper, "ctx := &runtime.NativeCallContext{Env: env, State: state}")
	if directIndex < 0 || contextIndex < 0 || contextIndex < directIndex {
		t.Fatalf("known native method helper must construct NativeCallContext only after the direct-path return")
	}

	var callHelpers bytes.Buffer
	gen.renderRuntimeCallHelpers(&callHelpers)
	callSource := callHelpers.String()
	start := strings.Index(callSource, "func __able_static_generic_union_method_call(")
	if start < 0 {
		t.Fatal("static generic-union method helper was not rendered")
	}
	segment := callSource[start:]
	if end := strings.Index(segment, "func __able_compiled_thunk_value("); end >= 0 {
		segment = segment[:end]
	}
	if !strings.Contains(segment, "__able_call_known_native_method_fast(obj, entry, args)") {
		t.Fatalf("static generic-union method helper does not use the known-method path:\n%s", segment)
	}
	if strings.Contains(segment, "runtime.NativeBoundMethodValue{Receiver: obj, Method: *entry.fn}") {
		t.Fatalf("static generic-union method helper still boxes its known method:\n%s", segment)
	}
	if !strings.Contains(segment, "bridge.CallStaticGenericUnionMember(__able_runtime, obj, methodName, args, call)") {
		t.Fatalf("static generic-union method helper no longer preserves the dynamic fallback:\n%s", segment)
	}
}

func TestCompilerStaticGenericUnionKnownMethodPreservesCallSemantics(t *testing.T) {
	source := strings.Join([]string{
		"union Choice T = nil | T",
		"",
		"methods Choice T {",
		"  fn plus(self: Self, amount: i32) -> i32 {",
		"    self match {",
		"      case nil => amount,",
		"      case value => (value as i32) + amount,",
		"    }",
		"  }",
		"",
		"  fn touch(self: Self) -> void {}",
		"",
		"  fn fail(self: Self) -> void { raise \"known-method-error\" }",
		"}",
		"",
		"fn main() -> void {",
		"  present: Choice i32 = 4_i32",
		"  print(present.plus(3_i32))",
		"  present.touch()",
		"  do {",
		"    present.fail()",
		"  } rescue {",
		"    case err => { print(err) }",
		"  }",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-known-generic-union-method-", source, Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   "main.able",
	})
	if stdout != "7\nknown-method-error\n" {
		t.Fatalf("known generic-union method output = %q", stdout)
	}
}
