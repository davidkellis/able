package compiler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (e *irGoEmitter) emitValueRef(ref IRValueRef) (string, error) {
	switch v := ref.(type) {
	case IRValueUse:
		if v.Value == nil {
			return "", fmt.Errorf("compiler: value reference is nil")
		}
		return e.nameForValue(v.Value), nil
	case IRConst:
		return e.emitLiteral(v.Literal)
	case IRVoid:
		return "runtime.VoidValue{}", nil
	case IRGlobal:
		if v.Name == "" {
			return "", fmt.Errorf("compiler: global name is empty")
		}
		return fmt.Sprintf("__able_global(rt, %q)", v.Name), nil
	default:
		return "", fmt.Errorf("compiler: unsupported IR value ref %T", ref)
	}
}

func (e *irGoEmitter) emitArgs(args []IRValueRef) (string, error) {
	if len(args) == 0 {
		return "nil", nil
	}
	values := make([]string, 0, len(args))
	for _, arg := range args {
		expr, err := e.emitValueRef(arg)
		if err != nil {
			return "", err
		}
		values = append(values, expr)
	}
	return "[]runtime.Value{" + strings.Join(values, ", ") + "}", nil
}

func (e *irGoEmitter) emitLiteral(lit ast.Literal) (string, error) {
	switch node := lit.(type) {
	case *ast.StringLiteral:
		if node == nil {
			return "", fmt.Errorf("compiler: string literal is nil")
		}
		return fmt.Sprintf("runtime.StringValue{Val: %q}", node.Value), nil
	case *ast.BooleanLiteral:
		if node == nil {
			return "", fmt.Errorf("compiler: bool literal is nil")
		}
		return fmt.Sprintf("runtime.BoolValue{Val: %t}", node.Value), nil
	case *ast.NilLiteral:
		return "runtime.NilValue{}", nil
	case *ast.IntegerLiteral:
		if node == nil || node.Value == nil {
			return "", fmt.Errorf("compiler: integer literal is nil")
		}
		suffix := integerSuffix(node)
		return fmt.Sprintf("__able_int_literal(%q, %s)", node.Value.String(), suffix), nil
	case *ast.FloatLiteral:
		if node == nil {
			return "", fmt.Errorf("compiler: float literal is nil")
		}
		suffix := floatSuffix(node)
		value := strconv.FormatFloat(node.Value, 'g', -1, 64)
		return fmt.Sprintf("__able_float_literal(%s, %s)", value, suffix), nil
	case *ast.CharLiteral:
		if node == nil {
			return "", fmt.Errorf("compiler: char literal is nil")
		}
		runes := []rune(node.Value)
		if len(runes) != 1 {
			return "", fmt.Errorf("compiler: invalid char literal")
		}
		return fmt.Sprintf("runtime.CharValue{Val: %q}", runes[0]), nil
	default:
		return "", fmt.Errorf("compiler: unsupported literal %T", lit)
	}
}

func integerSuffix(lit *ast.IntegerLiteral) string {
	if lit == nil || lit.IntegerType == nil {
		return "runtime.IntegerI32"
	}
	switch *lit.IntegerType {
	case ast.IntegerTypeI8:
		return "runtime.IntegerI8"
	case ast.IntegerTypeI16:
		return "runtime.IntegerI16"
	case ast.IntegerTypeI32:
		return "runtime.IntegerI32"
	case ast.IntegerTypeI64:
		return "runtime.IntegerI64"
	case ast.IntegerTypeI128:
		return "runtime.IntegerI128"
	case ast.IntegerTypeU8:
		return "runtime.IntegerU8"
	case ast.IntegerTypeU16:
		return "runtime.IntegerU16"
	case ast.IntegerTypeU32:
		return "runtime.IntegerU32"
	case ast.IntegerTypeU64:
		return "runtime.IntegerU64"
	case ast.IntegerTypeU128:
		return "runtime.IntegerU128"
	default:
		return "runtime.IntegerI32"
	}
}

