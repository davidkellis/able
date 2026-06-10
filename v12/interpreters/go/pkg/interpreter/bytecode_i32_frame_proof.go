//go:build !(js && wasm)

package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

// bytecodeI32FrameProof is immutable lowering metadata for a future VM-v2
// frame. It is deliberately separate from the legacy i32 register sidecar:
// recording a proof must not change the current VM representation or opcode
// selection.
type bytecodeI32FrameProof struct {
	slots []bool
}

func (p *bytecodeI32FrameProof) hasSlot(slot int) bool {
	return p != nil && slot >= 0 && slot < len(p.slots) && p.slots[slot]
}

func bytecodeAttachI32FrameProof(layout *bytecodeFrameLayout, def *ast.FunctionDefinition, instructions []bytecodeInstruction, inferred bytecodeInferenceFacts) {
	if layout == nil || def == nil || len(inferred) == 0 || layout.slotCount <= 0 {
		return
	}

	proof := &bytecodeI32FrameProof{slots: make([]bool, layout.slotCount)}
	for slot, param := range def.Params {
		if slot >= len(layout.slotKinds) || layout.slotKinds[slot] != bytecodeCellKindI32 || param == nil {
			continue
		}
		if pattern, ok := param.Name.(ast.Node); ok && bytecodeInferenceIsConcreteI32(inferred[pattern]) {
			proof.slots[slot] = true
		}
	}

	seenWrite := make([]bool, layout.slotCount)
	validWrite := make([]bool, layout.slotCount)
	invalidWrite := make([]bool, layout.slotCount)
	for _, instr := range instructions {
		if !bytecodeInstructionWritesI32FrameSlot(instr) || instr.target < 0 || instr.target >= layout.slotCount {
			continue
		}
		slot := instr.target
		if slot >= len(layout.slotKinds) || layout.slotKinds[slot] != bytecodeCellKindI32 {
			continue
		}
		seenWrite[slot] = true
		assignment, ok := instr.node.(*ast.AssignmentExpression)
		if !ok || assignment == nil ||
			!bytecodeInferenceIsConcreteI32(inferred[assignment.Left]) ||
			!bytecodeInferenceIsConcreteI32(inferred[assignment.Right]) {
			invalidWrite[slot] = true
			continue
		}
		validWrite[slot] = true
	}
	for slot := range proof.slots {
		if seenWrite[slot] {
			proof.slots[slot] = validWrite[slot] && !invalidWrite[slot]
		} else if slot >= len(def.Params) {
			proof.slots[slot] = false
		}
	}

	layout.i32FrameProof = proof
}

func bytecodeInstructionWritesI32FrameSlot(instr bytecodeInstruction) bool {
	switch instr.op {
	case bytecodeOpStoreSlot,
		bytecodeOpStoreSlotNew,
		bytecodeOpStoreSlotI32,
		bytecodeOpCompoundAssignSlotI32:
		return true
	default:
		return false
	}
}

func bytecodeInferenceIsConcreteI32(typ typechecker.Type) bool {
	integer, ok := typ.(typechecker.IntegerType)
	return ok && integer.Suffix == "i32"
}
