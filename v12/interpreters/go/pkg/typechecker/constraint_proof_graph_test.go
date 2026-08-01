package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestConstraintProofGraphClosesAnchoredRecursiveSelfObligation(t *testing.T) {
	iface := InterfaceType{InterfaceName: "Recursive"}
	box := StructType{StructName: "Box"}
	checker := New()
	checker.implementations = []ImplementationSpec{{
		InterfaceName: "Recursive",
		Interface:     iface,
		Target:        box,
		Obligations: []ConstraintObligation{{
			Owner:      "impl Recursive for Box interface requirements",
			TypeParam:  "Self",
			Subject:    box,
			Constraint: iface,
		}},
	}}

	ok, detail := checker.typeImplementsInterface(box, iface, nil)
	if !ok {
		t.Fatalf("anchored recursive proof failed: %s", detail)
	}
}

func TestConstraintProofGraphClosesMutuallyAnchoredObligations(t *testing.T) {
	first := InterfaceType{InterfaceName: "First"}
	second := InterfaceType{InterfaceName: "Second"}
	box := StructType{StructName: "Box"}
	checker := New()
	checker.implementations = []ImplementationSpec{
		{
			InterfaceName: "First",
			Interface:     first,
			Target:        box,
			Obligations: []ConstraintObligation{{
				Owner:      "impl First for Box interface requirements",
				TypeParam:  "Self",
				Subject:    box,
				Constraint: second,
			}},
		},
		{
			InterfaceName: "Second",
			Interface:     second,
			Target:        box,
			Obligations: []ConstraintObligation{{
				Owner:      "impl Second for Box interface requirements",
				TypeParam:  "Self",
				Subject:    box,
				Constraint: first,
			}},
		},
	}

	ok, detail := checker.typeImplementsInterface(box, first, nil)
	if !ok {
		t.Fatalf("mutually anchored recursive proof failed: %s", detail)
	}
}

func TestConstraintProofGraphRejectsGrowingCycle(t *testing.T) {
	iface := InterfaceType{InterfaceName: "Recursive"}
	box := StructType{StructName: "Box"}
	wrapper := StructType{
		StructName: "Wrapper",
		TypeParams: []GenericParamSpec{{Name: "T"}},
	}
	param := TypeParameterType{ParameterName: "T"}
	checker := New()
	checker.implementations = []ImplementationSpec{{
		InterfaceName: "Recursive",
		Interface:     iface,
		TypeParams:    []GenericParamSpec{{Name: "T"}},
		Target:        param,
		Obligations: []ConstraintObligation{{
			Owner:     "impl Recursive for T",
			TypeParam: "T",
			Subject: AppliedType{
				Base:      wrapper,
				Arguments: []Type{param},
			},
			Constraint: iface,
		}},
	}}

	ok, detail := checker.typeImplementsInterface(box, iface, nil)
	if ok {
		t.Fatal("growing recursive proof unexpectedly succeeded")
	}
	if !strings.Contains(detail, "not well-founded") {
		t.Fatalf("unexpected growing-cycle detail: %s", detail)
	}
}

func TestInterfaceWhereSelfObligationIsAppliedToImplementation(t *testing.T) {
	recursive := ast.Iface(
		"Recursive",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"identity",
				[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
				ast.Ty("Self"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint(ast.Ty("Self"), ast.InterfaceConstr(ast.Ty("Recursive"))),
		},
		nil,
		false,
	)
	box := ast.StructDef("Box", nil, ast.StructKindSingleton, nil, nil, false)
	implementation := ast.Impl(
		"Recursive",
		ast.Ty("Box"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"identity",
				[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
				[]ast.Statement{ast.Ret(ast.ID("self"))},
				ast.Ty("Self"),
				nil,
				nil,
				false,
				false,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	checker := New()
	diags, err := checker.CheckModule(ast.NewModule(
		[]ast.Statement{recursive, box, implementation},
		nil,
		nil,
	))
	if err != nil {
		t.Fatalf("CheckModule: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("anchored interface where Self obligation failed: %v", diags)
	}
}
