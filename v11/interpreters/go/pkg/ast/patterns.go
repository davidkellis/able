package ast

// Patterns

type Pattern interface {
	Node
	patternNode()
}

type patternMarker struct{}

func (patternMarker) patternNode() {}

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
