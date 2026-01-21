package ast

// Type expressions

type TypeExpression interface {
	Node
	typeExpressionNode()
}

type typeExpressionMarker struct{}

func (typeExpressionMarker) typeExpressionNode() {}

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

	TypeParam   TypeExpression         `json:"typeParam"`
	Constraints []*InterfaceConstraint `json:"constraints"`
}

func NewWhereClauseConstraint(typeParam TypeExpression, constraints []*InterfaceConstraint) *WhereClauseConstraint {
	return &WhereClauseConstraint{nodeImpl: newNodeImpl(NodeWhereClauseConstraint), TypeParam: typeParam, Constraints: constraints}
}
