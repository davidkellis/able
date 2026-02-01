package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

type generator struct {
	opts      Options
	structs   map[string]*structInfo
	functions map[string]*functionInfo
	warnings  []string
	mangler   *nameMangler
	needsAst  bool
}

type compileContext struct {
	params     map[string]paramInfo
	locals     map[string]paramInfo
	functions  map[string]*functionInfo
	parent     *compileContext
	temps      *int
	reason     string
	loopDepth  int
	rethrowVar string
}

func newGenerator(opts Options) *generator {
	return &generator{
		opts:      opts,
		structs:   make(map[string]*structInfo),
		functions: make(map[string]*functionInfo),
		mangler:   newNameMangler(),
	}
}

func (g *generator) collect(program *driver.Program) error {
	if program == nil || program.Entry == nil || program.Entry.AST == nil {
		return fmt.Errorf("compiler: missing entry module")
	}
	module := program.Entry.AST
	for _, stmt := range module.Body {
		def, ok := stmt.(*ast.StructDefinition)
		if !ok || def == nil || def.ID == nil {
			continue
		}
		name := def.ID.Name
		if name == "" {
			continue
		}
		if _, exists := g.structs[name]; exists {
			g.warnings = append(g.warnings, fmt.Sprintf("compiler: duplicate struct %s; skipping", name))
			continue
		}
		goName := g.mangler.unique(exportIdent(name))
		g.structs[name] = &structInfo{
			Name:   name,
			GoName: goName,
			Kind:   def.Kind,
			Node:   def,
		}
	}

	mapper := NewTypeMapper(g.structs)
	for _, info := range g.structs {
		fields := make([]fieldInfo, 0, len(info.Node.Fields))
		supported := info.Kind != ast.StructKindPositional
		for idx, field := range info.Node.Fields {
			fieldName := ""
			if field.Name != nil {
				fieldName = field.Name.Name
			} else {
				fieldName = fmt.Sprintf("field_%d", idx+1)
			}
			goFieldName := exportIdent(fieldName)
			goType, ok := mapper.Map(field.FieldType)
			if !ok {
				supported = false
			}
			fields = append(fields, fieldInfo{
				Name:      fieldName,
				GoName:    goFieldName,
				GoType:    goType,
				Supported: ok,
			})
		}
		info.Fields = fields
		info.Supported = supported
	}

	functions := make(map[string][]*ast.FunctionDefinition)
	for _, stmt := range module.Body {
		def, ok := stmt.(*ast.FunctionDefinition)
		if !ok || def == nil || def.ID == nil {
			continue
		}
		name := def.ID.Name
		if name == "" {
			continue
		}
		functions[name] = append(functions[name], def)
	}

	for name, defs := range functions {
		if len(defs) != 1 {
			g.warnings = append(g.warnings, fmt.Sprintf("compiler: function %s has %d overloads; leaving in interpreter", name, len(defs)))
			continue
		}
		info := &functionInfo{
			Name:       name,
			GoName:     g.mangler.unique("fn_" + sanitizeIdent(name)),
			Definition: defs[0],
		}
		g.fillFunctionInfo(info, mapper)
		g.functions[name] = info
	}
	g.resolveCompileableFunctions()
	g.detectAstNeeds()
	return nil
}

func (g *generator) fillFunctionInfo(info *functionInfo, mapper *TypeMapper) {
	if info == nil || info.Definition == nil {
		return
	}
	def := info.Definition
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	for idx, param := range def.Params {
		name := fmt.Sprintf("arg%d", idx)
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			supported = false
		}
		goName := safeParamName(name, idx)
		goType, ok := mapper.Map(param.ParamType)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    goName,
			GoType:    goType,
			Supported: ok,
		})
	}
	retType, ok := mapper.Map(def.ReturnType)
	if !ok || retType == "" {
		supported = false
	}
	info.Params = params
	info.ReturnType = retType
	info.SupportedTypes = supported
	info.Arity = len(params)

	if !supported {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
	}
}

func (g *generator) bodyCompileable(info *functionInfo, retType string) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	def := info.Definition
	if def.Body == nil {
		info.Reason = "missing function body"
		return false
	}
	ctx := newCompileContext(info, g.functions)
	_, _, ok := g.compileBody(ctx, info)
	if !ok {
		info.Reason = ctx.reason
		if info.Reason == "" {
			info.Reason = "unsupported function body"
		}
	}
	return ok
}

func (g *generator) resolveCompileableFunctions() {
	pending := make(map[string]*functionInfo)
	for name, info := range g.functions {
		if info == nil {
			continue
		}
		if !info.SupportedTypes {
			info.Compileable = false
			continue
		}
		pending[name] = info
	}
	for {
		progress := false
		for name, info := range pending {
			if info == nil {
				delete(pending, name)
				continue
			}
			if info.Compileable {
				delete(pending, name)
				continue
			}
			if ok := g.bodyCompileable(info, info.ReturnType); ok {
				info.Compileable = true
				info.Reason = ""
				progress = true
			}
		}
		if !progress {
			break
		}
	}
	for _, info := range pending {
		if info == nil {
			continue
		}
		if info.Reason == "" {
			info.Reason = "unsupported function body"
		}
		info.Compileable = false
	}
}

