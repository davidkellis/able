package interpreter

import (
	"able/interpreter-go/pkg/typechecker"
)

// TypecheckConfig configures optional typechecker integration.
type TypecheckConfig struct {
	Checker  *typechecker.Checker
	FailFast bool
}

// EnableTypechecker wires a typechecker run before module evaluation.
func (i *Interpreter) EnableTypechecker(cfg TypecheckConfig) {
	if cfg.Checker != nil {
		i.typechecker = cfg.Checker
	} else if i.typechecker == nil {
		i.typechecker = typechecker.New()
	}
	i.typecheckerEnabled = true
	i.typecheckerStrict = cfg.FailFast
	i.typecheckDiagnostics = nil
}

// DisableTypechecker stops running the typechecker before evaluation.
func (i *Interpreter) DisableTypechecker() {
	i.typecheckerEnabled = false
	i.typecheckerStrict = false
	i.typecheckDiagnostics = nil
}

// TypecheckDiagnostics returns the diagnostics produced by the last typechecker run.
func (i *Interpreter) TypecheckDiagnostics() []typechecker.Diagnostic {
	if !i.typecheckerEnabled || len(i.typecheckDiagnostics) == 0 {
		return nil
	}
	out := make([]typechecker.Diagnostic, len(i.typecheckDiagnostics))
	copy(out, i.typecheckDiagnostics)
	return out
}
