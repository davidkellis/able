package compiler

import (
	"bytes"
	"fmt"
)

// renderRuntimeStructFieldHelpers emits the common named-struct lookup used at
// the interpreter/compiler boundary. The interpreter may store named fields in
// either a map or positional slots, so generated conversions must support both
// representations.
func (g *generator) renderRuntimeStructFieldHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func __able_struct_named_field_value(inst *runtime.StructInstanceValue, name string) (runtime.Value, bool) {\n")
	fmt.Fprintf(buf, "\tif inst == nil || name == \"\" {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif inst.Fields != nil {\n")
	fmt.Fprintf(buf, "\t\tif value, ok := inst.Fields[name]; ok {\n")
	fmt.Fprintf(buf, "\t\t\treturn value, true\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif inst.Positional == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.Kind != ast.StructKindNamed {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tidx, ok := inst.Definition.NamedFieldIndices[name]\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\tfor fieldIndex, field := range inst.Definition.Node.Fields {\n")
	fmt.Fprintf(buf, "\t\t\tif field != nil && field.Name != nil && field.Name.Name == name {\n")
	fmt.Fprintf(buf, "\t\t\t\tidx, ok = fieldIndex, true\n")
	fmt.Fprintf(buf, "\t\t\t\tbreak\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif !ok || idx < 0 || idx >= len(inst.Positional) {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn inst.Positional[idx], true\n")
	fmt.Fprintf(buf, "}\n\n")
}
