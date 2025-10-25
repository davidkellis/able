package parser

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/parser/language"
)

var infixOperatorSets = map[string][]string{
	"logical_or_expression":     {"||"},
	"logical_and_expression":    {"&&"},
	"bitwise_or_expression":     {"|"},
	"bitwise_xor_expression":    {"^"},
	"bitwise_and_expression":    {"&"},
	"equality_expression":       {"==", "!="},
	"comparison_expression":     {">", "<", ">=", "<="},
	"shift_expression":          {"<<", ">>"},
	"additive_expression":       {"+", "-"},
	"multiplicative_expression": {"*", "/", "%"},
	"exponent_expression":       {"**"},
}

// ModuleParser wraps a tree-sitter parser configured for Able v10 modules.
type ModuleParser struct {
	parser *sitter.Parser
}

// NewModuleParser constructs a parser with the Able language loaded.
func NewModuleParser() (*ModuleParser, error) {
	lang := language.Able()
	if lang == nil {
		return nil, fmt.Errorf("parser: able language not available")
	}

	p := sitter.NewParser()
	if err := p.SetLanguage(lang); err != nil {
		return nil, fmt.Errorf("parser: %w", err)
	}

	return &ModuleParser{parser: p}, nil
}

// Close releases parser resources.
func (p *ModuleParser) Close() {
	if p == nil || p.parser == nil {
		return
	}
	p.parser.Close()
}

// ParseModule parses Able source into the canonical AST module.
func (p *ModuleParser) ParseModule(source []byte) (*ast.Module, error) {
	if p == nil || p.parser == nil {
		return nil, fmt.Errorf("parser: nil parser")
	}

	tree := p.parser.Parse(source, nil)
	defer tree.Close()

	root := tree.RootNode()
	if root == nil || root.Kind() != "source_file" {
		return nil, fmt.Errorf("parser: unexpected root node")
	}
	if root.HasError() {
		return nil, fmt.Errorf("parser: syntax errors present")
	}

	var (
		modulePackage *ast.PackageStatement
		imports       = make([]*ast.ImportStatement, 0)
		body          = make([]ast.Statement, 0)
	)

	for i := uint(0); i < root.NamedChildCount(); i++ {
		node := root.NamedChild(i)
		switch node.Kind() {
		case "package_statement":
			pkg, err := parsePackageStatement(node, source)
			if err != nil {
				return nil, err
			}
			modulePackage = pkg
		case "import_statement":
			kindNode := node.ChildByFieldName("kind")
			if kindNode == nil {
				return nil, fmt.Errorf("parser: import missing kind")
			}

			path, err := parseQualifiedIdentifier(node.ChildByFieldName("path"), source)
			if err != nil {
				return nil, err
			}

			isWildcard, selectors, alias, err := parseImportClause(node.ChildByFieldName("clause"), source)
			if err != nil {
				return nil, err
			}

			switch kindNode.Kind() {
			case "import":
				imports = append(imports, ast.NewImportStatement(path, isWildcard, selectors, alias))
			case "dynimport":
				body = append(body, ast.NewDynImportStatement(path, isWildcard, selectors, alias))
			default:
				return nil, fmt.Errorf("parser: unsupported import kind %q", kindNode.Kind())
			}
		case "function_definition":
			fn, err := parseFunctionDefinition(node, source)
			if err != nil {
				return nil, err
			}
			body = append(body, fn)
		default:
			if !node.IsNamed() {
				continue
			}
			stmt, err := parseStatement(node, source)
			if err != nil {
				return nil, err
			}
			if stmt == nil {
				return nil, fmt.Errorf("parser: unsupported top-level node %q", node.Kind())
			}
			body = append(body, stmt)
		}
	}

	return ast.NewModule(body, imports, modulePackage), nil
}

func parsePackageStatement(node *sitter.Node, source []byte) (*ast.PackageStatement, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil package statement")
	}

	var parts []*ast.Identifier
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		id, err := parseIdentifier(child, source)
		if err != nil {
			return nil, err
		}
		parts = append(parts, id)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("parser: empty package statement")
	}

	return ast.NewPackageStatement(parts, false), nil
}

func parseQualifiedIdentifier(node *sitter.Node, source []byte) ([]*ast.Identifier, error) {
	if node == nil || node.Kind() != "qualified_identifier" {
		return nil, fmt.Errorf("parser: expected qualified identifier")
	}

	var parts []*ast.Identifier
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		id, err := parseIdentifier(child, source)
		if err != nil {
			return nil, err
		}
		parts = append(parts, id)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("parser: empty qualified identifier")
	}
	return parts, nil
}

