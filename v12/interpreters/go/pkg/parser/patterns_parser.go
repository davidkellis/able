package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
)

func (ctx *parseContext) parsePattern(node *sitter.Node) (ast.Pattern, error) {
	pattern, err := ctx.parsePatternInternal(node)
	if err != nil {
		return nil, wrapParseError(node, err)
	}
	return pattern, nil
}

func (ctx *parseContext) parsePatternInternal(node *sitter.Node) (ast.Pattern, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil pattern")
	}

	source := ctx.source

	if node.Kind() == "pattern" || node.Kind() == "pattern_base" {
		if node.NamedChildCount() == 0 {
			text := strings.TrimSpace(sliceContent(node, source))
			if text == "_" {
				return annotatePattern(ast.Wc(), node), nil
			}
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				if child == nil || isIgnorableNode(child) {
					continue
				}
				if child.IsNamed() {
					return ctx.parsePattern(child)
				}
				if strings.TrimSpace(sliceContent(child, source)) == "_" {
					return annotatePattern(ast.Wc(), child), nil
				}
			}
			return nil, fmt.Errorf("parser: empty %s", node.Kind())
		}
		return ctx.parsePattern(node.NamedChild(0))
	}

	switch node.Kind() {
	case "identifier":
		expr, err := parseIdentifier(node, source)
		if err != nil {
			return nil, err
		}
		return annotatePattern(expr, node), nil
	case "_":
		return annotatePattern(ast.Wc(), node), nil
	case "literal_pattern":
		pattern, err := ctx.parseLiteralPattern(node)
		if err != nil {
			return nil, err
		}
		return annotatePattern(pattern, node), nil
	case "struct_pattern":
		return ctx.parseStructPattern(node)
	case "array_pattern":
		return ctx.parseArrayPattern(node)
	case "parenthesized_pattern":
		if inner := firstNamedChild(node); inner != nil {
			return ctx.parsePattern(inner)
		}
		return nil, fmt.Errorf("parser: empty parenthesized pattern")
	case "typed_pattern":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed typed pattern")
		}
		innerPattern, err := ctx.parsePattern(node.NamedChild(0))
		if err != nil {
			return nil, err
		}
		typeExpr := ctx.parseTypeExpression(node.NamedChild(1))
		if typeExpr == nil {
			return nil, fmt.Errorf("parser: typed pattern missing type expression")
		}
		return annotatePattern(ast.NewTypedPattern(innerPattern, typeExpr), node), nil
	case "pattern", "pattern_base":
		pattern, err := ctx.parsePattern(node.NamedChild(0))
		if err != nil {
			return nil, err
		}
		return annotatePattern(pattern, node), nil
	default:
		return nil, fmt.Errorf("parser: unsupported pattern kind %q", node.Kind())
	}
}

func (ctx *parseContext) parseLiteralPattern(node *sitter.Node) (ast.Pattern, error) {
	if node == nil || node.Kind() != "literal_pattern" {
		return nil, fmt.Errorf("parser: expected literal_pattern node")
	}
	literalNode := firstNamedChild(node)
	if literalNode == nil {
		return nil, fmt.Errorf("parser: literal pattern missing literal")
	}

	litExpr, err := ctx.parseExpression(literalNode)
	if err != nil {
		return nil, err
	}

	literal, ok := litExpr.(ast.Literal)
	if !ok {
		return nil, fmt.Errorf("parser: literal pattern must contain literal, found %T", litExpr)
	}

	pattern := ast.NewLiteralPattern(literal)
	annotateSpan(pattern, node)
	return pattern, nil
}

func (ctx *parseContext) parseStructPattern(node *sitter.Node) (ast.Pattern, error) {
	if node == nil || node.Kind() != "struct_pattern" {
		return nil, fmt.Errorf("parser: expected struct_pattern node")
	}

	var structType *ast.Identifier
	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		parts, err := parseQualifiedIdentifier(typeNode, ctx.source)
		if err != nil {
			return nil, err
		}
		if len(parts) == 0 {
			return nil, fmt.Errorf("parser: struct pattern type missing identifier")
		}
		structType = collapseQualifiedIdentifier(parts)
		if structType == nil {
			return nil, fmt.Errorf("parser: struct pattern type missing identifier")
		}
	}

	fields := make([]*ast.StructPatternField, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
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
			field, err := ctx.parseStructPatternField(elem)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
			continue
		}
		pattern, err := ctx.parsePattern(elem)
		if err != nil {
			return nil, err
		}
		var fieldName *ast.Identifier
		if structType != nil {
			if ident, ok := pattern.(*ast.Identifier); ok && ident != nil {
				fieldName = ident
			}
		}
		field := ast.NewStructPatternField(pattern, fieldName, nil, nil)
		annotateSpan(field, elem)
		fields = append(fields, field)
	}

	pattern := ast.NewStructPattern(fields, false, structType)
	ctx.normalizeStructPattern(pattern)
	annotatePattern(pattern, node)
	return pattern, nil
}

func hasPositionalStructFields(fields []*ast.StructPatternField) bool {
	for _, field := range fields {
		if field.FieldName == nil {
			return true
		}
	}
	return false
}

