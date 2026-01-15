package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

// ModuleDiagnostic ties a diagnostic to the package/files that produced it.
type ModuleDiagnostic struct {
	Package    string
	Files      []string
	Diagnostic Diagnostic
	Source     SourceHint
}

// SourceHint provides a best-effort reference to the originating file.
type SourceHint struct {
	Path      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// DescribeModuleDiagnostic formats a module diagnostic for human-readable output.
func DescribeModuleDiagnostic(diag ModuleDiagnostic) string {
	message := diag.Diagnostic.Message
	if diag.Package != "" {
		message = fmt.Sprintf("%s: %s", diag.Package, message)
	}
	location := formatSourceHint(diag.Source)
	if location == "" && len(diag.Files) > 0 {
		location = fmt.Sprintf("e.g., %s", diag.Files[0])
	}
	prefix := "typechecker: "
	if diag.Diagnostic.Severity == SeverityWarning {
		prefix = "warning: typechecker: "
	}
	if location != "" {
		return fmt.Sprintf("%s%s %s", prefix, location, message)
	}
	return fmt.Sprintf("%s%s", prefix, message)
}

func formatSourceHint(hint SourceHint) string {
	path := strings.TrimSpace(hint.Path)
	line := hint.Line
	column := hint.Column
	switch {
	case path != "" && line > 0 && column > 0:
		return fmt.Sprintf("%s:%d:%d", path, line, column)
	case path != "" && line > 0:
		return fmt.Sprintf("%s:%d", path, line)
	case path != "":
		return path
	case line > 0 && column > 0:
		return fmt.Sprintf("line %d, column %d", line, column)
	case line > 0:
		return fmt.Sprintf("line %d", line)
	default:
		return ""
	}
}

// ExportedSymbolSummary summarises a binding exposed by a package.
type ExportedSymbolSummary struct {
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
}

// ExportedGenericParamSummary summarises a generic parameter and its constraints.
type ExportedGenericParamSummary struct {
	Name        string   `json:"name"`
	Constraints []string `json:"constraints,omitempty"`
}

// ExportedWhereConstraintSummary records a where-clause requirement.
type ExportedWhereConstraintSummary struct {
	TypeParam   string   `json:"typeParam"`
	Constraints []string `json:"constraints,omitempty"`
}

// ExportedObligationSummary captures solver obligations that arose while collecting exports.
type ExportedObligationSummary struct {
	Owner      string `json:"owner,omitempty"`
	TypeParam  string `json:"typeParam"`
	Constraint string `json:"constraint"`
	Subject    string `json:"subject"`
	Context    string `json:"context,omitempty"`
}

// ExportedFunctionSummary describes the callable surface of a function or method.
type ExportedFunctionSummary struct {
	Parameters  []string                         `json:"parameters,omitempty"`
	ReturnType  string                           `json:"returnType"`
	TypeParams  []ExportedGenericParamSummary    `json:"typeParams,omitempty"`
	Where       []ExportedWhereConstraintSummary `json:"where,omitempty"`
	Obligations []ExportedObligationSummary      `json:"obligations,omitempty"`
}

// ExportedStructSummary summarises a public struct definition.
type ExportedStructSummary struct {
	TypeParams []ExportedGenericParamSummary    `json:"typeParams,omitempty"`
	Fields     map[string]string                `json:"fields,omitempty"`
	Positional []string                         `json:"positional,omitempty"`
	Where      []ExportedWhereConstraintSummary `json:"where,omitempty"`
}

// ExportedInterfaceSummary summarises a public interface definition.
type ExportedInterfaceSummary struct {
	TypeParams []ExportedGenericParamSummary      `json:"typeParams,omitempty"`
	Methods    map[string]ExportedFunctionSummary `json:"methods,omitempty"`
	Where      []ExportedWhereConstraintSummary   `json:"where,omitempty"`
}

// ExportedImplementationSummary summarises a public impl block.
type ExportedImplementationSummary struct {
	ImplName      string                             `json:"implName,omitempty"`
	InterfaceName string                             `json:"interface"`
	Target        string                             `json:"target"`
	InterfaceArgs []string                           `json:"interfaceArgs,omitempty"`
	TypeParams    []ExportedGenericParamSummary      `json:"typeParams,omitempty"`
	Methods       map[string]ExportedFunctionSummary `json:"methods,omitempty"`
	Where         []ExportedWhereConstraintSummary   `json:"where,omitempty"`
	Obligations   []ExportedObligationSummary        `json:"obligations,omitempty"`
}

// ExportedMethodSetSummary summarises a public methods block.
type ExportedMethodSetSummary struct {
	TypeParams  []ExportedGenericParamSummary      `json:"typeParams,omitempty"`
	Target      string                             `json:"target"`
	Methods     map[string]ExportedFunctionSummary `json:"methods,omitempty"`
	Where       []ExportedWhereConstraintSummary   `json:"where,omitempty"`
	Obligations []ExportedObligationSummary        `json:"obligations,omitempty"`
}

// PackageSummary captures the public API surface exported by a package.
type PackageSummary struct {
	Name            string                              `json:"name"`
	Visibility      string                              `json:"visibility"`
	Symbols         map[string]ExportedSymbolSummary    `json:"symbols"`
	PrivateSymbols  map[string]ExportedSymbolSummary    `json:"privateSymbols"`
	Structs         map[string]ExportedStructSummary    `json:"structs"`
	Interfaces      map[string]ExportedInterfaceSummary `json:"interfaces"`
	Functions       map[string]ExportedFunctionSummary  `json:"functions"`
	Implementations []ExportedImplementationSummary     `json:"implementations"`
	MethodSets      []ExportedMethodSetSummary          `json:"methodSets"`
}

// CheckResult aggregates diagnostics and package summaries for a program check.
type CheckResult struct {
	Diagnostics []ModuleDiagnostic
	Packages    map[string]PackageSummary
}

type packageExports struct {
	name       string
	visibility string
	symbols    map[string]Type
	private    map[string]Type
	impls      []ImplementationSpec
	methodSets []MethodSetSpec
	structs    map[string]StructType
	interfaces map[string]InterfaceType
	functions  map[string]FunctionType
}

type aliasDeclInfo struct {
	node    ast.Node
	origins map[ast.Node]string
	path    string
}