func parseImportClause(node *sitter.Node, source []byte) (bool, []*ast.ImportSelector, *ast.Identifier, error) {
	if node == nil {
		return false, nil, nil, nil
	}

	var (
		isWildcard bool
		selectors  []*ast.ImportSelector
		alias      *ast.Identifier
	)

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		switch child.Kind() {
		case "import_selector":
			selector, err := parseImportSelector(child, source)
			if err != nil {
				return false, nil, nil, err
			}
			selectors = append(selectors, selector)
		case "import_wildcard_clause":
			isWildcard = true
		case "identifier":
			if alias != nil {
				return false, nil, nil, fmt.Errorf("parser: multiple aliases in import clause")
			}
			var err error
			alias, err = parseIdentifier(child, source)
			if err != nil {
				return false, nil, nil, err
			}
		default:
			return false, nil, nil, fmt.Errorf("parser: unsupported import clause node %q", child.Kind())
		}
	}

	if isWildcard && len(selectors) > 0 {
		return false, nil, nil, fmt.Errorf("parser: wildcard import cannot include selectors")
	}
	if alias != nil && len(selectors) > 0 {
		return false, nil, nil, fmt.Errorf("parser: alias cannot be combined with selectors")
	}

	return isWildcard, selectors, alias, nil
}

func parseImportSelector(node *sitter.Node, source []byte) (*ast.ImportSelector, error) {
	if node == nil || node.Kind() != "import_selector" {
		return nil, fmt.Errorf("parser: expected import_selector node")
	}

	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty import selector")
	}

	name, err := parseIdentifier(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}

	var alias *ast.Identifier
	if node.NamedChildCount() > 1 {
		alias, err = parseIdentifier(node.NamedChild(1), source)
		if err != nil {
			return nil, err
		}
	}

	return ast.NewImportSelector(name, alias), nil
}

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
		if child == nil {
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
		if child == targetNode || child == typeParamsNode || child == whereNode {
			continue
		}
		switch child.Kind() {
		case "function_definition":
			fn, err := parseFunctionDefinition(child, source)
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
				fn, err := parseFunctionDefinition(memberChild, source)
				if err != nil {
					return nil, err
				}
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
		if child == interfaceNode || child == interfaceArgsNode || child == targetNode || child == typeParamsNode || child == whereNode {
			continue
		}
		switch child.Kind() {
		case "function_definition":
			fn, err := parseFunctionDefinition(child, source)
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
				fn, err := parseFunctionDefinition(memberChild, source)
				if err != nil {
					return nil, err
				}
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

func parsePattern(node *sitter.Node, source []byte) (ast.Pattern, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil pattern")
	}

	if node.Kind() == "pattern" || node.Kind() == "pattern_base" {
		if node.NamedChildCount() == 0 {
			return nil, fmt.Errorf("parser: empty %s", node.Kind())
		}
		return parsePattern(node.NamedChild(0), source)
	}

	switch node.Kind() {
	case "identifier":
		return parseIdentifier(node, source)
	case "_":
		return ast.Wc(), nil
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

func parseBlock(node *sitter.Node, source []byte) (*ast.BlockExpression, error) {
	if node == nil {
		return ast.NewBlockExpression(nil), nil
	}

	statements := make([]ast.Statement, 0)
	for i := uint(0); i < node.NamedChildCount(); {
		child := node.NamedChild(i)
		i++
		if child == nil || !child.IsNamed() {
			continue
		}
		var (
			stmt ast.Statement
			err  error
		)
		if child.Kind() == "break_statement" {
			stmt, err = parseStatement(child, source)
			if err != nil {
				return nil, err
			}
			if brk, ok := stmt.(*ast.BreakStatement); ok && brk != nil && brk.Value == nil {
				if next := nextNamedSibling(node, i-1); next != nil && next.Kind() == "expression_statement" {
					exprNode := firstNamedChild(next)
					if exprNode != nil {
						expr, exprErr := parseExpression(exprNode, source)
						if exprErr != nil {
							return nil, exprErr
						}
						brk.Value = expr
						i++
					}
				}
			}
		} else {
			stmt, err = parseStatement(child, source)
			if err != nil {
				return nil, err
			}
		}
		if stmt != nil {
			if lambda, ok := stmt.(*ast.LambdaExpression); ok && len(statements) > 0 {
				switch prev := statements[len(statements)-1].(type) {
				case *ast.FunctionCall:
					if len(prev.Arguments) == 0 || prev.Arguments[len(prev.Arguments)-1] != lambda {
						prev.Arguments = append(prev.Arguments, lambda)
					}
					prev.IsTrailingLambda = true
					continue
				case ast.Expression:
					call := ast.NewFunctionCall(prev, nil, nil, true)
					call.Arguments = []ast.Expression{lambda}
					statements[len(statements)-1] = call
					continue
				}
			}
			statements = append(statements, stmt)
		}
	}

	return ast.NewBlockExpression(statements), nil
}

func parseStatement(node *sitter.Node, source []byte) (ast.Statement, error) {
	switch node.Kind() {
	case "expression_statement":
		exprNode := firstNamedChild(node)
		if exprNode == nil {
			return nil, fmt.Errorf("parser: expression statement missing expression")
		}
		expr, err := parseExpression(exprNode, source)
		if err != nil {
			return nil, err
		}
		return expr, nil
	case "return_statement":
		valueNode := firstNamedChild(node)
		if valueNode == nil {
			return ast.NewReturnStatement(nil), nil
		}
		expr, err := parseExpression(valueNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewReturnStatement(expr), nil
	case "while_statement":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed while statement")
		}
		conditionNode := node.NamedChild(0)
		bodyNode := node.NamedChild(1)
		condition, err := parseExpression(conditionNode, source)
		if err != nil {
			return nil, err
		}
		body, err := parseBlock(bodyNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewWhileLoop(condition, body), nil
	case "for_statement":
		if node.NamedChildCount() < 3 {
			return nil, fmt.Errorf("parser: malformed for statement")
		}
		patternNode := node.NamedChild(0)
		iterNode := node.NamedChild(1)
		bodyNode := node.NamedChild(2)
		pattern, err := parsePattern(patternNode, source)
		if err != nil {
			return nil, err
		}
		iterable, err := parseExpression(iterNode, source)
		if err != nil {
			return nil, err
		}
		body, err := parseBlock(bodyNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewForLoop(pattern, iterable, body), nil
	case "break_statement":
		labelNode := node.ChildByFieldName("label")
		var label *ast.Identifier
		if labelNode != nil {
			lbl, err := parseLabel(labelNode, source)
			if err != nil {
				return nil, err
			}
			label = lbl
		}
		valueNode := node.ChildByFieldName("value")
		var value ast.Expression
		if valueNode != nil {
			expr, err := parseExpression(valueNode, source)
			if err != nil {
				return nil, err
			}
			value = expr
		}
		return ast.NewBreakStatement(label, value), nil
	case "continue_statement":
		return ast.NewContinueStatement(nil), nil
	case "raise_statement":
		valueNode := firstNamedChild(node)
		if valueNode == nil {
			return nil, fmt.Errorf("parser: raise statement missing expression")
		}
		expr, err := parseExpression(valueNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewRaiseStatement(expr), nil
	case "struct_definition":
		return parseStructDefinition(node, source)
	case "methods_definition":
		return parseMethodsDefinition(node, source)
	case "implementation_definition":
		return parseImplementationDefinition(node, source)
	case "named_implementation_definition":
		return parseNamedImplementationDefinition(node, source)
	default:
		// For now, ignore unsupported statements in blocks.
		return nil, nil
	}
}

func parseImplicitMemberExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	memberNode := node.ChildByFieldName("member")
	if memberNode == nil {
		return nil, fmt.Errorf("parser: implicit member missing identifier")
	}
	member, err := parseIdentifier(memberNode, source)
	if err != nil {
		return nil, err
	}
	return ast.NewImplicitMemberExpression(member), nil
}

func parsePlaceholderExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	raw := strings.TrimSpace(sliceContent(node, source))
	if raw == "" {
		return nil, fmt.Errorf("parser: empty placeholder expression")
	}
	if raw == "@" {
		return ast.NewPlaceholderExpression(nil), nil
	}
	if strings.HasPrefix(raw, "@") {
		value := raw[1:]
		if value == "" {
			return ast.NewPlaceholderExpression(nil), nil
		}
		index, err := strconv.Atoi(value)
		if err != nil || index <= 0 {
			return nil, fmt.Errorf("parser: invalid placeholder index %q", raw)
		}
		return ast.NewPlaceholderExpression(&index), nil
	}
	return nil, fmt.Errorf("parser: unsupported placeholder token %q", raw)
}

func parseInterpolatedString(node *sitter.Node, source []byte) (ast.Expression, error) {
	parts := make([]ast.Expression, 0)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "interpolation_text":
			text := sliceContent(child, source)
			if text != "" {
				parts = append(parts, ast.Str(text))
			}
		case "string_interpolation":
			exprNode := child.ChildByFieldName("expression")
			if exprNode == nil {
				return nil, fmt.Errorf("parser: interpolation missing expression")
			}
			expr, err := parseExpression(exprNode, source)
			if err != nil {
				return nil, err
			}
			parts = append(parts, expr)
		}
	}
	return ast.NewStringInterpolation(parts), nil
}

func parseExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil expression node")
	}

	switch node.Kind() {
	case "identifier":
		return parseIdentifier(node, source)
	case "number_literal":
		return parseNumberLiteral(node, source)
	case "boolean_literal":
		return parseBooleanLiteral(node, source)
	case "nil_literal":
		return parseNilLiteral(node, source)
	case "string_literal":
		return parseStringLiteral(node, source)
	case "array_literal":
		return parseArrayLiteral(node, source)
	case "block":
		return parseBlock(node, source)
	case "do_expression":
		return parseDoExpression(node, source)
	case "lambda_expression":
		return parseLambdaExpression(node, source)
	case "postfix_expression":
		return parsePostfixExpression(node, source)
	case "call_target":
		return parsePostfixExpression(node, source)
	case "member_access":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed member access")
		}
		objectExpr, err := parseExpression(node.NamedChild(0), source)
		if err != nil {
			return nil, err
		}
		memberExpr, err := parseExpression(node.NamedChild(1), source)
		if err != nil {
			return nil, err
		}
		return ast.NewMemberAccessExpression(objectExpr, memberExpr), nil
	case "proc_expression":
		return parseProcExpression(node, source)
	case "spawn_expression":
		return parseSpawnExpression(node, source)
	case "if_expression":
		return parseIfExpression(node, source)
	case "range_expression":
		return parseRangeExpression(node, source)
	case "assignment_expression":
		return parseAssignmentExpression(node, source)
	case "unary_expression":
		return parseUnaryExpression(node, source)
	case "implicit_member_expression":
		return parseImplicitMemberExpression(node, source)
	case "placeholder_expression":
		return parsePlaceholderExpression(node, source)
	case "topic_reference":
		return ast.NewTopicReferenceExpression(), nil
	case "interpolated_string":
		return parseInterpolatedString(node, source)
	case "parenthesized_expression":
		if child := firstNamedChild(node); child != nil {
			return parseExpression(child, source)
		}
		return nil, fmt.Errorf("parser: empty parenthesized expression")
	case "pipe_expression":
		return parsePipeExpression(node, source)
	case "matchable_expression":
		if child := firstNamedChild(node); child != nil {
			return parseExpression(child, source)
		}
	}

	if operators, ok := infixOperatorSets[node.Kind()]; ok {
		return parseInfixExpression(node, source, operators)
	}

	if child := firstNamedChild(node); child != nil && child != node {
		return parseExpression(child, source)
	}

	if id, ok := findIdentifier(node, source); ok {
		return id, nil
	}

	return nil, fmt.Errorf("parser: unsupported expression kind %q", node.Kind())
}

func parsePostfixExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty postfix expression")
	}

	result, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}

	var (
		pendingTypeArgs []ast.TypeExpression
		lastCall        *ast.FunctionCall
	)

	for i := uint(1); i < node.NamedChildCount(); i++ {
		suffix := node.NamedChild(i)
		switch suffix.Kind() {
		case "member_access":
			memberNode := firstNamedChild(suffix)
			if memberNode == nil {
				return nil, fmt.Errorf("parser: member access missing member")
			}
			memberExpr, err := parseExpression(memberNode, source)
			if err != nil {
				return nil, err
			}
			result = ast.NewMemberAccessExpression(result, memberExpr)
			lastCall = nil
		case "type_arguments":
			typeArgs, err := parseTypeArgumentList(suffix, source)
			if err != nil {
				return nil, err
			}
			pendingTypeArgs = typeArgs
			lastCall = nil
		case "index_suffix":
			if suffix.NamedChildCount() == 0 {
				return nil, fmt.Errorf("parser: index expression missing index value")
			}
			if suffix.NamedChildCount() > 1 {
				return nil, fmt.Errorf("parser: slice expressions are not supported yet")
			}
			indexExpr, err := parseExpression(suffix.NamedChild(0), source)
			if err != nil {
				return nil, err
			}
			result = ast.NewIndexExpression(result, indexExpr)
			lastCall = nil
		case "call_suffix":
			args, err := parseCallArguments(suffix, source)
			if err != nil {
				return nil, err
			}
			typeArgs := pendingTypeArgs
			pendingTypeArgs = nil

			callExpr := ast.NewFunctionCall(result, args, typeArgs, false)
			result = callExpr
			lastCall = callExpr
		case "lambda_expression":
			lambdaExpr, err := parseLambdaExpression(suffix, source)
			if err != nil {
				return nil, err
			}

			typeArgs := pendingTypeArgs
			pendingTypeArgs = nil

			if lastCall != nil && !lastCall.IsTrailingLambda {
				lastCall.Arguments = append(lastCall.Arguments, lambdaExpr)
				lastCall.IsTrailingLambda = true
				result = lastCall
			} else {
				callExpr := ast.NewFunctionCall(result, nil, typeArgs, true)
				callExpr.Arguments = append(callExpr.Arguments, lambdaExpr)
				result = callExpr
				lastCall = callExpr
			}
		default:
			return nil, fmt.Errorf("parser: unsupported postfix suffix %q", suffix.Kind())
		}
	}

	if len(pendingTypeArgs) > 0 {
		return nil, fmt.Errorf("parser: dangling type arguments in expression")
	}

	return result, nil
}

