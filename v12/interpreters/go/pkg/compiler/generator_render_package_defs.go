package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderCompiledPackageDefinitionFiles() (map[string][]byte, error) {
	if g == nil || !g.hasFunctions() {
		return nil, nil
	}
	out := make(map[string][]byte)
	packageList := g.registrationPackageList()
	for idx, pkgName := range packageList {
		fileName := fmt.Sprintf("compiled_pkg_defs_%s_%d.go", sanitizeIdent(strings.TrimSpace(pkgName)), idx)
		if strings.TrimSpace(pkgName) == "" {
			fileName = fmt.Sprintf("compiled_pkg_defs_entry_%d.go", idx)
		}
		src, err := g.renderCompiledPackageDefinitionFile(pkgName, idx)
		if err != nil {
			return nil, err
		}
		out[fileName] = src
	}
	return out, nil
}

func (g *generator) renderCompiledPackageDefinitionFile(pkgName string, idx int) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	imports := []string{
		"able/interpreter-go/pkg/ast",
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
		"fmt",
		"math/big",
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "var (\n")
	fmt.Fprintf(&buf, "\t_ = ast.NewIdentifier\n")
	fmt.Fprintf(&buf, "\t_ = big.NewInt\n")
	fmt.Fprintf(&buf, "\t_ = bridge.New\n")
	fmt.Fprintf(&buf, "\t_ = fmt.Errorf\n")
	fmt.Fprintf(&buf, "\t_ runtime.Value\n")
	fmt.Fprintf(&buf, ")\n\n")
	fnName := g.packageDefinitionRegistrarFuncName(pkgName, idx)
	fmt.Fprintf(&buf, "func %s(rt *bridge.Runtime, interp *interpreter.Interpreter, pkgEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n", fnName)
	fmt.Fprintf(&buf, "\t_ = rt\n")
	fmt.Fprintf(&buf, "\t_ = interp\n")
	fmt.Fprintf(&buf, "\t_ = pkgEnv\n")
	fmt.Fprintf(&buf, "\t_ = __able_bootstrapped_metadata\n")
	for _, info := range g.sortedStructInfosForPackage(pkgName) {
		defExpr, ok := g.renderStructDefinitionExpr(info)
		if !ok {
			continue
		}
		localDefVar := sanitizeIdent(info.Name) + "_def"
		fmt.Fprintf(&buf, "\tif existingDef, ok := pkgEnv.StructDefinition(%q); ok {\n", info.Name)
		fmt.Fprintf(&buf, "\t\tif _, err := pkgEnv.Get(%q); err != nil {\n", info.Name)
		fmt.Fprintf(&buf, "\t\t\tpkgEnv.Define(%q, existingDef)\n", info.Name)
		fmt.Fprintf(&buf, "\t\t}\n")
		fmt.Fprintf(&buf, "\t\tinterp.RegisterPackageSymbol(%q, %q, existingDef)\n", pkgName, info.Name)
		fmt.Fprintf(&buf, "\t} else {\n")
		fmt.Fprintf(&buf, "\t\t%s := %s\n", localDefVar, defExpr)
		fmt.Fprintf(&buf, "\t\tpkgEnv.Define(%q, %s)\n", info.Name, localDefVar)
		fmt.Fprintf(&buf, "\t\tpkgEnv.DefineStruct(%q, %s)\n", info.Name, localDefVar)
		fmt.Fprintf(&buf, "\t\tinterp.RegisterPackageSymbol(%q, %q, %s)\n", pkgName, info.Name, localDefVar)
		fmt.Fprintf(&buf, "\t}\n")
	}
	for _, def := range g.sortedInterfaceDefsForPackage(pkgName) {
		defExpr, ok := g.renderInterfaceDefinitionExpr(def, "pkgEnv")
		if !ok || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		localVar := fmt.Sprintf("__able_iface_%s", sanitizeIdent(def.ID.Name))
		fmt.Fprintf(&buf, "\t%s := %s\n", localVar, defExpr)
		fmt.Fprintf(&buf, "\tinterp.RegisterInterfaceDefinition(%q, %s)\n", def.ID.Name, localVar)
		fmt.Fprintf(&buf, "\tif _, err := pkgEnv.Get(%q); err != nil {\n", def.ID.Name)
		fmt.Fprintf(&buf, "\t\tpkgEnv.Define(%q, %s)\n", def.ID.Name, localVar)
		fmt.Fprintf(&buf, "\t}\n")
	}
	for _, def := range g.sortedUnionDefsForPackage(pkgName) {
		defExpr, ok := g.renderUnionDefinitionExpr(def)
		if !ok || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		localVar := fmt.Sprintf("__able_union_%s", sanitizeIdent(def.ID.Name))
		fmt.Fprintf(&buf, "\t%s := %s\n", localVar, defExpr)
		fmt.Fprintf(&buf, "\tinterp.RegisterUnionDefinition(%q, &%s)\n", def.ID.Name, localVar)
		fmt.Fprintf(&buf, "\tinterp.RegisterPackageSymbol(%q, %q, %s)\n", pkgName, def.ID.Name, localVar)
		fmt.Fprintf(&buf, "\tif _, err := pkgEnv.Get(%q); err != nil {\n", def.ID.Name)
		fmt.Fprintf(&buf, "\t\tpkgEnv.Define(%q, %s)\n", def.ID.Name, localVar)
		fmt.Fprintf(&buf, "\t}\n")
		// Seed union variant constructors: make each variant struct available by name
		variantTypes := collectUnionVariantTypes(def.Variants)
		for _, variantName := range variantTypes {
			if variantName == def.ID.Name {
				continue
			}
			fmt.Fprintf(&buf, "\tif _, err := pkgEnv.Get(%q); err != nil {\n", variantName)
			fmt.Fprintf(&buf, "\t\tif variantDef, ok := pkgEnv.StructDefinition(%q); ok {\n", variantName)
			fmt.Fprintf(&buf, "\t\t\tpkgEnv.Define(%q, variantDef)\n", variantName)
			fmt.Fprintf(&buf, "\t\t}\n")
			fmt.Fprintf(&buf, "\t}\n")
		}
	}
	// Register type aliases so bridge.MatchType can expand them in no-bootstrap mode.
	if aliases, ok := g.typeAliases[pkgName]; ok {
		aliasNames := make([]string, 0, len(aliases))
		for name := range aliases {
			aliasNames = append(aliasNames, name)
		}
		sort.Strings(aliasNames)
		for _, aliasName := range aliasNames {
			target := aliases[aliasName]
			targetExpr, ok := g.renderTypeExpression(target)
			if !ok || targetExpr == "" {
				continue
			}
			// Find the original alias AST node to get generic params
			genericParamsExpr := g.renderTypeAliasGenericParams(pkgName, aliasName)
			if genericParamsExpr != "" {
				fmt.Fprintf(&buf, "\tinterp.RegisterTypeAlias(%q, &ast.TypeAliasDefinition{ID: ast.NewIdentifier(%q), TargetType: %s, GenericParams: %s})\n", aliasName, aliasName, targetExpr, genericParamsExpr)
			} else {
				fmt.Fprintf(&buf, "\tinterp.RegisterTypeAlias(%q, &ast.TypeAliasDefinition{ID: ast.NewIdentifier(%q), TargetType: %s})\n", aliasName, aliasName, targetExpr)
			}
		}
	}
	g.renderNamedImplNamespaceSeeds(&buf, "pkgEnv", pkgName)
	g.renderModuleBindingSeeds(&buf, "pkgEnv", pkgName)
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n")
	return formatSource(buf.Bytes())
}

