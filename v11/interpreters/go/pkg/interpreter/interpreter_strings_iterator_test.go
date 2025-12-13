package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/runtime"
)

func TestStringIteratorRunsAfterStdlibImport(t *testing.T) {
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: filepath.Join("..", "..", "..", "..", "stdlib", "src"), Kind: driver.RootStdlib},
		{Path: filepath.Join("..", "..", "..", "..", "kernel", "src"), Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	stdlibProgram, err := loader.Load(filepath.Join("..", "..", "..", "..", "stdlib", "src", "text", "string.able"))
	if err != nil {
		t.Fatalf("load stdlib string: %v", err)
	}

	interp := New()
	if _, _, _, err := interp.EvaluateProgram(stdlibProgram, ProgramEvaluationOptions{SkipTypecheck: true}); err != nil {
		t.Fatalf("evaluate stdlib string: %v", err)
	}

	iter, err := interp.resolveIteratorValue(runtime.StringValue{Val: "hey"}, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("resolve iterator: %v", err)
	}
	defer iter.Close()

	var bytes []int64
	for {
		val, done, err := iter.Next()
		if err != nil {
			t.Fatalf("iterator step: %v", err)
		}
		if done {
			break
		}
		intVal, ok := val.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("expected integer element, got %T", val)
		}
		bytes = append(bytes, intVal.Val.Int64())
	}

	if len(bytes) != 3 {
		t.Fatalf("expected 3 bytes, got %d (%v)", len(bytes), bytes)
	}
	if bytes[0] != int64('h') || bytes[1] != int64('e') || bytes[2] != int64('y') {
		t.Fatalf("unexpected bytes: %v", bytes)
	}
}
