package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func parseFunctionDefinition(node *sitter.Node, source []byte) (*ast.FunctionDefinition, error) {
	if node == nil || node.Kind() != "function_definition" {
		return nil, fmt.Errorf("parser: expected function_definition node")
	}

	name, generics, params, returnType, whereClause, body, isMethodShorthand, isPrivate, err := parseFunctionCore(node, source)
	if err != nil {
		return nil, err
	}

	fn := ast.NewFunctionDefinition(
		name,
		params,
		body,
		returnType,
		generics,
		whereClause,
		isMethodShorthand,
		isPrivate,
	)

	return fn, nil
}

func parseFunctionCore(node *sitter.Node, source []byte) (*ast.Identifier, []*ast.GenericParameter, []*ast.FunctionParameter, ast.TypeExpression, []*ast.WhereClauseConstraint, *ast.BlockExpression, bool, bool, error) {
	name, err := parseIdentifier(node.ChildByFieldName("name"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	params, err := parseParameterList(node.ChildByFieldName("parameters"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	bodyNode := node.ChildByFieldName("body")
	fnBody, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	isPrivate := false
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "private" {
			isPrivate = true
			break
		}
	}

	returnType := parseReturnType(node.ChildByFieldName("return_type"), source)
	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	methodShorthand := node.ChildByFieldName("method_shorthand") != nil

	return name, generics, params, returnType, whereClause, fnBody, methodShorthand, isPrivate, nil
}

func parseParameterList(node *sitter.Node, source []byte) ([]*ast.FunctionParameter, error) {
	if node == nil {
		return nil, nil
	}

	count := node.NamedChildCount()
	if count == 0 {
		return nil, nil
	}

	params := make([]*ast.FunctionParameter, 0, count)
	for i := uint(0); i < count; i++ {
		paramNode := node.NamedChild(i)
		param, err := parseParameter(paramNode, source)
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}
	return params, nil
}

func parseStructDefinition(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "struct_definition" {
		return nil, fmt.Errorf("parser: expected struct_definition node")
	}

	nameNode := node.ChildByFieldName("name")
	id, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, err
	}

	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), source)
	if err != nil {
		return nil, err
	}

	isPrivate := hasLeadingPrivate(node)

	kind := ast.StructKindSingleton
	fields := make([]*ast.StructFieldDefinition, 0)

	if recordNode := node.ChildByFieldName("record"); recordNode != nil {
		kind = ast.StructKindNamed
		for i := uint(0); i < recordNode.NamedChildCount(); i++ {
			fieldNode := recordNode.NamedChild(i)
			if fieldNode == nil || fieldNode.Kind() != "struct_field" {
				continue
			}
			field, err := parseStructFieldDefinition(fieldNode, source)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
	} else if tupleNode := node.ChildByFieldName("tuple"); tupleNode != nil {
		kind = ast.StructKindPositional
		for i := uint(0); i < tupleNode.NamedChildCount(); i++ {
			child := tupleNode.NamedChild(i)
			if child == nil || !child.IsNamed() {
				continue
			}
			fieldType := parseTypeExpression(child, source)
			if fieldType == nil {
				return nil, fmt.Errorf("parser: unsupported tuple field type")
			}
			fields = append(fields, ast.NewStructFieldDefinition(fieldType, nil))
		}
	}

	return ast.NewStructDefinition(id, fields, kind, generics, whereClause, isPrivate), nil
}

func parseStructFieldDefinition(node *sitter.Node, source []byte) (*ast.StructFieldDefinition, error) {
	if node == nil || node.Kind() != "struct_field" {
		return nil, fmt.Errorf("parser: expected struct_field node")
	}

	var name *ast.Identifier
	var fieldType ast.TypeExpression

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "identifier":
			if name == nil {
				id, err := parseIdentifier(child, source)
				if err != nil {
					return nil, err
				}
				name = id
			}
		default:
			if fieldType == nil {
				fieldType = parseTypeExpression(child, source)
			}
		}
	}

	if fieldType == nil {
		return nil, fmt.Errorf("parser: struct field missing type")
	}

	return ast.NewStructFieldDefinition(fieldType, name), nil
}

