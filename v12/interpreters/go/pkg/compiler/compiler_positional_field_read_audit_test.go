package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCompilerGeneratedNamedFieldReadsUseSharedAccessor(t *testing.T) {
	result := compileNoFallbackSource(t, "package demo\n\nfn main() -> i32 { 1 }\n")

	tests := []struct {
		helper string
		want   []string
		avoid  []string
	}{
		{
			helper: "__able_string_bytes_from_struct",
			want:   []string{`__able_struct_named_field_value(inst, "bytes")`},
			avoid:  []string{`inst.Fields["bytes"]`, "inst.Positional[0]"},
		},
		{
			helper: "__able_array_values",
			want:   []string{`__able_struct_named_field_value(inst, "storage_handle")`},
			avoid:  []string{`inst.Fields["storage_handle"]`},
		},
		{
			helper: "__able_struct_Array_apply",
			want:   []string{`__able_struct_named_field_value(inst, "storage_handle")`},
			avoid:  []string{`if handleVal, ok := inst.Fields["storage_handle"]`},
		},
		{
			helper: "__able_try_member_set",
			want:   []string{`__able_struct_named_field_value(inst, "storage_handle")`},
			avoid:  []string{`if handleVal, ok := inst.Fields["storage_handle"]`},
		},
		{
			helper: "__able_try_member_get",
			want: []string{
				`__able_struct_named_field_value(inst, "storage_handle")`,
				"__able_struct_named_field_value(inst, name)",
			},
			avoid: []string{
				`inst.Fields["storage_handle"]`,
				"inst.Fields[name]",
				"inst.Positional[idx]",
			},
		},
		{
			helper: "__able_try_member_get_method",
			want:   []string{"__able_struct_named_field_value(inst, name)"},
			avoid:  []string{"inst.Fields[name]"},
		},
		{
			helper: "__able_future_error_details",
			want:   []string{`__able_struct_named_field_value(v, "details")`},
			avoid:  []string{`v.Fields["details"]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.helper, func(t *testing.T) {
			body, ok := findCompiledFunction(result, tt.helper)
			if !ok {
				t.Fatalf("could not find generated helper %s", tt.helper)
			}
			for _, fragment := range tt.want {
				if !strings.Contains(body, fragment) {
					t.Fatalf("%s should contain %q:\n%s", tt.helper, fragment, body)
				}
			}
			for _, fragment := range tt.avoid {
				if strings.Contains(body, fragment) {
					t.Fatalf("%s should avoid representation-specific read %q:\n%s", tt.helper, fragment, body)
				}
			}
		})
	}

	mainSource := compileMainSource(t, "demo", "package demo\n\nfn main() -> i32 { 1 }\n")
	formatStart := strings.Index(mainSource, "func formatRuntimeValue(")
	formatEnd := strings.Index(mainSource, "\nfunc main()")
	if formatStart < 0 || formatEnd <= formatStart {
		t.Fatalf("could not isolate formatRuntimeValue in generated main:\n%s", mainSource)
	}
	formatBody := mainSource[formatStart:formatEnd]
	if !strings.Contains(formatBody, `__able_struct_named_field_value(v, "storage_handle")`) {
		t.Fatalf("formatRuntimeValue should use the shared Array field accessor:\n%s", mainSource)
	}
	if strings.Contains(formatBody, `v.Fields["storage_handle"]`) {
		t.Fatalf("formatRuntimeValue should avoid a representation-specific Array field read:\n%s", mainSource)
	}
	arrayShapeStart := strings.Index(mainSource, "func isArrayStructInstance(")
	arrayShapeEnd := strings.Index(mainSource, "\nfunc isCallableRuntimeValue(")
	if arrayShapeStart < 0 || arrayShapeEnd <= arrayShapeStart {
		t.Fatalf("could not isolate isArrayStructInstance in generated main:\n%s", mainSource)
	}
	arrayShapeBody := mainSource[arrayShapeStart:arrayShapeEnd]
	for _, field := range []string{"storage_handle", "length", "capacity"} {
		if !strings.Contains(arrayShapeBody, `__able_struct_named_field_value(v, "`+field+`")`) {
			t.Fatalf("isArrayStructInstance should use the shared accessor for %s:\n%s", field, arrayShapeBody)
		}
		if strings.Contains(arrayShapeBody, `v.Fields["`+field+`"]`) {
			t.Fatalf("isArrayStructInstance should avoid a representation-specific %s read:\n%s", field, arrayShapeBody)
		}
	}
}

func TestCompilerRuntimeFunctionalUpdateReadsPositionalNamedFields(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	ctx := newCompileContext(gen, nil, nil, nil, "demo", nil)
	source := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Int(1), ast.NewIdentifier("value"), false),
		},
		false,
		ast.NewIdentifier("Record"),
		nil,
		nil,
	)
	update := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Int(2), ast.NewIdentifier("value"), false),
		},
		false,
		ast.NewIdentifier("Record"),
		[]ast.Expression{source},
		nil,
	)

	lines, _, _, ok := gen.compileStructLiteralRuntime(ctx, update)
	if !ok {
		t.Fatalf("runtime functional update did not compile: %s", ctx.reason)
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "__able_struct_named_field_value(") {
		t.Fatalf("runtime functional update should use the shared named-field accessor:\n%s", joined)
	}
	if strings.Contains(joined, ".Fields == nil") {
		t.Fatalf("runtime functional update should accept positional-backed named structs:\n%s", joined)
	}
}