func (ctx *parseContext) normalizeStructPattern(pattern *ast.StructPattern) {
	if pattern == nil || ctx == nil {
		return
	}
	structType := pattern.StructType
	structKind, hasStructKind := ctx.resolveStructKind(structType)
	if structType != nil && (!hasStructKind || structKind != ast.StructKindPositional) {
		for _, field := range pattern.Fields {
			if field == nil || field.FieldName != nil || field.Pattern == nil {
				continue
			}
			if ident, ok := field.Pattern.(*ast.Identifier); ok && ident != nil {
				field.FieldName = ident
			}
		}
	}
	isPositional := false
	if structType == nil {
		isPositional = true
	} else if hasStructKind && structKind == ast.StructKindPositional {
		isPositional = true
	} else if hasStructKind && structKind == ast.StructKindNamed {
		isPositional = false
	} else {
		isPositional = hasPositionalStructFields(pattern.Fields)
	}
	pattern.IsPositional = isPositional
}

func extractIdentifierName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	end := 0
	for end < len(raw) {
		ch := raw[end]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || (ch >= '0' && ch <= '9') {
			end++
			continue
		}
		break
	}
	return strings.TrimSpace(raw[:end])
}

func (ctx *parseContext) parseStructPatternField(node *sitter.Node) (*ast.StructPatternField, error) {
	if node == nil || node.Kind() != "struct_pattern_field" {
		return nil, fmt.Errorf("parser: expected struct_pattern_field node")
	}

	var fieldName *ast.Identifier
	if nameNode := node.ChildByFieldName("field"); nameNode != nil {
		name, err := parseIdentifier(nameNode, ctx.source)
		if err != nil {
			return nil, err
		}
		fieldName = name
	}

	var binding *ast.Identifier
	if bindingNode := node.ChildByFieldName("binding"); bindingNode != nil {
		id, err := parseIdentifier(bindingNode, ctx.source)
		if err != nil {
			return nil, err
		}
		binding = id
	}

	var typeAnnotation ast.TypeExpression
	if typeNode := node.ChildByFieldName("type"); typeNode != nil {
		if typeExpr := ctx.parseTypeExpression(typeNode); typeExpr != nil {
			typeAnnotation = typeExpr
		}
	}

	raw := strings.TrimSpace(sliceContent(node, ctx.source))
	prefix := raw
	if idx := strings.Index(prefix, ":"); idx >= 0 {
		prefix = strings.TrimSpace(prefix[:idx])
	}
	if strings.Contains(prefix, "::") {
		parts := strings.SplitN(prefix, "::", 2)
		if len(parts) == 2 {
			left := extractIdentifierName(strings.TrimSpace(parts[0]))
			right := extractIdentifierName(strings.TrimSpace(parts[1]))
			if left != "" && right != "" {
				if fieldName == nil || fieldName.Name != left {
					fieldName = ast.ID(left)
				}
				binding = ast.ID(right)
			}
		}
	}

	var pattern ast.Pattern
	if valueNode := node.ChildByFieldName("value"); valueNode != nil {
		valuePattern, err := ctx.parsePattern(valueNode)
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

	if sp, ok := pattern.(*ast.StructPattern); ok && sp.StructType == nil {
		if simple, ok := typeAnnotation.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
			sp.StructType = simple.Name
			ctx.normalizeStructPattern(sp)
		}
	}

	var bindingForField *ast.Identifier
	if valueNode := node.ChildByFieldName("value"); valueNode != nil && binding != nil {
		bindingForField = binding
	}

	field := ast.NewStructPatternField(pattern, fieldName, bindingForField, typeAnnotation)
	annotateSpan(field, node)
	return field, nil
}

func (ctx *parseContext) parseArrayPattern(node *sitter.Node) (ast.Pattern, error) {
	if node == nil || node.Kind() != "array_pattern" {
		return nil, fmt.Errorf("parser: expected array_pattern node")
	}

	elements := make([]ast.Pattern, 0)
	var rest ast.Pattern

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		if child.Kind() == "array_pattern_rest" {
			if rest != nil {
				return nil, fmt.Errorf("parser: multiple array rest patterns")
			}
			rp, err := ctx.parseArrayPatternRest(child)
			if err != nil {
				return nil, err
			}
			rest = rp
			continue
		}
		pattern, err := ctx.parsePattern(child)
		if err != nil {
			return nil, err
		}
		elements = append(elements, pattern)
	}

	pattern := ast.NewArrayPattern(elements, rest)
	annotatePattern(pattern, node)
	return pattern, nil
}

func (ctx *parseContext) parseArrayPatternRest(node *sitter.Node) (ast.Pattern, error) {
	if node == nil || node.Kind() != "array_pattern_rest" {
		return nil, fmt.Errorf("parser: expected array_pattern_rest node")
	}

	if node.NamedChildCount() == 0 {
		return annotatePattern(ast.Wc(), node), nil
	}

	restNode := node.NamedChild(0)
	pattern, err := ctx.parsePattern(restNode)
	if err != nil {
		return nil, err
	}
	return annotatePattern(pattern, node), nil
}

// Legacy wrapper for modules that still call the function form directly.
