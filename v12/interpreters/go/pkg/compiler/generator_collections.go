package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func memberAccessMismatchReason(objectExpr string, objectType string, memberName string, fieldType string, expected string) string {
	return fmt.Sprintf("member access type mismatch: expr=%s object=%s member=%s field=%s expected=%s", objectExpr, objectType, memberName, fieldType, expected)
}

func (g *generator) compileArrayLiteral(ctx *compileContext, lit *ast.ArrayLiteral, expected string) ([]string, string, string, bool) {
	if lit == nil {
		ctx.setReason("missing array literal")
		return nil, "", "", false
	}
	monoSpec, monoExpected := g.monoArraySpecForGoType(expected)
	var returnType string
	switch {
	case monoExpected:
		returnType = monoSpec.GoType
	case expected == "" && g.isArrayStructType("*Array"):
		// Return *Array so assignments can detect the struct type and set OriginGoType.
		returnType = "*Array"
	case expected == "" || expected == "runtime.Value" || expected == "any":
		returnType = "runtime.Value"
	default:
		if g.typeCategory(expected) != "struct" {
			ctx.setReason("array literal type mismatch")
			return nil, "", "", false
		}
		baseName, ok := g.structBaseName(expected)
		if !ok || baseName != "Array" {
			ctx.setReason("array literal type mismatch")
			return nil, "", "", false
		}
		returnType = expected
	}
	elementExprs := make([]string, 0, len(lit.Elements))
	elementTypes := make([]string, 0, len(lit.Elements))
	var elementLines []string
	for _, element := range lit.Elements {
		elemExpected := ""
		if monoExpected && monoSpec != nil {
			elemExpected = monoSpec.ElemGoType
		}
		elLines, expr, goType, ok := g.compileExprLines(ctx, element, elemExpected)
		if !ok {
			return nil, "", "", false
		}
		elementLines = append(elementLines, elLines...)
		elementExprs = append(elementExprs, expr)
		elementTypes = append(elementTypes, goType)
	}
	if !monoExpected && expected == "" {
		if inferredSpec, ok := g.inferMonoArraySpecForElementTypes(elementTypes); ok && inferredSpec != nil {
			monoSpec = inferredSpec
			monoExpected = true
			returnType = inferredSpec.GoType
		}
	}
	if monoExpected && monoSpec != nil {
		lines := append([]string{}, elementLines...)
		arrTemp := ctx.newTemp()
		sliceExpr := fmt.Sprintf("[]%s{%s}", monoSpec.ElemGoType, strings.Join(elementExprs, ", "))
		arrExpr := fmt.Sprintf("&%s{Length: int32(%d), Capacity: int32(%d), Storage_handle: int64(0), Elements: %s}", monoSpec.GoName, len(elementExprs), len(elementExprs), sliceExpr)
		lines = append(lines, fmt.Sprintf("%s := %s", arrTemp, arrExpr))
		lines = append(lines, g.staticArraySyncCall(returnType, arrTemp))
		return lines, arrTemp, returnType, true
	}
	// Build []runtime.Value slice directly — no interpreter array store.
	lines := append([]string{}, elementLines...)
	valueExprs := make([]string, 0, len(elementExprs))
	for idx, expr := range elementExprs {
		goType := elementTypes[idx]
		if goType == "runtime.Value" {
			valueExprs = append(valueExprs, expr)
			continue
		}
		elemConvLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, goType)
		if !ok {
			ctx.setReason("array literal element unsupported")
			return nil, "", "", false
		}
		lines = append(lines, elemConvLines...)
		valueExprs = append(valueExprs, valueExpr)
	}
	sliceExpr := fmt.Sprintf("[]runtime.Value{%s}", strings.Join(valueExprs, ", "))
	if returnType == "runtime.Value" {
		arrTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := &runtime.ArrayValue{Elements: %s}", arrTemp, sliceExpr))
		return lines, arrTemp, "runtime.Value", true
	}
	arrTemp := ctx.newTemp()
	arrExpr := fmt.Sprintf("&Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: int64(0), Elements: %s}", len(valueExprs), len(valueExprs), sliceExpr)
	if !strings.HasPrefix(returnType, "*") {
		arrExpr = fmt.Sprintf("Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: int64(0), Elements: %s}", len(valueExprs), len(valueExprs), sliceExpr)
	}
	lines = append(lines, fmt.Sprintf("%s := %s", arrTemp, arrExpr))
	return lines, arrTemp, returnType, true
}

