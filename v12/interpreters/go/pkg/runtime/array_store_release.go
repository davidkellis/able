package runtime

// arrayStoreRemoveBackingStateLocked deletes every backing representation and
// metadata entry for a handle after its final explicit lease is released.
// Callers hold arrayStoreMu exclusively.
func arrayStoreRemoveBackingStateLocked(handle int64) {
	if handle == 0 {
		return
	}
	delete(arrayStates, handle)
	delete(monoArrayI32States, handle)
	delete(monoArrayI64States, handle)
	delete(monoArrayBoolStates, handle)
	delete(monoArrayCharStates, handle)
	delete(monoArrayU8States, handle)
	delete(monoArrayU32States, handle)
	delete(monoArrayU64States, handle)
	delete(monoArrayF64States, handle)
	removeArrayHandleKind(handle)
	cacheArrayHandleRevision(handle, nil)
}
