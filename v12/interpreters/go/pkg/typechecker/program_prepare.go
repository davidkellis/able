//go:build !(js && wasm)

package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/driver"
)

// PrepareProgramForEvaluation applies the declaration-side AST normalization
// that runtime evaluation depends on, without running the full module checker.
// This keeps SkipTypecheck execution semantically aligned with a fully checked
// program while avoiding the full diagnostic pipeline.
func PrepareProgramForEvaluation(program *driver.Program) error {
	if program == nil {
		return fmt.Errorf("typechecker: program is nil")
	}
	pc := NewProgramChecker()
	for _, mod := range program.Modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		env, impls, methods, _ := pc.buildPrelude(mod.AST.Imports, mod.Package)
		checker := New()
		checker.SetPrelude(env, impls, methods)
		checker.SetNodeOrigins(mod.NodeOrigins)
		checker.collectDeclarations(mod.AST)
		pc.captureExports(mod, checker)
	}
	return nil
}
