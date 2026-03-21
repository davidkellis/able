package compiler

import "able/interpreter-go/pkg/ast"

type structInfo struct {
	Name        string
	Package     string
	GoName      string
	TypeExpr    ast.TypeExpression
	Kind        ast.StructKind
	Fields      []fieldInfo
	Node        *ast.StructDefinition
	Supported   bool
	Specialized bool
}

type fieldInfo struct {
	Name      string
	GoName    string
	GoType    string
	TypeExpr  ast.TypeExpression
	Supported bool
}

type paramInfo struct {
	Name         string
	GoName       string
	GoType       string
	TypeExpr     ast.TypeExpression
	Supported    bool
	OriginGoType string // underlying struct type when GoType is runtime.Value
}

type functionInfo struct {
	Name           string
	Package        string
	QualifiedName  string
	GoName         string
	Params         []paramInfo
	ParamFacts     map[string]integerFact
	ReturnType     string
	TypeBindings   map[string]ast.TypeExpression
	SupportedTypes bool
	Arity          int
	Definition     *ast.FunctionDefinition
	Compileable    bool
	Reason         string
	HasOriginal    bool
	InternalOnly   bool
}

type FallbackInfo struct {
	Name   string
	Reason string
}

type overloadInfo struct {
	Name          string
	Package       string
	QualifiedName string
	Entries       []*functionInfo
	MinArity      int
}

type methodInfo struct {
	TargetName   string
	TargetType   ast.TypeExpression
	MethodName   string
	ReceiverType string
	ExpectsSelf  bool
	Info         *functionInfo
}

type implMethodInfo struct {
	InterfaceName     string
	InterfaceArgs     []ast.TypeExpression
	InterfaceGenerics []*ast.GenericParameter
	TargetType        ast.TypeExpression
	ImplName          string
	IsDefault         bool
	ImplGenerics      []*ast.GenericParameter
	WhereClause       []*ast.WhereClauseConstraint
	MethodName        string
	Info              *functionInfo
	ImplDefinition    *ast.ImplementationDefinition
}

type implDefinitionInfo struct {
	Definition *ast.ImplementationDefinition
	Package    string
}
