package runtime

import "testing"

func TestArrayValueKind(t *testing.T) {
	arr := &ArrayValue{Elements: []Value{StringValue{Val: "a"}}}
	if arr.Kind() != KindArray {
		t.Fatalf("expected KindArray, got %v", arr.Kind())
	}
}

func TestRangeValueKind(t *testing.T) {
	rangeVal := RangeValue{Start: IntegerValue{Val: bigInt(1), TypeSuffix: IntegerI32}, End: IntegerValue{Val: bigInt(3), TypeSuffix: IntegerI32}, Inclusive: true}
	if rangeVal.Kind() != KindRange {
		t.Fatalf("expected KindRange, got %v", rangeVal.Kind())
	}
}

func TestStructInstanceKind(t *testing.T) {
	structDef := StructDefinitionValue{}
	s := &StructInstanceValue{Definition: &structDef, Fields: map[string]Value{"x": IntegerValue{Val: bigInt(1)}}, Positional: nil}
	if s.Kind() != KindStructInstance {
		t.Fatalf("expected KindStructInstance, got %v", s.Kind())
	}
}
