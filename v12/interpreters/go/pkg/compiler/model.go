package compiler

import "able/interpreter-go/pkg/ast"

type structInfo struct {
	Name      string
	GoName    string
	Kind      ast.StructKind
	Fields    []fieldInfo
	Node      *ast.StructDefinition
	Supported bool
}

type fieldInfo struct {
	Name      string
	GoName    string
	GoType    string
	Supported bool
}

type paramInfo struct {
	Name      string
	GoName    string
	GoType    string
	Supported bool
}

type functionInfo struct {
	Name           string
	GoName         string
	Params         []paramInfo
	ReturnType     string
	SupportedTypes bool
	Arity          int
	Definition     *ast.FunctionDefinition
	Compileable    bool
	Reason         string
}
