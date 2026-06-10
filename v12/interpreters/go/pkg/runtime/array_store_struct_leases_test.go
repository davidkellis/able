package runtime

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestArrayStoreStructInstanceLeaseTracksAliasTransferAndFuture(t *testing.T) {
	newArrayStoreTestScope(t)

	definition := &StructDefinitionValue{Node: ast.StructDef("Array", nil, ast.StructKindNamed, nil, nil, false)}
	instance := &StructInstanceValue{Definition: definition}
	first := ArrayStoreNew()
	second := ArrayStoreNew()

	if err := ArrayStoreTrackStructInstanceLease(instance, first); err != nil {
		t.Fatalf("track first struct handle: %v", err)
	}
	alias := instance
	if err := ArrayStoreTrackStructInstanceLease(alias, first); err != nil {
		t.Fatalf("track aliased struct handle: %v", err)
	}
	returned := passArrayStructLease(alias)
	if returned != instance {
		t.Fatalf("returned Array instance = %p, want %p", returned, instance)
	}
	if err := ArrayStoreTrackStructInstanceLease(returned, first); err != nil {
		t.Fatalf("track returned Array instance: %v", err)
	}
	future := NewFuture()
	future.Resolve(returned)
	result, errValue, status := future.Await()
	if errValue != nil || status != FutureResolved || result != instance {
		t.Fatalf("future result = (%#v, %#v, %v), want aliased Array instance", result, errValue, status)
	}
	if err := ArrayStoreTrackStructInstanceLease(result.(*StructInstanceValue), first); err != nil {
		t.Fatalf("track future Array result: %v", err)
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 1 || stats.OwnersByHandle[first] != 1 {
		t.Fatalf("aliased/future struct leases = %#v, want one owner", stats)
	}

	if err := ArrayStoreTrackStructInstanceLease(instance, second); err != nil {
		t.Fatalf("transfer struct lease: %v", err)
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 1 || stats.OwnersByHandle[first] != 0 || stats.OwnersByHandle[second] != 1 {
		t.Fatalf("transferred struct leases = %#v, want second handle owner", stats)
	}
	if err := ArrayStoreReleaseStructInstanceLease(instance); err != nil {
		t.Fatalf("release struct lease: %v", err)
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 0 {
		t.Fatalf("released struct lease = %#v, want empty ledger", stats)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 0 || state.HandleCount != 0 || state.RevisionCount != 0 {
		t.Fatalf("struct final release left backing state: %#v", state)
	}
}

func passArrayStructLease(value *StructInstanceValue) *StructInstanceValue {
	return value
}

func TestArrayStoreStructInstanceLeaseRejectsNonArray(t *testing.T) {
	newArrayStoreTestScope(t)

	definition := &StructDefinitionValue{Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)}
	instance := &StructInstanceValue{Definition: definition}
	handle := ArrayStoreNew()
	if err := ArrayStoreTrackStructInstanceLease(instance, handle); err == nil {
		t.Fatal("expected non-Array struct lease owner to be rejected")
	}
}

func TestArrayStoreStructInstanceLeaseRegistersTokenCleanup(t *testing.T) {
	newArrayStoreTestScope(t)

	definition := &StructDefinitionValue{Node: ast.StructDef("Array", nil, ast.StructKindNamed, nil, nil, false)}
	instance := &StructInstanceValue{Definition: definition}
	handle := ArrayStoreNew()
	if err := ArrayStoreTrackStructInstanceLease(instance, handle); err != nil {
		t.Fatalf("track Array struct lease: %v", err)
	}
	sidecar, ok := instance.Native.(*arrayStructInstanceLeaseSidecar)
	if !ok || sidecar == nil || !sidecar.lease.cleanupRegistered {
		t.Fatalf("Array struct sidecar cleanup = %#v, want registered cleanup", instance.Native)
	}

	releaseArrayStoreLeaseByCleanupForTest(&sidecar.lease)
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 0 || len(stats.OwnersByHandle) != 0 {
		t.Fatalf("Array struct cleanup leases = %#v, want empty ledger", stats)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 0 {
		t.Fatalf("Array struct final cleanup left backing state: %#v", state)
	}
}
