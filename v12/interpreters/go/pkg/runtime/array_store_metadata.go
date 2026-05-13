package runtime

import "sync"

const (
	arrayMetadataBoxI32Max = 65536
	arrayMetadataBoxU64Max = 262144
)

var (
	arrayMetadataBoxI32Once sync.Once
	arrayMetadataBoxedI32   []Value
	arrayMetadataBoxU64Once sync.Once
	arrayMetadataBoxedU64   []Value
)

func initArrayMetadataI32BoxCache() {
	size := arrayMetadataBoxI32Max + 1
	arrayMetadataBoxedI32 = make([]Value, size)
	for idx := range arrayMetadataBoxedI32 {
		arrayMetadataBoxedI32[idx] = NewSmallInt(int64(idx), IntegerI32)
	}
}

func boxedArrayMetadataI32Value(value int) (Value, bool) {
	if value < 0 || value > arrayMetadataBoxI32Max {
		return nil, false
	}
	arrayMetadataBoxI32Once.Do(initArrayMetadataI32BoxCache)
	return arrayMetadataBoxedI32[value], true
}

func initArrayMetadataU64BoxCache() {
	size := arrayMetadataBoxU64Max + 1
	arrayMetadataBoxedU64 = make([]Value, size)
	for idx := range arrayMetadataBoxedU64 {
		arrayMetadataBoxedU64[idx] = NewSmallInt(int64(idx), IntegerU64)
	}
}

func BoxedArrayMetadataU64Value(value int64) (Value, bool) {
	if value < 0 || value > arrayMetadataBoxU64Max {
		return nil, false
	}
	arrayMetadataBoxU64Once.Do(initArrayMetadataU64BoxCache)
	return arrayMetadataBoxedU64[int(value)], true
}

func (s *ArrayState) BoxedLengthValue() Value {
	if s == nil {
		if boxed, ok := boxedArrayMetadataI32Value(0); ok {
			return boxed
		}
		return NewSmallInt(0, IntegerI32)
	}
	length := len(s.Values)
	if s.cachedLengthBox != nil && s.cachedLength == length {
		return s.cachedLengthBox
	}
	if boxed, ok := boxedArrayMetadataI32Value(length); ok {
		s.cachedLength = length
		s.cachedLengthBox = boxed
		return boxed
	}
	boxed := NewSmallInt(int64(length), IntegerI32)
	s.cachedLength = length
	s.cachedLengthBox = boxed
	return boxed
}

func (s *ArrayState) BoxedCapacityValue() Value {
	if s == nil {
		if boxed, ok := boxedArrayMetadataI32Value(0); ok {
			return boxed
		}
		return NewSmallInt(0, IntegerI32)
	}
	if s.cachedCapacityBox != nil && s.cachedCapacity == s.Capacity {
		return s.cachedCapacityBox
	}
	if boxed, ok := boxedArrayMetadataI32Value(s.Capacity); ok {
		s.cachedCapacity = s.Capacity
		s.cachedCapacityBox = boxed
		return boxed
	}
	boxed := NewSmallInt(int64(s.Capacity), IntegerI32)
	s.cachedCapacity = s.Capacity
	s.cachedCapacityBox = boxed
	return boxed
}