func floatSuffix(lit *ast.FloatLiteral) string {
	if lit == nil || lit.FloatType == nil {
		return "runtime.FloatF64"
	}
	switch *lit.FloatType {
	case ast.FloatTypeF32:
		return "runtime.FloatF32"
	case ast.FloatTypeF64:
		return "runtime.FloatF64"
	default:
		return "runtime.FloatF64"
	}
}

func (e *irGoEmitter) functionName(fn *IRFunction) string {
	if fn == nil {
		return ""
	}
	if name, ok := e.funcNames[fn]; ok {
		return name
	}
	pkg := sanitizeIdent(fn.Package)
	name := sanitizeIdent(fn.Name)
	if pkg == "" {
		pkg = "pkg"
	}
	if name == "" {
		name = "fn"
	}
	result := e.mangler.unique(fmt.Sprintf("__able_ir_%s_%s", pkg, name))
	e.funcNames[fn] = result
	return result
}

func (e *irGoEmitter) sortedLabels(fn *IRFunction) []string {
	labels := make([]string, 0, len(fn.Blocks))
	for label := range fn.Blocks {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	if fn.EntryLabel == "" {
		return labels
	}
	ordered := make([]string, 0, len(labels))
	ordered = append(ordered, fn.EntryLabel)
	for _, label := range labels {
		if label == fn.EntryLabel {
			continue
		}
		ordered = append(ordered, label)
	}
	return ordered
}

func (e *irGoEmitter) sortedDeclaredValues() []*IRValue {
	if len(e.declaredVals) == 0 {
		return nil
	}
	values := append([]*IRValue{}, e.declaredVals...)
	sort.Slice(values, func(i, j int) bool {
		if values[i] == nil || values[j] == nil {
			return values[j] != nil
		}
		if values[i].ID == values[j].ID {
			return values[i].Name < values[j].Name
		}
		return values[i].ID < values[j].ID
	})
	return values
}

func (e *irGoEmitter) valueNameFromRef(ref IRValueRef) (string, bool) {
	switch v := ref.(type) {
	case IRValueUse:
		if v.Value == nil || v.Value.Name == "" {
			return "", false
		}
		return v.Value.Name, true
	default:
		return "", false
	}
}

func (e *irGoEmitter) collectPatternSlots(pattern ast.Pattern) ([]*IRSlot, error) {
	if pattern == nil {
		return nil, nil
	}
	seen := make(map[*IRSlot]struct{})
	var walk func(pat ast.Pattern) error
	walk = func(pat ast.Pattern) error {
		if pat == nil {
			return nil
		}
		switch p := pat.(type) {
		case *ast.Identifier:
			if p.Name == "" || p.Name == "_" {
				return nil
			}
			slot := e.patternSlots[p]
			if slot == nil {
				return fmt.Errorf("compiler: missing slot for pattern %s", p.Name)
			}
			seen[slot] = struct{}{}
		case *ast.TypedPattern:
			return walk(p.Pattern)
		case *ast.StructPattern:
			for _, field := range p.Fields {
				if field == nil {
					continue
				}
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					slot := e.patternSlots[field]
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for binding %s", field.Binding.Name)
					}
					seen[slot] = struct{}{}
				}
				if err := walk(field.Pattern); err != nil {
					return err
				}
			}
		case *ast.ArrayPattern:
			for _, elem := range p.Elements {
				if err := walk(elem); err != nil {
					return err
				}
			}
			if p.RestPattern != nil {
				return walk(p.RestPattern)
			}
		}
		return nil
	}
	if err := walk(pattern); err != nil {
		return nil, err
	}
	slots := make([]*IRSlot, 0, len(seen))
	for slot := range seen {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		if slots[i] == nil || slots[j] == nil {
			return slots[j] != nil
		}
		if slots[i].Name == slots[j].Name {
			return fmt.Sprintf("%p", slots[i]) < fmt.Sprintf("%p", slots[j])
		}
		return slots[i].Name < slots[j].Name
	})
	return slots, nil
}

