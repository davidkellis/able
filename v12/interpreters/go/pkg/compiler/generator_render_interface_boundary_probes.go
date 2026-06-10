package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderNativeInterfaceKnownAdapterProbe(
	buf *bytes.Buffer,
	adapter *nativeInterfaceAdapter,
	renderedAdapterTypes map[string]struct{},
) {
	if g == nil || buf == nil || adapter == nil || adapter.TypeExpr == nil {
		return
	}
	if adapter.GoType != "" {
		renderedAdapterTypes[adapter.GoType] = struct{}{}
	}
	renderedAdapterType, ok := g.renderTypeExpression(adapter.TypeExpr)
	if !ok {
		return
	}
	fmt.Fprintf(buf, "\tif coerced, ok, err := bridge.MatchType(rt, %s, base); err != nil {\n", renderedAdapterType)
	fmt.Fprintf(buf, "\t\treturn nil, false, err\n")
	fmt.Fprintf(buf, "\t} else if ok {\n")
	if g.renderNativeInterfaceRuntimeToGoValueTryError(buf, "converted", "coerced", adapter.GoType, "\t\t") {
		fmt.Fprintf(buf, "\t\treturn %s(converted), true, nil\n", adapter.WrapHelper)
	} else {
		fmt.Fprintf(buf, "\t\t_ = coerced\n")
	}
	fmt.Fprintf(buf, "\t}\n")
}
