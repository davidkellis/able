package semanticabi

import (
	"reflect"
	"testing"
	"unsafe"

	ableruntime "able/interpreter-go/pkg/runtime"
)

func TestCellIsSixteenBytesAndPointerFree(t *testing.T) {
	if size := unsafe.Sizeof(Cell{}); size != 16 {
		t.Fatalf("Cell size = %d, want 16", size)
	}
	typeOfCell := reflect.TypeOf(Cell{})
	for index := 0; index < typeOfCell.NumField(); index++ {
		kind := typeOfCell.Field(index).Type.Kind()
		if kind == reflect.Pointer || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Func || kind == reflect.String {
			t.Fatalf("Cell field %s has pointer-bearing kind %s", typeOfCell.Field(index).Name, kind)
		}
	}
	id, err := NewObjectID(7, 3)
	if err != nil {
		t.Fatal(err)
	}
	if id.Index() != 7 || id.Generation() != 3 || !id.Valid() {
		t.Fatalf("ObjectID round trip = index %d generation %d valid %v", id.Index(), id.Generation(), id.Valid())
	}
	cell, err := ReferenceCell(TagKindArray, 9, id)
	if err != nil {
		t.Fatal(err)
	}
	if cell.Payload != uint64(id) || cell.Aux != 9 {
		t.Fatalf("reference cell = %#v", cell)
	}
	if _, err := ImmediateCell(TagKindArray, 0, 0); err == nil {
		t.Fatal("ImmediateCell accepted shared-heap tag")
	}
	if _, err := ReferenceCell(TagKindInteger, 0, id); err == nil {
		t.Fatal("ReferenceCell accepted immediate tag")
	}
}

func TestGeneratedKindManifestExactlyMatchesRuntimeKinds(t *testing.T) {
	runtimeKinds := map[string]ableruntime.Kind{
		"KindString": ableruntime.KindString, "KindBool": ableruntime.KindBool,
		"KindChar": ableruntime.KindChar, "KindNil": ableruntime.KindNil,
		"KindVoid": ableruntime.KindVoid, "KindInteger": ableruntime.KindInteger,
		"KindFloat": ableruntime.KindFloat, "KindArray": ableruntime.KindArray,
		"KindHashMap": ableruntime.KindHashMap, "KindHasher": ableruntime.KindHasher,
		"KindFunction": ableruntime.KindFunction, "KindNativeFunction": ableruntime.KindNativeFunction,
		"KindFunctionOverload": ableruntime.KindFunctionOverload, "KindStructDefinition": ableruntime.KindStructDefinition,
		"KindTypeRef": ableruntime.KindTypeRef, "KindStructInstance": ableruntime.KindStructInstance,
		"KindInterfaceDefinition": ableruntime.KindInterfaceDefinition, "KindInterfaceValue": ableruntime.KindInterfaceValue,
		"KindUnionDefinition": ableruntime.KindUnionDefinition, "KindPackage": ableruntime.KindPackage,
		"KindDynPackage": ableruntime.KindDynPackage, "KindDynRef": ableruntime.KindDynRef,
		"KindError": ableruntime.KindError, "KindHostHandle": ableruntime.KindHostHandle,
		"KindBoundMethod": ableruntime.KindBoundMethod, "KindNativeBoundMethod": ableruntime.KindNativeBoundMethod,
		"KindImplementationNamespace": ableruntime.KindImplementationNamespace, "KindFuture": ableruntime.KindFuture,
		"KindIterator": ableruntime.KindIterator, "KindIteratorEnd": ableruntime.KindIteratorEnd,
		"KindPartialFunction": ableruntime.KindPartialFunction,
	}
	if len(KindManifest) != len(runtimeKinds) {
		t.Fatalf("kind manifest has %d entries, runtime has %d", len(KindManifest), len(runtimeKinds))
	}
	classCounts := map[KindClass]int{}
	seenRuntime := make(map[ableruntime.Kind]bool, len(runtimeKinds))
	for index, descriptor := range KindManifest {
		runtimeKind, ok := runtimeKinds[descriptor.Name]
		if !ok {
			t.Fatalf("manifest contains unknown runtime kind %q", descriptor.Name)
		}
		if runtimeKind != ableruntime.Kind(descriptor.RuntimeOrdinal) {
			t.Fatalf("%s runtime ordinal = %d, manifest = %d", descriptor.Name, runtimeKind, descriptor.RuntimeOrdinal)
		}
		if descriptor.Tag != uint32(index+1) {
			t.Fatalf("%s tag = %d, want %d", descriptor.Name, descriptor.Tag, index+1)
		}
		if seenRuntime[runtimeKind] {
			t.Fatalf("runtime kind %d appears more than once", runtimeKind)
		}
		seenRuntime[runtimeKind] = true
		classCounts[descriptor.Class]++
	}
	if classCounts[KindImmediate] != 7 || classCounts[KindSharedHeap] != 20 || classCounts[KindHostRegistry] != 4 {
		t.Fatalf("kind class counts = %#v, want immediate=7 heap=20 host=4", classCounts)
	}
}