func collectUnionVariantTypes(variants []ast.TypeExpression) []string {
	var names []string
	for _, v := range variants {
		switch vt := v.(type) {
		case *ast.SimpleTypeExpression:
			if vt != nil && vt.Name != nil && vt.Name.Name != "" {
				names = append(names, vt.Name.Name)
			}
		case *ast.GenericTypeExpression:
			if vt != nil {
				if base, ok := vt.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && base.Name.Name != "" {
					names = append(names, base.Name.Name)
				}
			}
		case *ast.UnionTypeExpression:
			if vt != nil {
				names = append(names, collectUnionVariantTypes(vt.Members)...)
			}
		}
	}
	return names
}

func (g *generator) renderTypeAliasGenericParams(pkgName string, aliasName string) string {
	if g == nil || aliasName == "" {
		return ""
	}
	params := g.typeAliasGenericParams[pkgName][aliasName]
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, gp := range params {
		if gp == nil || gp.Name == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("ast.NewGenericParameter(ast.NewIdentifier(%q), nil)", gp.Name.Name))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("[]*ast.GenericParameter{%s}", strings.Join(parts, ", "))
}

func (g *generator) renderModuleBindingSeeds(buf *bytes.Buffer, envVar string, pkgName string) {
	if g == nil || buf == nil {
		return
	}
	bindings := g.moduleBindings[pkgName]
	if len(bindings) == 0 {
		return
	}
	fmt.Fprintf(buf, "\t// Module-level constant bindings\n")
	for _, binding := range bindings {
		fmt.Fprintf(buf, "\tif _, err := %s.Get(%q); err != nil {\n", envVar, binding.Name)
		fmt.Fprintf(buf, "\t\t%s.Define(%q, %s)\n", envVar, binding.Name, binding.GoValue)
		fmt.Fprintf(buf, "\t}\n")
	}
}
