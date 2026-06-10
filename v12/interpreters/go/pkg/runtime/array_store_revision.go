package runtime

type ArrayStoreRevisionCursor struct {
	handle   int64
	revision *uint64
}

func ArrayStoreRevisionCursorIfAvailable(handle int64) (ArrayStoreRevisionCursor, uint64, bool, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	arrayHandleRevisionMu.Lock()
	defer arrayHandleRevisionMu.Unlock()
	revision, ok, err := arrayHandleRevisionPointer(handle)
	if err != nil || !ok || revision == nil {
		return ArrayStoreRevisionCursor{}, 0, ok, err
	}
	return ArrayStoreRevisionCursor{handle: handle, revision: revision}, *revision, true, nil
}

func (cursor ArrayStoreRevisionCursor) Matches(handle int64, expected uint64) bool {
	return handle != 0 && handle == cursor.handle && cursor.revision != nil && *cursor.revision == expected
}

func (cursor ArrayStoreRevisionCursor) MatchesKnownHandle(handle int64, expected uint64) bool {
	return handle != 0 && cursor.handle == handle && cursor.revision != nil && *cursor.revision == expected
}

// ArrayStoreRevisionIfAvailable reports the current mutation revision for a
// handle-backed array without forcing typed arrays through dynamic materialization.
func ArrayStoreRevisionIfAvailable(handle int64) (uint64, bool, error) {
	if handle == 0 {
		return 0, false, nil
	}
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	arrayHandleRevisionMu.Lock()
	defer arrayHandleRevisionMu.Unlock()
	if arrayHandleRevisionHotHandle == handle && arrayHandleRevisionHot != nil {
		return *arrayHandleRevisionHot, true, nil
	}
	revision, ok, err := arrayHandleRevisionPointer(handle)
	if err != nil || !ok {
		return 0, ok, err
	}
	return *revision, true, nil
}

// ArrayStoreRevisionMatchesIfAvailable compares a handle-backed array's current
// mutation revision with expected without materializing typed arrays.
func ArrayStoreRevisionMatchesIfAvailable(handle int64, expected uint64) (bool, bool, error) {
	if handle == 0 {
		return false, false, nil
	}
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	arrayHandleRevisionMu.Lock()
	defer arrayHandleRevisionMu.Unlock()
	if arrayHandleRevisionHotHandle == handle && arrayHandleRevisionHot != nil {
		return *arrayHandleRevisionHot == expected, true, nil
	}
	revision, ok, err := arrayHandleRevisionPointer(handle)
	if err != nil || !ok {
		return false, ok, err
	}
	return *revision == expected, true, nil
}
