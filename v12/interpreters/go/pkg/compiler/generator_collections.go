package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileArrayLiteral(ctx *compileContext, lit *ast.ArrayLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing array literal")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("array literal type mismatch")
		return "", "", false
	}
	elements := make([]string, 0, len(lit.Elements))
	for _, element := range lit.Elements {
		expr, goType, ok := g.compileExpr(ctx, element, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("array literal element unsupported")
			return "", "", false
		}
		elements = append(elements, valueExpr)
	}
	expr := fmt.Sprintf("&runtime.ArrayValue{Elements: []runtime.Value{%s}}", strings.Join(elements, ", "))
	return expr, "runtime.Value", true
}

func (g *generator) compileMapLiteral(ctx *compileContext, lit *ast.MapLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing map literal")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("map literal type mismatch")
		return "", "", false
	}
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
	buf.WriteString("func() runtime.Value {\n")
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
	buf.WriteString("\tdef, err := __able_runtime.StructDefinition(\"HashMap\")\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tpanic(err)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\thandleVal, err := __able_runtime.Call(\"__able_hash_map_new\", nil)\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tpanic(err)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tinst := &runtime.StructInstanceValue{Definition: def, Fields: map[string]runtime.Value{\"handle\": handleVal}}\n")
	for idx, element := range elements {
		switch element.kind {
		case "entry":
			keyTemp := fmt.Sprintf("__able_map_key_%d", idx)
			valueTemp := fmt.Sprintf("__able_map_value_%d", idx)
			buf.WriteString(fmt.Sprintf("\t%s := %s\n", keyTemp, element.key))
			buf.WriteString(fmt.Sprintf("\t%s := %s\n", valueTemp, element.value))
			buf.WriteString(fmt.Sprintf("\tkeyType = mergeType(keyType, typeFromValue(%s))\n", keyTemp))
			buf.WriteString(fmt.Sprintf("\tvalueType = mergeType(valueType, typeFromValue(%s))\n", valueTemp))
			buf.WriteString(fmt.Sprintf("\t_, err = __able_runtime.Call(\"__able_hash_map_set\", []runtime.Value{%s, %s, %s})\n", "handleVal", keyTemp, valueTemp))
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
			buf.WriteString(fmt.Sprintf("\t\t\t_, err := __able_runtime.Call(\"__able_hash_map_set\", []runtime.Value{%s, args[0], args[1]})\n", "handleVal"))
			buf.WriteString("\t\t\tif err != nil {\n")
			buf.WriteString("\t\t\t\treturn nil, err\n")
			buf.WriteString("\t\t\t}\n")
			buf.WriteString("\t\t\treturn runtime.NilValue{}, nil\n")
			buf.WriteString("\t\t},\n")
			buf.WriteString("\t}\n")
			buf.WriteString(fmt.Sprintf("\t_, err = __able_runtime.Call(\"__able_hash_map_for_each\", []runtime.Value{%s, %s})\n", handleTemp, callbackTemp))
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
	buf.WriteString("\tinst.TypeArguments = []ast.TypeExpression{keyType, valueType}\n")
	buf.WriteString("\treturn inst\n")
	buf.WriteString("}()")
	return buf.String(), "runtime.Value", true
}

func (g *generator) compileMemberAccess(ctx *compileContext, expr *ast.MemberAccessExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing member access")
		return "", "", false
	}
	if expr.Safe {
		ctx.setReason("safe member access unsupported")
		return "", "", false
	}
	member, ok := expr.Member.(*ast.Identifier)
	if !ok || member == nil || member.Name == "" {
		ctx.setReason("unsupported member access")
		return "", "", false
	}
	objectExpr, objectType, ok := g.compileExpr(ctx, expr.Object, "")
	if !ok {
		return "", "", false
	}
	info := g.structInfoByGoName(objectType)
	if info == nil {
		ctx.setReason("unsupported member access")
		return "", "", false
	}
	field := g.fieldInfo(info, member.Name)
	if field == nil {
		ctx.setReason("unknown struct field")
		return "", "", false
	}
	if !g.typeMatches(expected, field.GoType) {
		ctx.setReason("member access type mismatch")
		return "", "", false
	}
	return fmt.Sprintf("%s.%s", objectExpr, field.GoName), field.GoType, true
}

