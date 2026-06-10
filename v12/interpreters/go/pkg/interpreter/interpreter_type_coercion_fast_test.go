package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterCoerceValueToTypeWouldBeNoOp(t *testing.T) {
	interp := New()
	interp.interfaces["Box"] = &runtime.InterfaceDefinitionValue{}

	if !interp.coerceValueToTypeWouldBeNoOp(ast.Gen(ast.Ty("Array"), ast.Ty("i32"))) {
		t.Fatalf("expected Array i32 coercion to be a no-op")
	}
	if interp.coerceValueToTypeWouldBeNoOp(ast.Gen(ast.Ty("Box"), ast.Ty("i32"))) {
		t.Fatalf("expected generic interface coercion to require runtime work")
	}
	if interp.coerceValueToTypeWouldBeNoOp(ast.Ty("i32")) {
		t.Fatalf("expected simple primitive coercion to remain active")
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast(t *testing.T) {
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)

	casted, ok, err := castValueToCanonicalSimpleTypeFast("i32", intVal)
	if err != nil {
		t.Fatalf("unexpected same-type cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected i32 fast cast to handle same-type integer")
	}
	if !valuesEqual(casted, intVal) {
		t.Fatalf("unexpected same-type cast result: got=%#v want=%#v", casted, intVal)
	}

	floatCasted, ok, err := castValueToCanonicalSimpleTypeFast("f64", intVal)
	if err != nil {
		t.Fatalf("unexpected integer-to-float cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected f64 fast cast to handle integer input")
	}
	floatVal, ok := floatCasted.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float result, got %T", floatCasted)
	}
	if floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
		t.Fatalf("unexpected float cast result: got=%#v", floatVal)
	}
}

func TestInterpreterCoerceValueToTypeUnsignedIntegerWideningUsesValueRange(t *testing.T) {
	interp := New()
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	u64Val, err := interp.coerceValueToType(ast.Ty("u64"), value)
	if err != nil {
		t.Fatalf("unexpected u64 coercion error: %v", err)
	}
	assertIntValue(t, u64Val, runtime.IntegerU64, 7)

	usizeVal, err := interp.coerceValueToType(ast.Ty("usize"), value)
	if err != nil {
		t.Fatalf("unexpected usize coercion error: %v", err)
	}
	assertIntValue(t, usizeVal, runtime.IntegerUsize, 7)
}

func TestCoerceIntegerValueToTargetKindUsesRawIntegerCarrierFastPath(t *testing.T) {
	i32Val, ok := coerceIntegerValueToTargetKindIfInRange(bytecodeRawI32SlotValue(7), runtime.IntegerI32)
	if !ok {
		t.Fatalf("expected raw i32 to coerce to i32")
	}
	assertIntValue(t, i32Val, runtime.IntegerI32, 7)

	i64Val, ok := coerceIntegerValueToTargetKindIfInRange(bytecodeRawI32SlotValue(7), runtime.IntegerI64)
	if !ok {
		t.Fatalf("expected raw i32 to coerce to i64")
	}
	assertIntValue(t, i64Val, runtime.IntegerI64, 7)
}

func TestCoerceIntegerValueToTargetKindPreservesHighBitRawUnsignedValues(t *testing.T) {
	value := bytecodeRawU64ResultValue(^uint64(0))

	coerced, ok := coerceIntegerValueToTargetKindIfInRange(value, runtime.IntegerU64)
	if !ok {
		t.Fatalf("expected high-bit raw u64 to coerce to u64")
	}
	intVal, ok := coerced.(runtime.IntegerValue)
	if !ok || intVal.TypeSuffix != runtime.IntegerU64 || intVal.IsSmall() {
		t.Fatalf("unexpected high-bit raw u64 coercion result: %#v", coerced)
	}
	if got := intVal.BigInt().String(); got != "18446744073709551615" {
		t.Fatalf("high-bit raw u64 coercion = %s, want 18446744073709551615", got)
	}

	if coerced, ok := coerceIntegerValueToTargetKindIfInRange(value, runtime.IntegerI64); ok {
		t.Fatalf("high-bit raw u64 should not coerce to i64, got %#v", coerced)
	}
}

func TestCoerceIntegerValueToTargetKindRawI32AvoidsAlloc(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		coerced, ok := coerceIntegerValueToTargetKindIfInRange(bytecodeRawI32SlotValue(7), runtime.IntegerI32)
		if !ok {
			t.Fatalf("expected raw i32 to coerce to i32")
		}
		assertIntValue(t, coerced, runtime.IntegerI32, 7)
	})
	if allocs != 0 {
		t.Fatalf("expected repeated raw i32 coercions to avoid allocations, got %.2f", allocs)
	}
}

