//go:build js && wasm

package interpreter

import "able/interpreter-go/pkg/ast"

func (i *Interpreter) pushBytecodeInferenceFacts(facts bytecodeInferenceFacts) func() {
	if i == nil {
		return func() {}
	}
	i.bytecodeInferenceFactsMu.Lock()
	previous := i.bytecodeInferenceFacts
	i.bytecodeInferenceFacts = facts
	i.bytecodeInferenceFactsMu.Unlock()
	return func() {
		i.bytecodeInferenceFactsMu.Lock()
		i.bytecodeInferenceFacts = previous
		i.bytecodeInferenceFactsMu.Unlock()
	}
}

func (i *Interpreter) pushBytecodeMethodSelections(_ bytecodeMethodSelections) func() {
	return func() {}
}

func (i *Interpreter) bytecodeGenericUnionMethodCallProven(_ *ast.FunctionCall) bool {
	return false
}

func (i *Interpreter) bytecodeInferenceFactsSnapshot() bytecodeInferenceFacts {
	return nil
}

func (i *Interpreter) installRuntimeInferenceFacts(_ bytecodeInferenceFacts) {}

func (i *Interpreter) runtimeInferenceFactsSnapshot() bytecodeInferenceFacts { return nil }

func (i *Interpreter) RegisterStaticCallReceiverType(_ *ast.FunctionCall, _ ast.TypeExpression) {}
