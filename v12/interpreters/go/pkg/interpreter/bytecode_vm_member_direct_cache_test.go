package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_MemberMethodDirectCacheSurvivesHotEviction(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected first member-method cache store")
	}
	if _, ok := vm.storeCachedMemberMethod(
		program,
		2,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected second member-method cache store")
	}
	vm.memberMethodCache = nil

	cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver)
	if !ok || cached.template != method {
		t.Fatalf("expected direct member-method cache hit after hot eviction, got %#v/%v", cached.template, ok)
	}
	if !vm.memberMethodHot.valid || vm.memberMethodHot.program != program || vm.memberMethodHot.ip != 1 {
		t.Fatalf("expected direct hit to promote hot entry, got %#v", vm.memberMethodHot)
	}
}

func TestBytecodeVM_MemberMethodDirectCacheRespectsShapeInvalidation(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env := runtime.NewEnvironmentWithValueCapacity(parent, 0)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store")
	}
	if _, ok := vm.storeCachedMemberMethod(
		program,
		2,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected hot-evicting member-method cache store")
	}
	parent.Define("ping", runtime.NewSmallInt(1, runtime.IntegerI32))
	vm.memberMethodCache = nil

	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver); ok {
		t.Fatalf("expected stale direct member-method cache miss after shape change, got %#v", cached.template)
	}
}

func TestBytecodeVM_MemberMethodDirectCacheRefreshesAfterUnrelatedShapeChange(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	env := runtime.NewEnvironmentWithValueCapacity(parent, 0)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store")
	}
	if _, ok := vm.storeCachedMemberMethod(
		program,
		2,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected hot-evicting member-method cache store")
	}
	parent.Define("other", runtime.NewSmallInt(1, runtime.IntegerI32))
	vm.memberMethodCache = nil

	cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver)
	if !ok || cached.template != method {
		t.Fatalf("expected direct member-method cache hit after unrelated shape change, got %#v/%v", cached.template, ok)
	}
	direct := vm.memberMethodDirect[bytecodeMemberMethodDirectIndex(1)]
	if direct.bindingShapeVersion != env.BindingShapeRevision() {
		t.Fatalf("direct binding shape version = %d, want %d", direct.bindingShapeVersion, env.BindingShapeRevision())
	}
}

func TestBytecodeVM_MemberMethodDirectCacheReusesAcrossSiblingCallScopes(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	firstEnv := runtime.NewEnvironmentWithSingleBinding(parent, 1, "argument", runtime.NewSmallInt(1, runtime.IntegerI32))
	secondEnv := runtime.NewEnvironmentWithSingleBinding(parent, 1, "argument", runtime.NewSmallInt(2, runtime.IntegerI32))
	vm := newBytecodeVM(interp, firstEnv)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(parent)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store in first sibling scope")
	}
	vm.env = secondEnv
	cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver)
	if !ok || cached.template != method {
		t.Fatalf("expected dependency-valid cache hit in sibling scope, got %#v/%v", cached.template, ok)
	}
}

func TestBytecodeVM_MemberMethodDirectCacheRejectsSiblingNameShadowing(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	firstEnv := runtime.NewEnvironmentWithSingleBinding(parent, 1, "argument", runtime.NewSmallInt(1, runtime.IntegerI32))
	vm := newBytecodeVM(interp, firstEnv)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(parent)

	if _, ok := vm.storeCachedMemberMethod(
		program,
		1,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatalf("expected member-method cache store in first sibling scope")
	}

	vm.env = runtime.NewEnvironmentWithSingleBinding(parent, 1, "ping", runtime.NewSmallInt(9, runtime.IntegerI32))
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver); ok {
		t.Fatalf("expected sibling member-name shadowing to invalidate cache, got %#v", cached.template)
	}

	vm.env = runtime.NewEnvironmentWithSingleBinding(parent, 1, "S", runtime.NewSmallInt(9, runtime.IntegerI32))
	if cached, ok := vm.lookupCachedMemberMethodEntry(program, 1, "ping", true, receiver); ok {
		t.Fatalf("expected sibling receiver-type shadowing to invalidate cache, got %#v", cached.template)
	}
}
