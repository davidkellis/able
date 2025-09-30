package ast

import "math/big"

type NodeType string

const (
	NodeIdentifier               NodeType = "Identifier"
	NodeStringLiteral            NodeType = "StringLiteral"
	NodeIntegerLiteral           NodeType = "IntegerLiteral"
	NodeFloatLiteral             NodeType = "FloatLiteral"
	NodeBooleanLiteral           NodeType = "BooleanLiteral"
	NodeNilLiteral               NodeType = "NilLiteral"
	NodeCharLiteral              NodeType = "CharLiteral"
	NodeArrayLiteral             NodeType = "ArrayLiteral"
	NodeSimpleTypeExpression     NodeType = "SimpleTypeExpression"
	NodeGenericTypeExpression    NodeType = "GenericTypeExpression"
	NodeFunctionTypeExpression   NodeType = "FunctionTypeExpression"
	NodeNullableTypeExpression   NodeType = "NullableTypeExpression"
	NodeResultTypeExpression     NodeType = "ResultTypeExpression"
	NodeUnionTypeExpression      NodeType = "UnionTypeExpression"
	NodeWildcardTypeExpression   NodeType = "WildcardTypeExpression"
	NodeInterfaceConstraint      NodeType = "InterfaceConstraint"
	NodeGenericParameter         NodeType = "GenericParameter"
	NodeWhereClauseConstraint    NodeType = "WhereClauseConstraint"
	NodeWildcardPattern          NodeType = "WildcardPattern"
	NodeLiteralPattern           NodeType = "LiteralPattern"
	NodeStructPatternField       NodeType = "StructPatternField"
	NodeStructPattern            NodeType = "StructPattern"
	NodeArrayPattern             NodeType = "ArrayPattern"
	NodeTypedPattern             NodeType = "TypedPattern"
	NodeUnaryExpression          NodeType = "UnaryExpression"
	NodeBinaryExpression         NodeType = "BinaryExpression"
	NodeFunctionCall             NodeType = "FunctionCall"
	NodeBlockExpression          NodeType = "BlockExpression"
	NodeAssignmentExpression     NodeType = "AssignmentExpression"
	NodeRangeExpression          NodeType = "RangeExpression"
	NodeStringInterpolation      NodeType = "StringInterpolation"
	NodeMemberAccessExpression   NodeType = "MemberAccessExpression"
	NodeIndexExpression          NodeType = "IndexExpression"
	NodeLambdaExpression         NodeType = "LambdaExpression"
	NodeProcExpression           NodeType = "ProcExpression"
	NodeSpawnExpression          NodeType = "SpawnExpression"
	NodePropagationExpression    NodeType = "PropagationExpression"
	NodeOrElseExpression         NodeType = "OrElseExpression"
	NodeBreakpointExpression     NodeType = "BreakpointExpression"
	NodeOrClause                 NodeType = "OrClause"
	NodeIfExpression             NodeType = "IfExpression"
	NodeMatchClause              NodeType = "MatchClause"
	NodeMatchExpression          NodeType = "MatchExpression"
	NodeWhileLoop                NodeType = "WhileLoop"
	NodeForLoop                  NodeType = "ForLoop"
	NodeBreakStatement           NodeType = "BreakStatement"
	NodeContinueStatement        NodeType = "ContinueStatement"
	NodeRaiseStatement           NodeType = "RaiseStatement"
	NodeRescueExpression         NodeType = "RescueExpression"
	NodeEnsureExpression         NodeType = "EnsureExpression"
	NodeRethrowStatement         NodeType = "RethrowStatement"
	NodeStructFieldDefinition    NodeType = "StructFieldDefinition"
	NodeStructDefinition         NodeType = "StructDefinition"
	NodeStructFieldInitializer   NodeType = "StructFieldInitializer"
	NodeStructLiteral            NodeType = "StructLiteral"
	NodeUnionDefinition          NodeType = "UnionDefinition"
	NodeFunctionParameter        NodeType = "FunctionParameter"
	NodeFunctionDefinition       NodeType = "FunctionDefinition"
	NodeFunctionSignature        NodeType = "FunctionSignature"
	NodeInterfaceDefinition      NodeType = "InterfaceDefinition"
	NodeImplementationDefinition NodeType = "ImplementationDefinition"
	NodeMethodsDefinition        NodeType = "MethodsDefinition"
	NodePackageStatement         NodeType = "PackageStatement"
	NodeImportSelector           NodeType = "ImportSelector"
	NodeImportStatement          NodeType = "ImportStatement"
	NodeModule                   NodeType = "Module"
	NodeReturnStatement          NodeType = "ReturnStatement"
	NodeDynImportStatement       NodeType = "DynImportStatement"
	NodePreludeStatement         NodeType = "PreludeStatement"
	NodeExternFunctionBody       NodeType = "ExternFunctionBody"
)

