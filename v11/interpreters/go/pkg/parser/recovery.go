package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser/language"
)

func recoverableInterfaceBaseErrors(root *sitter.Node, source []byte) bool {
	if root == nil || !root.HasError() {
		return true
	}
	ok := true
	walkNodes(root, func(node *sitter.Node) {
		if node == nil || !ok {
			return
		}
		if node.IsMissing() {
			ok = false
			return
		}
		if !node.IsError() {
			return
		}
		iface := nearestInterfaceDefinition(node)
		if iface == nil {
			ok = false
			return
		}
		if _, _, recovered := recoverInterfaceBaseSelfType(iface, source); !recovered {
			ok = false
		}
	})
	return ok
}

func nearestInterfaceDefinition(node *sitter.Node) *sitter.Node {
	parent := node
	for parent != nil {
		if parent.Kind() == "interface_definition" {
			return parent
		}
		parent = parent.Parent()
	}
	return nil
}

func recoverInterfaceBaseSelfType(node *sitter.Node, source []byte) (ast.TypeExpression, *sitter.Node, bool) {
	if node == nil || node.Kind() != "interface_definition" {
		return nil, nil, false
	}
	var errorNode *sitter.Node
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil || child.Kind() != "ERROR" {
			continue
		}
		text := strings.TrimSpace(sliceContent(child, source))
		if strings.HasSuffix(text, ":") {
			errorNode = child
			break
		}
	}
	if errorNode == nil {
		return nil, nil, false
	}

	var baseNode *sitter.Node
	foundError := false
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if sameNode(child, errorNode) {
			foundError = true
			continue
		}
		if !foundError || !child.IsNamed() {
			continue
		}
		if parseTypeExpression(child, source) != nil {
			baseNode = child
			break
		}
	}
	if baseNode == nil {
		return nil, nil, false
	}

	var selfType ast.TypeExpression
	if selfNode := firstTypeExpressionChild(errorNode, source); selfNode != nil {
		selfType = parseTypeExpression(selfNode, source)
	}
	if selfType == nil {
		selfType = parseTypeExpressionFromText(interfaceBaseSelfTypeText(errorNode, source))
	}
	if selfType == nil {
		return nil, nil, false
	}
	return selfType, baseNode, true
}

func firstTypeExpressionChild(node *sitter.Node, source []byte) *sitter.Node {
	if node == nil {
		return nil
	}
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || !child.IsNamed() {
			continue
		}
		if parseTypeExpression(child, source) != nil {
			return child
		}
	}
	return nil
}

func interfaceBaseSelfTypeText(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	text := strings.TrimSpace(sliceContent(node, source))
	text = strings.TrimSuffix(text, ":")
	return strings.TrimSpace(text)
}

func parseTypeExpressionFromText(text string) ast.TypeExpression {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	source := []byte(fmt.Sprintf("type __Recovered = %s\n", text))
	parser := sitter.NewParser()
	if err := parser.SetLanguage(language.Able()); err != nil {
		return nil
	}
	defer parser.Close()
	tree := parser.Parse(source, nil)
	defer tree.Close()
	root := tree.RootNode()
	if root == nil || root.HasError() {
		return nil
	}
	var target *sitter.Node
	for i := uint(0); i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		if child == nil || child.Kind() != "type_alias_definition" {
			continue
		}
		target = child.ChildByFieldName("target")
		break
	}
	if target == nil {
		return nil
	}
	return parseTypeExpression(target, source)
}
