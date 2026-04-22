package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestOverloadArgSignatureDistinguishesFloatSuffix(t *testing.T) {
	f32Sig := overloadArgSignatureForValues([]runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF32},
	})
	f64Sig := overloadArgSignatureForValues([]runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
	})
	if f32Sig == f64Sig {
		t.Fatalf("float overload signatures should differ for f32 vs f64")
	}
}

func TestOverloadArgSignatureDistinguishesHostHandleType(t *testing.T) {
	procSig := overloadArgSignatureForValues([]runtime.Value{
		&runtime.HostHandleValue{HandleType: "ProcHandle"},
	})
	socketSig := overloadArgSignatureForValues([]runtime.Value{
		&runtime.HostHandleValue{HandleType: "SocketHandle"},
	})
	if procSig == socketSig {
		t.Fatalf("host handle overload signatures should differ for distinct handle types")
	}
}
