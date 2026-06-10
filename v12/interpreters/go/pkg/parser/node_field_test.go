package parser

import (
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/parser/language"
)

func TestChildByFieldNameMatchesTreeSitterLookup(t *testing.T) {
	lang := language.Able()
	initializeNodeFieldIDs(lang)

	if got, want := len(nodeFieldIDs), int(lang.FieldCount()); got != want {
		t.Fatalf("cached field count = %d, want %d", got, want)
	}
	fieldNames := make([]string, 0, lang.FieldCount())
	for rawID := 1; rawID <= int(lang.FieldCount()); rawID++ {
		id := uint16(rawID)
		name := lang.FieldNameForId(id)
		if name == "" {
			t.Fatalf("field %d has no name", id)
		}
		if got := nodeFieldIDs[name]; got != id {
			t.Fatalf("cached field %q = %d, want %d", name, got, id)
		}
		fieldNames = append(fieldNames, name)
	}

	p := sitter.NewParser()
	defer p.Close()
	if err := p.SetLanguage(lang); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}
	tree := p.Parse([]byte("fn add(a: Int, b: Int): Int = a + b\n"), nil)
	defer tree.Close()

	var visit func(*sitter.Node)
	visit = func(node *sitter.Node) {
		for _, fieldName := range fieldNames {
			got := childByFieldName(node, fieldName)
			want := node.ChildByFieldName(fieldName)
			if (got == nil) != (want == nil) {
				t.Fatalf("childByFieldName(%q) nil mismatch: got %v, want %v", fieldName, got, want)
			}
			if got != nil && got.Id() != want.Id() {
				t.Fatalf("childByFieldName(%q) id = %d, want %d", fieldName, got.Id(), want.Id())
			}
		}
		for i := uint(0); i < node.ChildCount(); i++ {
			visit(node.Child(i))
		}
	}
	visit(tree.RootNode())

	if got := childByFieldName(nil, "name"); got != nil {
		t.Fatalf("childByFieldName(nil) = %v, want nil", got)
	}
}
