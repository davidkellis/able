package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileLocalMethodsDefinitionStatement(ctx *compileContext, def *ast.MethodsDefinition) ([]string, bool) {
	if def == nil || def.TargetType == nil || len(def.Definitions) == 0 {
		ctx.setReason("unsupported local methods definition")
		return nil, false
	}
	defExpr, ok := g.renderMethodsDefinitionExpr(def)
	if !ok {
		ctx.setReason("unsupported local methods definition")
		return nil, false
	}
	envName, lines := localDefinitionEnvSetup(ctx)
	errName := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("if _, %s := bridge.RegisterMethodsDefinition(__able_runtime, %s, %s); %s != nil { panic(%s) }", errName, defExpr, envName, errName, errName))
	return lines, true
}

func (g *generator) compileLocalImplementationDefinitionStatement(ctx *compileContext, def *ast.ImplementationDefinition) ([]string, bool) {
	if def == nil || def.InterfaceName == nil || strings.TrimSpace(def.InterfaceName.Name) == "" || def.TargetType == nil {
		ctx.setReason("unsupported local implementation definition")
		return nil, false
	}
	defExpr, ok := g.renderImplementationDefinitionExpr(def)
	if !ok {
		ctx.setReason("unsupported local implementation definition")
		return nil, false
	}
	envName, lines := localDefinitionEnvSetup(ctx)
	errName := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("if _, %s := bridge.RegisterImplementationDefinition(__able_runtime, %s, %s); %s != nil { panic(%s) }", errName, defExpr, envName, errName, errName))
	return lines, true
}
