package interpreter

import (
	"fmt"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_MemberMethodCacheTracksStructDefinition(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structS := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	structT := ast.StructDef(
		"T",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	sPing := ast.Fn(
		"ping",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Int(11),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	tPing := ast.Fn(
		"ping",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Int(22),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	callPing := ast.Fn(
		"call_ping",
		[]*ast.FunctionParameter{
			ast.Param("value", nil),
		},
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("value"), "ping")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structS,
		structT,
		ast.Methods(ast.Ty("S"), []*ast.FunctionDefinition{sPing}, nil, nil),
		ast.Methods(ast.Ty("T"), []*ast.FunctionDefinition{tPing}, nil, nil),
		callPing,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "n"),
			}, false, "S", nil, nil),
		),
		ast.Assign(
			ast.ID("t"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(2), "n"),
			}, false, "T", nil, nil),
		),
		ast.Call("call_ping", ast.ID("s")),
		ast.Call("call_ping", ast.ID("t")),
		ast.Call("call_ping", ast.ID("t")),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode member cache receiver-definition mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 22 {
		t.Fatalf("expected second call_ping to use T.ping and return 22, got %#v", got)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheMiss < 2 {
		t.Fatalf("expected receiver-definition changes to force cache misses, got misses=%d", stats.MemberMethodCacheMiss)
	}
	if stats.MemberMethodCacheHits == 0 {
		t.Fatalf("expected member method cache hit after repeating the T receiver")
	}
}

func TestBytecodeVM_MemberMethodCacheWorksInPackageEnv(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	ping := ast.Fn(
		"ping",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Int(7),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	callPing := ast.Fn(
		"call_ping",
		[]*ast.FunctionParameter{
			ast.Param("s", ast.Ty("S")),
		},
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("s"), "ping")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		ast.Methods(ast.Ty("S"), []*ast.FunctionDefinition{ping}, nil, nil),
		callPing,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "n"),
			}, false, "S", nil, nil),
		),
		ast.Call("call_ping", ast.ID("s")),
		ast.Call("call_ping", ast.ID("s")),
	}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	packageEnv := runtime.NewEnvironment(interp.GlobalEnvironment())
	vm := newBytecodeVM(interp, packageEnv)
	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("bytecode execution failed: %v", err)
	}
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode package-env member cache mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheMiss == 0 {
		t.Fatalf("expected package-env member method cache misses > 0")
	}
	if stats.MemberMethodCacheHits == 0 {
		t.Fatalf("expected package-env member method cache hits > 0")
	}
}

func TestBytecodeVM_MemberMethodCacheKeysPrimitiveReceiversByType(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}

	receiverI32A := runtime.NewSmallInt(7, runtime.IntegerI32)
	receiverI32B := runtime.NewSmallInt(11, runtime.IntegerI32)
	receiverI64 := runtime.NewSmallInt(11, runtime.IntegerI64)
	native := runtime.NativeFunctionValue{
		Name:  "eq",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.BoolValue{Val: true}, nil
		},
	}

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"eq",
		true,
		receiverI32A,
		runtime.NativeBoundMethodValue{Receiver: receiverI32A, Method: native},
	); !ok {
		t.Fatalf("expected primitive member-method cache store to succeed")
	}

	if _, ok := vm.lookupCachedMemberMethodEntry(program, 1, "eq", true, receiverI32B); !ok {
		t.Fatalf("expected i32 receiver cache hit across distinct values of the same primitive type")
	}
	if _, ok := vm.lookupCachedMemberMethodEntry(program, 1, "eq", true, receiverI64); ok {
		t.Fatalf("expected primitive member-method cache miss when the integer suffix changes")
	}
}

