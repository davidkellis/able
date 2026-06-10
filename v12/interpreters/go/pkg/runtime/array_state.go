package runtime

type ArrayState struct {
	Values                []Value
	Capacity              int
	Revision              uint64
	ValuesMaterialized    bool
	ElementTypeToken      uint16
	ElementTypeTokenKnown bool
	CachedI32Values       []int32
	CachedI32ValuesValid  []bool
	CachedI32ValuesCount  int
	CachedI32ValuesKnown  bool
	cachedLength          int
	cachedLengthBox       Value
	cachedCapacity        int
	cachedCapacityBox     Value
}

type monoArrayKind uint8

const (
	monoArrayKindDynamic monoArrayKind = iota
	monoArrayKindI32
	monoArrayKindI64
	monoArrayKindBool
	monoArrayKindChar
	monoArrayKindU8
	monoArrayKindU32
	monoArrayKindU64
	monoArrayKindF64
)

type monoArrayState[T any] struct {
	Values   []T
	Capacity int
	Revision uint64
}

type monoArrayI32State = monoArrayState[int32]
type monoArrayI64State = monoArrayState[int64]
type monoArrayBoolState = monoArrayState[bool]
type monoArrayCharState = monoArrayState[rune]
type monoArrayU8State = monoArrayState[uint8]
type monoArrayU32State = monoArrayState[uint32]
type monoArrayU64State = monoArrayState[uint64]
type monoArrayF64State = monoArrayState[float64]
