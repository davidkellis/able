package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestIsIteratorEndRecognizesSingletonStructDefinition(t *testing.T) {
	interp := New()
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef("IteratorEnd", nil, ast.StructKindNamed, nil, nil, false),
	}
	if !interp.isIteratorEnd(def) {
		t.Fatalf("expected singleton IteratorEnd definition to be recognized")
	}
}