func (g *generator) compileBody(ctx *compileContext, info *functionInfo) ([]string, string, bool) {
	if info == nil || info.Definition == nil || info.Definition.Body == nil {
		ctx.setReason("missing function body")
		return nil, "", false
	}
	statements := info.Definition.Body.Body
	if len(statements) == 0 {
		if g.isVoidType(info.ReturnType) {
			return nil, "struct{}{}", true
		}
		ctx.setReason("empty body requires void return")
		return nil, "", false
	}
	lines := make([]string, 0, len(statements))
	for idx, stmt := range statements {
		isLast := idx == len(statements)-1
		if ret, ok := stmt.(*ast.ReturnStatement); ok {
			if !isLast {
				ctx.setReason("return must be final statement")
				return nil, "", false
			}
			return g.compileReturnStatement(ctx, info.ReturnType, ret, lines)
		}
		if isLast {
			if expr, ok := stmt.(ast.Expression); ok && expr != nil {
				return g.compileImplicitReturn(ctx, info.ReturnType, expr, lines)
			}
			if g.isVoidType(info.ReturnType) {
				stmtLines, ok := g.compileStatement(ctx, stmt)
				if !ok {
					return nil, "", false
				}
				lines = append(lines, stmtLines...)
				return lines, "struct{}{}", true
			}
			ctx.setReason("missing return expression")
			return nil, "", false
		}
		stmtLines, ok := g.compileStatement(ctx, stmt)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
	}
	ctx.setReason("missing return expression")
	return nil, "", false
}

func (g *generator) compileReturnStatement(ctx *compileContext, returnType string, ret *ast.ReturnStatement, lines []string) ([]string, string, bool) {
	if ret == nil {
		ctx.setReason("missing return")
		return nil, "", false
	}
	if ret.Argument == nil {
		if g.isVoidType(returnType) {
			return lines, "struct{}{}", true
		}
		ctx.setReason("missing return expression")
		return nil, "", false
	}
	if g.isVoidType(returnType) {
		stmtLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", ret.Argument)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
		}
		return lines, "struct{}{}", true
	}
	exprLines, expr, _, ok := g.compileTailExpression(ctx, returnType, ret.Argument)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, exprLines...)
	return lines, expr, true
}

func (g *generator) compileImplicitReturn(ctx *compileContext, returnType string, expr ast.Expression, lines []string) ([]string, string, bool) {
	if g.isVoidType(returnType) {
		stmtLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", expr)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
		}
		return lines, "struct{}{}", true
	}
	stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, returnType, expr)
	if !ok {
		return nil, "", false
	}
	if !g.typeMatches(returnType, valueType) {
		ctx.setReason("assignment return type mismatch")
		return nil, "", false
	}
	lines = append(lines, stmtLines...)
	return lines, valueExpr, true
}

func (g *generator) compileStatement(ctx *compileContext, stmt ast.Statement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing statement")
		return nil, false
	}
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
		lines, _, _, ok := g.compileAssignment(ctx, s)
		return lines, ok
	case *ast.WhileLoop:
		return g.compileWhileLoop(ctx, s)
	case *ast.ForLoop:
		return g.compileForLoop(ctx, s)
	case *ast.BreakStatement:
		return g.compileBreakStatement(ctx, s)
	case *ast.ContinueStatement:
		return g.compileContinueStatement(ctx, s)
	case *ast.RaiseStatement:
		return g.compileRaiseStatement(ctx, s)
	case *ast.RethrowStatement:
		return g.compileRethrowStatement(ctx, s)
	case *ast.IfExpression:
		return g.compileIfStatement(ctx, s)
	case *ast.BlockExpression:
		return g.compileBlockStatement(ctx, s)
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			valueLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", expr)
			if !ok {
				return nil, false
			}
			if valueExpr == "" {
				return valueLines, true
			}
			lines := append(valueLines, fmt.Sprintf("_ = %s", valueExpr))
			return lines, true
		}
		ctx.setReason("unsupported statement")
		return nil, false
	}
}

func newCompileContext(info *functionInfo, functions map[string]*functionInfo) *compileContext {
	counter := 0
	ctx := &compileContext{
		params:    make(map[string]paramInfo),
		locals:    make(map[string]paramInfo),
		functions: functions,
		temps:     &counter,
		loopDepth: 0,
	}
	if info != nil {
		for _, param := range info.Params {
			if param.Name == "" {
				continue
			}
			ctx.params[param.Name] = param
		}
	}
	return ctx
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
		locals:     make(map[string]paramInfo),
		functions:  c.functions,
		parent:     c,
		temps:      c.temps,
		loopDepth:  c.loopDepth,
		rethrowVar: c.rethrowVar,
	}
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
