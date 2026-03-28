package compiler

func (ctx *compileContext) setNominalCoercionOrigin(temp string, origin nominalCoercionOrigin) {
	if ctx == nil || temp == "" || origin.Expr == "" || origin.GoType == "" {
		return
	}
	if ctx.coercedNominalOrigins == nil {
		ctx.coercedNominalOrigins = make(map[string]nominalCoercionOrigin)
	}
	ctx.coercedNominalOrigins[temp] = origin
}

func (ctx *compileContext) nominalCoercionOrigin(temp string) (nominalCoercionOrigin, bool) {
	if ctx == nil || temp == "" || ctx.coercedNominalOrigins == nil {
		return nominalCoercionOrigin{}, false
	}
	origin, ok := ctx.coercedNominalOrigins[temp]
	return origin, ok
}
