package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDiagnosticRegistrationReservesNodeOrigins(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> i32 {",
			"  1 + 2",
			"}",
			"",
		}, "\n"),
	})
	if !strings.Contains(compiledSrc, "bridge.ReserveNodeOrigins(__able_runtime,") {
		t.Fatal("expected generated diagnostic registration to pre-size node origins")
	}
}
