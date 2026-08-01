package typechecker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func checkInterfaceUpcastSource(t *testing.T, upcast string) []ModuleDiagnostic {
	t.Helper()
	source := strings.Join([]string{
		"package app",
		"",
		"interface Show for Self {",
		"  fn show(self: Self) -> String",
		"}",
		"",
		"interface Copyable for Self {",
		"  fn clone(self: Self) -> Self",
		"}",
		"",
		"struct Point { x: i32 }",
		"struct Wrapper <T> { value: T }",
		"",
		"impl Show for Point {",
		"  fn show(self: Self) -> String { \"point\" }",
		"}",
		"",
		"impl Copyable for Point {",
		"  fn clone(self: Self) -> Self { self }",
		"}",
		"",
		"impl <T: Show> Show for Wrapper T {",
		"  fn show(self: Self) -> String { \"show\" }",
		"}",
		"",
		"impl <T: Copyable> Show for Wrapper T {",
		"  fn show(self: Self) -> String { \"copy\" }",
		"}",
		"",
		"fn main() -> void {",
		"  wrapped := Wrapper { value: Point { x: 1 } }",
		"  " + upcast,
		"}",
		"",
	}, "\n")

	entryPath := filepath.Join(t.TempDir(), "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	t.Cleanup(func() { loader.Close() })
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("loader.Load: %v", err)
	}
	result, err := NewProgramChecker().Check(program)
	if err != nil {
		t.Fatalf("ProgramChecker.Check: %v", err)
	}
	return result.Diagnostics
}

func requireInterfaceUpcastAmbiguity(t *testing.T, diagnostics []ModuleDiagnostic) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Diagnostic.Message, "ambiguous implementations of Show for Wrapper") {
			return
		}
	}
	t.Fatalf("expected interface-upcast ambiguity diagnostic, got %v", diagnostics)
}

func TestTypedInterfaceUpcastRejectsAmbiguousImplementations(t *testing.T) {
	requireInterfaceUpcastAmbiguity(t, checkInterfaceUpcastSource(t, "captured: Show = wrapped"))
}

func TestExplicitInterfaceUpcastRejectsKnownAmbiguousImplementations(t *testing.T) {
	requireInterfaceUpcastAmbiguity(t, checkInterfaceUpcastSource(t, "captured := wrapped as Show"))
}
