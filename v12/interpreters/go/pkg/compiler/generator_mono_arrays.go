package compiler

type monoArrayElemKind uint8

const (
	monoArrayElemKindUnknown monoArrayElemKind = iota
	monoArrayElemKindI32
	monoArrayElemKindI64
	monoArrayElemKindBool
	monoArrayElemKindU8
)

func (g *generator) monoArraysEnabled() bool {
	return g != nil && g.opts.ExperimentalMonoArrays
}
