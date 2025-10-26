package interpreter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"

	"able/interpreter10-go/pkg/ast"
)

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
	case "PreludeStatement":
		target, _ := node["target"].(string)
		code, _ := node["code"].(string)
		return ast.NewPreludeStatement(ast.HostTarget(target), code), nil
	case "ExternFunctionBody":
		sigRaw, ok := node["signature"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("extern function body missing signature")
		}
		sigNode, err := decodeNode(sigRaw)
		if err != nil {
			return nil, err
		}
		signature, ok := sigNode.(*ast.FunctionDefinition)
		if !ok {
			return nil, fmt.Errorf("invalid extern signature %T", sigNode)
		}
		target, _ := node["target"].(string)
		body, _ := node["body"].(string)
		return ast.NewExternFunctionBody(ast.HostTarget(target), signature, body), nil
	case "StringLiteral":
		val, _ := node["value"].(string)
		return ast.NewStringLiteral(val), nil
	case "CharLiteral":
		val, _ := node["value"].(string)
		return ast.NewCharLiteral(val), nil
	case "NilLiteral":
		return ast.NewNilLiteral(), nil
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
	case "UnaryExpression":
		op, _ := node["operator"].(string)
		operandNode, ok := node["operand"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unary expression missing operand")
		}
		decoded, err := decodeNode(operandNode)
		if err != nil {
			return nil, err
		}
		expr, ok := decoded.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid unary operand %T", decoded)
		}
		return ast.NewUnaryExpression(ast.UnaryOperator(op), expr), nil

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
	case "LambdaExpression":
		paramsRaw, _ := node["params"].([]any)
		params := make([]*ast.FunctionParameter, 0, len(paramsRaw))
		for _, raw := range paramsRaw {
			paramNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid lambda parameter %T", raw)
			}
			decoded, err := decodeNode(paramNode)
			if err != nil {
				return nil, err
			}
			param, ok := decoded.(*ast.FunctionParameter)
			if !ok {
				return nil, fmt.Errorf("invalid lambda parameter %T", decoded)
			}
			params = append(params, param)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, err
		}
		bodyExpr, ok := bodyNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid lambda body %T", bodyNode)
		}
		var returnType ast.TypeExpression
		if rtRaw, ok := node["returnType"].(map[string]any); ok {
			RT, err := decodeTypeExpression(rtRaw)
			if err != nil {
				return nil, err
			}
			returnType = RT
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid lambda generic param %T", raw)
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
					return nil, fmt.Errorf("invalid lambda where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		isVerbose, _ := node["isVerboseSyntax"].(bool)
		return ast.NewLambdaExpression(params, bodyExpr, returnType, generics, whereClause, isVerbose), nil
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
	case "ImplementationDefinition":
		ifaceNode, err := decodeNode(node["interfaceName"].(map[string]any))
		if err != nil {
			return nil, err
		}
		ifaceID, ok := ifaceNode.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid implementation interface name %T", ifaceNode)
		}
		targetRaw, ok := node["targetType"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("implementation definition missing target type")
		}
		targetType, err := decodeTypeExpression(targetRaw)
		if err != nil {
			return nil, err
		}
		defsVal, _ := node["definitions"].([]any)
		defs := make([]*ast.FunctionDefinition, 0, len(defsVal))
		for _, raw := range defsVal {
			fnNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid implementation function %T", raw)
			}
			decoded, err := decodeNode(fnNode)
			if err != nil {
				return nil, err
			}
			fn, ok := decoded.(*ast.FunctionDefinition)
			if !ok {
				return nil, fmt.Errorf("invalid implementation function %T", decoded)
			}
			defs = append(defs, fn)
		}
		var implName *ast.Identifier
		if implRaw, ok := node["implName"].(map[string]any); ok {
			decoded, err := decodeNode(implRaw)
			if err != nil {
				return nil, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid implementation name %T", decoded)
			}
			implName = id
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid implementation generic param %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, err
				}
				generics = append(generics, gp)
			}
		}
		var interfaceArgs []ast.TypeExpression
		if iaRaw, ok := node["interfaceArgs"].([]any); ok {
			interfaceArgs = make([]ast.TypeExpression, 0, len(iaRaw))
			for _, raw := range iaRaw {
				argNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid implementation interface arg %T", raw)
				}
				expr, err := decodeTypeExpression(argNode)
				if err != nil {
					return nil, err
				}
				interfaceArgs = append(interfaceArgs, expr)
			}
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid implementation where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewImplementationDefinition(ifaceID, targetType, defs, implName, generics, interfaceArgs, whereClause, isPrivate), nil
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
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid struct generic param %T", raw)
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
					return nil, fmt.Errorf("invalid struct where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		kindStr, _ := node["kind"].(string)
		kind := ast.StructKind(kindStr)
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewStructDefinition(id, fields, kind, generics, whereClause, isPrivate), nil
	case "UnionDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("invalid union identifier %T", idNode)
		}
		variantsVal, _ := node["variants"].([]any)
		variants := make([]ast.TypeExpression, 0, len(variantsVal))
		for _, raw := range variantsVal {
			variantNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid union variant %T", raw)
			}
			variant, err := decodeTypeExpression(variantNode)
			if err != nil {
				return nil, err
			}
			variants = append(variants, variant)
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid union generic param %T", raw)
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
					return nil, fmt.Errorf("invalid union where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewUnionDefinition(id, variants, generics, whereClause, isPrivate), nil

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
	case "IteratorLiteral":
		bodyVal, _ := node["body"].([]any)
		body := make([]ast.Statement, 0, len(bodyVal))
		for _, raw := range bodyVal {
			stmtNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid iterator body statement %T", raw)
			}
			decoded, err := decodeNode(stmtNode)
			if err != nil {
				return nil, err
			}
			stmt, ok := decoded.(ast.Statement)
			if !ok {
				return nil, fmt.Errorf("invalid iterator body node %T", decoded)
			}
			body = append(body, stmt)
		}
		return ast.NewIteratorLiteral(body), nil

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
	case "ImplicitMemberExpression":
		memberNode, err := decodeNode(node["member"].(map[string]any))
		if err != nil {
			return nil, err
		}
		ident, ok := memberNode.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("implicit member expects identifier, got %T", memberNode)
		}
		return ast.NewImplicitMemberExpression(ident), nil
	case "IndexExpression":
		objectNode, err := decodeNode(node["object"].(map[string]any))
		if err != nil {
			return nil, err
		}
		object, ok := objectNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid index object %T", objectNode)
		}
		indexNode, err := decodeNode(node["index"].(map[string]any))
		if err != nil {
			return nil, err
		}
		indexExpr, ok := indexNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid index expression %T", indexNode)
		}
		return ast.NewIndexExpression(object, indexExpr), nil
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
	case "BreakpointExpression":
		var label *ast.Identifier
		if labelRaw, ok := node["label"].(map[string]any); ok {
			decoded, err := decodeNode(labelRaw)
			if err != nil {
				return nil, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("invalid breakpoint label %T", decoded)
			}
			label = id
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, fmt.Errorf("breakpoint body must be block expression, got %T", bodyNode)
		}
		return ast.NewBreakpointExpression(label, body), nil
	case "YieldStatement":
		var expr ast.Expression
		if exprRaw, ok := node["expression"].(map[string]any); ok {
			decoded, err := decodeNode(exprRaw)
			if err != nil {
				return nil, err
			}
			exprVal, ok := decoded.(ast.Expression)
			if !ok {
				return nil, fmt.Errorf("invalid yield expression %T", decoded)
			}
			expr = exprVal
		}
		return ast.NewYieldStatement(expr), nil

	case "RethrowStatement":
		return ast.NewRethrowStatement(), nil
	default:
		return nil, fs.ErrInvalid
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
