//go:build js && wasm

package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// TypecheckConfig is retained on js/wasm so callers can compile against the
// interpreter API. Typechecking requires the native source loader and is not
// part of the pre-parsed-AST WASM runtime.
type TypecheckConfig struct {
	// Checker cannot have the native *typechecker.Checker type on js/wasm
	// because that package requires the native loader. It is retained only so
	// nil-config call sites can share their configuration shape.
	Checker  any
	FailFast bool
}

func (i *Interpreter) EnableTypechecker(cfg TypecheckConfig) {
	i.typecheckerEnabled = true
	i.typecheckerStrict = cfg.FailFast
	i.typecheckDiagnostics = nil
}

func (i *Interpreter) DisableTypechecker() {
	i.typecheckerEnabled = false
	i.typecheckerStrict = false
	i.typecheckDiagnostics = nil
}

func (i *Interpreter) TypecheckDiagnostics() []interpreterTypecheckDiagnostic {
	if len(i.typecheckDiagnostics) == 0 {
		return nil
	}
	out := make([]interpreterTypecheckDiagnostic, len(i.typecheckDiagnostics))
	copy(out, i.typecheckDiagnostics)
	return out
}

func (i *Interpreter) prepareModuleTypechecking(_ *ast.Module) (func(), error) {
	i.typecheckDiagnostics = nil
	if i.typecheckerEnabled {
		return nil, fmt.Errorf("typechecker is unavailable on js/wasm; provide a checked pre-parsed AST module")
	}
	return nil, nil
}
