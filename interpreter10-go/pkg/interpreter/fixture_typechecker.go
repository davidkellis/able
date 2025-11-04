package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func formatModuleDiagnostics(diags []ModuleDiagnostic) []string {
	if len(diags) == 0 {
		return nil
	}
	msgs := make([]string, len(diags))
	for i, diag := range diags {
		msgs[i] = formatModuleDiagnostic(diag)
	}
	return msgs
}

func formatModuleDiagnostic(diag ModuleDiagnostic) string {
	location := formatSourceHint(diag.Source)
	if location != "" {
		return fmt.Sprintf("typechecker: %s %s", location, diag.Diagnostic.Message)
	}
	return fmt.Sprintf("typechecker: %s", diag.Diagnostic.Message)
}

func formatSourceHint(hint typechecker.SourceHint) string {
	path := normalizeSourcePath(strings.TrimSpace(hint.Path))
	line := hint.Line
	column := hint.Column
	switch {
	case path != "" && line > 0 && column > 0:
		return fmt.Sprintf("%s:%d:%d", path, line, column)
	case path != "" && line > 0:
		return fmt.Sprintf("%s:%d", path, line)
	case path != "":
		return path
	case line > 0 && column > 0:
		return fmt.Sprintf("line %d, column %d", line, column)
	case line > 0:
		return fmt.Sprintf("line %d", line)
	default:
		return ""
	}
}

var (
	repoRootOnce sync.Once
	repoRootPath string
	repoRootErr  error
)

func normalizeSourcePath(raw string) string {
	if raw == "" {
		return ""
	}
	path := raw
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	root := repositoryRoot()
	if root != "" {
		if rel, err := filepath.Rel(root, path); err == nil {
			path = rel
		}
	}
	return filepath.ToSlash(path)
}

func repositoryRoot() string {
	repoRootOnce.Do(func() {
		root, err := filepath.Abs(filepath.Join("..", "..", ".."))
		if err != nil {
			repoRootErr = err
			return
		}
		repoRootPath = root
	})
	if repoRootErr != nil {
		return ""
	}
	return repoRootPath
}

func checkFixtureTypecheckDiagnostics(t *testing.T, mode fixtureTypecheckMode, expected []string, diags []ModuleDiagnostic, skipTS bool) []string {
	if mode == typecheckModeOff {
		return nil
	}
	t.Helper()
	actual := formatModuleDiagnostics(diags)
	for _, msg := range actual {
		t.Log(msg)
	}

	if skipTS {
		return actual
	}

	if len(expected) == 0 {
		return actual
	}
	if len(actual) == 0 {
		t.Fatalf("fixture expected typechecker diagnostics %v but none were produced", expected)
	}
	expectedKeys := diagnosticKeys(expected)
	actualKeys := diagnosticKeys(actual)
	if len(expectedKeys) != len(actualKeys) {
		t.Fatalf("fixture expected typechecker diagnostics %v, got %v", expected, actual)
	}
	for i := range expectedKeys {
		if expectedKeys[i] != actualKeys[i] {
			t.Fatalf("fixture expected typechecker diagnostics %v, got %v", expected, actual)
		}
	}
	return actual
}
