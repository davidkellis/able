package parser

import (
	"sync"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

var (
	nodeKindNamesOnce sync.Once
	nodeKindNames     []string
)

// initializeNodeKindNames resolves the grammar's stable symbol names once.
// Tree-sitter nodes expose compact symbol IDs, so AST mapping does not need to
// convert the same C strings on every kind check.
func initializeNodeKindNames(language *sitter.Language) {
	if language == nil {
		return
	}
	nodeKindNamesOnce.Do(func() {
		names := make([]string, int(language.NodeKindCount()))
		for id := range names {
			names[id] = language.NodeKindForId(uint16(id))
		}
		nodeKindNames = names
	})
}

func nodeKind(node *sitter.Node) string {
	if node == nil {
		return ""
	}
	id := int(node.KindId())
	if id >= 0 && id < len(nodeKindNames) {
		return nodeKindNames[id]
	}
	return node.Kind()
}