func parseCallArguments(node *sitter.Node, source []byte) ([]ast.Expression, error) {
	args := make([]ast.Expression, 0)

	for j := uint(0); j < node.NamedChildCount(); j++ {
		child := node.NamedChild(j)
		if !child.IsNamed() {
			continue
		}
		argExpr, err := parseExpression(child, source)
		if err != nil {
			return nil, err
		}
		args = append(args, argExpr)
	}

	return args, nil
}

func parseTypeArgumentList(node *sitter.Node, source []byte) ([]ast.TypeExpression, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil type argument list")
	}

	var args []ast.TypeExpression
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if !child.IsNamed() {
			continue
		}
		typeExpr := parseTypeExpression(child, source)
		if typeExpr == nil {
			return nil, fmt.Errorf("parser: unsupported type argument kind %q", child.Kind())
		}
		args = append(args, typeExpr)
	}

	return args, nil
}

func parseBooleanLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, source))
	switch value {
	case "true":
		return ast.Bool(true), nil
	case "false":
		return ast.Bool(false), nil
	default:
		return nil, fmt.Errorf("parser: invalid boolean literal %q", value)
	}
}

func parseNilLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, source))
	if value != "nil" {
		return nil, fmt.Errorf("parser: invalid nil literal %q", value)
	}
	return ast.Nil(), nil
}

func parseArrayLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	var elements []ast.Expression
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if !child.IsNamed() {
			continue
		}
		element, err := parseExpression(child, source)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	return ast.NewArrayLiteral(elements), nil
}

func parseDoExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: do expression missing body")
	}
	return parseBlock(bodyNode, source)
}

func parseProcExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: proc expression missing body")
	}
	body, err := parseExpression(bodyNode, source)
	if err != nil {
		return nil, err
	}
	return ast.NewProcExpression(body), nil
}

func parseSpawnExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: spawn expression missing body")
	}
	body, err := parseExpression(bodyNode, source)
	if err != nil {
		return nil, err
	}
	return ast.NewSpawnExpression(body), nil
}

func parseIfExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	conditionNode := node.ChildByFieldName("condition")
	if conditionNode == nil {
		return nil, fmt.Errorf("parser: if expression missing condition")
	}
	condition, err := parseExpression(conditionNode, source)
	if err != nil {
		return nil, err
	}
	bodyNode := node.ChildByFieldName("consequence")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: if expression missing body")
	}
	body, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}
	clauses := make([]*ast.OrClause, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Kind() == "or_clause" {
			clause, err := parseOrClause(child, source)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
		}
	}
	if elseNode := findElseBlock(node, bodyNode); elseNode != nil {
		elseBody, err := parseBlock(elseNode, source)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, ast.NewOrClause(elseBody, nil))
	}
	return ast.NewIfExpression(condition, body, clauses), nil
}

