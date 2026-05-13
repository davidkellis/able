package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_DirectFloatArithmeticFastPath(t *testing.T) {
	f32 := runtime.FloatF32
	cases := []struct {
		name  string
		op    string
		left  runtime.Value
		right runtime.Value
		want  runtime.FloatValue
	}{
		{
			name:  "f64_add",
			op:    "+",
			left:  runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF64},
			right: runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64},
			want:  runtime.FloatValue{Val: 3.75, TypeSuffix: runtime.FloatF64},
		},
		{
			name:  "f64_multiply",
			op:    "*",
			left:  runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
			right: runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64},
			want:  runtime.FloatValue{Val: 7.5, TypeSuffix: runtime.FloatF64},
		},
		{
			name:  "f32_subtract_normalizes",
			op:    "-",
			left:  runtime.FloatValue{Val: 1.1, TypeSuffix: runtime.FloatF32},
			right: &runtime.FloatValue{Val: 0.2, TypeSuffix: runtime.FloatF32},
			want:  runtime.FloatValue{Val: normalizeFloat(runtime.FloatF32, 0.9), TypeSuffix: runtime.FloatF32},
		},
		{
			name:  "mixed_widens_to_f64",
			op:    "+",
			left:  runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF32},
			right: runtime.FloatValue{Val: 2.25, TypeSuffix: runtime.FloatF64},
			want:  runtime.FloatValue{Val: 3.75, TypeSuffix: runtime.FloatF64},
		},
		{
			name:  "explicit_f32_literal_shape",
			op:    "*",
			left:  runtime.FloatValue{Val: 1.5, TypeSuffix: f32},
			right: runtime.FloatValue{Val: 2, TypeSuffix: f32},
			want:  runtime.FloatValue{Val: normalizeFloat(f32, 3), TypeSuffix: f32},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, handled := bytecodeDirectFloatArithmeticFast(tc.op, tc.left, tc.right)
			if !handled {
				t.Fatalf("expected direct float fast path to handle %s", tc.op)
			}
			floatVal, ok := got.(runtime.FloatValue)
			if !ok {
				t.Fatalf("direct float result = %#v, want FloatValue", got)
			}
			if floatVal != tc.want {
				t.Fatalf("direct float result = %#v, want %#v", floatVal, tc.want)
			}
		})
	}
}

func TestBytecodeVM_DirectFloatArithmeticFastPathFallsBackForNonFloat(t *testing.T) {
	if _, handled := bytecodeDirectFloatArithmeticFast("+", runtime.FloatValue{Val: 1, TypeSuffix: runtime.FloatF64}, runtime.NewSmallInt(2, runtime.IntegerI32)); handled {
		t.Fatalf("expected mixed float/integer to fall back to existing numeric promotion path")
	}
	if _, handled := bytecodeDirectFloatArithmeticFast("/", runtime.FloatValue{Val: 1, TypeSuffix: runtime.FloatF64}, runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64}); handled {
		t.Fatalf("expected division to fall back to existing division-by-zero checked path")
	}
}

func TestBytecodeVM_DirectFloatBinaryParity(t *testing.T) {
	f32 := ast.FloatTypeF32
	module := ast.Mod([]ast.Statement{
		ast.Bin("+",
			ast.Bin("*", ast.Flt(2.5), ast.Flt(4.0)),
			ast.Bin("-", ast.FltTyped(3.5, &f32), ast.FltTyped(1.25, &f32)),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode direct-float binary mismatch: got=%#v want=%#v", got, want)
	}
}
