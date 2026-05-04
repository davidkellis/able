package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayMemberFastPathDetectsCanonicalMethods(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	lenKey := bytecodeMemberMethodCacheKey{
		member:        "len",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverArray,
	}
	lenDef := ast.Fn("len", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
	}, []ast.Statement{ast.Int(0)}, ast.Ty("i32"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		lenDef: "/tmp/able/v12/kernel/src/kernel.able",
	})

	got := vm.memberMethodFastPathFor(lenKey, &runtime.FunctionValue{Declaration: lenDef})
	if got != bytecodeMemberMethodFastPathArrayLen {
		t.Fatalf("len fast path = %d, want ArrayLen", got)
	}

	lenDefWrongOrigin := ast.Fn("len", nil, []ast.Statement{ast.Int(0)}, ast.Ty("i32"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		lenDefWrongOrigin: "/tmp/project/kernel.able",
	})
	got = vm.memberMethodFastPathFor(lenKey, &runtime.FunctionValue{Declaration: lenDefWrongOrigin})
	if got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("custom Array.len should not receive canonical fast path, got %d", got)
	}

	key := bytecodeMemberMethodCacheKey{
		member:        "get",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverArray,
	}
	getDef := ast.Fn("get", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Nullable(ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		getDef: "/tmp/able-stdlib/src/collections/array.able",
	})

	got = vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: getDef})
	if got != bytecodeMemberMethodFastPathArrayGet {
		t.Fatalf("fast path = %d, want ArrayGet", got)
	}

	getDefWrongOrigin := ast.Fn("get", nil, []ast.Statement{ast.Nil()}, ast.Nullable(ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		getDefWrongOrigin: "/tmp/project/collections/array.able",
	})
	got = vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: getDefWrongOrigin})
	if got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("custom Array.get should not receive canonical fast path, got %d", got)
	}

	indexGetDef := ast.Fn("get", nil, []ast.Statement{ast.Nil()}, ast.Result(ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		indexGetDef: "/tmp/able-stdlib/src/collections/array.able",
	})
	got = vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: indexGetDef})
	if got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("Index.get result method should not receive nullable Array.get fast path, got %d", got)
	}

	pushKey := bytecodeMemberMethodCacheKey{
		member:        "push",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverArray,
	}
	pushDef := ast.Fn("push", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("value", ast.Ty("T")),
	}, []ast.Statement{ast.Nil()}, ast.Ty("void"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		pushDef: "/tmp/able/v12/kernel/src/kernel.able",
	})
	got = vm.memberMethodFastPathFor(pushKey, &runtime.FunctionValue{Declaration: pushDef})
	if got != bytecodeMemberMethodFastPathArrayPush {
		t.Fatalf("push fast path = %d, want ArrayPush", got)
	}

	pushDefWrongOrigin := ast.Fn("push", nil, []ast.Statement{ast.Nil()}, ast.Ty("void"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		pushDefWrongOrigin: "/tmp/project/kernel.able",
	})
	got = vm.memberMethodFastPathFor(pushKey, &runtime.FunctionValue{Declaration: pushDefWrongOrigin})
	if got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("custom Array.push should not receive canonical fast path, got %d", got)
	}

	pushDefWrongReturn := ast.Fn("push", nil, []ast.Statement{ast.Nil()}, ast.Ty("nil"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		pushDefWrongReturn: "/tmp/able/v12/kernel/src/kernel.able",
	})
	got = vm.memberMethodFastPathFor(pushKey, &runtime.FunctionValue{Declaration: pushDefWrongReturn})
	if got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("Array.push with wrong return type should not receive fast path, got %d", got)
	}
}

func TestBytecodeVM_CanonicalStdlibOriginCheckDoesNotAllocate(t *testing.T) {
	var ok bool
	allocs := testing.AllocsPerRun(1000, func() {
		ok = isCanonicalAbleStdlibOrigin("/tmp/able-stdlib/src/collections/array.able", "collections/array.able")
	})
	if !ok {
		t.Fatalf("expected canonical stdlib origin to match")
	}
	if allocs != 0 {
		t.Fatalf("canonical stdlib origin check allocated %.2f times per run", allocs)
	}
	if isCanonicalAbleStdlibOrigin("/tmp/project/src/collections/array.able", "collections/array.able") {
		t.Fatalf("custom origin should not match canonical stdlib path")
	}
}

