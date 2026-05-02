package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CallMemberUsesCanonicalStringByteIteratorNextFastPath(t *testing.T) {
	interp, vm, iface, iter := setupCanonicalStringByteIteratorCallMemberFast(t)

	vm.stack = []runtime.Value{iface}
	_, err := vm.execCallMember(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("call-member string byte iterator next fast path failed: %v", err)
	}
	if !valuesEqual(vm.stack[0], runtime.NewSmallInt(97, runtime.IntegerU8)) {
		t.Fatalf("next result = %#v, want u8 97", vm.stack[0])
	}
	if offset, ok := bytecodeI32StructField(iter, "offset"); !ok || offset != 1 {
		t.Fatalf("offset after next = %d/%v, want 1/true", offset, ok)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{iface}
	_, err = vm.execCallMember(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("call-member string byte iterator next end fast path failed: %v", err)
	}
	if _, ok := vm.stack[0].(runtime.IteratorEndValue); !ok {
		t.Fatalf("next end result = %#v, want IteratorEnd", vm.stack[0])
	}
}

func TestBytecodeVM_CanonicalStringByteIteratorNextFastPathRequiresIteratorU8Interface(t *testing.T) {
	_, vm, iface, _ := setupCanonicalStringByteIteratorCallMemberFast(t)

	iface.InterfaceArgs = nil
	vm.stack = []runtime.Value{iface}
	_, handled, err := vm.execCanonicalStringByteIteratorNextCallMemberFast(
		bytecodeInstruction{name: "next", argCount: 0},
		0,
		nil,
	)
	if err != nil {
		t.Fatalf("non-Iterator-u8 fast path check failed: %v", err)
	}
	if handled {
		t.Fatalf("fast path handled interface without canonical Iterator u8 args")
	}
}

func setupCanonicalStringByteIteratorCallMemberFast(t *testing.T) (*Interpreter, *bytecodeVM, *runtime.InterfaceValue, *runtime.StructInstanceValue) {
	t.Helper()

	interp := NewBytecode()
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

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{runtime.StringValue{Val: "a"}}
	_, handled, err := vm.execCachedMemberMethodFastPath(
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
	iface, ok := vm.stack[0].(*runtime.InterfaceValue)
	if !ok || iface == nil {
		t.Fatalf("bytes result = %#v, want Iterator interface", vm.stack[0])
	}
	iter, ok := iface.Underlying.(*runtime.StructInstanceValue)
	if !ok || iter == nil {
		t.Fatalf("bytes underlying = %#v, want RawStringBytesIter", iface.Underlying)
	}
	return interp, vm, iface, iter
}
