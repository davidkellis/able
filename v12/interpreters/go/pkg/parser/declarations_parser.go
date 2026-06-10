package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
)

func (ctx *parseContext) parseFunctionDefinition(node *sitter.Node) (*ast.FunctionDefinition, error) {
	if node == nil || nodeKind(node) != "function_definition" {
		return nil, fmt.Errorf("parser: expected function_definition node")
	}

	name, generics, params, returnType, whereClause, body, isMethodShorthand, isPrivate, err := ctx.parseFunctionCore(node)
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

	annotateSpan(fn, node)
	return fn, nil
}

func (ctx *parseContext) parseFunctionCore(node *sitter.Node) (*ast.Identifier, []*ast.GenericParameter, []*ast.FunctionParameter, ast.TypeExpression, []*ast.WhereClauseConstraint, *ast.BlockExpression, bool, bool, error) {
	source := ctx.source
	name, err := parseIdentifier(childByFieldName(node, "name"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	params, err := ctx.parseParameterList(childByFieldName(node, "parameters"))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	bodyNode := childByFieldName(node, "body")
	fnBody, err := ctx.parseBlock(bodyNode)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	isPrivate := false
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		if nodeKind(child) == "private" {
			isPrivate = true
			break
		}
	}

	returnType := ctx.parseReturnType(childByFieldName(node, "return_type"))
	generics, err := parseTypeParameters(childByFieldName(node, "type_parameters"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	whereClause, err := parseWhereClause(childByFieldName(node, "where_clause"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	methodShorthand := childByFieldName(node, "method_shorthand") != nil

	return name, generics, params, returnType, whereClause, fnBody, methodShorthand, isPrivate, nil
}

func (ctx *parseContext) parseParameterList(node *sitter.Node) ([]*ast.FunctionParameter, error) {
	if node == nil {
		return make([]*ast.FunctionParameter, 0), nil
	}

	count := node.NamedChildCount()
	params := make([]*ast.FunctionParameter, 0, count)
	for i := uint(0); i < count; i++ {
		paramNode := node.NamedChild(i)
		param, err := ctx.parseParameter(paramNode)
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}
	return params, nil
}

func (ctx *parseContext) parseStructDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "struct_definition" {
		return nil, fmt.Errorf("parser: expected struct_definition node")
	}

	source := ctx.source
	nameNode := childByFieldName(node, "name")
	id, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	generics, err := parseTypeParameters(childByFieldName(node, "type_parameters"), source)
	if err != nil {
		return nil, err
	}

	whereClause, err := parseWhereClause(childByFieldName(node, "where_clause"), source)
	if err != nil {
		return nil, err
	}

	isPrivate := hasLeadingPrivate(node)

	kind := ast.StructKindSingleton
	fields := make([]*ast.StructFieldDefinition, 0)

	if recordNode := childByFieldName(node, "record"); recordNode != nil {
		kind = ast.StructKindNamed
		for i := uint(0); i < recordNode.NamedChildCount(); i++ {
			fieldNode := recordNode.NamedChild(i)
			if fieldNode == nil || isIgnorableNode(fieldNode) || nodeKind(fieldNode) != "struct_field" {
				continue
			}
			field, err := ctx.parseStructFieldDefinition(fieldNode)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
	} else if tupleNode := childByFieldName(node, "tuple"); tupleNode != nil {
		kind = ast.StructKindPositional
		for i := uint(0); i < tupleNode.NamedChildCount(); i++ {
			child := tupleNode.NamedChild(i)
			if child == nil || !child.IsNamed() || isIgnorableNode(child) {
				continue
			}
			if nodeKind(child) == "ERROR" && strings.TrimSpace(sliceContent(child, ctx.source)) == "" {
				continue
			}
			fieldType := ctx.parseTypeExpression(child)
			if fieldType == nil {
				return nil, fmt.Errorf("parser: unsupported tuple field type")
			}
			field := ast.NewStructFieldDefinition(fieldType, nil)
			annotateSpan(field, child)
			fields = append(fields, field)
		}
	}
	if len(fields) == 0 {
		kind = ast.StructKindSingleton
	}

	if id != nil && ctx != nil && ctx.structKinds != nil {
		ctx.structKinds[id.Name] = kind
	}

	structDef := ast.NewStructDefinition(id, fields, kind, generics, whereClause, isPrivate)
	annotateSpan(structDef, node)
	return structDef, nil
}

func (ctx *parseContext) parseTypeAliasDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "type_alias_definition" {
		return nil, fmt.Errorf("parser: expected type_alias_definition node")
	}

	source := ctx.source
	nameNode := childByFieldName(node, "name")
	id, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	generics, err := parseTypeParameters(childByFieldName(node, "type_parameters"), source)
	if err != nil {
		return nil, err
	}

	whereClause, err := parseWhereClause(childByFieldName(node, "where_clause"), source)
	if err != nil {
		return nil, err
	}

	targetNode := childByFieldName(node, "target")
	targetType := ctx.parseTypeExpression(targetNode)
	if targetType == nil {
		return nil, fmt.Errorf("parser: type alias missing target type")
	}

	isPrivate := hasLeadingPrivate(node)
	alias := ast.NewTypeAliasDefinition(id, targetType, generics, whereClause, isPrivate)
	annotateSpan(alias, node)
	return alias, nil
}

func (ctx *parseContext) parseStructFieldDefinition(node *sitter.Node) (*ast.StructFieldDefinition, error) {
	if node == nil || nodeKind(node) != "struct_field" {
		return nil, fmt.Errorf("parser: expected struct_field node")
	}

	var name *ast.Identifier
	var fieldType ast.TypeExpression

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		switch nodeKind(child) {
		case "identifier":
			if name == nil {
				id, err := parseIdentifier(child, ctx.source)
				if err != nil {
					return nil, err
				}
				name = id
			}
		default:
			if fieldType == nil {
				fieldType = ctx.parseTypeExpression(child)
			}
		}
	}

	if fieldType == nil {
		return nil, fmt.Errorf("parser: struct field missing type")
	}

	field := ast.NewStructFieldDefinition(fieldType, name)
	annotateSpan(field, node)
	return field, nil
}

func (ctx *parseContext) parseMethodsDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "methods_definition" {
		return nil, fmt.Errorf("parser: expected methods_definition node")
	}

	source := ctx.source
	generics, err := parseTypeParameters(childByFieldName(node, "type_parameters"), source)
	if err != nil {
		return nil, err
	}

	whereClause, err := parseWhereClause(childByFieldName(node, "where_clause"), source)
	if err != nil {
		return nil, err
	}

	targetType := ctx.parseTypeExpression(childByFieldName(node, "target"))
	if targetType == nil {
		return nil, fmt.Errorf("parser: methods definition missing target type")
	}

	definitions := make([]*ast.FunctionDefinition, 0)

	targetNode := childByFieldName(node, "target")
	typeParamsNode := childByFieldName(node, "type_parameters")
	whereNode := childByFieldName(node, "where_clause")

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		fieldName := node.FieldNameForChild(uint32(i))
		if (fieldName == "target" || fieldName == "type_parameters" || fieldName == "where_clause") && nodeKind(child) != "function_definition" && nodeKind(child) != "method_member" {
			continue
		}
		if sameNode(child, targetNode) || sameNode(child, typeParamsNode) || sameNode(child, whereNode) {
			continue
		}
		switch nodeKind(child) {
		case "function_definition":
			fn, err := ctx.parseFunctionDefinition(child)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, fn)
		case "method_member":
			for j := uint(0); j < child.NamedChildCount(); j++ {
				memberChild := child.NamedChild(j)
				if memberChild == nil || nodeKind(memberChild) != "function_definition" {
					continue
				}
				fn, err := ctx.parseFunctionDefinition(memberChild)
				if err != nil {
					return nil, err
				}
				definitions = append(definitions, fn)
			}
		}
	}

	methods := ast.NewMethodsDefinition(targetType, definitions, generics, whereClause)
	annotateSpan(methods, node)
	return methods, nil
}