type Node interface {
	NodeType() NodeType
	isNode()
}

type nodeImpl struct {
	Type NodeType `json:"type"`
}

func newNodeImpl(kind NodeType) nodeImpl {
	return nodeImpl{Type: kind}
}

func (n nodeImpl) NodeType() NodeType { return n.Type }
func (nodeImpl) isNode()              {}

// Marker interfaces.

type Expression interface {
	Node
	expressionNode()
	statementNode()
}

type expressionMarker struct{}

func (expressionMarker) expressionNode() {}

type Statement interface {
	Node
	statementNode()
}

type statementMarker struct{}

func (statementMarker) statementNode() {}

type Pattern interface {
	Node
	patternNode()
}

type patternMarker struct{}

func (patternMarker) patternNode() {}

type TypeExpression interface {
	Node
	typeExpressionNode()
}

type typeExpressionMarker struct{}

func (typeExpressionMarker) typeExpressionNode() {}

type Literal interface {
	Expression
	literalNode()
}

type literalMarker struct{}

func (literalMarker) literalNode() {}

// Identifier

type Identifier struct {
	nodeImpl
	expressionMarker
	statementMarker
	patternMarker
	assignmentTargetMarker

	Name string `json:"name"`
}

func NewIdentifier(name string) *Identifier {
	return &Identifier{nodeImpl: newNodeImpl(NodeIdentifier), Name: name}
}

// Literals

type IntegerType string

const (
	IntegerTypeI8   IntegerType = "i8"
	IntegerTypeI16  IntegerType = "i16"
	IntegerTypeI32  IntegerType = "i32"
	IntegerTypeI64  IntegerType = "i64"
	IntegerTypeI128 IntegerType = "i128"
	IntegerTypeU8   IntegerType = "u8"
	IntegerTypeU16  IntegerType = "u16"
	IntegerTypeU32  IntegerType = "u32"
	IntegerTypeU64  IntegerType = "u64"
	IntegerTypeU128 IntegerType = "u128"
)

type FloatType string

const (
	FloatTypeF32 FloatType = "f32"
	FloatTypeF64 FloatType = "f64"
)

type StringLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Value string `json:"value"`
}

func NewStringLiteral(value string) *StringLiteral {
	return &StringLiteral{nodeImpl: newNodeImpl(NodeStringLiteral), Value: value}
}

type IntegerLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Value       *big.Int     `json:"value"`
	IntegerType *IntegerType `json:"integerType,omitempty"`
}

func NewIntegerLiteral(value *big.Int, integerType *IntegerType) *IntegerLiteral {
	return &IntegerLiteral{nodeImpl: newNodeImpl(NodeIntegerLiteral), Value: value, IntegerType: integerType}
}

type FloatLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Value     float64    `json:"value"`
	FloatType *FloatType `json:"floatType,omitempty"`
}

func NewFloatLiteral(value float64, floatType *FloatType) *FloatLiteral {
	return &FloatLiteral{nodeImpl: newNodeImpl(NodeFloatLiteral), Value: value, FloatType: floatType}
}

type BooleanLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Value bool `json:"value"`
}