func TestGeneratedOpcodeManifestIsDenseAndTyped(t *testing.T) {
	if len(OpManifest) != 38 {
		t.Fatalf("opcode manifest has %d entries, want 38", len(OpManifest))
	}
	for index, descriptor := range OpManifest {
		if descriptor.Opcode != uint16(index+1) {
			t.Fatalf("opcode %s = %d, want %d", descriptor.Name, descriptor.Opcode, index+1)
		}
		for _, operand := range descriptor.Operands {
			if operand < OperandImmediate || operand > OperandCallTarget {
				t.Fatalf("opcode %s has invalid operand class %d", descriptor.Name, operand)
			}
		}
		if descriptor.Variadic > OperandCallTarget {
			t.Fatalf("opcode %s has invalid variadic operand class %d", descriptor.Name, descriptor.Variadic)
		}
		for _, write := range descriptor.Writes {
			if int(write) >= len(descriptor.Operands) || descriptor.Operands[write] != OperandRegister {
				t.Fatalf("opcode %s has invalid write operand %d", descriptor.Name, write)
			}
		}
	}
}

func TestGeneratedHeapAndHostLayoutsAreExhaustive(t *testing.T) {
	shared := make(map[uint32]bool)
	internal := 0
	for index, layout := range ObjectLayoutManifest {
		if layout.LayoutID != uint16(index+1) {
			t.Fatalf("layout %s id = %d, want %d", layout.Name, layout.LayoutID, index+1)
		}
		if layout.Mutability != LayoutImmutable && layout.Mutability != LayoutMutable {
			t.Fatalf("layout %s has invalid mutability %d", layout.Name, layout.Mutability)
		}
		if layout.RuntimeTag == 0 {
			internal++
			continue
		}
		kind, ok := KindByTag(layout.RuntimeTag)
		if !ok || kind.Class != KindSharedHeap {
			t.Fatalf("layout %s tag %d is not shared heap", layout.Name, layout.RuntimeTag)
		}
		if shared[layout.RuntimeTag] {
			t.Fatalf("shared tag %d has more than one layout", layout.RuntimeTag)
		}
		shared[layout.RuntimeTag] = true
		for _, field := range layout.Fields {
			if field.Name == "" || field.Storage < FieldScalar || field.Storage > FieldObjects {
				t.Fatalf("layout %s has invalid field %#v", layout.Name, field)
			}
		}
	}
	if len(shared) != 20 || internal != 7 {
		t.Fatalf("layout counts shared=%d internal=%d, want 20 and 7", len(shared), internal)
	}
	for _, kind := range KindManifest {
		if kind.Class == KindSharedHeap && !shared[kind.Tag] {
			t.Fatalf("shared kind %s has no layout", kind.Name)
		}
	}

	hosts := make(map[uint32]bool)
	for _, layout := range HostLayoutManifest {
		kind, ok := KindByTag(layout.RuntimeTag)
		if !ok || kind.Class != KindHostRegistry {
			t.Fatalf("host layout %s tag %d is not host registry", layout.Name, layout.RuntimeTag)
		}
		if hosts[layout.RuntimeTag] {
			t.Fatalf("host tag %d has more than one layout", layout.RuntimeTag)
		}
		hosts[layout.RuntimeTag] = true
	}
	if len(hosts) != 4 {
		t.Fatalf("host layouts = %d, want 4", len(hosts))
	}
	for _, kind := range KindManifest {
		if kind.Class == KindHostRegistry && !hosts[kind.Tag] {
			t.Fatalf("host kind %s has no layout", kind.Name)
		}
	}
}
