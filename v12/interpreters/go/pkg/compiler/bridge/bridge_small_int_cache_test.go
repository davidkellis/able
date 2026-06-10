package bridge

import (
	"sync"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var bridgeSmallIntCacheSink runtime.Value

func TestCommonI32BoxInitializesOnlyRequestedSlots(t *testing.T) {
	commonI32Boxes = [commonI32BoxMax - commonI32BoxMin + 1]lazyIntegerBox{}

	if got := ToInt(42, runtime.IntegerI32); got == nil {
		t.Fatal("ToInt(42, i32) returned nil")
	}
	initialized := 0
	for index := range commonI32Boxes {
		if commonI32Boxes[index].value != nil {
			initialized++
		}
	}
	if initialized != 1 {
		t.Fatalf("initialized common i32 slots = %d, want 1", initialized)
	}
}

func TestToIntCommonI32BoxPreservesValueAndAvoidsAllocation(t *testing.T) {
	for _, value := range []int64{commonI32BoxMin, -1, 0, 42, commonI32BoxMax} {
		got := ToInt(value, runtime.IntegerI32)
		integer, ok := got.(runtime.IntegerValue)
		if !ok || !integer.IsSmall() || integer.Int64Fast() != value || integer.TypeSuffix != runtime.IntegerI32 {
			t.Fatalf("ToInt(%d, i32) = %#v, want small i32", value, got)
		}
	}

	allocations := testing.AllocsPerRun(1000, func() {
		bridgeSmallIntCacheSink = ToInt(42, runtime.IntegerI32)
	})
	if allocations > 0.1 {
		t.Fatalf("ToInt common i32 allocations/run = %f, want <= 0.1", allocations)
	}
}

func TestToIntCommonI32BoxFallsBackOutsideRangeAndForOtherTypes(t *testing.T) {
	for _, test := range []struct {
		value  int64
		suffix runtime.IntegerType
	}{
		{commonI32BoxMin - 1, runtime.IntegerI32},
		{commonI32BoxMax + 1, runtime.IntegerI32},
		{42, runtime.IntegerI64},
	} {
		got := ToInt(test.value, test.suffix)
		integer, ok := got.(runtime.IntegerValue)
		if !ok || !integer.IsSmall() || integer.Int64Fast() != test.value || integer.TypeSuffix != test.suffix {
			t.Fatalf("ToInt(%d, %s) = %#v, want small integer with original suffix", test.value, test.suffix, got)
		}
	}
}

func TestToIntConcurrentCommonValues(t *testing.T) {
	values := []int64{commonI32BoxMin, -1, 0, 42, commonI32BoxMax}
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		for _, value := range values {
			wait.Add(1)
			go func() {
				defer wait.Done()
				for iteration := 0; iteration < 100; iteration++ {
					got := ToInt(value, runtime.IntegerI32)
					integer, ok := got.(runtime.IntegerValue)
					if !ok || integer.Int64Fast() != value || integer.TypeSuffix != runtime.IntegerI32 {
						t.Errorf("ToInt(%d, i32) = %#v, want concurrent small i32", value, got)
						return
					}
				}
			}()
		}
	}
	wait.Wait()
}

func TestToDynamicI64PreservesValueAndAvoidsAllocation(t *testing.T) {
	for _, value := range []int64{commonI32BoxMin, -1, 0, 42, commonI32BoxMax} {
		got := ToDynamicI64(value)
		integer, ok := got.(runtime.IntegerValue)
		if !ok || !integer.IsSmall() || integer.Int64Fast() != value || integer.TypeSuffix != runtime.IntegerI64 {
			t.Fatalf("ToDynamicI64(%d) = %#v, want small i64", value, got)
		}
	}

	allocations := testing.AllocsPerRun(1000, func() {
		bridgeSmallIntCacheSink = ToDynamicI64(42)
	})
	if allocations > 0.1 {
		t.Fatalf("ToDynamicI64 common allocations/run = %f, want <= 0.1", allocations)
	}
}

func TestToDynamicI64FallsBackOutsideCommonRange(t *testing.T) {
	for _, value := range []int64{commonI32BoxMin - 1, commonI32BoxMax + 1} {
		got := ToDynamicI64(value)
		integer, ok := got.(runtime.IntegerValue)
		if !ok || !integer.IsSmall() || integer.Int64Fast() != value || integer.TypeSuffix != runtime.IntegerI64 {
			t.Fatalf("ToDynamicI64(%d) = %#v, want small i64", value, got)
		}
	}
}

func TestToDynamicI64ConcurrentCommonValues(t *testing.T) {
	values := []int64{commonI32BoxMin, -1, 0, 42, commonI32BoxMax}
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		for _, value := range values {
			wait.Add(1)
			go func() {
				defer wait.Done()
				for iteration := 0; iteration < 100; iteration++ {
					got := ToDynamicI64(value)
					integer, ok := got.(runtime.IntegerValue)
					if !ok || integer.Int64Fast() != value || integer.TypeSuffix != runtime.IntegerI64 {
						t.Errorf("ToDynamicI64(%d) = %#v, want concurrent small i64", value, got)
						return
					}
				}
			}()
		}
	}
	wait.Wait()
}
