package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func BenchmarkBytecodeVMForIteratorAdapter(b *testing.B) {
	interp := NewBytecode()
	module := mustParseModuleSource(b, `
package benchiter

struct IteratorEnd {}

struct Counter {
  current: i32,
  limit: i32
}

methods Counter {
  fn next(self: Self) -> i32 | IteratorEnd {
    if self.current >= self.limit {
      IteratorEnd {}
    } else {
      value := self.current
      self.current = self.current + 1
      value
    }
  }
}

fn make_counter() -> Counter {
  Counter { current: 0, limit: 2_000_000_000 }
}
`)
	driverModule := &driver.Module{
		Package: "benchiter",
		AST:     module,
		Files:   []string{"benchiter/main.able"},
	}
	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(module, driverModule.Files[0], origins)
	driverModule.NodeOrigins = origins
	program := &driver.Program{
		Entry:   driverModule,
		Modules: []*driver.Module{driverModule},
	}
	b.Cleanup(func() {
		programEvaluationStateCache.Delete(program)
		loadedModuleBytecodeCache.Delete(driverModule)
	})

	_, env, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    true,
		AllowDiagnostics: false,
	})
	if err != nil {
		b.Fatalf("evaluate module: %v", err)
	}
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	makeCounter, err := env.Get("make_counter")
	if err != nil {
		b.Fatalf("lookup make_counter: %v", err)
	}

	b.Run("struct_next_frame", func(b *testing.B) {
		counter, err := interp.CallFunction(makeCounter, nil)
		if err != nil {
			b.Fatalf("make_counter: %v", err)
		}
		vm := newBytecodeVM(interp, env)
		if err := vm.pushForIterator(counter); err != nil {
			b.Fatalf("pushForIterator: %v", err)
		}

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			value, done, err := vm.nextForIterator()
			if err != nil {
				b.Fatalf("nextForIterator: %v", err)
			}
			if done {
				b.Fatalf("iterator ended at %d", i)
			}
			if _, ok := value.(runtime.IntegerValue); !ok {
				b.Fatalf("next value = %T (%#v), want runtime.IntegerValue", value, value)
			}
		}
	})
}
