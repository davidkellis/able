package runtime

import "testing"

func TestArrayValueKind(t *testing.T) {
	arr := &ArrayValue{Elements: []Value{StringValue{Val: "a"}}}
	if arr.Kind() != KindArray {
		t.Fatalf("expected KindArray, got %v", arr.Kind())
	}
}

func TestStructInstanceKind(t *testing.T) {
	structDef := StructDefinitionValue{}
	s := &StructInstanceValue{Definition: &structDef, Fields: map[string]Value{"x": IntegerValue{Val: bigInt(1)}}, Positional: nil}
	if s.Kind() != KindStructInstance {
		t.Fatalf("expected KindStructInstance, got %v", s.Kind())
	}
}