func parseMethodsDefinition(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "methods_definition" {
		return nil, fmt.Errorf("parser: expected methods_definition node")
	}

	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, err
	}

	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), source)
	if err != nil {
		return nil, err
	}

	targetType := parseTypeExpression(node.ChildByFieldName("target"), source)
	if targetType == nil {
		return nil, fmt.Errorf("parser: methods definition missing target type")
	}

	definitions := make([]*ast.FunctionDefinition, 0)

	targetNode := node.ChildByFieldName("target")
	typeParamsNode := node.ChildByFieldName("type_parameters")
	whereNode := node.ChildByFieldName("where_clause")

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		fieldName := node.FieldNameForChild(uint32(i))
		if (fieldName == "target" || fieldName == "type_parameters" || fieldName == "where_clause") && child.Kind() != "function_definition" && child.Kind() != "method_member" {
			continue
		}
		if sameNode(child, targetNode) || sameNode(child, typeParamsNode) || sameNode(child, whereNode) {
			continue
		}
		switch child.Kind() {
		case "function_definition":
			fn, err := parseFunctionDefinition(child, source)
			if err != nil {
				return nil, err
			}
			fn.IsMethodShorthand = true
			definitions = append(definitions, fn)
		case "method_member":
			for j := uint(0); j < child.NamedChildCount(); j++ {
				memberChild := child.NamedChild(j)
				if memberChild == nil || memberChild.Kind() != "function_definition" {
					continue
				}
				fn, err := parseFunctionDefinition(memberChild, source)
				if err != nil {
					return nil, err
				}
				fn.IsMethodShorthand = true
				definitions = append(definitions, fn)
			}
		}
	}

	return ast.NewMethodsDefinition(targetType, definitions, generics, whereClause), nil
}

func parseImplementationDefinitionNode(node *sitter.Node, source []byte) (*ast.ImplementationDefinition, error) {
	if node == nil || node.Kind() != "implementation_definition" {
		return nil, fmt.Errorf("parser: expected implementation_definition node")
	}

	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, err
	}

	interfaceNode := node.ChildByFieldName("interface")
	parts, err := parseQualifiedIdentifier(interfaceNode, source)
	if err != nil || len(parts) == 0 {
		return nil, fmt.Errorf("parser: invalid interface identifier")
	}
	var interfaceName *ast.Identifier
	if len(parts) == 1 {
		interfaceName = parts[0]
	} else {
		interfaceName = ast.ID(strings.Join(identifiersToStrings(parts), "."))
	}

	interfaceArgs, err := parseInterfaceArguments(node.ChildByFieldName("interface_args"), source)
	if err != nil {
		return nil, err
	}

	targetType := parseTypeExpression(node.ChildByFieldName("target"), source)
	if targetType == nil {
		return nil, fmt.Errorf("parser: implementation missing target type")
	}

	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), source)
	if err != nil {
		return nil, err
	}

	definitions := make([]*ast.FunctionDefinition, 0)

	interfaceArgsNode := node.ChildByFieldName("interface_args")
	targetNode := node.ChildByFieldName("target")
	typeParamsNode := node.ChildByFieldName("type_parameters")
	whereNode := node.ChildByFieldName("where_clause")

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		fieldName := node.FieldNameForChild(uint32(i))
		if (fieldName == "interface" || fieldName == "interface_args" || fieldName == "target" || fieldName == "type_parameters" || fieldName == "where_clause") && child.Kind() != "function_definition" && child.Kind() != "method_member" {
			continue
		}
		if sameNode(child, interfaceNode) || sameNode(child, interfaceArgsNode) || sameNode(child, targetNode) || sameNode(child, typeParamsNode) || sameNode(child, whereNode) {
			continue
		}
		switch child.Kind() {
		case "function_definition":
			fn, err := parseFunctionDefinition(child, source)
			if err != nil {
				return nil, err
			}
			fn.IsMethodShorthand = true
			definitions = append(definitions, fn)
		case "method_member":
			for j := uint(0); j < child.NamedChildCount(); j++ {
				memberChild := child.NamedChild(j)
				if memberChild == nil || memberChild.Kind() != "function_definition" {
					continue
				}
				fn, err := parseFunctionDefinition(memberChild, source)
				if err != nil {
					return nil, err
				}
				fn.IsMethodShorthand = true
				definitions = append(definitions, fn)
			}
		}
	}

	impl := ast.NewImplementationDefinition(interfaceName, targetType, definitions, nil, generics, interfaceArgs, whereClause, hasLeadingPrivate(node))
	return impl, nil
}

