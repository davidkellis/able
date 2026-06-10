package runtime

import (
	"sync"
	"sync/atomic"
)

// IteratorValue represents a lazily evaluated iterator produced by generator literals.
type IteratorValue struct {
	next      func() (Value, bool, error)
	nextRaw   func() (RawValue, bool, error)
	closer    func()
	retained  []Value
	closeOnce sync.Once
	closed    atomic.Bool
}

// IteratorHostDriver is the explicit host boundary for iterator behavior.
// Retained values enumerate semantic roots captured by the opaque callbacks.
type IteratorHostDriver struct {
	Next     func() (Value, bool, error)
	NextRaw  func() (RawValue, bool, error)
	Finalize func()
	Retained []Value
}

// NewIteratorValue constructs an iterator with the provided driver function.
func NewIteratorValue(step func() (Value, bool, error), finalize func()) *IteratorValue {
	if step == nil {
		step = func() (Value, bool, error) { return IteratorEnd, true, nil }
	}
	return &IteratorValue{next: step, closer: finalize}
}

func NewIteratorValueWithRaw(step func() (RawValue, bool, error), finalize func()) *IteratorValue {
	if step == nil {
		step = func() (RawValue, bool, error) {
			return NewRawValue(IteratorEnd), true, nil
		}
	}
	return &IteratorValue{nextRaw: step, closer: finalize}
}

func NewIteratorValueFromHostDriver(driver IteratorHostDriver, closed bool) *IteratorValue {
	value := &IteratorValue{}
	value.RestoreHostDriver(driver, closed)
	return value
}

// RestoreHostDriver initializes a newly constructed iterator wrapper from the
// explicit host boundary used by the shared semantic heap.
func (v *IteratorValue) RestoreHostDriver(driver IteratorHostDriver, closed bool) {
	if v == nil {
		return
	}
	v.next, v.nextRaw, v.closer = driver.Next, driver.NextRaw, driver.Finalize
	v.retained = append(v.retained[:0], driver.Retained...)
	if v.next == nil && v.nextRaw == nil {
		v.next = func() (Value, bool, error) { return IteratorEnd, true, nil }
	}
	v.closed.Store(closed)
}

func (v *IteratorValue) HostDriverSnapshot() (IteratorHostDriver, bool) {
	if v == nil {
		return IteratorHostDriver{}, true
	}
	driver := IteratorHostDriver{Next: v.next, NextRaw: v.nextRaw, Finalize: v.closer, Retained: append([]Value(nil), v.retained...)}
	return driver, v.closed.Load()
}

func (v *IteratorValue) Kind() Kind { return KindIterator }

// Next advances the iterator. The bool result reports whether iteration has completed.
func (v *IteratorValue) Next() (Value, bool, error) {
	raw, done, err := v.NextRaw()
	if err != nil {
		return nil, done, err
	}
	return raw.Materialize(), done, nil
}

func (v *IteratorValue) NextRaw() (RawValue, bool, error) {
	if v == nil {
		return NewRawValue(IteratorEnd), true, nil
	}
	if v.closed.Load() {
		return NewRawValue(IteratorEnd), true, nil
	}
	step := v.next
	rawStep := v.nextRaw
	if rawStep != nil {
		return rawStep()
	}
	if step == nil {
		return NewRawValue(IteratorEnd), true, nil
	}
	value, done, err := step()
	if err != nil {
		return RawValue{}, done, err
	}
	return NewRawValue(value), done, nil
}

// Close releases any resources held by the iterator.
func (v *IteratorValue) Close() {
	if v == nil {
		return
	}
	v.closeOnce.Do(func() {
		v.closed.Store(true)
		if v.closer != nil {
			v.closer()
		}
	})
}
