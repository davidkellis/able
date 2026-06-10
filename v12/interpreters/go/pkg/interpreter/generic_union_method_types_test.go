package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm(t *testing.T) {
	root := repositoryRoot()
	entryPath := filepath.Join(root, "v12", "examples", "benchmarks", "option_result_config", "option_result_config.able")
	stdlibRoot := filepath.Join(filepath.Dir(root), "able-stdlib")
	if _, err := os.Stat(entryPath); err != nil {
		t.Fatalf("option/result configuration benchmark source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stdlibRoot, "src", "package.yml")); err != nil {
		t.Fatalf("canonical stdlib source: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", stdlibRoot)
	t.Setenv("ABLE_SOURCE_ROOT_ONLY", "1")

	searchPaths, err := buildExecSearchPaths(entryPath, filepath.Dir(entryPath), fixtureManifest{})
	if err != nil {
		t.Fatalf("build benchmark search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("new loader: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load benchmark program: %v", err)
	}

	for _, mode := range []struct {
		name string
		new  func() *Interpreter
	}{
		{name: "treewalker", new: New},
		{name: "bytecode", new: NewBytecode},
	} {
		t.Run(mode.name, func(t *testing.T) {
			interp := mode.new()
			var stdout []string
			registerTestPrint(interp, &stdout)
			_, entryEnv, check, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
			if err != nil {
				t.Fatalf("evaluate benchmark program: %v", err)
			}
			if len(check.Diagnostics) != 0 {
				t.Fatalf("unexpected benchmark diagnostics: %v", check.Diagnostics)
			}
			if entryEnv == nil {
				t.Fatal("expected benchmark entry environment")
			}
			mainValue, err := entryEnv.Get("main")
			if err != nil {
				t.Fatalf("lookup benchmark main: %v", err)
			}
			for run := 0; run < 2; run++ {
				if _, err := interp.CallFunction(mainValue, nil); err != nil {
					t.Fatalf("warm benchmark main call %d: %v", run+1, err)
				}
			}
			if len(stdout) != 2 || stdout[0] != "1024:18221610432" || stdout[1] != "1024:18221610432" {
				t.Fatalf("unexpected benchmark stdout: %v", stdout)
			}
		})
	}
}
