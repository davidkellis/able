package interpreter

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/typechecker"
)

const fixtureTypecheckEnv = "ABLE_TYPECHECK_FIXTURES"

type fixtureTypecheckMode int

const (
	typecheckModeOff fixtureTypecheckMode = iota
	typecheckModeWarn
	typecheckModeStrict
)

func configureFixtureTypechecker(interp *Interpreter) fixtureTypecheckMode {
	modeVal, ok := os.LookupEnv(fixtureTypecheckEnv)
	if !ok {
		return typecheckModeOff
	}
	mode := strings.TrimSpace(strings.ToLower(modeVal))
	switch mode {
	case "", "0", "off", "false":
		return typecheckModeOff
	case "strict", "fail", "error", "1", "true":
		interp.EnableTypechecker(TypecheckConfig{FailFast: true})
		return typecheckModeStrict
	case "warn", "warning":
		interp.EnableTypechecker(TypecheckConfig{})
		return typecheckModeWarn
	default:
		interp.EnableTypechecker(TypecheckConfig{})
		return typecheckModeWarn
	}
}

func collectDiagnosticMessages(diags []typechecker.Diagnostic) []string {
	if len(diags) == 0 {
		return nil
	}
	msgs := make([]string, len(diags))
	for i, diag := range diags {
		msgs[i] = diag.Message
	}
	return msgs
}

func checkFixtureTypecheckDiagnostics(t *testing.T, mode fixtureTypecheckMode, expected []string, diags []typechecker.Diagnostic) {
	if mode == typecheckModeOff {
		return
	}
	t.Helper()
	actual := collectDiagnosticMessages(diags)
	for _, msg := range actual {
		t.Logf("typechecker: %s", msg)
	}

	if len(expected) == 0 {
		if len(actual) > 0 {
			t.Fatalf("typechecker produced diagnostics %v but fixture did not provide expectations", actual)
		}
		return
	}
	if len(actual) == 0 {
		t.Fatalf("fixture expected typechecker diagnostics %v but none were produced", expected)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("fixture expected typechecker diagnostics %v, got %v", expected, actual)
	}
}
