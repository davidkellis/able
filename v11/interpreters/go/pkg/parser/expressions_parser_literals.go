package parser

import (
	"able/interpreter-go/pkg/ast"
	"fmt"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"strconv"
	"strings"
)

func (ctx *parseContext) parseImplicitMemberExpression(node *sitter.Node) (ast.Expression, error) {
	memberNode := node.ChildByFieldName("member")
	if memberNode == nil {
		return nil, fmt.Errorf("parser: implicit member missing identifier")
	}
	member, err := parseIdentifier(memberNode, ctx.source)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewImplicitMemberExpression(member), node), nil
}

func (ctx *parseContext) parsePlaceholderExpression(node *sitter.Node) (ast.Expression, error) {
	raw := strings.TrimSpace(sliceContent(node, ctx.source))
	if raw == "" {
		return nil, fmt.Errorf("parser: empty placeholder expression")
	}
	if raw == "@" {
		defaultIndex := 1
		return annotateExpression(ast.NewPlaceholderExpression(&defaultIndex), node), nil
	}
	if strings.HasPrefix(raw, "@") {
		value := raw[1:]
		if value == "" {
			defaultIndex := 1
			return annotateExpression(ast.NewPlaceholderExpression(&defaultIndex), node), nil
		}
		index, err := strconv.Atoi(value)
		if err != nil || index <= 0 {
			return nil, fmt.Errorf("parser: invalid placeholder index %q", raw)
		}
		return annotateExpression(ast.NewPlaceholderExpression(&index), node), nil
	}
	return nil, fmt.Errorf("parser: unsupported placeholder token %q", raw)
}

func (ctx *parseContext) parseInterpolatedString(node *sitter.Node) (ast.Expression, error) {
	parts := make([]ast.Expression, 0)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		switch child.Kind() {
		case "interpolation_text":
			text := unescapeInterpolationText(sliceContent(child, ctx.source))
			if text != "" {
				parts = append(parts, annotateExpression(ast.Str(text), child))
			}
		case "string_interpolation":
			exprNode := child.ChildByFieldName("expression")
			if exprNode == nil {
				return nil, fmt.Errorf("parser: interpolation missing expression")
			}
			expr, err := ctx.parseExpression(exprNode)
			if err != nil {
				return nil, err
			}
			parts = append(parts, expr)
		}
	}
	return annotateExpression(ast.NewStringInterpolation(parts), node), nil
}

func unescapeInterpolationText(text string) string {
	if !strings.Contains(text, "\\") {
		return text
	}
	var builder strings.Builder
	builder.Grow(len(text))
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if ch != '\\' {
			builder.WriteByte(ch)
			continue
		}
		if i+1 >= len(text) {
			builder.WriteByte('\\')
			break
		}
		next := text[i+1]
		switch next {
		case '`':
			builder.WriteByte('`')
			i++
		case '$':
			builder.WriteByte('$')
			i++
		case '\\':
			builder.WriteByte('\\')
			i++
		default:
			builder.WriteByte('\\')
			builder.WriteByte(next)
			i++
		}
	}
	return builder.String()
}

func (ctx *parseContext) parseIteratorLiteral(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "iterator_literal" {
		return nil, fmt.Errorf("parser: expected iterator_literal node")
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: iterator literal missing body")
	}

	var binding *ast.Identifier
	if bindingNode := bodyNode.ChildByFieldName("binding"); bindingNode != nil {
		id, err := parseIdentifier(bindingNode, ctx.source)
		if err != nil {
			return nil, err
		}
		binding = id
	}

	var elementType ast.TypeExpression
	if elementTypeNode := node.ChildByFieldName("element_type"); elementTypeNode != nil {
		elementType = parseTypeExpression(elementTypeNode, ctx.source)
	}

	block, err := ctx.parseBlock(bodyNode)
	if err != nil {
		return nil, err
	}

	literal := ast.NewIteratorLiteral(block.Body)
	literal.Binding = binding
	literal.ElementType = elementType
	return annotateExpression(literal, node), nil
}

func (ctx *parseContext) parseBooleanLiteral(node *sitter.Node) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, ctx.source))
	switch value {
	case "true":
		return annotateExpression(ast.Bool(true), node), nil
	case "false":
		return annotateExpression(ast.Bool(false), node), nil
	default:
		return nil, fmt.Errorf("parser: invalid boolean literal %q", value)
	}
}

func (ctx *parseContext) parseNilLiteral(node *sitter.Node) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, ctx.source))
	if value != "nil" {
		return nil, fmt.Errorf("parser: invalid nil literal %q", value)
	}
	return annotateExpression(ast.Nil(), node), nil
}

