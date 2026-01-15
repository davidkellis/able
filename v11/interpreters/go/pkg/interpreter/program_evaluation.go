package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

// ProgramEvaluationOptions configures EvaluateProgram behaviour.
type ProgramEvaluationOptions struct {
	// SkipTypecheck bypasses the program-wide typechecker. Use this when diagnostics
	// have already been collected via TypecheckProgram.
	SkipTypecheck bool
	// AllowDiagnostics permits evaluation to proceed even when the typechecker
	// reports diagnostics. Diagnostics are still returned to the caller.
	AllowDiagnostics bool
}

// EvaluateProgram executes the modules in the provided program according to their
// dependency order. If typechecking is enabled (default), diagnostics are returned
// and evaluation is skipped when issues are detected.
func (i *Interpreter) EvaluateProgram(program *driver.Program, opts ProgramEvaluationOptions) (runtime.Value, *runtime.Environment, ProgramCheckResult, error) {
	if program == nil {
		return nil, nil, ProgramCheckResult{}, fmt.Errorf("interpreter: program is nil")
	}
	if program.Entry == nil || program.Entry.AST == nil {
		return nil, nil, ProgramCheckResult{}, fmt.Errorf("interpreter: program missing entry module")
	}
	i.SetNodeOrigins(mergeNodeOrigins(program.Modules))

	var check ProgramCheckResult
	if !opts.SkipTypecheck {
		var err error
		check, err = TypecheckProgram(program)
		if err != nil {
			return nil, nil, ProgramCheckResult{}, err
		}
		if len(check.Diagnostics) > 0 && !opts.AllowDiagnostics {
			return nil, nil, check, nil
		}
	}

	prevEnabled := i.typecheckerEnabled
	prevStrict := i.typecheckerStrict
	prevChecker := i.typechecker
	if prevEnabled {
		i.DisableTypechecker()
		defer i.EnableTypechecker(TypecheckConfig{Checker: prevChecker, FailFast: prevStrict})
	}

	var entryEnv *runtime.Environment
	var entryValue runtime.Value = runtime.NilValue{}
	for _, mod := range program.Modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		val, env, err := i.EvaluateModule(mod.AST)
		if err != nil {
			source := "<unknown>"
			if len(mod.Files) > 0 {
				source = mod.Files[0]
			}
			return nil, nil, check, fmt.Errorf("interpreter: evaluation error in package %s (e.g., %s): %w", mod.Package, source, err)
		}
		if mod.Package == program.Entry.Package {
			entryEnv = env
			entryValue = val
		}
	}

	if entryEnv == nil {
		entryEnv = i.GlobalEnvironment()
	}
	return entryValue, entryEnv, check, nil
}

func mergeNodeOrigins(modules []*driver.Module) map[ast.Node]string {
	if len(modules) == 0 {
		return nil
	}
	origins := make(map[ast.Node]string)
	for _, mod := range modules {
		if mod == nil || len(mod.NodeOrigins) == 0 {
			continue
		}
		for node, origin := range mod.NodeOrigins {
			if node == nil {
				continue
			}
			origins[node] = origin
		}
	}
	if len(origins) == 0 {
		return nil
	}
	return origins
}
