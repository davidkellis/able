package compiler

import "fmt"

func (g *generator) staticSliceLenExpr(sliceExpr string) string {
	return fmt.Sprintf("__able_slice_len(%s)", sliceExpr)
}

func (g *generator) staticSliceCapExpr(sliceExpr string) string {
	return fmt.Sprintf("__able_slice_cap(%s)", sliceExpr)
}

func (g *generator) staticStringLenBytesExpr(stringExpr string) string {
	return fmt.Sprintf("__able_string_len_bytes(%s)", stringExpr)
}

func (g *generator) staticArrayLengthExpr(arrayExpr string) string {
	return g.staticSliceLenExpr(fmt.Sprintf("%s.Elements", arrayExpr))
}

func (g *generator) staticArrayCapacityExpr(arrayExpr string) string {
	return g.staticSliceCapExpr(fmt.Sprintf("%s.Elements", arrayExpr))
}
