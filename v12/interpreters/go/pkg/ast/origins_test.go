package ast

import "testing"

func TestAnnotateOriginsPreservesPartialTableSemantics(t *testing.T) {
	name := NewIdentifier("main")
	value := NewIdentifier("value")
	body := NewBlockExpression([]Statement{value})
	fn := NewFunctionDefinition(name, nil, body, nil, nil, nil, false, false)
	module := NewModule([]Statement{fn}, nil, nil)
	table := map[Node]string{module: "existing.able"}

	AnnotateOrigins(module, "new.able", table)
	if got := table[module]; got != "existing.able" {
		t.Fatalf("module origin = %q, want existing.able", got)
	}
	for _, node := range []Node{fn, name, body, value} {
		if got := table[node]; got != "new.able" {
			t.Fatalf("descendant %T origin = %q, want new.able", node, got)
		}
	}
}

func TestAnnotateOriginsSkippingKnownRequiresCompleteSubtrees(t *testing.T) {
	name := NewIdentifier("main")
	fn := NewFunctionDefinition(name, nil, NewBlockExpression(nil), nil, nil, nil, false, false)
	module := NewModule([]Statement{fn}, nil, nil)
	table := map[Node]string{module: "existing.able"}

	AnnotateOriginsSkippingKnown(module, "new.able", table)
	if _, ok := table[fn]; ok {
		t.Fatal("skip-known traversal unexpectedly filled a partial descendant")
	}
}
