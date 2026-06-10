package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeArrayOwnershipProgramMetadata(t *testing.T) {
	tests := []struct {
		name            string
		instructions    []bytecodeInstruction
		observesArrays  bool
		releaseEligible bool
	}{
		{
			name:         "ordinary_scalar",
			instructions: []bytecodeInstruction{{op: bytecodeOpConst, value: runtime.NewSmallInt(1, runtime.IntegerI32)}},
		},
		{
			name:            "array_literal",
			instructions:    []bytecodeInstruction{{op: bytecodeOpArrayLiteral, argCount: 0}},
			observesArrays:  true,
			releaseEligible: true,
		},
		{
			name:            "canonical_array_new_opcode",
			instructions:    []bytecodeInstruction{{op: bytecodeOpCallMemberArrayNew}},
			observesArrays:  true,
			releaseEligible: true,
		},
		{
			name: "canonical_static_array_new",
			instructions: []bytecodeInstruction{{
				op:   bytecodeOpCallStaticMember,
				name: "new",
				node: ast.NewFunctionCall(ast.Member(ast.ID("Array"), "new"), nil, nil, false),
			}},
			observesArrays:  true,
			releaseEligible: true,
		},
		{
			name: "kernel_array_literal",
			instructions: []bytecodeInstruction{{
				op:   bytecodeOpStructLiteralNamedFast,
				node: ast.StructLit(nil, false, "Array", nil, nil),
			}},
			observesArrays:  true,
			releaseEligible: true,
		},
		{
			name: "capture_barrier",
			instructions: []bytecodeInstruction{
				{op: bytecodeOpArrayLiteral, argCount: 0},
				{op: bytecodeOpMakeFunction},
			},
			observesArrays: true,
		},
		{
			name: "dynamic_barrier",
			instructions: []bytecodeInstruction{
				{op: bytecodeOpArrayLiteral, argCount: 0},
				{op: bytecodeOpDynImport},
			},
			observesArrays: true,
		},
		{
			name: "spawn_barrier",
			instructions: []bytecodeInstruction{
				{op: bytecodeOpArrayLiteral, argCount: 0},
				{op: bytecodeOpSpawn},
			},
			observesArrays: true,
		},
		{
			name: "aggregate_barrier",
			instructions: []bytecodeInstruction{
				{op: bytecodeOpArrayLiteral, argCount: 0},
				{op: bytecodeOpMapLiteral},
			},
			observesArrays: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: tc.instructions})
			metadata := program.arrayOwnershipMetadata
			if got := metadata.observesArrays(); got != tc.observesArrays {
				t.Fatalf("observes arrays = %v, want %v", got, tc.observesArrays)
			}
			if got := metadata.releaseEligible(); got != tc.releaseEligible {
				t.Fatalf("release eligible = %v, want %v", got, tc.releaseEligible)
			}
		})
	}
}

func TestBytecodeArrayOwnershipProfileActivatesEligibleInlineCallee(t *testing.T) {
	interp := NewBytecode()
	profile := interp.enableBytecodeArrayOwnershipProfile()
	defer interp.disableBytecodeArrayOwnershipProfile()
	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(vm)

	vm.pushCallFrame(1, nil, nil, vm.env, nil, 0, 0, false, false)
	if vm.arrayOwnershipObserver != nil {
		t.Fatal("ineligible caller frame should not allocate an ownership observer")
	}
	eligible := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayLiteral, argCount: 0},
	}})
	vm.ensureBytecodeArrayOwnershipForProgram(eligible)
	observer := vm.arrayOwnershipObserver
	if observer == nil || observer.current == nil {
		t.Fatal("eligible inline callee should activate ownership observation")
	}
	if len(profile.observers) != 1 {
		t.Fatalf("profile observers = %d, want 1", len(profile.observers))
	}
	if vm.callFrames[0].arrayOwnershipParent == nil {
		t.Fatal("eligible inline callee should receive a reconstructed caller context")
	}
}

func TestBytecodeArrayOwnershipMetadataDetectsLoweredStaticArrayNew(t *testing.T) {
	module := mustParseModuleSource(t, `
fn make() {
  Array.new()
}
`)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(firstFunctionDefinition(t, module))
	if err != nil {
		t.Fatalf("lower static Array.new: %v", err)
	}
	if !program.arrayOwnershipMetadata.observesArrays() {
		t.Fatalf("lowered static Array.new metadata = %#v, want canonical creation", program.arrayOwnershipMetadata)
	}
}
