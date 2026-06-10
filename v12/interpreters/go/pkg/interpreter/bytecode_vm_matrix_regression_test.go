package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func nestedF64ArrayForTest(t *testing.T, interp *Interpreter, value runtime.Value) [][]float64 {
	t.Helper()
	outer, ok := value.(*runtime.ArrayValue)
	if !ok || outer == nil {
		t.Fatalf("expected nested array value, got %T (%#v)", value, value)
	}
	outerValues, err := interp.ArrayElements(outer)
	if err != nil {
		t.Fatalf("read outer array elements: %v", err)
	}
	rows := make([][]float64, len(outerValues))
	for rowIdx, rowValue := range outerValues {
		row, ok := rowValue.(*runtime.ArrayValue)
		if !ok || row == nil {
			t.Fatalf("expected row array at %d, got %T (%#v)", rowIdx, rowValue, rowValue)
		}
		rowValues, err := interp.ArrayElements(row)
		if err != nil {
			t.Fatalf("read row %d elements: %v", rowIdx, err)
		}
		rows[rowIdx] = make([]float64, len(rowValues))
		for colIdx, cell := range rowValues {
			switch fv := bytecodeSlotReadValue(cell).(type) {
			case runtime.FloatValue:
				rows[rowIdx][colIdx] = fv.Val
			case *runtime.FloatValue:
				if fv == nil {
					t.Fatalf("nil float cell at [%d][%d]", rowIdx, colIdx)
				}
				rows[rowIdx][colIdx] = fv.Val
			default:
				t.Fatalf("expected float at [%d][%d], got %T (%#v)", rowIdx, colIdx, cell, cell)
			}
		}
	}
	return rows
}

func preloadArrayStdlibForTest(t *testing.T, interp *Interpreter) {
	t.Helper()
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: stdlibRoot, Kind: driver.RootStdlib},
		{Path: kernelRoot, Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init failed: %v", err)
	}
	stdlibProgram, err := loader.Load(filepath.Join(stdlibRoot, "collections", "array.able"))
	if err != nil {
		t.Fatalf("load stdlib array failed: %v", err)
	}
	if _, _, _, err := interp.EvaluateProgram(stdlibProgram, ProgramEvaluationOptions{}); err != nil {
		t.Fatalf("evaluate stdlib array failed: %v", err)
	}
}

func TestBytecodeVM_BuildMatrixSmallParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn build_matrix(n: i32) -> Array (Array f64) {
  t := 1.0 / (n as f64) / (n as f64)
  m: Array (Array f64) := Array.new()
  i := 0
  loop {
    if i >= n { break }
    row: Array f64 := Array.new()
    j := 0
    loop {
      if j >= n { break }
      row.push(t * ((i - j) as f64) * ((i + j) as f64))
      j = j + 1
    }
    m.push(row)
    i = i + 1
  }
  m
}

fn main() -> f64 {
  build_matrix(6).get(3)!.get(2)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode build_matrix mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BuildMatrixReturnedNestedArrayParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn build_matrix(n: i32) -> Array (Array f64) {
  t := 1.0 / (n as f64) / (n as f64)
  m: Array (Array f64) := Array.new()
  i := 0
  loop {
    if i >= n { break }
    row: Array f64 := Array.new()
    j := 0
    loop {
      if j >= n { break }
      row.push(t * ((i - j) as f64) * ((i + j) as f64))
      j = j + 1
    }
    m.push(row)
    i = i + 1
  }
  m
}

build_matrix(6)
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)
	wantRows := nestedF64ArrayForTest(t, treeInterp, want)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	gotRows := nestedF64ArrayForTest(t, byteInterp, got)

	if len(gotRows) != len(wantRows) {
		t.Fatalf("build_matrix rows length mismatch: got=%d want=%d", len(gotRows), len(wantRows))
	}
	for rowIdx := range wantRows {
		if len(gotRows[rowIdx]) != len(wantRows[rowIdx]) {
			t.Fatalf("build_matrix row %d length mismatch: got=%d want=%d", rowIdx, len(gotRows[rowIdx]), len(wantRows[rowIdx]))
		}
		for colIdx := range wantRows[rowIdx] {
			if gotRows[rowIdx][colIdx] != wantRows[rowIdx][colIdx] {
				t.Fatalf("build_matrix[%d][%d] mismatch: got=%v want=%v row=%v", rowIdx, colIdx, gotRows[rowIdx][colIdx], wantRows[rowIdx][colIdx], gotRows[rowIdx])
			}
		}
	}
}

