//go:build !(js && wasm)

package interpreter

import (
	"os"
	"testing"

	ableRuntime "able/interpreter-go/pkg/runtime"
)

func BenchmarkFib30Bytecode(b *testing.B) {
	src := []byte(`fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(30)
`)
	tmpFile, err := os.CreateTemp("", "fib*.able")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(src); err != nil {
		b.Fatal(err)
	}
	tmpFile.Close()

	module, err := parseSourceModule(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp := New()
		interp.execMode = execModeBytecode
		_, _, err := interp.EvaluateModule(module)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFib30BytecodeRuntimeOnly(b *testing.B) {
	src := []byte(`fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	tmpFile, err := os.CreateTemp("", "fib-runtime*.able")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(src); err != nil {
		b.Fatal(err)
	}
	tmpFile.Close()

	module, err := parseSourceModule(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	interp := NewBytecode()
	_, env, err := interp.EvaluateModule(module)
	if err != nil {
		b.Fatal(err)
	}
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	fibValue, err := env.Get("fib")
	if err != nil {
		b.Fatal(err)
	}
	args := []ableRuntime.Value{ableRuntime.NewSmallInt(30, ableRuntime.IntegerI32)}

	got, err := interp.CallFunction(fibValue, args)
	if err != nil {
		b.Fatalf("warmup fib: %v", err)
	}
	if want := ableRuntime.NewSmallInt(832040, ableRuntime.IntegerI32); !valuesEqual(got, want) {
		b.Fatalf("warmup fib result mismatch: got=%#v want=%#v", got, want)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := interp.CallFunction(fibValue, args); err != nil {
			b.Fatalf("call fib: %v", err)
		}
	}
}
