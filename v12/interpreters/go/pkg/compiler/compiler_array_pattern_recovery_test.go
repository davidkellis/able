package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func TestCompilerMatchArrayPatternRecoversNativeCarrierFromRecoverableSubjectType(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()

	pattern := ast.ArrP([]ast.Pattern{ast.ID("head")}, ast.ID("tail"))
	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"
	ctx.matchSubjectTypeExpr = ast.Gen(ast.Ty("Array"), ast.Ty("i32"))
	lowered, loweredOK := gen.lowerCarrierType(ctx, ctx.matchSubjectTypeExpr)
	recoveredLines, recoveredTemp, recoveredType, recovered := gen.recoverStaticArrayPatternSubject(ctx, "subject", "runtime.Value")
	if !recovered {
		t.Fatalf("expected array pattern subject recovery helper to succeed (type expr=%q lowered=%q loweredOK=%t)", typeExpressionToString(ctx.matchSubjectTypeExpr), lowered, loweredOK)
	}
	if recoveredTemp == "" || recoveredType == "" || len(recoveredLines) == 0 {
		t.Fatalf("expected array pattern subject recovery helper to produce a concrete carrier conversion")
	}
	condLines, _, bindLines, ok := gen.compileMatchPattern(ctx, pattern, "subject", "runtime.Value")
	if !ok {
		t.Fatalf("expected match array pattern to compile, got reason %q", ctx.reason)
	}

	joined := strings.Join(append(condLines, bindLines...), "\n")
	if !strings.Contains(joined, "__able_array_i32_from(") {
		t.Fatalf("expected match array pattern to recover the native array carrier:\n%s", joined)
	}
	if strings.Contains(joined, "__able_array_values(") {
		t.Fatalf("expected match array pattern to avoid runtime array value extraction when a native carrier is recoverable:\n%s", joined)
	}

	head, ok := ctx.lookup("head")
	if !ok {
		t.Fatalf("expected head binding to be recorded in compile context")
	}
	if head.GoType != "int32" {
		t.Fatalf("expected head binding GoType=int32, got %q", head.GoType)
	}

	tail, ok := ctx.lookup("tail")
	if !ok {
		t.Fatalf("expected tail binding to be recorded in compile context")
	}
	if tail.GoType != "*__able_array_i32" {
		t.Fatalf("expected tail binding GoType=*__able_array_i32, got %q", tail.GoType)
	}
	if got := typeExpressionToString(tail.TypeExpr); got != "Array<i32>" {
		t.Fatalf("expected tail binding type expr Array<i32>, got %q", got)
	}
}

func TestCompilerAssignmentArrayPatternRecoversNativeCarrierFromRecoverableSubjectType(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()

	pattern := ast.ArrP([]ast.Pattern{ast.ID("head")}, ast.ID("tail"))
	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"
	ctx.expectedTypeExpr = ast.Gen(ast.Ty("Array"), ast.Ty("i32"))
	lowered, loweredOK := gen.lowerCarrierType(ctx, ctx.expectedTypeExpr)
	recoveredLines, recoveredTemp, recoveredType, recovered := gen.recoverStaticArrayPatternSubject(ctx, "subject", "runtime.Value")
	if !recovered {
		t.Fatalf("expected assignment array pattern subject recovery helper to succeed (type expr=%q lowered=%q loweredOK=%t)", typeExpressionToString(ctx.expectedTypeExpr), lowered, loweredOK)
	}
	if recoveredTemp == "" || recoveredType == "" || len(recoveredLines) == 0 {
		t.Fatalf("expected assignment array pattern subject recovery helper to produce a concrete carrier conversion")
	}
	condLines, _, ok := gen.compileMatchPatternCondition(ctx, pattern, "subject", "runtime.Value")
	if !ok {
		t.Fatalf("expected assignment array pattern condition to compile, got reason %q", ctx.reason)
	}
	bindLines, ok := gen.compileAssignmentPatternBindings(ctx, pattern, "subject", "runtime.Value", patternBindingMode{
		declare:  true,
		newNames: map[string]struct{}{"head": {}, "tail": {}},
	})
	if !ok {
		t.Fatalf("expected assignment array pattern bindings to compile, got reason %q", ctx.reason)
	}

	joined := strings.Join(append(condLines, bindLines...), "\n")
	if !strings.Contains(joined, "__able_array_i32_from(") {
		t.Fatalf("expected pattern assignment array destructuring to recover the native array carrier:\n%s", joined)
	}
	if strings.Contains(joined, "__able_array_values(") {
		t.Fatalf("expected pattern assignment array destructuring to avoid runtime array value extraction when a native carrier is recoverable:\n%s", joined)
	}

	head, ok := ctx.lookup("head")
	if !ok {
		t.Fatalf("expected head binding to be recorded in compile context")
	}
	if head.GoType != "int32" {
		t.Fatalf("expected head binding GoType=int32, got %q", head.GoType)
	}

	tail, ok := ctx.lookup("tail")
	if !ok {
		t.Fatalf("expected tail binding to be recorded in compile context")
	}
	if tail.GoType != "*__able_array_i32" {
		t.Fatalf("expected tail binding GoType=*__able_array_i32, got %q", tail.GoType)
	}
	if got := typeExpressionToString(tail.TypeExpr); got != "Array<i32>" {
		t.Fatalf("expected tail binding type expr Array<i32>, got %q", got)
	}
}

func TestCompilerMatchExpressionArrayPatternRecoversNativeCarrierFromInferredSubject(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()

	subject := ast.ID("source")
	pattern := ast.ArrP([]ast.Pattern{ast.ID("head")}, ast.ID("tail"))
	match := ast.Match(subject, ast.Mc(pattern, ast.Int(1)), ast.Mc(ast.Wc(), ast.Int(0)))

	gen.setTypecheckInference(map[string]typechecker.InferenceMap{
		"demo": {
			subject: typechecker.ArrayType{Element: typechecker.IntegerType{Suffix: "i32"}},
		},
	})

	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "int32"
	lines, _, _, ok := gen.compileMatchExpression(ctx, match, "")
	if !ok {
		t.Fatalf("expected match expression to compile, got reason %q", ctx.reason)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "__able_array_i32_from(") {
		t.Fatalf("expected match expression array pattern to recover the native array carrier:\n%s", joined)
	}
	if strings.Contains(joined, "__able_array_values(") {
		t.Fatalf("expected match expression array pattern to avoid runtime array value extraction when a native carrier is recoverable:\n%s", joined)
	}
}

func TestCompilerPatternAssignmentArrayRecoversNativeCarrierFromInferredSubject(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()

	subject := ast.ID("source")
	pattern := ast.ArrP([]ast.Pattern{ast.ID("head")}, ast.ID("tail"))
	assign := ast.Assign(pattern, subject)

	gen.setTypecheckInference(map[string]typechecker.InferenceMap{
		"demo": {
			subject: typechecker.ArrayType{Element: typechecker.IntegerType{Suffix: "i32"}},
		},
	})

	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"
	lines, _, _, ok := gen.compilePatternAssignment(ctx, assign, pattern)
	if !ok {
		t.Fatalf("expected pattern assignment to compile, got reason %q", ctx.reason)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "__able_array_i32_from(") {
		t.Fatalf("expected pattern assignment array destructuring to recover the native array carrier:\n%s", joined)
	}
	if strings.Contains(joined, "__able_array_values(") {
		t.Fatalf("expected pattern assignment array destructuring to avoid runtime array value extraction when a native carrier is recoverable:\n%s", joined)
	}
}
