package typechecker

import "able/interpreter-go/pkg/ast"

// InferenceMap tracks the resolved type of each AST node.
type InferenceMap map[ast.Node]Type

// set records a type for a node.
func (m InferenceMap) set(node ast.Node, typ Type) {
	if node == nil || typ == nil {
		return
	}
	m[node] = typ
}

// get retrieves a type for a node.
func (m InferenceMap) get(node ast.Node) (Type, bool) {
	typ, ok := m[node]
	return typ, ok
}
