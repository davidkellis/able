package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileArrayLiteral(ctx *compileContext, lit *ast.ArrayLiteral, expected string) ([]string, string, string, bool) {
	if lit == nil {
		ctx.setReason("missing array literal")
		return nil, "", "", false
	}
	var returnType string
	switch {
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
	elementExpectedType := ""
	if kind, ok := g.monoArrayKindForLiteral(ctx, nil); ok {
		elementExpectedType = g.monoArrayElemGoType(kind)
	}
	var elementLines []string
	for _, element := range lit.Elements {
		elLines, expr, goType, ok := g.compileExprLines(ctx, element, elementExpectedType)
		if !ok {
			return nil, "", "", false
		}
		elementLines = append(elementLines, elLines...)
		elementExprs = append(elementExprs, expr)
		elementTypes = append(elementTypes, goType)
	}
	if returnType != "runtime.Value" {
		if monoKind, ok := g.monoArrayKindForLiteral(ctx, elementTypes); ok {
			monoGoType := g.monoArrayElemGoType(monoKind)
			capacityExpr := fmt.Sprintf("%d", len(lit.Elements))
			newHandleExpr, ok := g.monoArrayNewWithCapacityExpr(monoKind, capacityExpr)
			if !ok {
				ctx.setReason("array literal unsupported mono kind")
				return nil, "", "", false
			}
			lines := append([]string{}, elementLines...)
			handleTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", handleTemp, newHandleExpr))
			for idx, elementExpr := range elementExprs {
				coercedExpr, ok := g.coerceExprToGoType(elementExpr, elementTypes[idx], monoGoType)
				if !ok {
					ctx.setReason("array literal element unsupported")
					return nil, "", "", false
				}
				valueTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, coercedExpr))
				writeExpr, ok := g.monoArrayWriteExpr(monoKind, handleTemp, fmt.Sprintf("%d", idx), valueTemp)
				if !ok {
					ctx.setReason("array literal mono write unsupported")
					return nil, "", "", false
				}
				lines = append(lines, fmt.Sprintf("__able_panic_on_error(%s)", writeExpr))
			}
			arrTemp := ctx.newTemp()
			arrExpr := fmt.Sprintf("&Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: bridge.ToInt(%s, runtime.IntegerType(\"i64\"))}", len(lit.Elements), len(lit.Elements), handleTemp)
			if !strings.HasPrefix(returnType, "*") {
				arrExpr = fmt.Sprintf("Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: bridge.ToInt(%s, runtime.IntegerType(\"i64\"))}", len(lit.Elements), len(lit.Elements), handleTemp)
			}
			lines = append(lines, fmt.Sprintf("%s := %s", arrTemp, arrExpr))
			return lines, arrTemp, returnType, true
		}
	}
	callNode := g.diagNodeName(lit, "*ast.ArrayLiteral", "array")
	lines := append([]string{}, elementLines...)
	lines = append(lines, "if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }")
	handleTemp := ctx.newTemp()
	capacityExpr := fmt.Sprintf("bridge.ToInt(int64(%d), runtime.IntegerType(\"i32\"))", len(lit.Elements))
	lines = append(lines, fmt.Sprintf("%s := __able_extern_array_with_capacity([]runtime.Value{%s}, %s)", handleTemp, capacityExpr, callNode))
	for idx, expr := range elementExprs {
		goType := elementTypes[idx]
		elemConvLines, valueExpr, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("array literal element unsupported")
			return nil, "", "", false
		}
		lines = append(lines, elemConvLines...)
		valueTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
		indexExpr := fmt.Sprintf("bridge.ToInt(int64(%d), runtime.IntegerType(\"i32\"))", idx)
		lines = append(lines, fmt.Sprintf("_ = __able_extern_array_write([]runtime.Value{%s, %s, %s}, %s)", handleTemp, indexExpr, valueTemp, callNode))
	}
	if returnType == "runtime.Value" {
		arrTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, err := __able_struct_Array_to(__able_runtime, &Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: %s})", arrTemp, len(lit.Elements), len(lit.Elements), handleTemp))
		lines = append(lines, "if err != nil { panic(err) }")
		return lines, arrTemp, "runtime.Value", true
	}
	arrTemp := ctx.newTemp()
	arrExpr := fmt.Sprintf("&Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: %s}", len(lit.Elements), len(lit.Elements), handleTemp)
	if !strings.HasPrefix(returnType, "*") {
		arrExpr = fmt.Sprintf("Array{Length: int32(%d), Capacity: int32(%d), Storage_handle: %s}", len(lit.Elements), len(lit.Elements), handleTemp)
	}
	lines = append(lines, fmt.Sprintf("%s := %s", arrTemp, arrExpr))
	return lines, arrTemp, returnType, true
}

