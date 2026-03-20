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
	return g != nil && g.opts.ExperimentalMonoArrays
}