func NewBooleanLiteral(value bool) *BooleanLiteral {
	return &BooleanLiteral{nodeImpl: newNodeImpl(NodeBooleanLiteral), Value: value}
}

type NilLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker
}

func NewNilLiteral() *NilLiteral {
	return &NilLiteral{nodeImpl: newNodeImpl(NodeNilLiteral)}
}

type CharLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Value string `json:"value"`
}

func NewCharLiteral(value string) *CharLiteral {
	return &CharLiteral{nodeImpl: newNodeImpl(NodeCharLiteral), Value: value}
}

type ArrayLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Elements []Expression `json:"elements"`
}

func NewArrayLiteral(elements []Expression) *ArrayLiteral {
	return &ArrayLiteral{nodeImpl: newNodeImpl(NodeArrayLiteral), Elements: elements}
}

// Type expressions

type SimpleTypeExpression struct {
	nodeImpl
	typeExpressionMarker

	Name *Identifier `json:"name"`
}

func NewSimpleTypeExpression(name *Identifier) *SimpleTypeExpression {
	return &SimpleTypeExpression{nodeImpl: newNodeImpl(NodeSimpleTypeExpression), Name: name}
}

type GenericTypeExpression struct {
	nodeImpl
	typeExpressionMarker

	Base      TypeExpression   `json:"base"`
	Arguments []TypeExpression `json:"arguments"`
}

func NewGenericTypeExpression(base TypeExpression, arguments []TypeExpression) *GenericTypeExpression {
	return &GenericTypeExpression{nodeImpl: newNodeImpl(NodeGenericTypeExpression), Base: base, Arguments: arguments}
}

type FunctionTypeExpression struct {
	nodeImpl
	typeExpressionMarker

	ParamTypes []TypeExpression `json:"paramTypes"`
	ReturnType TypeExpression   `json:"returnType"`
}

func NewFunctionTypeExpression(paramTypes []TypeExpression, returnType TypeExpression) *FunctionTypeExpression {
	return &FunctionTypeExpression{nodeImpl: newNodeImpl(NodeFunctionTypeExpression), ParamTypes: paramTypes, ReturnType: returnType}
}

type NullableTypeExpression struct {
	nodeImpl
	typeExpressionMarker

	InnerType TypeExpression `json:"innerType"`
}

func NewNullableTypeExpression(inner TypeExpression) *NullableTypeExpression {
	return &NullableTypeExpression{nodeImpl: newNodeImpl(NodeNullableTypeExpression), InnerType: inner}
}

type ResultTypeExpression struct {
	nodeImpl
	typeExpressionMarker

	InnerType TypeExpression `json:"innerType"`
}

func NewResultTypeExpression(inner TypeExpression) *ResultTypeExpression {
	return &ResultTypeExpression{nodeImpl: newNodeImpl(NodeResultTypeExpression), InnerType: inner}
}

type UnionTypeExpression struct {
	nodeImpl
	typeExpressionMarker

	Members []TypeExpression `json:"members"`
}

func NewUnionTypeExpression(members []TypeExpression) *UnionTypeExpression {
	return &UnionTypeExpression{nodeImpl: newNodeImpl(NodeUnionTypeExpression), Members: members}
}

type WildcardTypeExpression struct {
	nodeImpl
	typeExpressionMarker
}

func NewWildcardTypeExpression() *WildcardTypeExpression {
	return &WildcardTypeExpression{nodeImpl: newNodeImpl(NodeWildcardTypeExpression)}
}

// Generics and constraints

type InterfaceConstraint struct {
	nodeImpl

	InterfaceType TypeExpression `json:"interfaceType"`
}

func NewInterfaceConstraint(interfaceType TypeExpression) *InterfaceConstraint {
	return &InterfaceConstraint{nodeImpl: newNodeImpl(NodeInterfaceConstraint), InterfaceType: interfaceType}
}

type GenericParameter struct {
	nodeImpl

	Name        *Identifier            `json:"name"`
	Constraints []*InterfaceConstraint `json:"constraints,omitempty"`
}

