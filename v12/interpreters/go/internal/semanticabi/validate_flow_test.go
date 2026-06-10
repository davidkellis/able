package semanticabi

import (
	"strings"
	"testing"
)

func TestFlowValidatorRejectsMalformedControlAndDataFlow(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Image)
		want   string
	}{
		{name: "missing-terminator", mutate: func(image *Image) {
			image.Instructions = image.Instructions[:len(image.Instructions)-1]
			image.Functions[0].InstructionCount--
		}, want: "block 2 lacks terminator"},
		{name: "unreachable-block", mutate: func(image *Image) {
			image.Instructions[4].Operands[2] = 1
		}, want: "block 2 is unreachable"},
		{name: "undefined-register", mutate: func(image *Image) {
			image.Instructions[2].Operands[3] = 3
		}, want: "reads undefined register 3"},
		{name: "call-arity", mutate: func(image *Image) {
			image.Instructions[6].Operands = image.Instructions[6].Operands[:3]
		}, want: "call arity 0, target requires 1"},
		{name: "result-type", mutate: func(image *Image) {
			image.RegisterTypes[2] = 1
		}, want: "result register type 1 conflicts with declared type 0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			image := validFlowImage()
			test.mutate(image)
			first := Validate(image)
			second := Validate(image)
			if first == nil || second == nil || first.Error() != second.Error() || !strings.Contains(first.Error(), test.want) {
				t.Fatalf("Validate errors = %v and %v, want %q", first, second, test.want)
			}
		})
	}
}

func TestFlowValidatorAcceptsReachableTypedGraph(t *testing.T) {
	if err := Validate(validFlowImage()); err != nil {
		t.Fatal(err)
	}
}

func validFlowImage() *Image {
	return &Image{
		Header:    CurrentHeader(77),
		Strings:   []string{"flow.able", "flow", "main", "i32", "bool", "+", "callee"},
		Types:     []TypeRecord{{Name: 3}, {Name: 4}},
		Sources:   []SourceRecord{{File: 0, StartLine: 1, StartColumn: 1, EndLine: 4, EndColumn: 1, Callsite: NoIndex}},
		Constants: []ConstantRecord{{Tag: TagKindInteger, Data: []byte{0, 1}}},
		Packages:  []PackageRecord{{Name: 1, FunctionCount: 1}},
		Functions: []FunctionRecord{{
			Name: 2, Source: 0, RegisterCount: 4, BlockCount: 3,
			InstructionCount: 8, Flags: FunctionFlagShadowEligible | FunctionFlagFlowValidated,
		}},
		RegisterTypes: []uint32{0, 0, 0, 1},
		CallTargets: []CallTarget{{
			Kind: CallTargetLocal, PackageName: 1, OwnerType: NoIndex,
			Name: 6, Arity: 1, ReturnType: 0,
		}},
		Instructions: []Instruction{
			{Opcode: OpLoadConst, Source: 0, Block: 0, Operands: []uint32{0, 0}},
			{Opcode: OpLoadConst, Source: 0, Block: 0, Operands: []uint32{1, 0}},
			{Opcode: OpBinaryValue, Source: 0, Block: 0, Operands: []uint32{2, 0, 5, 0, 1}},
			{Opcode: OpTypeTest, Source: 0, Block: 0, Operands: []uint32{3, 2, 0}},
			{Opcode: OpBranch, Source: 0, Block: 0, Operands: []uint32{3, 1, 2}},
			{Opcode: OpReturnValue, Source: 0, Block: 1, Operands: []uint32{2}},
			{Opcode: OpInvoke, Source: 0, Block: 2, Operands: []uint32{2, 0, 0, 0}},
			{Opcode: OpReturnValue, Source: 0, Block: 2, Operands: []uint32{2}},
		},
	}
}
