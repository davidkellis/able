package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestReserveNodeOriginsPreservesRegisteredOrigins(t *testing.T) {
	interp := New()
	interp.ReserveNodeOrigins(2)
	first := ast.ID("first")
	second := ast.ID("second")
	interp.AddNodeOrigin(first, "first.able")
	interp.AddNodeOrigin(second, "second.able")
	interp.ReserveNodeOrigins(1)

	if got := interp.nodeOrigins[first]; got != "first.able" {
		t.Fatalf("first origin = %q, want first.able", got)
	}
	if got := interp.nodeOrigins[second]; got != "second.able" {
		t.Fatalf("second origin = %q, want second.able", got)
	}
}