func (ctx *parseContext) parseImplementationDefinitionNode(node *sitter.Node) (*ast.ImplementationDefinition, error) {
	if node == nil || nodeKind(node) != "implementation_definition" {
		return nil, fmt.Errorf("parser: expected implementation_definition node")
	}

	source := ctx.source
	generics, err := parseTypeParameters(childByFieldName(node, "type_parameters"), source)
	if err != nil {
		return nil, err
	}

	interfaceNode := childByFieldName(node, "interface")
	parts, err := ctx.parseQualifiedIdentifier(interfaceNode)
	if err != nil || len(parts) == 0 {
		return nil, fmt.Errorf("parser: invalid interface identifier")
	}
	interfaceName := collapseQualifiedIdentifier(parts)

	interfaceArgs, err := ctx.parseInterfaceArguments(childByFieldName(node, "interface_args"))
	if err != nil {
		return nil, err
	}

	targetType := ctx.parseTypeExpression(childByFieldName(node, "target"))
	if targetType == nil {
		return nil, fmt.Errorf("parser: implementation missing target type")
	}

	whereClause, err := parseWhereClause(childByFieldName(node, "where_clause"), source)
	if err != nil {
		return nil, err
	}

	definitions := make([]*ast.FunctionDefinition, 0)

	interfaceArgsNode := childByFieldName(node, "interface_args")
	targetNode := childByFieldName(node, "target")
	typeParamsNode := childByFieldName(node, "type_parameters")
	whereNode := childByFieldName(node, "where_clause")

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		fieldName := node.FieldNameForChild(uint32(i))
		if (fieldName == "interface" || fieldName == "interface_args" || fieldName == "target" || fieldName == "type_parameters" || fieldName == "where_clause") && nodeKind(child) != "function_definition" && nodeKind(child) != "method_member" {
			continue
		}
		if sameNode(child, interfaceNode) || sameNode(child, interfaceArgsNode) || sameNode(child, targetNode) || sameNode(child, typeParamsNode) || sameNode(child, whereNode) {
			continue
		}
		switch nodeKind(child) {
		case "function_definition":
			fn, err := ctx.parseFunctionDefinition(child)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, fn)
		case "method_member":
			for j := uint(0); j < child.NamedChildCount(); j++ {
				memberChild := child.NamedChild(j)
				if memberChild == nil || nodeKind(memberChild) != "function_definition" {
					continue
				}
				fn, err := ctx.parseFunctionDefinition(memberChild)
				if err != nil {
					return nil, err
				}
				definitions = append(definitions, fn)
			}
		}
	}

	impl := ast.NewImplementationDefinition(interfaceName, targetType, definitions, nil, generics, interfaceArgs, whereClause, hasLeadingPrivate(node))
	annotateSpan(impl, node)
	return impl, nil
}

