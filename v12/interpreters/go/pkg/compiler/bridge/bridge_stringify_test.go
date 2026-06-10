package bridge

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestStringifyStandalonePrimitives(t *testing.T) {
	tests := []struct {
		name  string
		value runtime.Value
		want  string
	}{
		{name: "string", value: runtime.StringValue{Val: "able"}, want: "able"},
		{name: "bool", value: runtime.BoolValue{Val: true}, want: "true"},
		{name: "char", value: runtime.CharValue{Val: 'λ'}, want: "λ"},
		{name: "small integer", value: runtime.NewSmallInt(-17, runtime.IntegerI32), want: "-17"},
		{name: "wide integer", value: runtime.NewBigIntValue(new(big.Int).Lsh(big.NewInt(1), 100), runtime.IntegerU128), want: "1267650600228229401496703205376"},
		{name: "float", value: runtime.FloatValue{Val: 12.3456789, TypeSuffix: runtime.FloatF64}, want: "12.3456789"},
		{name: "nil", value: runtime.NilValue{}, want: "nil"},
	}

	rt := New(nil)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Stringify(rt, test.value)
			if err != nil {
				t.Fatalf("Stringify() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("Stringify() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestStringifyStandaloneNonPrimitiveStillRequiresInterpreter(t *testing.T) {
	_, err := Stringify(New(nil), &runtime.ArrayValue{})
	if err == nil {
		t.Fatal("Stringify() unexpectedly accepted a non-primitive without an interpreter")
	}
}