func (g *generator) compileMapLiteral(ctx *compileContext, lit *ast.MapLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing map literal")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" {
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
	buf.WriteString("\thandleVal, err := __able_hash_map_new_impl(nil)\n")
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
	buf.WriteString("\tinst.TypeArguments = []ast.TypeExpression{keyType, valueType}\n")
	buf.WriteString("\treturn inst\n")
	buf.WriteString("}()")
	return buf.String(), "runtime.Value", true
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
		objConvLines, objValue, ok := g.runtimeValueLines(ctx, objectExpr, objectType)
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
			lines = append(lines,
				fmt.Sprintf("%s := %s", objTemp, objValue),
				fmt.Sprintf("%s := %s", memberTemp, memberValue),
				fmt.Sprintf("var %s runtime.Value", resultTemp),
				fmt.Sprintf("if __able_is_nil(%s) { %s = runtime.NilValue{} } else { %s = __able_member_get(%s, %s) }", objTemp, resultTemp, resultTemp, objTemp, memberTemp),
			)
			baseExpr = resultTemp
		} else {
			baseExpr = fmt.Sprintf("__able_member_get(%s, %s)", objValue, memberValue)
		}
		if expected == "" || expected == "runtime.Value" || expected == "any" {
			return lines, baseExpr, "runtime.Value", true
		}
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, baseExpr, expected)
		if !ok {
			ctx.setReason("member access type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	// Handle member access on Go builtin string (Able String struct fields).
	if objectType == "string" {
		if memberName := g.memberName(expr.Member); memberName != "" {
			if fieldExpr, fieldType, ok := g.stringBuiltinFieldAccess(objectExpr, memberName); ok {
				if !g.typeMatches(expected, fieldType) {
					ctx.setReason("member access type mismatch")
					return nil, "", "", false
				}
				return objLines, fieldExpr, fieldType, true
			}
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
		if !g.typeMatches(expected, field.GoType) {
			ctx.setReason("member access type mismatch")
			return nil, "", "", false
		}
		return objLines, fmt.Sprintf("%s.%s", objectExpr, field.GoName), field.GoType, true
	}
	memberValue, ok := g.memberAssignmentRuntimeValue(ctx, expr.Member)
	if !ok {
		ctx.setReason("unknown struct field")
		return nil, "", "", false
	}
	objConvLines, objValue, ok := g.runtimeValueLines(ctx, objectExpr, objectType)
	if !ok {
		ctx.setReason("unknown struct field")
		return nil, "", "", false
	}
	baseExpr := fmt.Sprintf("__able_member_get(%s, %s)", objValue, memberValue)
	if expected == "" || expected == "runtime.Value" {
		lines := append([]string{}, objLines...)
		lines = append(lines, objConvLines...)
		return lines, baseExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, baseExpr, expected)
	if !ok {
		ctx.setReason("member access type mismatch")
		return nil, "", "", false
	}
	lines := append([]string{}, objLines...)
	lines = append(lines, objConvLines...)
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
	if !g.typeMatches(expected, field.GoType) {
		return nil, "", "", false
	}
	baseName, ok := g.structBaseName(info.OriginGoType)
	if !ok {
		return nil, "", "", false
	}
	// CSE: reuse existing extraction temp if available.
	if cached, ok := ctx.originExtractions[objIdent.Name]; ok {
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
	return lines, fmt.Sprintf("%s.%s", extractTemp, field.GoName), field.GoType, true
}

func (g *generator) compileIndexExpression(ctx *compileContext, expr *ast.IndexExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing index expression")
		return nil, "", "", false
	}
	objLines, objExpr, objType, ok := g.compileExprLines(ctx, expr.Object, "")
	if !ok {
		return nil, "", "", false
	}
	if g.isArrayStructType(objType) {
		if monoKind, monoEnabled := g.monoArrayKindForObject(ctx, expr.Object, objType); monoEnabled {
			idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, expr.Index, "")
			if !ok {
				return nil, "", "", false
			}
			objTemp := ctx.newTemp()
			idxTemp := ctx.newTemp()
			indexTemp := ctx.newTemp()
			handleRawTemp := ctx.newTemp()
			handleTemp := ctx.newTemp()
			lengthTemp := ctx.newTemp()
			nativeTemp := ctx.newTemp()
			resultTemp := ctx.newTemp()
			readExpr, readType, ok := g.monoArrayReadExpr(monoKind, handleTemp, indexTemp)
			if !ok {
				ctx.setReason("index expression unsupported mono read")
				return nil, "", "", false
			}
			monoConvLines, runtimeExpr, ok := g.runtimeValueLines(ctx, nativeTemp, readType)
			if !ok {
				ctx.setReason("index expression unsupported mono conversion")
				return nil, "", "", false
			}
			lines := append([]string{}, objLines...)
			lines = append(lines, idxLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
			lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
			if !ok {
				ctx.setReason("index expression unsupported")
				return nil, "", "", false
			}
			lines = append(lines, monoConvLines...)
			lines = append(lines,
				fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
				fmt.Sprintf("%s, %s_err := bridge.AsInt(%s, 64)", handleTemp, handleTemp, handleRawTemp),
			fmt.Sprintf("if %s_err != nil { panic(%s_err) }", handleTemp, handleTemp),
				fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
				"if err != nil { panic(err) }",
				fmt.Sprintf("var %s runtime.Value", resultTemp),
				fmt.Sprintf("if %s < 0 || %s >= %s { %s = __able_index_error(%s, %s) } else {", indexTemp, indexTemp, lengthTemp, resultTemp, indexTemp, lengthTemp),
				fmt.Sprintf("\t%s, err := %s", nativeTemp, readExpr),
				"\tif err != nil { panic(err) }",
				fmt.Sprintf("\t%s = %s", resultTemp, runtimeExpr),
				"}",
			)
			if expected == "" || expected == "runtime.Value" {
				return lines, resultTemp, "runtime.Value", true
			}
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
			if !ok {
				ctx.setReason("index expression type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, expr.Index, "")
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		handleRawTemp := ctx.newTemp()
		handleTemp := ctx.newTemp()
		lengthTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		readTemp := ctx.newTemp()
		lines := append([]string{}, objLines...)
		lines = append(lines, idxLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			ctx.setReason("index expression unsupported")
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
			fmt.Sprintf("%s, %s_err := bridge.AsInt(%s, 64)", handleTemp, handleTemp, handleRawTemp),
			fmt.Sprintf("if %s_err != nil { panic(%s_err) }", handleTemp, handleTemp),
			fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
			"if err != nil { panic(err) }",
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s < 0 || %s >= %s { %s = __able_index_error(%s, %s) } else {", indexTemp, indexTemp, lengthTemp, resultTemp, indexTemp, lengthTemp),
			fmt.Sprintf("\t%s, err := runtime.ArrayStoreRead(%s, %s)", readTemp, handleTemp, indexTemp),
			"\tif err != nil { panic(err) }",
			fmt.Sprintf("\tif %s == nil { %s = runtime.NilValue{} } else { %s = %s }", readTemp, resultTemp, resultTemp, readTemp),
			"}",
		)
		if expected == "" || expected == "runtime.Value" {
			return lines, resultTemp, "runtime.Value", true
		}
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("index expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	objConvLines, objValue, ok := g.runtimeValueLines(ctx, objExpr, objType)
	if !ok {
		ctx.setReason("index object unsupported")
		return nil, "", "", false
	}
	idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, expr.Index, "")
	if !ok {
		return nil, "", "", false
	}
	idxConvLines, idxValue, ok := g.runtimeValueLines(ctx, idxExpr, idxType)
	if !ok {
		ctx.setReason("index expression unsupported")
		return nil, "", "", false
	}
	lines := append([]string{}, objLines...)
	lines = append(lines, objConvLines...)
	lines = append(lines, idxLines...)
	lines = append(lines, idxConvLines...)
	baseExpr := fmt.Sprintf("__able_index(%s, %s)", objValue, idxValue)
	if expected == "" || expected == "runtime.Value" {
		return lines, baseExpr, "runtime.Value", true
	}
	convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, baseExpr, expected)
	if !ok {
		ctx.setReason("index expression type mismatch")
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}

func (g *generator) compileArrayMethodIntrinsicCall(
	ctx *compileContext,
	objNode ast.Expression,
	objExpr string,
	objType string,
	methodName string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if g == nil || ctx == nil || !g.isArrayStructType(objType) {
		return nil, "", "", false
	}
	monoKind, monoEnabled := g.monoArrayKindForObject(ctx, objNode, objType)
	switch methodName {
	case "len":
		return g.compileArrayMethodLenIntrinsic(ctx, objExpr, args, expected, callNode)
	case "push":
		return g.compileArrayMethodPushIntrinsic(ctx, objExpr, monoKind, monoEnabled, args, expected, callNode)
	case "pop":
		return g.compileArrayMethodPopIntrinsic(ctx, objExpr, args, expected, callNode)
	case "first":
		return g.compileArrayMethodFirstLastIntrinsic(ctx, objExpr, args, expected, callNode, true)
	case "last":
		return g.compileArrayMethodFirstLastIntrinsic(ctx, objExpr, args, expected, callNode, false)
	case "is_empty":
		return g.compileArrayMethodIsEmptyIntrinsic(ctx, objExpr, args, expected, callNode)
	case "clear":
		return g.compileArrayMethodClearIntrinsic(ctx, objExpr, args, expected, callNode)
	case "capacity":
		return g.compileArrayMethodCapacityIntrinsic(ctx, objExpr, args, expected, callNode)
	case "get":
		if len(args) != 1 {
			return nil, "", "", false
		}
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, args[0], "")
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		handleRawTemp := ctx.newTemp()
		handleTemp := ctx.newTemp()
		lengthTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := append(idxLines, []string{
			fmt.Sprintf("__able_push_call_frame(%s)", callNode),
			fmt.Sprintf("%s := %s", objTemp, objExpr),
		}...)
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
			fmt.Sprintf("%s, %s_err := bridge.AsInt(%s, 64)", handleTemp, handleTemp, handleRawTemp),
			fmt.Sprintf("if %s_err != nil { panic(%s_err) }", handleTemp, handleTemp),
			fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
			"if err != nil { panic(err) }",
			fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
			fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
		)
		if monoEnabled {
			readExpr, readType, ok := g.monoArrayReadExpr(monoKind, handleTemp, indexTemp)
			if !ok {
				return nil, "", "", false
			}
			nativeTemp := ctx.newTemp()
			runtimeExpr, ok := g.runtimeValueExpr(nativeTemp, readType)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines,
				fmt.Sprintf("if %s >= 0 && %s < %s {", indexTemp, indexTemp, lengthTemp),
				fmt.Sprintf("\t%s, err := %s", nativeTemp, readExpr),
				"\tif err != nil { panic(err) }",
				fmt.Sprintf("\t%s = %s", resultTemp, runtimeExpr),
				"}",
			)
		} else {
			readTemp := ctx.newTemp()
			lines = append(lines,
				fmt.Sprintf("if %s >= 0 && %s < %s {", indexTemp, indexTemp, lengthTemp),
				fmt.Sprintf("\t%s, err := runtime.ArrayStoreRead(%s, %s)", readTemp, handleTemp, indexTemp),
				"\tif err != nil { panic(err) }",
				fmt.Sprintf("\tif %s != nil { %s = %s }", readTemp, resultTemp, readTemp),
				"}",
			)
		}
		lines = append(lines, "__able_pop_call_frame()")
		if expected == "" || expected == "runtime.Value" {
			return lines, resultTemp, "runtime.Value", true
		}
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	case "set":
		if len(args) != 2 {
			return nil, "", "", false
		}
		setIdxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, args[0], "")
		if !ok {
			return nil, "", "", false
		}
		setValLines, valueExpr, valueType, ok := g.compileExprLines(ctx, args[1], "")
		if !ok {
			return nil, "", "", false
		}
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		indexTemp := ctx.newTemp()
		valueTemp := ctx.newTemp()
		handleRawTemp := ctx.newTemp()
		handleTemp := ctx.newTemp()
		lengthTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := append(setIdxLines, setValLines...)
		lines = append(lines, []string{
			fmt.Sprintf("__able_push_call_frame(%s)", callNode),
			fmt.Sprintf("%s := %s", objTemp, objExpr),
		}...)
		lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
			fmt.Sprintf("%s, %s_err := bridge.AsInt(%s, 64)", handleTemp, handleTemp, handleRawTemp),
			fmt.Sprintf("if %s_err != nil { panic(%s_err) }", handleTemp, handleTemp),
			fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
			"if err != nil { panic(err) }",
			fmt.Sprintf("%s.Length = int32(%s)", objTemp, lengthTemp),
			fmt.Sprintf("var %s runtime.Value = runtime.NilValue{}", resultTemp),
		)
		if monoEnabled {
			monoGoType := g.monoArrayElemGoType(monoKind)
			if monoGoType == "" {
				return nil, "", "", false
			}
			coercedValueExpr, ok := g.coerceExprToGoType(valueExpr, valueType, monoGoType)
			if !ok {
				return nil, "", "", false
			}
			runtimeAssignedExpr, ok := g.runtimeValueExpr(valueTemp, monoGoType)
			if !ok {
				return nil, "", "", false
			}
			writeExpr, ok := g.monoArrayWriteExpr(monoKind, handleTemp, indexTemp, valueTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, coercedValueExpr))
			lines = append(lines, fmt.Sprintf("if %s < 0 || %s >= %s { %s = __able_index_error(%s, %s) } else { __able_panic_on_error(%s); %s = %s }", indexTemp, indexTemp, lengthTemp, resultTemp, indexTemp, lengthTemp, writeExpr, resultTemp, runtimeAssignedExpr))
		} else {
			valConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, valConvLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
			lines = append(lines, fmt.Sprintf("if %s < 0 || %s >= %s { %s = __able_index_error(%s, %s) } else { __able_panic_on_error(runtime.ArrayStoreWrite(%s, %s, %s)) }", indexTemp, indexTemp, lengthTemp, resultTemp, indexTemp, lengthTemp, handleTemp, indexTemp, valueTemp))
		}
		lines = append(lines, "__able_pop_call_frame()")
		if expected == "" || expected == "runtime.Value" {
			return lines, resultTemp, "runtime.Value", true
		}
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	default:
		return nil, "", "", false
	}
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
		baseName, ok := g.structBaseName(expected)
		if !ok {
			baseName = strings.TrimPrefix(expected, "*")
		}
		return fmt.Sprintf("func() %s { val := %s; v, err := __able_struct_%s_from(val); if err != nil { panic(err) }; return v }()", expected, valueExpr, baseName), true
	}
	return "", false
}

