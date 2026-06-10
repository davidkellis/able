package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeIntegerLoadCarrierForSlot(t *testing.T) {
	smallPointer := runtime.NewSmallInt(7, runtime.IntegerI64)
	bigValue := runtime.NewBigIntValue(new(big.Int).Lsh(big.NewInt(1), 80), runtime.IntegerI128)
	tests := []struct {
		name  string
		value runtime.Value
		want  bytecodeIntegerLoadCarrier
	}{
		{name: "raw i64 cell", value: &bytecodeRawI64SlotCell{Val: 3}, want: bytecodeIntegerLoadCarrierRawI64Cell},
		{name: "raw integer cell", value: &bytecodeRawIntegerSlotCell{Raw: 3, TypeSuffix: runtime.IntegerU16}, want: bytecodeIntegerLoadCarrierRawIntegerCell},
		{name: "raw i32 value", value: bytecodeRawI32SlotValue(3), want: bytecodeIntegerLoadCarrierRawI32Value},
		{name: "raw integer value", value: bytecodeRawIntegerValue{Raw: 3, TypeSuffix: runtime.IntegerU16}, want: bytecodeIntegerLoadCarrierRawIntegerValue},
		{name: "small pointer", value: &smallPointer, want: bytecodeIntegerLoadCarrierSmallIntegerPointer},
		{name: "small value", value: runtime.NewSmallInt(3, runtime.IntegerI64), want: bytecodeIntegerLoadCarrierSmallIntegerValue},
		{name: "big integer", value: bigValue, want: bytecodeIntegerLoadCarrierBigInteger},
		{name: "mismatch", value: runtime.StringValue{Val: "3"}, want: bytecodeIntegerLoadCarrierOther},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := &bytecodeVM{slots: []runtime.Value{test.value}}
			if got := bytecodeIntegerLoadCarrierForSlot(vm, 0); got != test.want {
				t.Fatalf("carrier = %s, want %s", got.String(), test.want.String())
			}
		})
	}

	registerVM := &bytecodeVM{
		slots:            []runtime.Value{nil},
		i32Registers:     []int32{3},
		i32RegisterValid: []bool{true},
	}
	if got := bytecodeIntegerLoadCarrierForSlot(registerVM, 0); got != bytecodeIntegerLoadCarrierI32Register {
		t.Fatalf("register carrier = %s", got.String())
	}

	sidecarVM := &bytecodeVM{slots: []runtime.Value{nil}}
	sidecarVM.restoreValueSlotI32Frame(sidecarVM.slots, []int32{3}, []bool{true})
	if got := bytecodeIntegerLoadCarrierForSlot(sidecarVM, 0); got != bytecodeIntegerLoadCarrierI32Sidecar {
		t.Fatalf("sidecar carrier = %s", got.String())
	}
}

func TestBytecodeIntegerLoadConsumerFindsSourceEnclosingOperation(t *testing.T) {
	left := ast.ID("left")
	right := ast.ID("right")
	add := ast.Bin("+", left, right)
	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpLoadSlot, node: left},
		{op: bytecodeOpLoadSlot, node: right},
		{op: bytecodeOpBinary, operator: "+", node: add},
	}}
	if got := bytecodeIntegerLoadConsumerForProgram(program, 0); got.consumer != bytecodeIntegerLoadConsumerArithmetic || got.op != bytecodeOpBinary {
		t.Fatalf("left consumer = %+v, want Binary arithmetic", got)
	}
	if got := bytecodeIntegerLoadConsumerForProgram(program, 1); got.consumer != bytecodeIntegerLoadConsumerArithmetic || got.op != bytecodeOpBinary {
		t.Fatalf("right consumer = %+v, want Binary arithmetic", got)
	}
}

func TestBytecodeIntegerLoadConsumerAttributesArraySlotOperandRole(t *testing.T) {
	receiver := ast.ID("items")
	index := ast.ID("index")
	value := ast.ID("value")
	call := ast.NewFunctionCall(ast.Member(receiver, "write_slot"), []ast.Expression{index, value}, nil, false)
	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpLoadSlot, node: receiver},
		{op: bytecodeOpLoadSlot, node: index},
		{op: bytecodeOpLoadSlot, node: value},
		{
			op: bytecodeOpCallMemberArraySlot, argCount: 2, node: call,
			memberFastPath: bytecodeMemberMethodFastPathArrayWriteSlot,
		},
	}}
	wants := []bytecodeIntegerLoadOperandRole{
		bytecodeIntegerLoadOperandRoleReceiver,
		bytecodeIntegerLoadOperandRoleIndex,
		bytecodeIntegerLoadOperandRoleValue,
	}
	for ip, want := range wants {
		got := bytecodeIntegerLoadConsumerForProgram(program, ip)
		if got.op != bytecodeOpCallMemberArraySlot || got.operation != "write-slot" || got.role != want {
			t.Fatalf("load %d use = %+v, want CallMemberArraySlot role %s", ip, got, want.String())
		}
	}
}

func TestBytecodeProgramReachRecordsProvenIntegerLoadShape(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	identifier := ast.ID("count")
	add := ast.Bin("+", identifier, ast.Int(1))
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpLoadSlot, target: 0, node: identifier},
			{op: bytecodeOpConst, value: runtime.NewSmallInt(1, runtime.IntegerI64), node: add.Right},
			{op: bytecodeOpBinary, operator: "+", node: add},
		},
		frameLayout: &bytecodeFrameLayout{slotKinds: []bytecodeCellKind{bytecodeCellKindValue}},
	}
	program = interp.annotateBytecodeProgramReachWithScalarChecks(
		program, "function", "add", nil,
		[]bytecodeSimpleTypeCheck{bytecodeSimpleTypeCheckI64},
	)
	interp.recordBytecodeProgramEntry(program)
	vm := &bytecodeVM{interp: interp, slots: []runtime.Value{&bytecodeRawI64SlotCell{Val: 3}}}
	vm.recordProvenIntegerLoadShape(program, 0, &program.instructions[0])

	rows := interp.BytecodeStats().ProgramReach
	if len(rows) != 1 || len(rows[0].IntegerLoadShapes) != 1 {
		t.Fatalf("unexpected shape rows: %+v", rows)
	}
	shape := rows[0].IntegerLoadShapes[0]
	if shape.Carrier != "raw-i64-slot-cell" || shape.Consumer != "arithmetic" || shape.ConsumerOpcode != "Binary" || shape.ConsumerOperation != "" || shape.ConsumerOperandRole != "" || shape.DynamicInstructions != 1 {
		t.Fatalf("unexpected shape: %+v", shape)
	}

	interp.ResetBytecodeStats()
	interp.recordBytecodeProgramEntry(program)
	if shapes := interp.BytecodeStats().ProgramReach[0].IntegerLoadShapes; len(shapes) != 0 {
		t.Fatalf("shapes after reset: %+v", shapes)
	}
}
