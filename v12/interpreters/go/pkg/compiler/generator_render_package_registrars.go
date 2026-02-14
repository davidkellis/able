package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func (g *generator) renderCompiledPackageRegistrarFiles() (map[string][]byte, error) {
	if g == nil || !g.hasFunctions() {
		return nil, nil
	}
	out := make(map[string][]byte)
	packageList := g.registrationPackageList()
	for idx, pkgName := range packageList {
		fileName := fmt.Sprintf("compiled_pkg_registrar_%s_%d.go", sanitizeIdent(strings.TrimSpace(pkgName)), idx)
		if strings.TrimSpace(pkgName) == "" {
			fileName = fmt.Sprintf("compiled_pkg_registrar_entry_%d.go", idx)
		}
		src, err := g.renderCompiledPackageRegistrarFile(pkgName, idx)
		if err != nil {
			return nil, err
		}
		out[fileName] = src
	}
	return out, nil
}

func (g *generator) renderCompiledPackageRegistrarFile(pkgName string, idx int) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	imports := []string{
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	if !g.renderRegisterPackageRegistrar(&buf, pkgName, idx) {
		return nil, fmt.Errorf("compiler: failed to render package registrar for %q", pkgName)
	}
	return formatSource(buf.Bytes())
}
