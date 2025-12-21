package interpreter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/runtime"
)

func TestExecFixtures(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v11", "fixtures", "exec")
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
			runExecFixture(t, dir)
		})
	}
}

func collectExecFixtures(t *testing.T, root string) []string {
	t.Helper()
	if root == "" {
		return nil
	}
	var dirs []string
	var walk func(string)
	walk = func(current string) {
		entries, err := os.ReadDir(current)
		if err != nil {
			return
		}
		hasManifest := false
		for _, entry := range entries {
			if entry.Type().IsRegular() && entry.Name() == "manifest.json" {
				hasManifest = true
				break
			}
		}
		if hasManifest {
			dirs = append(dirs, current)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				walk(filepath.Join(current, entry.Name()))
			}
		}
	}
	walk(root)
	return dirs
}

func runExecFixture(t *testing.T, dir string) {
	t.Helper()

	manifest := readManifest(t, dir)
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)

	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: stdlibRoot, Kind: driver.RootStdlib},
		{Path: kernelRoot, Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	interp := New()
	var stdout []string
	registerPrint(interp, &stdout)

	exitCode := 0
	var runtimeErr error

	entryEnv := interp.GlobalEnvironment()
	_, entryEnv, _, err = interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    true,
		AllowDiagnostics: true,
	})
	if err != nil {
		runtimeErr = err
		exitCode = 1
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
			runtimeErr = err
			exitCode = 1
		}
	}

	expected := manifest.Expect

	if runtimeErr != nil {
		if expected.Exit == nil || exitCode != *expected.Exit {
			t.Fatalf("runtime error: %v", runtimeErr)
		}
	}

	if expected.Stdout != nil {
		if !reflect.DeepEqual(stdout, expected.Stdout) {
			t.Fatalf("stdout mismatch: expected %v, got %v", expected.Stdout, stdout)
		}
	}

	if expected.Stderr != nil {
		actualErrs := []string{}
		if runtimeErr != nil {
			actualErrs = append(actualErrs, extractErrorMessage(runtimeErr))
		}
		if !reflect.DeepEqual(actualErrs, expected.Stderr) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expected.Stderr, actualErrs)
		}
	}

	if expected.Exit != nil {
		if exitCode != *expected.Exit {
			t.Fatalf("exit code mismatch: expected %d, got %d", *expected.Exit, exitCode)
		}
	} else if runtimeErr != nil {
		t.Fatalf("runtime error: %v", runtimeErr)
	}
}
