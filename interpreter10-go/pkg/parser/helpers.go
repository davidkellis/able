package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func parseIdentifier(node *sitter.Node, source []byte) (*ast.Identifier, error) {
	if node == nil || node.Kind() != "identifier" {
		return nil, fmt.Errorf("parser: expected identifier")
	}
	content := sliceContent(node, source)
	return ast.ID(content), nil
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
	return ast.ID(content), nil
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
