package runtime

import (
	"fmt"
	goruntime "runtime"
	"unsafe"
)

// ArrayStoreLease identifies one handle-bearing runtime owner. The final
// release removes ArrayStore backing state and its metadata.
//
// Do not copy a live lease. Handle-bearing runtime values and generated Array
// carriers are pointer-owned, so their lease has a stable address.
type ArrayStoreLease struct {
	handle            int64
	generation        uint64
	tracked           bool
	cleanupRegistered bool
	cleanup           goruntime.Cleanup
}

// arrayStoreLeaseKey identifies an ownership token without retaining its
// wrapper. The generation keeps an arbitrarily delayed cleanup for an old
// wrapper from affecting a new wrapper at the same address.
type arrayStoreLeaseKey struct {
	address    uintptr
	generation uint64
}

// ArrayStoreLeaseStats is a copy of the current diagnostic lease ledger.
// OwnersByHandle may be inspected or modified by callers without affecting the
// registry.
type ArrayStoreLeaseStats struct {
	OwnerCount     int
	OwnersByHandle map[int64]int
}

var arrayStoreLeaseHandles map[arrayStoreLeaseKey]int64
var arrayStoreLeaseCounts map[int64]int
var arrayStoreNextLeaseGeneration uint64 = 1

func ensureArrayStoreLeaseLedger() {
	if arrayStoreLeaseHandles == nil {
		arrayStoreLeaseHandles = make(map[arrayStoreLeaseKey]int64)
	}
	if arrayStoreLeaseCounts == nil {
		arrayStoreLeaseCounts = make(map[int64]int)
	}
}

// ArrayStoreUpdateLease records that lease now owns handle. Passing zero
// releases the lease, stops any registered cleanup, and removes backing state
// when it was the final owner. This function is safe for concurrent ArrayStore
// users.
func ArrayStoreUpdateLease(lease *ArrayStoreLease, handle int64) error {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		arrayStoreStopLeaseCleanupLocked(lease)
	}
	return arrayStoreUpdateLeaseLocked(lease, handle)
}

// ArrayStoreMoveLease transfers source's handle token to target without an
// observable zero-owner interval. It is for explicit replacement moves of
// pointer-owned Array carriers; callers still copy the carrier data and then
// register target's cleanup through ArrayStoreTrackLeaseOwner.
func ArrayStoreMoveLease(target *ArrayStoreLease, source *ArrayStoreLease) error {
	if target == nil || source == nil {
		return fmt.Errorf("array lease is nil")
	}
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if target == source {
		return nil
	}
	arrayStoreStopLeaseCleanupLocked(target)
	if err := arrayStoreUpdateLeaseLocked(target, 0); err != nil {
		return err
	}
	if !source.tracked {
		return nil
	}
	sourceKey := arrayStoreLeaseKeyForLocked(source)
	handle := source.handle
	arrayStoreStopLeaseCleanupLocked(source)
	delete(arrayStoreLeaseHandles, sourceKey)
	source.handle = 0
	source.tracked = false
	source.generation = 0

	target.handle = handle
	target.tracked = true
	targetKey := arrayStoreLeaseKeyForLocked(target)
	arrayStoreLeaseHandles[targetKey] = handle
	return nil
}

// ArrayStoreTrackLeaseOwner records handle as owned by lease and registers a
// token-only cleanup for owner. It is for pointer-owned wrappers that embed an
// ArrayStoreLease, including generated compiler carriers. The cleanup never
// retains owner or removes backing state; it only removes the diagnostic
// ledger entry and removes backing state if that was the final owner.
func ArrayStoreTrackLeaseOwner[T any](owner *T, lease *ArrayStoreLease, handle int64) error {
	if owner == nil {
		return fmt.Errorf("array lease owner is nil")
	}
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	return arrayStoreTrackLeaseOwnerLocked(owner, lease, handle)
}

// ArrayStoreTrackArrayValueLease records the handle-bearing ownership of arr.
// Callers that create a view through ArrayStore factories receive this
// automatically; the function is exported for generated bridge boundaries.
func ArrayStoreTrackArrayValueLease(arr *ArrayValue, handle int64) error {
	if arr == nil {
		return fmt.Errorf("array lease owner is nil")
	}
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	return arrayStoreTrackArrayValueLeaseLocked(arr, handle)
}

func arrayStoreTrackArrayValueLeaseLocked(arr *ArrayValue, handle int64) error {
	if arr == nil {
		return fmt.Errorf("array lease owner is nil")
	}
	return arrayStoreTrackLeaseOwnerLocked(arr, &arr.Lease, handle)
}

func arrayStoreTrackLeaseOwnerLocked[T any](owner *T, lease *ArrayStoreLease, handle int64) error {
	if owner == nil {
		return fmt.Errorf("array lease owner is nil")
	}
	if handle == 0 {
		arrayStoreStopLeaseCleanupLocked(lease)
	}
	if err := arrayStoreUpdateLeaseLocked(lease, handle); err != nil {
		return err
	}
	arrayStoreRegisterLeaseCleanupLocked(owner, lease)
	goruntime.KeepAlive(owner)
	return nil
}

// ArrayStoreReleaseArrayValueLease releases arr's ownership of its backing
// state. The final lease removes that state and its metadata.
func ArrayStoreReleaseArrayValueLease(arr *ArrayValue) error {
	if arr == nil {
		return fmt.Errorf("array lease owner is nil")
	}
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	arrayStoreStopLeaseCleanupLocked(&arr.Lease)
	return arrayStoreUpdateLeaseLocked(&arr.Lease, 0)
}

