package parser

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser/language"
)

// ModuleParser wraps a tree-sitter parser configured for Able v12 modules.
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
	if root == nil {
		return nil, fmt.Errorf("parser: unexpected root node")
	}
	if root.Kind() != "source_file" {
		if root.HasError() {
			return nil, syntaxError(root)
		}
		return nil, fmt.Errorf("parser: unexpected root node")
	}
	if root.HasError() && !recoverableInterfaceBaseErrors(root, source) && !recoverableWhitespaceErrors(root, source) {
		return nil, syntaxError(root)
	}

	ctx := newParseContext(source)

	var (
		modulePackage *ast.PackageStatement
		imports       = make([]*ast.ImportStatement, 0)
		body          = make([]ast.Statement, 0)
	)

	for i := uint(0); i < root.NamedChildCount(); i++ {
		node := root.NamedChild(i)
		if isIgnorableNode(node) {
			continue
		}
		switch node.Kind() {
		case "package_statement":
			pkg, err := ctx.parsePackageStatement(node)
			if err != nil {
				return nil, wrapParseError(node, err)
			}
			modulePackage = pkg
		case "import_statement":
			kindNode := node.ChildByFieldName("kind")
			if kindNode == nil {
				return nil, wrapParseError(node, fmt.Errorf("parser: import missing kind"))
			}

			path, err := ctx.parseQualifiedIdentifier(node.ChildByFieldName("path"))
			if err != nil {
				return nil, wrapParseError(node, err)
			}

			aliasNode := node.ChildByFieldName("alias")
			var alias *ast.Identifier
			if aliasNode != nil {
				alias, err = parseIdentifier(aliasNode, ctx.source)
				if err != nil {
					return nil, wrapParseError(aliasNode, err)
				}
			}

			isWildcard, selectors, err := ctx.parseImportClause(node.ChildByFieldName("clause"))
			if err != nil {
				return nil, wrapParseError(node, err)
			}

			if alias != nil && len(selectors) > 0 {
				return nil, wrapParseError(node, fmt.Errorf("parser: alias cannot be combined with selectors"))
			}

			if alias == nil && !isWildcard && len(selectors) == 0 && hasLegacyImportAlias(node, ctx.source) {
				return nil, wrapParseError(node, fmt.Errorf("parser: legacy import alias syntax is unsupported; use :: for renames"))
			}

			switch kindNode.Kind() {
			case "import":
				stmt := ast.NewImportStatement(path, isWildcard, selectors, alias)
				annotateSpan(stmt, node)
				imports = append(imports, stmt)
			case "dynimport":
				dyn := ast.NewDynImportStatement(path, isWildcard, selectors, alias)
				annotateSpan(dyn, node)
				body = append(body, dyn)
			default:
				return nil, wrapParseError(kindNode, fmt.Errorf("parser: unsupported import kind %q", kindNode.Kind()))
			}
		case "function_definition":
			fn, err := ctx.parseFunctionDefinition(node)
			if err != nil {
				return nil, wrapParseError(node, err)
			}
			body = append(body, fn)
		default:
			if !node.IsNamed() {
				continue
			}
			stmt, err := ctx.parseStatement(node)
			if err != nil {
				return nil, wrapParseError(node, err)
			}
			if stmt == nil {
				return nil, fmt.Errorf("parser: unsupported top-level node %q", node.Kind())
			}
			if stmt != nil {
				if lambda, ok := stmt.(*ast.LambdaExpression); ok && len(body) > 0 {
					switch prev := body[len(body)-1].(type) {
					case *ast.AssignmentExpression:
						switch rhs := prev.Right.(type) {
						case *ast.FunctionCall:
							if len(rhs.Arguments) == 0 || rhs.Arguments[len(rhs.Arguments)-1] != lambda {
								rhs.Arguments = append(rhs.Arguments, lambda)
							}
							rhs.IsTrailingLambda = true
							continue
						case ast.Expression:
							call := ast.NewFunctionCall(rhs, nil, nil, true)
							call.Arguments = []ast.Expression{lambda}
							prev.Right = call
							continue
						}
					case *ast.FunctionCall:
						if len(prev.Arguments) == 0 || prev.Arguments[len(prev.Arguments)-1] != lambda {
							prev.Arguments = append(prev.Arguments, lambda)
						}
						prev.IsTrailingLambda = true
						continue
					case ast.Expression:
						call := ast.NewFunctionCall(prev, nil, nil, true)
						call.Arguments = []ast.Expression{lambda}
						body[len(body)-1] = call
						continue
					}
				}
				body = append(body, stmt)
			}
		}
	}

	module := ast.NewModule(body, imports, modulePackage)
	module.Body = repairTypeAliasTargets(module.Body, source)
	annotateSpan(module, root)
	return module, nil
}