func TestBytecodeVM_ArrayMemberFastPathLenAndGetSemantics(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)

	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayLen,
		bytecodeInstruction{name: "len", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array len fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected array len fast path to handle call")
	}
	intVal, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok || intVal.Int64Fast() != 2 || intVal.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("len result = %#v, want i32 2", vm.stack[0])
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayGet,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array get fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected array get fast path to handle call")
	}
	want := runtime.StringValue{Val: "one"}
	if !valuesEqual(vm.stack[0], want) {
		t.Fatalf("get result = %#v, want %#v", vm.stack[0], want)
	}
	if state, tracked := bytecodeTrackedArrayState(arr); !tracked || len(state.Values) != 2 {
		t.Fatalf("array get test receiver should use tracked state, tracked=%v state=%#v", tracked, state)
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(-1, runtime.IntegerI32)}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayGet,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array get negative fast path failed: %v", err)
	}
	if !handled || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("negative get result = %#v, want nil", vm.stack[0])
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.StringValue{Val: "two"}}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayPush,
		bytecodeInstruction{name: "push", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array push fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected array push fast path to handle call")
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("push result = %#v, want void", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after push: %v", err)
	}
	if len(state.Values) != 3 {
		t.Fatalf("array length after push = %d, want 3", len(state.Values))
	}
	if want := (runtime.StringValue{Val: "two"}); !valuesEqual(state.Values[2], want) {
		t.Fatalf("array pushed value = %#v, want %#v", state.Values[2], want)
	}
}

func TestBytecodeVM_CanonicalArrayGetOverloadFastPath(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)

	nullableGet := ast.Fn("get", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Nullable(ast.Ty("T")), nil, nil, false, false)
	resultGet := ast.Fn("get", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Result(ast.Ty("T")), nil, nil, false, false)
	customGet := ast.Fn("get", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Nullable(ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		nullableGet: "/tmp/able-stdlib/src/collections/array.able",
		resultGet:   "/tmp/able-stdlib/src/collections/array.able",
		customGet:   "/tmp/project/collections/array.able",
	})

	overload := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{
		{Declaration: resultGet, Closure: env, MethodPriority: -1},
		{Declaration: nullableGet, Closure: env},
	}}
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	_, handled, err := vm.execCanonicalArrayGetOverloadMemberFast(
		overload,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("canonical Array.get overload fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected canonical Array.get overload fast path to handle call")
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("Array.get overload result = %#v, want %#v", vm.stack[0], want)
	}

	rejected := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{
		{Declaration: customGet, Closure: env},
		{Declaration: resultGet, Closure: env, MethodPriority: -1},
	}}
	if vm.isCanonicalNullableArrayGetOverload(rejected) {
		t.Fatalf("custom Array.get overload should not use canonical fast path")
	}
}

func TestBytecodeVM_CanonicalArrayGetOverloadCachesFunctionPair(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	nullableGet := ast.Fn("get", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Nullable(ast.Ty("T")), nil, nil, false, false)
	resultGet := ast.Fn("get", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Result(ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		nullableGet: "/tmp/able-stdlib/src/collections/array.able",
		resultGet:   "/tmp/able-stdlib/src/collections/array.able",
	})

	resultFn := &runtime.FunctionValue{Declaration: resultGet, Closure: env, MethodPriority: -1}
	nullableFn := &runtime.FunctionValue{Declaration: nullableGet, Closure: env}
	first := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{resultFn, nullableFn}}
	if !vm.isCanonicalNullableArrayGetOverload(first) {
		t.Fatalf("expected first canonical Array.get overload wrapper to validate")
	}
	if vm.arrayGetOverloadPairNullable != nullableFn || vm.arrayGetOverloadPairResult != resultFn || !vm.arrayGetOverloadPairOK {
		t.Fatalf("expected canonical Array.get pair cache to store underlying functions")
	}

	second := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{resultFn, nullableFn}}
	if !vm.isCanonicalNullableArrayGetOverload(second) {
		t.Fatalf("expected second canonical Array.get overload wrapper to validate from function-pair cache")
	}
	if vm.arrayGetOverloadHot != second {
		t.Fatalf("expected pointer hot cache to refresh to second overload wrapper")
	}
}

