package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestFinalizeBytecodeProgramMetadataMarksFollowingPropagation(t *testing.T) {
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst},
		{op: bytecodeOpPropagation},
		{op: bytecodeOpConst},
	}})

	if program == nil {
		t.Fatalf("finalized program is nil")
	}
	if len(program.followedByPropagation) != len(program.instructions) {
		t.Fatalf("metadata length = %d, want %d", len(program.followedByPropagation), len(program.instructions))
	}
	if !program.followedByPropagation[0] {
		t.Fatalf("instruction 0 should be marked as followed by propagation")
	}
	if program.followedByPropagation[1] || program.followedByPropagation[2] {
		t.Fatalf("unexpected propagation metadata: %#v", program.followedByPropagation)
	}
}

func TestFinalizeBytecodeProgramMetadataAvoidsAllocationWithoutPropagation(t *testing.T) {
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst},
		{op: bytecodeOpReturn},
	}})

	if program == nil {
		t.Fatalf("finalized program is nil")
	}
	if program.followedByPropagation != nil {
		t.Fatalf("metadata allocated for program without propagation: %#v", program.followedByPropagation)
	}
}

func TestFinalizeBytecodeProgramMetadataMarksIntegerConstValidationNeed(t *testing.T) {
	withoutIntegerConst := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst, value: runtime.BoolValue{Val: true}},
		{op: bytecodeOpReturn},
	}})
	if !withoutIntegerConst.integerConstValidationKnown {
		t.Fatalf("expected finalized program to mark integer const validation metadata known")
	}
	if withoutIntegerConst.hasIntegerConstValidation {
		t.Fatalf("non-integer const should not require integer validation")
	}

	withIntegerConst := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst, value: runtime.NewSmallInt(1, runtime.IntegerI32)},
		{op: bytecodeOpReturn},
	}})
	if !withIntegerConst.integerConstValidationKnown || !withIntegerConst.hasIntegerConstValidation {
		t.Fatalf("integer const should require lazy integer validation metadata")
	}
}

func TestFinalizeBytecodeProgramMetadataPrebuildsSlotConstImmediateTable(t *testing.T) {
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{
			op:              bytecodeOpBinaryIntLessEqualSlotConst,
			intImmediate:    runtime.NewSmallInt(7, runtime.IntegerI32),
			hasIntImmediate: true,
		},
		{op: bytecodeOpReturn},
	}})
	if !program.slotConstIntImmTableKnown {
		t.Fatalf("expected finalized program to mark slot-const immediate metadata known")
	}
	if program.slotConstIntImmTable == nil {
		t.Fatalf("expected finalized program to prebuild slot-const immediate table")
	}
	if got, ok := bytecodeSlotConstImmediateAt(program.instructions[0], 0, program.slotConstIntImmTable); !ok {
		t.Fatalf("expected prebuilt slot-const immediate")
	} else if raw, rawOK := got.ToInt64(); !rawOK || raw != 7 {
		t.Fatalf("prebuilt slot-const immediate = %v ok=%v, want 7", got, rawOK)
	}

	noSlotConst := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst, value: runtime.BoolValue{Val: true}},
	}})
	if !noSlotConst.slotConstIntImmTableKnown {
		t.Fatalf("expected finalized program without slot-const ops to mark metadata known")
	}
	if noSlotConst.slotConstIntImmTable != nil {
		t.Fatalf("program without slot-const immediates should not allocate a table")
	}
}