func (g *generator) compileMapLiteral(ctx *compileContext, lit *ast.MapLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing map literal")
		return "", "", false
	}
	hashMapType, ok := g.nativeStructCarrierType(ctx.packageName, "HashMap")
	if !ok {
		ctx.setReason("map literal type mismatch")
		return "", "", false
	}
	returnType := hashMapType
	if expected != "" && expected != "runtime.Value" && expected != "any" {
		if expectedInfo := g.structInfoByGoName(expected); expectedInfo != nil && expectedInfo.Name == "HashMap" {
			returnType = expected
		} else if baseName, ok := g.structBaseName(expected); ok && baseName == "HashMap" {
			returnType = expected
		}
	}
	hashMapBase := strings.TrimPrefix(returnType, "*")
	type mapElement struct {
		kind   string
		key    string
		value  string
		spread string
	}
	elements := make([]mapElement, 0, len(lit.Elements))
	for _, element := range lit.Elements {
		switch entry := element.(type) {
		case *ast.MapLiteralEntry:
			if entry == nil || entry.Key == nil || entry.Value == nil {
				ctx.setReason("unsupported map literal entry")
				return "", "", false
			}
			keyExpr, keyType, ok := g.compileExpr(ctx, entry.Key, "")
			if !ok {
				return "", "", false
			}
			keyRuntime, ok := g.runtimeValueExpr(keyExpr, keyType)
			if !ok {
				ctx.setReason("map literal key unsupported")
				return "", "", false
			}
			valueExpr, valueType, ok := g.compileExpr(ctx, entry.Value, "")
			if !ok {
				return "", "", false
			}
			valueRuntime, ok := g.runtimeValueExpr(valueExpr, valueType)
			if !ok {
				ctx.setReason("map literal value unsupported")
				return "", "", false
			}
			elements = append(elements, mapElement{kind: "entry", key: keyRuntime, value: valueRuntime})
		case *ast.MapLiteralSpread:
			if entry == nil || entry.Expression == nil {
				ctx.setReason("unsupported map literal spread")
				return "", "", false
			}
			spreadExpr, spreadType, ok := g.compileExpr(ctx, entry.Expression, "")
			if !ok {
				return "", "", false
			}
			spreadRuntime, ok := g.runtimeValueExpr(spreadExpr, spreadType)
			if !ok {
				ctx.setReason("map literal spread unsupported")
				return "", "", false
			}
			elements = append(elements, mapElement{kind: "spread", spread: spreadRuntime})
		default:
			ctx.setReason("unsupported map literal element")
			return "", "", false
		}
	}
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("func() %s {\n", returnType))
	buf.WriteString("\tif __able_runtime == nil {\n")
	buf.WriteString("\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tvar keyType ast.TypeExpression\n")
	buf.WriteString("\tvar valueType ast.TypeExpression\n")
	buf.WriteString("\tvar typeExprEqual func(a, b ast.TypeExpression) bool\n")
	buf.WriteString("\ttypeExprEqual = func(a, b ast.TypeExpression) bool {\n")
	buf.WriteString("\t\tswitch ta := a.(type) {\n")
	buf.WriteString("\t\tcase nil:\n")
	buf.WriteString("\t\t\treturn b == nil\n")
	buf.WriteString("\t\tcase *ast.SimpleTypeExpression:\n")
	buf.WriteString("\t\t\tother, ok := b.(*ast.SimpleTypeExpression)\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif ta.Name == nil || other.Name == nil {\n")
	buf.WriteString("\t\t\t\treturn ta.Name == other.Name\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ta.Name.Name == other.Name.Name\n")
	buf.WriteString("\t\tcase *ast.GenericTypeExpression:\n")
	buf.WriteString("\t\t\tother, ok := b.(*ast.GenericTypeExpression)\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif !typeExprEqual(ta.Base, other.Base) {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif len(ta.Arguments) != len(other.Arguments) {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tfor idx := range ta.Arguments {\n")
	buf.WriteString("\t\t\t\tif !typeExprEqual(ta.Arguments[idx], other.Arguments[idx]) {\n")
	buf.WriteString("\t\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn true\n")
	buf.WriteString("\t\tcase *ast.NullableTypeExpression:\n")
	buf.WriteString("\t\t\tother, ok := b.(*ast.NullableTypeExpression)\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn typeExprEqual(ta.InnerType, other.InnerType)\n")
	buf.WriteString("\t\tcase *ast.ResultTypeExpression:\n")
	buf.WriteString("\t\t\tother, ok := b.(*ast.ResultTypeExpression)\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn typeExprEqual(ta.InnerType, other.InnerType)\n")
	buf.WriteString("\t\tcase *ast.UnionTypeExpression:\n")
	buf.WriteString("\t\t\tother, ok := b.(*ast.UnionTypeExpression)\n")
	buf.WriteString("\t\t\tif !ok || len(ta.Members) != len(other.Members) {\n")
	buf.WriteString("\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tfor idx := range ta.Members {\n")
	buf.WriteString("\t\t\t\tif !typeExprEqual(ta.Members[idx], other.Members[idx]) {\n")
	buf.WriteString("\t\t\t\t\treturn false\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn true\n")
	buf.WriteString("\t\tcase *ast.WildcardTypeExpression:\n")
	buf.WriteString("\t\t\t_, ok := b.(*ast.WildcardTypeExpression)\n")
	buf.WriteString("\t\t\treturn ok\n")
	buf.WriteString("\t\tdefault:\n")
	buf.WriteString("\t\t\treturn false\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tmergeType := func(current, next ast.TypeExpression) ast.TypeExpression {\n")
	buf.WriteString("\t\tif current == nil {\n")
	buf.WriteString("\t\t\treturn next\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif next == nil {\n")
	buf.WriteString("\t\t\treturn current\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif _, ok := current.(*ast.WildcardTypeExpression); ok {\n")
	buf.WriteString("\t\t\treturn current\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif _, ok := next.(*ast.WildcardTypeExpression); ok {\n")
	buf.WriteString("\t\t\treturn next\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif typeExprEqual(current, next) {\n")
	buf.WriteString("\t\t\treturn current\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\treturn ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tvar typeFromValue func(val runtime.Value) ast.TypeExpression\n")
	buf.WriteString("\ttypeFromValue = func(val runtime.Value) ast.TypeExpression {\n")
	buf.WriteString("\t\tif val == nil {\n")
	buf.WriteString("\t\t\treturn nil\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tswitch v := val.(type) {\n")
	buf.WriteString("\t\tcase runtime.StringValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"String\")\n")
	buf.WriteString("\t\tcase *runtime.StringValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"String\")\n")
	buf.WriteString("\t\tcase runtime.BoolValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"bool\")\n")
	buf.WriteString("\t\tcase *runtime.BoolValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"bool\")\n")
	buf.WriteString("\t\tcase runtime.CharValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"char\")\n")
	buf.WriteString("\t\tcase *runtime.CharValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"char\")\n")
	buf.WriteString("\t\tcase runtime.NilValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"nil\")\n")
	buf.WriteString("\t\tcase *runtime.NilValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"nil\")\n")
	buf.WriteString("\t\tcase runtime.VoidValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"void\")\n")
	buf.WriteString("\t\tcase *runtime.VoidValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(\"void\")\n")
	buf.WriteString("\t\tcase runtime.IntegerValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(string(v.TypeSuffix))\n")
	buf.WriteString("\t\tcase *runtime.IntegerValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ast.Ty(string(v.TypeSuffix))\n")
	buf.WriteString("\t\tcase runtime.FloatValue:\n")
	buf.WriteString("\t\t\treturn ast.Ty(string(v.TypeSuffix))\n")
	buf.WriteString("\t\tcase *runtime.FloatValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ast.Ty(string(v.TypeSuffix))\n")
	buf.WriteString("\t\tcase *runtime.InterfaceValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn typeFromValue(v.Underlying)\n")
	buf.WriteString("\t\tcase runtime.InterfaceValue:\n")
	buf.WriteString("\t\t\treturn typeFromValue(v.Underlying)\n")
	buf.WriteString("\t\tcase *runtime.ArrayValue:\n")
	buf.WriteString("\t\t\tif v == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tvar elemType ast.TypeExpression\n")
	buf.WriteString("\t\t\tfor _, elem := range v.Elements {\n")
	buf.WriteString("\t\t\t\tinferred := typeFromValue(elem)\n")
	buf.WriteString("\t\t\t\tif inferred == nil {\n")
	buf.WriteString("\t\t\t\t\tcontinue\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tif elemType == nil {\n")
	buf.WriteString("\t\t\t\t\telemType = inferred\n")
	buf.WriteString("\t\t\t\t\tcontinue\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tif !typeExprEqual(elemType, inferred) {\n")
	buf.WriteString("\t\t\t\t\telemType = ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t\t\t\t\tbreak\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif elemType == nil {\n")
	buf.WriteString("\t\t\t\telemType = ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn ast.Gen(ast.Ty(\"Array\"), elemType)\n")
	buf.WriteString("\t\tcase *runtime.StructInstanceValue:\n")
	buf.WriteString("\t\t\tif v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {\n")
	buf.WriteString("\t\t\t\treturn nil\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tbase := ast.Ty(v.Definition.Node.ID.Name)\n")
	buf.WriteString("\t\t\tgenerics := v.Definition.Node.GenericParams\n")
	buf.WriteString("\t\t\tif len(generics) == 0 {\n")
	buf.WriteString("\t\t\t\treturn base\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\targs := v.TypeArguments\n")
	buf.WriteString("\t\t\tif len(args) != len(generics) {\n")
	buf.WriteString("\t\t\t\tfilled := make([]ast.TypeExpression, len(generics))\n")
	buf.WriteString("\t\t\t\tfor idx := range filled {\n")
	buf.WriteString("\t\t\t\t\tfilled[idx] = ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\targs = filled\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif len(args) > 0 {\n")
	buf.WriteString("\t\t\t\tcloned := make([]ast.TypeExpression, len(args))\n")
	buf.WriteString("\t\t\t\tfor idx, arg := range args {\n")
	buf.WriteString("\t\t\t\t\tif arg == nil {\n")
	buf.WriteString("\t\t\t\t\t\tcloned[idx] = ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t\t\t\t\t} else {\n")
	buf.WriteString("\t\t\t\t\t\tcloned[idx] = arg\n")
	buf.WriteString("\t\t\t\t\t}\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\treturn ast.Gen(base, cloned...)\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn base\n")
	buf.WriteString("\t\tdefault:\n")
	buf.WriteString("\t\t\treturn ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\thandleVal, err := __able_hash_map_new_impl(nil)\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tpanic(err)\n")
	buf.WriteString("\t}\n")
	for idx, element := range elements {
		switch element.kind {
		case "entry":
			keyTemp := fmt.Sprintf("__able_map_key_%d", idx)
			valueTemp := fmt.Sprintf("__able_map_value_%d", idx)
			buf.WriteString(fmt.Sprintf("\t%s := %s\n", keyTemp, element.key))
			buf.WriteString(fmt.Sprintf("\t%s := %s\n", valueTemp, element.value))
			buf.WriteString(fmt.Sprintf("\tkeyType = mergeType(keyType, typeFromValue(%s))\n", keyTemp))
			buf.WriteString(fmt.Sprintf("\tvalueType = mergeType(valueType, typeFromValue(%s))\n", valueTemp))
			buf.WriteString(fmt.Sprintf("\t_, err = __able_hash_map_set_impl([]runtime.Value{%s, %s, %s})\n", "handleVal", keyTemp, valueTemp))
			buf.WriteString("\tif err != nil {\n")
			buf.WriteString("\t\tpanic(err)\n")
			buf.WriteString("\t}\n")
		case "spread":
			spreadTemp := fmt.Sprintf("__able_map_spread_%d", idx)
			handleTemp := fmt.Sprintf("__able_map_spread_handle_%d", idx)
			callbackTemp := fmt.Sprintf("__able_map_spread_cb_%d", idx)
			buf.WriteString(fmt.Sprintf("\t%s := %s\n", spreadTemp, element.spread))
			buf.WriteString(fmt.Sprintf("\t%s := func(val runtime.Value) runtime.Value {\n", handleTemp))
			buf.WriteString("\t\tcurrent := val\n")
			buf.WriteString("\t\tswitch v := current.(type) {\n")
			buf.WriteString("\t\tcase *runtime.InterfaceValue:\n")
			buf.WriteString("\t\t\tif v != nil {\n")
			buf.WriteString("\t\t\t\tcurrent = v.Underlying\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\tcase runtime.InterfaceValue:\n")
			buf.WriteString("\t\t\tcurrent = v.Underlying\n")
			buf.WriteString("\t\t}\n")
			buf.WriteString("\t\tswitch inst := current.(type) {\n")
			buf.WriteString("\t\tcase *runtime.StructInstanceValue:\n")
			buf.WriteString("\t\t\tif inst == nil || inst.Fields == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != \"HashMap\" {\n")
			buf.WriteString("\t\t\t\tpanic(fmt.Errorf(\"map literal spread expects HashMap value\"))\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\t\tif len(inst.TypeArguments) >= 2 {\n")
			buf.WriteString("\t\t\t\tkeyType = mergeType(keyType, inst.TypeArguments[0])\n")
			buf.WriteString("\t\t\t\tvalueType = mergeType(valueType, inst.TypeArguments[1])\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\t\thandle, ok := inst.Fields[\"handle\"]\n")
			buf.WriteString("\t\t\tif !ok {\n")
			buf.WriteString("\t\t\t\tpanic(fmt.Errorf(\"map literal spread expects HashMap value\"))\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\t\treturn handle\n")
			buf.WriteString("\t\tdefault:\n")
			buf.WriteString("\t\t\tpanic(fmt.Errorf(\"map literal spread expects HashMap value\"))\n")
			buf.WriteString("\t\t}\n")
			buf.WriteString(fmt.Sprintf("\t}(%s)\n", spreadTemp))
			buf.WriteString(fmt.Sprintf("\t%s := runtime.NativeFunctionValue{\n", callbackTemp))
			buf.WriteString("\t\tName: \"__able_map_spread_cb\",\n")
			buf.WriteString("\t\tArity: 2,\n")
			buf.WriteString("\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
			buf.WriteString("\t\t\tif len(args) != 2 {\n")
			buf.WriteString("\t\t\t\treturn nil, fmt.Errorf(\"map literal spread callback expects key and value\")\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\t\tkeyType = mergeType(keyType, typeFromValue(args[0]))\n")
			buf.WriteString("\t\t\tvalueType = mergeType(valueType, typeFromValue(args[1]))\n")
			buf.WriteString(fmt.Sprintf("\t\t\t_, err := __able_hash_map_set_impl([]runtime.Value{%s, args[0], args[1]})\n", "handleVal"))
			buf.WriteString("\t\t\tif err != nil {\n")
			buf.WriteString("\t\t\t\treturn nil, err\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\t\treturn runtime.NilValue{}, nil\n")
			buf.WriteString("\t\t},\n")
			buf.WriteString("\t}\n")
			buf.WriteString(fmt.Sprintf("\t_, err = __able_hash_map_for_each_impl([]runtime.Value{%s, %s})\n", handleTemp, callbackTemp))
			buf.WriteString("\tif err != nil {\n")
			buf.WriteString("\t\tpanic(err)\n")
			buf.WriteString("\t}\n")
		}
	}
	buf.WriteString("\tif keyType == nil {\n")
	buf.WriteString("\t\tkeyType = ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif valueType == nil {\n")
	buf.WriteString("\t\tvalueType = ast.NewWildcardTypeExpression()\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\t_ = keyType\n")
	buf.WriteString("\t_ = valueType\n")
	buf.WriteString("\thandleRaw, err := __able_hash_map_handle_from_value(handleVal)\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tpanic(err)\n")
	buf.WriteString("\t}\n")
	buf.WriteString(fmt.Sprintf("\treturn &%s{Handle: handleRaw}\n", hashMapBase))
	buf.WriteString("}()")
	return buf.String(), returnType, true
}

