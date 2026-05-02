package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func qualifiedName(pkg string, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func runtimeLookupPackageName(pkg string) string {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return ""
	}
	parts := strings.Split(pkg, ".")
	if len(parts) >= 2 && parts[len(parts)-1] == parts[len(parts)-2] {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return pkg
}

func (c *compileContext) setReason(reason string) {
	if c == nil || reason == "" {
		return
	}
	if c.reason == "" {
		c.reason = reason
	}
	// Propagate reason to all ancestor contexts for diagnostics.
	for p := c.parent; p != nil; p = p.parent {
		if p.reason != "" {
			break
		}
		p.reason = reason
	}
}

func (c *compileContext) lookup(name string) (paramInfo, bool) {
	if c == nil {
		return paramInfo{}, false
	}
	if local, ok := c.locals[name]; ok {
		return local, true
	}
	if c.parent != nil {
		return c.parent.lookup(name)
	}
	if param, ok := c.params[name]; ok {
		return param, true
	}
	return paramInfo{}, false
}

func (c *compileContext) lookupCurrent(name string) (paramInfo, bool) {
	if c == nil {
		return paramInfo{}, false
	}
	if local, ok := c.locals[name]; ok {
		return local, true
	}
	if c.parent == nil {
		if param, ok := c.params[name]; ok {
			return param, true
		}
	}
	return paramInfo{}, false
}

func (c *compileContext) updateBinding(name string, info paramInfo) bool {
	if c == nil || name == "" {
		return false
	}
	if c.locals != nil {
		if _, ok := c.locals[name]; ok {
			c.locals[name] = info
			return true
		}
	}
	if c.parent != nil {
		if c.parent.updateBinding(name, info) {
			return true
		}
	}
	if c.params != nil {
		if _, ok := c.params[name]; ok {
			c.params[name] = info
			return true
		}
	}
	return false
}

func (c *compileContext) child() *compileContext {
	if c == nil {
		return nil
	}
	return &compileContext{
		function:               c.function,
		locals:                 make(map[string]paramInfo),
		integerFacts:           cloneIntegerFacts(c.integerFacts),
		functions:              c.functions,
		overloads:              c.overloads,
		packageName:            c.packageName,
		blockStatements:        c.blockStatements,
		statementIndex:         c.statementIndex,
		parent:                 c,
		temps:                  c.temps,
		loopDepth:              c.loopDepth,
		loopLabel:              c.loopLabel,
		loopBreakValueTemp:     c.loopBreakValueTemp,
		loopBreakValueType:     c.loopBreakValueType,
		loopBreakProbe:         c.loopBreakProbe,
		rethrowVar:             c.rethrowVar,
		rethrowErrVar:          c.rethrowErrVar,
		breakpoints:            c.breakpoints,
		breakpointGoLabels:     c.breakpointGoLabels,
		breakpointResultTemps:  c.breakpointResultTemps,
		breakpointResultTypes:  c.breakpointResultTypes,
		breakpointResultProbes: c.breakpointResultProbes,
		implicitReceiver:       c.implicitReceiver,
		hasImplicitReceiver:    c.hasImplicitReceiver,
		placeholderParams:      c.placeholderParams,
		inPlaceholder:          c.inPlaceholder,
		returnType:             c.returnType,
		returnTypeExpr:         c.returnTypeExpr,
		expectedTypeExpr:       c.expectedTypeExpr,
		matchSubjectTypeExpr:   c.matchSubjectTypeExpr,
		controlMode:            c.controlMode,
		controlCaptureVar:      c.controlCaptureVar,
		controlCaptureLabel:    c.controlCaptureLabel,
		controlCaptureBreak:    c.controlCaptureBreak,
		rethrowControlVar:      c.rethrowControlVar,
		genericNames:           c.genericNames,
		typeBindings:           c.typeBindings,
		implSiblings:           c.implSiblings,
		analysisOnly:           c.analysisOnly,
		closureScope:           c.closureScope,
	}
}

func (c *compileContext) closureChild() *compileContext {
	if c == nil {
		return nil
	}
	child := c.child()
	if child != nil {
		child.closureScope = true
	}
	return child
}

func (c *compileContext) probeChild() *compileContext {
	if c == nil {
		return nil
	}
	child := c.child()
	if child == nil || child.temps == nil {
		return child
	}
	temps := *child.temps
	child.temps = &temps
	return child
}

func (c *compileContext) substituteTypeBindings(expr ast.TypeExpression) ast.TypeExpression {
	if c == nil || len(c.typeBindings) == 0 || expr == nil {
		return expr
	}
	return substituteTypeParams(expr, c.typeBindings)
}

func (g *generator) typeExprContextInContext(ctx *compileContext, expr ast.TypeExpression) (string, ast.TypeExpression) {
	if g == nil || expr == nil {
		return "", expr
	}
	if ctx == nil {
		return "", expr
	}
	return g.normalizeTypeExprContextForPackage(ctx.packageName, ctx.substituteTypeBindings(expr))
}

func (g *generator) typeExprInContext(ctx *compileContext, expr ast.TypeExpression) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	if ctx == nil {
		return expr
	}
	_, normalized := g.typeExprContextInContext(ctx, expr)
	return normalized
}

func (g *generator) mapTypeExpressionInContext(ctx *compileContext, expr ast.TypeExpression) (string, bool) {
	if g == nil {
		return "", false
	}
	if ctx == nil {
		return g.mapTypeExpressionInPackage("", expr)
	}
	resolvedPkg, normalized := g.typeExprContextInContext(ctx, expr)
	return g.mapTypeExpressionInPackage(resolvedPkg, normalized)
}

func (c *compileContext) pushBreakpoint(label string) {
	if c == nil || label == "" {
		return
	}
	if c.breakpoints == nil {
		c.breakpoints = make(map[string]int)
	}
	c.breakpoints[label]++
}

func (c *compileContext) popBreakpoint(label string) {
	if c == nil || label == "" || c.breakpoints == nil {
		return
	}
	count := c.breakpoints[label]
	if count <= 1 {
		delete(c.breakpoints, label)
		return
	}
	c.breakpoints[label] = count - 1
}

func (c *compileContext) hasBreakpoint(label string) bool {
	if c == nil || label == "" || c.breakpoints == nil {
		return false
	}
	return c.breakpoints[label] > 0
}

func (c *compileContext) newTemp() string {
	if c == nil || c.temps == nil {
		return "__able_tmp"
	}
	for {
		name := fmt.Sprintf("__able_tmp_%d", *c.temps)
		*c.temps++
		if _, exists := c.lookup(name); !exists {
			return name
		}
	}
}