func TestIntegerValueToFloat64FastSmallIntegerWithoutAlloc(t *testing.T) {
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)
	allocs := testing.AllocsPerRun(1000, func() {
		if got := integerValueToFloat64Fast(intVal); got != 7 {
			t.Fatalf("integerValueToFloat64Fast() = %v, want 7", got)
		}
		if got := integerRefToFloat64Fast(&intVal); got != 7 {
			t.Fatalf("integerRefToFloat64Fast() = %v, want 7", got)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected repeated small integer to float helpers to allocate zero, got %.2f", allocs)
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast_SmallIntegerToFloatAvoidsBigIntPath(t *testing.T) {
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)
	allocs := testing.AllocsPerRun(1000, func() {
		casted, ok, err := castValueToCanonicalSimpleTypeFast("f64", intVal)
		if err != nil {
			t.Fatalf("unexpected f64 cast error: %v", err)
		}
		if !ok {
			t.Fatalf("expected f64 fast cast to handle small integer input")
		}
		floatVal, ok := casted.(runtime.FloatValue)
		if !ok || floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
			t.Fatalf("unexpected f64 cast result: %#v", casted)
		}
	})
	if allocs > 2 {
		t.Fatalf("expected repeated small integer to f64 casts to avoid the big-int path, got %.2f allocations", allocs)
	}
}

func TestInlineCoerceValueBySimpleTypeSmallIntegerToFloatAvoidsBigIntPath(t *testing.T) {
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)
	allocs := testing.AllocsPerRun(1000, func() {
		casted, ok, err := inlineCoerceValueBySimpleType("f64", &intVal)
		if err != nil {
			t.Fatalf("unexpected f64 inline coercion error: %v", err)
		}
		if !ok {
			t.Fatalf("expected f64 inline coercion to handle small integer input")
		}
		floatVal, ok := casted.(runtime.FloatValue)
		if !ok || floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
			t.Fatalf("unexpected f64 inline coercion result: %#v", casted)
		}
	})
	if allocs > 1 {
		t.Fatalf("expected repeated inline small integer to f64 coercions to avoid the big-int path, got %.2f allocations", allocs)
	}
}

func TestInlineCoerceValueBySimpleTypeUnsignedIntegerWideningUsesValueRange(t *testing.T) {
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	u64Val, ok, err := inlineCoerceValueBySimpleType("u64", value)
	if err != nil {
		t.Fatalf("unexpected u64 inline coercion error: %v", err)
	}
	if !ok {
		t.Fatalf("expected u64 inline coercion to handle fitting i32 input")
	}
	assertIntValue(t, u64Val, runtime.IntegerU64, 7)

	usizeVal, ok, err := inlineCoerceValueBySimpleType("usize", value)
	if err != nil {
		t.Fatalf("unexpected usize inline coercion error: %v", err)
	}
	if !ok {
		t.Fatalf("expected usize inline coercion to handle fitting i32 input")
	}
	assertIntValue(t, usizeVal, runtime.IntegerUsize, 7)
}

func TestInterpreterCoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath(t *testing.T) {
	interp := New()
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)
	targetType := ast.Ty("f64")
	allocs := testing.AllocsPerRun(1000, func() {
		coerced, err := interp.coerceValueToType(targetType, intVal)
		if err != nil {
			t.Fatalf("unexpected f64 coercion error: %v", err)
		}
		floatVal, ok := coerced.(runtime.FloatValue)
		if !ok || floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
			t.Fatalf("unexpected f64 coercion result: %#v", coerced)
		}
	})
	if allocs > 2 {
		t.Fatalf("expected repeated coerceValueToType small integer to f64 coercions to stay at or below the current fast-path floor, got %.2f allocations", allocs)
	}
}

func TestInterpreterCoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath(t *testing.T) {
	interp := New()
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)
	targetType := ast.Ty("f64")
	env := interp.GlobalEnvironment()
	allocs := testing.AllocsPerRun(1000, func() {
		coerced, err := interp.coerceReturnValue(targetType, intVal, nil, env)
		if err != nil {
			t.Fatalf("unexpected f64 return coercion error: %v", err)
		}
		floatVal, ok := coerced.(runtime.FloatValue)
		if !ok || floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
			t.Fatalf("unexpected f64 return coercion result: %#v", coerced)
		}
	})
	if allocs > 2 {
		t.Fatalf("expected repeated coerceReturnValue small integer to f64 coercions to stay at or below the current fast-path floor, got %.2f allocations", allocs)
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast_SmallIntegerWraps(t *testing.T) {
	negVal := runtime.NewSmallInt(-1, runtime.IntegerI16)

	u8Casted, ok, err := castValueToCanonicalSimpleTypeFast("u8", negVal)
	if err != nil {
		t.Fatalf("unexpected u8 cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected u8 fast cast to handle small integer input")
	}
	u8Val, ok := u8Casted.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", u8Casted)
	}
	if !u8Val.IsSmall() || u8Val.TypeSuffix != runtime.IntegerU8 || u8Val.Int64Fast() != 255 {
		t.Fatalf("unexpected u8 cast result: %#v", u8Val)
	}

	u64Casted, ok, err := castValueToCanonicalSimpleTypeFast("u64", negVal)
	if err != nil {
		t.Fatalf("unexpected u64 cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected u64 fast cast to handle small integer input")
	}
	u64Val, ok := u64Casted.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", u64Casted)
	}
	if u64Val.IsSmall() {
		t.Fatalf("expected wrapped u64 cast to fall back to big integer storage")
	}
	if got := u64Val.BigInt().String(); got != "18446744073709551615" {
		t.Fatalf("unexpected u64 wrap result: %s", got)
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast_SmallIntegerWrapsWithoutAlloc(t *testing.T) {
	negVal := runtime.NewSmallInt(-1, runtime.IntegerI16)
	allocs := testing.AllocsPerRun(1000, func() {
		casted, ok, err := castValueToCanonicalSimpleTypeFast("u8", negVal)
		if err != nil {
			t.Fatalf("unexpected u8 cast error: %v", err)
		}
		if !ok {
			t.Fatalf("expected u8 fast cast to handle small integer input")
		}
		u8Val, ok := casted.(runtime.IntegerValue)
		if !ok || !u8Val.IsSmall() || u8Val.Int64Fast() != 255 {
			t.Fatalf("unexpected u8 cast result: %#v", casted)
		}
	})
	if allocs > 1 {
		t.Fatalf("expected repeated small u8 casts to stay at or below one allocation, got %.2f", allocs)
	}
}

func TestInterpreterFastExactNamedStructTypeMatch(t *testing.T) {
	interp := New()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}

	matched, ok := fastExactNamedStructTypeMatch(interp, ast.Ty("Node"), value)
	if !ok || !matched {
		t.Fatalf("expected exact named struct fast match")
	}
}

