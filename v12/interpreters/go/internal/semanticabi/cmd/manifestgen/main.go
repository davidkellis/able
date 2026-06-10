package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type opSpec struct {
	name       string
	operands   []string
	variadic   string
	writes     []uint8
	terminator bool
}

type fieldSpec struct {
	name    string
	storage string
}

type layoutSpec struct {
	name        string
	runtimeKind string
	mutability  string
	fields      []fieldSpec
}

type hostSpec struct {
	name         string
	runtimeKind  string
	mutable      bool
	cancelable   bool
	retainsCells bool
}

func fields(values ...string) []fieldSpec {
	result := make([]fieldSpec, 0, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		result = append(result, fieldSpec{name: values[index], storage: values[index+1]})
	}
	return result
}

var objectLayouts = []layoutSpec{
	{name: "String", runtimeKind: "KindString", mutability: "LayoutImmutable", fields: fields("utf8", "FieldBytes")},
	{name: "Array", runtimeKind: "KindArray", mutability: "LayoutMutable", fields: fields("elements", "FieldCells", "element_type", "FieldScalar", "revision", "FieldScalar")},
	{name: "HashMap", runtimeKind: "KindHashMap", mutability: "LayoutMutable", fields: fields("entries", "FieldCells", "hashes", "FieldScalar", "revision", "FieldScalar")},
	{name: "Hasher", runtimeKind: "KindHasher", mutability: "LayoutMutable", fields: fields("state", "FieldScalar")},
	{name: "Function", runtimeKind: "KindFunction", mutability: "LayoutImmutable", fields: fields("declaration", "FieldScalar", "environment", "FieldObject", "method_set", "FieldObject")},
	{name: "FunctionOverload", runtimeKind: "KindFunctionOverload", mutability: "LayoutImmutable", fields: fields("functions", "FieldObjects")},
	{name: "StructDefinition", runtimeKind: "KindStructDefinition", mutability: "LayoutImmutable", fields: fields("definition", "FieldScalar", "field_names", "FieldBytes")},
	{name: "TypeRef", runtimeKind: "KindTypeRef", mutability: "LayoutImmutable", fields: fields("name", "FieldBytes", "type_arguments", "FieldScalar")},
	{name: "StructInstance", runtimeKind: "KindStructInstance", mutability: "LayoutMutable", fields: fields("definition", "FieldObject", "fields", "FieldCells", "type_arguments", "FieldScalar")},
	{name: "InterfaceDefinition", runtimeKind: "KindInterfaceDefinition", mutability: "LayoutImmutable", fields: fields("definition", "FieldScalar", "environment", "FieldObject", "qualified_name", "FieldBytes")},
	{name: "InterfaceValue", runtimeKind: "KindInterfaceValue", mutability: "LayoutImmutable", fields: fields("interface", "FieldObject", "underlying", "FieldCell", "methods", "FieldCells", "type_arguments", "FieldScalar")},
	{name: "UnionDefinition", runtimeKind: "KindUnionDefinition", mutability: "LayoutImmutable", fields: fields("definition", "FieldScalar")},
	{name: "Package", runtimeKind: "KindPackage", mutability: "LayoutMutable", fields: fields("name", "FieldBytes", "name_path", "FieldBytes", "identity", "FieldBytes", "flags", "FieldScalar", "public_names", "FieldBytes", "public_bindings", "FieldCells")},
	{name: "DynPackage", runtimeKind: "KindDynPackage", mutability: "LayoutImmutable", fields: fields("name", "FieldBytes", "name_path", "FieldBytes", "identity", "FieldBytes", "flags", "FieldScalar")},
	{name: "DynRef", runtimeKind: "KindDynRef", mutability: "LayoutImmutable", fields: fields("package", "FieldBytes", "name", "FieldBytes")},
	{name: "Error", runtimeKind: "KindError", mutability: "LayoutImmutable", fields: fields("type", "FieldScalar", "payload", "FieldCells", "message", "FieldBytes", "context", "FieldObject")},
	{name: "BoundMethod", runtimeKind: "KindBoundMethod", mutability: "LayoutImmutable", fields: fields("receiver", "FieldCell", "method", "FieldCell")},
	{name: "ImplementationNamespace", runtimeKind: "KindImplementationNamespace", mutability: "LayoutImmutable", fields: fields("definition", "FieldScalar", "methods", "FieldCells")},
	{name: "Iterator", runtimeKind: "KindIterator", mutability: "LayoutMutable", fields: fields("driver", "FieldCell", "state", "FieldObject", "retained", "FieldCells", "closed", "FieldScalar")},
	{name: "PartialFunction", runtimeKind: "KindPartialFunction", mutability: "LayoutImmutable", fields: fields("target", "FieldCell", "bound_arguments", "FieldCells", "call", "FieldScalar")},
	{name: "Environment", mutability: "LayoutMutable", fields: fields("parent", "FieldObject", "bindings", "FieldObjects")},
	{name: "BindingCell", mutability: "LayoutMutable", fields: fields("name", "FieldBytes", "value", "FieldCell")},
	{name: "SequenceStorage", mutability: "LayoutMutable", fields: fields("values", "FieldCells", "revision", "FieldScalar")},
	{name: "MapStorage", mutability: "LayoutMutable", fields: fields("entries", "FieldCells", "hashes", "FieldScalar", "revision", "FieldScalar")},
	{name: "IteratorState", mutability: "LayoutMutable", fields: fields("values", "FieldCells", "references", "FieldObjects", "position", "FieldScalar")},
	{name: "ErrorContext", mutability: "LayoutImmutable", fields: fields("values", "FieldCells", "causes", "FieldObjects", "source", "FieldBytes")},
	{name: "WideScalar", mutability: "LayoutImmutable", fields: fields("format", "FieldScalar", "payload", "FieldBytes")},
}