func (ctx *parseContext) parseArrayLiteral(node *sitter.Node) (ast.Expression, error) {
	elements := make([]ast.Expression, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || !child.IsNamed() || isIgnorableNode(child) {
			continue
		}
		element, err := ctx.parseExpression(child)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	return annotateExpression(ast.NewArrayLiteral(elements), node), nil
}

func extractStructLiteralType(expr ast.TypeExpression) (*ast.Identifier, []ast.TypeExpression, error) {
	switch typed := expr.(type) {
	case *ast.SimpleTypeExpression:
		if typed.Name == nil {
			return nil, nil, fmt.Errorf("parser: struct literal missing type identifier")
		}
		return typed.Name, nil, nil
	case *ast.GenericTypeExpression:
		base, baseArgs, err := extractStructLiteralType(typed.Base)
		if err != nil {
			return nil, nil, err
		}
		args := make([]ast.TypeExpression, 0, len(baseArgs)+len(typed.Arguments))
		args = append(args, baseArgs...)
		args = append(args, typed.Arguments...)
		return base, args, nil
	default:
		return nil, nil, fmt.Errorf("parser: struct literal type must be nominal")
	}
}

func (ctx *parseContext) parseStructLiteral(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "struct_literal" {
		return nil, fmt.Errorf("parser: expected struct literal node")
	}

	typeNode := node.ChildByFieldName("type")
	if typeNode == nil {
		return nil, fmt.Errorf("parser: struct literal missing type")
	}
	typeExpr := ctx.parseTypeExpression(typeNode)
	if typeExpr == nil {
		return nil, fmt.Errorf("parser: invalid struct literal type")
	}
	structType, typeArgs, err := extractStructLiteralType(typeExpr)
	if err != nil {
		return nil, err
	}

	fields := make([]*ast.StructFieldInitializer, 0)
	var functionalUpdates []ast.Expression

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		fieldName := node.FieldNameForChild(uint32(i))
		if fieldName == "type" || sameNode(child, typeNode) {
			continue
		}

		var elem *sitter.Node
		if child.Kind() == "struct_literal_element" {
			elem = firstNamedChild(child)
		} else {
			elem = child
		}
		if elem == nil {
			continue
		}

		switch elem.Kind() {
		case "struct_literal_field":
			nameNode := elem.ChildByFieldName("name")
			if nameNode == nil {
				return nil, fmt.Errorf("parser: struct literal field missing name")
			}
			name, err := parseIdentifier(nameNode, ctx.source)
			if err != nil {
				return nil, err
			}
			valueNode := elem.ChildByFieldName("value")
			if valueNode == nil {
				return nil, fmt.Errorf("parser: struct literal field missing value")
			}
			value, err := ctx.parseExpression(valueNode)
			if err != nil {
				return nil, err
			}
			field := ast.NewStructFieldInitializer(value, name, false)
			annotateSpan(field, elem)
			fields = append(fields, field)
		case "struct_literal_shorthand_field":
			nameNode := elem.ChildByFieldName("name")
			if nameNode == nil {
				nameNode = firstNamedChild(elem)
			}
			name, err := parseIdentifier(nameNode, ctx.source)
			if err != nil {
				return nil, err
			}
			field := ast.NewStructFieldInitializer(nil, name, true)
			annotateSpan(field, elem)
			fields = append(fields, field)
		case "struct_literal_spread":
			exprNode := firstNamedChild(elem)
			if exprNode == nil {
				return nil, fmt.Errorf("parser: struct spread missing expression")
			}
			expr, err := ctx.parseExpression(exprNode)
			if err != nil {
				return nil, err
			}
			functionalUpdates = append(functionalUpdates, expr)
		default:
			expr, err := ctx.parseExpression(elem)
			if err != nil {
				return nil, err
			}
			field := ast.NewStructFieldInitializer(expr, nil, false)
			annotateSpan(field, elem)
			fields = append(fields, field)
		}
	}

	positional := false
	for _, field := range fields {
		if field.Name == nil {
			positional = true
			break
		}
	}

	return annotateExpression(ast.NewStructLiteral(fields, positional, structType, functionalUpdates, typeArgs), node), nil
}

func (ctx *parseContext) parseMapLiteral(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "map_literal" {
		return nil, fmt.Errorf("parser: expected map literal node")
	}
	elements := make([]ast.MapLiteralElement, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		if child.Kind() == "map_literal_element" {
			if nested := firstNamedChild(child); nested != nil {
				child = nested
			}
		}
		switch child.Kind() {
		case "map_literal_entry":
			keyNode := child.ChildByFieldName("key")
			valueNode := child.ChildByFieldName("value")
			if keyNode == nil || valueNode == nil {
				return nil, fmt.Errorf("parser: map literal entry missing key or value")
			}
			keyExpr, err := ctx.parseExpression(keyNode)
			if err != nil {
				return nil, err
			}
			valueExpr, err := ctx.parseExpression(valueNode)
			if err != nil {
				return nil, err
			}
			entry := ast.NewMapLiteralEntry(keyExpr, valueExpr)
			annotateSpan(entry, child)
			elements = append(elements, entry)
		case "map_literal_spread":
			exprNode := child.ChildByFieldName("expression")
			if exprNode == nil {
				exprNode = firstNamedChild(child)
			}
			if exprNode == nil {
				return nil, fmt.Errorf("parser: map literal spread missing expression")
			}
			spreadExpr, err := ctx.parseExpression(exprNode)
			if err != nil {
				return nil, err
			}
			spread := ast.NewMapLiteralSpread(spreadExpr)
			annotateSpan(spread, child)
			elements = append(elements, spread)
		default:
			return nil, fmt.Errorf("parser: unsupported map literal element %s", child.Kind())
		}
	}
	return annotateExpression(ast.NewMapLiteral(elements), node), nil
}
