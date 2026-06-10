package runtime

import "testing"

func TestIteratorValueCloseReturnsEndAndRunsCloserOnce(t *testing.T) {
	nextCalls := 0
	closeCalls := 0
	iter := NewIteratorValue(func() (Value, bool, error) {
		nextCalls++
		return NewSmallInt(1, IntegerI32), false, nil
	}, func() {
		closeCalls++
	})

	iter.Close()
	iter.Close()

	if closeCalls != 1 {
		t.Fatalf("closer calls = %d, want 1", closeCalls)
	}
	value, done, err := iter.Next()
	if err != nil {
		t.Fatalf("Next after Close: %v", err)
	}
	if !done {
		t.Fatalf("Next after Close done = false, want true")
	}
	if _, ok := value.(IteratorEndValue); !ok {
		t.Fatalf("Next after Close value = %#v, want IteratorEnd", value)
	}
	if nextCalls != 0 {
		t.Fatalf("closed iterator called next %d times, want 0", nextCalls)
	}
}

func TestIteratorValueRawCloseReturnsEndAndRunsCloserOnce(t *testing.T) {
	nextCalls := 0
	closeCalls := 0
	iter := NewIteratorValueWithRaw(func() (RawValue, bool, error) {
		nextCalls++
		return NewRawIntegerValue(IntegerI64, 1), false, nil
	}, func() {
		closeCalls++
	})

	iter.Close()
	iter.Close()

	if closeCalls != 1 {
		t.Fatalf("closer calls = %d, want 1", closeCalls)
	}
	value, done, err := iter.NextRaw()
	if err != nil {
		t.Fatalf("NextRaw after Close: %v", err)
	}
	if !done {
		t.Fatalf("NextRaw after Close done = false, want true")
	}
	if got := value.Materialize(); got != IteratorEnd {
		t.Fatalf("NextRaw after Close value = %#v, want IteratorEnd", got)
	}
	if nextCalls != 0 {
		t.Fatalf("closed iterator called raw next %d times, want 0", nextCalls)
	}
}

func TestIteratorHostDriverSnapshotPreservesRetainedRootsAndState(t *testing.T) {
	retained := &ArrayValue{Elements: []Value{NewSmallInt(1, IntegerI32)}}
	closed := NewIteratorValueFromHostDriver(IteratorHostDriver{
		Next:     func() (Value, bool, error) { return NewSmallInt(2, IntegerI32), false, nil },
		Retained: []Value{retained},
	}, true)
	driver, isClosed := closed.HostDriverSnapshot()
	if !isClosed || len(driver.Retained) != 1 || driver.Retained[0] != retained {
		t.Fatalf("driver snapshot = (%+v, %v)", driver, isClosed)
	}
	driver.Retained[0] = NilValue{}
	again, _ := closed.HostDriverSnapshot()
	if again.Retained[0] != retained {
		t.Fatal("snapshot exposed mutable retained-root storage")
	}
}
