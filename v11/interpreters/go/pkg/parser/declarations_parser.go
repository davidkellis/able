package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
)

func (ctx *parseContext) parseFunctionDefinition(node *sitter.Node) (*ast.FunctionDefinition, error) {
	if node == nil || node.Kind() != "function_definition" {
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
	name, err := parseIdentifier(node.ChildByFieldName("name"), source)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	params, err := ctx.parseParameterList(node.ChildByFieldName("parameters"))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, false, false, err
	}

	bodyNode := node.ChildByFieldName("body")
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
		if child.Kind() == "private" {
			isPrivate = true
			break
		}
	}

	returnType := ctx.parseReturnType(node.ChildByFieldName("return_type"))
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
	if node == nil || node.Kind() != "struct_definition" {
		return nil, fmt.Errorf("parser: expected struct_definition node")
	}

	source := ctx.source
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
			if fieldNode == nil || isIgnorableNode(fieldNode) || fieldNode.Kind() != "struct_field" {
				continue
			}
			field, err := ctx.parseStructFieldDefinition(fieldNode)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
	} else if tupleNode := node.ChildByFieldName("tuple"); tupleNode != nil {
		kind = ast.StructKindPositional
		for i := uint(0); i < tupleNode.NamedChildCount(); i++ {
			child := tupleNode.NamedChild(i)
			if child == nil || !child.IsNamed() || isIgnorableNode(child) {
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

	if id != nil && ctx != nil && ctx.structKinds != nil {
		ctx.structKinds[id.Name] = kind
	}

	structDef := ast.NewStructDefinition(id, fields, kind, generics, whereClause, isPrivate)
	annotateSpan(structDef, node)
	return structDef, nil
}

func (ctx *parseContext) parseTypeAliasDefinition(node *sitter.Node) (ast.Statement, error) {
	if node == nil || node.Kind() != "type_alias_definition" {
		return nil, fmt.Errorf("parser: expected type_alias_definition node")
	}

	source := ctx.source
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

	targetNode := node.ChildByFieldName("target")
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
	if node == nil || node.Kind() != "struct_field" {
		return nil, fmt.Errorf("parser: expected struct_field node")
	}

	var name *ast.Identifier
	var fieldType ast.TypeExpression

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		switch child.Kind() {
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
	if node == nil || node.Kind() != "methods_definition" {
		return nil, fmt.Errorf("parser: expected methods_definition node")
	}

	source := ctx.source
	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, err
	}

	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), source)
	if err != nil {
		return nil, err
	}

	targetType := ctx.parseTypeExpression(node.ChildByFieldName("target"))
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
			fn, err := ctx.parseFunctionDefinition(child)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, fn)
		case "method_member":
			for j := uint(0); j < child.NamedChildCount(); j++ {
				memberChild := child.NamedChild(j)
				if memberChild == nil || memberChild.Kind() != "function_definition" {
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
	if node == nil || node.Kind() != "implementation_definition" {
		return nil, fmt.Errorf("parser: expected implementation_definition node")
	}

	source := ctx.source
	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), source)
	if err != nil {
		return nil, err
	}

	interfaceNode := node.ChildByFieldName("interface")
	parts, err := ctx.parseQualifiedIdentifier(interfaceNode)
	if err != nil || len(parts) == 0 {
		return nil, fmt.Errorf("parser: invalid interface identifier")
	}
	interfaceName := collapseQualifiedIdentifier(parts)

	interfaceArgs, err := ctx.parseInterfaceArguments(node.ChildByFieldName("interface_args"))
	if err != nil {
		return nil, err
	}

	targetType := ctx.parseTypeExpression(node.ChildByFieldName("target"))
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
			fn, err := ctx.parseFunctionDefinition(child)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, fn)
		case "method_member":
			for j := uint(0); j < child.NamedChildCount(); j++ {
				memberChild := child.NamedChild(j)
				if memberChild == nil || memberChild.Kind() != "function_definition" {
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
	if node == nil || node.Kind() != "named_implementation_definition" {
		return nil, fmt.Errorf("parser: expected named implementation node")
	}
	nameNode := node.ChildByFieldName("name")
	implNode := node.ChildByFieldName("implementation")
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
			return nil, fmt.Errorf("parser: unsupported interface argument kind %q", child.Kind())
		}
		args = append(args, expr)
	}
	return args, nil
}

