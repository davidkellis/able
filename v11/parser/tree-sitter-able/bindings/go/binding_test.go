package tree_sitter_able_test

import (
	"testing"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_able "github.com/davidkellis/able/bindings/go"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_able.Language())
	if language == nil {
		t.Errorf("Error loading Able grammar")
	}
}
