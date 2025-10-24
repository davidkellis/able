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
		imports       []*ast.ImportStatement
		body          []ast.Statement
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
			if node.IsNamed() {
				return nil, fmt.Errorf("parser: unsupported top-level node %q", node.Kind())
			}
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

func parseFunctionDefinition(node *sitter.Node, source []byte) (*ast.FunctionDefinition, error) {
	if node == nil || node.Kind() != "function_definition" {
		return nil, fmt.Errorf("parser: expected function_definition node")
	}

	name, err := parseIdentifier(node.ChildByFieldName("name"), source)
	if err != nil {
		return nil, err
	}

	params, err := parseParameterList(node.ChildByFieldName("parameters"), source)
	if err != nil {
		return nil, err
	}

	bodyNode := node.ChildByFieldName("body")
	fnBody, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}

	isPrivate := false
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "private" {
			isPrivate = true
			break
		}
	}

	fn := ast.NewFunctionDefinition(
		name,
		params,
		fnBody,
		parseReturnType(node.ChildByFieldName("return_type"), source),
		nil, // generics
		nil, // where clause
		false,
		isPrivate,
	)

	return fn, nil
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

func parseParameter(node *sitter.Node, source []byte) (*ast.FunctionParameter, error) {
	if node == nil || node.Kind() != "parameter" {
		return nil, fmt.Errorf("parser: expected parameter node")
	}

	patternNode := node.ChildByFieldName("pattern")
	pattern, err := parsePattern(patternNode, source)
	if err != nil {
		return nil, err
	}

	// TODO: parse parameter types
	return ast.NewFunctionParameter(pattern, nil), nil
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
	default:
		return nil, fmt.Errorf("parser: unsupported pattern kind %q", node.Kind())
	}
}

func parseBlock(node *sitter.Node, source []byte) (*ast.BlockExpression, error) {
	if node == nil {
		return ast.NewBlockExpression(nil), nil
	}

	var statements []ast.Statement
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if !child.IsNamed() {
			continue
		}
		stmt, err := parseStatement(child, source)
		if err != nil {
			return nil, err
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
	default:
		// For now, ignore unsupported statements in blocks.
		return nil, nil
	}
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
	case "string_literal":
		return parseStringLiteral(node, source)
	case "lambda_expression":
		return parseLambdaExpression(node, source)
	case "postfix_expression":
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
	case "additive_expression":
		if node.NamedChildCount() == 1 {
			return parseExpression(node.NamedChild(0), source)
		}
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed additive expression")
		}
		leftNode := node.NamedChild(0)
		rightNode := node.NamedChild(1)
		leftExpr, err := parseExpression(leftNode, source)
		if err != nil {
			return nil, err
		}
		rightExpr, err := parseExpression(rightNode, source)
		if err != nil {
			return nil, err
		}
		operator := detectOperatorBetween(leftNode, rightNode, source)
		if operator == "" {
			return nil, fmt.Errorf("parser: unknown operator in additive expression")
		}
		return ast.NewBinaryExpression(operator, leftExpr, rightExpr), nil
	case "multiplicative_expression",
		"range_expression",
		"logical_or_expression",
		"logical_and_expression",
		"bitwise_or_expression",
		"bitwise_xor_expression",
		"bitwise_and_expression",
		"equality_expression",
		"comparison_expression",
		"shift_expression",
		"unary_expression",
		"exponent_expression",
		"matchable_expression",
		"pipe_expression",
		"assignment_expression":
		if child := firstNamedChild(node); child != nil {
			return parseExpression(child, source)
		}
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
	var args []ast.Expression

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
		typeExpr := parseReturnType(child, source)
		if typeExpr == nil {
			return nil, fmt.Errorf("parser: unsupported type argument kind %q", child.Kind())
		}
		args = append(args, typeExpr)
	}

	return args, nil
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

func detectOperatorBetween(left, right *sitter.Node, source []byte) string {
	if left == nil || right == nil {
		return ""
	}
	start := int(left.EndByte())
	end := int(right.StartByte())
	if start < 0 || end < start || end > len(source) {
		return ""
	}
	segment := strings.TrimSpace(string(source[start:end]))
	if strings.Contains(segment, "+") {
		return "+"
	}
	if strings.Contains(segment, "-") {
		return "-"
	}
	return ""
}

func parseReturnType(node *sitter.Node, source []byte) ast.TypeExpression {
	if node == nil {
		return nil
	}
	switch node.Kind() {
	case "return_type", "type_expression", "type_identifier":
		if child := firstNamedChild(node); child != nil && child != node {
			return parseReturnType(child, source)
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
			return parseReturnType(child, source)
		}
	}
	return nil
}

func identifiersToStrings(ids []*ast.Identifier) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.Name
	}
	return result
}
