package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
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
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "var (\n")
	fmt.Fprintf(&buf, "\t_ = ast.NewIdentifier\n")
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
		fmt.Fprintf(&buf, "\tif _, ok := pkgEnv.StructDefinition(%q); !ok {\n", info.Name)
		fmt.Fprintf(&buf, "\t\tpkgEnv.DefineStruct(%q, %s)\n", info.Name, defExpr)
		fmt.Fprintf(&buf, "\t}\n")
	}
	for _, def := range g.sortedInterfaceDefsForPackage(pkgName) {
		defExpr, ok := g.renderInterfaceDefinitionExpr(def, "pkgEnv")
		if !ok || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		fmt.Fprintf(&buf, "\tif _, err := pkgEnv.Get(%q); err != nil {\n", def.ID.Name)
		fmt.Fprintf(&buf, "\t\tpkgEnv.Define(%q, %s)\n", def.ID.Name, defExpr)
		fmt.Fprintf(&buf, "\t}\n")
	}
	for _, def := range g.sortedUnionDefsForPackage(pkgName) {
		defExpr, ok := g.renderUnionDefinitionExpr(def)
		if !ok || def == nil || def.ID == nil || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		fmt.Fprintf(&buf, "\tif _, err := pkgEnv.Get(%q); err != nil {\n", def.ID.Name)
		fmt.Fprintf(&buf, "\t\tpkgEnv.Define(%q, %s)\n", def.ID.Name, defExpr)
		fmt.Fprintf(&buf, "\t}\n")
	}
	g.renderNamedImplNamespaceSeeds(&buf, "pkgEnv", pkgName)
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n")
	return formatSource(buf.Bytes())
}
