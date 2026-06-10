package compiler

type monoArrayElemKind uint8

const (
	monoArrayElemKindUnknown monoArrayElemKind = iota
	monoArrayElemKindI8
	monoArrayElemKindI16
	monoArrayElemKindI32
	monoArrayElemKindI64
	monoArrayElemKindU16
	monoArrayElemKindU32
	monoArrayElemKindU64
	monoArrayElemKindISize
	monoArrayElemKindUSize
	monoArrayElemKindF32
	monoArrayElemKindF64
	monoArrayElemKindBool
	monoArrayElemKindU8
	monoArrayElemKindChar
	monoArrayElemKindString
)

func (g *generator) monoArraysEnabled() bool {
	// Static Array values are part of the compiler's native carrier contract.
	// The legacy ExperimentalMonoArrays options remain in the public Options
	// struct for source compatibility, but they may no longer select the
	// runtime-store representation for statically representable arrays.
	return g != nil
}
