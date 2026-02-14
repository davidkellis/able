package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderCompiledPackageAggregatorsFile() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)

	imports := g.importsForCompiledPackageAggregators()
	if len(imports) > 0 {
		fmt.Fprintf(&buf, "import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		fmt.Fprintf(&buf, ")\n\n")
	}
	if !g.renderRegisterPackageRegistrars(&buf) {
		return nil, fmt.Errorf("compiler: failed to render package registrars")
	}

	return formatSource(buf.Bytes())
}