func (g *generator) compileMemberAccess(ctx *compileContext, expr *ast.MemberAccessExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing member access")
		return nil, "", "", false
	}
	objLines, objectExpr, objectType, ok := g.compileExprLines(ctx, expr.Object, "")
	if !ok {
		return nil, "", "", false
	}
	if recoverLines, recoveredExpr, recoveredType, recovered := g.recoverDispatchExpr(ctx, expr.Object, objectExpr, objectType); recovered {
		objLines = append(objLines, recoverLines...)
		objectExpr = recoveredExpr
		objectType = recoveredType
	}
	objectCategory := g.typeCategory(objectType)
	if objectCategory == "runtime" || objectCategory == "any" {
		// When object is runtime.Value with known origin struct type, extract
		// and access the field directly instead of going through dynamic dispatch.
		if !expr.Safe {
			if fieldLines, fieldExpr, fieldType, ok := g.compileOriginStructFieldAccess(ctx, expr, objectExpr, expected); ok {
				lines := append([]string{}, objLines...)
				lines = append(lines, fieldLines...)
				return lines, fieldExpr, fieldType, true
			}
		}
		memberValue, ok := g.memberAssignmentRuntimeValue(ctx, expr.Member)
		if !ok {
			ctx.setReason("unsupported member access")
			return nil, "", "", false
		}
		objConvLines, objValue, ok := g.lowerRuntimeValue(ctx, objectExpr, objectType)
		if !ok {
			ctx.setReason("unsupported member access")
			return nil, "", "", false
		}
		lines := append([]string{}, objLines...)
		lines = append(lines, objConvLines...)
		var baseExpr string
		if expr.Safe {
			objTemp := ctx.newTemp()
			memberTemp := ctx.newTemp()
			resultTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			lines = append(lines,
				fmt.Sprintf("%s := %s", objTemp, objValue),
				fmt.Sprintf("%s := %s", memberTemp, memberValue),
				fmt.Sprintf("var %s runtime.Value", resultTemp),
				fmt.Sprintf("var %s *__ableControl", controlTemp),
				fmt.Sprintf("if __able_is_nil(%s) { %s = runtime.NilValue{} } else { %s, %s = __able_member_get(%s, %s) }", objTemp, resultTemp, resultTemp, controlTemp, objTemp, memberTemp),
			)
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, controlLines...)
			baseExpr = resultTemp
		} else {
			var ok bool
			lines, baseExpr, ok = g.appendRuntimeMemberGetControlLines(ctx, lines, objValue, memberValue)
			if !ok {
				return nil, "", "", false
			}
		}
		if expected == "" || expected == "runtime.Value" || expected == "any" {
			return lines, baseExpr, "runtime.Value", true
		}
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, baseExpr, expected)
		if !ok {
			ctx.setReason(memberAccessMismatchReason(objectExpr, objectType, g.memberName(expr.Member), "runtime.Value", expected))
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	// Handle member access on Go builtin string (Able String struct fields).
	if objectType == "string" {
		if memberName := g.memberName(expr.Member); memberName != "" {
			if fieldExpr, fieldType, ok := g.stringBuiltinFieldAccess(objectExpr, memberName); ok {
				if g.typeMatches(expected, fieldType) {
					return objLines, fieldExpr, fieldType, true
				}
				// Allow integer cast for len_bytes (uint) to the expected integer type.
				if expected != "" && g.isIntegerType(expected) && g.isIntegerType(fieldType) {
					castExpr := fmt.Sprintf("%s(%s)", expected, fieldExpr)
					return objLines, castExpr, expected, true
				}
				ctx.setReason(memberAccessMismatchReason(objectExpr, objectType, memberName, fieldType, expected))
				return nil, "", "", false
			}
		}
	}
	memberName := g.memberName(expr.Member)
	if objectType == "runtime.ErrorValue" && memberName != "" {
		if lines, fieldExpr, fieldType, ok := g.compileNativeErrorMemberAccess(ctx, objectExpr, memberName, expected); ok {
			allLines := append([]string{}, objLines...)
			allLines = append(allLines, lines...)
			return allLines, fieldExpr, fieldType, true
		}
	}
	if !expr.Safe && memberName != "" {
		if ifaceMethod, ok := g.nativeInterfaceMethodForGoType(objectType, memberName); ok {
			lines, callableExpr, callableType, ok := g.compileNativeInterfaceBoundMethodValue(ctx, objectExpr, objectType, ifaceMethod)
			if ok {
				allLines := append([]string{}, objLines...)
				allLines = append(allLines, lines...)
				return allLines, callableExpr, callableType, true
			}
		}
		if lines, callableExpr, callableType, ok := g.compileStaticIteratorControllerBoundMethodValue(ctx, objectExpr, objectType, memberName, expected); ok {
			allLines := append([]string{}, objLines...)
			allLines = append(allLines, lines...)
			return allLines, callableExpr, callableType, true
		}
	}
	info := g.structInfoByGoName(objectType)
	if info == nil {
		ctx.setReason("unsupported member access")
		return nil, "", "", false
	}
	field, ok := g.structFieldForMember(info, expr.Member)
	if !ok {
		ctx.setReason("unsupported member access")
		return nil, "", "", false
	}
	if field != nil {
		fieldExpr := fmt.Sprintf("%s.%s", objectExpr, field.GoName)
		fieldType := field.GoType
		if fieldType == "runtime.Value" && expected != "" && expected != "runtime.Value" && expected != "any" {
			lines := append([]string{}, objLines...)
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, fieldExpr, expected)
			if ok {
				lines = append(lines, convLines...)
				return lines, converted, expected, true
			}
		}
		if !g.typeMatches(expected, field.GoType) {
			coerceLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, append([]string{}, objLines...), fieldExpr, fieldType, expected)
			if ok && (expected == "" || g.typeMatches(expected, coercedType)) {
				return coerceLines, coercedExpr, coercedType, true
			}
			ctx.setReason(memberAccessMismatchReason(objectExpr, objectType, memberName, fieldType, expected))
			return nil, "", "", false
		}
		if expr.Safe && strings.HasPrefix(objectType, "*") {
			objTemp := ctx.newTemp()
			resultType := expected
			if resultType == "" {
				if inferredType, ok := g.safeNavigationCarrierType(fieldType); ok {
					resultType = inferredType
				} else {
					resultType = "any"
				}
			}
			resultTemp := ctx.newTemp()
			nilExpr := safeNilReturnExpr(resultType)
			if wrapped, ok := g.nativeUnionNilExpr(resultType); ok {
				nilExpr = wrapped
			}
			lines := append([]string{}, objLines...)
			lines = append(lines,
				fmt.Sprintf("%s := %s", objTemp, objectExpr),
				fmt.Sprintf("var %s %s", resultTemp, resultType),
			)
			lines = append(lines, fmt.Sprintf("if %s == nil {", objTemp))
			lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, nilExpr))
			lines = append(lines, "} else {")
			coerceLines, coercedExpr, ok := g.safeNavigationCoerceSuccessExpr(ctx, fmt.Sprintf("%s.%s", objTemp, field.GoName), fieldType, resultType)
			if !ok {
				ctx.setReason(memberAccessMismatchReason(objectExpr, objectType, memberName, fieldType, resultType))
				return nil, "", "", false
			}
			lines = append(lines, indentLines(coerceLines, 1)...)
			lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, coercedExpr))
			lines = append(lines, "}")
			return lines, resultTemp, resultType, true
		}
		return objLines, fieldExpr, fieldType, true
	}
	if !expr.Safe && memberName != "" {
		if method := g.methodForReceiver(objectType, memberName); method != nil {
			lines, callableExpr, callableType, ok := g.compileNativeBoundMethodValue(ctx, objectExpr, objectType, method)
			if ok {
				allLines := append([]string{}, objLines...)
				allLines = append(allLines, lines...)
				return allLines, callableExpr, callableType, true
			}
		}
		if method := g.compileableInterfaceMethodForConcreteReceiver(objectType, memberName); method != nil {
			lines, callableExpr, callableType, ok := g.compileNativeBoundMethodValue(ctx, objectExpr, objectType, method)
			if ok {
				allLines := append([]string{}, objLines...)
				allLines = append(allLines, lines...)
				return allLines, callableExpr, callableType, true
			}
		}
	}
	memberValue, ok := g.memberAssignmentRuntimeValue(ctx, expr.Member)
	if !ok {
		ctx.setReason("unknown struct field")
		return nil, "", "", false
	}
	objConvLines, objValue, ok := g.lowerRuntimeValue(ctx, objectExpr, objectType)
	if !ok {
		ctx.setReason("unknown struct field")
		return nil, "", "", false
	}
	lines := append([]string{}, objLines...)
	lines = append(lines, objConvLines...)
	baseExpr := ""
	lines, baseExpr, ok = g.appendRuntimeMemberGetControlLines(ctx, lines, objValue, memberValue)
	if !ok {
		return nil, "", "", false
	}
	if expected == "" || expected == "runtime.Value" {
		return lines, baseExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, baseExpr, expected)
	if !ok {
		ctx.setReason(memberAccessMismatchReason(objectExpr, objectType, memberName, "runtime.Value", expected))
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

// compileOriginStructFieldAccess resolves a field access on a runtime.Value variable
// whose underlying struct type is known. Extracts the struct and accesses the Go field directly.
func (g *generator) compileOriginStructFieldAccess(ctx *compileContext, expr *ast.MemberAccessExpression, objExpr string, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Object == nil {
		return nil, "", "", false
	}
	objIdent, ok := expr.Object.(*ast.Identifier)
	if !ok || objIdent == nil || objIdent.Name == "" {
		return nil, "", "", false
	}
	info, ok := ctx.lookup(objIdent.Name)
	if !ok || info.OriginGoType == "" {
		return nil, "", "", false
	}
	structInfo := g.structInfoByGoName(info.OriginGoType)
	if structInfo == nil {
		return nil, "", "", false
	}
	field, ok := g.structFieldForMember(structInfo, expr.Member)
	if !ok || field == nil {
		return nil, "", "", false
	}
	if field.GoType != "runtime.Value" && !g.typeMatches(expected, field.GoType) {
		return nil, "", "", false
	}
	baseName, ok := g.structBaseName(info.OriginGoType)
	if !ok {
		return nil, "", "", false
	}
	// CSE: reuse existing extraction temp if available.
	if cached, ok := ctx.originExtractions[objIdent.Name]; ok {
		if field.GoType == "runtime.Value" && expected != "" && expected != "runtime.Value" && expected != "any" {
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, fmt.Sprintf("%s.%s", cached, field.GoName), expected)
			if !ok {
				return nil, "", "", false
			}
			return convLines, converted, expected, true
		}
		return nil, fmt.Sprintf("%s.%s", cached, field.GoName), field.GoType, true
	}
	extractTemp := ctx.newTemp()
	extractErrTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", extractTemp, extractErrTemp, baseName, objExpr),
		fmt.Sprintf("if %s != nil { panic(%s) }", extractErrTemp, extractErrTemp),
	}
	// Cache extraction for subsequent accesses in this scope.
	if ctx.originExtractions == nil {
		ctx.originExtractions = make(map[string]string)
	}
	ctx.originExtractions[objIdent.Name] = extractTemp
	if field.GoType == "runtime.Value" && expected != "" && expected != "runtime.Value" && expected != "any" {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, fmt.Sprintf("%s.%s", extractTemp, field.GoName), expected)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, fmt.Sprintf("%s.%s", extractTemp, field.GoName), field.GoType, true
}
