//go:build js && wasm

package interpreter

import "able/interpreter-go/pkg/ast"

// bytecodeI32FrameProof is native typechecker-derived metadata. The WASM
// runtime has no native typechecker, so lowering retains the frame layout but
// omits that optional proof.
type bytecodeI32FrameProof struct {
	slots []bool
}

func (p *bytecodeI32FrameProof) hasSlot(slot int) bool {
	return p != nil && slot >= 0 && slot < len(p.slots) && p.slots[slot]
}

func bytecodeAttachI32FrameProof(_ *bytecodeFrameLayout, _ *ast.FunctionDefinition, _ []bytecodeInstruction, _ bytecodeInferenceFacts) {
}
