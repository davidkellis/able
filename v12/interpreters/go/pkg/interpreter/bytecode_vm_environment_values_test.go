package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_EnvironmentNameWritesMaterializeRawI32Values(t *testing.T) {
	for _, tc := range []struct {
		name string
		op   bytecodeOp
		seed bool
	}{
		{name: "declare", op: bytecodeOpDeclareName},
		{name: "assign", op: bytecodeOpAssignName, seed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := NewBytecode()
			env := runtime.NewEnvironment(interp.GlobalEnvironment())
			if tc.seed {
				env.Define("value", runtime.NewSmallInt(0, runtime.IntegerI32))
			}
			vm := newBytecodeVM(interp, env)
			_, err := vm.run(&bytecodeProgram{instructions: []bytecodeInstruction{
				{op: bytecodeOpConst, value: bytecodeRawI32SlotCachedValue(37)},
				{op: tc.op, name: "value"},
			}})
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			got, err := env.Get("value")
			if err != nil {
				t.Fatalf("get value: %v", err)
			}
			integer, ok := got.(runtime.IntegerValue)
			if !ok || integer.TypeSuffix != runtime.IntegerI32 || integer.BigInt().Int64() != 37 {
				t.Fatalf("environment value = %#v, want materialized i32 37", got)
			}
		})
	}
}

func TestMaterializeRuntimeValueBoxesBytecodeRawI32(t *testing.T) {
	raw := bytecodeRawI32SlotCachedValue(41)
	materializer, ok := raw.(interface {
		MaterializeRuntimeValue() runtime.Value
	})
	if !ok {
		t.Fatalf("raw i32 carrier %T does not expose boundary materialization", raw)
	}
	got := materializer.MaterializeRuntimeValue()
	integer, ok := got.(runtime.IntegerValue)
	if !ok || integer.TypeSuffix != runtime.IntegerI32 || integer.BigInt().Int64() != 41 {
		t.Fatalf("materialized value = %#v, want i32 41", got)
	}
}

func TestBytecodeVM_CompletedRunMaterializesRawScalarResult(t *testing.T) {
	for _, tc := range []struct {
		name       string
		returnNode ast.Node
	}{
		{name: "result"},
		{name: "return signal", returnNode: ast.Int(41)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := NewBytecode()
			vm := newBytecodeVM(interp, interp.GlobalEnvironment())
			got, err := vm.run(&bytecodeProgram{instructions: []bytecodeInstruction{
				{op: bytecodeOpConst, value: bytecodeRawI64ResultValue(41)},
				{op: bytecodeOpReturn, node: tc.returnNode},
			}})
			if signal, ok := err.(returnSignal); ok {
				got = signal.value
				err = nil
			}
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			integer, ok := got.(runtime.IntegerValue)
			if !ok || integer.TypeSuffix != runtime.IntegerI64 || integer.BigInt().Int64() != 41 {
				t.Fatalf("completed run value = %#v, want materialized i64 41", got)
			}
		})
	}
}
