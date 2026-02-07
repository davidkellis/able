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
	TypeExpr  ast.TypeExpression
	Supported bool
}

type functionInfo struct {
	Name           string
	Package        string
	QualifiedName  string
	GoName         string
	Params         []paramInfo
	ReturnType     string
	SupportedTypes bool
	Arity          int
	Definition     *ast.FunctionDefinition
	Compileable    bool
	Reason         string
}

type FallbackInfo struct {
	Name   string
	Reason string
}

type overloadInfo struct {
	Name     string
	Package  string
	QualifiedName string
	Entries  []*functionInfo
	MinArity int
}

type methodInfo struct {
	TargetName   string
	TargetType   ast.TypeExpression
	MethodName   string
	ReceiverType string
	ExpectsSelf  bool
	Info         *functionInfo
}
