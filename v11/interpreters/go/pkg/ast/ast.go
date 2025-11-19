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
	NodeMapLiteralEntry          NodeType = "MapLiteralEntry"
	NodeMapLiteralSpread         NodeType = "MapLiteralSpread"
	NodeMapLiteral               NodeType = "MapLiteral"
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
	NodeAwaitExpression          NodeType = "AwaitExpression"
	NodeOrElseExpression         NodeType = "OrElseExpression"
	NodeBreakpointExpression     NodeType = "BreakpointExpression"
	NodeOrClause                 NodeType = "OrClause"
	NodeIfExpression             NodeType = "IfExpression"
	NodeMatchClause              NodeType = "MatchClause"
	NodeMatchExpression          NodeType = "MatchExpression"
	NodeWhileLoop                NodeType = "WhileLoop"
	NodeForLoop                  NodeType = "ForLoop"
	NodeLoopExpression           NodeType = "LoopExpression"
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
	NodeTypeAliasDefinition      NodeType = "TypeAliasDefinition"
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
	NodeIteratorLiteral          NodeType = "IteratorLiteral"
	NodeImplicitMemberExpression NodeType = "ImplicitMemberExpression"
	NodePlaceholderExpression    NodeType = "PlaceholderExpression"
	NodeTopicReferenceExpression NodeType = "TopicReferenceExpression"
	NodeYieldStatement           NodeType = "YieldStatement"
	NodePreludeStatement         NodeType = "PreludeStatement"
	NodeExternFunctionBody       NodeType = "ExternFunctionBody"
)

type Node interface {
	NodeType() NodeType
	Span() Span
	isNode()
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Span struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type nodeImpl struct {
	Type NodeType `json:"type"`
	span Span
}

func newNodeImpl(kind NodeType) nodeImpl {
	return nodeImpl{Type: kind}
}

func (n nodeImpl) NodeType() NodeType { return n.Type }
func (n nodeImpl) Span() Span         { return n.span }
func (nodeImpl) isNode()              {}
func (n *nodeImpl) setSpan(span Span) { n.span = span }

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

	Value any `json:"value"`
}

func NewNilLiteral() *NilLiteral {
	return &NilLiteral{nodeImpl: newNodeImpl(NodeNilLiteral), Value: nil}
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

type MapLiteralElement interface {
	Node
	mapLiteralElementNode()
}

type mapLiteralElementMarker struct{}

func (mapLiteralElementMarker) mapLiteralElementNode() {}

type MapLiteralEntry struct {
	nodeImpl
	mapLiteralElementMarker

	Key   Expression `json:"key"`
	Value Expression `json:"value"`
}

func NewMapLiteralEntry(key, value Expression) *MapLiteralEntry {
	return &MapLiteralEntry{nodeImpl: newNodeImpl(NodeMapLiteralEntry), Key: key, Value: value}
}

type MapLiteralSpread struct {
	nodeImpl
	mapLiteralElementMarker

	Expression Expression `json:"expression"`
}

func NewMapLiteralSpread(expr Expression) *MapLiteralSpread {
	return &MapLiteralSpread{nodeImpl: newNodeImpl(NodeMapLiteralSpread), Expression: expr}
}

type MapLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker
	literalMarker

	Elements []MapLiteralElement `json:"elements"`
}

func NewMapLiteral(elements []MapLiteralElement) *MapLiteral {
	return &MapLiteral{nodeImpl: newNodeImpl(NodeMapLiteral), Elements: elements}
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
	IsInferred  bool                   `json:"isInferred,omitempty"`
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
	if args == nil {
		args = make([]Expression, 0)
	}
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

type IteratorLiteral struct {
	nodeImpl
	expressionMarker
	statementMarker

	Body        []Statement    `json:"body"`
	Binding     *Identifier    `json:"binding,omitempty"`
	ElementType TypeExpression `json:"elementType,omitempty"`
}

func NewIteratorLiteral(body []Statement) *IteratorLiteral {
	return &IteratorLiteral{nodeImpl: newNodeImpl(NodeIteratorLiteral), Body: body}
}

type ImplicitMemberExpression struct {
	nodeImpl
	expressionMarker
	statementMarker
	assignmentTargetMarker

	Member *Identifier `json:"member"`
}

func NewImplicitMemberExpression(member *Identifier) *ImplicitMemberExpression {
	return &ImplicitMemberExpression{nodeImpl: newNodeImpl(NodeImplicitMemberExpression), Member: member}
}

type PlaceholderExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Index *int `json:"index,omitempty"`
}

func NewPlaceholderExpression(index *int) *PlaceholderExpression {
	return &PlaceholderExpression{nodeImpl: newNodeImpl(NodePlaceholderExpression), Index: index}
}

type TopicReferenceExpression struct {
	nodeImpl
	expressionMarker
	statementMarker
}

func NewTopicReferenceExpression() *TopicReferenceExpression {
	return &TopicReferenceExpression{nodeImpl: newNodeImpl(NodeTopicReferenceExpression)}
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
	AssignmentBitXor  AssignmentOperator = `\xor=`
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
	Safe   bool       `json:"safe,omitempty"`
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

type AwaitExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Expression Expression `json:"expression"`
}

func NewAwaitExpression(expr Expression) *AwaitExpression {
	return &AwaitExpression{nodeImpl: newNodeImpl(NodeAwaitExpression), Expression: expr}
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

type LoopExpression struct {
	nodeImpl
	expressionMarker
	statementMarker

	Body *BlockExpression `json:"body"`
}

func NewLoopExpression(body *BlockExpression) *LoopExpression {
	return &LoopExpression{nodeImpl: newNodeImpl(NodeLoopExpression), Body: body}
}

type BreakStatement struct {
	nodeImpl
	statementMarker

	Label *Identifier `json:"label,omitempty"`
	Value Expression  `json:"value,omitempty"`
}

func NewBreakStatement(label *Identifier, value Expression) *BreakStatement {
	return &BreakStatement{nodeImpl: newNodeImpl(NodeBreakStatement), Label: label, Value: value}
}

type ContinueStatement struct {
	nodeImpl
	statementMarker

	Label *Identifier `json:"label,omitempty"`
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

type YieldStatement struct {
	nodeImpl
	statementMarker

	Expression Expression `json:"expression,omitempty"`
}

func NewYieldStatement(expression Expression) *YieldStatement {
	return &YieldStatement{nodeImpl: newNodeImpl(NodeYieldStatement), Expression: expression}
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
