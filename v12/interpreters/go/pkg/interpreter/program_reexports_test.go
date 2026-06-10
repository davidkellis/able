package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterSourceReexportsPreserveImportedFunctionIdentity(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wrapper string
		new     func() *Interpreter
	}{
		{name: "named treewalker", wrapper: "import sample.source.{value};\nexport value;", new: New},
		{name: "wildcard treewalker", wrapper: "export * from sample.source;", new: New},
		{name: "named bytecode", wrapper: "import sample.source.{value};\nexport value;", new: NewBytecode},
		{name: "wildcard bytecode", wrapper: "export * from sample.source;", new: NewBytecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeTestFile(t, filepath.Join(root, "package.yml"), "name: sample\n")
			writeTestFile(t, filepath.Join(root, "source.able"), "package source\nfn value() -> i32 { 42 }\n")
			writeTestFile(t, filepath.Join(root, "wrapper.able"), "package wrapper\n"+tc.wrapper+"\n")
			writeTestFile(t, filepath.Join(root, "main.able"), "package main\nimport sample.wrapper.{value};\nfn main() -> i32 { value() }\n")

			loader, err := driver.NewLoader(nil)
			if err != nil {
				t.Fatalf("NewLoader error: %v", err)
			}
			defer loader.Close()
			program, err := loader.Load(filepath.Join(root, "main.able"))
			if err != nil {
				t.Fatalf("Load error: %v", err)
			}

			interp := tc.new()
			_, entryEnv, check, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
			if err != nil {
				t.Fatalf("EvaluateProgram error: %v", err)
			}
			if len(check.Diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %v", check.Diagnostics)
			}
			mainValue, err := entryEnv.Get("main")
			if err != nil {
				t.Fatalf("entry main binding: %v", err)
			}
			got, err := interp.CallFunction(mainValue, nil)
			if err != nil {
				t.Fatalf("main call: %v", err)
			}
			integer, ok := got.(runtime.IntegerValue)
			if !ok || integer.BigInt().Int64() != 42 {
				t.Fatalf("re-exported call result = %T(%v), want i32 42", got, got)
			}
		})
	}
}