// expectRuntimeValueExprLines converts a runtime.Value expression to a concrete Go type
// using setup lines instead of an IIFE. Returns (lines, expression, ok).
func (g *generator) expectRuntimeValueExprLines(ctx *compileContext, valueExpr string, expected string) ([]string, string, bool) {
	// runtime.Value satisfies any — no conversion needed.
	if expected == "any" {
		return nil, valueExpr, true
	}
	valTemp := ctx.newTemp()
	vTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	switch g.typeCategory(expected) {
	case "bool":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsBool(%s)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, vTemp, true
	case "string":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsString(%s)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, vTemp, true
	case "rune":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsRune(%s)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, vTemp, true
	case "float32":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsFloat(%s)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, fmt.Sprintf("float32(%s)", vTemp), true
	case "float64":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsFloat(%s)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, vTemp, true
	case "int":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsInt(%s, bridge.NativeIntBits)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, fmt.Sprintf("int(%s)", vTemp), true
	case "uint":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsUint(%s, bridge.NativeIntBits)", vTemp, errTemp, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, fmt.Sprintf("uint(%s)", vTemp), true
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(expected)
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsInt(%s, %d)", vTemp, errTemp, valTemp, bits),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, fmt.Sprintf("%s(%s)", expected, vTemp), true
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(expected)
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsUint(%s, %d)", vTemp, errTemp, valTemp, bits),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, fmt.Sprintf("%s(%s)", expected, vTemp), true
	case "struct":
		baseName, ok := g.structBaseName(expected)
		if !ok {
			baseName = strings.TrimPrefix(expected, "*")
		}
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", vTemp, errTemp, baseName, valTemp),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, vTemp, true
	}
	return nil, "", false
}