func parseImplementationDefinition(node *sitter.Node, source []byte) (ast.Statement, error) {
	return parseImplementationDefinitionNode(node, source)
}

func parseNamedImplementationDefinition(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "named_implementation_definition" {
		return nil, fmt.Errorf("parser: expected named implementation node")
	}
	nameNode := node.ChildByFieldName("name")
	implNode := node.ChildByFieldName("implementation")
	if implNode == nil {
		return nil, fmt.Errorf("parser: named implementation missing implementation body")
	}
	impl, err := parseImplementationDefinitionNode(implNode, source)
	if err != nil {
		return nil, err
	}
	if nameNode != nil {
		name, err := parseIdentifier(nameNode, source)
		if err != nil {
			return nil, err
		}
		impl.ImplName = name
	}
	return impl, nil
}

func parseInterfaceArguments(node *sitter.Node, source []byte) ([]ast.TypeExpression, error) {
	if node == nil {
		return nil, nil
	}
	var args []ast.TypeExpression
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		expr := parseTypeExpression(child, source)
		if expr == nil {
			return nil, fmt.Errorf("parser: unsupported interface argument kind %q", child.Kind())
		}
		args = append(args, expr)
	}
	return args, nil
}

func parseParameter(node *sitter.Node, source []byte) (*ast.FunctionParameter, error) {
	if node == nil || node.Kind() != "parameter" {
		return nil, fmt.Errorf("parser: expected parameter node")
	}

	patternNode := node.ChildByFieldName("pattern")
	pattern, err := parsePattern(patternNode, source)
	if err != nil {
		return nil, err
	}

	var paramType ast.TypeExpression
	if typed, ok := pattern.(*ast.TypedPattern); ok {
		pattern = typed.Pattern
		paramType = typed.TypeAnnotation
	}

	return ast.NewFunctionParameter(pattern, paramType), nil
}

func parseUnionDefinition(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "union_definition" {
		return nil, fmt.Errorf("parser: expected union_definition node")
	}

	nameNode := node.ChildByFieldName("name")
	name, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	typeParamsNode := node.ChildByFieldName("type_parameters")
	typeParams, err := parseTypeParameters(typeParamsNode, source)
	if err != nil {
		return nil, err
	}

	variants := make([]ast.TypeExpression, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if sameNode(child, nameNode) || sameNode(child, typeParamsNode) {
			continue
		}
		variant := parseTypeExpression(child, source)
		if variant == nil {
			return nil, fmt.Errorf("parser: invalid union variant")
		}
		variants = append(variants, variant)
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("parser: union definition requires variants")
	}

	return ast.NewUnionDefinition(name, variants, typeParams, nil, hasLeadingPrivate(node)), nil
}

func parseInterfaceDefinition(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "interface_definition" {
		return nil, fmt.Errorf("parser: expected interface_definition node")
	}

	nameNode := node.ChildByFieldName("name")
	name, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	typeParamsNode := node.ChildByFieldName("type_parameters")
	typeParams, err := parseTypeParameters(typeParamsNode, source)
	if err != nil {
		return nil, err
	}

	var selfType ast.TypeExpression
	selfNode := node.ChildByFieldName("self_type")
	if selfNode != nil {
		selfType = parseTypeExpression(selfNode, source)
	}

	whereNode := node.ChildByFieldName("where_clause")
	whereClause, err := parseWhereClause(whereNode, source)
	if err != nil {
		return nil, err
	}

	compositeNode := node.ChildByFieldName("composite")
	signatures := make([]*ast.FunctionSignature, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if sameNode(child, nameNode) || sameNode(child, typeParamsNode) || sameNode(child, selfNode) || sameNode(child, whereNode) || sameNode(child, compositeNode) {
			continue
		}
		if child.Kind() != "interface_member" {
			continue
		}
		sigNode := child.ChildByFieldName("signature")
		if sigNode == nil {
			return nil, fmt.Errorf("parser: interface member missing signature")
		}
		signature, err := parseFunctionSignature(sigNode, source)
		if err != nil {
			return nil, err
		}
		if defaultBody := child.ChildByFieldName("default_body"); defaultBody != nil {
			body, err := parseBlock(defaultBody, source)
			if err != nil {
				return nil, err
			}
			signature.DefaultImpl = body
		}
		signatures = append(signatures, signature)
	}

	var baseInterfaces []ast.TypeExpression
	if compositeNode != nil {
		bounds, err := parseTypeBoundList(compositeNode, source)
		if err != nil {
			return nil, err
		}
		baseInterfaces = append(baseInterfaces, bounds...)
	}

	return ast.NewInterfaceDefinition(name, signatures, typeParams, selfType, whereClause, baseInterfaces, hasLeadingPrivate(node)), nil
}

