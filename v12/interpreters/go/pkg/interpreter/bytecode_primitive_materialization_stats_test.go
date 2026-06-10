package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodePrimitiveMaterializationAttributesReasonCarrierAndSource(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	interp.bytecodePrimitiveMaterializationStatsEnabled = true
	node := ast.ID("value")
	ast.SetSpan(node, ast.Span{Start: ast.Position{Line: 12, Column: 5}})
	interp.SetNodeOrigins(map[ast.Node]string{node: "/work/main.able"})
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{{op: bytecodeOpArrayLiteral, node: node}},
		reach:        &bytecodeProgramReach{kind: "function", name: "main"},
	}
	vm := &bytecodeVM{interp: interp, currentProgram: program}

	got := vm.materializePrimitiveValue(
		bytecodeMaterializationCandidateStatic,
		bytecodeMaterializationReasonCollection,
		bytecodeRawI32SlotCachedValue(42),
	)
	if !valuesEqual(got, runtime.NewSmallInt(42, runtime.IntegerI32)) {
		t.Fatalf("materialized value = %#v", got)
	}

	rows := interp.BytecodeStats().PrimitiveMaterializations
	if len(rows) != 1 {
		t.Fatalf("materialization rows = %+v, want one", rows)
	}
	row := rows[0]
	if row.Class != bytecodeMaterializationCandidateStatic ||
		row.Reason != bytecodeMaterializationReasonCollection ||
		row.Carrier != "i32_slot_value" ||
		row.Suffix != string(runtime.IntegerI32) ||
		row.Opcode != bytecodeOpName(bytecodeOpArrayLiteral) ||
		row.IP != 0 ||
		row.Program != "function:main" ||
		row.Origin != "/work/main.able" ||
		row.Line != 12 ||
		row.Column != 5 ||
		row.Count != 1 {
		t.Fatalf("materialization row = %+v", row)
	}
}

func TestBytecodePrimitiveMaterializationIsOptInAndCountsOnlyRawCarriers(t *testing.T) {
	interp := NewBytecode()
	vm := &bytecodeVM{interp: interp}
	vm.materializePrimitiveValue(
		bytecodeMaterializationCandidateStatic,
		bytecodeMaterializationReasonCast,
		bytecodeRawF64SlotValue(1.5),
	)
	if rows := interp.BytecodeStats().PrimitiveMaterializations; len(rows) != 0 {
		t.Fatalf("disabled materialization rows = %+v", rows)
	}

	interp.bytecodeStatsEnabled = true
	interp.bytecodePrimitiveMaterializationStatsEnabled = true
	vm.materializePrimitiveValue(
		bytecodeMaterializationCandidateStatic,
		bytecodeMaterializationReasonCast,
		runtime.NewSmallInt(1, runtime.IntegerI32),
	)
	if rows := interp.BytecodeStats().PrimitiveMaterializations; len(rows) != 0 {
		t.Fatalf("ordinary runtime value produced rows = %+v", rows)
	}

	vm.materializePrimitiveValue(
		bytecodeMaterializationCandidateStatic,
		bytecodeMaterializationReasonCast,
		bytecodeRawF64SlotValue(1.5),
	)
	rows := interp.BytecodeStats().PrimitiveMaterializations
	if len(rows) != 1 || rows[0].Carrier != "float_slot_value" || rows[0].Suffix != string(runtime.FloatF64) {
		t.Fatalf("raw float rows = %+v", rows)
	}
	interp.ResetBytecodeStats()
	if rows := interp.BytecodeStats().PrimitiveMaterializations; len(rows) != 0 {
		t.Fatalf("rows after reset = %+v", rows)
	}
}

func TestBytecodePrimitiveReturnMaterializationAttributesCallerConsumer(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	interp.bytecodePrimitiveMaterializationStatsEnabled = true
	node := ast.ID("consumer")
	ast.SetSpan(node, ast.Span{Start: ast.Position{Line: 22, Column: 7}})
	interp.SetNodeOrigins(map[ast.Node]string{node: "/work/caller.able"})
	caller := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpCall},
			{op: bytecodeOpStoreSlotI32, node: node},
		},
		reach: &bytecodeProgramReach{kind: "function", name: "caller"},
	}
	callee := &bytecodeProgram{
		instructions: []bytecodeInstruction{{op: bytecodeOpReturn}},
		reach:        &bytecodeProgramReach{kind: "function", name: "callee"},
	}
	vm := &bytecodeVM{
		interp:         interp,
		currentProgram: callee,
		callFrameKinds: []bytecodeCallFrameKind{bytecodeCallFrameKindFull},
		callFrames: []bytecodeCallFrame{{
			returnIP: 1,
			program:  caller,
		}},
	}

	vm.materializePrimitiveValue(
		bytecodeMaterializationCandidateStatic,
		bytecodeMaterializationReasonStaticReturn,
		bytecodeRawI32SlotCachedValue(42),
	)

	rows := interp.BytecodeStats().PrimitiveMaterializations
	if len(rows) != 1 {
		t.Fatalf("materialization rows = %+v, want one", rows)
	}
	row := rows[0]
	if row.ReturnFrame != "full" ||
		row.ConsumerOpcode != bytecodeOpName(bytecodeOpStoreSlotI32) ||
		row.ConsumerIP != 1 ||
		row.ConsumerProgram != "function:caller" ||
		row.ConsumerOrigin != "/work/caller.able" ||
		row.ConsumerLine != 22 ||
		row.ConsumerColumn != 7 {
		t.Fatalf("return consumer row = %+v", row)
	}
}

func TestBytecodePrimitiveMaterializationHasDedicatedObserver(t *testing.T) {
	t.Setenv(bytecodePrimitiveMaterializationStatsEnv, "1")
	t.Setenv("ABLE_BYTECODE_STATS", "")
	interp := NewBytecode()
	vm := &bytecodeVM{interp: interp}

	vm.materializePrimitiveValue(
		bytecodeMaterializationCandidateStatic,
		bytecodeMaterializationReasonStaticReturn,
		bytecodeRawI32SlotCachedValue(7),
	)

	snapshot := interp.BytecodeStats()
	if snapshot.Enabled {
		t.Fatalf("full bytecode stats unexpectedly enabled")
	}
	if !snapshot.PrimitiveMaterializationsEnabled || len(snapshot.PrimitiveMaterializations) != 1 {
		t.Fatalf("dedicated materialization snapshot = %+v", snapshot)
	}
}