// memberName extracts the string name from a member access expression.
func (g *generator) memberName(member ast.Expression) string {
	if ident, ok := member.(*ast.Identifier); ok && ident != nil {
		return ident.Name
	}
	return ""
}

// stringBuiltinFieldAccess maps Able String struct field names to Go string equivalents.
func (g *generator) stringBuiltinFieldAccess(objectExpr, fieldName string) (string, string, bool) {
	switch fieldName {
	case "len_bytes":
		return fmt.Sprintf("uint(len(%s))", objectExpr), "uint", true
	}
	return "", "", false
}

func (g *generator) isArrayStructType(goType string) bool {
	if g == nil || g.typeCategory(goType) != "struct" {
		return false
	}
	baseName, ok := g.structBaseName(goType)
	return ok && baseName == "Array"
}

func (g *generator) structInfoByGoName(goName string) *structInfo {
	if goName == "" {
		return nil
	}
	if strings.HasPrefix(goName, "*") {
		goName = strings.TrimPrefix(goName, "*")
	}
	for _, info := range g.structs {
		if info != nil && info.GoName == goName {
			return info
		}
	}
	return nil
}

func (g *generator) nativeIndexIntExpr(expr string, goType string) (string, bool) {
	switch g.typeCategory(goType) {
	case "int", "int8", "int16", "int32", "uint8", "uint16":
		return fmt.Sprintf("int(%s)", expr), true
	case "int64":
		return fmt.Sprintf("func() int { raw := int64(%s); idx := int(raw); if int64(idx) != raw { panic(fmt.Errorf(\"index overflows native int\")) }; return idx }()", expr), true
	case "uint", "uint32", "uint64":
		return fmt.Sprintf("func() int { raw := uint64(%s); idx := int(raw); if uint64(idx) != raw { panic(fmt.Errorf(\"index overflows native int\")) }; return idx }()", expr), true
	default:
		return "", false
	}
}

