package typechecker

import "testing"

func TestBuiltinFixedIntegerArithmeticMethodSignatures(t *testing.T) {
	for _, suffix := range []string{
		"i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128",
	} {
		valueType := IntegerType{Suffix: suffix}
		for _, mode := range []string{"wrapping", "saturating", "checked"} {
			for _, operation := range []string{"add", "sub", "mul"} {
				name := mode + "_" + operation
				method, ok := builtinFixedIntegerArithmeticMethod(valueType, name)
				if !ok {
					t.Fatalf("%s.%s was not recognized", suffix, name)
				}
				if len(method.Params) != 1 || !typesEquivalentForSignature(method.Params[0], valueType) {
					t.Fatalf("%s.%s parameter = %#v, want %s", suffix, name, method.Params, suffix)
				}
				if mode == "checked" {
					nullable, ok := method.Return.(NullableType)
					if !ok || !typesEquivalentForSignature(nullable.Inner, valueType) {
						t.Fatalf("%s.%s return = %#v, want ?%s", suffix, name, method.Return, suffix)
					}
					continue
				}
				if !typesEquivalentForSignature(method.Return, valueType) {
					t.Fatalf("%s.%s return = %#v, want %s", suffix, name, method.Return, suffix)
				}
			}
		}
	}
}

func TestBuiltinFixedIntegerArithmeticExcludesSizeDependentAndUnknownMethods(t *testing.T) {
	for _, suffix := range []string{"isize", "usize"} {
		if _, ok := builtinFixedIntegerArithmeticMethod(
			IntegerType{Suffix: suffix},
			"wrapping_add",
		); ok {
			t.Fatalf("unexpected intrinsic on %s", suffix)
		}
	}
	for _, name := range []string{"wrapping_div", "saturating_pow", "checked_neg", "add"} {
		if _, ok := builtinFixedIntegerArithmeticMethod(
			IntegerType{Suffix: "i32"},
			name,
		); ok {
			t.Fatalf("unexpected intrinsic %s", name)
		}
	}
}

func TestBuiltinFixedIntegerArithmeticCannotBeReplacedByMethodSet(t *testing.T) {
	checker := New()
	i32Type := IntegerType{Suffix: "i32"}
	checker.methodSets = append(checker.methodSets, MethodSetSpec{
		Target: i32Type,
		Methods: map[string]FunctionType{
			"wrapping_add": {
				Params: []Type{i32Type, i32Type},
				Return: PrimitiveType{Kind: PrimitiveBool},
			},
		},
	})

	method, ok, detail := checker.lookupMethod(i32Type, "wrapping_add", true, false)
	if !ok || detail != "" {
		t.Fatalf("lookupMethod failed: ok=%v detail=%q", ok, detail)
	}
	if !typesEquivalentForSignature(method.Return, i32Type) {
		t.Fatalf("reserved intrinsic return = %#v, want i32", method.Return)
	}
}