func (ctx *parseContext) parseImplementationDefinition(node *sitter.Node) (ast.Statement, error) {
	return ctx.parseImplementationDefinitionNode(node)
}

func (ctx *parseContext) parseNamedImplementationDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "named_implementation_definition" {
		return nil, fmt.Errorf("parser: expected named implementation node")
	}
	nameNode := childByFieldName(node, "name")
	implNode := childByFieldName(node, "implementation")
	if implNode == nil {
		return nil, fmt.Errorf("parser: named implementation missing implementation body")
	}
	impl, err := ctx.parseImplementationDefinitionNode(implNode)
	if err != nil {
		return nil, err
	}
	if nameNode != nil {
		name, err := parseIdentifier(nameNode, ctx.source)
		if err != nil {
			return nil, err
		}
		impl.ImplName = name
	}
	annotateSpan(impl, node)
	return impl, nil
}

func (ctx *parseContext) parseInterfaceArguments(node *sitter.Node) ([]ast.TypeExpression, error) {
	if node == nil {
		return nil, nil
	}
	var args []ast.TypeExpression
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		genericNode := findTopLevelGenericApplication(child)
		if genericNode != nil {
			snippet := strings.TrimSpace(sliceContent(genericNode, ctx.source))
			detail := ""
			if snippet != "" {
				detail = fmt.Sprintf("; wrap %q in parentheses", snippet)
			}
			return nil, fmt.Errorf("parser: interface arguments require parenthesized generic applications%s", detail)
		}
		expr := ctx.parseTypeExpression(child)
		if expr == nil {
			return nil, fmt.Errorf("parser: unsupported interface argument kind %q", nodeKind(child))
		}
		args = append(args, expr)
	}
	return args, nil
}

