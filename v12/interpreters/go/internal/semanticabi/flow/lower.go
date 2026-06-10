package flow

import (
	"fmt"
	"strings"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
)

type Coverage struct {
	Function     string
	ASTNodes     int
	Instructions int
	Registers    int
	Blocks       int
	CallTargets  int
	HostEffects  []string
	ASTFallbacks int
	Unsupported  []string
}

type Options struct {
	// HostFunctions contains resolved fully-qualified extern/native function
	// names and their return types. It is semantic input, not name inference.
	HostFunctions map[string]string
}

func (coverage Coverage) Complete() bool {
	return coverage.ASTNodes > 0 && coverage.Instructions > 0 && coverage.Registers > 0 &&
		coverage.Blocks > 1 && coverage.ASTFallbacks == 0 && len(coverage.Unsupported) == 0
}

type block struct {
	id           uint32
	instructions []semanticabi.Instruction
	terminated   bool
}

type binding struct {
	register uint32
	typeID   uint32
}

type loopTarget struct {
	continueBlock  uint32
	breakBlock     uint32
	resultRegister uint32
}

type lowerer struct {
	module        *ast.Module
	function      *ast.FunctionDefinition
	functionName  string
	packageName   string
	image         *semanticabi.Image
	blocks        []*block
	current       uint32
	scopes        []map[string]binding
	loops         []loopTarget
	registerTypes []uint32
	tables        tableBuilder
	imports       map[string]string
	localReturns  map[string]string
	globalTypes   map[string]string
	hostFunctions map[string]string
	rootSource    uint32
	coverage      Coverage
}

func LowerFunction(module *ast.Module, functionName, sourcePath string, programID uint64) (*semanticabi.Image, Coverage, error) {
	return LowerFunctionWithOptions(module, functionName, sourcePath, programID, Options{})
}

func LowerFunctionWithOptions(module *ast.Module, functionName, sourcePath string, programID uint64, options Options) (*semanticabi.Image, Coverage, error) {
	if module == nil {
		return nil, Coverage{}, fmt.Errorf("semanticabi flow: nil module")
	}
	if functionName == "" || sourcePath == "" || programID == 0 {
		return nil, Coverage{}, fmt.Errorf("semanticabi flow: function, source path, and nonzero program id are required")
	}
	function := findFunction(module, functionName)
	if function == nil {
		return nil, Coverage{}, fmt.Errorf("semanticabi flow: function %q not found", functionName)
	}
	lower := &lowerer{
		module:        module,
		function:      function,
		functionName:  functionName,
		packageName:   modulePackageName(module),
		image:         &semanticabi.Image{Header: semanticabi.CurrentHeader(programID)},
		imports:       collectImports(module),
		localReturns:  collectFunctionReturns(module),
		globalTypes:   collectGlobalTypes(module),
		hostFunctions: copyStringMap(options.HostFunctions),
		coverage:      Coverage{Function: functionName, HostEffects: make([]string, 0)},
	}
	lower.tables.init(lower.image, sourcePath)
	if err := lower.auditASTCoverage(function); err != nil {
		return nil, lower.coverage, err
	}
	rootSource, err := lower.tables.source(function, semanticabi.NoIndex)
	if err != nil {
		return nil, lower.coverage, err
	}
	lower.rootSource = rootSource
	lower.newBlock()
	lower.pushScope()
	for _, parameter := range function.Params {
		name, ok := patternBindingName(parameter.Name)
		if !ok {
			return nil, lower.coverage, fmt.Errorf("semanticabi flow: unsupported parameter pattern %s", parameter.Name.NodeType())
		}
		typeID := lower.tables.typeIndex(typeNameOrDynamic(parameter.ParamType))
		register := lower.newRegister(typeID)
		lower.bind(name, binding{register: register, typeID: typeID})
	}
	result, hasResult, err := lower.lowerBlock(function.Body)
	if err != nil {
		return nil, lower.coverage, err
	}
	if !lower.currentBlock().terminated {
		if !hasResult {
			result, err = lower.emitVoid(function)
			if err != nil {
				return nil, lower.coverage, err
			}
		}
		if err := lower.emit(function, semanticabi.OpReturnValue, result); err != nil {
			return nil, lower.coverage, err
		}
	}
	lower.popScope()
	if err := lower.finish(rootSource); err != nil {
		return nil, lower.coverage, err
	}
	return lower.image, lower.coverage, nil
}

func copyStringMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func (lower *lowerer) finish(rootSource uint32) error {
	for _, current := range lower.blocks {
		if len(current.instructions) == 0 {
			return fmt.Errorf("semanticabi flow: block %d is empty", current.id)
		}
		lower.image.Instructions = append(lower.image.Instructions, current.instructions...)
	}
	lower.image.RegisterTypes = append(lower.image.RegisterTypes, lower.registerTypes...)
	packageSymbol := lower.tables.symbol(lower.packageName)
	functionSymbol := lower.tables.symbol(lower.functionName)
	lower.image.Packages = []semanticabi.PackageRecord{{Name: packageSymbol, FunctionCount: 1}}
	lower.image.Functions = []semanticabi.FunctionRecord{{
		Name: functionSymbol, Package: 0, Source: rootSource,
		ParameterCount: uint32(len(lower.function.Params)), RegisterStart: 0,
		RegisterCount: uint32(len(lower.registerTypes)), BlockCount: uint32(len(lower.blocks)),
		InstructionCount: uint32(len(lower.image.Instructions)),
		Flags:            semanticabi.FunctionFlagShadowEligible | semanticabi.FunctionFlagFlowValidated,
	}}
	lower.coverage.Instructions = len(lower.image.Instructions)
	lower.coverage.Registers = len(lower.registerTypes)
	lower.coverage.Blocks = len(lower.blocks)
	lower.coverage.CallTargets = len(lower.image.CallTargets)
	if !lower.coverage.Complete() {
		return fmt.Errorf("semanticabi flow: function %q did not satisfy complete coverage", lower.functionName)
	}
	return semanticabi.Validate(lower.image)
}

