package interpreter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type fixtureManifest struct {
	Description string   `json:"description"`
	Entry       string   `json:"entry"`
	Setup       []string `json:"setup"`
	Expect      struct {
		Result *struct {
			Kind  string      `json:"kind"`
			Value interface{} `json:"value"`
		} `json:"result"`
		Stdout []string `json:"stdout"`
		Errors []string `json:"errors"`
	} `json:"expect"`
}

func TestFixtureParityStringLiteral(t *testing.T) {
	root := filepath.Join("..", "..", "..", "fixtures", "ast")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fixtureDir := filepath.Join(root, entry.Name())
		walkFixtures(t, fixtureDir, func(dir string) {
			interp := New()
			var stdout []string
			registerPrint(interp, &stdout)
			manifest := readManifest(t, dir)
			entryFile := manifest.Entry
			if entryFile == "" {
				entryFile = "module.json"
			}
			modulePath := filepath.Join(dir, entryFile)
			module := readModule(t, modulePath)
			if len(manifest.Setup) > 0 {
				for _, setupFile := range manifest.Setup {
					setupPath := filepath.Join(dir, setupFile)
					setupModule := readModule(t, setupPath)
					if _, _, err := interp.EvaluateModule(setupModule); err != nil {
						t.Fatalf("fixture %s setup module %s failed: %v", dir, setupFile, err)
					}
				}
			}
			result, _, err := interp.EvaluateModule(module)
			if len(manifest.Expect.Errors) > 0 {
				if err == nil {
					t.Fatalf("fixture %s expected evaluation error", dir)
				}
				msg := extractErrorMessage(err)
				if !contains(manifest.Expect.Errors, msg) {
					t.Fatalf("fixture %s expected error message in %v, got %s", dir, manifest.Expect.Errors, msg)
				}
				return
			}
			if err != nil {
				t.Fatalf("fixture %s evaluation error: %v", dir, err)
			}
			assertResult(t, dir, manifest, result, stdout)
		})
	}
}

func walkFixtures(t *testing.T, dir string, fn func(string)) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	hasModule := false
	for _, entry := range entries {
		if entry.Type().IsRegular() && entry.Name() == "module.json" {
			hasModule = true
		}
	}
	if hasModule {
		fn(dir)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			walkFixtures(t, filepath.Join(dir, entry.Name()), fn)
		}
	}
}

func readManifest(t *testing.T, dir string) fixtureManifest {
	t.Helper()
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fixtureManifest{}
		}
		t.Fatalf("read manifest %s: %v", manifestPath, err)
	}
	var manifest fixtureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse manifest %s: %v", manifestPath, err)
	}
	return manifest
}

func readModule(t *testing.T, path string) *ast.Module {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read module %s: %v", path, err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse module %s: %v", path, err)
	}
	node, err := decodeNode(raw)
	if err != nil {
		t.Fatalf("decode module %s: %v", path, err)
	}
	mod, ok := node.(*ast.Module)
	if !ok {
		t.Fatalf("decoded node is not module: %T", node)
	}
	return mod
}

