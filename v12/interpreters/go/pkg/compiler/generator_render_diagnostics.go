package compiler

import (
	"bytes"
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderDiagnosticNodeRegistration(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.diagNodes) == 0 {
		return
	}
	fmt.Fprintf(buf, "func __able_register_diag_nodes() {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbridge.ReserveNodeOrigins(__able_runtime, %d)\n", len(g.diagNodes))
	for _, info := range g.diagNodes {
		span := info.Span
		fmt.Fprintf(buf, "\tast.SetSpan(%s, ast.Span{Start: ast.Position{Line: %d, Column: %d}, End: ast.Position{Line: %d, Column: %d}})\n", info.Name, span.Start.Line, span.Start.Column, span.End.Line, span.End.Column)
		if info.Origin != "" {
			fmt.Fprintf(buf, "\tbridge.RegisterNodeOrigin(__able_runtime, %s, %q)\n", info.Name, info.Origin)
		}
		if call, ok := info.Node.(*ast.FunctionCall); ok && call != nil {
			if receiverType := g.staticCallReceiverTypeHints[call]; receiverType != nil {
				if rendered, ok := g.renderTypeExpression(receiverType); ok {
					fmt.Fprintf(buf, "\tbridge.RegisterStaticCallReceiverType(__able_runtime, %s, %s)\n", info.Name, rendered)
				}
			}
		}
	}
	fmt.Fprintf(buf, "}\n\n")
}
