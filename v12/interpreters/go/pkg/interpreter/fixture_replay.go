//go:build !(js && wasm)

package interpreter

import (
	"fmt"
	"path/filepath"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

// FixtureReplayResult captures the observable result of replaying a fixture.
type FixtureReplayResult struct {
	Value         runtime.Value
	Stdout        []string
	Diagnostics   []string
	TypecheckMode string
	RuntimeError  string
}

// ReplayFixture evaluates a fixture directory using the shared fixture runtime
// wiring and returns the observable output needed by tooling wrappers.
func ReplayFixture(dir, entry string, setup []string, executor Executor) (FixtureReplayResult, error) {
	var result FixtureReplayResult

	modulePath := filepath.Join(dir, entry)
	entryModuleAST, entryOrigin, err := LoadFixtureModule(modulePath)
	if err != nil {
		return result, fmt.Errorf("load module %s: %w", modulePath, err)
	}

	setupModules := make([]*driver.Module, 0, len(setup))
	for _, setupFile := range setup {
		setupPath := filepath.Join(dir, setupFile)
		setupModuleAST, setupOrigin, err := LoadFixtureModule(setupPath)
		if err != nil {
			return result, fmt.Errorf("load setup %s: %w", setupPath, err)
		}
		setupModules = append(setupModules, fixtureDriverModule(setupModuleAST, setupOrigin))
	}

	entryModule := fixtureDriverModule(entryModuleAST, entryOrigin)
	program, err := buildFixtureProgram(setupModules, entryModule)
	if err != nil {
		return result, err
	}

	interp := NewWithExecutor(executor)
	mode := configureFixtureTypechecker(interp)
	result.TypecheckMode = fixtureTypecheckModeString(mode)

	var stdout []string
	registerPrint(interp, &stdout)

	value, _, check, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    mode == typecheckModeOff,
		AllowDiagnostics: mode != typecheckModeOff,
	})
	result.Diagnostics = formatModuleDiagnostics(check.Diagnostics)
	result.Stdout = stdout
	if err != nil {
		result.RuntimeError = DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err))
		return result, nil
	}
	result.Value = value
	return result, nil
}

func fixtureTypecheckModeString(mode fixtureTypecheckMode) string {
	switch mode {
	case typecheckModeWarn:
		return "warn"
	case typecheckModeStrict:
		return "strict"
	default:
		return "off"
	}
}