func decodeNode(node map[string]any) (ast.Node, error) {
	typ, _ := node["type"].(string)
	switch typ {
	case "Module":
		importsVal, _ := node["imports"].([]any)
		imports := make([]*ast.ImportStatement, 0, len(importsVal))
		for _, raw := range importsVal {
			child, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid import entry %T", raw)
			}
			imp, err := decodeImportStatement(child)
			if err != nil {
				return nil, err
			}
			imports = append(imports, imp)
		}
		bodyVal, _ := node["body"].([]any)
		stmts := make([]ast.Statement, 0, len(bodyVal))
		for _, raw := range bodyVal {
			child, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, err
			}
			stmt, ok := child.(ast.Statement)
			if !ok {
				return nil, fs.ErrInvalid
			}
			stmts = append(stmts, stmt)
		}
		var pkg *ast.PackageStatement
		if pkgNode, ok := node["package"].(map[string]any); ok {
			decoded, err := decodePackageStatement(pkgNode)
			if err != nil {
				return nil, err
			}
			pkg = decoded
		}
		return ast.NewModule(stmts, imports, pkg), nil
	case "StringLiteral":
		val, _ := node["value"].(string)
		return ast.NewStringLiteral(val), nil
	case "BooleanLiteral":
		val, _ := node["value"].(bool)
		return ast.NewBooleanLiteral(val), nil
	case "IntegerLiteral":
		bi := parseBigInt(node["value"])
		var suffixPtr *ast.IntegerType
		if s, ok := node["integerType"].(string); ok {
			stype := ast.IntegerType(s)
			suffixPtr = &stype
		}
		return ast.NewIntegerLiteral(bi, suffixPtr), nil
	case "FloatLiteral":
		val, _ := node["value"].(float64)
		var suffixPtr *ast.FloatType
		if s, ok := node["floatType"].(string); ok {
			stype := ast.FloatType(s)
			suffixPtr = &stype
		}
		return ast.NewFloatLiteral(val, suffixPtr), nil
	case "ArrayLiteral":
		elementsVal, _ := node["elements"].([]any)
		exprs := make([]ast.Expression, 0, len(elementsVal))
		for _, raw := range elementsVal {
			child, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, err
			}
			expr, ok := child.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid array element %T", child)
			}
			exprs = append(exprs, expr)
		}
		return ast.NewArrayLiteral(exprs), nil
	case "Identifier":
		name, _ := node["name"].(string)
		return ast.NewIdentifier(name), nil
	case "AssignmentExpression":
		op, _ := node["operator"].(string)
		leftNode, err := decodeNode(node["left"].(map[string]any))
		if err != nil {
			return nil, err
		}
		rightNode, err := decodeNode(node["right"].(map[string]any))
		if err != nil {
			return nil, err
		}
		left, ok := leftNode.(ast.AssignmentTarget)
		if !ok {
			return nil, fmt.Errorf("invalid assignment target %T", leftNode)
		}
		right, ok := rightNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid assignment expression right %T", rightNode)
		}
		return ast.NewAssignmentExpression(ast.AssignmentOperator(op), left, right), nil
	case "BlockExpression":
		bodyVal, _ := node["body"].([]any)
		stmts := make([]ast.Statement, 0, len(bodyVal))
		for _, raw := range bodyVal {
			child, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, err
			}
			stmt, ok := child.(ast.Statement)
			if !ok {
				return nil, fs.ErrInvalid
			}
			stmts = append(stmts, stmt)
		}
		return ast.NewBlockExpression(stmts), nil
	case "BinaryExpression":
		op, _ := node["operator"].(string)
		leftNode, err := decodeNode(node["left"].(map[string]any))
		if err != nil {
			return nil, err
		}
		rightNode, err := decodeNode(node["right"].(map[string]any))
		if err != nil {
			return nil, err
		}
		left, ok := leftNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid binary left %T", leftNode)
		}
		right, ok := rightNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid binary right %T", rightNode)
		}
		return ast.NewBinaryExpression(op, left, right), nil
	case "StringInterpolation":
		partsVal, _ := node["parts"].([]any)
		parts := make([]ast.Expression, 0, len(partsVal))
		for _, raw := range partsVal {
			childNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid interpolation part %T", raw)
			}
			decoded, err := decodeNode(childNode)
			if err != nil {
				return nil, err
			}
			expr, ok := decoded.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid interpolation expression %T", decoded)
			}
			parts = append(parts, expr)
		}
		return ast.NewStringInterpolation(parts), nil
	case "RangeExpression":
		startNode, err := decodeNode(node["start"].(map[string]any))
		if err != nil {
			return nil, err
		}
		endNode, err := decodeNode(node["end"].(map[string]any))
		if err != nil {
			return nil, err
		}
		start, ok := startNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid range start %T", startNode)
		}
		endExpr, ok := endNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid range end %T", endNode)
		}
		inclusive, _ := node["inclusive"].(bool)
		return ast.NewRangeExpression(start, endExpr, inclusive), nil
	case "MatchExpression":
		subjectNode, err := decodeNode(node["subject"].(map[string]any))
		if err != nil {
			return nil, err
		}
		subject, ok := subjectNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid match subject %T", subjectNode)
		}
		clausesVal, _ := node["clauses"].([]any)
		clauses := make([]*ast.MatchClause, 0, len(clausesVal))
		for _, raw := range clausesVal {
			clauseNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid match clause %T", raw)
			}
			clause, err := decodeMatchClause(clauseNode)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
		}
		return ast.NewMatchExpression(subject, clauses), nil
	case "PropagationExpression":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		expr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid propagation expression %T", exprNode)
		}
		return ast.NewPropagationExpression(expr), nil
	case "OrElseExpression":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		expression, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid or-else expression %T", exprNode)
		}
		handlerNode, err := decodeNode(node["handler"].(map[string]any))
		if err != nil {
			return nil, err
		}
		handler, ok := handlerNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("or-else handler must be block, got %T", handlerNode)
		}
		var binding *ast.Identifier
		if bindingRaw, ok := node["errorBinding"].(map[string]any); ok {
			decoded, err := decodeNode(bindingRaw)
			if err != nil {
				return nil, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid or-else binding %T", decoded)
			}
			binding = id
		}
		return ast.NewOrElseExpression(expression, handler, binding), nil
	case "EnsureExpression":
		tryNode, err := decodeNode(node["tryExpression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		tryExpr, ok := tryNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid ensure try expression %T", tryNode)
		}
		blockNode, err := decodeNode(node["ensureBlock"].(map[string]any))
		if err != nil {
			return nil, err
		}
		ensureBlock, ok := blockNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("ensure block must be block expression, got %T", blockNode)
		}
		return ast.NewEnsureExpression(tryExpr, ensureBlock), nil
	case "ProcExpression":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		expr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid proc expression body %T", exprNode)
		}
		return ast.NewProcExpression(expr), nil
	case "SpawnExpression":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		expr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid spawn expression body %T", exprNode)
		}
		return ast.NewSpawnExpression(expr), nil
	case "WhileLoop":
		condNode, err := decodeNode(node["condition"].(map[string]any))
		if err != nil {
			return nil, err
		}
		condition, ok := condNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid while condition %T", condNode)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, err
		}
		block, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("while body must be block expression, got %T", bodyNode)
		}
		return ast.NewWhileLoop(condition, block), nil
	case "ForLoop":
		pattern, err := decodePattern(node["pattern"].(map[string]any))
		if err != nil {
			return nil, err
		}
		iterNode, err := decodeNode(node["iterable"].(map[string]any))
		if err != nil {
			return nil, err
		}
		iterable, ok := iterNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid for-loop iterable %T", iterNode)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("for-loop body must be block expression, got %T", bodyNode)
		}
		return ast.NewForLoop(pattern, iterable, body), nil
	case "IfExpression":
		condNode, err := decodeNode(node["ifCondition"].(map[string]any))
		if err != nil {
			return nil, err
		}
		condition, ok := condNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid if condition %T", condNode)
		}
		bodyNode, err := decodeNode(node["ifBody"].(map[string]any))
		if err != nil {
			return nil, err
		}
		ifBody, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("if body must be block, got %T", bodyNode)
		}
		orClauseVals, _ := node["orClauses"].([]any)
		clauses := make([]*ast.OrClause, 0, len(orClauseVals))
		for _, raw := range orClauseVals {
			clauseNode := raw.(map[string]any)
			var clauseCond ast.Expression
			if condRaw, ok := clauseNode["condition"].(map[string]any); ok {
				exprNode, err := decodeNode(condRaw)
				if err != nil {
					return nil, err
				}
				expr, ok := exprNode.(ast.Expression)
				if !ok {
					return nil, fmt.Errorf("invalid or-clause condition %T", exprNode)
				}
				clauseCond = expr
			}
			bodyNode, err := decodeNode(clauseNode["body"].(map[string]any))
			if err != nil {
				return nil, err
			}
			body, ok := bodyNode.(*ast.BlockExpression)
			if !ok {
				return nil, fmt.Errorf("or-clause body must be block, got %T", bodyNode)
			}
			clauses = append(clauses, ast.NewOrClause(body, clauseCond))
		}
		return ast.NewIfExpression(condition, ifBody, clauses), nil
	case "FunctionCall":
		calleeNode, err := decodeNode(node["callee"].(map[string]any))
		if err != nil {
			return nil, err
		}
		callee, ok := calleeNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid callee %T", calleeNode)
		}
		argsVal, _ := node["arguments"].([]any)
		args := make([]ast.Expression, 0, len(argsVal))
		for _, raw := range argsVal {
			argMap, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid argument node %T", raw)
			}
			argNode, err := decodeNode(argMap)
			if err != nil {
				return nil, err
			}
			expr, ok := argNode.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid argument %T", argNode)
			}
			args = append(args, expr)
		}
		var typeArgs []ast.TypeExpression
		if taRaw, ok := node["typeArguments"].([]any); ok {
			typeArgs = make([]ast.TypeExpression, 0, len(taRaw))
			for _, raw := range taRaw {
				taNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid type argument %T", raw)
				}
				taExpr, err := decodeTypeExpression(taNode)
				if err != nil {
					return nil, err
				}
				typeArgs = append(typeArgs, taExpr)
			}
		}
		return ast.NewFunctionCall(callee, args, typeArgs, false), nil
	case "FunctionSignature":
		sig, err := decodeFunctionSignature(node)
		if err != nil {
			return nil, err
		}
		return sig, nil
	case "FunctionParameter":
		nameNode, err := decodeNode(node["name"].(map[string]any))
		if err != nil {
			return nil, err
		}
		pattern, ok := nameNode.(ast.Pattern)
		if !ok {
			return nil, fmt.Errorf("invalid function parameter name %T", nameNode)
		}
		var paramType ast.TypeExpression
		if typeRaw, ok := node["paramType"].(map[string]any); ok {
			typeExpr, err := decodeTypeExpression(typeRaw)
			if err != nil {
				return nil, err
			}
			paramType = typeExpr
		}
		return ast.NewFunctionParameter(pattern, paramType), nil
	case "FunctionDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid function identifier %T", idNode)
		}
		paramsVal, _ := node["params"].([]any)
		params := make([]*ast.FunctionParameter, 0, len(paramsVal))
		for _, raw := range paramsVal {
			paramNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid function parameter node %T", raw)
			}
			decoded, err := decodeNode(paramNode)
			if err != nil {
				return nil, err
			}
			param, ok := decoded.(*ast.FunctionParameter)
			if !ok {
				return nil, fmt.Errorf("invalid function parameter %T", decoded)
			}
			params = append(params, param)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("function body must be block, got %T", bodyNode)
		}
		var returnType ast.TypeExpression
		if retRaw, ok := node["returnType"].(map[string]any); ok {
			rt, err := decodeTypeExpression(retRaw)
			if err != nil {
				return nil, err
			}
			returnType = rt
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid generic parameter node %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, err
				}
				generics = append(generics, gp)
			}
		}
		isMethodShorthand, _ := node["isMethodShorthand"].(bool)
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewFunctionDefinition(id, params, body, returnType, generics, nil, isMethodShorthand, isPrivate), nil
	case "StructDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid struct identifier %T", idNode)
		}
		fieldsVal, _ := node["fields"].([]any)
		fields := make([]*ast.StructFieldDefinition, 0, len(fieldsVal))
		for _, raw := range fieldsVal {
			fieldNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid struct field %T", raw)
			}
			field, err := decodeStructFieldDefinition(fieldNode)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
		kindStr, _ := node["kind"].(string)
		kind := ast.StructKind(kindStr)
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewStructDefinition(id, fields, kind, nil, nil, isPrivate), nil
	case "MethodsDefinition":
		targetRaw, ok := node["targetType"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("methods definition missing target type")
		}
		targetType, err := decodeTypeExpression(targetRaw)
		if err != nil {
			return nil, err
		}
		defsVal, _ := node["definitions"].([]any)
		defs := make([]*ast.FunctionDefinition, 0, len(defsVal))
		for _, raw := range defsVal {
			defNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid methods definition entry %T", raw)
			}
			decoded, err := decodeNode(defNode)
			if err != nil {
				return nil, err
			}
			fn, ok := decoded.(*ast.FunctionDefinition)
			if !ok {
				return nil, fmt.Errorf("invalid methods definition function %T", decoded)
			}
			defs = append(defs, fn)
		}
		return ast.NewMethodsDefinition(targetType, defs, nil, nil), nil
	case "InterfaceDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid interface identifier %T", idNode)
		}
		sigsVal, _ := node["signatures"].([]any)
		signatures := make([]*ast.FunctionSignature, 0, len(sigsVal))
		for _, raw := range sigsVal {
			sigNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid interface signature %T", raw)
			}
			sig, err := decodeFunctionSignature(sigNode)
			if err != nil {
				return nil, err
			}
			signatures = append(signatures, sig)
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid interface generic %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, err
				}
				generics = append(generics, gp)
			}
		}
		var selfType ast.TypeExpression
		if selfRaw, ok := node["selfTypePattern"].(map[string]any); ok {
			typeExpr, err := decodeTypeExpression(selfRaw)
			if err != nil {
				return nil, err
			}
			selfType = typeExpr
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		var baseInterfaces []ast.TypeExpression
		if baseRaw, ok := node["baseInterfaces"].([]any); ok {
			baseInterfaces = make([]ast.TypeExpression, 0, len(baseRaw))
			for _, raw := range baseRaw {
				baseNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid base interface %T", raw)
				}
				typeExpr, err := decodeTypeExpression(baseNode)
				if err != nil {
					return nil, err
				}
				baseInterfaces = append(baseInterfaces, typeExpr)
			}
		}
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewInterfaceDefinition(id, signatures, generics, selfType, whereClause, baseInterfaces, isPrivate), nil
	case "StructLiteral":
		fieldsVal, _ := node["fields"].([]any)
		fields := make([]*ast.StructFieldInitializer, 0, len(fieldsVal))
		for _, raw := range fieldsVal {
			fieldNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid struct field initializer %T", raw)
			}
			field, err := decodeStructFieldInitializer(fieldNode)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
		var structType *ast.Identifier
		if typeRaw, ok := node["structType"].(map[string]any); ok {
			typeNode, err := decodeNode(typeRaw)
			if err != nil {
				return nil, err
			}
			id, ok := typeNode.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid struct type %T", typeNode)
			}
			structType = id
		}
		isPositional, _ := node["isPositional"].(bool)
		var update ast.Expression
		if updateRaw, ok := node["functionalUpdateSource"].(map[string]any); ok {
			updNode, err := decodeNode(updateRaw)
			if err != nil {
				return nil, err
			}
			expr, ok := updNode.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid functional update source %T", updNode)
			}
			update = expr
		}
		typeArgsVal, _ := node["typeArguments"].([]any)
		typeArgs := make([]ast.TypeExpression, 0, len(typeArgsVal))
		for _, raw := range typeArgsVal {
			typeNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid type argument %T", raw)
			}
			typeExpr, err := decodeTypeExpression(typeNode)
			if err != nil {
				return nil, err
			}
			typeArgs = append(typeArgs, typeExpr)
		}
		return ast.NewStructLiteral(fields, isPositional, structType, update, typeArgs), nil
	case "StructFieldDefinition":
		return decodeStructFieldDefinition(node)
	case "StructFieldInitializer":
		return decodeStructFieldInitializer(node)
	case "ContinueStatement":
		var label *ast.Identifier
		if labelRaw, ok := node["label"].(map[string]any); ok {
			decoded, err := decodeNode(labelRaw)
			if err != nil {
				return nil, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid continue label %T", decoded)
			}
			label = id
		}
		return ast.NewContinueStatement(label), nil
	case "ImportStatement":
		imp, err := decodeImportStatement(node)
		if err != nil {
			return nil, err
		}
		return imp, nil
	case "DynImportStatement":
		imp, err := decodeDynImportStatement(node)
		if err != nil {
			return nil, err
		}
		return imp, nil
	case "MemberAccessExpression":
		objectNode, err := decodeNode(node["object"].(map[string]any))
		if err != nil {
			return nil, err
		}
		object, ok := objectNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid member access object %T", objectNode)
		}
		memberNode, err := decodeNode(node["member"].(map[string]any))
		if err != nil {
			return nil, err
		}
		memberExpr, ok := memberNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid member expression %T", memberNode)
		}
		return ast.NewMemberAccessExpression(object, memberExpr), nil
	case "WildcardPattern", "LiteralPattern", "StructPattern", "ArrayPattern", "TypedPattern":
		return decodePattern(node)
	case "ReturnStatement":
		var argument ast.Expression
		if argRaw, ok := node["argument"].(map[string]any); ok {
			decoded, err := decodeNode(argRaw)
			if err != nil {
				return nil, err
			}
			expr, ok := decoded.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid return argument %T", decoded)
			}
			argument = expr
		}
		return ast.NewReturnStatement(argument), nil
	case "RaiseStatement":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		expr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid raise expression %T", exprNode)
		}
		return ast.NewRaiseStatement(expr), nil
	case "BreakStatement":
		var label *ast.Identifier
		if labelRaw, ok := node["label"].(map[string]any); ok {
			decoded, err := decodeNode(labelRaw)
			if err != nil {
				return nil, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid break label %T", decoded)
			}
			label = id
		}
		var value ast.Expression
		if valueRaw, ok := node["value"].(map[string]any); ok {
			decoded, err := decodeNode(valueRaw)
			if err != nil {
				return nil, err
			}
			expr, ok := decoded.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid break value %T", decoded)
			}
			value = expr
		}
		return ast.NewBreakStatement(label, value), nil
	case "RescueExpression":
		monNode, err := decodeNode(node["monitoredExpression"].(map[string]any))
		if err != nil {
			return nil, err
		}
		monExpr, ok := monNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid rescue expression %T", monNode)
		}
		clausesVal, _ := node["clauses"].([]any)
		clauses := make([]*ast.MatchClause, 0, len(clausesVal))
		for _, raw := range clausesVal {
			clauseNode, err := decodeMatchClause(raw.(map[string]any))
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clauseNode)
		}
		return ast.NewRescueExpression(monExpr, clauses), nil
	default:
		return nil, fs.ErrInvalid
	}
}

