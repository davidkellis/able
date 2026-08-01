package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestInterfaceMethodDynamicSafetyClassifiesRecursiveSelf(t *testing.T) {
	self := TypeParameterType{ParameterName: "Self"}
	pair := StructType{StructName: "Pair"}
	cases := []struct {
		name   string
		method FunctionType
		safe   bool
	}{
		{
			name:   "ordinary result",
			method: FunctionType{Params: []Type{self}, Return: PrimitiveType{Kind: PrimitiveString}},
			safe:   true,
		},
		{
			name:   "exact Self result",
			method: FunctionType{Params: []Type{self}, Return: self},
			safe:   true,
		},
		{
			name:   "additional Self parameter",
			method: FunctionType{Params: []Type{self, self}, Return: self},
		},
		{
			name:   "nullable parameter",
			method: FunctionType{Params: []Type{self, NullableType{Inner: self}}, Return: self},
		},
		{
			name: "nominal result",
			method: FunctionType{
				Params: []Type{self},
				Return: AppliedType{Base: pair, Arguments: []Type{self}},
			},
		},
		{
			name:   "nullable result",
			method: FunctionType{Params: []Type{self}, Return: NullableType{Inner: self}},
		},
		{
			name: "union result",
			method: FunctionType{
				Params: []Type{self},
				Return: UnionLiteralType{Members: []Type{
					self,
					PrimitiveType{Kind: PrimitiveNil},
				}},
			},
		},
		{
			name: "callable result",
			method: FunctionType{
				Params: []Type{self},
				Return: FunctionType{Params: []Type{self}, Return: self},
			},
		},
		{
			name: "higher-kinded Self application",
			method: FunctionType{
				Params: []Type{self},
				Return: AppliedType{
					Base:      self,
					Arguments: []Type{PrimitiveType{Kind: PrimitiveString}},
				},
			},
		},
		{
			name: "method constraint",
			method: FunctionType{
				Params: []Type{self},
				Return: PrimitiveType{Kind: PrimitiveString},
				Where: []WhereConstraintSpec{{
					TypeParam:   "T",
					Subject:     TypeParameterType{ParameterName: "T"},
					Constraints: []Type{AppliedType{Base: InterfaceType{InterfaceName: "Convert"}, Arguments: []Type{self}}},
				}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			safe, _ := interfaceMethodDynamicSafety(tc.method)
			if safe != tc.safe {
				t.Fatalf("dynamic safety = %t, want %t", safe, tc.safe)
			}
		})
	}
}

func TestInterfaceValueMemberRejectsStaticOnlySelfSignature(t *testing.T) {
	self := TypeParameterType{ParameterName: "Self"}
	iface := InterfaceType{
		InterfaceName: "DuplicatePair",
		Methods: map[string]FunctionType{
			"duplicate_pair": {
				Params: []Type{self},
				Return: AppliedType{
					Base:      StructType{StructName: "Pair"},
					Arguments: []Type{self},
				},
			},
		},
	}
	checker := New()
	checker.global.Define("DuplicatePair", iface)
	env := checker.global.Extend()
	env.Define("value", iface)

	member := ast.Member(ast.ID("value"), "duplicate_pair")
	diags, inferred := checker.checkMemberAccessWithOptions(env, member, true)
	if len(diags) != 1 {
		t.Fatalf("expected one static-only diagnostic, got %v", diags)
	}
	if diags[0].Code != DiagnosticCodeStaticOnlyInterfaceMethod {
		t.Fatalf("diagnostic code = %q", diags[0].Code)
	}
	if !strings.Contains(diags[0].Message, "result contains Self") {
		t.Fatalf("unexpected diagnostic: %s", diags[0].Message)
	}
	if !isUnknownType(inferred) {
		t.Fatalf("unsafe interface member inferred as %s", formatType(inferred))
	}
}

func TestInterfaceValueMemberAllowsExactSelfResult(t *testing.T) {
	self := TypeParameterType{ParameterName: "Self"}
	iface := InterfaceType{
		InterfaceName: "Clone",
		Methods: map[string]FunctionType{
			"clone": {Params: []Type{self}, Return: self},
		},
	}
	checker := New()
	checker.global.Define("Clone", iface)
	env := checker.global.Extend()
	env.Define("value", iface)

	diags, inferred := checker.checkMemberAccessWithOptions(env, ast.Member(ast.ID("value"), "clone"), true)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	fn, ok := inferred.(FunctionType)
	if !ok {
		t.Fatalf("inferred %T, want FunctionType", inferred)
	}
	if _, _, ok := interfaceFromType(fn.Return); !ok {
		t.Fatalf("exact Self result inferred as %s, want interface view", formatType(fn.Return))
	}
}

func TestInterfaceTypeNamespaceDoesNotApplyDynamicSafety(t *testing.T) {
	self := TypeParameterType{ParameterName: "Self"}
	iface := InterfaceType{
		InterfaceName: "DuplicatePair",
		Methods: map[string]FunctionType{
			"duplicate_pair": {
				Params: []Type{self},
				Return: AppliedType{
					Base:      StructType{StructName: "Pair"},
					Arguments: []Type{self},
				},
			},
		},
	}
	checker := New()
	checker.global.Define("DuplicatePair", iface)
	env := checker.global.Extend()

	diags, inferred := checker.checkMemberAccessWithOptions(
		env,
		ast.Member(ast.ID("DuplicatePair"), "duplicate_pair"),
		true,
	)
	if len(diags) != 0 {
		t.Fatalf("static interface namespace should remain callable: %v", diags)
	}
	if _, ok := inferred.(FunctionType); !ok {
		t.Fatalf("inferred %T, want FunctionType", inferred)
	}
}
