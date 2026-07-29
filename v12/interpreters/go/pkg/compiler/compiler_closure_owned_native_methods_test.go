package compiler

import (
	"strings"
	"testing"
)

func TestCompilerClosureOwnedKernelMethodsFollowSchedulerContextActivation(t *testing.T) {
	awaitSource := strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  handle := spawn { 42 }",
		"  await [handle] as i32",
		"}",
		"",
	}, "\n")
	defaultResult := compileNoFallbackSource(t, awaitSource)
	experimentalResult := compileNoFallbackSourceWithCompilerOptions(
		t,
		awaitSource,
		Options{
			PackageName:                  "main",
			ExperimentalExecutionContext: true,
		},
	)

	defaultSource := string(defaultResult.Files["compiled.go"])
	experimentalSource := string(experimentalResult.Files["compiled.go"])
	methods := []struct {
		field string
		local string
	}{
		{field: "is_ready", local: "isReady"},
		{field: "register", local: "register"},
		{field: "commit", local: "commit"},
		{field: "is_default", local: "isDefault"},
		{field: "wake", local: "wakeFn"},
		{field: "cancel", local: "cancelMethod"},
	}
	for _, method := range methods {
		bound := "inst.Fields[" + quoteGoString(method.field) +
			"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: " +
			method.local + "}"
		direct := "inst.Fields[" + quoteGoString(method.field) +
			"] = " + method.local
		for mode, compiledSource := range map[string]string{
			"default":  defaultSource,
			"force_on": experimentalSource,
		} {
			if strings.Contains(compiledSource, bound) {
				t.Fatalf("%s source retains bound %s method", mode, method.field)
			}
			if !strings.Contains(compiledSource, direct) {
				t.Fatalf("%s source is missing receiver-free %s method", mode, method.field)
			}
		}
	}
}

func TestCompilerClosureOwnedKernelMethodCaptureParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping closure-owned kernel callable execution test in short mode")
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.concurrency.{Await}",
		"",
		"fn main() -> void {",
		"  handle := spawn {",
		"    arm := Await.default({ => 42_i32 })",
		"    ready := arm.is_ready",
		"    commit := arm.commit",
		"    print(ready())",
		"    print(commit())",
		"  }",
		"  future_flush()",
		"  handle.value()",
		"}",
		"",
	}, "\n")
	for _, testCase := range []struct {
		name         string
		experimental bool
	}{
		{name: "default"},
		{name: "execution_context", experimental: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got := compileAndRunExecSourceWithOptions(
				t,
				"ablec-closure-owned-native-capture",
				source,
				Options{
					PackageName:                  "main",
					EmitMain:                     true,
					ExperimentalExecutionContext: testCase.experimental,
				},
			)
			if got != "true\n42\n" {
				t.Fatalf("captured kernel callable output = %q", got)
			}
		})
	}
}

func quoteGoString(value string) string {
	return `"` + value + `"`
}