func decodePattern(node map[string]any) (ast.Pattern, error) {
	if node == nil {
		return nil, fmt.Errorf("pattern node is nil")
	}
	typ, _ := node["type"].(string)
	switch typ {
	case "Identifier":
		decoded, err := decodeNode(node)
		if err != nil {
			return nil, err
		}
		ident, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid identifier pattern %T", decoded)
		}
		return ident, nil
	case "WildcardPattern":
		return ast.NewWildcardPattern(), nil
	case "LiteralPattern":
		literalRaw, ok := node["literal"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("literal pattern missing literal")
		}
		decoded, err := decodeNode(literalRaw)
		if err != nil {
			return nil, err
		}
		literal, ok := decoded.(ast.Literal)
		if !ok {
			return nil, fmt.Errorf("invalid literal pattern literal %T", decoded)
		}
		return ast.NewLiteralPattern(literal), nil
	case "StructPattern":
		fieldsVal, _ := node["fields"].([]any)
		fields := make([]*ast.StructPatternField, 0, len(fieldsVal))
		for _, raw := range fieldsVal {
			fieldNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid struct pattern field %T", raw)
			}
			field, err := decodeStructPatternField(fieldNode)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
		var structType *ast.Identifier
		if structTypeRaw, ok := node["structType"].(map[string]any); ok {
			decoded, err := decodeNode(structTypeRaw)
			if err != nil {
				return nil, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid struct pattern type %T", decoded)
			}
			structType = id
		}
		isPositional, _ := node["isPositional"].(bool)
		return ast.NewStructPattern(fields, isPositional, structType), nil
	case "ArrayPattern":
		elementsVal, _ := node["elements"].([]any)
		elements := make([]ast.Pattern, 0, len(elementsVal))
		for _, raw := range elementsVal {
			elemNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid array pattern element %T", raw)
			}
			elem, err := decodePattern(elemNode)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
		}
		var rest ast.Pattern
		if restRaw, ok := node["restPattern"]; ok && restRaw != nil {
			restNode, ok := restRaw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid array rest pattern %T", restRaw)
			}
			decodedRest, err := decodePattern(restNode)
			if err != nil {
				return nil, err
			}
			rest = decodedRest
		}
		return ast.NewArrayPattern(elements, rest), nil
	case "TypedPattern":
		patternRaw, ok := node["pattern"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("typed pattern missing pattern")
		}
		pattern, err := decodePattern(patternRaw)
		if err != nil {
			return nil, err
		}
		annotationRaw, ok := node["typeAnnotation"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("typed pattern missing typeAnnotation")
		}
		annotation, err := decodeTypeExpression(annotationRaw)
		if err != nil {
			return nil, err
		}
		return ast.NewTypedPattern(pattern, annotation), nil
	default:
		return nil, fmt.Errorf("unsupported pattern type %s", typ)
	}
}