func TestInterpreterCoerceValueToTypeExactNamedStructReturnsSameValue(t *testing.T) {
	interp := New()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}

	coerced, err := interp.coerceValueToType(ast.Ty("Node"), value)
	if err != nil {
		t.Fatalf("coerceValueToType: %v", err)
	}
	if coerced != value {
		t.Fatalf("expected exact named struct coercion to reuse value")
	}
}

func TestInterpreterCoerceReturnValueExactNamedStructReturnsSameValue(t *testing.T) {
	interp := New()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}

	coerced, err := interp.coerceReturnValue(ast.Ty("Node"), value, nil, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("coerceReturnValue: %v", err)
	}
	if coerced != value {
		t.Fatalf("expected exact named struct return coercion to reuse value")
	}
}

func TestInterpreterCoerceValueToTypeExactNamedStructUnwrapsErrorPayload(t *testing.T) {
	interp := New()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	payload := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	errVal := runtime.ErrorValue{
		Message: "boom",
		Payload: map[string]runtime.Value{"value": payload},
	}

	coerced, err := interp.coerceValueToType(ast.Ty("Node"), errVal)
	if err != nil {
		t.Fatalf("coerceValueToType: %v", err)
	}
	if coerced != payload {
		t.Fatalf("expected exact named struct coercion to unwrap payload")
	}
}

func TestInlineCoercionUnnecessaryWithInterpreterExactNamedStruct(t *testing.T) {
	interp := New()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}

	if !inlineCoercionUnnecessaryWithInterpreter(interp, ast.Ty("Node"), value) {
		t.Fatalf("expected exact named struct inline coercion fast path")
	}
	if !inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(interp, "Node", value) {
		t.Fatalf("expected exact named struct simple-type inline coercion fast path")
	}
}

func TestInlineCoercionUnnecessaryWithInterpreterDoesNotTreatErrorAsExactNamedStruct(t *testing.T) {
	interp := New()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	payload := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	errVal := runtime.ErrorValue{
		Payload: map[string]runtime.Value{
			"value": payload,
		},
	}

	if !inlineCoercionUnnecessaryWithInterpreter(interp, ast.Ty("Node"), errVal) {
		t.Fatalf("expected error payload exact named struct inline coercion fast path")
	}
	if !inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(interp, "Node", errVal) {
		t.Fatalf("expected error payload exact named struct simple-type inline coercion fast path")
	}
	if inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(interp, "Error", payload) {
		t.Fatalf("did not expect exact named struct helper to treat payload as Error")
	}
}

func TestInlineExactNamedStructNoCoercionBytecodeExactDef(t *testing.T) {
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	defVal := &runtime.StructDefinitionValue{Node: nodeDef}
	value := &runtime.StructInstanceValue{
		Definition: defVal,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}

	if !inlineExactNamedStructNoCoercionBytecodeExactDef(defVal, value) {
		t.Fatalf("expected bytecode exact-def helper to accept exact struct instance")
	}
	singletonDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Leaf",
			nil,
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	if !inlineExactNamedStructNoCoercionBytecodeExactDef(singletonDef, singletonDef) {
		t.Fatalf("expected bytecode exact-def helper to accept exact singleton struct definition")
	}
}

func TestInlineExactNamedStructNoCoercionBytecodeExactDefDoesNotUnwrapErrorPayload(t *testing.T) {
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	payload := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	errVal := runtime.ErrorValue{
		Payload: map[string]runtime.Value{
			"value": payload,
		},
	}

	if inlineExactNamedStructNoCoercionBytecodeExactDef(payload.Definition, errVal) {
		t.Fatalf("did not expect bytecode exact-def helper to treat error payload as a no-coercion hit")
	}
}
