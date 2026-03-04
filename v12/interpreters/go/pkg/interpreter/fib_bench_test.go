//go:build !(js && wasm)

package interpreter

import (
	"os"
	"testing"
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
