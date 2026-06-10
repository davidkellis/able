package flow

import (
	"fmt"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
)

type sourceKey struct {
	startLine, startColumn uint32
	endLine, endColumn     uint32
	callsite               uint32
}

type callKey struct {
	kind                    semanticabi.CallTargetKind
	packageName, ownerType  uint32
	name, arity, returnType uint32
}

type tableBuilder struct {
	image           *semanticabi.Image
	symbols         map[string]uint32
	types           map[string]uint32
	sources         map[sourceKey]uint32
	constants       map[string]uint32
	capabilities    map[string]uint32
	callTargets     map[callKey]uint32
	sourcePathIndex uint32
}

func (tables *tableBuilder) init(image *semanticabi.Image, sourcePath string) {
	tables.image = image
	tables.symbols = make(map[string]uint32)
	tables.types = make(map[string]uint32)
	tables.sources = make(map[sourceKey]uint32)
	tables.constants = make(map[string]uint32)
	tables.capabilities = make(map[string]uint32)
	tables.callTargets = make(map[callKey]uint32)
	tables.sourcePathIndex = tables.symbol(sourcePath)
	tables.typeIndex("dynamic")
	tables.typeIndex("void")
	tables.typeIndex("bool")
}

func (tables *tableBuilder) symbol(value string) uint32 {
	if index, ok := tables.symbols[value]; ok {
		return index
	}
	index := uint32(len(tables.image.Strings))
	tables.image.Strings = append(tables.image.Strings, value)
	tables.symbols[value] = index
	return index
}

func (tables *tableBuilder) typeIndex(name string) uint32 {
	if name == "" {
		name = "dynamic"
	}
	if index, ok := tables.types[name]; ok {
		return index
	}
	index := uint32(len(tables.image.Types))
	tables.image.Types = append(tables.image.Types, semanticabi.TypeRecord{Name: tables.symbol(name)})
	tables.types[name] = index
	return index
}

func (tables *tableBuilder) source(node ast.Node, callsite uint32) (uint32, error) {
	span := node.Span()
	positions := []int{span.Start.Line, span.Start.Column, span.End.Line, span.End.Column}
	for _, position := range positions {
		if position < 0 {
			return 0, fmt.Errorf("semanticabi flow: negative source position for %s", node.NodeType())
		}
	}
	key := sourceKey{
		startLine: uint32(span.Start.Line), startColumn: uint32(span.Start.Column),
		endLine: uint32(span.End.Line), endColumn: uint32(span.End.Column), callsite: callsite,
	}
	if index, ok := tables.sources[key]; ok {
		return index, nil
	}
	index := uint32(len(tables.image.Sources))
	tables.image.Sources = append(tables.image.Sources, semanticabi.SourceRecord{
		File: tables.sourcePathIndex, StartLine: key.startLine, StartColumn: key.startColumn,
		EndLine: key.endLine, EndColumn: key.endColumn, Callsite: callsite,
	})
	tables.sources[key] = index
	return index, nil
}

func (tables *tableBuilder) constant(tag, aux uint32, data []byte) uint32 {
	key := fmt.Sprintf("%d:%d:%x", tag, aux, data)
	if index, ok := tables.constants[key]; ok {
		return index
	}
	index := uint32(len(tables.image.Constants))
	tables.image.Constants = append(tables.image.Constants, semanticabi.ConstantRecord{
		Tag: tag, Aux: aux, Data: append([]byte(nil), data...),
	})
	tables.constants[key] = index
	return index
}

func (tables *tableBuilder) capability(name string) uint32 {
	if index, ok := tables.capabilities[name]; ok {
		return index
	}
	index := uint32(len(tables.image.Capabilities))
	tables.image.Capabilities = append(tables.image.Capabilities, semanticabi.HostCapability{
		Name: tables.symbol(name), EffectKind: 1,
	})
	tables.capabilities[name] = index
	return index
}

func (tables *tableBuilder) callTarget(kind semanticabi.CallTargetKind, packageName string, ownerType uint32, name string, arity uint32, returnType uint32) uint32 {
	packageSymbol := semanticabi.NoIndex
	if packageName != "" {
		packageSymbol = tables.symbol(packageName)
	}
	key := callKey{kind: kind, packageName: packageSymbol, ownerType: ownerType, name: tables.symbol(name), arity: arity, returnType: returnType}
	if index, ok := tables.callTargets[key]; ok {
		return index
	}
	index := uint32(len(tables.image.CallTargets))
	tables.image.CallTargets = append(tables.image.CallTargets, semanticabi.CallTarget{
		Kind: kind, PackageName: packageSymbol, OwnerType: ownerType, Name: key.name,
		Arity: arity, ReturnType: returnType,
	})
	tables.callTargets[key] = index
	return index
}

func callsiteFor(node ast.Node, rootSource uint32) uint32 {
	if node.NodeType() == ast.NodeFunctionCall {
		return rootSource
	}
	return semanticabi.NoIndex
}
