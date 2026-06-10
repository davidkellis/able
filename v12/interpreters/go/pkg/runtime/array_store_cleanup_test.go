package runtime

// releaseArrayValueLeaseByCleanupForTest deterministically exercises the same
// token-only callback used by runtime.AddCleanup. It avoids GC scheduling, and
// stops the registered cleanup before simulating an unreachable owner.
func releaseArrayValueLeaseByCleanupForTest(arr *ArrayValue) {
	if arr == nil {
		return
	}
	releaseArrayStoreLeaseByCleanupForTest(&arr.Lease)
}

// releaseArrayStoreLeaseByCleanupForTest deterministically exercises the same
// token-only callback used by runtime.AddCleanup for any lease-owning wrapper.
// It avoids GC scheduling, and stops the registered cleanup before simulating
// an unreachable owner.
func releaseArrayStoreLeaseByCleanupForTest(lease *ArrayStoreLease) {
	if lease == nil {
		return
	}
	arrayStoreMu.Lock()
	key := arrayStoreLeaseKeyForLocked(lease)
	arrayStoreStopLeaseCleanupLocked(lease)
	lease.handle = 0
	lease.tracked = false
	arrayStoreMu.Unlock()

	arrayStoreReleaseLeaseByCleanup(key)
}
