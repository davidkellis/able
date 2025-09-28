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
	Description string `json:"description"`
	Entry       string `json:"entry"`
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
		return ast.NewModule(stmts, nil, nil), nil
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
			argNode, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, err
			}
			expr, ok := argNode.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid argument %T", argNode)
			}
			args = append(args, expr)
		}
		return ast.NewFunctionCall(callee, args, nil, false), nil
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
