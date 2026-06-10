package semanticabi

import "fmt"

// Cell is the cross-engine value representation. Its fields are deliberately
// limited to fixed-width scalars so it never contains a Go or native pointer.
type Cell struct {
	Tag     uint32
	Aux     uint32
	Payload uint64
}

// ObjectID is a generation-checked shared-heap or host-registry identity.
// Index and generation zero are both reserved for invalid/uninitialized data.
type ObjectID uint64

// CellAuxIndirect marks an otherwise-immediate value whose payload is an
// ObjectID for immutable scalar storage. Remaining Aux bits are kind-specific.
const CellAuxIndirect uint32 = 1 << 31

func NewObjectID(index, generation uint32) (ObjectID, error) {
	if index == 0 {
		return 0, fmt.Errorf("semanticabi: object index zero is invalid")
	}
	if generation == 0 {
		return 0, fmt.Errorf("semanticabi: object generation zero is invalid")
	}
	return ObjectID(uint64(generation)<<32 | uint64(index)), nil
}

func (id ObjectID) Index() uint32      { return uint32(id) }
func (id ObjectID) Generation() uint32 { return uint32(uint64(id) >> 32) }
func (id ObjectID) Valid() bool        { return id.Index() != 0 && id.Generation() != 0 }

func ImmediateCell(tag, aux uint32, payload uint64) (Cell, error) {
	descriptor, ok := KindByTag(tag)
	if !ok {
		return Cell{}, fmt.Errorf("semanticabi: unknown value tag %d", tag)
	}
	if descriptor.Class != KindImmediate {
		return Cell{}, fmt.Errorf("semanticabi: tag %s is not immediate", descriptor.Name)
	}
	return Cell{Tag: tag, Aux: aux, Payload: payload}, nil
}

func ReferenceCell(tag, aux uint32, id ObjectID) (Cell, error) {
	descriptor, ok := KindByTag(tag)
	if !ok {
		return Cell{}, fmt.Errorf("semanticabi: unknown value tag %d", tag)
	}
	if descriptor.Class == KindImmediate {
		return Cell{}, fmt.Errorf("semanticabi: tag %s is not reference-backed", descriptor.Name)
	}
	if !id.Valid() {
		return Cell{}, fmt.Errorf("semanticabi: invalid reference identity")
	}
	return Cell{Tag: tag, Aux: aux, Payload: uint64(id)}, nil
}

func IndirectImmediateCell(tag, format uint32, id ObjectID) (Cell, error) {
	descriptor, ok := KindByTag(tag)
	if !ok {
		return Cell{}, fmt.Errorf("semanticabi: unknown value tag %d", tag)
	}
	if descriptor.Class != KindImmediate {
		return Cell{}, fmt.Errorf("semanticabi: tag %s is not immediate", descriptor.Name)
	}
	if format&CellAuxIndirect != 0 {
		return Cell{}, fmt.Errorf("semanticabi: indirect format uses reserved bit")
	}
	if !id.Valid() {
		return Cell{}, fmt.Errorf("semanticabi: invalid indirect identity")
	}
	return Cell{Tag: tag, Aux: CellAuxIndirect | format, Payload: uint64(id)}, nil
}

func (cell Cell) IsIndirectImmediate() bool { return cell.Aux&CellAuxIndirect != 0 }
