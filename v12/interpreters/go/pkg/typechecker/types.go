package typechecker

import (
	"able/interpreter-go/pkg/ast"
	"math/big"
)

// Type represents an Able v12 type understood by the checker.
type Type interface {
	Name() string
}

type PrimitiveKind string

const (
	PrimitiveNil        PrimitiveKind = "Nil"
	PrimitiveBool       PrimitiveKind = "Bool"
	PrimitiveChar       PrimitiveKind = "Char"
	PrimitiveString     PrimitiveKind = "String"
	PrimitiveInt        PrimitiveKind = "Int"
	PrimitiveFloat      PrimitiveKind = "Float"
	PrimitiveIoHandle   PrimitiveKind = "IoHandle"
	PrimitiveProcHandle PrimitiveKind = "ProcHandle"
)

type PrimitiveType struct {
	Kind PrimitiveKind
}

func (p PrimitiveType) Name() string { return string(p.Kind) }

type IntegerType struct {
	Suffix   string
	Literal  *big.Int
	Explicit bool
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
	Name            string
	Constraints     []Type
	ConstraintNodes []ast.TypeExpression
	IsInferred      bool
}

// WhereConstraintSpec records a where-clause constraint (e.g. `where T: Display`).
type WhereConstraintSpec struct {
	TypeParam       string
	Subject         Type
	Constraints     []Type
	ConstraintNodes []ast.TypeExpression
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
	TypeArgs   []Type
}

func (s StructInstanceType) Name() string { return "StructInstance:" + s.StructName }

type InterfaceType struct {
	InterfaceName   string
	TypeParams      []GenericParamSpec
	Where           []WhereConstraintSpec
	Methods         map[string]FunctionType
	DefaultMethods  map[string]bool
	SelfTypePattern ast.TypeExpression
}

func (i InterfaceType) Name() string { return "Interface:" + i.InterfaceName }

type AliasType struct {
	AliasName   string
	TypeParams  []GenericParamSpec
	Target      Type
	Where       []WhereConstraintSpec
	Obligations []ConstraintObligation
	Definition  *ast.TypeAliasDefinition
}

func (a AliasType) Name() string {
	if a.AliasName == "" {
		return "Alias"
	}
	return "Alias:" + a.AliasName
}

type PackageType struct {
	Package        string
	Symbols        map[string]Type
	PrivateSymbols map[string]Type
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
	Variants   []Type
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

type FunctionOverloadType struct {
	Overloads []FunctionType
}

func (f FunctionOverloadType) Name() string { return "FunctionOverload" }

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
	ImplName                string
	InterfaceName           string
	Interface               InterfaceType
	TypeParams              []GenericParamSpec
	Target                  Type
	InterfaceArgs           []Type
	Methods                 map[string]FunctionType
	Where                   []WhereConstraintSpec
	MethodWhereClauseCounts map[string]int
	Obligations             []ConstraintObligation
	UnionVariants           []string
	IsBuiltin               bool
	Definition              *ast.ImplementationDefinition
}

type ImplementationNamespaceType struct {
	Impl *ImplementationSpec
}

func (i ImplementationNamespaceType) Name() string {
	if i.Impl != nil && i.Impl.ImplName != "" {
		return "Implementation:" + i.Impl.ImplName
	}
	return "Implementation"
}

type MethodSetSpec struct {
	TypeParams    []GenericParamSpec
	Target        Type
	Methods       map[string]FunctionType
	TypeQualified map[string]bool
	Where         []WhereConstraintSpec
	Obligations   []ConstraintObligation
	Definition    *ast.MethodsDefinition
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

type MapType struct {
	Key   Type
	Value Type
}

func (m MapType) Name() string {
	return "Map"
}

type RangeType struct {
	Element Type
	Bounds  []Type
}

func (r RangeType) Name() string {
	elem := "unknown"
	if r.Element != nil && !isUnknownType(r.Element) {
		elem = r.Element.Name()
	}
	return "Range[" + elem + "]"
}

type IteratorType struct {
	Element Type
}

func (i IteratorType) Name() string {
	elem := "unknown"
	if i.Element != nil && !isUnknownType(i.Element) {
		elem = i.Element.Name()
	}
	return "Iterator[" + elem + "]"
}

type UnknownType struct{}

func (UnknownType) Name() string { return "Unknown" }