func (g *generator) compileIndexExpression(ctx *compileContext, expr *ast.IndexExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing index expression")
		return "", "", false
	}
	objExpr, objType, ok := g.compileExpr(ctx, expr.Object, "")
	if !ok {
		return "", "", false
	}
	objValue, ok := g.runtimeValueExpr(objExpr, objType)
	if !ok {
		ctx.setReason("index object unsupported")
		return "", "", false
	}
	idxExpr, idxType, ok := g.compileExpr(ctx, expr.Index, "")
	if !ok {
		return "", "", false
	}
	idxValue, ok := g.runtimeValueExpr(idxExpr, idxType)
	if !ok {
		ctx.setReason("index expression unsupported")
		return "", "", false
	}
	baseExpr := fmt.Sprintf("__able_index(%s, %s)", objValue, idxValue)
	if expected == "" || expected == "runtime.Value" {
		return baseExpr, "runtime.Value", true
	}
	converted, ok := g.expectRuntimeValueExpr(baseExpr, expected)
	if !ok {
		ctx.setReason("index expression type mismatch")
		return "", "", false
	}
	return converted, expected, true
}

func (g *generator) expectRuntimeValueExpr(valueExpr string, expected string) (string, bool) {
	switch g.typeCategory(expected) {
	case "bool":
		return fmt.Sprintf("func() bool { val := %s; v, err := bridge.AsBool(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "string":
		return fmt.Sprintf("func() string { val := %s; v, err := bridge.AsString(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "rune":
		return fmt.Sprintf("func() rune { val := %s; v, err := bridge.AsRune(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "float32":
		return fmt.Sprintf("func() float32 { val := %s; v, err := bridge.AsFloat(val); if err != nil { panic(err) }; return float32(v) }()", valueExpr), true
	case "float64":
		return fmt.Sprintf("func() float64 { val := %s; v, err := bridge.AsFloat(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "int":
		return fmt.Sprintf("func() int { val := %s; v, err := bridge.AsInt(val, bridge.NativeIntBits); if err != nil { panic(err) }; return int(v) }()", valueExpr), true
	case "uint":
		return fmt.Sprintf("func() uint { val := %s; v, err := bridge.AsUint(val, bridge.NativeIntBits); if err != nil { panic(err) }; return uint(v) }()", valueExpr), true
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(expected)
		return fmt.Sprintf("func() %s { val := %s; v, err := bridge.AsInt(val, %d); if err != nil { panic(err) }; return %s(v) }()", expected, valueExpr, bits, expected), true
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(expected)
		return fmt.Sprintf("func() %s { val := %s; v, err := bridge.AsUint(val, %d); if err != nil { panic(err) }; return %s(v) }()", expected, valueExpr, bits, expected), true
	case "struct":
		return fmt.Sprintf("func() %s { val := %s; v, err := __able_struct_%s_from(val); if err != nil { panic(err) }; return v }()", expected, valueExpr, expected), true
	}
	return "", false
}

func (g *generator) structInfoByGoName(goName string) *structInfo {
	if goName == "" {
		return nil
	}
	for _, info := range g.structs {
		if info != nil && info.GoName == goName {
			return info
		}
	}
	return nil
}

func (g *generator) runtimeValueExpr(expr string, goType string) (string, bool) {
	switch g.typeCategory(goType) {
	case "runtime":
		return expr, true
	case "bool":
		return fmt.Sprintf("bridge.ToBool(%s)", expr), true
	case "string":
		return fmt.Sprintf("bridge.ToString(%s)", expr), true
	case "rune":
		return fmt.Sprintf("bridge.ToRune(%s)", expr), true
	case "float32":
		return fmt.Sprintf("bridge.ToFloat32(%s)", expr), true
	case "float64":
		return fmt.Sprintf("bridge.ToFloat64(%s)", expr), true
	case "int":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))", expr), true
	case "uint":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))", expr), true
	case "int8":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\"))", expr), true
	case "int16":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\"))", expr), true
	case "int32":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\"))", expr), true
	case "int64":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\"))", expr), true
	case "uint8":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\"))", expr), true
	case "uint16":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\"))", expr), true
	case "uint32":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\"))", expr), true
	case "uint64":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\"))", expr), true
	case "struct":
		return fmt.Sprintf("func() runtime.Value { if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }; v, err := __able_struct_%s_to(__able_runtime, %s); if err != nil { panic(err) }; return v }()", goType, expr), true
	default:
		return "", false
	}
}
