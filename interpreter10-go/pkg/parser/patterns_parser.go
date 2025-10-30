package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func parsePattern(node *sitter.Node, source []byte) (ast.Pattern, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil pattern")
	}

	if node.Kind() == "pattern" || node.Kind() == "pattern_base" {
		if node.NamedChildCount() == 0 {
			text := strings.TrimSpace(sliceContent(node, source))
			if text == "_" {
				return ast.Wc(), nil
			}
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				if child == nil {
					continue
				}
				if child.IsNamed() {
					return parsePattern(child, source)
				}
				if strings.TrimSpace(sliceContent(child, source)) == "_" {
					return ast.Wc(), nil
				}
			}
			return nil, fmt.Errorf("parser: empty %s", node.Kind())
		}
		return parsePattern(node.NamedChild(0), source)
	}

	switch node.Kind() {
	case "identifier":
		return parseIdentifier(node, source)
	case "_":
		return ast.Wc(), nil
	case "literal_pattern":
		return parseLiteralPattern(node, source)
	case "struct_pattern":
		return parseStructPattern(node, source)
	case "array_pattern":
		return parseArrayPattern(node, source)
	case "parenthesized_pattern":
		if inner := firstNamedChild(node); inner != nil {
			return parsePattern(inner, source)
		}
		return nil, fmt.Errorf("parser: empty parenthesized pattern")
	case "typed_pattern":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed typed pattern")
		}
		innerPattern, err := parsePattern(node.NamedChild(0), source)
		if err != nil {
			return nil, err
		}
		typeExpr := parseTypeExpression(node.NamedChild(1), source)
		if typeExpr == nil {
			return nil, fmt.Errorf("parser: typed pattern missing type expression")
		}
		return ast.NewTypedPattern(innerPattern, typeExpr), nil
	case "pattern", "pattern_base":
		return parsePattern(node.NamedChild(0), source)
	default:
		return nil, fmt.Errorf("parser: unsupported pattern kind %q", node.Kind())
	}
}

func parseLiteralPattern(node *sitter.Node, source []byte) (ast.Pattern, error) {
	if node == nil || node.Kind() != "literal_pattern" {
		return nil, fmt.Errorf("parser: expected literal_pattern node")
	}
	literalNode := firstNamedChild(node)
	if literalNode == nil {
		return nil, fmt.Errorf("parser: literal pattern missing literal")
	}

	litExpr, err := parseExpression(literalNode, source)
	if err != nil {
		return nil, err
	}

	literal, ok := litExpr.(ast.Literal)
	if !ok {
		return nil, fmt.Errorf("parser: literal pattern must contain literal, found %T", litExpr)
	}

	return ast.NewLiteralPattern(literal), nil
}

func parseStructPattern(node *sitter.Node, source []byte) (ast.Pattern, error) {
	if node == nil || node.Kind() != "struct_pattern" {
		return nil, fmt.Errorf("parser: expected struct_pattern node")
	}

	var structType *ast.Identifier
	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		parts, err := parseQualifiedIdentifier(typeNode, source)
		if err != nil {
			return nil, err
		}
		if len(parts) == 0 {
			return nil, fmt.Errorf("parser: struct pattern type missing identifier")
		}
		structType = parts[len(parts)-1]
		if len(parts) > 1 {
			structType = ast.ID(strings.Join(identifiersToStrings(parts), "."))
		}
	}

	fields := make([]*ast.StructPatternField, 0)
	isPositional := false
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if field := node.FieldNameForChild(uint32(i)); field == "type" {
			continue
		}
		if sameNode(child, typeNode) {
			continue
		}
		elem := child
		if child.Kind() == "struct_pattern_element" {
			elem = firstNamedChild(child)
			if elem == nil {
				continue
			}
		}

		if elem.Kind() == "struct_pattern_field" {
			if elem.ChildByFieldName("binding") == nil && elem.ChildByFieldName("value") == nil {
				fieldNode := elem.ChildByFieldName("field")
				if fieldNode == nil {
					return nil, fmt.Errorf("parser: struct pattern field missing identifier")
				}
				pat, err := parseIdentifier(fieldNode, source)
				if err != nil {
					return nil, err
				}
				fields = append(fields, ast.NewStructPatternField(pat, nil, nil))
				continue
			}
			field, err := parseStructPatternField(elem, source)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
			continue
		}
		pattern, err := parsePattern(elem, source)
		if err != nil {
			return nil, err
		}
		fields = append(fields, ast.NewStructPatternField(pattern, nil, nil))
	}

	isPositional = false
	for _, field := range fields {
		if field.FieldName == nil {
			isPositional = true
			break
		}
	}

	return ast.NewStructPattern(fields, isPositional, structType), nil
}

func parseStructPatternField(node *sitter.Node, source []byte) (*ast.StructPatternField, error) {
	if node == nil || node.Kind() != "struct_pattern_field" {
		return nil, fmt.Errorf("parser: expected struct_pattern_field node")
	}

	var fieldName *ast.Identifier
	if nameNode := node.ChildByFieldName("field"); nameNode != nil {
		name, err := parseIdentifier(nameNode, source)
		if err != nil {
			return nil, err
		}
		fieldName = name
	}

	var binding *ast.Identifier
	if bindingNode := node.ChildByFieldName("binding"); bindingNode != nil {
		id, err := parseIdentifier(bindingNode, source)
		if err != nil {
			return nil, err
		}
		binding = id
	}

	var pattern ast.Pattern
	if valueNode := node.ChildByFieldName("value"); valueNode != nil {
		valuePattern, err := parsePattern(valueNode, source)
		if err != nil {
			return nil, err
		}
		pattern = valuePattern
	} else {
		switch {
		case binding != nil:
			pattern = binding
		case fieldName != nil:
			pattern = fieldName
		default:
			pattern = ast.Wc()
		}
	}

	return ast.NewStructPatternField(pattern, fieldName, binding), nil
}

func parseArrayPattern(node *sitter.Node, source []byte) (ast.Pattern, error) {
	if node == nil || node.Kind() != "array_pattern" {
		return nil, fmt.Errorf("parser: expected array_pattern node")
	}

	elements := make([]ast.Pattern, 0)
	var rest ast.Pattern

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Kind() == "array_pattern_rest" {
			if rest != nil {
				return nil, fmt.Errorf("parser: multiple array rest patterns")
			}
			rp, err := parseArrayPatternRest(child, source)
			if err != nil {
				return nil, err
			}
			rest = rp
			continue
		}
		pattern, err := parsePattern(child, source)
		if err != nil {
			return nil, err
		}
		elements = append(elements, pattern)
	}

	return ast.NewArrayPattern(elements, rest), nil
}

func parseArrayPatternRest(node *sitter.Node, source []byte) (ast.Pattern, error) {
	if node == nil || node.Kind() != "array_pattern_rest" {
		return nil, fmt.Errorf("parser: expected array_pattern_rest node")
	}

	if node.NamedChildCount() == 0 {
		return ast.Wc(), nil
	}

	restNode := node.NamedChild(0)
	return parsePattern(restNode, source)
}
