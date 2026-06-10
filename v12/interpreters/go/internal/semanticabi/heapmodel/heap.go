package heapmodel

import (
	"fmt"
	"sort"

	"able/interpreter-go/internal/semanticabi"
)

type FieldValue struct {
	Scalar  uint64
	Bytes   []byte
	Cells   []semanticabi.Cell
	Objects []semanticabi.ObjectID
}

type Object struct {
	LayoutID uint16
	Fields   []FieldValue
	Revision uint64
	Ready    bool
}

type slot struct {
	generation uint32
	object     *Object
}

type Heap struct {
	slots      []slot
	free       []uint32
	rootFrames map[uint64][]semanticabi.Cell
	nextRoot   uint64
	hosts      HostRegistry
}

type Stats struct {
	LiveObjects int `json:"live_objects"`
	FreeSlots   int `json:"free_slots"`
	LiveHosts   int `json:"live_hosts"`
}

type Collection struct {
	Before    int `json:"before"`
	Reachable int `json:"reachable"`
	Collected int `json:"collected"`
}

func New() *Heap {
	h := &Heap{slots: make([]slot, 1), rootFrames: make(map[uint64][]semanticabi.Cell)}
	h.hosts.slots = make([]hostSlot, 1)
	return h
}

func (h *Heap) AllocateTag(tag uint32, fields []FieldValue) (semanticabi.ObjectID, error) {
	layout, ok := semanticabi.ObjectLayoutByTag(tag)
	if !ok {
		return 0, fmt.Errorf("heapmodel: tag %d has no shared-heap layout", tag)
	}
	return h.AllocateLayout(layout.LayoutID, fields)
}

func (h *Heap) AllocateLayout(layoutID uint16, fields []FieldValue) (semanticabi.ObjectID, error) {
	id, err := h.ReserveLayout(layoutID)
	if err != nil {
		return 0, err
	}
	if err := h.Initialize(id, fields); err != nil {
		return 0, err
	}
	return id, nil
}

// ReserveLayout publishes an identity before its fields are encoded. This is
// required for arbitrary cyclic graphs; the object cannot resolve or trace
// until Initialize completes exactly once.
func (h *Heap) ReserveLayout(layoutID uint16) (semanticabi.ObjectID, error) {
	if _, ok := semanticabi.ObjectLayoutByID(layoutID); !ok {
		return 0, fmt.Errorf("heapmodel: unknown object layout %d", layoutID)
	}
	index, generation, err := h.reserveSlot()
	if err != nil {
		return 0, err
	}
	h.slots[index].object = &Object{LayoutID: layoutID}
	return semanticabi.NewObjectID(index, generation)
}

func (h *Heap) Initialize(id semanticabi.ObjectID, fields []FieldValue) error {
	s, err := h.resolveSlotRaw(id)
	if err != nil {
		return err
	}
	if s.object.Ready {
		return fmt.Errorf("heapmodel: object id %d:%d is already initialized", id.Index(), id.Generation())
	}
	layout, _ := semanticabi.ObjectLayoutByID(s.object.LayoutID)
	if err := validateFields(layout, fields); err != nil {
		return err
	}
	s.object.Fields = cloneFields(fields)
	s.object.Ready = true
	return nil
}

func (h *Heap) reserveSlot() (uint32, uint32, error) {
	if len(h.free) == 0 {
		index := uint32(len(h.slots))
		h.slots = append(h.slots, slot{generation: 1})
		return index, 1, nil
	}
	index := h.free[0]
	h.free = h.free[1:]
	s := &h.slots[index]
	if s.generation == ^uint32(0) {
		return 0, 0, fmt.Errorf("heapmodel: generation exhausted for slot %d", index)
	}
	s.generation++
	return index, s.generation, nil
}

func (h *Heap) Resolve(id semanticabi.ObjectID) (Object, error) {
	s, err := h.resolveSlot(id)
	if err != nil {
		return Object{}, err
	}
	return cloneObject(*s.object), nil
}

func (h *Heap) Mutate(id semanticabi.ObjectID, fields []FieldValue) error {
	s, err := h.resolveSlot(id)
	if err != nil {
		return err
	}
	layout, _ := semanticabi.ObjectLayoutByID(s.object.LayoutID)
	if layout.Mutability != semanticabi.LayoutMutable {
		return fmt.Errorf("heapmodel: layout %s is immutable", layout.Name)
	}
	if err := validateFields(layout, fields); err != nil {
		return err
	}
	s.object.Fields = cloneFields(fields)
	s.object.Revision++
	return nil
}

func (h *Heap) Cell(tag uint32, id semanticabi.ObjectID) (semanticabi.Cell, error) {
	layout, ok := semanticabi.ObjectLayoutByTag(tag)
	if !ok {
		return semanticabi.Cell{}, fmt.Errorf("heapmodel: tag %d has no shared-heap layout", tag)
	}
	object, err := h.Resolve(id)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	if object.LayoutID != layout.LayoutID {
		actual, _ := semanticabi.ObjectLayoutByID(object.LayoutID)
		return semanticabi.Cell{}, fmt.Errorf("heapmodel: tag %s cannot reference layout %s", layout.Name, actual.Name)
	}
	return semanticabi.ReferenceCell(tag, 0, id)
}

