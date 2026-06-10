package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestBytecodeVM_ForStructIteratorFrameClosesOnBreak(t *testing.T) {
	module := mustParseModuleSource(t, `
package iterator_frame_close

struct IteratorEnd {}

struct Counter {
  current: i32,
  limit: i32,
  closed: i32
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

  fn close(self: Self) -> void {
    self.closed = self.closed + 10
  }
}

counter := Counter { current: 1, limit: 4, closed: 0 }
sum := 0
for item in counter {
  sum = sum + item
  break
}
sum + counter.closed
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode struct iterator close mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ForGenericStructIteratorUsesCachedBytecodeNext(t *testing.T) {
	interp := NewBytecode()
	module := mustParseModuleSource(t, `
package iterator_frame_generic

struct IteratorEnd {}

struct Counter T {
  current: i32,
  limit: i32,
  value: T
}

methods Counter T {
  fn next(self: Self) -> T | IteratorEnd {
    if self.current >= self.limit {
      IteratorEnd {}
    } else {
      self.current = self.current + 1
      self.value
    }
  }
}

fn make_counter() -> Counter i32 {
  Counter { current: 0, limit: 3, value: 7 }
}
`)
	driverModule := &driver.Module{
		Package: "iterator_frame_generic",
		AST:     module,
		Files:   []string{"iterator_frame_generic/main.able"},
	}
	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(module, driverModule.Files[0], origins)
	driverModule.NodeOrigins = origins
	program := &driver.Program{
		Entry:   driverModule,
		Modules: []*driver.Module{driverModule},
	}
	t.Cleanup(func() {
		programEvaluationStateCache.Delete(program)
		loadedModuleBytecodeCache.Delete(driverModule)
	})

	_, env, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{SkipTypecheck: true})
	if err != nil {
		t.Fatalf("evaluate program: %v", err)
	}
	makeCounter, err := env.Get("make_counter")
	if err != nil {
		t.Fatalf("lookup make_counter: %v", err)
	}
	counter, err := interp.CallFunction(makeCounter, nil)
	if err != nil {
		t.Fatalf("make_counter: %v", err)
	}

	vm := newBytecodeVM(interp, env)
	if err := vm.pushForIterator(counter); err != nil {
		t.Fatalf("pushForIterator: %v", err)
	}
	if len(vm.iterStack) != 1 {
		t.Fatalf("iter stack len = %d, want 1", len(vm.iterStack))
	}
	frame := vm.iterStack[0]
	if frame.nextBytecode == nil {
		if frame.nextFn == nil {
			t.Fatalf("expected resolved next function")
		}
		decl, _ := frame.nextFn.Declaration.(*ast.FunctionDefinition)
		prog, _ := frame.nextFn.Bytecode.(*bytecodeProgram)
		var layout *bytecodeFrameLayout
		if prog != nil {
			layout = prog.frameLayout
		}
		plan := interp.functionRuntimeGenericBindingPlan(frame.nextFn)
		declGenerics := -1
		if decl != nil {
			declGenerics = len(decl.GenericParams)
		}
		methodSetGenerics := -1
		if frame.nextFn.MethodSet != nil {
			methodSetGenerics = len(frame.nextFn.MethodSet.GenericParams)
		}
		paramSlots := -1
		if layout != nil {
			paramSlots = layout.paramSlots
		}
		t.Fatalf("expected cached bytecode next; decl_generics=%d methodset_generics=%d explicit=%t call_local=%t constraints=%t closure_nil=%t layout_nil=%t param_slots=%d method_shorthand=%t needs_env_scopes=%t",
			declGenerics,
			methodSetGenerics,
			plan.explicitUsed,
			plan.callLocalUsed,
			plan.hasGenericConstraints,
			frame.nextFn.Closure == nil,
			layout == nil,
			paramSlots,
			layout != nil && layout.methodShorthand,
			layout != nil && layout.needsEnvScopes,
		)
	}
}
