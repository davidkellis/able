package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeIndexMethodCacheEntry struct {
	globalRevision      uint64
	receiverKind        bytecodeMemberReceiverKind
	arrayElemType       uint16
	receiverTypeKey     string
	receiverArrayHandle int64
	receiverArrayRev    uint64
	receiverArrayRevOK  bool
	receiverArrayCursor runtime.ArrayStoreRevisionCursor
	methodCacheVersion  uint64
	method              runtime.Value
	hasMethod           bool
	fastPath            bytecodeIndexMethodFastPathKind
}

type bytecodeIndexMethodCacheTable struct {
	get []bytecodeIndexMethodCacheEntry
	set []bytecodeIndexMethodCacheEntry
}

const bytecodeIndexMethodDirectEntries = 16

type bytecodeInlineIndexMethodCacheEntry struct {
	valid               bool
	program             *bytecodeProgram
	ip                  int
	methodKind          bytecodeIndexMethodCacheKind
	globalRevision      uint64
	receiverKind        bytecodeMemberReceiverKind
	arrayElemType       uint16
	receiverTypeKey     string
	receiverArrayHandle int64
	receiverArrayRev    uint64
	receiverArrayRevOK  bool
	receiverArrayCursor runtime.ArrayStoreRevisionCursor
	methodCacheVersion  uint64
	resolvedMethod      runtime.Value
	hasMethod           bool
	fastPath            bytecodeIndexMethodFastPathKind
}

type bytecodeIndexMethodCacheKind uint8

const (
	bytecodeIndexMethodCacheUnknown bytecodeIndexMethodCacheKind = iota
	bytecodeIndexMethodCacheGet
	bytecodeIndexMethodCacheSet
)

type bytecodeIndexMethodFastPathKind uint8

const (
	bytecodeIndexMethodFastPathNone bytecodeIndexMethodFastPathKind = iota
	bytecodeIndexMethodFastPathCanonicalArrayGet
	bytecodeIndexMethodFastPathCanonicalArraySet
)

const (
	bytecodeIndexTypeUnknown uint16 = iota
	bytecodeIndexTypeI8
	bytecodeIndexTypeI16
	bytecodeIndexTypeI32
	bytecodeIndexTypeI64
	bytecodeIndexTypeI128
	bytecodeIndexTypeU8
	bytecodeIndexTypeU16
	bytecodeIndexTypeU32
	bytecodeIndexTypeU64
	bytecodeIndexTypeU128
	bytecodeIndexTypeIsize
	bytecodeIndexTypeUsize
	bytecodeIndexTypeF32
	bytecodeIndexTypeF64
	bytecodeIndexTypeString
	bytecodeIndexTypeBool
	bytecodeIndexTypeChar
	bytecodeIndexTypeNil
	bytecodeIndexTypeVoid
)

func bytecodeIndexMethodCacheKindFor(methodName string) bytecodeIndexMethodCacheKind {
	switch methodName {
	case "get":
		return bytecodeIndexMethodCacheGet
	case "set":
		return bytecodeIndexMethodCacheSet
	default:
		return bytecodeIndexMethodCacheUnknown
	}
}

func bytecodeArrayReceiverForIndexCache(value runtime.Value) (*runtime.ArrayValue, bool) {
	switch v := value.(type) {
	case *runtime.ArrayValue:
		return v, v != nil
	case runtime.InterfaceValue:
		if arr, ok := v.Underlying.(*runtime.ArrayValue); ok && arr != nil {
			return arr, true
		}
	case *runtime.InterfaceValue:
		if v != nil {
			if arr, ok := v.Underlying.(*runtime.ArrayValue); ok && arr != nil {
				return arr, true
			}
		}
	}
	return nil, false
}