func (g *generator) appendIndexIntLines(ctx *compileContext, lines []string, idxExpr string, idxType string, idxTemp string, indexTemp string) ([]string, bool) {
	lines = append(lines, fmt.Sprintf("%s := %s", idxTemp, idxExpr))
	if nativeExpr, ok := g.nativeIndexIntExpr(idxTemp, idxType); ok {
		lines = append(lines, fmt.Sprintf("%s := %s", indexTemp, nativeExpr))
		return lines, true
	}
	idxConvLines, idxValueExpr, ok := g.runtimeValueLines(ctx, idxTemp, idxType)
	if !ok {
		return nil, false
	}
	lines = append(lines, idxConvLines...)
	idxRuntimeTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", idxRuntimeTemp, idxValueExpr))
	idxRawTemp := ctx.newTemp()
	idxErrTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := bridge.AsInt(%s, 64)", idxRawTemp, idxErrTemp, idxRuntimeTemp))
	lines = append(lines, fmt.Sprintf("if %s != nil { panic(%s) }", idxErrTemp, idxErrTemp))
	lines = append(lines, fmt.Sprintf("%s := int(%s)", indexTemp, idxRawTemp))
	return lines, true
}

func (g *generator) runtimeValueExpr(expr string, goType string) (string, bool) {
	switch g.typeCategory(goType) {
	case "runtime":
		return expr, true
	case "any":
		return fmt.Sprintf("__able_any_to_value(%s)", expr), true
	case "void":
		return fmt.Sprintf("func() runtime.Value { _ = %s; return runtime.VoidValue{} }()", expr), true
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
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		if strings.HasPrefix(goType, "*") {
			return fmt.Sprintf("__able_any_to_value(%s)", expr), true
		}
		return fmt.Sprintf("func() runtime.Value { if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }; v, err := __able_struct_%s_to(__able_runtime, %s); if err != nil { panic(err) }; return v }()", baseName, expr), true
	default:
		return "", false
	}
}

// runtimeValueLines is like runtimeValueExpr but returns setup lines + a clean
// expression instead of wrapping multi-statement conversions in IIFEs.
// Use this at call sites that have a lines slice and compileContext available.
func (g *generator) runtimeValueLines(ctx *compileContext, expr string, goType string) ([]string, string, bool) {
	switch g.typeCategory(goType) {
	case "void":
		return []string{fmt.Sprintf("_ = %s", expr)}, "runtime.VoidValue{}", true
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		// Struct pointers can be nil for nullable types (?T). Handle by
		// converting to runtime.Value first, using __able_any_to_value which
		// already has nil and struct pointer handling.
		if strings.HasPrefix(goType, "*") {
			convTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, expr),
			}
			return lines, convTemp, true
		}
		convTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		lines := []string{
			"if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }",
			fmt.Sprintf("%s, %s := __able_struct_%s_to(__able_runtime, %s)", convTemp, errTemp, baseName, expr),
			fmt.Sprintf("if %s != nil { panic(%s) }", errTemp, errTemp),
		}
		return lines, convTemp, true
	default:
		converted, ok := g.runtimeValueExpr(expr, goType)
		return nil, converted, ok
	}
}