func TestBytecodeVM_MemberMethodCacheHitsForRepeatedPrimitiveReceiverCalls(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	program := mustLoadAbleProgramFromSource(t, `
import able.core.interfaces.{Less}

fn is_less(left: i32, right: i32) -> bool {
  left.cmp(right) == Less
}

fn main() -> bool {
  is_less(1, 2) && is_less(1, 2)
}

main()
`)

	want, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("tree evaluation failed: %v", err)
	}
	interp := NewBytecode()
	got, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode primitive member cache mismatch: got=%#v want=%#v", got, want)
	}
	if boolVal, ok := got.(runtime.BoolValue); !ok || !boolVal.Val {
		t.Fatalf("expected repeated primitive cmp call to return true, got %#v", got)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheMiss == 0 {
		t.Fatalf("expected primitive member-method cache miss on first call")
	}
	if stats.MemberMethodCacheHits == 0 {
		t.Fatalf("expected primitive member-method cache hit on repeated call")
	}
}

func TestBytecodeVM_LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}

	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	receiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	native := runtime.NativeFunctionValue{
		Name:  "ping",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.NewSmallInt(int64(len(args)), runtime.IntegerI32), nil
		},
	}

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.NativeBoundMethodValue{Receiver: receiver, Method: native},
	); !ok {
		t.Fatalf("expected member-method cache store to succeed")
	}

	cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver)
	if !ok {
		t.Fatalf("expected cached member-method entry")
	}
	if cached.template == nil {
		t.Fatalf("expected cached template")
	}
	if _, ok := cached.template.(runtime.NativeBoundMethodValue); ok {
		t.Fatalf("expected cached entry to keep unbound native template, got %#v", cached.template)
	}
	template, ok := cached.template.(runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("expected native function template, got %T", cached.template)
	}
	if template.Name != native.Name || template.Arity != native.Arity {
		t.Fatalf("unexpected cached template: got=%#v want=%#v", template, native)
	}
	if cached.dispatch != bytecodeMemberMethodDispatchExactNative {
		t.Fatalf("expected exact-native cached dispatch, got %v", cached.dispatch)
	}
	if cached.inlineFn != nil {
		t.Fatalf("expected exact-native cached dispatch to skip inline function, got %#v", cached.inlineFn)
	}

	target, ok := bytecodeResolveExactInjectedNativeCallTarget(cached.template, receiver, 0)
	if !ok {
		t.Fatalf("expected cached template to remain exact-native dispatchable")
	}
	if !target.hasReceiver {
		t.Fatalf("expected exact-native target to include receiver")
	}
	if target.injectedReceiver != receiver {
		t.Fatalf("unexpected injected receiver: got=%#v want=%#v", target.injectedReceiver, receiver)
	}
}

func TestBytecodeVM_LookupCachedMemberMethodEntryKeepsInlineDispatchForDirectFunctionTemplate(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}

	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	receiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	methodDef := ast.Fn(
		"ping",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{ast.Int(1)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	methodProgram, err := interp.lowerFunctionDefinitionBytecode(methodDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	methodFn := &runtime.FunctionValue{
		Declaration: methodDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(methodFn, methodProgram)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: methodFn},
	); !ok {
		t.Fatalf("expected member-method cache store to succeed")
	}

	cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver)
	if !ok {
		t.Fatalf("expected cached member-method entry")
	}
	if cached.template != methodFn {
		t.Fatalf("expected direct function template to be cached, got %#v want %#v", cached.template, methodFn)
	}
	if cached.dispatch != bytecodeMemberMethodDispatchInline {
		t.Fatalf("expected inline cached dispatch, got %v", cached.dispatch)
	}
	if cached.inlineFn != methodFn {
		t.Fatalf("expected cached inline function to match template, got %#v want %#v", cached.inlineFn, methodFn)
	}
}

func TestBytecodeVM_NonMethodMemberAccessSkipsMemberMethodCacheCounters(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	readField := ast.Fn(
		"read_field",
		[]*ast.FunctionParameter{
			ast.Param("s", ast.Ty("S")),
		},
		[]ast.Statement{
			ast.Member(ast.ID("s"), "n"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		readField,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(7), "n"),
			}, false, "S", nil, nil),
		),
		ast.Call("read_field", ast.ID("s")),
		ast.Call("read_field", ast.ID("s")),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode non-method member access mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheHits != 0 || stats.MemberMethodCacheMiss != 0 {
		t.Fatalf("expected non-method member access to skip member-method cache counters, got hits=%d misses=%d", stats.MemberMethodCacheHits, stats.MemberMethodCacheMiss)
	}
}

