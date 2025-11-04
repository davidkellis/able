package typechecker

// pushReturnType records the current function's expected return type.
func (c *Checker) pushReturnType(typ Type) {
	c.returnTypeStack = append(c.returnTypeStack, typ)
}

// popReturnType restores the previous expected return type.
func (c *Checker) popReturnType() {
	if len(c.returnTypeStack) == 0 {
		return
	}
	c.returnTypeStack = c.returnTypeStack[:len(c.returnTypeStack)-1]
}

// currentReturnType returns the innermost expected return type, if any.
func (c *Checker) currentReturnType() (Type, bool) {
	if len(c.returnTypeStack) == 0 {
		return nil, false
	}
	return c.returnTypeStack[len(c.returnTypeStack)-1], true
}

func (c *Checker) pushRescueContext() {
	c.rescueDepth++
}

func (c *Checker) popRescueContext() {
	if c.rescueDepth > 0 {
		c.rescueDepth--
	}
}

func (c *Checker) inRescueContext() bool {
	return c.rescueDepth > 0
}

func (c *Checker) pushLoopContext() {
	c.loopDepth++
	c.loopResultStack = append(c.loopResultStack, UnknownType{})
}

func (c *Checker) popLoopContext() Type {
	if c.loopDepth == 0 {
		return UnknownType{}
	}
	c.loopDepth--
	last := len(c.loopResultStack) - 1
	result := c.loopResultStack[last]
	c.loopResultStack = c.loopResultStack[:last]
	return result
}

func (c *Checker) inLoopContext() bool {
	return c.loopDepth > 0
}

func (c *Checker) recordBreakType(t Type) {
	if len(c.loopResultStack) == 0 {
		return
	}
	if t == nil {
		t = PrimitiveType{Kind: PrimitiveNil}
	}
	idx := len(c.loopResultStack) - 1
	c.loopResultStack[idx] = mergeCompatibleTypesSlice(c.loopResultStack[idx], t)
}

func (c *Checker) pushBreakpointLabel(label string) {
	if label == "" {
		return
	}
	c.breakpointStack = append(c.breakpointStack, label)
}

func (c *Checker) popBreakpointLabel() {
	if len(c.breakpointStack) == 0 {
		return
	}
	c.breakpointStack = c.breakpointStack[:len(c.breakpointStack)-1]
}

func (c *Checker) hasBreakpointLabel(label string) bool {
	for i := len(c.breakpointStack) - 1; i >= 0; i-- {
		if c.breakpointStack[i] == label {
			return true
		}
	}
	return false
}

func (c *Checker) pushAsyncContext() {
	c.asyncDepth++
}

func (c *Checker) popAsyncContext() {
	if c.asyncDepth > 0 {
		c.asyncDepth--
	}
}

func (c *Checker) inAsyncContext() bool {
	return c.asyncDepth > 0
}

func (c *Checker) pushConstraintScope(params []GenericParamSpec, where []WhereConstraintSpec) {
	scope := make(map[string][]Type)
	for _, param := range params {
		if param.Name == "" {
			continue
		}
		if len(param.Constraints) > 0 {
			copied := make([]Type, len(param.Constraints))
			copy(copied, param.Constraints)
			scope[param.Name] = append(scope[param.Name], copied...)
			continue
		}
		if _, ok := scope[param.Name]; !ok {
			scope[param.Name] = nil
		}
	}
	for _, clause := range where {
		if clause.TypeParam == "" {
			continue
		}
		if _, ok := scope[clause.TypeParam]; !ok {
			scope[clause.TypeParam] = nil
		}
		if len(clause.Constraints) > 0 {
			copied := make([]Type, len(clause.Constraints))
			copy(copied, clause.Constraints)
			scope[clause.TypeParam] = append(scope[clause.TypeParam], copied...)
		}
	}
	c.constraintStack = append(c.constraintStack, scope)
}

func (c *Checker) popConstraintScope() {
	if len(c.constraintStack) == 0 {
		return
	}
	c.constraintStack = c.constraintStack[:len(c.constraintStack)-1]
}

func (c *Checker) typeParamConstraints(name string) []Type {
	if name == "" {
		return nil
	}
	var out []Type
	for i := len(c.constraintStack) - 1; i >= 0; i-- {
		scope := c.constraintStack[i]
		if scope == nil {
			continue
		}
		if constraints, ok := scope[name]; ok {
			out = append(out, constraints...)
		}
	}
	return out
}

func (c *Checker) pushPipeContext() {
	c.pipeContextDepth++
}

func (c *Checker) popPipeContext() {
	if c.pipeContextDepth > 0 {
		c.pipeContextDepth--
	}
}

func (c *Checker) inPipeContext() bool {
	return c.pipeContextDepth > 0
}

func (c *Checker) typeParamInScope(name string) bool {
	if name == "" {
		return false
	}
	for i := len(c.constraintStack) - 1; i >= 0; i-- {
		scope := c.constraintStack[i]
		if scope == nil {
			continue
		}
		if _, ok := scope[name]; ok {
			return true
		}
	}
	return false
}
