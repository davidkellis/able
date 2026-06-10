package parser

import (
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/parser/language"
)

func TestNodeKindMatchesTreeSitterKind(t *testing.T) {
	lang := language.Able()
	initializeNodeKindNames(lang)

	p := sitter.NewParser()
	defer p.Close()
	if err := p.SetLanguage(lang); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}
	tree := p.Parse([]byte("fn add(a: Int, b: Int): Int = a + b\n"), nil)
	defer tree.Close()

	var visit func(*sitter.Node)
	visit = func(node *sitter.Node) {
		if got, want := nodeKind(node), node.Kind(); got != want {
			t.Fatalf("nodeKind(%d) = %q, want %q", node.KindId(), got, want)
		}
		for i := uint(0); i < node.ChildCount(); i++ {
			visit(node.Child(i))
		}
	}
	visit(tree.RootNode())
}
