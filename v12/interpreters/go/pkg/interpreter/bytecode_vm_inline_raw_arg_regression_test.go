package interpreter

import "testing"

func TestBytecodeVM_InlineNestedU64ArgumentSnapshotParity(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Bitmap T {
  value: u64
}

fn bitpos(hash: u64) -> u64 {
  bit := hash .& 31_u64
  1_u64 .<< bit
}

fn bitcount(value: u64) -> i32 {
  count := 0_i32
  current := value
  loop {
    if current == 0_u64 { break }
    count = count + ((current .& 1_u64) as i32)
    current = current .>> 1_u64
  }
  count
}

methods Bitmap T {
  fn with_bit(self: Self, bit: u64) -> Bitmap T {
    ignored := bitcount(self.value .& (bit - 1_u64))
    Bitmap { value: self.value .| bit }
  }
}

fn build<T>() -> Bitmap T {
  first := bitpos(7_u64)
  second := bitpos(20_u64)
  bitmap := Bitmap { value: 0_u64 }
  bitmap = (bitmap).with_bit(first)
  bitmap = (bitmap).with_bit(second)
  bitmap
}

fn main() -> u64 {
  build<i32>().value
}

main()
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, NewBytecode(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("nested inline u64 argument mismatch: got=%#v want=%#v", got, want)
	}
}
