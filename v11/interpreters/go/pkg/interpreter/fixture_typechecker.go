package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	if label := inferDiagnosticPackage(diag); label != "" {
		return fmt.Sprintf("typechecker: %s %s", label, diag.Diagnostic.Message)
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
	anchors := []string{}
	if root != "" {
		anchors = append(anchors, filepath.Join(root, "v11", "interpreters", "ts", "scripts"))
		anchors = append(anchors, root)
	}
	for _, anchor := range anchors {
		if anchor == "" {
			continue
		}
		if rel, err := filepath.Rel(anchor, path); err == nil {
			path = rel
			break
		}
	}
	return filepath.ToSlash(path)
}

func repositoryRoot() string {
	repoRootOnce.Do(func() {
		start := ""
		if _, file, _, ok := runtime.Caller(0); ok {
			start = filepath.Dir(file)
		} else if wd, err := os.Getwd(); err == nil {
			start = wd
		}
		dir := start
		for i := 0; i < 10 && dir != "" && dir != string(filepath.Separator); i++ {
			if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
				repoRootPath = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if repoRootPath == "" {
			repoRootErr = fmt.Errorf("repository root not found from %s", start)
		}
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

func inferDiagnosticPackage(diag ModuleDiagnostic) string {
	if diag.Package != "" {
		return diag.Package
	}
	candidates := []string{diag.Source.Path}
	candidates = append(candidates, diag.Files...)
	for _, path := range candidates {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		normalized := normalizeSourcePath(path)
		dir := filepath.Dir(normalized)
		if dir == "." || dir == "/" {
			continue
		}
		if base := filepath.Base(dir); base != "." && base != "/" {
			return base
		}
	}
	return ""
}
