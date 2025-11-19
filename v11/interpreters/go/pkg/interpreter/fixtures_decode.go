package interpreter

import (
	"fmt"
	"reflect"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

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
	var inferredGenerics []*ast.GenericParameter
	if igRaw, ok := node["inferredGenericParams"].([]any); ok {
		inferredGenerics = make([]*ast.GenericParameter, 0, len(igRaw))
		for _, raw := range igRaw {
			gpNode, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid signature inferred generic %T", raw)
			}
			gp, err := decodeGenericParameter(gpNode)
			if err != nil {
				return nil, err
			}
			inferredGenerics = append(inferredGenerics, gp)
		}
	}
	sig := ast.NewFunctionSignature(id, params, returnType, generics, whereClause, defaultImpl)
	if len(inferredGenerics) > 0 {
		sig.InferredGenericParams = attachInferredGenericParams(inferredGenerics, generics)
	}
	return sig, nil
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

func assertResult(t testingT, dir string, manifest fixtureManifest, result runtime.Value, stdout []string) {
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
	param := ast.NewGenericParameter(id, constraints)
	if inferred, ok := node["isInferred"].(bool); ok {
		param.IsInferred = inferred
	}
	return param, nil
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
		if item == value || strings.Contains(value, item) {
			return true
		}
	}
	return false
}