func decodeStructPatternField(node map[string]any) (*ast.StructPatternField, error) {
	patternRaw, ok := node["pattern"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("struct pattern field missing pattern")
	}
	pattern, err := decodePattern(patternRaw)
	if err != nil {
		return nil, err
	}
	var fieldName *ast.Identifier
	if fieldRaw, ok := node["fieldName"].(map[string]any); ok {
		decoded, err := decodeNode(fieldRaw)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid struct pattern field name %T", decoded)
		}
		fieldName = id
	}
	var binding *ast.Identifier
	if bindingRaw, ok := node["binding"].(map[string]any); ok {
		decoded, err := decodeNode(bindingRaw)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid struct pattern binding %T", decoded)
		}
		binding = id
	}
	return ast.NewStructPatternField(pattern, fieldName, binding), nil
}

func decodeTypeExpression(node map[string]any) (ast.TypeExpression, error) {
	if node == nil {
		return nil, fmt.Errorf("type expression node is nil")
	}
	typ, _ := node["type"].(string)
	switch typ {
	case "SimpleTypeExpression":
		nameRaw, ok := node["name"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("simple type missing name")
		}
		decoded, err := decodeNode(nameRaw)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid simple type name %T", decoded)
		}
		return ast.NewSimpleTypeExpression(id), nil
	case "GenericTypeExpression":
		baseRaw, ok := node["base"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("generic type missing base")
		}
		base, err := decodeTypeExpression(baseRaw)
		if err != nil {
			return nil, err
		}
		argsVal, _ := node["arguments"].([]any)
		args := make([]ast.TypeExpression, 0, len(argsVal))
		for _, raw := range argsVal {
			argNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid generic type argument %T", raw)
			}
			arg, err := decodeTypeExpression(argNode)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return ast.NewGenericTypeExpression(base, args), nil
	case "FunctionTypeExpression":
		paramsVal, _ := node["paramTypes"].([]any)
		params := make([]ast.TypeExpression, 0, len(paramsVal))
		for _, raw := range paramsVal {
			paramNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid function param type %T", raw)
			}
			param, err := decodeTypeExpression(paramNode)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
		returnRaw, ok := node["returnType"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("function type missing returnType")
		}
		returnType, err := decodeTypeExpression(returnRaw)
		if err != nil {
			return nil, err
		}
		return ast.NewFunctionTypeExpression(params, returnType), nil
	case "NullableTypeExpression":
		innerRaw, ok := node["innerType"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("nullable type missing innerType")
		}
		inner, err := decodeTypeExpression(innerRaw)
		if err != nil {
			return nil, err
		}
		return ast.NewNullableTypeExpression(inner), nil
	case "ResultTypeExpression":
		innerRaw, ok := node["innerType"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("result type missing innerType")
		}
		inner, err := decodeTypeExpression(innerRaw)
		if err != nil {
			return nil, err
		}
		return ast.NewResultTypeExpression(inner), nil
	case "UnionTypeExpression":
		membersVal, _ := node["members"].([]any)
		members := make([]ast.TypeExpression, 0, len(membersVal))
		for _, raw := range membersVal {
			memberNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid union member %T", raw)
			}
			member, err := decodeTypeExpression(memberNode)
			if err != nil {
				return nil, err
			}
			members = append(members, member)
		}
		return ast.NewUnionTypeExpression(members), nil
	case "WildcardTypeExpression":
		return ast.NewWildcardTypeExpression(), nil
	default:
		return nil, fmt.Errorf("unsupported type expression %s", typ)
	}
}