func NewGenericParameter(name *Identifier, constraints []*InterfaceConstraint) *GenericParameter {
	return &GenericParameter{nodeImpl: newNodeImpl(NodeGenericParameter), Name: name, Constraints: constraints}
}

type WhereClauseConstraint struct {
	nodeImpl

	TypeParam   *Identifier            `json:"typeParam"`
	Constraints []*InterfaceConstraint `json:"constraints"`
}

func NewWhereClauseConstraint(typeParam *Identifier, constraints []*InterfaceConstraint) *WhereClauseConstraint {
	return &WhereClauseConstraint{nodeImpl: newNodeImpl(NodeWhereClauseConstraint), TypeParam: typeParam, Constraints: constraints}
}

// Patterns

type WildcardPattern struct {
	nodeImpl
	patternMarker
	assignmentTargetMarker
}

func NewWildcardPattern() *WildcardPattern {
	return &WildcardPattern{nodeImpl: newNodeImpl(NodeWildcardPattern)}
}

type LiteralPattern struct {
	nodeImpl
	patternMarker
	assignmentTargetMarker

	Literal Literal `json:"literal"`
}

func NewLiteralPattern(literal Literal) *LiteralPattern {
	return &LiteralPattern{nodeImpl: newNodeImpl(NodeLiteralPattern), Literal: literal}
}

type StructPatternField struct {
	nodeImpl

	FieldName *Identifier `json:"fieldName,omitempty"`
	Pattern   Pattern     `json:"pattern"`
	Binding   *Identifier `json:"binding,omitempty"`
}

func NewStructPatternField(pattern Pattern, fieldName *Identifier, binding *Identifier) *StructPatternField {
	return &StructPatternField{nodeImpl: newNodeImpl(NodeStructPatternField), FieldName: fieldName, Pattern: pattern, Binding: binding}
}

type StructPattern struct {
	nodeImpl
	patternMarker
	assignmentTargetMarker

	StructType   *Identifier           `json:"structType,omitempty"`
	Fields       []*StructPatternField `json:"fields"`
	IsPositional bool                  `json:"isPositional"`
}

func NewStructPattern(fields []*StructPatternField, isPositional bool, structType *Identifier) *StructPattern {
	return &StructPattern{nodeImpl: newNodeImpl(NodeStructPattern), StructType: structType, Fields: fields, IsPositional: isPositional}
}

type ArrayPattern struct {
	nodeImpl
	patternMarker
	assignmentTargetMarker

	Elements    []Pattern `json:"elements"`
	RestPattern Pattern   `json:"restPattern,omitempty"`
}

func NewArrayPattern(elements []Pattern, rest Pattern) *ArrayPattern {
	return &ArrayPattern{nodeImpl: newNodeImpl(NodeArrayPattern), Elements: elements, RestPattern: rest}
}

type TypedPattern struct {
	nodeImpl
	patternMarker
	assignmentTargetMarker

	Pattern        Pattern        `json:"pattern"`
	TypeAnnotation TypeExpression `json:"typeAnnotation"`
}

func NewTypedPattern(pattern Pattern, typeAnnotation TypeExpression) *TypedPattern {
	return &TypedPattern{nodeImpl: newNodeImpl(NodeTypedPattern), Pattern: pattern, TypeAnnotation: typeAnnotation}
}

// Expressions

type UnaryOperator string

const (
	UnaryOperatorNegate UnaryOperator = "-"
	UnaryOperatorNot    UnaryOperator = "!"
	UnaryOperatorBitNot UnaryOperator = "~"
)

type UnaryExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Operator UnaryOperator `json:"operator"`
	Operand  Expression    `json:"operand"`
}

func NewUnaryExpression(operator UnaryOperator, operand Expression) *UnaryExpression {
	return &UnaryExpression{nodeImpl: newNodeImpl(NodeUnaryExpression), Operator: operator, Operand: operand}
}

type BinaryExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Operator string     `json:"operator"`
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
}

func NewBinaryExpression(operator string, left, right Expression) *BinaryExpression {
	return &BinaryExpression{nodeImpl: newNodeImpl(NodeBinaryExpression), Operator: operator, Left: left, Right: right}
}

