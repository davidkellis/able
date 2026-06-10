package ast

import "testing"

func TestWalkVisitsNestedNodesOnce(t *testing.T) {
	shared := NewIdentifier("shared")
	module := NewModule([]Statement{
		NewFunctionDefinition(shared, nil, NewBlockExpression([]Statement{shared}), nil, nil, nil, false, false),
	}, nil, NewPackageStatement([]*Identifier{NewIdentifier("sample")}, false))

	visits := make(map[Node]int)
	Walk(module, func(node Node) bool {
		visits[node]++
		return true
	})

	if visits[module] != 1 || visits[shared] != 1 {
		t.Fatalf("visit counts: module=%d shared=%d", visits[module], visits[shared])
	}
	if len(visits) < 5 {
		t.Fatalf("visited %d nodes, want nested module graph", len(visits))
	}
}

func TestWalkCanSkipOneSubtreeWithoutStoppingSiblings(t *testing.T) {
	firstChild := NewIdentifier("first_child")
	first := NewBlockExpression([]Statement{firstChild})
	secondChild := NewIdentifier("second_child")
	second := NewBlockExpression([]Statement{secondChild})
	module := NewModule([]Statement{first, second}, nil, nil)

	seen := make(map[Node]bool)
	Walk(module, func(node Node) bool {
		seen[node] = true
		return node != first
	})

	if seen[firstChild] {
		t.Fatal("visited child of pruned subtree")
	}
	if !seen[second] || !seen[secondChild] {
		t.Fatal("pruning first subtree stopped a sibling traversal")
	}
}
