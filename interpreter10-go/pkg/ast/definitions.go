package ast

// Definitions

type StructFieldDefinition struct {
	nodeImpl

	Name      *Identifier    `json:"name,omitempty"`
	FieldType TypeExpression `json:"fieldType"`
}

func NewStructFieldDefinition(fieldType TypeExpression, name *Identifier) *StructFieldDefinition {
	return &StructFieldDefinition{nodeImpl: newNodeImpl(NodeStructFieldDefinition), Name: name, FieldType: fieldType}
}

type StructKind string

const (
	StructKindSingleton  StructKind = "singleton"
	StructKindNamed      StructKind = "named"
	StructKindPositional StructKind = "positional"
)

type StructDefinition struct {
	nodeImpl
	statementMarker

	ID            *Identifier              `json:"id"`
	GenericParams []*GenericParameter      `json:"genericParams,omitempty"`
	Fields        []*StructFieldDefinition `json:"fields"`
	WhereClause   []*WhereClauseConstraint `json:"whereClause,omitempty"`
	Kind          StructKind               `json:"kind"`
	IsPrivate     bool                     `json:"isPrivate,omitempty"`
}

func NewStructDefinition(id *Identifier, fields []*StructFieldDefinition, kind StructKind, genericParams []*GenericParameter, whereClause []*WhereClauseConstraint, isPrivate bool) *StructDefinition {
	return &StructDefinition{nodeImpl: newNodeImpl(NodeStructDefinition), ID: id, Fields: fields, Kind: kind, GenericParams: genericParams, WhereClause: whereClause, IsPrivate: isPrivate}
}

type StructFieldInitializer struct {
	nodeImpl

	Name        *Identifier `json:"name,omitempty"`
	Value       Expression  `json:"value"`
	IsShorthand bool        `json:"isShorthand"`
}

func NewStructFieldInitializer(value Expression, name *Identifier, isShorthand bool) *StructFieldInitializer {
	return &StructFieldInitializer{nodeImpl: newNodeImpl(NodeStructFieldInitializer), Name: name, Value: value, IsShorthand: isShorthand}
}

type StructLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker

	StructType              *Identifier               `json:"structType,omitempty"`
	Fields                  []*StructFieldInitializer `json:"fields"`
	IsPositional            bool                      `json:"isPositional"`
	FunctionalUpdateSources []Expression              `json:"functionalUpdateSources,omitempty"`
	TypeArguments           []TypeExpression          `json:"typeArguments,omitempty"`
}

func NewStructLiteral(fields []*StructFieldInitializer, isPositional bool, structType *Identifier, functionalUpdateSources []Expression, typeArgs []TypeExpression) *StructLiteral {
	return &StructLiteral{nodeImpl: newNodeImpl(NodeStructLiteral), StructType: structType, Fields: fields, IsPositional: isPositional, FunctionalUpdateSources: functionalUpdateSources, TypeArguments: typeArgs}
}

type UnionDefinition struct {
	nodeImpl
	statementMarker

	ID            *Identifier              `json:"id"`
	GenericParams []*GenericParameter      `json:"genericParams,omitempty"`
	Variants      []TypeExpression         `json:"variants"`
	WhereClause   []*WhereClauseConstraint `json:"whereClause,omitempty"`
	IsPrivate     bool                     `json:"isPrivate,omitempty"`
}

func NewUnionDefinition(id *Identifier, variants []TypeExpression, genericParams []*GenericParameter, whereClause []*WhereClauseConstraint, isPrivate bool) *UnionDefinition {
	return &UnionDefinition{nodeImpl: newNodeImpl(NodeUnionDefinition), ID: id, Variants: variants, GenericParams: genericParams, WhereClause: whereClause, IsPrivate: isPrivate}
}

type FunctionParameter struct {
	nodeImpl

	Name      Pattern        `json:"name"`
	ParamType TypeExpression `json:"paramType,omitempty"`
}

func NewFunctionParameter(name Pattern, paramType TypeExpression) *FunctionParameter {
	return &FunctionParameter{nodeImpl: newNodeImpl(NodeFunctionParameter), Name: name, ParamType: paramType}
}

type FunctionDefinition struct {
	nodeImpl
	statementMarker

	ID                *Identifier              `json:"id"`
	GenericParams     []*GenericParameter      `json:"genericParams,omitempty"`
	Params            []*FunctionParameter     `json:"params"`
	ReturnType        TypeExpression           `json:"returnType,omitempty"`
	Body              *BlockExpression         `json:"body"`
	WhereClause       []*WhereClauseConstraint `json:"whereClause,omitempty"`
	IsMethodShorthand bool                     `json:"isMethodShorthand"`
	IsPrivate         bool                     `json:"isPrivate"`
}

