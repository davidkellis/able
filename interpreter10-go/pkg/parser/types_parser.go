package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func applyGenericType(base ast.TypeExpression, args []ast.TypeExpression) ast.TypeExpression {
	if base == nil {
		return nil
	}
	switch typed := base.(type) {
	case *ast.NullableTypeExpression:
		inner := applyGenericType(typed.InnerType, args)
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		inner := applyGenericType(typed.InnerType, args)
		return ast.NewResultTypeExpression(inner)
	default:
		return ast.NewGenericTypeExpression(base, args)
	}
}

func typeArgumentExpressions(node *sitter.Node, source []byte) []ast.TypeExpression {
	args, err := parseTypeArgumentList(node, source)
	if err != nil {
		return nil
	}
	return args
}

func parseReturnType(node *sitter.Node, source []byte) ast.TypeExpression {
	return parseTypeExpression(node, source)
}

func parseTypeExpression(node *sitter.Node, source []byte) ast.TypeExpression {
	if node == nil {
		return nil
	}
	switch node.Kind() {
	case "return_type", "type_expression", "type_prefix", "type_atom":
		if node.NamedChildCount() == 0 {
			break
		}
		if child := firstNamedChild(node); child != nil && child != node {
			expr := parseTypeExpression(child, source)
			if expr == nil {
				return nil
			}
			if node.Kind() == "type_prefix" {
				text := strings.TrimSpace(sliceContent(node, source))
				for len(text) > 0 && (text[0] == '?' || text[0] == '!') {
					switch text[0] {
					case '?':
						expr = ast.NewNullableTypeExpression(expr)
					case '!':
						expr = ast.NewResultTypeExpression(expr)
					}
					text = text[1:]
				}
			}
			return annotateTypeExpression(expr, node)
		}
	case "type_suffix":
		if node.NamedChildCount() > 1 {
			base := parseTypeExpression(node.NamedChild(0), source)
			var args []ast.TypeExpression
			for i := uint(1); i < node.NamedChildCount(); i++ {
				child := node.NamedChild(i)
				if child == nil || !child.IsNamed() {
					continue
				}
				if child.Kind() == "type_arguments" {
					args = append(args, typeArgumentExpressions(child, source)...)
					continue
				}
				arg := parseTypeExpression(child, source)
				if arg != nil {
					args = append(args, arg)
				}
			}
			if base != nil && len(args) > 0 {
				return annotateTypeExpression(applyGenericType(base, args), node)
			}
		}
		if child := firstNamedChild(node); child != nil && child != node {
			expr := parseTypeExpression(child, source)
			if expr == nil {
				return nil
			}
			return annotateTypeExpression(expr, node)
		}
	case "type_arrow":
		if node.NamedChildCount() >= 2 {
			paramTypes, ok := parseFunctionParameterTypes(node.NamedChild(0), source)
			if !ok {
				break
			}
			returnExpr := parseTypeExpression(node.NamedChild(1), source)
			if returnExpr != nil {
				return annotateTypeExpression(ast.NewFunctionTypeExpression(paramTypes, returnExpr), node)
			}
		}
		if child := firstNamedChild(node); child != nil && child != node {
			expr := parseTypeExpression(child, source)
			if expr == nil {
				return nil
			}
			return annotateTypeExpression(expr, node)
		}
	case "type_generic_application":
		if node.NamedChildCount() == 0 {
			break
		}
		base := parseTypeExpression(node.NamedChild(0), source)
		if base == nil {
			return nil
		}
		var args []ast.TypeExpression
		for i := uint(1); i < node.NamedChildCount(); i++ {
			arg := parseTypeExpression(node.NamedChild(i), source)
			if arg != nil {
				args = append(args, arg)
			}
		}
		if len(args) == 0 {
			return annotateTypeExpression(base, node)
		}
		return annotateTypeExpression(applyGenericType(base, args), node)
	case "type_union":
		var members []ast.TypeExpression
		for i := uint(0); i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			member := parseTypeExpression(child, source)
			if member != nil {
				members = append(members, member)
			}
		}
		if len(members) == 1 {
			return annotateTypeExpression(members[0], node)
		}
		if len(members) > 1 {
			return annotateTypeExpression(ast.NewUnionTypeExpression(members), node)
		}
	case "type_identifier":
		if child := firstNamedChild(node); child != nil && child != node {
			expr := parseTypeExpression(child, source)
			if expr == nil {
				return nil
			}
			return annotateTypeExpression(expr, node)
		}
	case "identifier":
		id, err := parseIdentifier(node, source)
		if err != nil {
			return nil
		}
		return annotateTypeExpression(ast.TyID(id), node)
	case "qualified_identifier":
		parts, err := parseQualifiedIdentifier(node, source)
		if err != nil || len(parts) == 0 {
			return nil
		}
		identifier := collapseQualifiedIdentifier(parts)
		if identifier == nil {
			return nil
		}
		return annotateTypeExpression(ast.TyID(identifier), node)
	default:
		if child := firstNamedChild(node); child != nil && child != node {
			if expr := parseTypeExpression(child, source); expr != nil {
				return annotateTypeExpression(expr, node)
			}
		}
	}
	text := strings.TrimSpace(sliceContent(node, source))
	if text == "" {
		return nil
	}
	return annotateTypeExpression(ast.Ty(strings.ReplaceAll(text, " ", "")), node)
}

func identifiersToStrings(ids []*ast.Identifier) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.Name
	}
	return result
}