var hostLayouts = []hostSpec{
	{name: "NativeFunction", runtimeKind: "KindNativeFunction", retainsCells: true},
	{name: "HostHandle", runtimeKind: "KindHostHandle", mutable: true, retainsCells: true},
	{name: "NativeBoundMethod", runtimeKind: "KindNativeBoundMethod", retainsCells: true},
	{name: "Future", runtimeKind: "KindFuture", mutable: true, cancelable: true, retainsCells: true},
}

var ops = []opSpec{
	{name: "DeclareFunction", operands: []string{"OperandSymbol", "OperandImmediate"}},
	{name: "Parameter", operands: []string{"OperandSymbol"}},
	{name: "Block"},
	{name: "Constant", operands: []string{"OperandConstant"}},
	{name: "LoadName", operands: []string{"OperandSymbol"}},
	{name: "TypeRef", operands: []string{"OperandType"}},
	{name: "Assign", operands: []string{"OperandSymbol"}},
	{name: "Unary", operands: []string{"OperandSymbol"}},
	{name: "Binary", operands: []string{"OperandSymbol"}},
	{name: "Cast"},
	{name: "Call", operands: []string{"OperandImmediate"}},
	{name: "Member", operands: []string{"OperandSymbol"}},
	{name: "If", operands: []string{"OperandBlock"}},
	{name: "Loop", operands: []string{"OperandBlock"}},
	{name: "Break", operands: []string{"OperandBlock"}},
	{name: "Return"},
	{name: "Match", operands: []string{"OperandImmediate"}},
	{name: "MatchClause", operands: []string{"OperandImmediate"}},
	{name: "Pattern", operands: []string{"OperandSymbol"}},
	{name: "Raise"},
	{name: "Array", operands: []string{"OperandImmediate"}},
	{name: "Interpolate", operands: []string{"OperandImmediate"}},
	{name: "HostEffect", operands: []string{"OperandCapability", "OperandImmediate"}},
	{name: "LoadConst", operands: []string{"OperandRegister", "OperandConstant"}, writes: []uint8{0}},
	{name: "LoadGlobal", operands: []string{"OperandRegister", "OperandType", "OperandSymbol"}, writes: []uint8{0}},
	{name: "MoveValue", operands: []string{"OperandRegister", "OperandRegister"}, writes: []uint8{0}},
	{name: "UnaryValue", operands: []string{"OperandRegister", "OperandType", "OperandSymbol", "OperandRegister"}, writes: []uint8{0}},
	{name: "BinaryValue", operands: []string{"OperandRegister", "OperandType", "OperandSymbol", "OperandRegister", "OperandRegister"}, writes: []uint8{0}},
	{name: "CastValue", operands: []string{"OperandRegister", "OperandType", "OperandRegister"}, writes: []uint8{0}},
	{name: "GetMemberValue", operands: []string{"OperandRegister", "OperandType", "OperandRegister", "OperandSymbol"}, writes: []uint8{0}},
	{name: "Invoke", operands: []string{"OperandRegister", "OperandCallTarget", "OperandType"}, variadic: "OperandRegister", writes: []uint8{0}},
	{name: "TypeTest", operands: []string{"OperandRegister", "OperandRegister", "OperandType"}, writes: []uint8{0}},
	{name: "Jump", operands: []string{"OperandBlock"}, terminator: true},
	{name: "Branch", operands: []string{"OperandRegister", "OperandBlock", "OperandBlock"}, terminator: true},
	{name: "ReturnValue", operands: []string{"OperandRegister"}, terminator: true},
	{name: "RaiseValue", operands: []string{"OperandRegister"}, terminator: true},
	{name: "MatchFail", terminator: true},
	{name: "HostEffectResume", operands: []string{"OperandRegister", "OperandCapability", "OperandType", "OperandBlock"}, variadic: "OperandRegister", writes: []uint8{0}, terminator: true},
}

