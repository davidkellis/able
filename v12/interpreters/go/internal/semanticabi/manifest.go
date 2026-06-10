package semanticabi

//go:generate go run ./cmd/manifestgen -runtime ../../pkg/runtime/values.go -out manifest_generated.go

type KindClass uint8

const (
	KindImmediate KindClass = iota + 1
	KindSharedHeap
	KindHostRegistry
)

type KindDescriptor struct {
	Name           string
	RuntimeOrdinal uint32
	Tag            uint32
	Class          KindClass
}

type LayoutMutability uint8

const (
	LayoutImmutable LayoutMutability = iota + 1
	LayoutMutable
)

type FieldStorage uint8

const (
	FieldScalar FieldStorage = iota + 1
	FieldBytes
	FieldCell
	FieldObject
	FieldCells
	FieldObjects
)

type LayoutFieldDescriptor struct {
	Name    string
	Storage FieldStorage
}

// ObjectLayoutDescriptor defines semantic ownership and tracing independently
// of the current Go runtime structs. RuntimeTag is zero for internal objects.
type ObjectLayoutDescriptor struct {
	Name       string
	LayoutID   uint16
	RuntimeTag uint32
	Mutability LayoutMutability
	Fields     []LayoutFieldDescriptor
}

type HostLayoutDescriptor struct {
	Name         string
	RuntimeTag   uint32
	Mutable      bool
	Cancelable   bool
	RetainsCells bool
}

type OperandKind uint8

const (
	OperandImmediate OperandKind = iota + 1
	OperandSymbol
	OperandType
	OperandConstant
	OperandBlock
	OperandCapability
	OperandRegister
	OperandCallTarget
)

type OpDescriptor struct {
	Name       string
	Opcode     uint16
	Operands   []OperandKind
	Variadic   OperandKind
	Writes     []uint8
	Terminator bool
}

func KindByTag(tag uint32) (KindDescriptor, bool) {
	if tag == 0 || int(tag) > len(KindManifest) {
		return KindDescriptor{}, false
	}
	descriptor := KindManifest[tag-1]
	return descriptor, descriptor.Tag == tag
}

func OpByCode(opcode uint16) (OpDescriptor, bool) {
	if opcode == 0 || int(opcode) > len(OpManifest) {
		return OpDescriptor{}, false
	}
	descriptor := OpManifest[opcode-1]
	return descriptor, descriptor.Opcode == opcode
}

func ObjectLayoutByID(id uint16) (ObjectLayoutDescriptor, bool) {
	if id == 0 || int(id) > len(ObjectLayoutManifest) {
		return ObjectLayoutDescriptor{}, false
	}
	descriptor := ObjectLayoutManifest[id-1]
	return descriptor, descriptor.LayoutID == id
}

func ObjectLayoutByTag(tag uint32) (ObjectLayoutDescriptor, bool) {
	for _, descriptor := range ObjectLayoutManifest {
		if descriptor.RuntimeTag == tag {
			return descriptor, true
		}
	}
	return ObjectLayoutDescriptor{}, false
}

func HostLayoutByTag(tag uint32) (HostLayoutDescriptor, bool) {
	for _, descriptor := range HostLayoutManifest {
		if descriptor.RuntimeTag == tag {
			return descriptor, true
		}
	}
	return HostLayoutDescriptor{}, false
}