func decodeFunctionSignature(node map[string]any) (*ast.FunctionSignature, error) {
	nameNode, err := decodeNode(node["name"].(map[string]any))
	if err != nil {
		return nil, err
	}
	id, ok := nameNode.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("invalid function signature name %T", nameNode)
	}
	paramsVal, _ := node["params"].([]any)
	params := make([]*ast.FunctionParameter, 0, len(paramsVal))
	for _, raw := range paramsVal {
		paramNode, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid function signature parameter %T", raw)
		}
		decoded, err := decodeNode(paramNode)
		if err != nil {
			return nil, err
		}
		param, ok := decoded.(*ast.FunctionParameter)
		if !ok {
			return nil, fmt.Errorf("invalid function signature parameter %T", decoded)
		}
		params = append(params, param)
	}
	var returnType ast.TypeExpression
	if retRaw, ok := node["returnType"].(map[string]any); ok {
		rt, err := decodeTypeExpression(retRaw)
		if err != nil {
			return nil, err
		}
		returnType = rt
	}
	var generics []*ast.GenericParameter
	if gpRaw, ok := node["genericParams"].([]any); ok {
		generics = make([]*ast.GenericParameter, 0, len(gpRaw))
		for _, raw := range gpRaw {
			gpNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid signature generic %T", raw)
			}
			gp, err := decodeGenericParameter(gpNode)
			if err != nil {
				return nil, err
			}
			generics = append(generics, gp)
		}
	}
	var whereClause []*ast.WhereClauseConstraint
	if wcRaw, ok := node["whereClause"].([]any); ok {
		whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
		for _, raw := range wcRaw {
			wcNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid signature where clause %T", raw)
			}
			wc, err := decodeWhereClauseConstraint(wcNode)
			if err != nil {
				return nil, err
			}
			whereClause = append(whereClause, wc)
		}
	}
	var defaultImpl *ast.BlockExpression
	if defRaw, ok := node["defaultImpl"].(map[string]any); ok {
		decoded, err := decodeNode(defRaw)
		if err != nil {
			return nil, err
		}
		block, ok := decoded.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("invalid default implementation %T", decoded)
		}
		defaultImpl = block
	}
	return ast.NewFunctionSignature(id, params, returnType, generics, whereClause, defaultImpl), nil
}

