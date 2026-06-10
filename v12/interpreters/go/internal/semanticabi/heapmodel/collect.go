package heapmodel

import (
	"fmt"
	"sort"

	"able/interpreter-go/internal/semanticabi"
)

func (h *Heap) Collect() (Collection, error) {
	before := h.Stats().LiveObjects
	marked := make(map[uint32]bool, before)
	frames := make([]uint64, 0, len(h.rootFrames))
	for frame := range h.rootFrames {
		frames = append(frames, frame)
	}
	sort.Slice(frames, func(i, j int) bool { return frames[i] < frames[j] })
	for _, frame := range frames {
		for _, cell := range h.rootFrames[frame] {
			if err := h.traceCell(cell, marked); err != nil {
				return Collection{}, fmt.Errorf("heapmodel: root frame %d: %w", frame, err)
			}
		}
	}
	for index := 1; index < len(h.hosts.slots); index++ {
		host := h.hosts.slots[index].object
		if host == nil {
			continue
		}
		for _, cell := range host.Retained {
			if err := h.traceCell(cell, marked); err != nil {
				return Collection{}, fmt.Errorf("heapmodel: host slot %d: %w", index, err)
			}
		}
	}
	collected := 0
	for index := 1; index < len(h.slots); index++ {
		if h.slots[index].object == nil || marked[uint32(index)] {
			continue
		}
		h.slots[index].object = nil
		h.free = insertFree(h.free, uint32(index))
		collected++
	}
	return Collection{Before: before, Reachable: len(marked), Collected: collected}, nil
}

func (h *Heap) traceCell(cell semanticabi.Cell, marked map[uint32]bool) error {
	descriptor, ok := semanticabi.KindByTag(cell.Tag)
	if !ok {
		return fmt.Errorf("unknown cell tag %d", cell.Tag)
	}
	if descriptor.Class == semanticabi.KindImmediate && cell.IsIndirectImmediate() {
		id := semanticabi.ObjectID(cell.Payload)
		object, err := h.resolveSlot(id)
		if err != nil {
			return err
		}
		if object.object.LayoutID != semanticabi.LayoutWideScalar {
			return fmt.Errorf("indirect immediate %s disagrees with object layout %d", descriptor.Name, object.object.LayoutID)
		}
		return h.traceObject(id, marked)
	}
	if descriptor.Class == semanticabi.KindImmediate {
		return nil
	}
	id := semanticabi.ObjectID(cell.Payload)
	if descriptor.Class == semanticabi.KindHostRegistry {
		host, err := h.hosts.resolve(id)
		if err != nil {
			return err
		}
		if host.Tag != cell.Tag {
			return fmt.Errorf("host cell tag %s disagrees with registry tag %d", descriptor.Name, host.Tag)
		}
		return nil
	}
	layout, ok := semanticabi.ObjectLayoutByTag(cell.Tag)
	if !ok {
		return fmt.Errorf("shared cell tag %s has no layout", descriptor.Name)
	}
	object, err := h.resolveSlot(id)
	if err != nil {
		return err
	}
	if object.object.LayoutID != layout.LayoutID {
		return fmt.Errorf("cell tag %s disagrees with object layout %d", descriptor.Name, object.object.LayoutID)
	}
	return h.traceObject(id, marked)
}

func (h *Heap) traceObject(id semanticabi.ObjectID, marked map[uint32]bool) error {
	if id == 0 {
		return nil
	}
	s, err := h.resolveSlot(id)
	if err != nil {
		return err
	}
	if marked[id.Index()] {
		return nil
	}
	marked[id.Index()] = true
	layout, _ := semanticabi.ObjectLayoutByID(s.object.LayoutID)
	for index, descriptor := range layout.Fields {
		field := s.object.Fields[index]
		switch descriptor.Storage {
		case semanticabi.FieldCell, semanticabi.FieldCells:
			for _, cell := range field.Cells {
				if err := h.traceCell(cell, marked); err != nil {
					return fmt.Errorf("layout %s field %s: %w", layout.Name, descriptor.Name, err)
				}
			}
		case semanticabi.FieldObject, semanticabi.FieldObjects:
			for _, child := range field.Objects {
				if err := h.traceObject(child, marked); err != nil {
					return fmt.Errorf("layout %s field %s: %w", layout.Name, descriptor.Name, err)
				}
			}
		}
	}
	return nil
}