type FunctionCall struct {
	nodeImpl
	expressionMarker
	statementMarker

	Callee           Expression       `json:"callee"`
	Arguments        []Expression     `json:"arguments"`
	TypeArguments    []TypeExpression `json:"typeArguments,omitempty"`
	IsTrailingLambda bool             `json:"isTrailingLambda"`
}

func NewFunctionCall(callee Expression, args []Expression, typeArgs []TypeExpression, isTrailingLambda bool) *FunctionCall {
	return &FunctionCall{nodeImpl: newNodeImpl(NodeFunctionCall), Callee: callee, Arguments: args, TypeArguments: typeArgs, IsTrailingLambda: isTrailingLambda}
}

type BlockExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Body []Statement `json:"body"`
}

func NewBlockExpression(body []Statement) *BlockExpression {
	return &BlockExpression{nodeImpl: newNodeImpl(NodeBlockExpression), Body: body}
}

type AssignmentOperator string

const (
	AssignmentDeclare AssignmentOperator = ":="
	AssignmentAssign  AssignmentOperator = "="
	AssignmentAdd     AssignmentOperator = "+="
	AssignmentSub     AssignmentOperator = "-="
	AssignmentMul     AssignmentOperator = "*="
	AssignmentDiv     AssignmentOperator = "/="
	AssignmentMod     AssignmentOperator = "%="
	AssignmentBitAnd  AssignmentOperator = "&="
	AssignmentBitOr   AssignmentOperator = "|="
	AssignmentBitXor  AssignmentOperator = "^="
	AssignmentShiftL  AssignmentOperator = "<<="
	AssignmentShiftR  AssignmentOperator = ">>="
)

type AssignmentTarget interface {
	Node
	assignmentTargetNode()
}

type assignmentTargetMarker struct{}

func (assignmentTargetMarker) assignmentTargetNode() {}

type AssignmentExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Operator AssignmentOperator `json:"operator"`
	Left     AssignmentTarget   `json:"left"`
	Right    Expression         `json:"right"`
}

func NewAssignmentExpression(operator AssignmentOperator, left AssignmentTarget, right Expression) *AssignmentExpression {
	return &AssignmentExpression{nodeImpl: newNodeImpl(NodeAssignmentExpression), Operator: operator, Left: left, Right: right}
}

type RangeExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Start     Expression `json:"start"`
	End       Expression `json:"end"`
	Inclusive bool       `json:"inclusive"`
}

func NewRangeExpression(start, end Expression, inclusive bool) *RangeExpression {
	return &RangeExpression{nodeImpl: newNodeImpl(NodeRangeExpression), Start: start, End: end, Inclusive: inclusive}
}

type StringInterpolation struct {
	nodeImpl
	expressionMarker
	statementMarker

	Parts []Expression `json:"parts"`
}

func NewStringInterpolation(parts []Expression) *StringInterpolation {
	return &StringInterpolation{nodeImpl: newNodeImpl(NodeStringInterpolation), Parts: parts}
}

type MemberAccessExpression struct {
	nodeImpl
	expressionMarker
	statementMarker
	assignmentTargetMarker

	Object Expression `json:"object"`
	Member Expression `json:"member"`
}

func NewMemberAccessExpression(object Expression, member Expression) *MemberAccessExpression {
	return &MemberAccessExpression{nodeImpl: newNodeImpl(NodeMemberAccessExpression), Object: object, Member: member}
}

type IndexExpression struct {
	nodeImpl
	expressionMarker
	statementMarker
	assignmentTargetMarker

	Object Expression `json:"object"`
	Index  Expression `json:"index"`
}

func NewIndexExpression(object, index Expression) *IndexExpression {
	return &IndexExpression{nodeImpl: newNodeImpl(NodeIndexExpression), Object: object, Index: index}
}

type LambdaExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	GenericParams   []*GenericParameter      `json:"genericParams,omitempty"`
	Params          []*FunctionParameter     `json:"params"`
	ReturnType      TypeExpression           `json:"returnType,omitempty"`
	Body            Expression               `json:"body"`
	WhereClause     []*WhereClauseConstraint `json:"whereClause,omitempty"`
	IsVerboseSyntax bool                     `json:"isVerboseSyntax"`
}

func NewLambdaExpression(params []*FunctionParameter, body Expression, returnType TypeExpression, generics []*GenericParameter, whereClause []*WhereClauseConstraint, isVerbose bool) *LambdaExpression {
	return &LambdaExpression{nodeImpl: newNodeImpl(NodeLambdaExpression), Params: params, Body: body, ReturnType: returnType, GenericParams: generics, WhereClause: whereClause, IsVerboseSyntax: isVerbose}
}

type ProcExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Expression Expression `json:"expression"`
}

func NewProcExpression(expr Expression) *ProcExpression {
	return &ProcExpression{nodeImpl: newNodeImpl(NodeProcExpression), Expression: expr}
}

type SpawnExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Expression Expression `json:"expression"`
}

func NewSpawnExpression(expr Expression) *SpawnExpression {
	return &SpawnExpression{nodeImpl: newNodeImpl(NodeSpawnExpression), Expression: expr}
}

type PropagationExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Expression Expression `json:"expression"`
}

func NewPropagationExpression(expression Expression) *PropagationExpression {
	return &PropagationExpression{nodeImpl: newNodeImpl(NodePropagationExpression), Expression: expression}
}

type OrElseExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Expression   Expression       `json:"expression"`
	Handler      *BlockExpression `json:"handler"`
	ErrorBinding *Identifier      `json:"errorBinding,omitempty"`
}

func NewOrElseExpression(expression Expression, handler *BlockExpression, errorBinding *Identifier) *OrElseExpression {
	return &OrElseExpression{nodeImpl: newNodeImpl(NodeOrElseExpression), Expression: expression, Handler: handler, ErrorBinding: errorBinding}
}

type BreakpointExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Label *Identifier      `json:"label"`
	Body  *BlockExpression `json:"body"`
}

func NewBreakpointExpression(label *Identifier, body *BlockExpression) *BreakpointExpression {
	return &BreakpointExpression{nodeImpl: newNodeImpl(NodeBreakpointExpression), Label: label, Body: body}
}

// Control flow

type OrClause struct {
	nodeImpl

	Condition Expression       `json:"condition,omitempty"`
	Body      *BlockExpression `json:"body"`
}

func NewOrClause(body *BlockExpression, condition Expression) *OrClause {
	return &OrClause{nodeImpl: newNodeImpl(NodeOrClause), Condition: condition, Body: body}
}

type IfExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	IfCondition Expression       `json:"ifCondition"`
	IfBody      *BlockExpression `json:"ifBody"`
	OrClauses   []*OrClause      `json:"orClauses"`
}

func NewIfExpression(ifCondition Expression, ifBody *BlockExpression, orClauses []*OrClause) *IfExpression {
	return &IfExpression{nodeImpl: newNodeImpl(NodeIfExpression), IfCondition: ifCondition, IfBody: ifBody, OrClauses: orClauses}
}

type MatchClause struct {
	nodeImpl

	Pattern Pattern    `json:"pattern"`
	Guard   Expression `json:"guard,omitempty"`
	Body    Expression `json:"body"`
}

func NewMatchClause(pattern Pattern, body Expression, guard Expression) *MatchClause {
	return &MatchClause{nodeImpl: newNodeImpl(NodeMatchClause), Pattern: pattern, Guard: guard, Body: body}
}

type MatchExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Subject Expression     `json:"subject"`
	Clauses []*MatchClause `json:"clauses"`
}

func NewMatchExpression(subject Expression, clauses []*MatchClause) *MatchExpression {
	return &MatchExpression{nodeImpl: newNodeImpl(NodeMatchExpression), Subject: subject, Clauses: clauses}
}