func TestBytecodeVM_CanonicalArrayGetCallCacheGuardsClosureEnv(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	closure := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, closure)
	program := &bytecodeProgram{}
	instr := bytecodeInstruction{name: "get", argCount: 1}
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)

	if vm.lookupCachedCanonicalArrayGetCall(program, 3, instr, arr) {
		t.Fatalf("unexpected canonical Array.get call-cache hit before store")
	}
	vm.storeCachedCanonicalArrayGetCall(program, 3, instr, arr)
	if !vm.lookupCachedCanonicalArrayGetCall(program, 3, instr, arr) {
		t.Fatalf("expected canonical Array.get call-cache hit after store")
	}
	vm.storeCachedCanonicalArrayGetCall(program, 5, instr, arr)
	vm.arrayGetCallCache = nil
	if !vm.lookupCachedCanonicalArrayGetCall(program, 3, instr, arr) {
		t.Fatalf("expected canonical Array.get hot cache to retain older call site")
	}
	if !vm.lookupCachedCanonicalArrayGetCall(program, 5, instr, arr) {
		t.Fatalf("expected canonical Array.get hot cache to retain newer call site")
	}

	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	_, handled, err := vm.execArrayGetMemberFast(instr, 0, 1, nil)
	if err != nil {
		t.Fatalf("cached canonical Array.get fast call failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected cached canonical Array.get fast call to handle array get")
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("cached canonical Array.get result = %#v, want %#v", vm.stack[0], want)
	}

	closure.Define("marker", runtime.NewSmallInt(1, runtime.IntegerI32))
	if vm.lookupCachedCanonicalArrayGetCall(program, 3, instr, arr) {
		t.Fatalf("expected closure env revision change to invalidate canonical Array.get call cache")
	}
	vm.storeCachedCanonicalArrayGetCall(program, 3, instr, arr)
	closure.SetRuntimeData(&implMethodContext{implName: "I"})
	if vm.lookupCachedCanonicalArrayGetCall(program, 3, instr, arr) {
		t.Fatalf("expected runtime impl context to bypass canonical Array.get call cache")
	}
}