func (h *Heap) resolveSlot(id semanticabi.ObjectID) (*slot, error) {
	s, err := h.resolveSlotRaw(id)
	if err != nil {
		return nil, err
	}
	if !s.object.Ready {
		return nil, fmt.Errorf("heapmodel: uninitialized object id %d:%d", id.Index(), id.Generation())
	}
	return s, nil
}

func (h *Heap) resolveSlotRaw(id semanticabi.ObjectID) (*slot, error) {
	if !id.Valid() || int(id.Index()) >= len(h.slots) {
		return nil, fmt.Errorf("heapmodel: invalid object id %d:%d", id.Index(), id.Generation())
	}
	s := &h.slots[id.Index()]
	if s.generation != id.Generation() {
		return nil, fmt.Errorf("heapmodel: stale object id %d:%d (current generation %d)", id.Index(), id.Generation(), s.generation)
	}
	if s.object == nil {
		return nil, fmt.Errorf("heapmodel: released object id %d:%d", id.Index(), id.Generation())
	}
	return s, nil
}

func (h *Heap) ProvisionalCell(tag uint32, id semanticabi.ObjectID) (semanticabi.Cell, error) {
	layout, ok := semanticabi.ObjectLayoutByTag(tag)
	if !ok {
		return semanticabi.Cell{}, fmt.Errorf("heapmodel: tag %d has no shared-heap layout", tag)
	}
	s, err := h.resolveSlotRaw(id)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	if s.object.LayoutID != layout.LayoutID {
		return semanticabi.Cell{}, fmt.Errorf("heapmodel: provisional tag %s disagrees with layout %d", layout.Name, s.object.LayoutID)
	}
	return semanticabi.ReferenceCell(tag, 0, id)
}

func (h *Heap) OpenRootFrame(cells ...semanticabi.Cell) uint64 {
	h.nextRoot++
	h.rootFrames[h.nextRoot] = append([]semanticabi.Cell(nil), cells...)
	return h.nextRoot
}

func (h *Heap) ReplaceRootFrame(frame uint64, cells ...semanticabi.Cell) error {
	if _, ok := h.rootFrames[frame]; !ok {
		return fmt.Errorf("heapmodel: unknown root frame %d", frame)
	}
	h.rootFrames[frame] = append([]semanticabi.Cell(nil), cells...)
	return nil
}

func (h *Heap) CloseRootFrame(frame uint64) error {
	if _, ok := h.rootFrames[frame]; !ok {
		return fmt.Errorf("heapmodel: unknown root frame %d", frame)
	}
	delete(h.rootFrames, frame)
	return nil
}

func (h *Heap) Hosts() *HostRegistry { return &h.hosts }

func (h *Heap) Stats() Stats {
	live := 0
	for index := 1; index < len(h.slots); index++ {
		if h.slots[index].object != nil {
			live++
		}
	}
	return Stats{LiveObjects: live, FreeSlots: len(h.free), LiveHosts: h.hosts.liveCount()}
}

func validateFields(layout semanticabi.ObjectLayoutDescriptor, fields []FieldValue) error {
	if len(fields) != len(layout.Fields) {
		return fmt.Errorf("heapmodel: layout %s has %d fields, got %d", layout.Name, len(layout.Fields), len(fields))
	}
	for index, descriptor := range layout.Fields {
		field := fields[index]
		switch descriptor.Storage {
		case semanticabi.FieldScalar:
			if len(field.Bytes)+len(field.Cells)+len(field.Objects) != 0 {
				return badField(layout, descriptor)
			}
		case semanticabi.FieldBytes:
			if len(field.Cells)+len(field.Objects) != 0 {
				return badField(layout, descriptor)
			}
		case semanticabi.FieldCell:
			if len(field.Cells) != 1 || len(field.Bytes)+len(field.Objects) != 0 {
				return badField(layout, descriptor)
			}
		case semanticabi.FieldObject:
			if len(field.Objects) != 1 || len(field.Bytes)+len(field.Cells) != 0 {
				return badField(layout, descriptor)
			}
		case semanticabi.FieldCells:
			if len(field.Bytes)+len(field.Objects) != 0 {
				return badField(layout, descriptor)
			}
		case semanticabi.FieldObjects:
			if len(field.Bytes)+len(field.Cells) != 0 {
				return badField(layout, descriptor)
			}
		default:
			return fmt.Errorf("heapmodel: layout %s field %s has unknown storage %d", layout.Name, descriptor.Name, descriptor.Storage)
		}
	}
	return nil
}

func badField(layout semanticabi.ObjectLayoutDescriptor, field semanticabi.LayoutFieldDescriptor) error {
	return fmt.Errorf("heapmodel: layout %s field %s does not match storage %d", layout.Name, field.Name, field.Storage)
}

func cloneFields(fields []FieldValue) []FieldValue {
	result := make([]FieldValue, len(fields))
	for index, field := range fields {
		result[index] = FieldValue{Scalar: field.Scalar, Bytes: append([]byte(nil), field.Bytes...), Cells: append([]semanticabi.Cell(nil), field.Cells...), Objects: append([]semanticabi.ObjectID(nil), field.Objects...)}
	}
	return result
}

func cloneObject(object Object) Object {
	object.Fields = cloneFields(object.Fields)
	return object
}

func insertFree(values []uint32, value uint32) []uint32 {
	index := sort.Search(len(values), func(index int) bool { return values[index] >= value })
	values = append(values, 0)
	copy(values[index+1:], values[index:])
	values[index] = value
	return values
}