func decodeWhereClauseConstraint(node map[string]any) (*ast.WhereClauseConstraint, error) {
	typeParamNode, err := decodeNode(node["typeParam"].(map[string]any))
	if err != nil {
		return nil, err
	}
	typeParam, ok := typeParamNode.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("invalid where clause type param %T", typeParamNode)
	}
	constraintsRaw, _ := node["constraints"].([]any)
	constraints := make([]*ast.InterfaceConstraint, 0, len(constraintsRaw))
	for _, raw := range constraintsRaw {
		consNode, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid where clause constraint %T", raw)
		}
		cons, err := decodeInterfaceConstraint(consNode)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, cons)
	}
	return ast.NewWhereClauseConstraint(typeParam, constraints), nil
}

func assertResult(t *testing.T, dir string, manifest fixtureManifest, result runtime.Value, stdout []string) {
	t.Helper()
	if manifest.Expect.Result == nil {
		if len(manifest.Expect.Stdout) > 0 {
			if !reflect.DeepEqual(stdout, manifest.Expect.Stdout) {
				t.Fatalf("fixture %s expected stdout %v, got %v", dir, manifest.Expect.Stdout, stdout)
			}
		}
		return
	}
	exp := manifest.Expect.Result
	switch exp.Kind {
	case "string":
		v, ok := result.(runtime.StringValue)
		if !ok {
			t.Fatalf("fixture %s expected string, got %T", dir, result)
		}
		if exp.Value != nil && v.Val != exp.Value {
			t.Fatalf("fixture %s expected value %v, got %v", dir, exp.Value, v.Val)
		}
	case "bool":
		v, ok := result.(runtime.BoolValue)
		if !ok {
			t.Fatalf("fixture %s expected bool, got %T", dir, result)
		}
		if exp.Value != nil {
			expected, _ := exp.Value.(bool)
			if v.Val != expected {
				t.Fatalf("fixture %s expected value %v, got %v", dir, exp.Value, v.Val)
			}
		}
	case "i32", "i64", "i16", "i8", "u32", "u16", "u8", "u64", "u128", "i128":
		v, ok := result.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("fixture %s expected integer, got %T", dir, result)
		}
		if string(v.TypeSuffix) != exp.Kind {
			t.Fatalf("fixture %s expected integer type %s, got %s", dir, exp.Kind, v.TypeSuffix)
		}
		if exp.Value != nil {
			expected := parseBigInt(exp.Value)
			if v.Val.Cmp(expected) != 0 {
				t.Fatalf("fixture %s expected value %v, got %v", dir, expected, v.Val)
			}
		}
	case "f32", "f64":
		v, ok := result.(runtime.FloatValue)
		if !ok {
			t.Fatalf("fixture %s expected float, got %T", dir, result)
		}
		if string(v.TypeSuffix) != exp.Kind {
			t.Fatalf("fixture %s expected float type %s, got %s", dir, exp.Kind, v.TypeSuffix)
		}
		if exp.Value != nil {
			expected, _ := exp.Value.(float64)
			if v.Val != expected {
				t.Fatalf("fixture %s expected value %v, got %v", dir, exp.Value, v.Val)
			}
		}
	default:
		if result.Kind().String() != exp.Kind {
			t.Fatalf("fixture %s expected kind %s, got %s", dir, exp.Kind, result.Kind())
		}
	}
	if len(manifest.Expect.Stdout) > 0 {
		if !reflect.DeepEqual(stdout, manifest.Expect.Stdout) {
			t.Fatalf("fixture %s expected stdout %v, got %v", dir, manifest.Expect.Stdout, stdout)
		}
	}
}