func parseFunctionSignature(node *sitter.Node, source []byte) (*ast.FunctionSignature, error) {
	if node == nil || node.Kind() != "function_signature" {
		return nil, fmt.Errorf("parser: expected function_signature node")
	}

	name, err := parseIdentifier(node.ChildByFieldName("name"), source)
	if err != nil {
		return nil, err
	}
	params, err := parseParameterList(node.ChildByFieldName("parameters"), source)
	if err != nil {
		return nil, err
	}
	returnType := parseReturnType(node.ChildByFieldName("return_type"), source)
	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, err
	}
	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), source)
	if err != nil {
		return nil, err
	}

	return ast.NewFunctionSignature(name, params, returnType, generics, whereClause, nil), nil
}

func parsePreludeStatement(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "prelude_statement" {
		return nil, fmt.Errorf("parser: expected prelude_statement node")
	}

	target, err := parseHostTarget(node.ChildByFieldName("target"), source)
	if err != nil {
		return nil, err
	}

	code, err := parseHostCodeBlock(node.ChildByFieldName("body"), source)
	if err != nil {
		return nil, err
	}

	return ast.NewPreludeStatement(target, code), nil
}

func parseExternFunction(node *sitter.Node, source []byte) (ast.Statement, error) {
	if node == nil || node.Kind() != "extern_function" {
		return nil, fmt.Errorf("parser: expected extern_function node")
	}

	target, err := parseHostTarget(node.ChildByFieldName("target"), source)
	if err != nil {
		return nil, err
	}

	signatureNode := node.ChildByFieldName("signature")
	if signatureNode == nil {
		return nil, fmt.Errorf("parser: extern function missing signature")
	}

	signature, err := parseFunctionSignature(signatureNode, source)
	if err != nil {
		return nil, err
	}

	body, err := parseHostCodeBlock(node.ChildByFieldName("body"), source)
	if err != nil {
		return nil, err
	}

	fn := ast.NewFunctionDefinition(
		signature.Name,
		signature.Params,
		nil,
		signature.ReturnType,
		signature.GenericParams,
		signature.WhereClause,
		false,
		false,
	)

	return ast.NewExternFunctionBody(target, fn, body), nil
}

func parseHostTarget(node *sitter.Node, source []byte) (ast.HostTarget, error) {
	if node == nil {
		return "", fmt.Errorf("parser: missing host target")
	}
	switch strings.TrimSpace(sliceContent(node, source)) {
	case "go":
		return ast.HostTargetGo, nil
	case "crystal":
		return ast.HostTargetCrystal, nil
	case "typescript":
		return ast.HostTargetTypeScript, nil
	case "python":
		return ast.HostTargetPython, nil
	case "ruby":
		return ast.HostTargetRuby, nil
	default:
		return "", fmt.Errorf("parser: unsupported host target")
	}
}

func parseHostCodeBlock(node *sitter.Node, source []byte) (string, error) {
	if node == nil || node.Kind() != "host_code_block" {
		return "", fmt.Errorf("parser: expected host_code_block node")
	}

	start := int(node.StartByte())
	end := int(node.EndByte())
	if start < 0 || end > len(source) || start >= end {
		return "", fmt.Errorf("parser: invalid host code block range")
	}

	content := strings.TrimSpace(string(source[start+1 : end-1]))
	return content, nil
}
