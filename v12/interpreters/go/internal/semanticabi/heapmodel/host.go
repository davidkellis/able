package heapmodel

import (
	"fmt"
	"sort"

	"able/interpreter-go/internal/semanticabi"
)

type HostObject struct {
	Tag      uint32
	Retained []semanticabi.Cell
	Revision uint64
	Canceled bool
}

type hostSlot struct {
	generation uint32
	object     *HostObject
}

type HostRegistry struct {
	slots []hostSlot
	free  []uint32
}

func (r *HostRegistry) Register(tag uint32, retained ...semanticabi.Cell) (semanticabi.ObjectID, error) {
	descriptor, ok := semanticabi.HostLayoutByTag(tag)
	if !ok {
		return 0, fmt.Errorf("heapmodel: tag %d has no host layout", tag)
	}
	if len(retained) != 0 && !descriptor.RetainsCells {
		return 0, fmt.Errorf("heapmodel: host layout %s cannot retain cells", descriptor.Name)
	}
	index, generation, err := r.reserve()
	if err != nil {
		return 0, err
	}
	r.slots[index].object = &HostObject{Tag: tag, Retained: append([]semanticabi.Cell(nil), retained...)}
	return semanticabi.NewObjectID(index, generation)
}

func (r *HostRegistry) Cell(tag uint32, id semanticabi.ObjectID) (semanticabi.Cell, error) {
	object, err := r.resolve(id)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	if object.Tag != tag {
		return semanticabi.Cell{}, fmt.Errorf("heapmodel: host tag %d cannot reference host tag %d", tag, object.Tag)
	}
	return semanticabi.ReferenceCell(tag, 0, id)
}

func (r *HostRegistry) Update(id semanticabi.ObjectID, retained ...semanticabi.Cell) error {
	object, err := r.resolve(id)
	if err != nil {
		return err
	}
	descriptor, _ := semanticabi.HostLayoutByTag(object.Tag)
	if !descriptor.Mutable {
		return fmt.Errorf("heapmodel: host layout %s is immutable", descriptor.Name)
	}
	object.Retained = append([]semanticabi.Cell(nil), retained...)
	object.Revision++
	return nil
}

func (r *HostRegistry) Cancel(id semanticabi.ObjectID) error {
	object, err := r.resolve(id)
	if err != nil {
		return err
	}
	descriptor, _ := semanticabi.HostLayoutByTag(object.Tag)
	if !descriptor.Cancelable {
		return fmt.Errorf("heapmodel: host layout %s is not cancelable", descriptor.Name)
	}
	object.Canceled = true
	object.Revision++
	return nil
}

func (r *HostRegistry) Release(id semanticabi.ObjectID) error {
	if _, err := r.resolve(id); err != nil {
		return err
	}
	r.slots[id.Index()].object = nil
	r.free = insertSorted(r.free, id.Index())
	return nil
}

func (r *HostRegistry) Resolve(id semanticabi.ObjectID) (HostObject, error) {
	object, err := r.resolve(id)
	if err != nil {
		return HostObject{}, err
	}
	result := *object
	result.Retained = append([]semanticabi.Cell(nil), object.Retained...)
	return result, nil
}

func (r *HostRegistry) reserve() (uint32, uint32, error) {
	if len(r.free) == 0 {
		index := uint32(len(r.slots))
		r.slots = append(r.slots, hostSlot{generation: 1})
		return index, 1, nil
	}
	index := r.free[0]
	r.free = r.free[1:]
	s := &r.slots[index]
	if s.generation == ^uint32(0) {
		return 0, 0, fmt.Errorf("heapmodel: host generation exhausted for slot %d", index)
	}
	s.generation++
	return index, s.generation, nil
}

func (r *HostRegistry) resolve(id semanticabi.ObjectID) (*HostObject, error) {
	if !id.Valid() || int(id.Index()) >= len(r.slots) {
		return nil, fmt.Errorf("heapmodel: invalid host id %d:%d", id.Index(), id.Generation())
	}
	s := &r.slots[id.Index()]
	if s.generation != id.Generation() {
		return nil, fmt.Errorf("heapmodel: stale host id %d:%d (current generation %d)", id.Index(), id.Generation(), s.generation)
	}
	if s.object == nil {
		return nil, fmt.Errorf("heapmodel: released host id %d:%d", id.Index(), id.Generation())
	}
	return s.object, nil
}

func (r *HostRegistry) liveCount() int {
	live := 0
	for index := 1; index < len(r.slots); index++ {
		if r.slots[index].object != nil {
			live++
		}
	}
	return live
}

func insertSorted(values []uint32, value uint32) []uint32 {
	index := sort.Search(len(values), func(index int) bool { return values[index] >= value })
	values = append(values, 0)
	copy(values[index+1:], values[index:])
	values[index] = value
	return values
}