func parseFunctionParameterTypes(node *sitter.Node, source []byte) ([]ast.TypeExpression, bool) {
	if node == nil {
		return nil, false
	}

	current := node
	for {
		if current == nil {
			return nil, false
		}
		if current.Kind() == "parenthesized_type" {
			if current.NamedChildCount() == 0 {
				return nil, false
			}
			params := make([]ast.TypeExpression, 0, current.NamedChildCount())
			for i := uint(0); i < current.NamedChildCount(); i++ {
				child := current.NamedChild(i)
				if child == nil {
					return nil, false
				}
				param := parseTypeExpression(child, source)
				if param == nil {
					return nil, false
				}
				params = append(params, param)
			}
			return params, true
		}
		if current.Kind() != "type_suffix" && current.Kind() != "type_prefix" && current.Kind() != "type_atom" {
			break
		}
		if child := firstNamedChild(current); child != nil && child != current {
			current = child
			continue
		}
		break
	}

	param := parseTypeExpression(node, source)
	if param == nil {
		return nil, false
	}
	return []ast.TypeExpression{param}, true
}

func parseTypeParameters(node *sitter.Node, source []byte) ([]*ast.GenericParameter, error) {
	if node == nil {
		return nil, nil
	}
	switch node.Kind() {
	case "declaration_type_parameters":
		if node.NamedChildCount() == 0 {
			return nil, nil
		}
		return parseTypeParameters(node.NamedChild(0), source)
	case "type_parameter_list":
		var params []*ast.GenericParameter
		for i := uint(0); i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child == nil || child.Kind() != "type_parameter" {
				continue
			}
			param, err := parseTypeParameter(child, source)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
		return params, nil
	case "generic_parameter_list":
		var params []*ast.GenericParameter
		for i := uint(0); i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child == nil || child.Kind() != "generic_parameter" {
				continue
			}
			param, err := parseGenericParameter(child, source)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
		return params, nil
	default:
		return nil, fmt.Errorf("parser: unsupported type parameter node %q", node.Kind())
	}
}

func parseTypeParameter(node *sitter.Node, source []byte) (*ast.GenericParameter, error) {
	if node == nil || node.Kind() != "type_parameter" {
		return nil, fmt.Errorf("parser: expected type_parameter node")
	}
	return buildGenericParameter(node, source)
}

func parseGenericParameter(node *sitter.Node, source []byte) (*ast.GenericParameter, error) {
	if node == nil || node.Kind() != "generic_parameter" {
		return nil, fmt.Errorf("parser: expected generic_parameter node")
	}
	return buildGenericParameter(node, source)
}

func buildGenericParameter(node *sitter.Node, source []byte) (*ast.GenericParameter, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil generic parameter")
	}
	var nameNode *sitter.Node
	if node.NamedChildCount() > 0 {
		nameNode = node.NamedChild(0)
	}
	if nameNode == nil || nameNode.Kind() != "identifier" {
		return nil, fmt.Errorf("parser: generic parameter missing identifier")
	}
	name, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	var constraints []*ast.InterfaceConstraint
	if node.NamedChildCount() > 1 {
		boundsNode := node.NamedChild(1)
		typeExprs, err := parseTypeBoundList(boundsNode, source)
		if err != nil {
			return nil, err
		}
		for _, expr := range typeExprs {
			constraints = append(constraints, ast.NewInterfaceConstraint(expr))
		}
	}

	param := ast.NewGenericParameter(name, constraints)
	annotateSpan(param, node)
	return param, nil
}

func parseTypeBoundList(node *sitter.Node, source []byte) ([]ast.TypeExpression, error) {
	if node == nil {
		return nil, nil
	}
	var bounds []ast.TypeExpression
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Kind() == "type_bound_list" {
			nested, err := parseTypeBoundList(child, source)
			if err != nil {
				return nil, err
			}
			bounds = append(bounds, nested...)
			continue
		}
		if expr := parseReturnType(child, source); expr != nil {
			bounds = append(bounds, expr)
		}
	}
	if len(bounds) == 0 {
		return nil, fmt.Errorf("parser: empty type bound list")
	}
	return bounds, nil
}

func parseWhereClause(node *sitter.Node, source []byte) ([]*ast.WhereClauseConstraint, error) {
	if node == nil {
		return nil, nil
	}
	if node.Kind() != "where_clause" {
		return nil, fmt.Errorf("parser: expected where clause, found %q", node.Kind())
	}
	var constraints []*ast.WhereClauseConstraint
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() != "where_constraint" {
			continue
		}
		constraint, err := parseWhereConstraint(child, source)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

func parseWhereConstraint(node *sitter.Node, source []byte) (*ast.WhereClauseConstraint, error) {
	if node == nil || node.Kind() != "where_constraint" {
		return nil, fmt.Errorf("parser: expected where_constraint node")
	}
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty where constraint")
	}
	nameNode := node.NamedChild(0)
	if nameNode == nil || nameNode.Kind() != "identifier" {
		return nil, fmt.Errorf("parser: where constraint missing identifier")
	}
	name, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}
	var constraintNode *sitter.Node
	if node.NamedChildCount() > 1 {
		constraintNode = node.NamedChild(1)
	}
	typeExprs, err := parseTypeBoundList(constraintNode, source)
	if err != nil {
		return nil, err
	}
	if len(typeExprs) == 0 {
		return nil, fmt.Errorf("parser: where constraint missing bounds")
	}
	interfaceConstraints := make([]*ast.InterfaceConstraint, 0, len(typeExprs))
	for _, expr := range typeExprs {
		interfaceConstraints = append(interfaceConstraints, ast.NewInterfaceConstraint(expr))
	}
	constraint := ast.NewWhereClauseConstraint(name, interfaceConstraints)
	annotateSpan(constraint, node)
	return constraint, nil
}