func TestBytecodeVM_InterfaceMemberMethodCacheUsesValidatedDictionary(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	parentEnv := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env := runtime.NewEnvironmentWithValueCapacity(parentEnv, 0)
	program := &bytecodeProgram{}
	method := bytecodeInterfaceCacheProbeNative(11)
	iface := bytecodeInterfaceCacheProbeValue(runtime.StringValue{Val: "beta"}, method)
	vm := newBytecodeVM(interp, env)

	for idx := 0; idx < 2; idx++ {
		got := bytecodeExecInterfaceProbe(t, vm, program, iface)
		if got != 11 {
			t.Fatalf("probe call %d = %d, want 11", idx+1, got)
		}
	}
	if iface.BoundMethod != nil || iface.BoundMethodName != "" {
		t.Fatalf("bytecode interface member cache should not materialize bound wrapper, got %q %#v", iface.BoundMethodName, iface.BoundMethod)
	}
	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheHits == 0 {
		t.Fatalf("expected validated interface member-method cache hit, got hits=%d misses=%d", stats.MemberMethodCacheHits, stats.MemberMethodCacheMiss)
	}
}

func TestBytecodeVM_InterfaceMemberMethodCacheRejectsShadowedDictionaryEntry(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	program := &bytecodeProgram{}
	first := bytecodeInterfaceCacheProbeNative(11)
	second := bytecodeInterfaceCacheProbeNative(22)
	iface := bytecodeInterfaceCacheProbeValue(runtime.StringValue{Val: "beta"}, first)
	vm := newBytecodeVM(interp, env)

	if got := bytecodeExecInterfaceProbe(t, vm, program, iface); got != 11 {
		t.Fatalf("first probe call = %d, want 11", got)
	}
	if got := bytecodeExecInterfaceProbe(t, vm, program, iface); got != 11 {
		t.Fatalf("cached first probe call = %d, want 11", got)
	}
	interfaceValueSetMethod(iface, "probe", second)
	if got := bytecodeExecInterfaceProbe(t, vm, program, iface); got != 22 {
		t.Fatalf("shadowed probe call = %d, want 22", got)
	}
	if got := bytecodeExecInterfaceProbe(t, vm, program, iface); got != 22 {
		t.Fatalf("cached shadowed probe call = %d, want 22", got)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheHits < 2 || stats.MemberMethodCacheMiss < 2 {
		t.Fatalf("expected hits and stale-entry miss, got hits=%d misses=%d", stats.MemberMethodCacheHits, stats.MemberMethodCacheMiss)
	}
	if iface.BoundMethod != nil || iface.BoundMethodName != "" {
		t.Fatalf("bytecode interface member cache should not materialize bound wrapper, got %q %#v", iface.BoundMethodName, iface.BoundMethod)
	}
}

func TestBytecodeVM_MemberMethodCacheAllowsNonImplRuntimeData(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env.SetRuntimeData("generator-payload")
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store with non-impl runtime data")
	}
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); !ok || cached.template != method {
		t.Fatalf("expected member-method cache lookup with non-impl runtime data, got %#v/%v", cached.template, ok)
	}
}

func TestBytecodeVM_MemberMethodCacheAllowsImplRuntimeDataWithContextValidation(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	firstCtx := &implMethodContext{
		implName:      "SImpl",
		interfaceName: "Pinger",
		target:        ast.Ty("S"),
		methods:       map[string]runtime.Value{},
	}
	secondCtx := &implMethodContext{
		implName:      "OtherSImpl",
		interfaceName: "Pinger",
		target:        ast.Ty("S"),
		methods:       map[string]runtime.Value{},
	}
	env.SetRuntimeData(firstCtx)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store with impl runtime data")
	}
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); !ok || cached.template != method {
		t.Fatalf("expected member-method cache hit with unchanged impl runtime data, got %#v/%v", cached.template, ok)
	}

	env.SetRuntimeData(secondCtx)
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); ok || cached.template != nil {
		t.Fatalf("expected member-method cache miss after impl runtime data change, got %#v/%v", cached.template, ok)
	}

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache restock with updated impl runtime data")
	}
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); !ok || cached.template != method {
		t.Fatalf("expected member-method cache hit after restock with updated impl runtime data, got %#v/%v", cached.template, ok)
	}
}

