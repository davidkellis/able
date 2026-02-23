package compiler

import "fmt"

func qualifiedName(pkg string, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func (c *compileContext) setReason(reason string) {
	if c == nil || reason == "" {
		return
	}
	if c.reason == "" {
		c.reason = reason
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

func (c *compileContext) child() *compileContext {
	if c == nil {
		return nil
	}
	return &compileContext{
		locals:              make(map[string]paramInfo),
		functions:           c.functions,
		overloads:           c.overloads,
		packageName:         c.packageName,
		parent:              c,
		temps:               c.temps,
		loopDepth:           c.loopDepth,
		rethrowVar:          c.rethrowVar,
		rethrowErrVar:       c.rethrowErrVar,
		breakpoints:         c.breakpoints,
		implicitReceiver:    c.implicitReceiver,
		hasImplicitReceiver: c.hasImplicitReceiver,
		placeholderParams:   c.placeholderParams,
		inPlaceholder:       c.inPlaceholder,
		returnType:          c.returnType,
		returnTypeExpr:      c.returnTypeExpr,
		genericNames:        c.genericNames,
		implSiblings:        c.implSiblings,
	}
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