var immediateKinds = nameSet(
	"KindBool", "KindChar", "KindNil", "KindVoid", "KindInteger", "KindFloat", "KindIteratorEnd",
)

var hostKinds = nameSet(
	"KindNativeFunction", "KindHostHandle", "KindNativeBoundMethod", "KindFuture",
)

func main() {
	runtimePath := flag.String("runtime", "", "path to pkg/runtime/values.go")
	outPath := flag.String("out", "manifest_generated.go", "generated output path")
	check := flag.Bool("check", false, "fail if output is stale")
	flag.Parse()
	if *runtimePath == "" {
		fatalf("-runtime is required")
	}
	kinds, err := parseKinds(*runtimePath)
	if err != nil {
		fatalf("%v", err)
	}
	generated, err := render(kinds)
	if err != nil {
		fatalf("%v", err)
	}
	if *check {
		current, err := os.ReadFile(*outPath)
		if err != nil {
			fatalf("read %s: %v", *outPath, err)
		}
		if !bytes.Equal(current, generated) {
			fatalf("%s is stale; run go generate ./internal/semanticabi", *outPath)
		}
		return
	}
	if err := os.WriteFile(*outPath, generated, 0o644); err != nil {
		fatalf("write %s: %v", *outPath, err)
	}
}

func parseKinds(path string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse runtime kinds: %w", err)
	}
	var kinds []string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, rawSpec := range general.Specs {
			spec, ok := rawSpec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range spec.Names {
				if strings.HasPrefix(name.Name, "Kind") {
					kinds = append(kinds, name.Name)
				}
			}
		}
	}
	if len(kinds) == 0 || kinds[0] != "KindString" || kinds[len(kinds)-1] != "KindPartialFunction" {
		return nil, fmt.Errorf("runtime Kind declaration was not found intact")
	}
	return kinds, nil
}

