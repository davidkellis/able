package runtime

import (
	"fmt"
	"sync"
	"testing"
)

func TestArrayStoreConcurrentIndependentHandles(t *testing.T) {
	const (
		workers    = 12
		iterations = 96
	)

	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for index := 0; index < iterations; index++ {
				handle := ArrayStoreNewWithCapacity(1)
				value := NewSmallInt(int64(worker*iterations+index), IntegerI32)
				if err := ArrayStoreWrite(handle, 0, value); err != nil {
					errs <- fmt.Errorf("write handle %d: %w", handle, err)
					return
				}
				size, err := ArrayStoreSize(handle)
				if err != nil {
					errs <- fmt.Errorf("size handle %d: %w", handle, err)
					return
				}
				if size != 1 {
					errs <- fmt.Errorf("size handle %d = %d, want 1", handle, size)
					return
				}
				got, err := ArrayStoreRead(handle, 0)
				if err != nil {
					errs <- fmt.Errorf("read handle %d: %w", handle, err)
					return
				}
				integer, ok := got.(IntegerValue)
				if !ok || integer.BigInt().Cmp(value.BigInt()) != 0 {
					errs <- fmt.Errorf("read handle %d = %#v, want %v", handle, got, value)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestArrayStoreConcurrentIndependentMonoI32Handles(t *testing.T) {
	const (
		workers    = 12
		iterations = 96
	)

	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for index := 0; index < iterations; index++ {
				handle := ArrayStoreMonoNewWithCapacityI32(1)
				want := int32(worker*iterations + index)
				if err := ArrayStoreMonoWriteI32(handle, 0, want); err != nil {
					errs <- fmt.Errorf("write handle %d: %w", handle, err)
					return
				}
				var info ArrayStoreMonoPrimitiveReadInfo
				available, err := ArrayStoreMonoPrimitiveReadInfoInto(handle, 0, &info)
				if err != nil || !available || !info.InBounds || info.Int64 != int64(want) {
					errs <- fmt.Errorf("primitive read handle %d = %#v, available=%t, err=%v", handle, info, available, err)
					return
				}
				got, err := ArrayStoreMonoReadI32(handle, 0)
				if err != nil || got != want {
					errs <- fmt.Errorf("read handle %d = %d, want %d, err=%v", handle, got, want, err)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestArrayStoreConcurrentIndependentMonoPrimitiveHandles(t *testing.T) {
	const (
		workers    = 12
		iterations = 96
	)

	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for index := 0; index < iterations; index++ {
				seed := worker*iterations + index
				var (
					handle int64
					want   int64
					kind   ArrayStoreMonoPrimitiveReadKind
					err    error
				)
				switch index % 8 {
				case 0:
					handle, want, kind = ArrayStoreMonoNewI32(), int64(int32(seed)), ArrayStoreMonoPrimitiveReadI32
					err = ArrayStoreMonoWriteI32(handle, 0, int32(want))
				case 1:
					handle, want, kind = ArrayStoreMonoNewI64(), int64(seed), ArrayStoreMonoPrimitiveReadI64
					err = ArrayStoreMonoWriteI64(handle, 0, want)
				case 2:
					handle, kind = ArrayStoreMonoNewBool(), ArrayStoreMonoPrimitiveReadBool
					want = int64(seed & 1)
					err = ArrayStoreMonoWriteBool(handle, 0, want != 0)
				case 3:
					handle, want, kind = ArrayStoreMonoNewChar(), int64('a'+rune(seed%26)), ArrayStoreMonoPrimitiveReadChar
					err = ArrayStoreMonoWriteChar(handle, 0, rune(want))
				case 4:
					handle, want, kind = ArrayStoreMonoNewU8(), int64(uint8(seed)), ArrayStoreMonoPrimitiveReadU8
					err = ArrayStoreMonoWriteU8(handle, 0, uint8(want))
				case 5:
					handle, want, kind = ArrayStoreMonoNewU32(), int64(uint32(seed)), ArrayStoreMonoPrimitiveReadU32
					err = ArrayStoreMonoWriteU32(handle, 0, uint32(want))
				case 6:
					handle, want, kind = ArrayStoreMonoNewU64(), int64(seed), ArrayStoreMonoPrimitiveReadU64
					err = ArrayStoreMonoWriteU64(handle, 0, uint64(want))
				case 7:
					handle, want, kind = ArrayStoreMonoNewF64(), int64(seed), ArrayStoreMonoPrimitiveReadF64
					err = ArrayStoreMonoWriteF64(handle, 0, float64(want))
				}
				if err != nil {
					errs <- fmt.Errorf("write handle %d: %w", handle, err)
					return
				}
				var info ArrayStoreMonoPrimitiveReadInfo
				available, err := ArrayStoreMonoPrimitiveReadInfoInto(handle, 0, &info)
				if err != nil || !available || !info.InBounds || info.Kind != kind {
					errs <- fmt.Errorf("primitive read handle %d = %#v, available=%t, err=%v", handle, info, available, err)
					return
				}
				switch kind {
				case ArrayStoreMonoPrimitiveReadBool:
					if info.Bool != (want != 0) {
						errs <- fmt.Errorf("bool handle %d = %t, want %t", handle, info.Bool, want != 0)
						return
					}
				case ArrayStoreMonoPrimitiveReadF64:
					if info.Float64 != float64(want) {
						errs <- fmt.Errorf("f64 handle %d = %f, want %d", handle, info.Float64, want)
						return
					}
				case ArrayStoreMonoPrimitiveReadU8, ArrayStoreMonoPrimitiveReadU32, ArrayStoreMonoPrimitiveReadU64:
					if info.Uint64 != uint64(want) {
						errs <- fmt.Errorf("unsigned handle %d = %d, want %d", handle, info.Uint64, want)
						return
					}
				default:
					if info.Int64 != want {
						errs <- fmt.Errorf("primitive handle %d = %d, want %d", handle, info.Int64, want)
						return
					}
				}
				switch index % 8 {
				case 0:
					got, err := ArrayStoreMonoReadI32(handle, 0)
					if err != nil || int64(got) != want {
						errs <- fmt.Errorf("i32 handle %d = %d, want %d, err=%v", handle, got, want, err)
						return
					}
				case 1:
					got, err := ArrayStoreMonoReadI64(handle, 0)
					if err != nil || got != want {
						errs <- fmt.Errorf("i64 handle %d = %d, want %d, err=%v", handle, got, want, err)
						return
					}
				case 2:
					got, err := ArrayStoreMonoReadBool(handle, 0)
					if err != nil || got != (want != 0) {
						errs <- fmt.Errorf("bool handle %d = %t, want %t, err=%v", handle, got, want != 0, err)
						return
					}
				case 3:
					got, err := ArrayStoreMonoReadChar(handle, 0)
					if err != nil || int64(got) != want {
						errs <- fmt.Errorf("char handle %d = %d, want %d, err=%v", handle, got, want, err)
						return
					}
				case 4:
					got, err := ArrayStoreMonoReadU8(handle, 0)
					if err != nil || int64(got) != want {
						errs <- fmt.Errorf("u8 handle %d = %d, want %d, err=%v", handle, got, want, err)
						return
					}
				case 5:
					got, err := ArrayStoreMonoReadU32(handle, 0)
					if err != nil || int64(got) != want {
						errs <- fmt.Errorf("u32 handle %d = %d, want %d, err=%v", handle, got, want, err)
						return
					}
				case 6:
					got, err := ArrayStoreMonoReadU64(handle, 0)
					if err != nil || int64(got) != want {
						errs <- fmt.Errorf("u64 handle %d = %d, want %d, err=%v", handle, got, want, err)
						return
					}
				case 7:
					got, err := ArrayStoreMonoReadF64(handle, 0)
					if err != nil || got != float64(want) {
						errs <- fmt.Errorf("f64 handle %d = %f, want %d, err=%v", handle, got, want, err)
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