func findTopLevelGenericApplication(node *sitter.Node) *sitter.Node {
	current := node
	for current != nil {
		if nodeKind(current) == "type_generic_application" {
			return current
		}
		if nodeKind(current) == "parenthesized_type" {
			return nil
		}
		if current.NamedChildCount() != 1 {
			return nil
		}
		child := current.NamedChild(0)
		if child == nil || !child.IsNamed() {
			return nil
		}
		current = child
	}
	return nil
}

func (ctx *parseContext) parseParameter(node *sitter.Node) (*ast.FunctionParameter, error) {
	if node == nil || nodeKind(node) != "parameter" {
		return nil, fmt.Errorf("parser: expected parameter node")
	}

	patternNode := childByFieldName(node, "pattern")
	pattern, err := ctx.parsePattern(patternNode)
	if err != nil {
		return nil, err
	}

	var paramType ast.TypeExpression
	if typed, ok := pattern.(*ast.TypedPattern); ok {
		pattern = typed.Pattern
		paramType = typed.TypeAnnotation
	}

	param := ast.NewFunctionParameter(pattern, paramType)
	annotateSpan(param, node)
	return param, nil
}

func (ctx *parseContext) parseUnionDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "union_definition" {
		return nil, fmt.Errorf("parser: expected union_definition node")
	}

	source := ctx.source
	nameNode := childByFieldName(node, "name")
	name, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	typeParamsNode := childByFieldName(node, "type_parameters")
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
		variant := ctx.parseTypeExpression(child)
		if variant == nil {
			return nil, fmt.Errorf("parser: invalid union variant")
		}
		variants = append(variants, variant)
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("parser: union definition requires variants")
	}

	union := ast.NewUnionDefinition(name, variants, typeParams, nil, hasLeadingPrivate(node))
	annotateStatement(union, node)
	return union, nil
}

func (ctx *parseContext) parseInterfaceDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "interface_definition" {
		return nil, fmt.Errorf("parser: expected interface_definition node")
	}

	source := ctx.source
	nameNode := childByFieldName(node, "name")
	name, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	typeParamsNode := childByFieldName(node, "type_parameters")
	typeParams, err := parseTypeParameters(typeParamsNode, source)
	if err != nil {
		return nil, err
	}

	whereNode := childByFieldName(node, "where_clause")
	whereClause, err := parseWhereClause(whereNode, source)
	if err != nil {
		return nil, err
	}

	selfNode := childByFieldName(node, "self_type")
	baseNode := childByFieldName(node, "base_interfaces")
	compositeNode := childByFieldName(node, "composite")
	var recoveredSelfType ast.TypeExpression
	if baseNode == nil && compositeNode == nil {
		if recoveredType, recoveredBase, ok := recoverInterfaceBaseSelfType(node, source); ok {
			recoveredSelfType = recoveredType
			selfNode = nil
			baseNode = recoveredBase
		}
	}
	if baseNode == nil {
		baseNode = compositeNode
	}
	signatures := make([]*ast.FunctionSignature, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if sameNode(child, nameNode) || sameNode(child, typeParamsNode) || sameNode(child, selfNode) || sameNode(child, whereNode) || sameNode(child, baseNode) {
			continue
		}
		if nodeKind(child) != "interface_member" {
			continue
		}
		sigNode := childByFieldName(child, "signature")
		if sigNode == nil {
			return nil, fmt.Errorf("parser: interface member missing signature")
		}
		signature, err := ctx.parseFunctionSignature(sigNode)
		if err != nil {
			return nil, err
		}
		if defaultBody := childByFieldName(child, "default_body"); defaultBody != nil {
			body, err := ctx.parseBlock(defaultBody)
			if err != nil {
				return nil, err
			}
			signature.DefaultImpl = body
		}
		signatures = append(signatures, signature)
	}

	var baseInterfaces []ast.TypeExpression
	if baseNode != nil {
		bounds, err := parseTypeBoundList(baseNode, source)
		if err != nil {
			return nil, err
		}
		baseInterfaces = append(baseInterfaces, bounds...)
	}

	var selfType ast.TypeExpression
	if recoveredSelfType != nil {
		selfType = recoveredSelfType
	} else if selfNode != nil {
		selfType = ctx.parseTypeExpression(selfNode)
	}

	iface := ast.NewInterfaceDefinition(name, signatures, typeParams, selfType, whereClause, baseInterfaces, hasLeadingPrivate(node))
	annotateStatement(iface, node)
	return iface, nil
}

