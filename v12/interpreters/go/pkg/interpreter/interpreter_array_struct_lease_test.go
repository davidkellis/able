package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestArrayStructStorageHandleAssignmentTransfersLeaseAcrossModes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		assign func(*Interpreter, *runtime.StructInstanceValue, runtime.Value) (runtime.Value, error)
	}{
		{
			name: "tree_walker",
			assign: func(interp *Interpreter, inst *runtime.StructInstanceValue, value runtime.Value) (runtime.Value, error) {
				return assignStructMember(interp, inst, ast.ID("storage_handle"), value, ast.AssignmentAssign, "", false)
			},
		},
		{
			name: "bytecode",
			assign: func(interp *Interpreter, inst *runtime.StructInstanceValue, value runtime.Value) (runtime.Value, error) {
				vm := newBytecodeVM(interp, interp.GlobalEnvironment())
				return vm.assignMemberValue(inst, ast.ID("storage_handle"), value, ast.AssignmentAssign, "", false)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := NewBytecode()
			first := runtime.ArrayStoreNew()
			second := runtime.ArrayStoreNew()
			inst := newArrayStructLeaseInstance(first)
			if err := runtime.ArrayStoreTrackStructInstanceLease(inst, first); err != nil {
				t.Fatalf("seed first struct lease: %v", err)
			}
			before := runtime.ArrayStoreLeaseStatsSnapshot()
			firstOwners := before.OwnersByHandle[first]
			secondOwners := before.OwnersByHandle[second]

			value := runtime.NewSmallInt(second, runtime.IntegerI64)
			if _, err := tc.assign(interp, inst, value); err != nil {
				t.Fatalf("assign storage_handle: %v", err)
			}
			after := runtime.ArrayStoreLeaseStatsSnapshot()
			if after.OwnersByHandle[first] != firstOwners-1 || after.OwnersByHandle[second] != secondOwners+1 {
				t.Fatalf("lease transfer = %#v, want first=%d second=%d", after, firstOwners-1, secondOwners+1)
			}
			field, ok := inst.Fields["storage_handle"].(runtime.IntegerValue)
			if !ok || field.Int64Fast() != second {
				t.Fatalf("storage_handle field = %#v, want %d", inst.Fields["storage_handle"], second)
			}
			if err := runtime.ArrayStoreReleaseStructInstanceLease(inst); err != nil {
				t.Fatalf("release transferred struct lease: %v", err)
			}
		})
	}
}

func TestArrayStructInstanceConversionTracksSidecarAndValueView(t *testing.T) {
	interp := New()
	handle := runtime.ArrayStoreNew()
	inst := newArrayStructLeaseInstance(handle)
	before := runtime.ArrayStoreLeaseStatsSnapshot()
	owners := before.OwnersByHandle[handle]

	value, err := interp.arrayValueFromStructInstance(inst)
	if err != nil {
		t.Fatalf("arrayValueFromStructInstance: %v", err)
	}
	after := runtime.ArrayStoreLeaseStatsSnapshot()
	if after.OwnersByHandle[handle] != owners+2 {
		t.Fatalf("struct conversion lease count = %#v, want %d", after, owners+2)
	}
	if err := runtime.ArrayStoreReleaseArrayValueLease(value); err != nil {
		t.Fatalf("release array view lease: %v", err)
	}
	if err := runtime.ArrayStoreReleaseStructInstanceLease(inst); err != nil {
		t.Fatalf("release struct lease: %v", err)
	}
}

func TestStringifyArrayStructTracksSidecarLease(t *testing.T) {
	interp := New()
	handle := runtime.ArrayStoreNew()
	inst := newArrayStructLeaseInstance(handle)
	before := runtime.ArrayStoreLeaseStatsSnapshot().OwnersByHandle[handle]

	rendered, ok := interp.stringifyArrayStruct(inst)
	if !ok || rendered != "[]" {
		t.Fatalf("stringify Array struct = (%q, %v), want ([], true)", rendered, ok)
	}
	after := runtime.ArrayStoreLeaseStatsSnapshot().OwnersByHandle[handle]
	if after != before+1 {
		t.Fatalf("stringify Array struct leases = %d, want %d", after, before+1)
	}
	if err := runtime.ArrayStoreReleaseStructInstanceLease(inst); err != nil {
		t.Fatalf("release stringified Array struct lease: %v", err)
	}
}

