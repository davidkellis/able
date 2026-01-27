package interpreter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

type execFixtureParityRun struct {
	Stdout               []string
	Stderr               []string
	ExitCode             int
	ExitSignaled         bool
	TypecheckDiagnostics []string
	TypecheckMode        fixtureTypecheckMode
	TypecheckOnly        bool
}

func TestExecFixtureParity(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	dirs := collectExecFixtures(t, root)
	for _, dir := range dirs {
		dir := dir
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			t.Fatalf("relative path for %s: %v", dir, err)
		}
		t.Run(filepath.ToSlash(rel), func(t *testing.T) {
			tree := runExecFixtureParity(t, dir, testExecTreewalker)
			bytecode := runExecFixtureParity(t, dir, testExecBytecode)
			compareExecFixtureParity(t, tree, bytecode)
		})
	}
}

func runExecFixtureParity(t *testing.T, dir string, execMode testExecMode) execFixtureParityRun {
	t.Helper()
	manifest := readManifest(t, dir)
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)

	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("exec search paths: %v", err)
	}

	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	outcome := execFixtureParityRun{}
	expectedTypecheck := manifest.Expect.TypecheckDiagnostics
	if expectedTypecheck != nil {
		check, err := TypecheckProgram(program)
		if err != nil {
			t.Fatalf("typecheck program: %v", err)
		}
		outcome.TypecheckDiagnostics = formatModuleDiagnostics(check.Diagnostics)
		if len(expectedTypecheck) > 0 {
			outcome.TypecheckOnly = true
			return outcome
		}
	}

	executor := selectFixtureExecutor(t, manifest.Executor)
	interp := newTestInterpreter(t, execMode, executor)
	outcome.TypecheckMode = configureFixtureTypechecker(interp)
	var stdout []string
	registerPrint(interp, &stdout)

	exitCode := 0
	exitSignaled := false
	var runtimeErr error

	entryEnv := interp.GlobalEnvironment()
	_, entryEnv, check, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    outcome.TypecheckMode == typecheckModeOff,
		AllowDiagnostics: outcome.TypecheckMode != typecheckModeOff,
	})
	outcome.TypecheckDiagnostics = formatModuleDiagnostics(check.Diagnostics)
	if err != nil {
		if code, ok := ExitCodeFromError(err); ok {
			exitCode = code
			exitSignaled = true
		} else {
			runtimeErr = err
			exitCode = 1
		}
	}

	var mainValue runtime.Value
	if runtimeErr == nil {
		env := entryEnv
		if env == nil {
			env = interp.GlobalEnvironment()
		}
		val, err := env.Get("main")
		if err != nil {
			runtimeErr = err
			exitCode = 1
		} else {
			mainValue = val
		}
	}

	if runtimeErr == nil {
		if _, err := interp.CallFunction(mainValue, nil); err != nil {
			if code, ok := ExitCodeFromError(err); ok {
				exitCode = code
				exitSignaled = true
			} else {
				runtimeErr = err
				exitCode = 1
			}
		}
	}

	outcome.Stdout = stdout
	outcome.ExitCode = exitCode
	outcome.ExitSignaled = exitSignaled
	if runtimeErr != nil {
		outcome.Stderr = []string{DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(runtimeErr))}
	}
	return outcome
}

func compareExecFixtureParity(t *testing.T, tree execFixtureParityRun, bytecode execFixtureParityRun) {
	t.Helper()
	if tree.TypecheckOnly != bytecode.TypecheckOnly {
		t.Fatalf("typecheck-only mismatch: treewalker=%v bytecode=%v", tree.TypecheckOnly, bytecode.TypecheckOnly)
	}
	if !reflect.DeepEqual(diagnosticKeys(tree.TypecheckDiagnostics), diagnosticKeys(bytecode.TypecheckDiagnostics)) {
		t.Fatalf("typecheck diagnostics mismatch: treewalker=%v bytecode=%v", tree.TypecheckDiagnostics, bytecode.TypecheckDiagnostics)
	}
	if tree.TypecheckOnly {
		return
	}
	if tree.TypecheckMode != bytecode.TypecheckMode {
		t.Fatalf("typecheck mode mismatch: treewalker=%v bytecode=%v", tree.TypecheckMode, bytecode.TypecheckMode)
	}
	if !reflect.DeepEqual(tree.Stdout, bytecode.Stdout) {
		t.Fatalf("stdout mismatch: treewalker=%v bytecode=%v", tree.Stdout, bytecode.Stdout)
	}
	if !reflect.DeepEqual(tree.Stderr, bytecode.Stderr) {
		t.Fatalf("stderr mismatch: treewalker=%v bytecode=%v", tree.Stderr, bytecode.Stderr)
	}
	if tree.ExitSignaled != bytecode.ExitSignaled || tree.ExitCode != bytecode.ExitCode {
		t.Fatalf("exit mismatch: treewalker=(%v,%d) bytecode=(%v,%d)", tree.ExitSignaled, tree.ExitCode, bytecode.ExitSignaled, bytecode.ExitCode)
	}
}
