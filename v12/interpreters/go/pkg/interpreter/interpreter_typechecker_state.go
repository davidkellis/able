//go:build !(js && wasm)

package interpreter

import "able/interpreter-go/pkg/typechecker"

// These aliases keep the AST evaluator independent of the optional
// typechecker package at its WASM boundary while preserving the native API.
type interpreterTypechecker = *typechecker.Checker
type interpreterTypecheckDiagnostic = typechecker.Diagnostic
type bytecodeInferenceFacts = typechecker.InferenceMap
type bytecodeMethodSelections = typechecker.MethodSelectionMap