func decodeGenericParameter(node map[string]any) (*ast.GenericParameter, error) {
	nameNode, ok := node["name"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("generic parameter missing name")
	}
	decodedName, err := decodeNode(nameNode)
	if err != nil {
		return nil, err
	}
	id, ok := decodedName.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("invalid generic parameter name %T", decodedName)
	}
	constraintsRaw, _ := node["constraints"].([]any)
	constraints := make([]*ast.InterfaceConstraint, 0, len(constraintsRaw))
	for _, raw := range constraintsRaw {
		constraintNode, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid interface constraint %T", raw)
		}
		constraint, err := decodeInterfaceConstraint(constraintNode)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return ast.NewGenericParameter(id, constraints), nil
}

func decodeInterfaceConstraint(node map[string]any) (*ast.InterfaceConstraint, error) {
	typeNode, ok := node["interfaceType"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("interface constraint missing interfaceType")
	}
	ifaceType, err := decodeTypeExpression(typeNode)
	if err != nil {
		return nil, err
	}
	return ast.NewInterfaceConstraint(ifaceType), nil
}

func parseBigInt(value interface{}) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}
	switch v := value.(type) {
	case float64:
		return big.NewInt(int64(v))
	case int64:
		return big.NewInt(v)
	case string:
		if bi, ok := new(big.Int).SetString(v, 10); ok {
			return bi
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return big.NewInt(i)
		}
	}
	return big.NewInt(0)
}

func decodeStructFieldDefinition(node map[string]any) (*ast.StructFieldDefinition, error) {
	fieldTypeRaw, ok := node["fieldType"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("struct field missing fieldType")
	}
	fieldType, err := decodeTypeExpression(fieldTypeRaw)
	if err != nil {
		return nil, err
	}
	var name *ast.Identifier
	if nameRaw, ok := node["name"].(map[string]any); ok {
		decoded, err := decodeNode(nameRaw)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid struct field name %T", decoded)
		}
		name = id
	}
	return ast.NewStructFieldDefinition(fieldType, name), nil
}

