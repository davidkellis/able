package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func (g *generator) renderRuntimeInterfaceConstraintSupport(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}

	knownTypeNames := g.runtimeInterfaceKnownTypeNames()
	if len(knownTypeNames) > 0 {
		fmt.Fprintf(buf, "var __able_known_type_names = map[string]struct{}{\n")
		for _, name := range knownTypeNames {
			fmt.Fprintf(buf, "\t%q: {},\n", name)
		}
		fmt.Fprintf(buf, "}\n\n")
	} else {
		fmt.Fprintf(buf, "var __able_known_type_names map[string]struct{}\n\n")
	}

	ifaceMethodNames := g.runtimeInterfaceMethodNames()
	if len(ifaceMethodNames) > 0 {
		fmt.Fprintf(buf, "var __able_interface_method_names = map[string][]string{\n")
		names := make([]string, 0, len(ifaceMethodNames))
		for name := range ifaceMethodNames {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			methods := ifaceMethodNames[name]
			fmt.Fprintf(buf, "\t%q: {", name)
			for idx, method := range methods {
				if idx > 0 {
					fmt.Fprintf(buf, ", ")
				}
				fmt.Fprintf(buf, "%q", method)
			}
			fmt.Fprintf(buf, "},\n")
		}
		fmt.Fprintf(buf, "}\n\n")
	} else {
		fmt.Fprintf(buf, "var __able_interface_method_names map[string][]string\n\n")
	}

	ifaceGenericParams := g.runtimeInterfaceGenericParamNames()
	if len(ifaceGenericParams) > 0 {
		fmt.Fprintf(buf, "var __able_interface_generic_param_names = map[string][]string{\n")
		names := make([]string, 0, len(ifaceGenericParams))
		for name := range ifaceGenericParams {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			params := ifaceGenericParams[name]
			fmt.Fprintf(buf, "\t%q: {", name)
			for idx, param := range params {
				if idx > 0 {
					fmt.Fprintf(buf, ", ")
				}
				fmt.Fprintf(buf, "%q", param)
			}
			fmt.Fprintf(buf, "},\n")
		}
		fmt.Fprintf(buf, "}\n\n")
	} else {
		fmt.Fprintf(buf, "var __able_interface_generic_param_names map[string][]string\n\n")
	}

	aliasDefs := g.runtimeInterfaceAliasDefinitions()
	if len(aliasDefs) > 0 {
		fmt.Fprintf(buf, "var __able_type_alias_defs = map[string]*ast.TypeAliasDefinition{\n")
		aliasNames := make([]string, 0, len(aliasDefs))
		for name := range aliasDefs {
			aliasNames = append(aliasNames, name)
		}
		sort.Strings(aliasNames)
		for _, name := range aliasNames {
			defExpr := aliasDefs[name]
			if strings.TrimSpace(defExpr) == "" {
				continue
			}
			fmt.Fprintf(buf, "\t%q: %s,\n", name, defExpr)
		}
		fmt.Fprintf(buf, "}\n\n")
	} else {
		fmt.Fprintf(buf, "var __able_type_alias_defs map[string]*ast.TypeAliasDefinition\n\n")
	}

	fmt.Fprintf(buf, "func __able_is_primitive_type_name(name string) bool {\n")
	fmt.Fprintf(buf, "\tswitch name {\n")
	fmt.Fprintf(buf, "\tcase \"bool\", \"String\", \"string\", \"IoHandle\", \"ProcHandle\", \"char\", \"nil\", \"void\":\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch name {\n")
	fmt.Fprintf(buf, "\tcase \"i8\", \"i16\", \"i32\", \"i64\", \"i128\", \"u8\", \"u16\", \"u32\", \"u64\", \"u128\", \"isize\", \"usize\":\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\tcase \"f32\", \"f64\":\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn false\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_is_known_type_name(name string) bool {\n")
	fmt.Fprintf(buf, "\tif name == \"\" || name == \"_\" || name == \"Self\" {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif __able_is_primitive_type_name(name) {\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif __able_known_type_names == nil {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t_, ok := __able_known_type_names[name]\n")
	fmt.Fprintf(buf, "\treturn ok\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_type_expr_has_unknown_names(expr ast.TypeExpression) bool {\n")
	fmt.Fprintf(buf, "\tswitch t := expr.(type) {\n")
	fmt.Fprintf(buf, "\tcase *ast.SimpleTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tif t.Name == nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn !__able_is_known_type_name(t.Name.Name)\n")
	fmt.Fprintf(buf, "\tcase *ast.GenericTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tif __able_type_expr_has_unknown_names(t.Base) {\n")
	fmt.Fprintf(buf, "\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tfor _, arg := range t.Arguments {\n")
	fmt.Fprintf(buf, "\t\t\tif __able_type_expr_has_unknown_names(arg) {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\tcase *ast.FunctionTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tif __able_type_expr_has_unknown_names(t.ReturnType) {\n")
	fmt.Fprintf(buf, "\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tfor _, param := range t.ParamTypes {\n")
	fmt.Fprintf(buf, "\t\t\tif __able_type_expr_has_unknown_names(param) {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\tcase *ast.NullableTypeExpression:\n")
	fmt.Fprintf(buf, "\t\treturn __able_type_expr_has_unknown_names(t.InnerType)\n")
	fmt.Fprintf(buf, "\tcase *ast.ResultTypeExpression:\n")
	fmt.Fprintf(buf, "\t\treturn __able_type_expr_has_unknown_names(t.InnerType)\n")
	fmt.Fprintf(buf, "\tcase *ast.UnionTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tfor _, member := range t.Members {\n")
	fmt.Fprintf(buf, "\t\t\tif __able_type_expr_has_unknown_names(member) {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_substitute_type_params(expr ast.TypeExpression, bindings map[string]ast.TypeExpression) ast.TypeExpression {\n")
	fmt.Fprintf(buf, "\tif expr == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch t := expr.(type) {\n")
	fmt.Fprintf(buf, "\tcase *ast.SimpleTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tif t.Name != nil {\n")
	fmt.Fprintf(buf, "\t\t\tif replacement, ok := bindings[t.Name.Name]; ok {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn replacement\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn t\n")
	fmt.Fprintf(buf, "\tcase *ast.GenericTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tbase := __able_substitute_type_params(t.Base, bindings)\n")
	fmt.Fprintf(buf, "\t\targs := make([]ast.TypeExpression, len(t.Arguments))\n")
	fmt.Fprintf(buf, "\t\tfor idx, arg := range t.Arguments {\n")
	fmt.Fprintf(buf, "\t\t\targs[idx] = __able_substitute_type_params(arg, bindings)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewGenericTypeExpression(base, args)\n")
	fmt.Fprintf(buf, "\tcase *ast.FunctionTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tparams := make([]ast.TypeExpression, len(t.ParamTypes))\n")
	fmt.Fprintf(buf, "\t\tfor idx, param := range t.ParamTypes {\n")
	fmt.Fprintf(buf, "\t\t\tparams[idx] = __able_substitute_type_params(param, bindings)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewFunctionTypeExpression(params, __able_substitute_type_params(t.ReturnType, bindings))\n")
	fmt.Fprintf(buf, "\tcase *ast.NullableTypeExpression:\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewNullableTypeExpression(__able_substitute_type_params(t.InnerType, bindings))\n")
	fmt.Fprintf(buf, "\tcase *ast.ResultTypeExpression:\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewResultTypeExpression(__able_substitute_type_params(t.InnerType, bindings))\n")
	fmt.Fprintf(buf, "\tcase *ast.UnionTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tmembers := make([]ast.TypeExpression, len(t.Members))\n")
	fmt.Fprintf(buf, "\t\tfor idx, member := range t.Members {\n")
	fmt.Fprintf(buf, "\t\t\tmembers[idx] = __able_substitute_type_params(member, bindings)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewUnionTypeExpression(members)\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn expr\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_expand_runtime_type_aliases(expr ast.TypeExpression, seen map[string]struct{}) ast.TypeExpression {\n")
	fmt.Fprintf(buf, "\tif expr == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif seen == nil {\n")
	fmt.Fprintf(buf, "\t\tseen = make(map[string]struct{})\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tkey := __able_type_expr_string(expr)\n")
	fmt.Fprintf(buf, "\tif key != \"\" {\n")
	fmt.Fprintf(buf, "\t\tif _, ok := seen[key]; ok {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch t := expr.(type) {\n")
	fmt.Fprintf(buf, "\tcase *ast.SimpleTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tif t == nil || t.Name == nil || __able_type_alias_defs == nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif def := __able_type_alias_defs[t.Name.Name]; def != nil && def.TargetType != nil && len(def.GenericParams) == 0 {\n")
	fmt.Fprintf(buf, "\t\t\tif key != \"\" {\n")
	fmt.Fprintf(buf, "\t\t\t\tseen[key] = struct{}{}\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn __able_expand_runtime_type_aliases(def.TargetType, seen)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn expr\n")
	fmt.Fprintf(buf, "\tcase *ast.GenericTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tbase := __able_expand_runtime_type_aliases(t.Base, seen)\n")
	fmt.Fprintf(buf, "\t\targs := make([]ast.TypeExpression, len(t.Arguments))\n")
	fmt.Fprintf(buf, "\t\tchanged := base != t.Base\n")
	fmt.Fprintf(buf, "\t\tfor idx, arg := range t.Arguments {\n")
	fmt.Fprintf(buf, "\t\t\targs[idx] = __able_expand_runtime_type_aliases(arg, seen)\n")
	fmt.Fprintf(buf, "\t\t\tif args[idx] != arg {\n")
	fmt.Fprintf(buf, "\t\t\t\tchanged = true\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif simple, ok := base.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && __able_type_alias_defs != nil {\n")
	fmt.Fprintf(buf, "\t\t\tif def := __able_type_alias_defs[simple.Name.Name]; def != nil && def.TargetType != nil && len(def.GenericParams) == len(args) {\n")
	fmt.Fprintf(buf, "\t\t\t\tif key != \"\" {\n")
	fmt.Fprintf(buf, "\t\t\t\t\tseen[key] = struct{}{}\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\tbindings := make(map[string]ast.TypeExpression)\n")
	fmt.Fprintf(buf, "\t\t\t\tfor idx, param := range def.GenericParams {\n")
	fmt.Fprintf(buf, "\t\t\t\t\tif param == nil || param.Name == nil || param.Name.Name == \"\" {\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\t\tbindings[param.Name.Name] = args[idx]\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\texpanded := __able_substitute_type_params(def.TargetType, bindings)\n")
	fmt.Fprintf(buf, "\t\t\t\treturn __able_expand_runtime_type_aliases(expanded, seen)\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !changed {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewGenericTypeExpression(base, args)\n")
	fmt.Fprintf(buf, "\tcase *ast.FunctionTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tparams := make([]ast.TypeExpression, len(t.ParamTypes))\n")
	fmt.Fprintf(buf, "\t\tchanged := false\n")
	fmt.Fprintf(buf, "\t\tfor idx, param := range t.ParamTypes {\n")
	fmt.Fprintf(buf, "\t\t\tparams[idx] = __able_expand_runtime_type_aliases(param, seen)\n")
	fmt.Fprintf(buf, "\t\t\tif params[idx] != param {\n")
	fmt.Fprintf(buf, "\t\t\t\tchanged = true\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tret := __able_expand_runtime_type_aliases(t.ReturnType, seen)\n")
	fmt.Fprintf(buf, "\t\tif ret != t.ReturnType {\n")
	fmt.Fprintf(buf, "\t\t\tchanged = true\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !changed {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewFunctionTypeExpression(params, ret)\n")
	fmt.Fprintf(buf, "\tcase *ast.NullableTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tinner := __able_expand_runtime_type_aliases(t.InnerType, seen)\n")
	fmt.Fprintf(buf, "\t\tif inner == t.InnerType {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewNullableTypeExpression(inner)\n")
	fmt.Fprintf(buf, "\tcase *ast.ResultTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tinner := __able_expand_runtime_type_aliases(t.InnerType, seen)\n")
	fmt.Fprintf(buf, "\t\tif inner == t.InnerType {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewResultTypeExpression(inner)\n")
	fmt.Fprintf(buf, "\tcase *ast.UnionTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tmembers := make([]ast.TypeExpression, len(t.Members))\n")
	fmt.Fprintf(buf, "\t\tchanged := false\n")
	fmt.Fprintf(buf, "\t\tfor idx, member := range t.Members {\n")
	fmt.Fprintf(buf, "\t\t\tmembers[idx] = __able_expand_runtime_type_aliases(member, seen)\n")
	fmt.Fprintf(buf, "\t\t\tif members[idx] != member {\n")
	fmt.Fprintf(buf, "\t\t\t\tchanged = true\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !changed {\n")
	fmt.Fprintf(buf, "\t\t\treturn expr\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn ast.NewUnionTypeExpression(members)\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn expr\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_collect_interface_constraint_strings(typeExpr ast.TypeExpression, memo map[string]struct{}) []string {\n")
	fmt.Fprintf(buf, "\tif typeExpr == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tkey := __able_type_expr_string(typeExpr)\n")
	fmt.Fprintf(buf, "\tif _, seen := memo[key]; seen {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmemo[key] = struct{}{}\n")
	fmt.Fprintf(buf, "\tresults := []string{key}\n")
	fmt.Fprintf(buf, "\tif simple, ok := typeExpr.(*ast.SimpleTypeExpression); ok && simple.Name != nil {\n")
	fmt.Fprintf(buf, "\t\tif bases := __able_interface_base_types[simple.Name.Name]; len(bases) > 0 {\n")
	fmt.Fprintf(buf, "\t\t\tfor _, base := range bases {\n")
	fmt.Fprintf(buf, "\t\t\t\tresults = append(results, __able_collect_interface_constraint_strings(base, memo)...)\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn results\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_constraint_key_set(constraints []__able_interface_constraint_spec) map[string]struct{} {\n")
	fmt.Fprintf(buf, "\tif len(constraints) == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tset := make(map[string]struct{})\n")
	fmt.Fprintf(buf, "\tfor _, c := range constraints {\n")
	fmt.Fprintf(buf, "\t\tif c.iface == nil {\n")
	fmt.Fprintf(buf, "\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\texpressions := __able_collect_interface_constraint_strings(c.iface, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\t\tfor _, expr := range expressions {\n")
	fmt.Fprintf(buf, "\t\t\tkey := fmt.Sprintf(\"%%s->%%s\", __able_type_expr_string(c.subject), expr)\n")
	fmt.Fprintf(buf, "\t\t\tset[key] = struct{}{}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn set\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_is_constraint_superset(a, b map[string]struct{}) bool {\n")
	fmt.Fprintf(buf, "\tif len(a) <= len(b) {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tfor key := range b {\n")
	fmt.Fprintf(buf, "\t\tif _, ok := a[key]; !ok {\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn true\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_is_proper_subset(a, b []string) bool {\n")
	fmt.Fprintf(buf, "\tif len(a) == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn len(b) > 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tsetA := make(map[string]struct{}, len(a))\n")
	fmt.Fprintf(buf, "\tfor _, val := range a {\n")
	fmt.Fprintf(buf, "\t\tsetA[val] = struct{}{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tsetB := make(map[string]struct{}, len(b))\n")
	fmt.Fprintf(buf, "\tfor _, val := range b {\n")
	fmt.Fprintf(buf, "\t\tsetB[val] = struct{}{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif len(setA) >= len(setB) {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tfor val := range setA {\n")
	fmt.Fprintf(buf, "\t\tif _, ok := setB[val]; !ok {\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn true\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_compare_interface_entries(a, b __able_interface_dispatch_entry) int {\n")
	fmt.Fprintf(buf, "\tif a.isConcrete && !b.isConcrete {\n")
	fmt.Fprintf(buf, "\t\treturn 1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif b.isConcrete && !a.isConcrete {\n")
	fmt.Fprintf(buf, "\t\treturn -1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif __able_is_constraint_superset(a.constraintKeys, b.constraintKeys) {\n")
	fmt.Fprintf(buf, "\t\treturn 1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif __able_is_constraint_superset(b.constraintKeys, a.constraintKeys) {\n")
	fmt.Fprintf(buf, "\t\treturn -1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\taUnion := a.unionVariants\n")
	fmt.Fprintf(buf, "\tbUnion := b.unionVariants\n")
	fmt.Fprintf(buf, "\tif len(aUnion) > 0 && len(bUnion) == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn -1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif len(aUnion) == 0 && len(bUnion) > 0 {\n")
	fmt.Fprintf(buf, "\t\treturn 1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif len(aUnion) > 0 && len(bUnion) > 0 {\n")
	fmt.Fprintf(buf, "\t\tif __able_is_proper_subset(aUnion, bUnion) {\n")
	fmt.Fprintf(buf, "\t\t\treturn 1\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif __able_is_proper_subset(bUnion, aUnion) {\n")
	fmt.Fprintf(buf, "\t\t\treturn -1\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif len(aUnion) != len(bUnion) {\n")
	fmt.Fprintf(buf, "\t\t\tif len(aUnion) < len(bUnion) {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn 1\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn -1\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif a.targetScore > b.targetScore {\n")
	fmt.Fprintf(buf, "\t\treturn 1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif a.targetScore < b.targetScore {\n")
	fmt.Fprintf(buf, "\t\treturn -1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif a.isBuiltin != b.isBuiltin {\n")
	fmt.Fprintf(buf, "\t\tif a.isBuiltin {\n")
	fmt.Fprintf(buf, "\t\t\treturn -1\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn 1\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn 0\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_interface_type_parts(expr ast.TypeExpression) (string, []ast.TypeExpression, bool) {\n")
	fmt.Fprintf(buf, "\texpr = __able_expand_runtime_type_aliases(expr, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\tswitch t := expr.(type) {\n")
	fmt.Fprintf(buf, "\tcase *ast.SimpleTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tif t == nil || t.Name == nil || t.Name.Name == \"\" {\n")
	fmt.Fprintf(buf, "\t\t\treturn \"\", nil, false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !__able_is_interface_type(t) {\n")
	fmt.Fprintf(buf, "\t\t\treturn \"\", nil, false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn t.Name.Name, nil, true\n")
	fmt.Fprintf(buf, "\tcase *ast.GenericTypeExpression:\n")
	fmt.Fprintf(buf, "\t\tbase, ok := t.Base.(*ast.SimpleTypeExpression)\n")
	fmt.Fprintf(buf, "\t\tif !ok || base == nil || base.Name == nil || base.Name.Name == \"\" {\n")
	fmt.Fprintf(buf, "\t\t\treturn \"\", nil, false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !__able_is_interface_type(base) {\n")
	fmt.Fprintf(buf, "\t\t\treturn \"\", nil, false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn base.Name.Name, t.Arguments, true\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn \"\", nil, false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_interface_generic_bindings(ifaceName string, ifaceArgs []ast.TypeExpression) map[string]ast.TypeExpression {\n")
	fmt.Fprintf(buf, "\tparams := __able_interface_generic_param_names[ifaceName]\n")
	fmt.Fprintf(buf, "\tif len(params) == 0 || len(ifaceArgs) == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbindings := make(map[string]ast.TypeExpression)\n")
	fmt.Fprintf(buf, "\tfor idx, name := range params {\n")
	fmt.Fprintf(buf, "\t\tif idx >= len(ifaceArgs) || name == \"\" {\n")
	fmt.Fprintf(buf, "\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tbindings[name] = ifaceArgs[idx]\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn bindings\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_interface_dispatch_entry_accepts_type(subject ast.TypeExpression, ifaceArgs []ast.TypeExpression, entry __able_interface_dispatch_entry, seen map[string]struct{}) bool {\n")
	fmt.Fprintf(buf, "\tif subject == nil {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tsubject = __able_expand_runtime_type_aliases(subject, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\ttemplate := __able_expand_runtime_type_aliases(entry.targetType, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\tbindings := make(map[string]ast.TypeExpression)\n")
	fmt.Fprintf(buf, "\tif !__able_match_type_template(template, subject, entry.genericNames, bindings) {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif len(ifaceArgs) > 0 {\n")
	fmt.Fprintf(buf, "\t\tif len(entry.interfaceArgs) == 0 || len(entry.interfaceArgs) != len(ifaceArgs) {\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tfor idx := range entry.interfaceArgs {\n")
	fmt.Fprintf(buf, "\t\t\tleft := __able_expand_runtime_type_aliases(entry.interfaceArgs[idx], make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\t\t\tright := __able_expand_runtime_type_aliases(ifaceArgs[idx], make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\t\t\tif !__able_match_type_template(left, right, entry.genericNames, bindings) {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn __able_enforce_constraints_seen(entry.constraints, bindings, seen)\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_type_expr_satisfies_interface_seen(subject ast.TypeExpression, iface ast.TypeExpression, seen map[string]struct{}) bool {\n")
	fmt.Fprintf(buf, "\tsubject = __able_expand_runtime_type_aliases(subject, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\tiface = __able_expand_runtime_type_aliases(iface, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\tif subject == nil || iface == nil {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif seen == nil {\n")
	fmt.Fprintf(buf, "\t\tseen = make(map[string]struct{})\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tkey := __able_type_expr_string(subject) + \"::\" + __able_type_expr_string(iface)\n")
	fmt.Fprintf(buf, "\tif _, ok := seen[key]; ok {\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tseen[key] = struct{}{}\n")
	fmt.Fprintf(buf, "\tifaceName, ifaceArgs, ok := __able_interface_type_parts(iface)\n")
	fmt.Fprintf(buf, "\tif !ok || ifaceName == \"\" {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmethods := __able_interface_method_names[ifaceName]\n")
	fmt.Fprintf(buf, "\tif len(methods) == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tfor _, methodName := range methods {\n")
	fmt.Fprintf(buf, "\t\tentries := []__able_interface_dispatch_entry(nil)\n")
	fmt.Fprintf(buf, "\t\tif perIface := __able_interface_dispatch[ifaceName]; perIface != nil {\n")
	fmt.Fprintf(buf, "\t\t\tentries = perIface[methodName]\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif len(entries) == 0 {\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tmatched := false\n")
	fmt.Fprintf(buf, "\t\tfor _, entry := range entries {\n")
	fmt.Fprintf(buf, "\t\t\tif __able_interface_dispatch_entry_accepts_type(subject, ifaceArgs, entry, seen) {\n")
	fmt.Fprintf(buf, "\t\t\t\tmatched = true\n")
	fmt.Fprintf(buf, "\t\t\t\tbreak\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !matched {\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbindings := __able_interface_generic_bindings(ifaceName, ifaceArgs)\n")
	fmt.Fprintf(buf, "\tif bases := __able_interface_base_types[ifaceName]; len(bases) > 0 {\n")
	fmt.Fprintf(buf, "\t\tfor _, base := range bases {\n")
	fmt.Fprintf(buf, "\t\t\tnext := base\n")
	fmt.Fprintf(buf, "\t\t\tif bindings != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\tnext = __able_substitute_type_params(base, bindings)\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tif !__able_type_expr_satisfies_interface_seen(subject, next, seen) {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn true\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_type_expr_satisfies_interface(subject ast.TypeExpression, iface ast.TypeExpression) bool {\n")
	fmt.Fprintf(buf, "\treturn __able_type_expr_satisfies_interface_seen(subject, iface, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_enforce_constraints_seen(constraints []__able_interface_constraint_spec, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {\n")
	fmt.Fprintf(buf, "\tif len(constraints) == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif bindings == nil {\n")
	fmt.Fprintf(buf, "\t\tbindings = map[string]ast.TypeExpression{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tfor _, spec := range constraints {\n")
	fmt.Fprintf(buf, "\t\tsubject := __able_expand_runtime_type_aliases(__able_substitute_type_params(spec.subject, bindings), make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\t\tiface := __able_expand_runtime_type_aliases(spec.iface, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "\t\tif __able_type_expr_has_unknown_names(subject) || __able_type_expr_has_unknown_names(iface) {\n")
	fmt.Fprintf(buf, "\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif !__able_type_expr_satisfies_interface_seen(subject, iface, seen) {\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn true\n")
	fmt.Fprintf(buf, "}\n\n")

	fmt.Fprintf(buf, "func __able_enforce_constraints(constraints []__able_interface_constraint_spec, bindings map[string]ast.TypeExpression) bool {\n")
	fmt.Fprintf(buf, "\treturn __able_enforce_constraints_seen(constraints, bindings, make(map[string]struct{}))\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) runtimeInterfaceKnownTypeNames() []string {
	if g == nil {
		return nil
	}
	known := make(map[string]struct{})
	for name := range g.structs {
		if strings.TrimSpace(name) != "" {
			known[name] = struct{}{}
		}
	}
	for name := range g.interfaces {
		if strings.TrimSpace(name) != "" {
			known[name] = struct{}{}
		}
	}
	for name := range g.unions {
		if strings.TrimSpace(name) != "" {
			known[name] = struct{}{}
		}
	}
	for _, perPkg := range g.typeAliases {
		for name := range perPkg {
			if strings.TrimSpace(name) != "" {
				known[name] = struct{}{}
			}
		}
	}
	if len(known) == 0 {
		return nil
	}
	names := make([]string, 0, len(known))
	for name := range known {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) runtimeInterfaceMethodNames() map[string][]string {
	if g == nil || len(g.interfaces) == 0 {
		return nil
	}
	out := make(map[string][]string)
	for name, def := range g.interfaces {
		if strings.TrimSpace(name) == "" || def == nil || len(def.Signatures) == 0 {
			continue
		}
		methods := make([]string, 0, len(def.Signatures))
		for _, sig := range def.Signatures {
			if sig == nil || sig.Name == nil || strings.TrimSpace(sig.Name.Name) == "" {
				continue
			}
			methods = append(methods, sig.Name.Name)
		}
		if len(methods) == 0 {
			continue
		}
		sort.Strings(methods)
		out[name] = methods
	}
	return out
}

func (g *generator) runtimeInterfaceGenericParamNames() map[string][]string {
	if g == nil || len(g.interfaces) == 0 {
		return nil
	}
	out := make(map[string][]string)
	for name, def := range g.interfaces {
		if strings.TrimSpace(name) == "" || def == nil || len(def.GenericParams) == 0 {
			continue
		}
		params := make([]string, 0, len(def.GenericParams))
		for _, gp := range def.GenericParams {
			if gp == nil || gp.Name == nil || strings.TrimSpace(gp.Name.Name) == "" {
				continue
			}
			params = append(params, gp.Name.Name)
		}
		if len(params) == 0 {
			continue
		}
		out[name] = params
	}
	return out
}

func (g *generator) runtimeInterfaceAliasDefinitions() map[string]string {
	if g == nil || len(g.typeAliases) == 0 {
		return nil
	}
	type renderedAlias struct {
		defExpr string
		unique  bool
	}
	rendered := make(map[string]renderedAlias)
	pkgs := make([]string, 0, len(g.typeAliases))
	for pkgName := range g.typeAliases {
		pkgs = append(pkgs, pkgName)
	}
	sort.Strings(pkgs)
	for _, pkgName := range pkgs {
		perPkg := g.typeAliases[pkgName]
		if len(perPkg) == 0 {
			continue
		}
		aliasNames := make([]string, 0, len(perPkg))
		for aliasName := range perPkg {
			aliasNames = append(aliasNames, aliasName)
		}
		sort.Strings(aliasNames)
		for _, aliasName := range aliasNames {
			target := perPkg[aliasName]
			if target == nil || strings.TrimSpace(aliasName) == "" {
				continue
			}
			targetExpr, ok := g.renderTypeExpression(target)
			if !ok || strings.TrimSpace(targetExpr) == "" {
				continue
			}
			defExpr := ""
			if genericParamsExpr := g.renderTypeAliasGenericParams(pkgName, aliasName); genericParamsExpr != "" {
				defExpr = fmt.Sprintf("&ast.TypeAliasDefinition{ID: ast.NewIdentifier(%q), TargetType: %s, GenericParams: %s}", aliasName, targetExpr, genericParamsExpr)
			} else {
				defExpr = fmt.Sprintf("&ast.TypeAliasDefinition{ID: ast.NewIdentifier(%q), TargetType: %s}", aliasName, targetExpr)
			}
			existing, ok := rendered[aliasName]
			if !ok {
				rendered[aliasName] = renderedAlias{defExpr: defExpr, unique: true}
				continue
			}
			if existing.defExpr != defExpr {
				rendered[aliasName] = renderedAlias{unique: false}
			}
		}
	}
	if len(rendered) == 0 {
		return nil
	}
	out := make(map[string]string)
	for name, entry := range rendered {
		if !entry.unique || strings.TrimSpace(entry.defExpr) == "" {
			continue
		}
		out[name] = entry.defExpr
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
