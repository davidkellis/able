package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderCompiledRegisterFile() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)

	imports := g.importsForCompiledRegister()
	if len(imports) > 0 {
		fmt.Fprintf(&buf, "import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		fmt.Fprintf(&buf, ")\n\n")
	}
	fmt.Fprintf(&buf, "var _ = ast.NewIdentifier\n\n")

	g.renderRegister(&buf)
	return formatSource(buf.Bytes())
}
