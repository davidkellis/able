package semanticabi

import "math"

const (
	FormatVersion   uint16 = 2
	SemanticVersion uint32 = 2
	SectionCount           = 8
	NoIndex         uint32 = math.MaxUint32
)

const (
	FunctionFlagShadowEligible uint32 = 1 << iota
	FunctionFlagFlowValidated
)

type CallTargetKind uint32

const (
	CallTargetLocal CallTargetKind = iota + 1
	CallTargetImported
	CallTargetMember
	CallTargetBuiltin
)

type Header struct {
	SemanticVersion uint32
	Identity        [32]byte
	ProgramID       uint64
}

func CurrentHeader(programID uint64) Header {
	return Header{
		SemanticVersion: SemanticVersion,
		Identity:        ManifestIdentity,
		ProgramID:       programID,
	}
}

type TypeRecord struct {
	Name  uint32
	Flags uint32
}

type SourceRecord struct {
	File        uint32
	StartLine   uint32
	StartColumn uint32
	EndLine     uint32
	EndColumn   uint32
	Callsite    uint32
}

type ConstantRecord struct {
	Tag  uint32
	Aux  uint32
	Data []byte
}

type PackageRecord struct {
	Name          uint32
	FunctionStart uint32
	FunctionCount uint32
}

type FunctionRecord struct {
	Name             uint32
	Package          uint32
	Source           uint32
	ParameterCount   uint32
	RegisterStart    uint32
	RegisterCount    uint32
	BlockCount       uint32
	InstructionStart uint32
	InstructionCount uint32
	Flags            uint32
}

type CallTarget struct {
	Kind        CallTargetKind
	PackageName uint32
	OwnerType   uint32
	Name        uint32
	Arity       uint32
	ReturnType  uint32
	Flags       uint32
}

type Instruction struct {
	Opcode   uint16
	Flags    uint16
	Source   uint32
	Block    uint32
	Operands []uint32
}

type HostCapability struct {
	Name       uint32
	EffectKind uint32
	Flags      uint32
}

type Image struct {
	Header        Header
	Strings       []string
	Types         []TypeRecord
	Sources       []SourceRecord
	Constants     []ConstantRecord
	Packages      []PackageRecord
	Functions     []FunctionRecord
	RegisterTypes []uint32
	CallTargets   []CallTarget
	Instructions  []Instruction
	Capabilities  []HostCapability
}
