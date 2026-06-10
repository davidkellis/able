package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

const typedBoundaryTelemetryEnv = "ABLE_COMPILER_TYPED_BOUNDARY_TELEMETRY"
const typedBoundaryTelemetryMaxShapes = 4096

type typedBoundaryShape struct {
	Category          string
	GeneratedFunction string
	AbleSource        string
	Carrier           string
	ImmediateConsumer string
	Reason            string
}

type typedBoundaryDiscriminatedShape struct {
	Value string
	Shape typedBoundaryShape
}

var typedBoundaryTelemetryCategories = []string{
	"any_to_runtime",
	"integer_from_runtime",
	"array_from_runtime",
	"array_to_runtime",
	"struct_from_runtime",
	"struct_to_runtime",
	"union_from_runtime",
	"union_to_runtime",
	"interface_from_runtime",
	"interface_to_runtime",
	"interface_lift_via_runtime",
	"callable_from_runtime",
	"callable_to_runtime",
	"control_from_error",
	"control_to_error",
}

func (g *generator) typedBoundaryTelemetryEnabled() bool {
	return g != nil && g.opts.EmitTypedBoundaryTelemetry
}

func typedBoundaryCounterName(category string) string {
	return "__able_typed_boundary_" + category
}

func typedBoundaryMarkerName(category string) string {
	return "__able_telemetry_typed_boundary_" + category
}

func (g *generator) renderTypedBoundaryTelemetryHelpers(buf *bytes.Buffer) {
	if !g.typedBoundaryTelemetryEnabled() {
		return
	}
	fmt.Fprintf(buf, "var __able_typed_boundary_shape_counts [%d]int64\n", typedBoundaryTelemetryMaxShapes)
	for _, category := range typedBoundaryTelemetryCategories {
		fmt.Fprintf(buf, "var %s int64\n", typedBoundaryCounterName(category))
	}
	fmt.Fprintf(buf, "\n")
	for _, category := range typedBoundaryTelemetryCategories {
		fmt.Fprintf(buf, "func %s() { atomic.AddInt64(&%s, 1) }\n", typedBoundaryMarkerName(category), typedBoundaryCounterName(category))
	}
	fmt.Fprintf(buf, "\nfunc __able_typed_boundary_telemetry_reset() {\n")
	for _, category := range typedBoundaryTelemetryCategories {
		fmt.Fprintf(buf, "\tatomic.StoreInt64(&%s, 0)\n", typedBoundaryCounterName(category))
	}
	fmt.Fprintf(buf, "\tfor idx := range __able_typed_boundary_shape_metadata {\n")
	fmt.Fprintf(buf, "\t\tatomic.StoreInt64(&__able_typed_boundary_shape_counts[idx], 0)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_typed_boundary_telemetry_snapshot() string {\n")
	fmt.Fprintf(buf, "\tcategories := fmt.Sprintf(`{")
	for idx, category := range typedBoundaryTelemetryCategories {
		if idx > 0 {
			fmt.Fprintf(buf, ",")
		}
		fmt.Fprintf(buf, "\"%s\":%%d", category)
	}
	fmt.Fprintf(buf, "}`,\n")
	for _, category := range typedBoundaryTelemetryCategories {
		fmt.Fprintf(buf, "\t\tatomic.LoadInt64(&%s),\n", typedBoundaryCounterName(category))
	}
	fmt.Fprintf(buf, "\t)\n")
	fmt.Fprintf(buf, "\tvar out strings.Builder\n")
	fmt.Fprintf(buf, "\tout.WriteString(`{\"categories\":`)\n")
	fmt.Fprintf(buf, "\tout.WriteString(categories)\n")
	fmt.Fprintf(buf, "\tout.WriteString(`,\"shapes\":[`)\n")
	fmt.Fprintf(buf, "\twrote := false\n")
	fmt.Fprintf(buf, "\tfor idx, shape := range __able_typed_boundary_shape_metadata {\n")
	fmt.Fprintf(buf, "\t\tcount := atomic.LoadInt64(&__able_typed_boundary_shape_counts[idx])\n")
	fmt.Fprintf(buf, "\t\tif count == 0 {\n")
	fmt.Fprintf(buf, "\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif wrote {\n")
	fmt.Fprintf(buf, "\t\t\tout.WriteByte(',')\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tfmt.Fprintf(&out, `{\"id\":%%d,\"category\":%%q,\"generated_function\":%%q,\"able_source\":%%q,\"carrier\":%%q,\"immediate_consumer\":%%q,\"reason\":%%q,\"count\":%%d}`, idx, shape.category, shape.generatedFunction, shape.ableSource, shape.carrier, shape.immediateConsumer, shape.reason, count)\n")
	fmt.Fprintf(buf, "\t\twrote = true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tout.WriteString(`]}`)\n")
	fmt.Fprintf(buf, "\treturn out.String()\n}\n\n")
}

func (g *generator) emitTypedBoundaryTelemetry(buf *bytes.Buffer, category string, indent string) {
	if !g.typedBoundaryTelemetryEnabled() {
		return
	}
	fmt.Fprintf(buf, "%s%s()\n", indent, typedBoundaryMarkerName(category))
}

func (g *generator) emitTypedBoundaryTelemetryShape(buf *bytes.Buffer, shape typedBoundaryShape, indent string) {
	if !g.typedBoundaryTelemetryEnabled() {
		return
	}
	g.emitTypedBoundaryTelemetry(buf, shape.Category, indent)
	idx := g.registerTypedBoundaryShape(shape)
	fmt.Fprintf(buf, "%satomic.AddInt64(&__able_typed_boundary_shape_counts[%d], 1)\n", indent, idx)
}