func parseOrClause(node *sitter.Node, source []byte) (*ast.OrClause, error) {
	bodyNode := node.ChildByFieldName("consequence")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: or clause missing body")
	}
	body, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}
	conditionNode := node.ChildByFieldName("condition")
	var condition ast.Expression
	if conditionNode != nil {
		condExpr, err := parseExpression(conditionNode, source)
		if err != nil {
			return nil, err
		}
		condition = condExpr
	}
	return ast.NewOrClause(body, condition), nil
}

func findElseBlock(ifNode *sitter.Node, consequence *sitter.Node) *sitter.Node {
	if ifNode == nil {
		return nil
	}
	var consequenceRangeStart, consequenceRangeEnd uint
	if consequence != nil {
		consequenceRangeStart = consequence.StartByte()
		consequenceRangeEnd = consequence.EndByte()
	}
	for i := uint(0); i < ifNode.NamedChildCount(); i++ {
		child := ifNode.NamedChild(i)
		if child.Kind() != "block" {
			continue
		}
		if consequence != nil && child.StartByte() == consequenceRangeStart && child.EndByte() == consequenceRangeEnd {
			continue
		}
		return child
	}
	return nil
}

func parseRangeExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil || node.NamedChildCount() < 2 {
		if child := firstNamedChild(node); child != nil {
			return parseExpression(child, source)
		}
		return nil, fmt.Errorf("parser: malformed range expression")
	}
	startExpr, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}
	endExpr, err := parseExpression(node.NamedChild(1), source)
	if err != nil {
		return nil, err
	}
	operatorText := strings.TrimSpace(sliceContent(operatorNode, source))
	inclusive := operatorText == "..."
	if operatorText != ".." && operatorText != "..." {
		return nil, fmt.Errorf("parser: unsupported range operator %q", operatorText)
	}
	return ast.NewRangeExpression(startExpr, endExpr, inclusive), nil
}

func parseAssignmentExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil {
		child := firstNamedChild(node)
		if child == nil {
			return nil, fmt.Errorf("parser: empty assignment expression")
		}
		return parseExpression(child, source)
	}
	leftNode := node.ChildByFieldName("left")
	rightNode := node.ChildByFieldName("right")
	if leftNode == nil || rightNode == nil {
		return nil, fmt.Errorf("parser: malformed assignment expression")
	}
	left, err := parseAssignmentTarget(leftNode, source)
	if err != nil {
		return nil, err
	}
	right, err := parseExpression(rightNode, source)
	if err != nil {
		return nil, err
	}
	operatorText := strings.TrimSpace(sliceContent(operatorNode, source))
	operator, err := mapAssignmentOperator(operatorText)
	if err != nil {
		return nil, err
	}
	return ast.NewAssignmentExpression(operator, left, right), nil
}

func parseAssignmentTarget(node *sitter.Node, source []byte) (ast.AssignmentTarget, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil assignment target")
	}
	switch node.Kind() {
	case "assignment_target":
		if child := firstNamedChild(node); child != nil {
			return parseAssignmentTarget(child, source)
		}
		return nil, fmt.Errorf("parser: empty assignment target")
	case "pattern", "pattern_base":
		pattern, err := parsePattern(node, source)
		if err != nil {
			return nil, err
		}
		target, ok := pattern.(ast.AssignmentTarget)
		if !ok {
			return nil, fmt.Errorf("parser: pattern cannot be used as assignment target: %T", pattern)
		}
		return target, nil
	default:
		expr, err := parseExpression(node, source)
		if err != nil {
			return nil, err
		}
		target, ok := expr.(ast.AssignmentTarget)
		if !ok {
			return nil, fmt.Errorf("parser: expression cannot be used as assignment target: %T", expr)
		}
		return target, nil
	}
}

func parseUnaryExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operandNode := firstNamedChild(node)
	if operandNode == nil {
		return nil, fmt.Errorf("parser: unary expression missing operand")
	}
	if int(node.StartByte()) == int(operandNode.StartByte()) {
		return parseExpression(operandNode, source)
	}
	operatorText := strings.TrimSpace(string(source[int(node.StartByte()):int(operandNode.StartByte())]))
	if operatorText == "" {
		return parseExpression(operandNode, source)
	}
	operand, err := parseExpression(operandNode, source)
	if err != nil {
		return nil, err
	}
	switch operatorText {
	case "-":
		return ast.NewUnaryExpression(ast.UnaryOperatorNegate, operand), nil
	case "!":
		return ast.NewUnaryExpression(ast.UnaryOperatorNot, operand), nil
	case "~":
		return ast.NewUnaryExpression(ast.UnaryOperatorBitNot, operand), nil
	default:
		return nil, fmt.Errorf("parser: unsupported unary operator %q", operatorText)
	}
}

func parsePipeExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty pipe expression")
	}
	if node.NamedChildCount() == 1 {
		return parseExpression(node.NamedChild(0), source)
	}
	return nil, fmt.Errorf("parser: pipe expressions with chaining are not supported yet")
}

func parseInfixExpression(node *sitter.Node, source []byte, operators []string) (ast.Expression, error) {
	count := node.NamedChildCount()
	if count == 0 {
		return nil, fmt.Errorf("parser: empty %s", node.Kind())
	}
	if count == 1 {
		return parseExpression(node.NamedChild(0), source)
	}
	result, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}
	prevNode := node.NamedChild(0)
	for i := uint(1); i < count; i++ {
		rightNode := node.NamedChild(i)
		rightExpr, err := parseExpression(rightNode, source)
		if err != nil {
			return nil, err
		}
		operator := extractOperatorBetween(prevNode, rightNode, source, operators)
		if operator == "" {
			return nil, fmt.Errorf("parser: could not determine operator between operands in %s", node.Kind())
		}
		result = ast.NewBinaryExpression(operator, result, rightExpr)
		prevNode = rightNode
	}
	return result, nil
}

func extractOperatorBetween(left, right *sitter.Node, source []byte, allowed []string) string {
	if left == nil || right == nil {
		return ""
	}
	start := int(left.EndByte())
	end := int(right.StartByte())
	if start < 0 || end < start || end > len(source) {
		return ""
	}
	segment := strings.TrimSpace(string(source[start:end]))
	if segment == "" {
		return ""
	}
	for _, op := range allowed {
		if segment == op {
			return op
		}
	}
	for _, op := range allowed {
		if strings.Contains(segment, op) {
			return op
		}
	}
	return ""
}

var assignmentOperatorMap = map[string]ast.AssignmentOperator{
	":=":  ast.AssignmentDeclare,
	"=":   ast.AssignmentAssign,
	"+=":  ast.AssignmentAdd,
	"-=":  ast.AssignmentSub,
	"*=":  ast.AssignmentMul,
	"/=":  ast.AssignmentDiv,
	"%=":  ast.AssignmentMod,
	"&=":  ast.AssignmentBitAnd,
	"|=":  ast.AssignmentBitOr,
	"^=":  ast.AssignmentBitXor,
	"<<=": ast.AssignmentShiftL,
	">>=": ast.AssignmentShiftR,
}

func mapAssignmentOperator(op string) (ast.AssignmentOperator, error) {
	if operator, ok := assignmentOperatorMap[op]; ok {
		return operator, nil
	}
	return "", fmt.Errorf("parser: unsupported assignment operator %q", op)
}

func parseLambdaExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "lambda_expression" {
		return nil, fmt.Errorf("parser: expected lambda expression")
	}

	var params []*ast.FunctionParameter
	if paramsNode := node.ChildByFieldName("parameters"); paramsNode != nil {
		for i := uint(0); i < paramsNode.NamedChildCount(); i++ {
			paramNode := paramsNode.NamedChild(i)
			if paramNode == nil || paramNode.Kind() != "lambda_parameter" {
				continue
			}
			param, err := parseLambdaParameter(paramNode, source)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
	}

	var returnType ast.TypeExpression
	if returnNode := node.ChildByFieldName("return_type"); returnNode != nil {
		returnType = parseReturnType(returnNode, source)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: lambda missing body")
	}

	bodyExpr, err := parseExpression(bodyNode, source)
	if err != nil {
		return nil, err
	}

	return ast.NewLambdaExpression(params, bodyExpr, returnType, nil, nil, false), nil
}

func parseLambdaParameter(node *sitter.Node, source []byte) (*ast.FunctionParameter, error) {
	if node == nil || node.Kind() != "lambda_parameter" {
		return nil, fmt.Errorf("parser: expected lambda parameter")
	}

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil, fmt.Errorf("parser: lambda parameter missing name")
	}

	id, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	return ast.NewFunctionParameter(id, nil), nil
}

func firstNamedChild(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}
	for i := uint(0); i < node.NamedChildCount(); i++ {
		return node.NamedChild(i)
	}
	return nil
}

