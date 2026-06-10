package runtime

import "sync"

var monoArrayElementTypeNameHotHandle int64
var monoArrayElementTypeNameHot string
var monoArrayElementTypeNameHotOK bool
var monoArrayElementTypeNameMu sync.Mutex

func clearMonoArrayElementTypeNameHot(handle int64, kind monoArrayKind) {
	if handle == 0 || monoArrayElementTypeNameHotHandle != handle {
		return
	}
	monoArrayElementTypeNameHotHandle = 0
	monoArrayElementTypeNameHot = ""
	monoArrayElementTypeNameHotOK = false
}

func ArrayStoreMonoElementTypeNameIfKnown(handle int64) (string, bool, error) {
	if handle == 0 {
		return "", false, nil
	}
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	monoArrayElementTypeNameMu.Lock()
	defer monoArrayElementTypeNameMu.Unlock()
	if monoArrayElementTypeNameHotHandle == handle {
		return monoArrayElementTypeNameHot, monoArrayElementTypeNameHotOK, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return "", false, err
	}
	typeName, ok := monoArrayElementTypeName(kind)
	if ok {
		monoArrayElementTypeNameHotHandle = handle
		monoArrayElementTypeNameHot = typeName
		monoArrayElementTypeNameHotOK = true
	}
	return typeName, ok, nil
}

func monoArrayElementTypeName(kind monoArrayKind) (string, bool) {
	switch kind {
	case monoArrayKindI32:
		return string(IntegerI32), true
	case monoArrayKindI64:
		return string(IntegerI64), true
	case monoArrayKindBool:
		return "bool", true
	case monoArrayKindChar:
		return "char", true
	case monoArrayKindU8:
		return string(IntegerU8), true
	case monoArrayKindU32:
		return string(IntegerU32), true
	case monoArrayKindU64:
		return string(IntegerU64), true
	case monoArrayKindF64:
		return string(FloatF64), true
	default:
		return "", false
	}
}