func TestBytecodeVM_CanonicalArrayGetCallCacheFeedsExecCallMember(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{}
	instr := bytecodeInstruction{name: "get", argCount: 1}
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)

	vm.storeCachedCanonicalArrayGetCall(program, 3, instr, arr)
	vm.arrayGetCallCache = nil
	vm.memberMethodCache = nil
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{}
	vm.ip = 3
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}

	newProg, err := vm.execCallMember(instr, program)
	if err != nil {
		t.Fatalf("execCallMember cached canonical Array.get failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("expected direct cached Array.get call to stay in current program")
	}
	if vm.ip != 4 {
		t.Fatalf("expected cached Array.get call to advance ip to 4, got %d", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected one Array.get result, got stack %#v", vm.stack)
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("cached canonical Array.get result = %#v, want %#v", vm.stack[0], want)
	}
}

func TestBytecodeVM_StaticArrayNewFastPathSemantics(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	arrayDef := ast.StructDef("Array", nil, ast.StructKindNamed, nil, nil, false)
	newDef := ast.Fn("new", nil, []ast.Statement{ast.Nil()}, ast.Gen(ast.Ty("Array"), ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		arrayDef: "/tmp/able/v12/kernel/src/kernel.able",
		newDef:   "/tmp/able/v12/kernel/src/kernel.able",
	})

	receiver := &runtime.StructDefinitionValue{Node: arrayDef}
	vm.stack = []runtime.Value{receiver}
	_, handled, err := vm.execStaticArrayNewMemberFast(
		bytecodeInstruction{name: "new", argCount: 0},
		receiver,
		&runtime.FunctionValue{Declaration: newDef},
		0,
		nil,
	)
	if err != nil {
		t.Fatalf("static Array.new fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected static Array.new fast path to handle canonical call")
	}
	arr, ok := vm.stack[0].(*runtime.ArrayValue)
	if !ok || arr == nil {
		t.Fatalf("Array.new result = %#v, want ArrayValue", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure Array.new result state: %v", err)
	}
	if len(state.Values) != 0 || state.Capacity != 0 {
		t.Fatalf("Array.new state = len %d cap %d, want empty", len(state.Values), state.Capacity)
	}

	wrongOrigin := ast.Fn("new", nil, []ast.Statement{ast.Nil()}, ast.Gen(ast.Ty("Array"), ast.Ty("T")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		arrayDef:    "/tmp/able/v12/kernel/src/kernel.able",
		wrongOrigin: "/tmp/project/kernel.able",
	})
	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{receiver}
	if _, handled, err := vm.execStaticArrayNewMemberFast(
		bytecodeInstruction{name: "new", argCount: 0},
		receiver,
		&runtime.FunctionValue{Declaration: wrongOrigin},
		0,
		nil,
	); err != nil || handled {
		t.Fatalf("custom Array.new fast path handled=%v err=%v, want unhandled nil error", handled, err)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{receiver, runtime.NewSmallInt(1, runtime.IntegerI32)}
	if _, handled, err := vm.execStaticArrayNewMemberFast(
		bytecodeInstruction{name: "new", argCount: 1},
		receiver,
		&runtime.FunctionValue{Declaration: newDef},
		0,
		nil,
	); err != nil || handled {
		t.Fatalf("Array.new with args fast path handled=%v err=%v, want unhandled nil error", handled, err)
	}
}

func TestBytecodeVM_StringMemberFastPathDetectsCanonicalMethods(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	cases := []struct {
		name       string
		returnType ast.TypeExpression
		want       bytecodeMemberMethodFastPathKind
	}{
		{name: "len_bytes", returnType: ast.Ty("u64"), want: bytecodeMemberMethodFastPathStringLenBytes},
		{name: "contains", returnType: ast.Ty("bool"), want: bytecodeMemberMethodFastPathStringContains},
		{name: "replace", returnType: ast.Ty("String"), want: bytecodeMemberMethodFastPathStringReplace},
		{name: "bytes", returnType: ast.Gen(ast.Ty("Iterator"), ast.Ty("u8")), want: bytecodeMemberMethodFastPathStringBytes},
	}
	for _, tc := range cases {
		key := bytecodeMemberMethodCacheKey{
			member:        tc.name,
			preferMethods: true,
			receiverKind:  bytecodeMemberReceiverString,
		}
		def := ast.Fn(tc.name, nil, []ast.Statement{ast.Nil()}, tc.returnType, nil, nil, false, false)
		interp.SetNodeOrigins(map[ast.Node]string{
			def: "/tmp/able-stdlib/src/text/string.able",
		})
		if got := vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: def}); got != tc.want {
			t.Fatalf("%s fast path = %d, want %d", tc.name, got, tc.want)
		}

		wrongOrigin := ast.Fn(tc.name, nil, []ast.Statement{ast.Nil()}, tc.returnType, nil, nil, false, false)
		interp.SetNodeOrigins(map[ast.Node]string{
			wrongOrigin: "/tmp/project/text/string.able",
		})
		if got := vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: wrongOrigin}); got != bytecodeMemberMethodFastPathNone {
			t.Fatalf("custom String.%s should not receive canonical fast path, got %d", tc.name, got)
		}
	}

	lenKey := bytecodeMemberMethodCacheKey{
		member:        "len_bytes",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverString,
	}
	lenWrongReturn := ast.Fn("len_bytes", nil, []ast.Statement{ast.Nil()}, ast.Ty("i32"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		lenWrongReturn: "/tmp/able-stdlib/src/text/string.able",
	})
	if got := vm.memberMethodFastPathFor(lenKey, &runtime.FunctionValue{Declaration: lenWrongReturn}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("String.len_bytes with wrong return type should not receive fast path, got %d", got)
	}

	bytesKey := bytecodeMemberMethodCacheKey{
		member:        "bytes",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverString,
	}
	bytesWrongReturn := ast.Fn("bytes", nil, []ast.Statement{ast.Nil()}, ast.Ty("StringBytesIter"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		bytesWrongReturn: "/tmp/able-stdlib/src/text/string.able",
	})
	if got := vm.memberMethodFastPathFor(bytesKey, &runtime.FunctionValue{Declaration: bytesWrongReturn}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("String.bytes with wrong return type should not receive fast path, got %d", got)
	}
}