func (g *generator) typedBoundaryTelemetryShapeLines(shape typedBoundaryShape) []string {
	if !g.typedBoundaryTelemetryEnabled() {
		return nil
	}
	idx := g.registerTypedBoundaryShape(shape)
	return []string{
		typedBoundaryMarkerName(shape.Category) + "()",
		fmt.Sprintf("atomic.AddInt64(&__able_typed_boundary_shape_counts[%d], 1)", idx),
	}
}

func (g *generator) emitTypedBoundaryTelemetryDiscriminated(
	buf *bytes.Buffer,
	category string,
	discriminator string,
	shapes []typedBoundaryDiscriminatedShape,
	indent string,
) {
	if !g.typedBoundaryTelemetryEnabled() {
		return
	}
	g.emitTypedBoundaryTelemetry(buf, category, indent)
	fmt.Fprintf(buf, "%sswitch %s {\n", indent, discriminator)
	for _, entry := range shapes {
		idx := g.registerTypedBoundaryShape(entry.Shape)
		fmt.Fprintf(buf, "%scase %q:\n", indent, entry.Value)
		fmt.Fprintf(buf, "%s\tatomic.AddInt64(&__able_typed_boundary_shape_counts[%d], 1)\n", indent, idx)
	}
	fmt.Fprintf(buf, "%s}\n", indent)
}

func (g *generator) registerTypedBoundaryShape(shape typedBoundaryShape) int {
	if g.typedBoundaryShapeIndexes == nil {
		g.typedBoundaryShapeIndexes = make(map[typedBoundaryShape]int)
	}
	if idx, ok := g.typedBoundaryShapeIndexes[shape]; ok {
		return idx
	}
	idx := len(g.typedBoundaryShapes)
	if idx >= typedBoundaryTelemetryMaxShapes {
		panic("compiler: typed-boundary telemetry shape capacity exceeded")
	}
	// Expression lowering can register candidates while comparing speculative
	// coercion paths. Only an emitted marker call, and ultimately a nonzero
	// snapshot count, proves that a candidate reached executable generated code.
	g.typedBoundaryShapes = append(g.typedBoundaryShapes, shape)
	g.typedBoundaryShapeIndexes[shape] = idx
	return idx
}

func (g *generator) renderTypedBoundaryTelemetryMetadata(buf *bytes.Buffer) {
	if !g.typedBoundaryTelemetryEnabled() {
		return
	}
	fmt.Fprintf(buf, "type __ableTypedBoundaryShape struct {\n")
	fmt.Fprintf(buf, "\tcategory string\n\tgeneratedFunction string\n\tableSource string\n")
	fmt.Fprintf(buf, "\tcarrier string\n\timmediateConsumer string\n\treason string\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "var __able_typed_boundary_shape_metadata = [...]__ableTypedBoundaryShape{\n")
	for _, shape := range g.typedBoundaryShapes {
		fmt.Fprintf(buf, "\t{category: %q, generatedFunction: %q, ableSource: %q, carrier: %q, immediateConsumer: %q, reason: %q},\n",
			shape.Category, shape.GeneratedFunction, shape.AbleSource, shape.Carrier, shape.ImmediateConsumer, shape.Reason)
	}
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) typedBoundaryAbleSource(pkgName string, node ast.Node, fallback string) string {
	origin := ""
	if node != nil && g.nodeOrigins != nil {
		origin = strings.TrimSpace(g.nodeOrigins[node])
	}
	if origin == "" {
		origin = strings.TrimSpace(pkgName)
	}
	if origin == "" {
		origin = "<compiler-runtime>"
	}
	if node == nil {
		if fallback != "" {
			return origin + "::" + fallback
		}
		return origin
	}
	span := node.Span()
	if span.Start.Line > 0 {
		return fmt.Sprintf("%s:%d:%d", origin, span.Start.Line, span.Start.Column)
	}
	if fallback != "" {
		return origin + "::" + fallback
	}
	return origin
}

func (g *generator) typedBoundaryCurrentSource(ctx *compileContext, fallback string) string {
	if ctx == nil {
		return g.typedBoundaryAbleSource("", nil, fallback)
	}
	var node ast.Node
	if ctx.statementIndex >= 0 && ctx.statementIndex < len(ctx.blockStatements) {
		node = ctx.blockStatements[ctx.statementIndex]
	} else if ctx.function != nil {
		node = ctx.function.Definition
	}
	return g.typedBoundaryAbleSource(ctx.packageName, node, fallback)
}

func (g *generator) typedBoundaryGeneratedFunction(ctx *compileContext) string {
	if ctx != nil && ctx.function != nil {
		if ctx.function.GoName != "" {
			return ctx.function.GoName
		}
		if ctx.function.QualifiedName != "" {
			return ctx.function.QualifiedName
		}
		if ctx.function.Name != "" {
			return ctx.function.Name
		}
	}
	return "<generated-static-context>"
}

func (g *generator) typedBoundaryCarrierLabel(goType string) string {
	if g != nil {
		if expr, ok := g.typeExprForGoType(goType); ok && expr != nil {
			return fmt.Sprintf("%s [%s]", goType, typeExpressionToString(expr))
		}
	}
	return goType
}

func (g *generator) typedBoundaryCarrierFullyConcrete(ctx *compileContext, goType string) bool {
	if g == nil || goType == "" || goType == "any" || goType == "runtime.Value" {
		return false
	}
	expr, ok := g.typeExprForGoType(goType)
	if !ok || expr == nil {
		return false
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
		expr = g.lowerNormalizedTypeExpr(ctx, expr)
	}
	if info := g.structInfoByGoName(goType); info != nil && info.Package != "" {
		pkgName = info.Package
	}
	return expr != nil && g.typeExprFullyBound(pkgName, expr)
}