func (ctx *parseContext) parseFunctionSignature(node *sitter.Node) (*ast.FunctionSignature, error) {
	if node == nil || nodeKind(node) != "function_signature" {
		return nil, fmt.Errorf("parser: expected function_signature node")
	}

	name, err := parseIdentifier(childByFieldName(node, "name"), ctx.source)
	if err != nil {
		return nil, err
	}
	params, err := ctx.parseParameterList(childByFieldName(node, "parameters"))
	if err != nil {
		return nil, err
	}
	returnType := parseReturnType(childByFieldName(node, "return_type"), ctx.source)
	generics, err := parseTypeParameters(childByFieldName(node, "type_parameters"), ctx.source)
	if err != nil {
		return nil, err
	}
	whereClause, err := parseWhereClause(childByFieldName(node, "where_clause"), ctx.source)
	if err != nil {
		return nil, err
	}

	signature := ast.NewFunctionSignature(name, params, returnType, generics, whereClause, nil)
	annotateSpan(signature, node)
	return signature, nil
}

func (ctx *parseContext) parsePreludeStatement(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "prelude_statement" {
		return nil, fmt.Errorf("parser: expected prelude_statement node")
	}

	target, err := ctx.parseHostTarget(childByFieldName(node, "target"))
	if err != nil {
		return nil, err
	}

	code, err := ctx.parseHostCodeBlock(childByFieldName(node, "body"))
	if err != nil {
		return nil, err
	}

	stmt := ast.NewPreludeStatement(target, code)
	annotateStatement(stmt, node)
	return stmt, nil
}

func (ctx *parseContext) parseExternFunction(node *sitter.Node) (ast.Statement, error) {
	if node == nil || nodeKind(node) != "extern_function" {
		return nil, fmt.Errorf("parser: expected extern_function node")
	}

	target, err := ctx.parseHostTarget(childByFieldName(node, "target"))
	if err != nil {
		return nil, err
	}

	signatureNode := childByFieldName(node, "signature")
	if signatureNode == nil {
		return nil, fmt.Errorf("parser: extern function missing signature")
	}

	signature, err := ctx.parseFunctionSignature(signatureNode)
	if err != nil {
		return nil, err
	}

	body, err := ctx.parseHostCodeBlock(childByFieldName(node, "body"))
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

	stmt := ast.NewExternFunctionBody(target, fn, body)
	annotateStatement(stmt, node)
	return stmt, nil
}

func (ctx *parseContext) parseHostTarget(node *sitter.Node) (ast.HostTarget, error) {
	if node == nil {
		return "", fmt.Errorf("parser: missing host target")
	}
	switch strings.TrimSpace(sliceContent(node, ctx.source)) {
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

func (ctx *parseContext) parseHostCodeBlock(node *sitter.Node) (string, error) {
	if node == nil || nodeKind(node) != "host_code_block" {
		return "", fmt.Errorf("parser: expected host_code_block node")
	}

	start := int(node.StartByte())
	end := int(node.EndByte())
	if start < 0 || end > len(ctx.source) || start >= end {
		return "", fmt.Errorf("parser: invalid host code block range")
	}

	content := strings.TrimSpace(string(ctx.source[start+1 : end-1]))
	return content, nil
}