func decodeStructFieldInitializer(node map[string]any) (*ast.StructFieldInitializer, error) {
	valueNode, ok := node["value"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("struct field initializer missing value")
	}
	decodedValue, err := decodeNode(valueNode)
	if err != nil {
		return nil, err
	}
	valueExpr, ok := decodedValue.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("invalid struct field value %T", decodedValue)
	}
	var name *ast.Identifier
	if nameRaw, ok := node["name"].(map[string]any); ok {
		decodedName, err := decodeNode(nameRaw)
		if err != nil {
			return nil, err
		}
		id, ok := decodedName.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid struct field initializer name %T", decodedName)
		}
		name = id
	}
	isShorthand, _ := node["isShorthand"].(bool)
	return ast.NewStructFieldInitializer(valueExpr, name, isShorthand), nil
}

func decodePackageStatement(node map[string]any) (*ast.PackageStatement, error) {
	namePathVal, _ := node["namePath"].([]any)
	namePath, err := decodeIdentifierList(namePathVal)
	if err != nil {
		return nil, err
	}
	isPrivate, _ := node["isPrivate"].(bool)
	return ast.NewPackageStatement(namePath, isPrivate), nil
}

func decodeImportStatement(node map[string]any) (*ast.ImportStatement, error) {
	pathVal, _ := node["packagePath"].([]any)
	packagePath, err := decodeIdentifierList(pathVal)
	if err != nil {
		return nil, err
	}
	isWildcard, _ := node["isWildcard"].(bool)
	selectorsVal, _ := node["selectors"].([]any)
	selectors := make([]*ast.ImportSelector, 0, len(selectorsVal))
	for _, raw := range selectorsVal {
		selectorNode, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid import selector %T", raw)
		}
		selector, err := decodeImportSelector(selectorNode)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selector)
	}
	var alias *ast.Identifier
	if aliasNode, ok := node["alias"].(map[string]any); ok {
		decoded, err := decodeNode(aliasNode)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid import alias %T", decoded)
		}
		alias = id
	}
	return ast.NewImportStatement(packagePath, isWildcard, selectors, alias), nil
}

func decodeDynImportStatement(node map[string]any) (*ast.DynImportStatement, error) {
	pathVal, _ := node["packagePath"].([]any)
	packagePath, err := decodeIdentifierList(pathVal)
	if err != nil {
		return nil, err
	}
	isWildcard, _ := node["isWildcard"].(bool)
	selectorsVal, _ := node["selectors"].([]any)
	selectors := make([]*ast.ImportSelector, 0, len(selectorsVal))
	for _, raw := range selectorsVal {
		selectorNode, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid dynimport selector %T", raw)
		}
		selector, err := decodeImportSelector(selectorNode)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selector)
	}
	var alias *ast.Identifier
	if aliasNode, ok := node["alias"].(map[string]any); ok {
		decoded, err := decodeNode(aliasNode)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid dynimport alias %T", decoded)
		}
		alias = id
	}
	return ast.NewDynImportStatement(packagePath, isWildcard, selectors, alias), nil
}

func decodeImportSelector(node map[string]any) (*ast.ImportSelector, error) {
	nameNode, ok := node["name"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("import selector missing name")
	}
	decodedName, err := decodeNode(nameNode)
	if err != nil {
		return nil, err
	}
	name, ok := decodedName.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("invalid import selector name %T", decodedName)
	}
	var alias *ast.Identifier
	if aliasNode, ok := node["alias"].(map[string]any); ok {
		decodedAlias, err := decodeNode(aliasNode)
		if err != nil {
			return nil, err
		}
		id, ok := decodedAlias.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid import selector alias %T", decodedAlias)
		}
		alias = id
	}
	return ast.NewImportSelector(name, alias), nil
}

func decodeIdentifierList(values []any) ([]*ast.Identifier, error) {
	ids := make([]*ast.Identifier, 0, len(values))
	for _, raw := range values {
		node, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid identifier entry %T", raw)
		}
		decoded, err := decodeNode(node)
		if err != nil {
			return nil, err
		}
		id, ok := decoded.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid identifier node %T", decoded)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func decodeMatchClause(node map[string]any) (*ast.MatchClause, error) {
	pattern, err := decodePattern(node["pattern"].(map[string]any))
	if err != nil {
		return nil, err
	}
	bodyNode, err := decodeNode(node["body"].(map[string]any))
	if err != nil {
		return nil, err
	}
	body, ok := bodyNode.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("invalid match clause body %T", bodyNode)
	}
	var guard ast.Expression
	if guardRaw, ok := node["guard"].(map[string]any); ok {
		guardNode, err := decodeNode(guardRaw)
		if err != nil {
			return nil, err
		}
		guardExpr, ok := guardNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid match clause guard %T", guardNode)
		}
		guard = guardExpr
	}
	return ast.NewMatchClause(pattern, body, guard), nil
}

func registerPrint(interp *Interpreter, buffer *[]string) {
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			var parts []string
			for _, arg := range args {
				parts = append(parts, formatRuntimeValue(arg))
			}
			*buffer = append(*buffer, strings.Join(parts, " "))
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}

func formatRuntimeValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}

func extractErrorMessage(err error) string {
	if rs, ok := err.(raiseSignal); ok {
		return rs.Error()
	}
	return err.Error()
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