func TestBytecodeVM_StringMemberFastPathSemantics(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()

	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{runtime.StringValue{Val: "science"}}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringLenBytes,
		bytecodeInstruction{name: "len_bytes", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string len_bytes fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string len_bytes fast path to handle call")
	}
	intVal, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok || intVal.Int64Fast() != 7 || intVal.TypeSuffix != runtime.IntegerU64 {
		t.Fatalf("len_bytes result = %#v, want u64 7", vm.stack[0])
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{
		runtime.StringValue{Val: "ceiling"},
		runtime.StringValue{Val: "ei"},
	}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringContains,
		bytecodeInstruction{name: "contains", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string contains fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string contains fast path to handle call")
	}
	boolVal, ok := vm.stack[0].(runtime.BoolValue)
	if !ok || !boolVal.Val {
		t.Fatalf("contains result = %#v, want true", vm.stack[0])
	}

	vm = newBytecodeVM(interp, env)
	receiver := runtime.StringValue{Val: "science"}
	vm.stack = []runtime.Value{
		&receiver,
		runtime.StringValue{Val: "en"},
		runtime.StringValue{Val: "X"},
	}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringReplace,
		bytecodeInstruction{name: "replace", argCount: 2},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string replace fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string replace fast path to handle call")
	}
	strVal, ok := vm.stack[0].(runtime.StringValue)
	if !ok || strVal.Val != "sciXce" {
		t.Fatalf("replace result = %#v, want sciXce", vm.stack[0])
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{
		runtime.StringValue{Val: "science"},
		runtime.StringValue{Val: ""},
		runtime.StringValue{Val: "X"},
	}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringReplace,
		bytecodeInstruction{name: "replace", argCount: 2},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string replace empty-needle fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string replace empty-needle fast path to handle call")
	}
	strVal, ok = vm.stack[0].(runtime.StringValue)
	if !ok || strVal.Val != "science" {
		t.Fatalf("replace empty-needle result = %#v, want science", vm.stack[0])
	}

	interp = NewBytecode()
	module := mustParseModuleSource(t, `
interface Iterator T {
  fn next(self: Self) -> T | IteratorEnd
}

struct IteratorEnd {}

private struct RawStringBytesIter {
  bytes: Array u8,
  offset: i32,
  len_bytes: i32
}

impl Iterator u8 for RawStringBytesIter {
  fn next(self: Self) -> u8 | IteratorEnd {
    IteratorEnd {}
  }
}
`)
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate iterator setup: %v", err)
	}
	iterDef, found := interp.lookupStructDefinition("RawStringBytesIter")
	if !found || iterDef == nil || iterDef.Node == nil {
		t.Fatalf("RawStringBytesIter definition missing after setup")
	}
	origins := map[ast.Node]string{
		iterDef.Node: "/tmp/able-stdlib/src/text/string.able",
	}
	for _, stmt := range module.Body {
		switch node := stmt.(type) {
		case *ast.InterfaceDefinition:
			if node.ID != nil && node.ID.Name == "Iterator" {
				origins[node] = "/tmp/able-stdlib/src/core/iteration.able"
			}
		case *ast.ImplementationDefinition:
			if node.InterfaceName == nil || node.InterfaceName.Name != "Iterator" {
				continue
			}
			for _, def := range node.Definitions {
				if def != nil && def.ID != nil && def.ID.Name == "next" {
					origins[def] = "/tmp/able-stdlib/src/text/string.able"
				}
			}
		}
	}
	interp.SetNodeOrigins(origins)
	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{runtime.StringValue{Val: "az"}}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringBytes,
		bytecodeInstruction{name: "bytes", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string bytes fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string bytes fast path to handle valid UTF-8")
	}
	if !vm.stringBytesIteratorNextSet || vm.stringBytesIteratorNextMethod == nil {
		t.Fatalf("expected String.bytes fast path to cache canonical Iterator.next")
	}
	iface, ok := vm.stack[0].(*runtime.InterfaceValue)
	if !ok || iface == nil {
		t.Fatalf("bytes result = %#v, want Iterator interface", vm.stack[0])
	}
	if iface.Methods == nil || iface.Methods["next"] == nil {
		t.Fatalf("bytes interface methods = %#v, want cached next method", iface.Methods)
	}
	iter, ok := iface.Underlying.(*runtime.StructInstanceValue)
	if !ok || iter == nil || iter.Definition != iterDef {
		t.Fatalf("bytes underlying = %#v, want RawStringBytesIter", iface.Underlying)
	}
	arr, ok := iter.Fields["bytes"].(*runtime.ArrayValue)
	if !ok || arr == nil {
		t.Fatalf("bytes field = %#v, want Array u8", iter.Fields["bytes"])
	}
	size, err := runtime.ArrayStoreSize(arr.Handle)
	if err != nil {
		t.Fatalf("bytes array size failed: %v", err)
	}
	if size != 2 {
		t.Fatalf("bytes array size = %d, want 2", size)
	}
	raw, ok, err := runtime.ArrayStoreMonoReadU8IfAvailable(arr.Handle, 0)
	if err != nil {
		t.Fatalf("mono u8 read failed: %v", err)
	}
	if !ok || raw != 'a' {
		t.Fatalf("mono u8 read = %d/%v, want 97/true", raw, ok)
	}
	first, err := runtime.ArrayStoreRead(arr.Handle, 0)
	if err != nil {
		t.Fatalf("bytes array read 0 failed: %v", err)
	}
	second, err := runtime.ArrayStoreRead(arr.Handle, 1)
	if err != nil {
		t.Fatalf("bytes array read 1 failed: %v", err)
	}
	if !valuesEqual(first, runtime.NewSmallInt(97, runtime.IntegerU8)) ||
		!valuesEqual(second, runtime.NewSmallInt(122, runtime.IntegerU8)) {
		t.Fatalf("bytes array values = %#v/%#v, want [97, 122]", first, second)
	}
	if offset, ok := bytecodeI32StructField(iter, "offset"); !ok || offset != 0 {
		t.Fatalf("bytes offset = %d/%v, want 0/true", offset, ok)
	}
	if length, ok := bytecodeI32StructField(iter, "len_bytes"); !ok || length != 2 {
		t.Fatalf("bytes len_bytes = %d/%v, want 2/true", length, ok)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{runtime.StringValue{Val: string([]byte{0xff})}}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringBytes,
		bytecodeInstruction{name: "bytes", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("invalid UTF-8 fallback returned error: %v", err)
	}
	if handled {
		t.Fatalf("String.bytes fast path should fall back for invalid UTF-8")
	}
}