type WhileLoop struct {
	nodeImpl
	statementMarker

	Condition Expression       `json:"condition"`
	Body      *BlockExpression `json:"body"`
}

func NewWhileLoop(condition Expression, body *BlockExpression) *WhileLoop {
	return &WhileLoop{nodeImpl: newNodeImpl(NodeWhileLoop), Condition: condition, Body: body}
}

type ForLoop struct {
	nodeImpl
	statementMarker

	Pattern  Pattern          `json:"pattern"`
	Iterable Expression       `json:"iterable"`
	Body     *BlockExpression `json:"body"`
}

func NewForLoop(pattern Pattern, iterable Expression, body *BlockExpression) *ForLoop {
	return &ForLoop{nodeImpl: newNodeImpl(NodeForLoop), Pattern: pattern, Iterable: iterable, Body: body}
}

type BreakStatement struct {
	nodeImpl
	statementMarker

	Label *Identifier `json:"label"`
	Value Expression  `json:"value"`
}

func NewBreakStatement(label *Identifier, value Expression) *BreakStatement {
	return &BreakStatement{nodeImpl: newNodeImpl(NodeBreakStatement), Label: label, Value: value}
}

type ContinueStatement struct {
	nodeImpl
	statementMarker

	Label *Identifier `json:"label"`
}

func NewContinueStatement(label *Identifier) *ContinueStatement {
	return &ContinueStatement{nodeImpl: newNodeImpl(NodeContinueStatement), Label: label}
}

// Error handling

type RaiseStatement struct {
	nodeImpl
	statementMarker

	Expression Expression `json:"expression"`
}

func NewRaiseStatement(expression Expression) *RaiseStatement {
	return &RaiseStatement{nodeImpl: newNodeImpl(NodeRaiseStatement), Expression: expression}
}

type RescueExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	MonitoredExpression Expression     `json:"monitoredExpression"`
	Clauses             []*MatchClause `json:"clauses"`
}

func NewRescueExpression(monitored Expression, clauses []*MatchClause) *RescueExpression {
	return &RescueExpression{nodeImpl: newNodeImpl(NodeRescueExpression), MonitoredExpression: monitored, Clauses: clauses}
}

type EnsureExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	TryExpression Expression       `json:"tryExpression"`
	EnsureBlock   *BlockExpression `json:"ensureBlock"`
}

func NewEnsureExpression(tryExpr Expression, ensureBlock *BlockExpression) *EnsureExpression {
	return &EnsureExpression{nodeImpl: newNodeImpl(NodeEnsureExpression), TryExpression: tryExpr, EnsureBlock: ensureBlock}
}

type RethrowStatement struct {
	nodeImpl
	statementMarker
}

func NewRethrowStatement() *RethrowStatement {
	return &RethrowStatement{nodeImpl: newNodeImpl(NodeRethrowStatement)}
}

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

	StructType             *Identifier               `json:"structType,omitempty"`
	Fields                 []*StructFieldInitializer `json:"fields"`
	IsPositional           bool                      `json:"isPositional"`
	FunctionalUpdateSource Expression                `json:"functionalUpdateSource,omitempty"`
	TypeArguments          []TypeExpression          `json:"typeArguments,omitempty"`
}

func NewStructLiteral(fields []*StructFieldInitializer, isPositional bool, structType *Identifier, functionalUpdateSource Expression, typeArgs []TypeExpression) *StructLiteral {
	return &StructLiteral{nodeImpl: newNodeImpl(NodeStructLiteral), StructType: structType, Fields: fields, IsPositional: isPositional, FunctionalUpdateSource: functionalUpdateSource, TypeArguments: typeArgs}
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

	Target    HostTarget          `json:"target"`
	Signature *FunctionDefinition `json:"signature"`
	Body      string              `json:"body"`
}

func NewExternFunctionBody(target HostTarget, signature *FunctionDefinition, body string) *ExternFunctionBody {
	return &ExternFunctionBody{nodeImpl: newNodeImpl(NodeExternFunctionBody), Target: target, Signature: signature, Body: body}
}
