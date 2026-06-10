package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_MemberMethodCacheCapacityKeepsInlineCaches(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{}
	receiver := bytecodeMemberCacheStructReceiver()
	method := bytecodeMemberCachePingFunction(env)

	vm.memberMethodCache = make(
		map[bytecodeMemberMethodCacheKey]bytecodeMemberMethodCacheEntry,
		bytecodeMemberMethodCacheMaxEntries,
	)
	for ip := 0; ip < bytecodeMemberMethodCacheMaxEntries; ip++ {
		vm.memberMethodCache[bytecodeMemberMethodCacheKey{
			program: program,
			ip:      ip,
			env:     env,
			member:  "retained",
		}] = bytecodeMemberMethodCacheEntry{}
	}

	const newIP = bytecodeMemberMethodCacheMaxEntries + 17
	if _, ok := vm.storeCachedMemberMethod(
		program,
		newIP,
		"ping",
		true,
		receiver,
		runtime.BoundMethodValue{Receiver: receiver, Method: method},
	); !ok {
		t.Fatal("expected saturated member-method cache store to keep inline entry")
	}
	if got := len(vm.memberMethodCache); got != bytecodeMemberMethodCacheMaxEntries {
		t.Fatalf("member-method cache length = %d, want %d", got, bytecodeMemberMethodCacheMaxEntries)
	}
	if !vm.memberMethodHot.valid || vm.memberMethodHot.ip != newIP {
		t.Fatalf("hot member-method cache was not refreshed: %#v", vm.memberMethodHot)
	}
	direct := vm.memberMethodDirect[bytecodeMemberMethodDirectIndex(newIP)]
	if !direct.valid || direct.ip != newIP {
		t.Fatalf("direct member-method cache was not refreshed: %#v", direct)
	}
}
