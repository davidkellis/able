package interpreter

import (
	"fmt"
	"io/fs"

	"able/interpreter-go/pkg/ast"
)

func decodeModuleNodes(node map[string]any, typ string) (ast.Node, bool, error) {
	switch typ {
	case "Module":
		importsVal, _ := node["imports"].([]any)
		imports := make([]*ast.ImportStatement, 0, len(importsVal))
		for _, raw := range importsVal {
			child, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid import entry %T", raw)
			}
			imp, err := decodeImportStatement(child)
			if err != nil {
				return nil, true, err
			}
			imports = append(imports, imp)
		}
		bodyVal, _ := node["body"].([]any)
		stmts := make([]ast.Statement, 0, len(bodyVal))
		for _, raw := range bodyVal {
			child, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, true, err
			}
			stmt, ok := child.(ast.Statement)
			if !ok {
				return nil, true, fs.ErrInvalid
			}
			stmts = append(stmts, stmt)
		}
		var pkg *ast.PackageStatement
		if pkgNode, ok := node["package"].(map[string]any); ok {
			decoded, err := decodePackageStatement(pkgNode)
			if err != nil {
				return nil, true, err
			}
			pkg = decoded
		}
		return ast.NewModule(stmts, imports, pkg), true, nil
	case "PreludeStatement":
		target, _ := node["target"].(string)
		code, _ := node["code"].(string)
		return ast.NewPreludeStatement(ast.HostTarget(target), code), true, nil
	case "ExternFunctionBody":
		sigRaw, ok := node["signature"].(map[string]any)
		if !ok {
			return nil, true, fmt.Errorf("extern function body missing signature")
		}
		sigNode, err := decodeNode(sigRaw)
		if err != nil {
			return nil, true, err
		}
		signature, ok := sigNode.(*ast.FunctionDefinition)
		if !ok {
			return nil, true, fmt.Errorf("invalid extern signature %T", sigNode)
		}
		target, _ := node["target"].(string)
		body, _ := node["body"].(string)
		return ast.NewExternFunctionBody(ast.HostTarget(target), signature, body), true, nil
	case "ImportStatement":
		stmt, err := decodeImportStatement(node)
		if err != nil {
			return nil, true, err
		}
		return stmt, true, nil
	case "DynImportStatement":
		stmt, err := decodeDynImportStatement(node)
		if err != nil {
			return nil, true, err
		}
		return stmt, true, nil
	default:
		return nil, false, nil
	}
}
