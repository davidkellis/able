package gobinding

import (
	"fmt"
	"reflect"
	"sort"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/heapmodel"
	"able/interpreter-go/pkg/runtime"
)

type Snapshot struct {
	Heap         *heapmodel.Heap
	Roots        []semanticabi.Cell
	metadata     []any
	hostMetadata map[semanticabi.ObjectID]any
}

type encoder struct {
	snapshot *Snapshot
	objects  map[pointerKey]semanticabi.ObjectID
	hosts    map[pointerKey]semanticabi.ObjectID
	envs     map[*runtime.Environment]semanticabi.ObjectID
}

type decoder struct {
	snapshot *Snapshot
	objects  map[semanticabi.ObjectID]any
	hosts    map[semanticabi.ObjectID]runtime.Value
	envs     map[semanticabi.ObjectID]*runtime.Environment
}

type pointerKey struct {
	typeName string
	pointer  uintptr
}

type UnsupportedError struct {
	Kind   string
	Reason string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("semanticabi Go binding: %s is not losslessly representable: %s", e.Kind, e.Reason)
}

func Encode(values ...runtime.Value) (*Snapshot, error) {
	snapshot := &Snapshot{Heap: heapmodel.New(), hostMetadata: make(map[semanticabi.ObjectID]any)}
	enc := &encoder{snapshot: snapshot, objects: make(map[pointerKey]semanticabi.ObjectID), hosts: make(map[pointerKey]semanticabi.ObjectID), envs: make(map[*runtime.Environment]semanticabi.ObjectID)}
	for _, value := range values {
		cell, err := enc.value(value)
		if err != nil {
			return nil, err
		}
		snapshot.Roots = append(snapshot.Roots, cell)
	}
	return snapshot, nil
}

func (s *Snapshot) Decode() ([]runtime.Value, error) {
	dec := &decoder{snapshot: s, objects: make(map[semanticabi.ObjectID]any), hosts: make(map[semanticabi.ObjectID]runtime.Value), envs: make(map[semanticabi.ObjectID]*runtime.Environment)}
	result := make([]runtime.Value, len(s.Roots))
	for index, cell := range s.Roots {
		value, err := dec.value(cell)
		if err != nil {
			return nil, err
		}
		result[index] = value
	}
	return result, nil
}

func RoundTrip(values ...runtime.Value) ([]runtime.Value, *Snapshot, error) {
	snapshot, err := Encode(values...)
	if err != nil {
		return nil, nil, err
	}
	decoded, err := snapshot.Decode()
	return decoded, snapshot, err
}

func (e *encoder) metadata(value any) uint64 {
	e.snapshot.metadata = append(e.snapshot.metadata, value)
	return uint64(len(e.snapshot.metadata))
}

func (d *decoder) metadata(index uint64) (any, error) {
	if index == 0 || index > uint64(len(d.snapshot.metadata)) {
		return nil, fmt.Errorf("semanticabi Go binding: metadata index %d is invalid", index)
	}
	return d.snapshot.metadata[index-1], nil
}

func keyOf(value any) (pointerKey, bool) {
	ref := reflect.ValueOf(value)
	if !ref.IsValid() || ref.Kind() != reflect.Pointer || ref.IsNil() {
		return pointerKey{}, false
	}
	return pointerKey{typeName: ref.Type().String(), pointer: ref.Pointer()}, true
}

func (e *encoder) reserve(value any, layoutID uint16, tag uint32) (semanticabi.ObjectID, semanticabi.Cell, bool, error) {
	key, keyed := keyOf(value)
	if keyed {
		if id, ok := e.objects[key]; ok {
			cell, err := e.snapshot.Heap.ProvisionalCell(tag, id)
			return id, cell, true, err
		}
	}
	id, err := e.snapshot.Heap.ReserveLayout(layoutID)
	if err != nil {
		return 0, semanticabi.Cell{}, false, err
	}
	if keyed {
		e.objects[key] = id
	}
	cell, err := e.snapshot.Heap.ProvisionalCell(tag, id)
	return id, cell, false, err
}

func fields(layoutID uint16) []heapmodel.FieldValue {
	result, err := heapmodel.NewFields(layoutID)
	if err != nil {
		panic(err)
	}
	return result
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