// ArrayStoreLeaseTracks reports whether lease already records handle. It is a
// cheap owner-local check for paths that already synchronize access to the
// surrounding ArrayValue or generated carrier.
func ArrayStoreLeaseTracks(lease *ArrayStoreLease, handle int64) bool {
	return lease != nil && lease.tracked && lease.handle == handle
}

// ArrayStoreLeaseTracksWithCleanup reports whether lease already records
// handle and has the matching token-only cleanup registered. It permits an
// owner-local fast path while still allowing an older tracked lease to be
// upgraded to cleanup ownership.
func ArrayStoreLeaseTracksWithCleanup(lease *ArrayStoreLease, handle int64) bool {
	return ArrayStoreLeaseTracks(lease, handle) && lease.cleanupRegistered
}

// ArrayStoreLeaseStatsSnapshot returns a lock-consistent copy of current
// diagnostic owners. It never initializes or mutates ArrayStore state.
func ArrayStoreLeaseStatsSnapshot() ArrayStoreLeaseStats {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()

	owners := make(map[int64]int, len(arrayStoreLeaseCounts))
	ownerCount := 0
	for handle, count := range arrayStoreLeaseCounts {
		owners[handle] = count
		ownerCount += count
	}
	return ArrayStoreLeaseStats{
		OwnerCount:     ownerCount,
		OwnersByHandle: owners,
	}
}

func arrayStoreUpdateLeaseLocked(lease *ArrayStoreLease, handle int64) error {
	if lease == nil {
		return fmt.Errorf("array lease is nil")
	}
	if handle < 0 {
		return fmt.Errorf("array handle must be non-negative")
	}
	ensureArrayStoreLeaseLedger()
	if handle != 0 {
		if !arrayStoreHandleDefinedLocked(handle) {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
	}

	key := arrayStoreLeaseKeyForLocked(lease)
	previous := int64(0)
	if lease.tracked {
		previous = lease.handle
	}
	if previous == handle {
		return nil
	}
	if previous != 0 {
		delete(arrayStoreLeaseHandles, key)
		if arrayStoreDecrementLeaseCountLocked(previous) {
			arrayStoreRemoveBackingStateLocked(previous)
		}
		lease.generation = 0
	}
	lease.handle = handle
	lease.tracked = handle != 0
	if handle == 0 {
		return nil
	}
	key = arrayStoreLeaseKeyForLocked(lease)
	arrayStoreLeaseHandles[key] = handle
	arrayStoreLeaseCounts[handle]++
	return nil
}

func arrayStoreLeaseKeyForLocked(lease *ArrayStoreLease) arrayStoreLeaseKey {
	if lease.generation == 0 {
		lease.generation = arrayStoreNextLeaseGeneration
		arrayStoreNextLeaseGeneration++
		if arrayStoreNextLeaseGeneration == 0 {
			arrayStoreNextLeaseGeneration++
		}
	}
	return arrayStoreLeaseKey{
		address:    uintptr(unsafe.Pointer(lease)),
		generation: lease.generation,
	}
}

func arrayStoreRegisterLeaseCleanupLocked[T any](owner *T, lease *ArrayStoreLease) {
	if owner == nil || lease == nil || lease.cleanupRegistered || !lease.tracked {
		return
	}
	key := arrayStoreLeaseKeyForLocked(lease)
	lease.cleanup = goruntime.AddCleanup(owner, arrayStoreReleaseLeaseByCleanup, key)
	lease.cleanupRegistered = true
}

func arrayStoreStopLeaseCleanupLocked(lease *ArrayStoreLease) {
	if lease == nil || !lease.cleanupRegistered {
		return
	}
	lease.cleanup.Stop()
	lease.cleanupRegistered = false
}

// arrayStoreReleaseLeaseByCleanup is intentionally token-only: its argument
// cannot retain an ArrayValue or its embedded lease. Cleanups can execute long
// after an address is reused, so they remove a ledger entry only when both its
// address and generation still match.
func arrayStoreReleaseLeaseByCleanup(key arrayStoreLeaseKey) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	arrayStoreReleaseLeaseByKeyLocked(key)
}

func arrayStoreReleaseLeaseByKeyLocked(key arrayStoreLeaseKey) {
	handle, ok := arrayStoreLeaseHandles[key]
	if !ok {
		return
	}
	delete(arrayStoreLeaseHandles, key)
	if arrayStoreDecrementLeaseCountLocked(handle) {
		arrayStoreRemoveBackingStateLocked(handle)
	}
}

func arrayStoreDecrementLeaseCountLocked(handle int64) bool {
	if handle == 0 {
		return false
	}
	count := arrayStoreLeaseCounts[handle]
	if count <= 1 {
		delete(arrayStoreLeaseCounts, handle)
		return true
	}
	arrayStoreLeaseCounts[handle] = count - 1
	return false
}

// arrayStoreHandleDefinedLocked intentionally checks backing-state maps rather
// than only the hot kind cache. A future last-owner release will clear backing
// state before a cached handle kind can be observed again.
func arrayStoreHandleDefinedLocked(handle int64) bool {
	if handle == 0 {
		return false
	}
	if _, ok := arrayStates[handle]; ok {
		return true
	}
	if _, ok := monoArrayI32States[handle]; ok {
		return true
	}
	if _, ok := monoArrayI64States[handle]; ok {
		return true
	}
	if _, ok := monoArrayBoolStates[handle]; ok {
		return true
	}
	if _, ok := monoArrayCharStates[handle]; ok {
		return true
	}
	if _, ok := monoArrayU8States[handle]; ok {
		return true
	}
	if _, ok := monoArrayU32States[handle]; ok {
		return true
	}
	if _, ok := monoArrayU64States[handle]; ok {
		return true
	}
	if _, ok := monoArrayF64States[handle]; ok {
		return true
	}
	return false
}
