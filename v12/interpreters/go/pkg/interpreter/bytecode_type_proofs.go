//go:build !(js && wasm)

package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

// bytecodeInferenceFactsForCheckedProgram keeps only facts produced for
// diagnostic-free modules. A warning/error elsewhere must not authorize an
// unboxed frame in the affected module.
func bytecodeInferenceFactsForCheckedProgram(program *driver.Program, check ProgramCheckResult) bytecodeInferenceFacts {
	if program == nil || len(check.Inferred) == 0 {
		return nil
	}
	invalidPackages := make(map[string]struct{}, len(check.Diagnostics))
	for _, diagnostic := range check.Diagnostics {
		invalidPackages[diagnostic.Package] = struct{}{}
	}
	facts := make(bytecodeInferenceFacts)
	for _, module := range program.Modules {
		if module == nil || module.AST == nil {
			continue
		}
		if _, invalid := invalidPackages[module.Package]; invalid {
			continue
		}
		for node, typ := range check.Inferred[module.Package] {
			if node != nil && typ != nil {
				facts[node] = typ
			}
		}
	}
	if len(facts) == 0 {
		return nil
	}
	return facts
}

func bytecodeMethodSelectionsForCheckedProgram(program *driver.Program, check ProgramCheckResult) bytecodeMethodSelections {
	if program == nil || len(check.Methods) == 0 {
		return nil
	}
	invalidPackages := make(map[string]struct{}, len(check.Diagnostics))
	for _, diagnostic := range check.Diagnostics {
		invalidPackages[diagnostic.Package] = struct{}{}
	}
	selections := make(bytecodeMethodSelections)
	for _, module := range program.Modules {
		if module == nil || module.AST == nil {
			continue
		}
		if _, invalid := invalidPackages[module.Package]; invalid {
			continue
		}
		for node, selection := range check.Methods[module.Package] {
			if node != nil {
				selections[node] = selection
			}
		}
	}
	if len(selections) == 0 {
		return nil
	}
	return selections
}

func (i *Interpreter) pushBytecodeInferenceFacts(facts bytecodeInferenceFacts) func() {
	if i == nil {
		return func() {}
	}
	var copied bytecodeInferenceFacts
	if len(facts) > 0 {
		copied = facts.Clone()
	}
	i.bytecodeInferenceFactsMu.Lock()
	previous := i.bytecodeInferenceFacts
	i.bytecodeInferenceFacts = copied
	i.bytecodeInferenceFactsMu.Unlock()
	return func() {
		i.bytecodeInferenceFactsMu.Lock()
		i.bytecodeInferenceFacts = previous
		i.bytecodeInferenceFactsMu.Unlock()
	}
}

func (i *Interpreter) pushBytecodeMethodSelections(selections bytecodeMethodSelections) func() {
	if i == nil {
		return func() {}
	}
	var copied bytecodeMethodSelections
	if len(selections) > 0 {
		copied = selections.Clone()
	}
	i.bytecodeInferenceFactsMu.Lock()
	previous := i.bytecodeMethodSelections
	i.bytecodeMethodSelections = copied
	i.bytecodeInferenceFactsMu.Unlock()
	return func() {
		i.bytecodeInferenceFactsMu.Lock()
		i.bytecodeMethodSelections = previous
		i.bytecodeInferenceFactsMu.Unlock()
	}
}

func (i *Interpreter) bytecodeGenericUnionMethodCallProven(call *ast.FunctionCall) bool {
	if i == nil || call == nil {
		return false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil {
		return false
	}
	i.bytecodeInferenceFactsMu.RLock()
	selection, ok := i.bytecodeMethodSelections[member]
	i.bytecodeInferenceFactsMu.RUnlock()
	return ok && selection.GenericNamedUnion
}

func (i *Interpreter) bytecodeInferenceFactsSnapshot() bytecodeInferenceFacts {
	if i == nil {
		return nil
	}
	i.bytecodeInferenceFactsMu.RLock()
	facts := i.bytecodeInferenceFacts
	i.bytecodeInferenceFactsMu.RUnlock()
	if len(facts) == 0 {
		return nil
	}
	return facts
}

// installRuntimeInferenceFacts retains the checked facts for the loaded
// program. Bytecode lowering consumes a scoped snapshot, while method calls
// may execute after EvaluateProgram has returned and therefore need the same
// facts for static generic-union receiver dispatch.
func (i *Interpreter) installRuntimeInferenceFacts(facts bytecodeInferenceFacts) {
	if i == nil {
		return
	}
	var copied bytecodeInferenceFacts
	if len(facts) > 0 {
		copied = facts.Clone()
	}
	i.bytecodeInferenceFactsMu.Lock()
	i.runtimeInferenceFacts = copied
	i.bytecodeInferenceFactsMu.Unlock()
}

func (i *Interpreter) runtimeInferenceFactsSnapshot() bytecodeInferenceFacts {
	if i == nil {
		return nil
	}
	i.bytecodeInferenceFactsMu.RLock()
	facts := i.runtimeInferenceFacts
	i.bytecodeInferenceFactsMu.RUnlock()
	if len(facts) == 0 {
		return nil
	}
	return facts
}

// RegisterStaticCallReceiverType supplies a checked receiver type for a call
// node synthesized by compiled Go output. Source-interpreted calls retain the
// original typechecker facts instead. The mapping is installed during compiled
// package registration and then read concurrently by calls.
func (i *Interpreter) RegisterStaticCallReceiverType(call *ast.FunctionCall, receiverType ast.TypeExpression) {
	if i == nil || call == nil || receiverType == nil {
		return
	}
	i.bytecodeInferenceFactsMu.Lock()
	if i.runtimeStaticCallReceiverTypes == nil {
		i.runtimeStaticCallReceiverTypes = make(map[*ast.FunctionCall]ast.TypeExpression)
	}
	i.runtimeStaticCallReceiverTypes[call] = receiverType
	i.bytecodeInferenceFactsMu.Unlock()
}

func (i *Interpreter) registeredStaticCallReceiverType(call *ast.FunctionCall) ast.TypeExpression {
	if i == nil || call == nil {
		return nil
	}
	i.bytecodeInferenceFactsMu.RLock()
	receiverType := i.runtimeStaticCallReceiverTypes[call]
	i.bytecodeInferenceFactsMu.RUnlock()
	return receiverType
}