func TestBytecodeVM_StringByteIteratorNextFastPathDetectsCanonicalMethod(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	key := bytecodeMemberMethodCacheKey{
		member:        "next",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverStruct,
	}
	nextDef := ast.Fn("next", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
	}, []ast.Statement{ast.Nil()}, ast.UnionT(ast.Ty("u8"), ast.Ty("IteratorEnd")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		nextDef: "/tmp/able-stdlib/src/text/string.able",
	})
	if got := vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: nextDef}); got != bytecodeMemberMethodFastPathStringByteIteratorNext {
		t.Fatalf("String byte iterator next fast path = %d, want StringByteIteratorNext", got)
	}

	wrongOrigin := ast.Fn("next", nil, []ast.Statement{ast.Nil()}, ast.UnionT(ast.Ty("u8"), ast.Ty("IteratorEnd")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		wrongOrigin: "/tmp/project/text/string.able",
	})
	if got := vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: wrongOrigin}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("custom next should not receive string byte iterator fast path, got %d", got)
	}

	wrongReturn := ast.Fn("next", nil, []ast.Statement{ast.Nil()}, ast.UnionT(ast.Ty("char"), ast.Ty("IteratorEnd")), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		wrongReturn: "/tmp/able-stdlib/src/text/string.able",
	})
	if got := vm.memberMethodFastPathFor(key, &runtime.FunctionValue{Declaration: wrongReturn}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("non-u8 iterator next should not receive string byte iterator fast path, got %d", got)
	}
}