func nextNamedSibling(parent *sitter.Node, currentIndex uint) *sitter.Node {
	if parent == nil {
		return nil
	}
	for j := currentIndex + 1; j < parent.NamedChildCount(); j++ {
		sibling := parent.NamedChild(j)
		if sibling != nil && sibling.IsNamed() {
			return sibling
		}
	}
	return nil
}

func findIdentifier(node *sitter.Node, source []byte) (*ast.Identifier, bool) {
	if node == nil {
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
	nameNode := firstNamedChild(node)
	if nameNode == nil {
		return nil, fmt.Errorf("parser: label missing identifier")
	}
	return parseIdentifier(nameNode, source)
}

func parseNumberLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	content := sliceContent(node, source)
	if content == "" {
		return nil, fmt.Errorf("parser: empty number literal")
	}
	sanitized := strings.ReplaceAll(content, "_", "")
	value := new(big.Int)
	if _, ok := value.SetString(sanitized, 10); !ok {
		return nil, fmt.Errorf("parser: invalid number literal %q", content)
	}
	return ast.IntBig(value, nil), nil
}

func parseStringLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	raw := sliceContent(node, source)
	unquoted, err := strconv.Unquote(raw)
	if err != nil {
		return nil, fmt.Errorf("parser: invalid string literal %q: %w", raw, err)
	}
	return ast.Str(unquoted), nil
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
			return parseTypeExpression(child, source)
		}
	case "type_suffix":
		if node.NamedChildCount() > 1 {
			base := parseTypeExpression(node.NamedChild(0), source)
			var args []ast.TypeExpression
			for i := uint(1); i < node.NamedChildCount(); i++ {
				arg := parseTypeExpression(node.NamedChild(i), source)
				if arg != nil {
					args = append(args, arg)
				}
			}
			if base != nil && len(args) > 0 {
				return ast.NewGenericTypeExpression(base, args)
			}
		}
		if child := firstNamedChild(node); child != nil && child != node {
			return parseTypeExpression(child, source)
		}
	case "type_arrow":
		if node.NamedChildCount() >= 2 {
			paramExpr := parseTypeExpression(node.NamedChild(0), source)
			returnExpr := parseTypeExpression(node.NamedChild(1), source)
			if paramExpr != nil && returnExpr != nil {
				return ast.NewFunctionTypeExpression([]ast.TypeExpression{paramExpr}, returnExpr)
			}
		}
		if child := firstNamedChild(node); child != nil && child != node {
			return parseTypeExpression(child, source)
		}
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
			return members[0]
		}
		if len(members) > 1 {
			return ast.NewUnionTypeExpression(members)
		}
	case "type_identifier":
		if child := firstNamedChild(node); child != nil && child != node {
			return parseTypeExpression(child, source)
		}
	case "identifier":
		id, err := parseIdentifier(node, source)
		if err != nil {
			return nil
		}
		return ast.TyID(id)
	case "qualified_identifier":
		parts, err := parseQualifiedIdentifier(node, source)
		if err != nil || len(parts) == 0 {
			return nil
		}
		if len(parts) == 1 {
			return ast.TyID(parts[0])
		}
		name := ast.ID(strings.Join(identifiersToStrings(parts), "."))
		return ast.TyID(name)
	default:
		if child := firstNamedChild(node); child != nil && child != node {
			if expr := parseTypeExpression(child, source); expr != nil {
				return expr
			}
		}
	}
	text := strings.TrimSpace(sliceContent(node, source))
	if text == "" {
		return nil
	}
	return ast.Ty(strings.ReplaceAll(text, " ", ""))
}

func identifiersToStrings(ids []*ast.Identifier) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.Name
	}
	return result
}

func parseTypeParameters(node *sitter.Node, source []byte) ([]*ast.GenericParameter, error) {
	if node == nil {
		return nil, nil
	}
	switch node.Kind() {
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
	default:
		return nil, fmt.Errorf("parser: unsupported type parameter node %q", node.Kind())
	}
}

func parseTypeParameter(node *sitter.Node, source []byte) (*ast.GenericParameter, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil type parameter")
	}
	var nameNode *sitter.Node
	if node.NamedChildCount() > 0 {
		nameNode = node.NamedChild(0)
	}
	if nameNode == nil || nameNode.Kind() != "identifier" {
		return nil, fmt.Errorf("parser: type parameter missing identifier")
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

	return ast.NewGenericParameter(name, constraints), nil
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
	return ast.NewWhereClauseConstraint(name, interfaceConstraints), nil
}