func findTopLevelGenericApplication(node *sitter.Node) *sitter.Node {
	current := node
	for current != nil {
		if current.Kind() == "type_generic_application" {
			return current
		}
		if current.Kind() == "parenthesized_type" {
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
	if node == nil || node.Kind() != "parameter" {
		return nil, fmt.Errorf("parser: expected parameter node")
	}

	patternNode := node.ChildByFieldName("pattern")
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
	if node == nil || node.Kind() != "union_definition" {
		return nil, fmt.Errorf("parser: expected union_definition node")
	}

	source := ctx.source
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
	if node == nil || node.Kind() != "interface_definition" {
		return nil, fmt.Errorf("parser: expected interface_definition node")
	}

	source := ctx.source
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
		selfType = ctx.parseTypeExpression(selfNode)
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
		signature, err := ctx.parseFunctionSignature(sigNode)
		if err != nil {
			return nil, err
		}
		if defaultBody := child.ChildByFieldName("default_body"); defaultBody != nil {
			body, err := ctx.parseBlock(defaultBody)
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

	iface := ast.NewInterfaceDefinition(name, signatures, typeParams, selfType, whereClause, baseInterfaces, hasLeadingPrivate(node))
	annotateStatement(iface, node)
	return iface, nil
}

func (ctx *parseContext) parseFunctionSignature(node *sitter.Node) (*ast.FunctionSignature, error) {
	if node == nil || node.Kind() != "function_signature" {
		return nil, fmt.Errorf("parser: expected function_signature node")
	}

	name, err := parseIdentifier(node.ChildByFieldName("name"), ctx.source)
	if err != nil {
		return nil, err
	}
	params, err := ctx.parseParameterList(node.ChildByFieldName("parameters"))
	if err != nil {
		return nil, err
	}
	returnType := parseReturnType(node.ChildByFieldName("return_type"), ctx.source)
	generics, err := parseTypeParameters(node.ChildByFieldName("type_parameters"), ctx.source)
	if err != nil {
		return nil, err
	}
	whereClause, err := parseWhereClause(node.ChildByFieldName("where_clause"), ctx.source)
	if err != nil {
		return nil, err
	}

	signature := ast.NewFunctionSignature(name, params, returnType, generics, whereClause, nil)
	annotateSpan(signature, node)
	return signature, nil
}

func (ctx *parseContext) parsePreludeStatement(node *sitter.Node) (ast.Statement, error) {
	if node == nil || node.Kind() != "prelude_statement" {
		return nil, fmt.Errorf("parser: expected prelude_statement node")
	}

	target, err := ctx.parseHostTarget(node.ChildByFieldName("target"))
	if err != nil {
		return nil, err
	}

	code, err := ctx.parseHostCodeBlock(node.ChildByFieldName("body"))
	if err != nil {
		return nil, err
	}

	stmt := ast.NewPreludeStatement(target, code)
	annotateStatement(stmt, node)
	return stmt, nil
}

func (ctx *parseContext) parseExternFunction(node *sitter.Node) (ast.Statement, error) {
	if node == nil || node.Kind() != "extern_function" {
		return nil, fmt.Errorf("parser: expected extern_function node")
	}

	target, err := ctx.parseHostTarget(node.ChildByFieldName("target"))
	if err != nil {
		return nil, err
	}

	signatureNode := node.ChildByFieldName("signature")
	if signatureNode == nil {
		return nil, fmt.Errorf("parser: extern function missing signature")
	}

	signature, err := ctx.parseFunctionSignature(signatureNode)
	if err != nil {
		return nil, err
	}

	body, err := ctx.parseHostCodeBlock(node.ChildByFieldName("body"))
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
	if node == nil || node.Kind() != "host_code_block" {
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
