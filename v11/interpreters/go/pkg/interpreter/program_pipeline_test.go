package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterPipelineFromSource(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "package.yml"), "name: sample\n")
	writeTestFile(t, filepath.Join(root, "shared.able"), `
package shared

fn welcome(name: string) -> string {
  `+"`welcome ${name}`"+`
}
`)
	writeTestFile(t, filepath.Join(root, "main.able"), `
package main

import sample.shared.{welcome}

fn main() -> void {
  print(welcome("Able"))
}
`)

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader error: %v", err)
	}
	defer loader.Close()

	entryPath := filepath.Join(root, "main.able")
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	check, err := TypecheckProgram(program)
	if err != nil {
		t.Fatalf("TypecheckProgram error: %v", err)
	}
	if len(check.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", check.Diagnostics)
	}

	interp := New()
	var logs []string
	registerTestPrint(interp, &logs)

	_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{SkipTypecheck: true})
	if err != nil {
		t.Fatalf("EvaluateProgram error: %v", err)
	}
	if entryEnv == nil {
		t.Fatalf("expected non-nil entry environment")
	}

	mainValue, err := entryEnv.Get("main")
	if err != nil {
		t.Fatalf("entry package missing main: %v", err)
	}
	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		t.Fatalf("call main: %v", err)
	}
	if len(logs) != 1 || logs[0] != "welcome Able" {
		t.Fatalf("unexpected stdout: %v", logs)
	}
}

func TestInterpreterPipelineTypecheckDiagnostics(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "package.yml"), "name: diag_demo\n")
	writeTestFile(t, filepath.Join(root, "main.able"), `
package main

fn main() -> i32 {
  "not an integer"
}
`)

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader error: %v", err)
	}
	defer loader.Close()

	entryPath := filepath.Join(root, "main.able")
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	check, err := TypecheckProgram(program)
	if err != nil {
		t.Fatalf("TypecheckProgram error: %v", err)
	}
	if len(check.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics for invalid program")
	}
	if want := "returns String, expected i32"; !strings.Contains(check.Diagnostics[0].Diagnostic.Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, check.Diagnostics[0].Diagnostic.Message)
	}

	interp := New()
	value, env, evalCheck, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("EvaluateProgram error: %v", err)
	}
	if len(evalCheck.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics from EvaluateProgram")
	}
	if value != nil || env != nil {
		t.Fatalf("expected evaluation to halt when diagnostics present: value=%v env=%v", value, env)
	}

	value, env, evalCheck, err = interp.EvaluateProgram(program, ProgramEvaluationOptions{
		AllowDiagnostics: true,
	})
	if err != nil {
		t.Fatalf("EvaluateProgram with AllowDiagnostics error: %v", err)
	}
	if len(evalCheck.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics when AllowDiagnostics=true")
	}
	if env == nil {
		t.Fatalf("expected entry environment when proceeding with diagnostics")
	}
	mainValue, err := env.Get("main")
	if err != nil {
		t.Fatalf("entry package missing main: %v", err)
	}
	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		t.Fatalf("call main after AllowDiagnostics: %v", err)
	}
	if value == nil {
		t.Fatalf("expected EvaluateProgram to return entry value when diagnostics allowed")
	}
}

func TestInterpreterPipelineLoopExpressionStatement(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "package.yml"), "name: loop_demo\n")
	writeTestFile(t, filepath.Join(root, "main.able"), `
package main

fn main() {
  counter := 3
  loop {
    if counter < 0 {
      break
    }
    counter = counter - 1
  }
}
`)

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader error: %v", err)
	}
	defer loader.Close()

	entryPath := filepath.Join(root, "main.able")
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	check, err := TypecheckProgram(program)
	if err != nil {
		t.Fatalf("TypecheckProgram error: %v", err)
	}
	if len(check.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", check.Diagnostics)
	}

	interp := New()
	if _, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{SkipTypecheck: true}); err != nil {
		t.Fatalf("EvaluateProgram error: %v", err)
	}
}

func registerTestPrint(interp *Interpreter, logs *[]string) {
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			var parts []string
			for _, arg := range args {
				parts = append(parts, formatTestRuntimeValue(arg))
			}
			*logs = append(*logs, strings.Join(parts, " "))
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}

func formatTestRuntimeValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.CharValue:
		return string(v.Val)
	case runtime.NilValue:
		return "nil"
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
