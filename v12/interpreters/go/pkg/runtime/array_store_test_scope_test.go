package runtime

import (
	"sync"
	"testing"
)

// arrayStoreTestScope gives a runtime-package test a fresh process-wide
// registry and restores the previous registry on cleanup. Lease generations
// remain monotonic across scopes so delayed cleanup callbacks cannot collide
// with restored entries. Tests using it must not call t.Parallel because
// ArrayStore is deliberately shared by design.
type arrayStoreTestScope struct {
	once sync.Once

	arrayStates            map[int64]*ArrayState
	monoArrayI32States     map[int64]*monoArrayI32State
	monoArrayI64States     map[int64]*monoArrayI64State
	monoArrayBoolStates    map[int64]*monoArrayBoolState
	monoArrayCharStates    map[int64]*monoArrayCharState
	monoArrayU8States      map[int64]*monoArrayU8State
	monoArrayU32States     map[int64]*monoArrayU32State
	monoArrayU64States     map[int64]*monoArrayU64State
	monoArrayF64States     map[int64]*monoArrayF64State
	arrayHandleKinds       map[int64]monoArrayKind
	arrayHandleRevisions   map[int64]*uint64
	arrayStoreLeaseHandles map[arrayStoreLeaseKey]int64
	arrayStoreLeaseCounts  map[int64]int
	arrayNextHandle        int64
}

func newArrayStoreTestScope(t testing.TB) *arrayStoreTestScope {
	t.Helper()

	arrayStoreMu.Lock()
	scope := &arrayStoreTestScope{
		arrayStates:            arrayStates,
		monoArrayI32States:     monoArrayI32States,
		monoArrayI64States:     monoArrayI64States,
		monoArrayBoolStates:    monoArrayBoolStates,
		monoArrayCharStates:    monoArrayCharStates,
		monoArrayU8States:      monoArrayU8States,
		monoArrayU32States:     monoArrayU32States,
		monoArrayU64States:     monoArrayU64States,
		monoArrayF64States:     monoArrayF64States,
		arrayHandleKinds:       arrayHandleKinds,
		arrayHandleRevisions:   arrayHandleRevisions,
		arrayStoreLeaseHandles: arrayStoreLeaseHandles,
		arrayStoreLeaseCounts:  arrayStoreLeaseCounts,
		arrayNextHandle:        arrayNextHandle,
	}
	arrayStates = make(map[int64]*ArrayState)
	monoArrayI32States = make(map[int64]*monoArrayI32State)
	monoArrayI64States = make(map[int64]*monoArrayI64State)
	monoArrayBoolStates = make(map[int64]*monoArrayBoolState)
	monoArrayCharStates = make(map[int64]*monoArrayCharState)
	monoArrayU8States = make(map[int64]*monoArrayU8State)
	monoArrayU32States = make(map[int64]*monoArrayU32State)
	monoArrayU64States = make(map[int64]*monoArrayU64State)
	monoArrayF64States = make(map[int64]*monoArrayF64State)
	arrayHandleKinds = make(map[int64]monoArrayKind)
	arrayHandleRevisions = make(map[int64]*uint64)
	arrayStoreLeaseHandles = make(map[arrayStoreLeaseKey]int64)
	arrayStoreLeaseCounts = make(map[int64]int)
	arrayNextHandle = 1
	arrayStoreMu.Unlock()

	t.Cleanup(scope.Close)
	return scope
}

func (s *arrayStoreTestScope) Close() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		arrayStoreMu.Lock()
		defer arrayStoreMu.Unlock()
		arrayStates = s.arrayStates
		monoArrayI32States = s.monoArrayI32States
		monoArrayI64States = s.monoArrayI64States
		monoArrayBoolStates = s.monoArrayBoolStates
		monoArrayCharStates = s.monoArrayCharStates
		monoArrayU8States = s.monoArrayU8States
		monoArrayU32States = s.monoArrayU32States
		monoArrayU64States = s.monoArrayU64States
		monoArrayF64States = s.monoArrayF64States
		arrayHandleKinds = s.arrayHandleKinds
		arrayHandleRevisions = s.arrayHandleRevisions
		arrayStoreLeaseHandles = s.arrayStoreLeaseHandles
		arrayStoreLeaseCounts = s.arrayStoreLeaseCounts
		arrayNextHandle = s.arrayNextHandle
	})
}

func TestArrayStoreTestScopeRestoresPreviousRegistry(t *testing.T) {
	before := ArrayStoreStatsSnapshot()
	scope := newArrayStoreTestScope(t)
	ArrayStoreMonoNewWithCapacityU8(8)
	if isolated := ArrayStoreStatsSnapshot(); isolated.HandleCount != 1 || isolated.U8.StateCount != 1 {
		t.Fatalf("isolated registry stats = %#v, want one u8 handle", isolated)
	}

	scope.Close()
	after := ArrayStoreStatsSnapshot()
	if after != before {
		t.Fatalf("restored registry stats = %#v, want %#v", after, before)
	}
}
