package ast

// AnnotateOrigins assigns the provided source path to every node reachable from root.
// The table map may be nil; when provided it is populated with node -> path entries.
func AnnotateOrigins(root Node, path string, table map[Node]string) {
	annotateOriginsConfigured(root, path, table, false)
}

// AnnotateOriginsSkippingKnown assigns origins while treating an existing
// table entry as proof that the node's entire subtree was annotated by an
// earlier complete traversal. Callers must not use this with a partial table.
func AnnotateOriginsSkippingKnown(root Node, path string, table map[Node]string) {
	annotateOriginsConfigured(root, path, table, true)
}

func annotateOriginsConfigured(root Node, path string, table map[Node]string, skipKnownSubtrees bool) {
	if root == nil || path == "" {
		return
	}
	if table == nil {
		table = make(map[Node]string)
	}
	Walk(root, func(node Node) bool {
		if _, ok := table[node]; !ok {
			table[node] = path
			return true
		}
		if skipKnownSubtrees {
			return false
		}
		return true
	})
}