func TestPositionalArrayStructStorageHandleAssignmentTransfersLease(t *testing.T) {
	for _, tc := range []struct {
		name   string
		assign func(*Interpreter, *runtime.StructInstanceValue, runtime.Value) (runtime.Value, error)
	}{
		{
			name: "tree_walker",
			assign: func(interp *Interpreter, inst *runtime.StructInstanceValue, value runtime.Value) (runtime.Value, error) {
				return assignStructMember(interp, inst, ast.Int(2), value, ast.AssignmentAssign, "", false)
			},
		},
		{
			name: "bytecode",
			assign: func(interp *Interpreter, inst *runtime.StructInstanceValue, value runtime.Value) (runtime.Value, error) {
				vm := newBytecodeVM(interp, interp.GlobalEnvironment())
				return vm.assignMemberValue(inst, ast.Int(2), value, ast.AssignmentAssign, "", false)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := runtime.ArrayStoreNew()
			second := runtime.ArrayStoreNew()
			inst := newPositionalArrayStructLeaseInstance(first)
			if err := runtime.ArrayStoreTrackStructInstanceLease(inst, first); err != nil {
				t.Fatalf("seed first positional struct lease: %v", err)
			}
			before := runtime.ArrayStoreLeaseStatsSnapshot()
			firstOwners := before.OwnersByHandle[first]
			secondOwners := before.OwnersByHandle[second]

			if _, err := tc.assign(NewBytecode(), inst, runtime.NewSmallInt(second, runtime.IntegerI64)); err != nil {
				t.Fatalf("assign positional storage_handle: %v", err)
			}
			after := runtime.ArrayStoreLeaseStatsSnapshot()
			if after.OwnersByHandle[first] != firstOwners-1 || after.OwnersByHandle[second] != secondOwners+1 {
				t.Fatalf("positional lease transfer = %#v, want first=%d second=%d", after, firstOwners-1, secondOwners+1)
			}
			if err := runtime.ArrayStoreReleaseStructInstanceLease(inst); err != nil {
				t.Fatalf("release positional struct lease: %v", err)
			}
		})
	}
}

func newArrayStructLeaseInstance(handle int64) *runtime.StructInstanceValue {
	definition := &runtime.StructDefinitionValue{Node: ast.StructDef("Array", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("i64"), "storage_handle"),
		ast.FieldDef(ast.Ty("i32"), "length"),
		ast.FieldDef(ast.Ty("i32"), "capacity"),
	}, ast.StructKindNamed, nil, nil, false)}
	return &runtime.StructInstanceValue{
		Definition: definition,
		Fields: map[string]runtime.Value{
			"storage_handle": runtime.NewSmallInt(handle, runtime.IntegerI64),
			"length":         runtime.NewSmallInt(0, runtime.IntegerI32),
			"capacity":       runtime.NewSmallInt(0, runtime.IntegerI32),
		},
	}
}

func newPositionalArrayStructLeaseInstance(handle int64) *runtime.StructInstanceValue {
	definition := &runtime.StructDefinitionValue{Node: ast.StructDef("Array", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("i32"), "length"),
		ast.FieldDef(ast.Ty("i32"), "capacity"),
		ast.FieldDef(ast.Ty("i64"), "storage_handle"),
	}, ast.StructKindNamed, nil, nil, false)}
	return &runtime.StructInstanceValue{
		Definition: definition,
		Positional: []runtime.Value{
			runtime.NewSmallInt(0, runtime.IntegerI32),
			runtime.NewSmallInt(0, runtime.IntegerI32),
			runtime.NewSmallInt(handle, runtime.IntegerI64),
		},
	}
}
