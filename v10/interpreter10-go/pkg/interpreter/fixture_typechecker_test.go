package interpreter

import "testing"

func TestConfigureFixtureTypecheckerWarnMode(t *testing.T) {
	t.Setenv(fixtureTypecheckEnv, "warn")
	interp := New()
	mode := configureFixtureTypechecker(interp)
	if mode != typecheckModeWarn {
		t.Fatalf("expected warn mode, got %v", mode)
	}

	module := buildTypeMismatchModule()
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("expected evaluation to succeed in warn mode, got %v", err)
	}
	diags := interp.TypecheckDiagnostics()
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics in warn mode")
	}
}

func TestConfigureFixtureTypecheckerStrictMode(t *testing.T) {
	t.Setenv(fixtureTypecheckEnv, "strict")
	interp := New()
	mode := configureFixtureTypechecker(interp)
	if mode != typecheckModeStrict {
		t.Fatalf("expected strict mode, got %v", mode)
	}

	module := buildTypeMismatchModule()
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected evaluation to fail in strict mode")
	}
	diags := interp.TypecheckDiagnostics()
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics recorded in strict mode")
	}
}

func TestConfigureFixtureTypecheckerOffMode(t *testing.T) {
	t.Setenv(fixtureTypecheckEnv, "off")
	interp := New()
	mode := configureFixtureTypechecker(interp)
	if mode != typecheckModeOff {
		t.Fatalf("expected off mode, got %v", mode)
	}

	module := buildTypeMismatchModule()
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("expected evaluation to succeed when typechecker disabled, got %v", err)
	}
	if diags := interp.TypecheckDiagnostics(); len(diags) != 0 {
		t.Fatalf("expected no diagnostics when typechecker disabled, got %v", diags)
	}
}
