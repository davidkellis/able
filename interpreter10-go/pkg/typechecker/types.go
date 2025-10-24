package typechecker

import "able/interpreter10-go/pkg/ast"

// Type represents an Able v10 type understood by the checker.
type Type interface {
	Name() string
}

type PrimitiveKind string

const (
	PrimitiveNil    PrimitiveKind = "Nil"
	PrimitiveBool   PrimitiveKind = "Bool"
	PrimitiveChar   PrimitiveKind = "Char"
	PrimitiveString PrimitiveKind = "String"
	PrimitiveInt    PrimitiveKind = "Int"
	PrimitiveFloat  PrimitiveKind = "Float"
)

type PrimitiveType struct {
	Kind PrimitiveKind
}

func (p PrimitiveType) Name() string { return string(p.Kind) }

type IntegerType struct {
	Suffix string
}

func (i IntegerType) Name() string {
	return "Int:" + i.Suffix
}

type FloatType struct {
	Suffix string
}

func (f FloatType) Name() string {
	return "Float:" + f.Suffix
}

// GenericParamSpec captures a generic parameter name and its interface constraints.
type GenericParamSpec struct {
	Name        string
	Constraints []Type
}

// WhereConstraintSpec records a where-clause constraint (e.g. `where T: Display`).
type WhereConstraintSpec struct {
	TypeParam   string
	Constraints []Type
}

type StructType struct {
	StructName string
	TypeParams []GenericParamSpec
	Fields     map[string]Type
	Positional []Type
	Where      []WhereConstraintSpec
}

func (s StructType) Name() string { return "Struct:" + s.StructName }

type StructInstanceType struct {
	StructName string
	Fields     map[string]Type
	Positional []Type
}

func (s StructInstanceType) Name() string { return "StructInstance:" + s.StructName }

type InterfaceType struct {
	InterfaceName string
	TypeParams    []GenericParamSpec
	Where         []WhereConstraintSpec
	Methods       map[string]FunctionType
}

func (i InterfaceType) Name() string { return "Interface:" + i.InterfaceName }

type PackageType struct {
	Package string
	Symbols map[string]Type
}

func (p PackageType) Name() string {
	if p.Package == "" {
		return "Package:<unknown>"
	}
	return "Package:" + p.Package
}

type UnionType struct {
	UnionName  string
	TypeParams []GenericParamSpec
	Where      []WhereConstraintSpec
}

func (u UnionType) Name() string { return "Union:" + u.UnionName }

type FunctionType struct {
	Params      []Type
	Return      Type
	TypeParams  []GenericParamSpec
	Where       []WhereConstraintSpec
	Obligations []ConstraintObligation
}

func (f FunctionType) Name() string { return "Function" }

type ProcType struct {
	Result Type
}

func (p ProcType) Name() string { return "Proc" }

type FutureType struct {
	Result Type
}

func (f FutureType) Name() string { return "Future" }

type AppliedType struct {
	Base      Type
	Arguments []Type
}

func (a AppliedType) Name() string { return a.Base.Name() + "<applied>" }

type NullableType struct {
	Inner Type
}

func (n NullableType) Name() string { return "Nullable(" + n.Inner.Name() + ")" }

type UnionLiteralType struct {
	Members []Type
}

func (u UnionLiteralType) Name() string { return "UnionLiteral" }

type ImplementationSpec struct {
	ImplName      string
	InterfaceName string
	TypeParams    []GenericParamSpec
	Target        Type
	InterfaceArgs []Type
	Methods       map[string]FunctionType
	Where         []WhereConstraintSpec
	Obligations   []ConstraintObligation
	Definition    *ast.ImplementationDefinition
}

type MethodSetSpec struct {
	TypeParams  []GenericParamSpec
	Target      Type
	Methods     map[string]FunctionType
	Where       []WhereConstraintSpec
	Obligations []ConstraintObligation
	Definition  *ast.MethodsDefinition
}

type ConstraintObligation struct {
	Owner      string
	TypeParam  string
	Constraint Type
	Subject    Type
	Context    string
	Node       ast.Node
}

type TypeParameterType struct {
	ParameterName string
}

func (t TypeParameterType) Name() string { return "TypeParam:" + t.ParameterName }

type ArrayType struct {
	Element Type
}

func (a ArrayType) Name() string {
	if a.Element == nil || isUnknownType(a.Element) {
		return "Array[unknown]"
	}
	return "Array[" + a.Element.Name() + "]"
}

type RangeType struct {
	Element Type
}

func (r RangeType) Name() string {
	elem := "unknown"
	if r.Element != nil && !isUnknownType(r.Element) {
		elem = r.Element.Name()
	}
	return "Range[" + elem + "]"
}

type UnknownType struct{}

func (UnknownType) Name() string { return "Unknown" }
