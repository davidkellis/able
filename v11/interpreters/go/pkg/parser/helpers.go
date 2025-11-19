package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

// parseContext mirrors the TypeScript ParseContext glue: it carries immutable
// parser state (currently the module source bytes) so helpers can share the
// same view of the file without threading arguments everywhere.
type parseContext struct {
	source []byte
}

func newParseContext(source []byte) *parseContext {
	return &parseContext{source: source}
}

func (ctx *parseContext) parseQualifiedIdentifier(node *sitter.Node) ([]*ast.Identifier, error) {
	return parseQualifiedIdentifier(node, ctx.source)
}

func parseIdentifier(node *sitter.Node, source []byte) (*ast.Identifier, error) {
	if node == nil || node.Kind() != "identifier" {
		return nil, fmt.Errorf("parser: expected identifier")
	}
	content := sliceContent(node, source)
	id := ast.ID(content)
	annotateSpan(id, node)
	return id, nil
}

func sliceContent(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	start := int(node.StartByte())
	end := int(node.EndByte())
	if start < 0 || end < start || end > len(source) {
		return ""
	}
	return string(source[start:end])
}

func hasLeadingPrivate(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		if child.Kind() == "private" {
			return true
		}
		if child.IsNamed() {
			break
		}
	}
	return false
}

func firstNamedChild(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && !isIgnorableNode(child) {
			return child
		}
	}
	return nil
}

func nextNamedSibling(parent *sitter.Node, currentIndex uint) *sitter.Node {
	if parent == nil {
		return nil
	}
	for j := currentIndex + 1; j < parent.NamedChildCount(); j++ {
		sibling := parent.NamedChild(j)
		if sibling != nil && sibling.IsNamed() && !isIgnorableNode(sibling) {
			return sibling
		}
	}
	return nil
}

func findIdentifier(node *sitter.Node, source []byte) (*ast.Identifier, bool) {
	if node == nil {
		return nil, false
	}

	if isIgnorableNode(node) {
		return nil, false
	}

	if node.Kind() == "identifier" {
		id, err := parseIdentifier(node, source)
		if err != nil {
			return nil, false
		}
		return id, true
	}

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if id, ok := findIdentifier(child, source); ok {
			return id, true
		}
	}

	return nil, false
}

func findNamedChildIndex(parent, target *sitter.Node) int {
	if parent == nil || target == nil {
		return -1
	}
	for idx := uint(0); idx < parent.NamedChildCount(); idx++ {
		child := parent.NamedChild(idx)
		if child == nil || !child.IsNamed() || isIgnorableNode(child) {
			continue
		}
		if sameNode(child, target) {
			return int(idx)
		}
	}
	return -1
}

func hasSemicolonBetween(source []byte, left, right *sitter.Node) bool {
	if left == nil || right == nil {
		return false
	}
	start := int(left.EndByte())
	end := int(right.StartByte())
	if start < 0 || end < start || end > len(source) {
		return false
	}
	for i := start; i < end; i++ {
		if source[i] == ';' {
			return true
		}
	}
	return false
}

func parseLabel(node *sitter.Node, source []byte) (*ast.Identifier, error) {
	if node == nil || node.Kind() != "label" {
		return nil, fmt.Errorf("parser: expected label")
	}
	content := strings.TrimSpace(sliceContent(node, source))
	if content == "" {
		return nil, fmt.Errorf("parser: empty label")
	}
	if strings.HasPrefix(content, "'") {
		content = content[1:]
	}
	if content == "" {
		return nil, fmt.Errorf("parser: label missing identifier")
	}
	id := ast.ID(content)
	annotateSpan(id, node)
	return id, nil
}

func collapseQualifiedIdentifier(parts []*ast.Identifier) *ast.Identifier {
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		return parts[0]
	}
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == nil || part.Name == "" {
			continue
		}
		names = append(names, part.Name)
	}
	if len(names) == 0 {
		return nil
	}
	id := ast.ID(strings.Join(names, "."))
	start := parts[0].Span()
	end := parts[len(parts)-1].Span()
	if !isZeroSpan(start) {
		span := start
		if !isZeroSpan(end) {
			span.End = end.End
		}
		ast.SetSpan(id, span)
	}
	return id
}

func sameNode(a, b *sitter.Node) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Kind() == b.Kind() && a.StartByte() == b.StartByte() && a.EndByte() == b.EndByte()
}

func isIgnorableNode(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind() {
	case "comment", "line_comment", "block_comment":
		return true
	default:
		return false
	}
}
