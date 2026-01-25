package runtime

import (
	"math/big"
	"testing"
)

func TestEnvironmentDefineAndGet(t *testing.T) {
	env := NewEnvironment(nil)
	value := StringValue{Val: "hello"}
	env.Define("greeting", value)

	got, err := env.Get("greeting")
	if err != nil {
		t.Fatalf("expected to retrieve binding: %v", err)
	}

	if gv, ok := got.(StringValue); !ok || gv.Val != "hello" {
		t.Fatalf("unexpected value returned: %#v", got)
	}
}

func TestEnvironmentAssignRespectsLexicalParent(t *testing.T) {
	env := NewEnvironment(nil)
	env.Define("counter", IntegerValue{Val: bigInt(1), TypeSuffix: IntegerI32})

	child := NewEnvironment(env)
	if err := child.Assign("counter", IntegerValue{Val: bigInt(2), TypeSuffix: IntegerI32}); err != nil {
		t.Fatalf("assign into parent failed: %v", err)
	}

	got, err := env.Get("counter")
	if err != nil {
		t.Fatalf("parent lookup failed: %v", err)
	}
	if iv, ok := got.(IntegerValue); !ok || iv.Val.Cmp(bigInt(2)) != 0 {
		t.Fatalf("unexpected counter value: %#v", got)
	}
}

func TestEnvironmentAssignUnknownFails(t *testing.T) {
	env := NewEnvironment(nil)
	err := env.Assign("missing", NilValue{})
	if err == nil {
		t.Fatalf("expected error when assigning undefined variable")
	}
	if err.Error() != "Undefined variable 'missing'" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
