package compiler

// lowerControlTransfer is the canonical control-envelope transfer synthesis
// entrypoint.
func (g *generator) lowerControlTransfer(ctx *compileContext, controlExpr string) ([]string, bool) {
	return g.controlTransferLines(ctx, controlExpr)
}

// lowerControlCheck is the canonical control-envelope guard synthesis entrypoint.
func (g *generator) lowerControlCheck(ctx *compileContext, controlExpr string) ([]string, bool) {
	return g.controlCheckLines(ctx, controlExpr)
}
