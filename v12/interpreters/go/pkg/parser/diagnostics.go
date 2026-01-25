package parser

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// SourceLocation captures a source span for parser diagnostics.
type SourceLocation struct {
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// ParseError includes a message plus a best-effort source location.
type ParseError struct {
	Message  string
	Location SourceLocation
}

func (e *ParseError) Error() string {
	return e.Message
}

func wrapParseError(node *sitter.Node, err error) error {
	if err == nil {
		return nil
	}
	var parseErr *ParseError
	if errors.As(err, &parseErr) {
		return parseErr
	}
	if node == nil {
		return err
	}
	return &ParseError{
		Message:  err.Error(),
		Location: locationForNode(node),
	}
}

func syntaxError(root *sitter.Node) *ParseError {
	missing := findFirstMissingNode(root)
	errorNode := missing
	if errorNode == nil {
		errorNode = findFirstErrorNode(root)
	}
	if errorNode == nil {
		errorNode = root
	}
	location := SourceLocation{}
	if errorNode != nil {
		location = locationForNode(errorNode)
	}
	expected := ""
	if missing != nil {
		expected = formatExpectedKind(missing.Kind())
	}
	message := "parser: syntax error"
	if expected != "" {
		message = fmt.Sprintf("parser: syntax error: expected %s", expected)
	}
	return &ParseError{
		Message:  message,
		Location: location,
	}
}

func locationForNode(node *sitter.Node) SourceLocation {
	if node == nil {
		return SourceLocation{}
	}
	start := node.StartPosition()
	end := node.EndPosition()
	return SourceLocation{
		Line:      int(start.Row) + 1,
		Column:    int(start.Column) + 1,
		EndLine:   int(end.Row) + 1,
		EndColumn: int(end.Column) + 1,
	}
}

func findFirstMissingNode(root *sitter.Node) *sitter.Node {
	var best *sitter.Node
	walkNodes(root, func(node *sitter.Node) {
		if node == nil || !node.IsMissing() {
			return
		}
		if best == nil || node.StartByte() < best.StartByte() {
			best = node
		}
	})
	return best
}

func findFirstErrorNode(root *sitter.Node) *sitter.Node {
	var best *sitter.Node
	walkNodes(root, func(node *sitter.Node) {
		if node == nil || !node.IsError() {
			return
		}
		if best == nil || node.StartByte() < best.StartByte() {
			best = node
		}
	})
	return best
}

func walkNodes(root *sitter.Node, visit func(node *sitter.Node)) {
	if root == nil {
		return
	}
	visit(root)
	for i := uint(0); i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child == nil {
			continue
		}
		walkNodes(child, visit)
	}
}

func formatExpectedKind(kind string) string {
	trimmed := strings.TrimSpace(kind)
	if trimmed == "" {
		return "token"
	}
	isSymbol := true
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			isSymbol = false
			break
		}
	}
	if len(trimmed) == 1 || isSymbol {
		return fmt.Sprintf("'%s'", trimmed)
	}
	return strings.ReplaceAll(trimmed, "_", " ")
}
