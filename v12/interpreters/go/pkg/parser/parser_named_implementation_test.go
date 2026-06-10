package parser

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestParseNamedImplementationDefinition(t *testing.T) {
	parser, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer parser.Close()

	module, err := parser.ParseModule([]byte("Fancy = impl Marker for Widget {}"))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}
	if len(module.Body) != 1 {
		t.Fatalf("expected one statement, got %d", len(module.Body))
	}
	implementation, ok := module.Body[0].(*ast.ImplementationDefinition)
	if !ok {
		t.Fatalf("expected ImplementationDefinition, got %T", module.Body[0])
	}
	if implementation.ImplName == nil || implementation.ImplName.Name != "Fancy" {
		t.Fatalf("expected implementation name Fancy, got %#v", implementation.ImplName)
	}
	if implementation.InterfaceName == nil || implementation.InterfaceName.Name != "Marker" {
		t.Fatalf("expected interface Marker, got %#v", implementation.InterfaceName)
	}
	target, ok := implementation.TargetType.(*ast.SimpleTypeExpression)
	if !ok || target.Name == nil || target.Name.Name != "Widget" {
		t.Fatalf("expected target Widget, got %#v", implementation.TargetType)
	}
}