func (e *irGoEmitter) emitPattern(pattern ast.Pattern, valueExpr string, okVar string, errVar string, slotTemps map[*IRSlot]string) error {
	if pattern == nil {
		return nil
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name == "" || p.Name == "_" {
			return nil
		}
		slot := e.patternSlots[p]
		if slot == nil {
			return fmt.Errorf("compiler: missing slot for pattern %s", p.Name)
		}
		temp, ok := slotTemps[slot]
		if !ok {
			return fmt.Errorf("compiler: missing temp for pattern %s", p.Name)
		}
		fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
		fmt.Fprintf(&e.buf, "\t\t%s = %s\n", temp, valueExpr)
		fmt.Fprintf(&e.buf, "\t}\n")
		return nil
	case *ast.WildcardPattern:
		return nil
	case *ast.LiteralPattern:
		litExpr, err := e.emitLiteral(p.Literal)
		if err != nil {
			return err
		}
		cmpTemp := e.mangler.unique("pattern_cmp")
		fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
		fmt.Fprintf(&e.buf, "\t\t%s, err := bridge.ApplyBinaryOperator(rt, %q, %s, %s)\n", cmpTemp, "==", valueExpr, litExpr)
		fmt.Fprintf(&e.buf, "\t\tif err != nil { return nil, err }\n")
		fmt.Fprintf(&e.buf, "\t\tif %s == nil { %s = runtime.NilValue{} }\n", cmpTemp, cmpTemp)
		fmt.Fprintf(&e.buf, "\t\tif !__able_truthy(rt, %s) {\n", cmpTemp)
		e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
		fmt.Fprintf(&e.buf, "\t\t}\n")
		fmt.Fprintf(&e.buf, "\t}\n")
		return nil
	case *ast.TypedPattern:
		rendered, ok := e.renderTypeExpression(p.TypeAnnotation)
		if !ok {
			return fmt.Errorf("compiler: unsupported typed pattern")
		}
		matchTemp := e.mangler.unique("pattern_match")
		coercedTemp := e.mangler.unique("pattern_coerce")
		fmt.Fprintf(&e.buf, "\tvar %s runtime.Value\n", coercedTemp)
		fmt.Fprintf(&e.buf, "\tvar %s bool\n", matchTemp)
		fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
		fmt.Fprintf(&e.buf, "\t\t%s, %s, err = bridge.MatchType(rt, %s, %s)\n", coercedTemp, matchTemp, rendered, valueExpr)
		fmt.Fprintf(&e.buf, "\t\tif err != nil { return nil, err }\n")
		fmt.Fprintf(&e.buf, "\t\tif !%s {\n", matchTemp)
		e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
		fmt.Fprintf(&e.buf, "\t\t}\n")
		fmt.Fprintf(&e.buf, "\t}\n")
		return e.emitPattern(p.Pattern, coercedTemp, okVar, errVar, slotTemps)
	case *ast.StructPattern:
		if p.StructType != nil && p.StructType.Name == "IteratorEnd" && len(p.Fields) == 0 {
			iterOK := e.mangler.unique("pattern_iter_end")
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\t%s := false\n", iterOK)
			fmt.Fprintf(&e.buf, "\t\tswitch %s.(type) {\n", valueExpr)
			fmt.Fprintf(&e.buf, "\t\tcase runtime.IteratorEndValue, *runtime.IteratorEndValue:\n")
			fmt.Fprintf(&e.buf, "\t\t\t%s = true\n", iterOK)
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t\tif !%s {\n", iterOK)
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
			return nil
		}
		instTemp := e.mangler.unique("pattern_struct")
		fmt.Fprintf(&e.buf, "\tvar %s *runtime.StructInstanceValue\n", instTemp)
		fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
		fmt.Fprintf(&e.buf, "\t\t%s = __able_struct_instance(%s)\n", instTemp, valueExpr)
		fmt.Fprintf(&e.buf, "\t\tif %s == nil {\n", instTemp)
		e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
		fmt.Fprintf(&e.buf, "\t\t}\n")
		fmt.Fprintf(&e.buf, "\t}\n")
		if p.StructType != nil && p.StructType.Name != "" {
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\tif %s.Definition == nil || %s.Definition.Node == nil || %s.Definition.Node.ID == nil || %s.Definition.Node.ID.Name != %q {\n", instTemp, instTemp, instTemp, instTemp, p.StructType.Name)
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
		}
		if p.IsPositional {
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\tif %s.Positional == nil || len(%s.Positional) != %d {\n", instTemp, instTemp, len(p.Fields))
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
			for idx, field := range p.Fields {
				if field == nil || field.Pattern == nil {
					return fmt.Errorf("compiler: invalid positional struct pattern")
				}
				fieldExpr := fmt.Sprintf("%s.Positional[%d]", instTemp, idx)
				if err := e.emitPattern(field.Pattern, fieldExpr, okVar, errVar, slotTemps); err != nil {
					return err
				}
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					slot := e.patternSlots[field]
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for binding %s", field.Binding.Name)
					}
					temp := slotTemps[slot]
					fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
					fmt.Fprintf(&e.buf, "\t\t%s = %s\n", temp, fieldExpr)
					fmt.Fprintf(&e.buf, "\t}\n")
				}
			}
			return nil
		}
		if len(p.Fields) > 0 {
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\tif %s.Fields == nil {\n", instTemp)
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
		}
		for _, field := range p.Fields {
			if field == nil || field.FieldName == nil || field.FieldName.Name == "" {
				return fmt.Errorf("compiler: named struct pattern missing field name")
			}
			fieldTemp := e.mangler.unique("pattern_field")
			fieldOk := e.mangler.unique("pattern_field_ok")
			fmt.Fprintf(&e.buf, "\tvar %s runtime.Value\n", fieldTemp)
			fmt.Fprintf(&e.buf, "\tvar %s bool\n", fieldOk)
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\t%s, %s = %s.Fields[%q]\n", fieldTemp, fieldOk, instTemp, field.FieldName.Name)
			fmt.Fprintf(&e.buf, "\t\tif !%s {\n", fieldOk)
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
			if err := e.emitPattern(field.Pattern, fieldTemp, okVar, errVar, slotTemps); err != nil {
				return err
			}
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				slot := e.patternSlots[field]
				if slot == nil {
					return fmt.Errorf("compiler: missing slot for binding %s", field.Binding.Name)
				}
				temp := slotTemps[slot]
				fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
				fmt.Fprintf(&e.buf, "\t\t%s = %s\n", temp, fieldTemp)
				fmt.Fprintf(&e.buf, "\t}\n")
			}
		}
		return nil
	case *ast.ArrayPattern:
		arrTemp := e.mangler.unique("pattern_array")
		fmt.Fprintf(&e.buf, "\tvar %s *runtime.ArrayValue\n", arrTemp)
		fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
		fmt.Fprintf(&e.buf, "\t\tswitch v := %s.(type) {\n", valueExpr)
		fmt.Fprintf(&e.buf, "\t\tcase *runtime.ArrayValue:\n")
		fmt.Fprintf(&e.buf, "\t\t\t%s = v\n", arrTemp)
		fmt.Fprintf(&e.buf, "\t\tcase runtime.ArrayValue:\n")
		fmt.Fprintf(&e.buf, "\t\t\t%s = &v\n", arrTemp)
		fmt.Fprintf(&e.buf, "\t\t}\n")
		fmt.Fprintf(&e.buf, "\t\tif %s == nil {\n", arrTemp)
		e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
		fmt.Fprintf(&e.buf, "\t\t}\n")
		fmt.Fprintf(&e.buf, "\t}\n")
		if p.RestPattern == nil {
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\tif %s == nil || len(%s.Elements) != %d {\n", arrTemp, arrTemp, len(p.Elements))
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
		} else {
			fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
			fmt.Fprintf(&e.buf, "\t\tif %s == nil || len(%s.Elements) < %d {\n", arrTemp, arrTemp, len(p.Elements))
			e.emitPatternMismatch(okVar, errVar, "pattern mismatch")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t}\n")
		}
		for idx, elem := range p.Elements {
			if elem == nil {
				return fmt.Errorf("compiler: invalid array pattern element")
			}
			elemExpr := fmt.Sprintf("%s.Elements[%d]", arrTemp, idx)
			if err := e.emitPattern(elem, elemExpr, okVar, errVar, slotTemps); err != nil {
				return err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				if rest.Name != "" && rest.Name != "_" {
					slot := e.patternSlots[rest]
					if slot == nil {
						return fmt.Errorf("compiler: missing slot for rest binding %s", rest.Name)
					}
					temp := slotTemps[slot]
					restTemp := e.mangler.unique("pattern_rest")
					fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
					fmt.Fprintf(&e.buf, "\t\t%s := append([]runtime.Value(nil), %s.Elements[%d:]...)\n", restTemp, arrTemp, len(p.Elements))
					fmt.Fprintf(&e.buf, "\t\t%s = &runtime.ArrayValue{Elements: %s}\n", temp, restTemp)
					fmt.Fprintf(&e.buf, "\t}\n")
				}
			case *ast.WildcardPattern:
				// ignore
			default:
				return fmt.Errorf("compiler: unsupported rest pattern type %s", rest.NodeType())
			}
		}
		return nil
	default:
		return fmt.Errorf("compiler: unsupported pattern %T", pattern)
	}
}

func (e *irGoEmitter) emitPatternMismatch(okVar string, errVar string, message string) {
	fmt.Fprintf(&e.buf, "\t\t%s = false\n", okVar)
	fmt.Fprintf(&e.buf, "\t\t%s = runtime.ErrorValue{Message: %q}\n", errVar, message)
}

func (e *irGoEmitter) renderTypeExpression(expr ast.TypeExpression) (string, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return "", false
		}
		return fmt.Sprintf("ast.Ty(%q)", t.Name.Name), true
	case *ast.GenericTypeExpression:
		if t == nil {
			return "", false
		}
		base, ok := e.renderTypeExpression(t.Base)
		if !ok {
			return "", false
		}
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			rendered, ok := e.renderTypeExpression(arg)
			if !ok {
				return "", false
			}
			args = append(args, rendered)
		}
		if len(args) == 0 {
			return fmt.Sprintf("ast.Gen(%s)", base), true
		}
		return fmt.Sprintf("ast.Gen(%s, %s)", base, strings.Join(args, ", ")), true
	case *ast.FunctionTypeExpression:
		if t == nil {
			return "", false
		}
		params := make([]string, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			rendered, ok := e.renderTypeExpression(param)
			if !ok {
				return "", false
			}
			params = append(params, rendered)
		}
		ret, ok := e.renderTypeExpression(t.ReturnType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.FnType([]ast.TypeExpression{%s}, %s)", strings.Join(params, ", "), ret), true
	case *ast.NullableTypeExpression:
		if t == nil {
			return "", false
		}
		inner, ok := e.renderTypeExpression(t.InnerType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.Nullable(%s)", inner), true
	case *ast.ResultTypeExpression:
		if t == nil {
			return "", false
		}
		inner, ok := e.renderTypeExpression(t.InnerType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.Result(%s)", inner), true
	case *ast.UnionTypeExpression:
		if t == nil {
			return "", false
		}
		members := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			rendered, ok := e.renderTypeExpression(member)
			if !ok {
				return "", false
			}
			members = append(members, rendered)
		}
		return fmt.Sprintf("ast.Union(%s)", strings.Join(members, ", ")), true
	case *ast.WildcardTypeExpression:
		return "ast.WildcardType()", true
	default:
		return "", false
	}
}