func NewFunctionDefinition(id *Identifier, params []*FunctionParameter, body *BlockExpression, returnType TypeExpression, generics []*GenericParameter, whereClause []*WhereClauseConstraint, isMethodShorthand, isPrivate bool) *FunctionDefinition {
	return &FunctionDefinition{nodeImpl: newNodeImpl(NodeFunctionDefinition), ID: id, Params: params, Body: body, ReturnType: returnType, GenericParams: generics, WhereClause: whereClause, IsMethodShorthand: isMethodShorthand, IsPrivate: isPrivate}
}

type FunctionSignature struct {
	nodeImpl

	Name          *Identifier              `json:"name"`
	GenericParams []*GenericParameter      `json:"genericParams,omitempty"`
	Params        []*FunctionParameter     `json:"params"`
	ReturnType    TypeExpression           `json:"returnType,omitempty"`
	WhereClause   []*WhereClauseConstraint `json:"whereClause,omitempty"`
	DefaultImpl   *BlockExpression         `json:"defaultImpl,omitempty"`
}

func NewFunctionSignature(name *Identifier, params []*FunctionParameter, returnType TypeExpression, generics []*GenericParameter, whereClause []*WhereClauseConstraint, defaultImpl *BlockExpression) *FunctionSignature {
	return &FunctionSignature{nodeImpl: newNodeImpl(NodeFunctionSignature), Name: name, Params: params, ReturnType: returnType, GenericParams: generics, WhereClause: whereClause, DefaultImpl: defaultImpl}
}

type InterfaceDefinition struct {
	nodeImpl
	statementMarker

	ID              *Identifier              `json:"id"`
	GenericParams   []*GenericParameter      `json:"genericParams,omitempty"`
	SelfTypePattern TypeExpression           `json:"selfTypePattern,omitempty"`
	Signatures      []*FunctionSignature     `json:"signatures"`
	WhereClause     []*WhereClauseConstraint `json:"whereClause,omitempty"`
	BaseInterfaces  []TypeExpression         `json:"baseInterfaces,omitempty"`
	IsPrivate       bool                     `json:"isPrivate"`
}

func NewInterfaceDefinition(id *Identifier, signatures []*FunctionSignature, generics []*GenericParameter, selfTypePattern TypeExpression, whereClause []*WhereClauseConstraint, baseInterfaces []TypeExpression, isPrivate bool) *InterfaceDefinition {
	return &InterfaceDefinition{nodeImpl: newNodeImpl(NodeInterfaceDefinition), ID: id, Signatures: signatures, GenericParams: generics, SelfTypePattern: selfTypePattern, WhereClause: whereClause, BaseInterfaces: baseInterfaces, IsPrivate: isPrivate}
}

type ImplementationDefinition struct {
	nodeImpl
	statementMarker

	ImplName      *Identifier              `json:"implName,omitempty"`
	GenericParams []*GenericParameter      `json:"genericParams,omitempty"`
	InterfaceName *Identifier              `json:"interfaceName"`
	InterfaceArgs []TypeExpression         `json:"interfaceArgs,omitempty"`
	TargetType    TypeExpression           `json:"targetType"`
	Definitions   []*FunctionDefinition    `json:"definitions"`
	WhereClause   []*WhereClauseConstraint `json:"whereClause,omitempty"`
	IsPrivate     bool                     `json:"isPrivate,omitempty"`
}

func NewImplementationDefinition(interfaceName *Identifier, targetType TypeExpression, definitions []*FunctionDefinition, implName *Identifier, generics []*GenericParameter, interfaceArgs []TypeExpression, whereClause []*WhereClauseConstraint, isPrivate bool) *ImplementationDefinition {
	return &ImplementationDefinition{nodeImpl: newNodeImpl(NodeImplementationDefinition), InterfaceName: interfaceName, TargetType: targetType, Definitions: definitions, ImplName: implName, GenericParams: generics, InterfaceArgs: interfaceArgs, WhereClause: whereClause, IsPrivate: isPrivate}
}

type MethodsDefinition struct {
	nodeImpl
	statementMarker

	TargetType    TypeExpression           `json:"targetType"`
	GenericParams []*GenericParameter      `json:"genericParams,omitempty"`
	Definitions   []*FunctionDefinition    `json:"definitions"`
	WhereClause   []*WhereClauseConstraint `json:"whereClause,omitempty"`
}