func TestBytecodeVM_MemberMethodCacheWorksInDeepEnv(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env := runtime.NewEnvironmentWithValueCapacity(parent, 0)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store in deep lexical env")
	}
	cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver)
	if !ok || cached.template != method {
		t.Fatalf("expected member-method cache hit in deep lexical env, got %#v/%v", cached.template, ok)
	}
}

func TestBytecodeVM_MemberMethodCacheInvalidatesOnDeepEnvShapeChange(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env := runtime.NewEnvironmentWithValueCapacity(parent, 0)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store in deep lexical env")
	}
	if _, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); !ok {
		t.Fatalf("expected member-method cache hit before shape change")
	}

	parent.Define("ping", runtime.NewSmallInt(1, runtime.IntegerI32))
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); ok {
		t.Fatalf("expected member-method cache miss after same-name binding shape change, got %#v", cached.template)
	}
}

func TestBytecodeVM_MemberMethodCacheIgnoresUnrelatedDeepEnvShapeChange(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env := runtime.NewEnvironmentWithValueCapacity(parent, 0)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store in deep lexical env")
	}
	if _, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); !ok {
		t.Fatalf("expected member-method cache hit before unrelated shape change")
	}

	parent.Define("other", runtime.NewSmallInt(1, runtime.IntegerI32))
	cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver)
	if !ok || cached.template != method {
		t.Fatalf("expected member-method cache hit after unrelated shape change, got %#v/%v", cached.template, ok)
	}
}

func TestBytecodeVM_MemberMethodCacheInvalidatesOnDeepEnvMemberOwnerChange(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	parent.Define("ping", runtime.NewSmallInt(1, runtime.IntegerI32))
	env := runtime.NewEnvironmentWithValueCapacity(parent, 0)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		0,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store in deep lexical env")
	}
	if _, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); !ok {
		t.Fatalf("expected member-method cache hit before owner mutation")
	}

	if err := parent.Assign("ping", bytecodeMemberCachePingFunction(env)); err != nil {
		t.Fatalf("assign same-name binding: %v", err)
	}
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 0, "ping", true, receiver); ok {
		t.Fatalf("expected member-method cache miss after same-name owner revision change, got %#v", cached.template)
	}
}