func TestBytecodeVM_StringByteIteratorNextFastPathSemantics(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	iterDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("RawStringBytesIter", []*ast.StructFieldDefinition{
			ast.FieldDef(ast.Gen(ast.Ty("Array"), ast.Ty("u8")), "bytes"),
			ast.FieldDef(ast.Ty("i32"), "offset"),
			ast.FieldDef(ast.Ty("i32"), "len_bytes"),
		}, ast.StructKindNamed, nil, nil, true),
	}
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(65, runtime.IntegerU8),
	}, 0)
	iter := &runtime.StructInstanceValue{
		Definition: iterDef,
		Fields: map[string]runtime.Value{
			"bytes":     arr,
			"offset":    runtime.NewSmallInt(0, runtime.IntegerI32),
			"len_bytes": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
	}
	iface := &runtime.InterfaceValue{Underlying: iter}

	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{iface}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringByteIteratorNext,
		bytecodeInstruction{name: "next", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string byte iterator next fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string byte iterator next fast path to handle call")
	}
	if !valuesEqual(vm.stack[0], runtime.NewSmallInt(65, runtime.IntegerU8)) {
		t.Fatalf("next result = %#v, want u8 65", vm.stack[0])
	}
	offset, ok := bytecodeI32StructField(iter, "offset")
	if !ok || offset != 1 {
		t.Fatalf("offset after next = %d/%v, want 1/true", offset, ok)
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{iface}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringByteIteratorNext,
		bytecodeInstruction{name: "next", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("string byte iterator next end fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected string byte iterator next end fast path to handle call")
	}
	if _, ok := vm.stack[0].(runtime.IteratorEndValue); !ok {
		t.Fatalf("next end result = %#v, want IteratorEnd", vm.stack[0])
	}
	offset, ok = bytecodeI32StructField(iter, "offset")
	if !ok || offset != 1 {
		t.Fatalf("offset after end = %d/%v, want 1/true", offset, ok)
	}

	monoIter := &runtime.StructInstanceValue{
		Definition: iterDef,
		Fields: map[string]runtime.Value{
			"bytes":     interp.newU8ArrayValueFromString("B"),
			"offset":    runtime.NewSmallInt(0, runtime.IntegerI32),
			"len_bytes": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
	}
	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{&runtime.InterfaceValue{Underlying: monoIter}}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathStringByteIteratorNext,
		bytecodeInstruction{name: "next", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono string byte iterator next fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono string byte iterator next fast path to handle call")
	}
	if !valuesEqual(vm.stack[0], runtime.NewSmallInt(66, runtime.IntegerU8)) {
		t.Fatalf("mono next result = %#v, want u8 66", vm.stack[0])
	}
	offset, ok = bytecodeI32StructField(monoIter, "offset")
	if !ok || offset != 1 {
		t.Fatalf("mono offset after next = %d/%v, want 1/true", offset, ok)
	}
}