func render(kinds []string) ([]byte, error) {
	var identity bytes.Buffer
	for ordinal, name := range kinds {
		fmt.Fprintf(&identity, "kind|%d|%s|%s\n", ordinal, name, classFor(name))
	}
	for index, op := range ops {
		fmt.Fprintf(&identity, "op|%d|%s|%s|%s|%v|%v\n", index+1, op.name, strings.Join(op.operands, ","), op.variadic, op.writes, op.terminator)
	}
	for index, layout := range objectLayouts {
		fmt.Fprintf(&identity, "layout|%d|%s|%s|%s", index+1, layout.name, layout.runtimeKind, layout.mutability)
		for _, field := range layout.fields {
			fmt.Fprintf(&identity, "|%s:%s", field.name, field.storage)
		}
		identity.WriteByte('\n')
	}
	for _, host := range hostLayouts {
		fmt.Fprintf(&identity, "host|%s|%s|%t|%t|%t\n", host.name, host.runtimeKind, host.mutable, host.cancelable, host.retainsCells)
	}
	digest := sha256.Sum256(identity.Bytes())

	var source bytes.Buffer
	source.WriteString("// Code generated by internal/semanticabi/cmd/manifestgen; DO NOT EDIT.\n\n")
	source.WriteString("package semanticabi\n\n")
	source.WriteString("const (\n")
	for ordinal, name := range kinds {
		fmt.Fprintf(&source, "\tTag%s uint32 = %d\n", name, ordinal+1)
	}
	source.WriteString(")\n\nconst (\n")
	for index, op := range ops {
		fmt.Fprintf(&source, "\tOp%s uint16 = %d\n", op.name, index+1)
	}
	source.WriteString(")\n\n")
	fmt.Fprintf(&source, "var ManifestIdentity = [32]byte{%s}\n\n", byteList(digest[:]))
	source.WriteString("var KindManifest = [...]KindDescriptor{\n")
	for ordinal, name := range kinds {
		fmt.Fprintf(&source, "\t{Name: %q, RuntimeOrdinal: %d, Tag: Tag%s, Class: %s},\n", name, ordinal, name, classFor(name))
	}
	source.WriteString("}\n\nvar OpManifest = [...]OpDescriptor{\n")
	for _, op := range ops {
		fmt.Fprintf(&source, "\t{Name: %q, Opcode: Op%s", op.name, op.name)
		if len(op.operands) != 0 {
			fmt.Fprintf(&source, ", Operands: []OperandKind{%s}", strings.Join(op.operands, ", "))
		}
		if op.variadic != "" {
			fmt.Fprintf(&source, ", Variadic: %s", op.variadic)
		}
		if len(op.writes) != 0 {
			writes := make([]string, len(op.writes))
			for index, value := range op.writes {
				writes[index] = fmt.Sprintf("%d", value)
			}
			fmt.Fprintf(&source, ", Writes: []uint8{%s}", strings.Join(writes, ", "))
		}
		if op.terminator {
			source.WriteString(", Terminator: true")
		}
		source.WriteString("},\n")
	}
	source.WriteString("}\n")
	source.WriteString("\nconst (\n")
	for index, layout := range objectLayouts {
		fmt.Fprintf(&source, "\tLayout%s uint16 = %d\n", layout.name, index+1)
	}
	source.WriteString(")\n\nvar ObjectLayoutManifest = [...]ObjectLayoutDescriptor{\n")
	for _, layout := range objectLayouts {
		runtimeTag := "uint32(0)"
		if layout.runtimeKind != "" {
			runtimeTag = "Tag" + layout.runtimeKind
		}
		fmt.Fprintf(&source, "\t{Name: %q, LayoutID: Layout%s, RuntimeTag: %s, Mutability: %s", layout.name, layout.name, runtimeTag, layout.mutability)
		if len(layout.fields) != 0 {
			source.WriteString(", Fields: []LayoutFieldDescriptor{")
			for _, field := range layout.fields {
				fmt.Fprintf(&source, "{Name: %q, Storage: %s},", field.name, field.storage)
			}
			source.WriteString("}")
		}
		source.WriteString("},\n")
	}
	source.WriteString("}\n\nvar HostLayoutManifest = [...]HostLayoutDescriptor{\n")
	for _, host := range hostLayouts {
		fmt.Fprintf(&source, "\t{Name: %q, RuntimeTag: Tag%s, Mutable: %t, Cancelable: %t, RetainsCells: %t},\n", host.name, host.runtimeKind, host.mutable, host.cancelable, host.retainsCells)
	}
	source.WriteString("}\n")
	return format.Source(source.Bytes())
}

func classFor(name string) string {
	if immediateKinds[name] {
		return "KindImmediate"
	}
	if hostKinds[name] {
		return "KindHostRegistry"
	}
	return "KindSharedHeap"
}

func nameSet(names ...string) map[string]bool {
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name] = true
	}
	return result
}

func byteList(values []byte) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = fmt.Sprintf("0x%02x", value)
	}
	return strings.Join(parts, ", ")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "manifestgen: "+format+"\n", args...)
	os.Exit(1)
}