func TestBytecodeVM_StaticMemberCallCacheHitsRepeatedCallSite(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	boxDef := ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)
	valueFn := ast.Fn(
		"value",
		nil,
		[]ast.Statement{ast.Int(7)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		boxDef,
		ast.Methods(ast.Ty("Box"), []*ast.FunctionDefinition{valueFn}, nil, nil),
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.Loop(
			ast.Iff(ast.Bin(">=", ast.ID("i"), ast.Int(4)), ast.Brk(nil, nil)),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("sum"),
				ast.Bin("+", ast.ID("sum"), ast.CallExpr(ast.Member(ast.ID("Box"), "value"))),
			),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("i"),
				ast.Bin("+", ast.ID("i"), ast.Int(1)),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode static member call result mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 28 {
		t.Fatalf("static member loop result = %#v, want 28", got)
	}

	stats := interp.BytecodeStats()
	if stats.CallMemberStaticCacheMisses == 0 {
		t.Fatalf("expected static member cache miss on first call, got stats=%#v", stats)
	}
	if stats.CallMemberStaticCacheHits < 3 {
		t.Fatalf("expected repeated static member cache hits, got hits=%d misses=%d",
			stats.CallMemberStaticCacheHits,
			stats.CallMemberStaticCacheMisses,
		)
	}
	if stats.CallMemberStaticInlineHits == 0 {
		t.Fatalf("expected cached static call to continue using inline dispatch")
	}
}

func TestBytecodeVM_StaticMemberCallUsesStackDirectFunctionForCompiledThunk(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"value",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
			[]ast.Statement{ast.Int(0)},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
		Bytecode: CompiledThunk(func(_ *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("compiled static value args = %d, want 1", len(args))
			}
			arg, ok := args[0].(runtime.IntegerValue)
			if !ok {
				return nil, fmt.Errorf("compiled static value arg = %T, want integer", args[0])
			}
			return runtime.NewSmallInt(arg.BigInt().Int64()+3, runtime.IntegerI32), nil
		}),
	}
	receiver := &runtime.StructDefinitionValue{Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)}
	vm.stack = []runtime.Value{receiver, runtime.NewSmallInt(4, runtime.IntegerI32)}

	newProg, err := vm.execStaticMemberCallable(
		fn,
		bytecodeInstruction{name: "value", argCount: 1},
		0,
		1,
		nil,
		nil,
		&bytecodeProgram{},
		true,
	)
	if err != nil {
		t.Fatalf("compiled static call failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("compiled static call unexpectedly inlined")
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	got, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok || got.BigInt().Int64() != 7 {
		t.Fatalf("compiled static call result = %#v, want 7", vm.stack[0])
	}

	stats := interp.BytecodeStats()
	if stats.DirectFunctionStackHits != 1 {
		t.Fatalf("DirectFunctionStackHits = %d, want 1", stats.DirectFunctionStackHits)
	}
	if stats.InlineResolvedMissNoBytecode != 1 {
		t.Fatalf("InlineResolvedMissNoBytecode = %d, want 1", stats.InlineResolvedMissNoBytecode)
	}
	if stats.CallMemberStaticGenericHits != 1 {
		t.Fatalf("CallMemberStaticGenericHits = %d, want 1", stats.CallMemberStaticGenericHits)
	}
}

func bytecodeMemberCacheStructReceiver() *runtime.StructInstanceValue {
	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	return &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: structDef},
		Fields: map[string]runtime.Value{
			"n": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
	}
}

func bytecodeMemberCachePingFunction(env *runtime.Environment) *runtime.FunctionValue {
	return &runtime.FunctionValue{
		Declaration: ast.Fn(
			"ping",
			[]*ast.FunctionParameter{
				ast.Param("self", ast.Ty("S")),
			},
			[]ast.Statement{ast.Int(1)},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
	}
}

func bytecodeInterfaceCacheProbeNative(value int64) *runtime.NativeFunctionValue {
	return &runtime.NativeFunctionValue{
		Name:       "probe",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("probe args = %d, want injected receiver", len(args))
			}
			if _, ok := args[0].(runtime.StringValue); !ok {
				return nil, fmt.Errorf("probe receiver = %T, want String", args[0])
			}
			return runtime.NewSmallInt(value, runtime.IntegerI32), nil
		},
	}
}

func bytecodeInterfaceCacheProbeValue(receiver runtime.Value, method runtime.Value) *runtime.InterfaceValue {
	return &runtime.InterfaceValue{
		Interface: &runtime.InterfaceDefinitionValue{
			Node: ast.Iface(
				"Probe",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"probe",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("String"))},
						ast.Ty("i32"),
						nil,
						nil,
						nil,
					),
				},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
		Underlying: receiver,
		Methods: map[string]runtime.Value{
			"probe": method,
		},
	}
}

func bytecodeExecInterfaceProbe(t *testing.T, vm *bytecodeVM, program *bytecodeProgram, iface *runtime.InterfaceValue) int64 {
	t.Helper()
	vm.ip = 0
	vm.stack = []runtime.Value{iface}
	newProg, err := vm.execCallMember(bytecodeInstruction{name: "probe", argCount: 0}, program)
	if err != nil {
		t.Fatalf("bytecode direct interface member call failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("unexpected inline program for direct interface member call")
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	intVal, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("stack result = %T (%#v), want integer", vm.stack[0], vm.stack[0])
	}
	return intVal.BigInt().Int64()
}