func repairTypeAliasTargets(body []ast.Statement, source []byte) []ast.Statement {
	if len(body) == 0 {
		return body
	}
	lines := strings.Split(string(source), "\n")
	repaired := make([]ast.Statement, 0, len(body))
	for i := 0; i < len(body); i++ {
		stmt := body[i]
		alias, ok := stmt.(*ast.TypeAliasDefinition)
		if !ok || alias.TargetType == nil || len(alias.GenericParams) == 0 {
			repaired = append(repaired, stmt)
			continue
		}
		genericNames := make(map[string]struct{})
		for _, gp := range alias.GenericParams {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				continue
			}
			genericNames[gp.Name.Name] = struct{}{}
		}
		if len(genericNames) == 0 {
			repaired = append(repaired, stmt)
			continue
		}
		span := alias.Span()
		end := span.End
		if end.Line == 0 {
			repaired = append(repaired, stmt)
			continue
		}
		target := alias.TargetType
		consumed := 0
		for j := i + 1; j < len(body); j++ {
			ident, ok := body[j].(*ast.Identifier)
			if !ok {
				break
			}
			idSpan := ident.Span()
			if idSpan.Start.Line != end.Line {
				break
			}
			lineText := ""
			if idx := end.Line - 1; idx >= 0 && idx < len(lines) {
				lineText = lines[idx]
			}
			startCol := end.Column - 1
			endCol := idSpan.Start.Column - 1
			if startCol < 0 {
				startCol = 0
			}
			if endCol < 0 {
				endCol = 0
			}
			if startCol > len(lineText) {
				startCol = len(lineText)
			}
			if endCol > len(lineText) {
				endCol = len(lineText)
			}
			if strings.TrimSpace(lineText[startCol:endCol]) != "" {
				break
			}
			if _, ok := genericNames[ident.Name]; !ok {
				break
			}
			extraArg := ast.Ty(ident.Name)
			switch t := target.(type) {
			case *ast.GenericTypeExpression:
				t.Arguments = append(t.Arguments, extraArg)
				target = t
			default:
				target = ast.Gen(target, extraArg)
			}
			end = idSpan.End
			consumed++
		}
		if consumed > 0 {
			alias.TargetType = target
			span.End = end
			ast.SetSpan(alias, span)
			i += consumed
		}
		repaired = append(repaired, alias)
	}
	return repaired
}

func (ctx *parseContext) parsePackageStatement(node *sitter.Node) (*ast.PackageStatement, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil package statement")
	}

	var parts []*ast.Identifier
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if isIgnorableNode(child) {
			continue
		}
		id, err := parseIdentifier(child, ctx.source)
		if err != nil {
			return nil, err
		}
		parts = append(parts, id)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("parser: empty package statement")
	}
	if len(parts) != 1 {
		return nil, fmt.Errorf("parser: package statement must use a single, unqualified name")
	}

	stmt := ast.NewPackageStatement(parts, false)
	annotateSpan(stmt, node)
	return stmt, nil
}

func parseQualifiedIdentifier(node *sitter.Node, source []byte) ([]*ast.Identifier, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: expected qualified identifier")
	}

	switch node.Kind() {
	case "qualified_identifier", "import_path":
	default:
		return nil, fmt.Errorf("parser: expected qualified identifier")
	}

	var parts []*ast.Identifier
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if isIgnorableNode(child) {
			continue
		}
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

func (ctx *parseContext) parseImportClause(node *sitter.Node) (bool, []*ast.ImportSelector, error) {
	if node == nil {
		return false, nil, nil
	}

	var (
		isWildcard bool
		selectors  []*ast.ImportSelector
	)

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if isIgnorableNode(child) {
			continue
		}
		switch child.Kind() {
		case "import_selector":
			selector, err := parseImportSelector(child, ctx.source)
			if err != nil {
				return false, nil, err
			}
			selectors = append(selectors, selector)
		case "import_wildcard_clause":
			isWildcard = true
		default:
			return false, nil, fmt.Errorf("parser: unsupported import clause node %q", child.Kind())
		}
	}

	if isWildcard && len(selectors) > 0 {
		return false, nil, fmt.Errorf("parser: wildcard import cannot include selectors")
	}

	return isWildcard, selectors, nil
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

	selector := ast.NewImportSelector(name, alias)
	annotateSpan(selector, node)
	return selector, nil
}
