package interpreter

import (
	"math"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestApplyEqualityInterface_PrimitivesCacheDirectPlans(t *testing.T) {
	tests := []struct {
		name          string
		op            string
		left          runtime.Value
		right         runtime.Value
		want          bool
		interfaceName string
	}{
		{name: "bool", op: "==", left: runtime.BoolValue{Val: true}, right: runtime.BoolValue{Val: true}, want: true, interfaceName: "Eq"},
		{name: "bool_not_equal", op: "!=", left: runtime.BoolValue{Val: true}, right: runtime.BoolValue{Val: false}, want: true, interfaceName: "Eq"},
		{name: "char", op: "==", left: runtime.CharValue{Val: 'a'}, right: runtime.CharValue{Val: 'b'}, want: false, interfaceName: "Eq"},
		{name: "string", op: "==", left: runtime.StringValue{Val: "able"}, right: runtime.StringValue{Val: "able"}, want: true, interfaceName: "Eq"},
		{name: "integer", op: "==", left: runtime.NewSmallInt(7, runtime.IntegerI32), right: runtime.NewSmallInt(7, runtime.IntegerI32), want: true, interfaceName: "Eq"},
		{name: "float", op: "==", left: runtime.FloatValue{Val: math.NaN(), TypeSuffix: runtime.FloatF64}, right: runtime.FloatValue{Val: math.NaN(), TypeSuffix: runtime.FloatF64}, want: false, interfaceName: "PartialEq"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			interp := NewBytecode()
			bootstrapPrimitiveInterfaceNativeTests(t, interp)

			result, handled, err := interp.applyEqualityInterface(test.op, test.left, test.right)
			if err != nil {
				t.Fatalf("primitive equality: %v", err)
			}
			boolResult, ok := result.(runtime.BoolValue)
			if !handled || !ok || boolResult.Val != test.want {
				t.Fatalf("result = %#v, handled=%v; want %v", result, handled, test.want)
			}

			info, ok := interp.getTypeInfoForValue(test.left)
			if !ok {
				t.Fatal("primitive type info unavailable")
			}
			entry, ok := interp.lookupEqualityDispatchCache(interp.cachedTypeInfoName(info))
			if !ok || entry.kind != equalityDispatchCacheMethod {
				t.Fatalf("equality cache entry = %#v, found=%v", entry, ok)
			}
			if entry.dispatch.interfaceName != test.interfaceName {
				t.Fatalf("cached interface = %q, want %q", entry.dispatch.interfaceName, test.interfaceName)
			}
			if !entry.primitive {
				t.Fatal("primitive equality cache entry did not retain a direct plan")
			}
			switch entry.method.(type) {
			case runtime.NativeFunctionValue, *runtime.NativeFunctionValue:
			default:
				t.Fatalf("cached primitive equality callable = %T, want native", entry.method)
			}
		})
	}
}

func TestApplyEqualityInterface_CustomNominalKeepsAbleCallable(t *testing.T) {
	program := mustLoadAbleProgramFromSource(t, `
import able.core.interfaces.{Eq}

struct Key { id: i32 }

impl Eq for Key {
  fn eq(self: Self, other: Self) -> bool { self.id == other.id }
}

fn main() -> bool { Key { id: 7 } == Key { id: 7 } }

main()
`)
	interp := NewBytecode()
	result, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("evaluate custom equality program: %v", err)
	}
	if boolResult, ok := result.(runtime.BoolValue); !ok || !boolResult.Val {
		t.Fatalf("custom equality result = %#v, want true", result)
	}
	entry, ok := interp.lookupEqualityDispatchCache("Key")
	if !ok || entry.kind != equalityDispatchCacheMethod {
		t.Fatalf("custom equality cache entry = %#v, found=%v", entry, ok)
	}
	if entry.primitive {
		t.Fatal("custom equality cache entry unexpectedly retained a primitive plan")
	}
	if _, ok := entry.method.(*runtime.FunctionValue); !ok {
		t.Fatalf("custom equality callable = %T, want Able function", entry.method)
	}
}
