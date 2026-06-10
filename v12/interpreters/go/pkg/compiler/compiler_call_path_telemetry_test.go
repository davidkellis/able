package compiler

import (
	"strings"
	"testing"
)

func TestCompilerCallPathTelemetryIsOptIn(t *testing.T) {
	source := `
fn main() -> void {
  print("telemetry")
}
`

	baseline := compileNoFallbackExecSourceWithOptions(t, "ablec-call-path-telemetry-baseline", source, Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   "main.able",
	})
	baselineSource := combinedGeneratedSource(baseline)
	for _, forbidden := range []string{
		"__able_call_path_telemetry",
		"__able_telemetry_fast_method_call",
		callPathTelemetryEnv,
	} {
		if strings.Contains(baselineSource, forbidden) {
			t.Fatalf("normal generated output must not contain %q", forbidden)
		}
	}

	telemetry := compileNoFallbackExecSourceWithOptions(t, "ablec-call-path-telemetry-enabled", source, Options{
		PackageName:           "main",
		EmitMain:              true,
		EntryPath:             "main.able",
		EmitCallPathTelemetry: true,
	})
	telemetrySource := combinedGeneratedSource(telemetry)
	for _, expected := range []string{
		"__able_call_path_telemetry_snapshot",
		"__able_telemetry_fast_method_call",
		"__able_telemetry_generic_union_method_call",
		"__able_telemetry_generic_union_fallback",
		callPathTelemetryEnv,
	} {
		if !strings.Contains(telemetrySource, expected) {
			t.Fatalf("telemetry output missing %q", expected)
		}
	}
}
