package interpreter

import (
	"math"
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
			assertFloatValue(t, got, tc.want.TypeSuffix, tc.want.Val)
			if tc.want.TypeSuffix == runtime.FloatF64 {
				if _, ok := got.(bytecodeRawF64SlotValue); !ok {
					t.Fatalf("direct float result = %#v, want raw f64 slot value", got)
				}
			} else {
				if _, ok := got.(bytecodeRawF32SlotValue); !ok {
					t.Fatalf("direct float result = %#v, want raw f32 slot value", got)
				}
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

func TestBytecodeVM_DirectFloatDivisionRawFastUsesIEEEDivisionByZero(t *testing.T) {
	nan, kind, handled, err := bytecodeDirectFloatDivisionRawFast(0, runtime.FloatF64, 0, runtime.FloatF64)
	if err != nil || !handled || kind != runtime.FloatF64 || !math.IsNaN(nan) {
		t.Fatalf("0.0/0.0 = (%v, %s, %v, %v), want f64 NaN without error", nan, kind, handled, err)
	}

	inf, kind, handled, err := bytecodeDirectFloatDivisionRawFast(1, runtime.FloatF64, 0, runtime.FloatF64)
	if err != nil || !handled || kind != runtime.FloatF64 || !math.IsInf(inf, 1) {
		t.Fatalf("1.0/0.0 = (%v, %s, %v, %v), want +Inf without error", inf, kind, handled, err)
	}
}

func TestEvaluateDivisionFastFloatZeroUsesIEEEDivisionByZero(t *testing.T) {
	nan, err := evaluateDivisionFast(
		runtime.FloatValue{Val: 0, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 0, TypeSuffix: runtime.FloatF64},
	)
	if err != nil {
		t.Fatalf("0.0/0.0 returned unexpected error: %v", err)
	}
	nanVal, ok := nan.(runtime.FloatValue)
	if !ok || nanVal.TypeSuffix != runtime.FloatF64 || !math.IsNaN(nanVal.Val) {
		t.Fatalf("0.0/0.0 = %#v, want f64 NaN", nan)
	}

	inf, err := evaluateDivisionFast(
		runtime.FloatValue{Val: 1, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 0, TypeSuffix: runtime.FloatF64},
	)
	if err != nil {
		t.Fatalf("1.0/0.0 returned unexpected error: %v", err)
	}
	infVal, ok := inf.(runtime.FloatValue)
	if !ok || infVal.TypeSuffix != runtime.FloatF64 || !math.IsInf(infVal.Val, 1) {
		t.Fatalf("1.0/0.0 = %#v, want f64 +Inf", inf)
	}
}

func TestBytecodeVM_DirectFloatCompareFastPath(t *testing.T) {
	f32 := runtime.FloatF32
	cases := []struct {
		name  string
		op    string
		left  runtime.Value
		right runtime.Value
		want  bool
	}{
		{
			name:  "f64_greater",
			op:    ">",
			left:  runtime.FloatValue{Val: 7.5, TypeSuffix: runtime.FloatF64},
			right: runtime.FloatValue{Val: 4.0, TypeSuffix: runtime.FloatF64},
			want:  true,
		},
		{
			name:  "mixed_less_equal",
			op:    "<=",
			left:  runtime.FloatValue{Val: 3.0, TypeSuffix: runtime.FloatF32},
			right: runtime.FloatValue{Val: 3.0, TypeSuffix: runtime.FloatF64},
			want:  true,
		},
		{
			name:  "pointer_not_equal",
			op:    "!=",
			left:  &runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF32},
			right: runtime.FloatValue{Val: normalizeFloat(f32, 1.5), TypeSuffix: runtime.FloatF32},
			want:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, handled := bytecodeDirectFloatCompareFast(tc.op, tc.left, tc.right)
			if !handled {
				t.Fatalf("expected direct float compare fast path to handle %s", tc.op)
			}
			if got.Val != tc.want {
				t.Fatalf("direct float compare result = %v, want %v", got.Val, tc.want)
			}
		})
	}
}

func TestBytecodeVM_DirectFloatCompareFastPathFallsBackForNonFloat(t *testing.T) {
	if _, handled := bytecodeDirectFloatCompareFast(">", runtime.FloatValue{Val: 1, TypeSuffix: runtime.FloatF64}, runtime.NewSmallInt(2, runtime.IntegerI32)); handled {
		t.Fatalf("expected mixed float/integer compare to fall back to existing numeric promotion path")
	}
}

func TestBytecodeVM_LoadRawFloatSlotAvoidsSnapshotAllocation(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		bytecodeRawFloatSlotValue(12.5, runtime.FloatF64),
	}
	vm.stack = make([]runtime.Value, 0, 1)
	instr := &bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}

	allocs := testing.AllocsPerRun(1000, func() {
		vm.stack = vm.stack[:0]
		vm.ip = 0
		if err := vm.execLoadSlotOpcode(instr); err != nil {
			t.Fatalf("load raw float slot failed: %v", err)
		}
	})
	if allocs != 0 {
		t.Fatalf("raw float slot load allocated %.2f times per run", allocs)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected one loaded stack value, got %d", len(vm.stack))
	}
	stackValue := vm.stack[0]
	assertFloatValue(t, stackValue, runtime.FloatF64, 12.5)
	if _, ok := stackValue.(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("loaded raw float stack value = %#v, want raw f64 slot value", stackValue)
	}
}

func TestBytecodeVM_StoreRawFloatSlotReusesCarrierWithoutAllocation(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	raw := bytecodeRawFloatSlotValue(21.5, runtime.FloatF64)
	vm.slots = []runtime.Value{nil}

	allocs := testing.AllocsPerRun(1000, func() {
		vm.slots[0] = nil
		got := vm.storeFloatSlotValue(0, raw)
		if got != raw {
			t.Fatalf("stored raw float = %#v, want original carrier %#v", got, raw)
		}
	})
	if allocs != 0 {
		t.Fatalf("raw float slot store allocated %.2f times per run", allocs)
	}
	if vm.slots[0] != raw {
		t.Fatalf("stored slot value = %#v, want original raw carrier %#v", vm.slots[0], raw)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 21.5)
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

func TestBytecodeVM_DirectFloatCompareParity(t *testing.T) {
	f32 := ast.FloatTypeF32
	module := ast.Mod([]ast.Statement{
		ast.Bin(">=",
			ast.Bin("+", ast.Flt(2.5), ast.FltTyped(1.25, &f32)),
			ast.Bin("-", ast.Flt(5.0), ast.Flt(1.0)),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode direct-float compare mismatch: got=%#v want=%#v", got, want)
	}
}