func TestBytecodeVM_ChainedFloatMultiplyWithCastsParity(t *testing.T) {
	module := mustParseModuleSource(t, `
fn main() -> f64 {
  t := 1.0 / 6.0 / 6.0
  i := 3
  j := 2
  t * ((i - j) as f64) * ((i + j) as f64)
}

main()
`)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode chained float multiply mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ArrayPushOfChainedFloatMultiplyParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn main() -> f64 {
  row: Array f64 := Array.new()
  t := 1.0 / 6.0 / 6.0
  i := 3
  j := 2
  row.push(t * ((i - j) as f64) * ((i + j) as f64))
  row.get(0)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode array push chained float multiply mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_LoopedArrayPushOfChainedFloatMultiplyParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn main() -> f64 {
  row: Array f64 := Array.new()
  t := 1.0 / 6.0 / 6.0
  i := 3
  j := 0
  loop {
    if j >= 3 { break }
    row.push(t * ((i - j) as f64) * ((i + j) as f64))
    j = j + 1
  }
  row.get(2)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode looped array push chained float multiply mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NestedArrayRowParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn main() -> f64 {
  m: Array (Array f64) := Array.new()
  i := 0
  loop {
    if i >= 4 { break }
    row: Array f64 := Array.new()
    row.push(i as f64)
    m.push(row)
    i = i + 1
  }
  m.get(3)!.get(0)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode nested array row mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NestedLoopFloatComputationParity(t *testing.T) {
	module := mustParseModuleSource(t, `
fn main() -> f64 {
  t := 1.0 / 6.0 / 6.0
  last := 0.0
  i := 0
  loop {
    if i >= 4 { break }
    j := 0
    loop {
      if j >= 3 { break }
      last = t * ((i - j) as f64) * ((i + j) as f64)
      j = j + 1
    }
    i = i + 1
  }
  last
}

main()
`)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode nested loop float computation mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NestedLoopRowBuildParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn main() -> f64 {
  t := 1.0 / 6.0 / 6.0
  last := 0.0
  i := 0
  loop {
    if i >= 4 { break }
    row: Array f64 := Array.new()
    j := 0
    loop {
      if j >= 3 { break }
      row.push(t * ((i - j) as f64) * ((i + j) as f64))
      j = j + 1
    }
    last = row.get(2)!
    i = i + 1
  }
  last
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode nested loop row build mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ReturnedRowArrayParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn build_row() -> Array f64 {
  row: Array f64 := Array.new()
  t := 1.0 / 6.0 / 6.0
  i := 3
  j := 0
  loop {
    if j >= 3 { break }
    row.push(t * ((i - j) as f64) * ((i + j) as f64))
    j = j + 1
  }
  row
}

fn main() -> f64 {
  build_row().get(2)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode returned row array mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ReturnedNestedArrayParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn build_rows() -> Array (Array f64) {
  m: Array (Array f64) := Array.new()
  i := 0
  loop {
    if i >= 4 { break }
    row: Array f64 := Array.new()
    row.push(i as f64)
    m.push(row)
    i = i + 1
  }
  m
}

fn main() -> f64 {
  build_rows().get(3)!.get(0)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode returned nested array mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_MatmulSmallParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn build_matrix(n: i32) -> Array (Array f64) {
  t := 1.0 / (n as f64) / (n as f64)
  m: Array (Array f64) := Array.new()
  i := 0
  loop {
    if i >= n { break }
    row: Array f64 := Array.new()
    j := 0
    loop {
      if j >= n { break }
      row.push(t * ((i - j) as f64) * ((i + j) as f64))
      j = j + 1
    }
    m.push(row)
    i = i + 1
  }
  m
}

fn matmul(a: Array (Array f64), b: Array (Array f64)) -> Array (Array f64) {
  n := a.len()
  c: Array (Array f64) := Array.new()
  i := 0
  loop {
    if i >= n { break }
    ci: Array f64 := Array.new()
    j := 0
    loop {
      if j >= n { break }
      ci.push(b.get(j)!.get(i)!)
      j = j + 1
    }
    c.push(ci)
    i = i + 1
  }

  d: Array (Array f64) := Array.new()
  i = 0
  loop {
    if i >= n { break }
    di: Array f64 := Array.new()
    ai := a.get(i)!
    j := 0
    loop {
      if j >= n { break }
      s := 0.0
      cj := c.get(j)!
      k := 0
      loop {
        if k >= n { break }
        s = s + ai.get(k)! * cj.get(k)!
        k = k + 1
      }
      di.push(s)
      j = j + 1
    }
    d.push(di)
    i = i + 1
  }
  d
}

fn main() -> f64 {
  d := matmul(build_matrix(6), build_matrix(6))
  d.get(3)!.get(3)!
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode matmul mismatch: got=%#v want=%#v", got, want)
	}
}
