package typechecker

import "able/interpreter-go/pkg/ast"

// MethodSelectionKind identifies the language mechanism that supplied a
// checked member method. It records semantic provenance, not runtime layout.
type MethodSelectionKind uint8

const (
	MethodSelectionUnknown MethodSelectionKind = iota
	MethodSelectionMethodSet
	MethodSelectionImplementation
)

// MethodSelection records the declaration chosen for a member access.
type MethodSelection struct {
	Kind              MethodSelectionKind
	MethodSet         *ast.MethodsDefinition
	Implementation    *ast.ImplementationDefinition
	Target            Type
	GenericNamedUnion bool
}

// MethodSelectionMap tracks selections by member-access AST node.
type MethodSelectionMap map[ast.Node]MethodSelection

func (m MethodSelectionMap) Clone() MethodSelectionMap {
	if m == nil {
		return nil
	}
	out := make(MethodSelectionMap, len(m))
	for node, selection := range m {
		out[node] = selection
	}
	return out
}

func (c *Checker) recordActiveMethodSelection(selection MethodSelection) {
	if c == nil || c.activeMemberAccess == nil || selection.Kind == MethodSelectionUnknown {
		return
	}
	if c.methodSelections == nil {
		c.methodSelections = make(MethodSelectionMap)
	}
	c.methodSelections[c.activeMemberAccess] = selection
}

func methodSetTargetsGenericNamedUnion(spec MethodSetSpec) bool {
	if len(spec.TypeParams) == 0 {
		return false
	}
	info, ok := structInfoFromType(spec.Target)
	return ok && info.isUnion && info.name != ""
}

func methodSelectionForImplementation(spec ImplementationSpec) MethodSelection {
	return MethodSelection{
		Kind:           MethodSelectionImplementation,
		Implementation: spec.Definition,
		Target:         spec.Target,
	}
}