func (lower *lowerer) newBlock() uint32 {
	id := uint32(len(lower.blocks))
	lower.blocks = append(lower.blocks, &block{id: id})
	lower.current = id
	return id
}

func (lower *lowerer) setCurrent(id uint32) { lower.current = id }

func (lower *lowerer) currentBlock() *block { return lower.blocks[lower.current] }

func (lower *lowerer) emit(node ast.Node, opcode uint16, operands ...uint32) error {
	return lower.emitTo(lower.current, node, opcode, operands...)
}

func (lower *lowerer) emitTo(blockID uint32, node ast.Node, opcode uint16, operands ...uint32) error {
	current := lower.blocks[blockID]
	if current.terminated {
		return fmt.Errorf("semanticabi flow: instruction after terminator in block %d", blockID)
	}
	source, err := lower.tables.source(node, callsiteFor(node, lower.rootSource))
	if err != nil {
		return err
	}
	descriptor, ok := semanticabi.OpByCode(opcode)
	if !ok {
		return fmt.Errorf("semanticabi flow: unknown opcode %d", opcode)
	}
	current.instructions = append(current.instructions, semanticabi.Instruction{
		Opcode: opcode, Source: source, Block: blockID, Operands: append([]uint32(nil), operands...),
	})
	current.terminated = descriptor.Terminator
	return nil
}

func (lower *lowerer) newRegister(typeID uint32) uint32 {
	register := uint32(len(lower.registerTypes))
	lower.registerTypes = append(lower.registerTypes, typeID)
	return register
}

func (lower *lowerer) registerType(register uint32) uint32 { return lower.registerTypes[register] }

func (lower *lowerer) pushScope() { lower.scopes = append(lower.scopes, make(map[string]binding)) }
func (lower *lowerer) popScope()  { lower.scopes = lower.scopes[:len(lower.scopes)-1] }

func (lower *lowerer) bind(name string, value binding) {
	lower.scopes[len(lower.scopes)-1][name] = value
}

func (lower *lowerer) lookup(name string) (binding, bool) {
	for index := len(lower.scopes) - 1; index >= 0; index-- {
		if value, ok := lower.scopes[index][name]; ok {
			return value, true
		}
	}
	return binding{}, false
}

func (lower *lowerer) auditASTCoverage(root ast.Node) error {
	var unsupported string
	ast.Walk(root, func(node ast.Node) bool {
		lower.coverage.ASTNodes++
		if !supportedNodeType(node.NodeType()) && unsupported == "" {
			unsupported = string(node.NodeType())
			lower.coverage.Unsupported = append(lower.coverage.Unsupported, unsupported)
			lower.coverage.ASTFallbacks++
		}
		return unsupported == ""
	})
	if unsupported != "" {
		return fmt.Errorf("semanticabi flow: unsupported AST node %s", unsupported)
	}
	return nil
}

func supportedNodeType(kind ast.NodeType) bool {
	switch kind {
	case ast.NodeFunctionDefinition, ast.NodeFunctionParameter, ast.NodeBlockExpression,
		ast.NodeIdentifier, ast.NodeStringLiteral, ast.NodeIntegerLiteral, ast.NodeFloatLiteral,
		ast.NodeBooleanLiteral, ast.NodeNilLiteral, ast.NodeCharLiteral,
		ast.NodeSimpleTypeExpression, ast.NodeGenericTypeExpression,
		ast.NodeAssignmentExpression, ast.NodeUnaryExpression, ast.NodeBinaryExpression,
		ast.NodeTypeCastExpression, ast.NodeFunctionCall, ast.NodeMemberAccessExpression,
		ast.NodeIfExpression, ast.NodeLoopExpression, ast.NodeBreakStatement,
		ast.NodeContinueStatement, ast.NodeReturnStatement, ast.NodeMatchExpression,
		ast.NodeMatchClause, ast.NodeTypedPattern, ast.NodeRaiseStatement:
		return true
	default:
		return false
	}
}

func findFunction(module *ast.Module, name string) *ast.FunctionDefinition {
	for _, statement := range module.Body {
		if function, ok := statement.(*ast.FunctionDefinition); ok && function.ID != nil && function.ID.Name == name {
			return function
		}
	}
	return nil
}

func modulePackageName(module *ast.Module) string {
	if module.Package == nil {
		return "<module>"
	}
	parts := make([]string, 0, len(module.Package.NamePath))
	for _, part := range module.Package.NamePath {
		if part != nil {
			parts = append(parts, part.Name)
		}
	}
	return strings.Join(parts, ".")
}
