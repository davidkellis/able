package heapmodel

import (
	"reflect"
	"strings"
	"testing"

	"able/interpreter-go/internal/semanticabi"
)

func TestConformanceVectorsAreDeterministic(t *testing.T) {
	first, err := RunConformanceVectors()
	if err != nil {
		t.Fatal(err)
	}
	second, err := RunConformanceVectors()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("conformance reports differ:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if first.SharedHeapLayouts != 20 || first.InternalHeapLayouts != 7 || first.HostLayouts != 4 || first.DeterministicVectorCount != 8 {
		t.Fatalf("conformance coverage = %#v", first)
	}
}

func TestCollectionRejectsStaleTracedEdgeWithoutSweeping(t *testing.T) {
	h := New()
	child, err := allocate(h, semanticabi.LayoutString, nil)
	if err != nil {
		t.Fatal(err)
	}
	childCell, err := h.Cell(semanticabi.TagKindString, child)
	if err != nil {
		t.Fatal(err)
	}
	holder, err := allocate(h, semanticabi.LayoutArray, func(fields []FieldValue) error {
		return SetCells(semanticabi.LayoutArray, fields, "elements", childCell)
	})
	if err != nil {
		t.Fatal(err)
	}
	holderCell, _ := h.Cell(semanticabi.TagKindArray, holder)
	childFrame := h.OpenRootFrame(childCell)
	if _, err := h.Collect(); err != nil {
		t.Fatal(err)
	}
	if err := h.CloseRootFrame(childFrame); err != nil {
		t.Fatal(err)
	}
	// Collect both, reuse the child's lowest slot, then deliberately retain the
	// old cell inside a newly rooted holder to prove generation checking.
	if _, err := h.Collect(); err != nil {
		t.Fatal(err)
	}
	_, err = allocate(h, semanticabi.LayoutString, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = allocate(h, semanticabi.LayoutString, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Resolve(holder); err == nil {
		// holder was expected to have collected; holderCell still carries its old ID.
		t.Fatal("holder unexpectedly remained live")
	}
	frame := h.OpenRootFrame(holderCell)
	_, err = h.Collect()
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("collection error = %v, want stale", err)
	}
	_ = h.CloseRootFrame(frame)
}

func TestHostRegistryChecksTagsMutabilityAndRelease(t *testing.T) {
	h := New()
	id, err := h.Hosts().Register(semanticabi.TagKindNativeFunction)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Hosts().Cell(semanticabi.TagKindFuture, id); err == nil {
		t.Fatal("host registry accepted mismatched tag")
	}
	if err := h.Hosts().Update(id); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("native function update error = %v", err)
	}
	if err := h.Hosts().Cancel(id); err == nil || !strings.Contains(err.Error(), "not cancelable") {
		t.Fatalf("native function cancel error = %v", err)
	}
	if err := h.Hosts().Release(id); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Hosts().Resolve(id); err == nil || !strings.Contains(err.Error(), "released") {
		t.Fatalf("released host resolution error = %v", err)
	}
}

func TestReservedIdentitySupportsCyclesButCannotResolveBeforeInitialize(t *testing.T) {
	h := New()
	id, err := h.ReserveLayout(semanticabi.LayoutArray)
	if err != nil {
		t.Fatal(err)
	}
	cell, err := h.ProvisionalCell(semanticabi.TagKindArray, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Resolve(id); err == nil || !strings.Contains(err.Error(), "uninitialized") {
		t.Fatalf("provisional resolve error = %v", err)
	}
	frame := h.OpenRootFrame(cell)
	if _, err := h.Collect(); err == nil || !strings.Contains(err.Error(), "uninitialized") {
		t.Fatalf("provisional collect error = %v", err)
	}
	fields, _ := NewFields(semanticabi.LayoutArray)
	if err := SetCells(semanticabi.LayoutArray, fields, "elements", cell); err != nil {
		t.Fatal(err)
	}
	if err := h.Initialize(id, fields); err != nil {
		t.Fatal(err)
	}
	if err := h.Initialize(id, fields); err == nil || !strings.Contains(err.Error(), "already initialized") {
		t.Fatalf("second initialize error = %v", err)
	}
	collection, err := h.Collect()
	if err != nil || collection.Reachable != 1 {
		t.Fatalf("initialized cycle collection = %+v: %v", collection, err)
	}
	_ = h.CloseRootFrame(frame)
}

func TestIndirectImmediateOwnsAndTracesWideScalar(t *testing.T) {
	h := New()
	fields, err := NewFields(semanticabi.LayoutWideScalar)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetScalar(semanticabi.LayoutWideScalar, fields, "format", 1); err != nil {
		t.Fatal(err)
	}
	if err := SetBytes(semanticabi.LayoutWideScalar, fields, "payload", []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	id, err := h.AllocateLayout(semanticabi.LayoutWideScalar, fields)
	if err != nil {
		t.Fatal(err)
	}
	cell, err := semanticabi.IndirectImmediateCell(semanticabi.TagKindInteger, 7, id)
	if err != nil {
		t.Fatal(err)
	}
	frame := h.OpenRootFrame(cell)
	collection, err := h.Collect()
	if err != nil || collection.Reachable != 1 || collection.Collected != 0 {
		t.Fatalf("rooted indirect collection = %+v: %v", collection, err)
	}
	if err := h.CloseRootFrame(frame); err != nil {
		t.Fatal(err)
	}
	collection, err = h.Collect()
	if err != nil || collection.Collected != 1 {
		t.Fatalf("unrooted indirect collection = %+v: %v", collection, err)
	}
}

func TestIndirectImmediateRejectsNonScalarBackingLayout(t *testing.T) {
	h := New()
	id, err := allocate(h, semanticabi.LayoutString, nil)
	if err != nil {
		t.Fatal(err)
	}
	cell, err := semanticabi.IndirectImmediateCell(semanticabi.TagKindInteger, 0, id)
	if err != nil {
		t.Fatal(err)
	}
	frame := h.OpenRootFrame(cell)
	defer func() { _ = h.CloseRootFrame(frame) }()
	if _, err := h.Collect(); err == nil || !strings.Contains(err.Error(), "disagrees with object layout") {
		t.Fatalf("wrong indirect backing error = %v", err)
	}
}
