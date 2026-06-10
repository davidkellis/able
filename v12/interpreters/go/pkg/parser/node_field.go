package parser

import (
	"sync"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

var (
	nodeFieldIDsOnce sync.Once
	nodeFieldIDs     map[string]uint16
)

// initializeNodeFieldIDs resolves the grammar's stable field IDs once. The
// binding's name-based node accessor otherwise allocates a C string and asks
// Tree-sitter to resolve the same field name on every lookup.
func initializeNodeFieldIDs(language *sitter.Language) {
	if language == nil {
		return
	}
	nodeFieldIDsOnce.Do(func() {
		fieldCount := int(language.FieldCount())
		ids := make(map[string]uint16, fieldCount)
		for rawID := 1; rawID <= fieldCount; rawID++ {
			id := uint16(rawID)
			if name := language.FieldNameForId(id); name != "" {
				ids[name] = id
			}
		}
		nodeFieldIDs = ids
	})
}

func childByFieldName(node *sitter.Node, name string) *sitter.Node {
	if node == nil {
		return nil
	}
	if id, ok := nodeFieldIDs[name]; ok {
		return node.ChildByFieldId(id)
	}
	return node.ChildByFieldName(name)
}