func NewMethodsDefinition(targetType TypeExpression, definitions []*FunctionDefinition, generics []*GenericParameter, whereClause []*WhereClauseConstraint) *MethodsDefinition {
	return &MethodsDefinition{nodeImpl: newNodeImpl(NodeMethodsDefinition), TargetType: targetType, Definitions: definitions, GenericParams: generics, WhereClause: whereClause}
}

// Packages & imports

type PackageStatement struct {
	nodeImpl
	statementMarker

	NamePath  []*Identifier `json:"namePath"`
	IsPrivate bool          `json:"isPrivate,omitempty"`
}

func NewPackageStatement(namePath []*Identifier, isPrivate bool) *PackageStatement {
	return &PackageStatement{nodeImpl: newNodeImpl(NodePackageStatement), NamePath: namePath, IsPrivate: isPrivate}
}

type ImportSelector struct {
	nodeImpl

	Name  *Identifier `json:"name"`
	Alias *Identifier `json:"alias,omitempty"`
}

func NewImportSelector(name *Identifier, alias *Identifier) *ImportSelector {
	return &ImportSelector{nodeImpl: newNodeImpl(NodeImportSelector), Name: name, Alias: alias}
}

type ImportStatement struct {
	nodeImpl
	statementMarker

	PackagePath []*Identifier     `json:"packagePath"`
	IsWildcard  bool              `json:"isWildcard"`
	Selectors   []*ImportSelector `json:"selectors,omitempty"`
	Alias       *Identifier       `json:"alias,omitempty"`
}

func NewImportStatement(packagePath []*Identifier, isWildcard bool, selectors []*ImportSelector, alias *Identifier) *ImportStatement {
	return &ImportStatement{nodeImpl: newNodeImpl(NodeImportStatement), PackagePath: packagePath, IsWildcard: isWildcard, Selectors: selectors, Alias: alias}
}

// Module root

type Module struct {
	nodeImpl

	Package *PackageStatement  `json:"package,omitempty"`
	Imports []*ImportStatement `json:"imports"`
	Body    []Statement        `json:"body"`
}

func NewModule(body []Statement, imports []*ImportStatement, pkg *PackageStatement) *Module {
	return &Module{nodeImpl: newNodeImpl(NodeModule), Package: pkg, Imports: imports, Body: body}
}

// Statements

type ReturnStatement struct {
	nodeImpl
	statementMarker

	Argument Expression `json:"argument,omitempty"`
}

func NewReturnStatement(argument Expression) *ReturnStatement {
	return &ReturnStatement{nodeImpl: newNodeImpl(NodeReturnStatement), Argument: argument}
}

type DynImportStatement struct {
	nodeImpl
	statementMarker

	PackagePath []*Identifier     `json:"packagePath"`
	IsWildcard  bool              `json:"isWildcard"`
	Selectors   []*ImportSelector `json:"selectors,omitempty"`
	Alias       *Identifier       `json:"alias,omitempty"`
}

func NewDynImportStatement(packagePath []*Identifier, isWildcard bool, selectors []*ImportSelector, alias *Identifier) *DynImportStatement {
	return &DynImportStatement{nodeImpl: newNodeImpl(NodeDynImportStatement), PackagePath: packagePath, IsWildcard: isWildcard, Selectors: selectors, Alias: alias}
}

// Host interop

type HostTarget string

const (
	HostTargetGo         HostTarget = "go"
	HostTargetCrystal    HostTarget = "crystal"
	HostTargetTypeScript HostTarget = "typescript"
	HostTargetPython     HostTarget = "python"
	HostTargetRuby       HostTarget = "ruby"
)

type PreludeStatement struct {
	nodeImpl
	statementMarker

	Target HostTarget `json:"target"`
	Code   string     `json:"code"`
}

func NewPreludeStatement(target HostTarget, code string) *PreludeStatement {
	return &PreludeStatement{nodeImpl: newNodeImpl(NodePreludeStatement), Target: target, Code: code}
}

type ExternFunctionBody struct {
	nodeImpl
	statementMarker

	Target    HostTarget          `json:"target"`
	Signature *FunctionDefinition `json:"signature"`
	Body      string              `json:"body"`
}

func NewExternFunctionBody(target HostTarget, signature *FunctionDefinition, body string) *ExternFunctionBody {
	return &ExternFunctionBody{nodeImpl: newNodeImpl(NodeExternFunctionBody), Target: target, Signature: signature, Body: body}
}
