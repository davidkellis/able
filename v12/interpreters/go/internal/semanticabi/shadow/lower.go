package shadow

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
)

type Coverage struct {
	Function     string
	VisitedNodes int
	LoweredNodes int
	Unsupported  []string
	HostEffects  []string
	ASTFallbacks int
}

func (coverage Coverage) Complete() bool {
	return coverage.VisitedNodes != 0 &&
		coverage.VisitedNodes == coverage.LoweredNodes &&
		len(coverage.Unsupported) == 0 && coverage.ASTFallbacks == 0
}

func LowerFunction(module *ast.Module, functionName, sourcePath string, programID uint64) (*semanticabi.Image, Coverage, error) {
	if module == nil {
		return nil, Coverage{}, fmt.Errorf("semanticabi shadow: nil module")
	}
	if functionName == "" {
		return nil, Coverage{}, fmt.Errorf("semanticabi shadow: empty function name")
	}
	if sourcePath == "" {
		return nil, Coverage{}, fmt.Errorf("semanticabi shadow: empty source path")
	}
	if programID == 0 {
		return nil, Coverage{}, fmt.Errorf("semanticabi shadow: program id zero is invalid")
	}
	function := findFunction(module, functionName)
	if function == nil {
		return nil, Coverage{}, fmt.Errorf("semanticabi shadow: function %q not found", functionName)
	}
	builder := newBuilder(sourcePath, programID)
	coverage := Coverage{Function: functionName}
	rootSource, err := builder.source(function, semanticabi.NoIndex)
	if err != nil {
		return nil, coverage, err
	}
	var lowerErr error
	ast.Walk(function, func(node ast.Node) bool {
		if lowerErr != nil {
			return false
		}
		coverage.VisitedNodes++
		if err := builder.lowerNode(node, rootSource, &coverage); err != nil {
			coverage.Unsupported = append(coverage.Unsupported, string(node.NodeType()))
			lowerErr = err
			return false
		}
		coverage.LoweredNodes++
		return true
	})
	if lowerErr != nil {
		return nil, coverage, lowerErr
	}
	if !coverage.Complete() {
		return nil, coverage, fmt.Errorf("semanticabi shadow: function %q is not completely lowered", functionName)
	}
	packageName := modulePackageName(module)
	packageSymbol := builder.symbol(packageName)
	functionSymbol := builder.symbol(functionName)
	registerCount := builder.localDeclarations + uint32(len(function.Params))
	builder.image.Packages = []semanticabi.PackageRecord{{
		Name:          packageSymbol,
		FunctionStart: 0,
		FunctionCount: 1,
	}}
	builder.image.Functions = []semanticabi.FunctionRecord{{
		Name:             functionSymbol,
		Package:          0,
		Source:           rootSource,
		ParameterCount:   uint32(len(function.Params)),
		RegisterCount:    registerCount,
		BlockCount:       1,
		InstructionStart: 0,
		InstructionCount: uint32(len(builder.image.Instructions)),
		Flags:            semanticabi.FunctionFlagShadowEligible,
	}}
	if err := semanticabi.Validate(builder.image); err != nil {
		return nil, coverage, err
	}
	return builder.image, coverage, nil
}

type sourceKey struct {
	startLine   uint32
	startColumn uint32
	endLine     uint32
	endColumn   uint32
	callsite    uint32
}

type builder struct {
	image             *semanticabi.Image
	symbols           map[string]uint32
	types             map[string]uint32
	sources           map[sourceKey]uint32
	constants         map[string]uint32
	capabilities      map[string]uint32
	sourcePathIndex   uint32
	localDeclarations uint32
}

func newBuilder(sourcePath string, programID uint64) *builder {
	result := &builder{
		image:        &semanticabi.Image{Header: semanticabi.CurrentHeader(programID)},
		symbols:      make(map[string]uint32),
		types:        make(map[string]uint32),
		sources:      make(map[sourceKey]uint32),
		constants:    make(map[string]uint32),
		capabilities: make(map[string]uint32),
	}
	result.sourcePathIndex = result.symbol(sourcePath)
	return result
}

func (builder *builder) lowerNode(node ast.Node, rootSource uint32, coverage *Coverage) error {
	callsite := semanticabi.NoIndex
	if node.NodeType() == ast.NodeFunctionCall {
		callsite = rootSource
	}
	source, err := builder.source(node, callsite)
	if err != nil {
		return err
	}
	emit := func(opcode uint16, operands ...uint32) {
		builder.image.Instructions = append(builder.image.Instructions, semanticabi.Instruction{
			Opcode: opcode, Source: source, Block: 0, Operands: operands,
		})
	}
	switch value := node.(type) {
	case *ast.FunctionDefinition:
		emit(semanticabi.OpDeclareFunction, builder.symbol(value.ID.Name), uint32(len(value.Params)))
	case *ast.FunctionParameter:
		name, ok := patternBindingName(value.Name)
		if !ok {
			name = string(value.Name.NodeType())
		}
		emit(semanticabi.OpParameter, builder.symbol(name))
	case *ast.BlockExpression:
		emit(semanticabi.OpBlock)
	case *ast.Identifier:
		emit(semanticabi.OpLoadName, builder.symbol(value.Name))
	case *ast.StringLiteral:
		emit(semanticabi.OpConstant, builder.constant(semanticabi.TagKindString, 0, []byte(value.Value)))
	case *ast.IntegerLiteral:
		emit(semanticabi.OpConstant, builder.constant(semanticabi.TagKindInteger, integerAux(value.IntegerType), signedMagnitude(value.Value)))
	case *ast.FloatLiteral:
		data := make([]byte, 8)
		binary.LittleEndian.PutUint64(data, math.Float64bits(value.Value))
		emit(semanticabi.OpConstant, builder.constant(semanticabi.TagKindFloat, floatAux(value.FloatType), data))
	case *ast.BooleanLiteral:
		payload := byte(0)
		if value.Value {
			payload = 1
		}
		emit(semanticabi.OpConstant, builder.constant(semanticabi.TagKindBool, 0, []byte{payload}))
	case *ast.NilLiteral:
		emit(semanticabi.OpConstant, builder.constant(semanticabi.TagKindNil, 0, nil))
	case *ast.CharLiteral:
		emit(semanticabi.OpConstant, builder.constant(semanticabi.TagKindChar, 0, []byte(value.Value)))
	case ast.TypeExpression:
		emit(semanticabi.OpTypeRef, builder.typeIndex(typeName(value)))
	case *ast.AssignmentExpression:
		emit(semanticabi.OpAssign, builder.symbol(string(value.Operator)))
		if value.Operator == ast.AssignmentDeclare {
			builder.localDeclarations++
		}
	case *ast.UnaryExpression:
		emit(semanticabi.OpUnary, builder.symbol(string(value.Operator)))
	case *ast.BinaryExpression:
		emit(semanticabi.OpBinary, builder.symbol(value.Operator))
	case *ast.TypeCastExpression:
		emit(semanticabi.OpCast)
	case *ast.FunctionCall:
		if capabilityName, ok := hostCapability(value.Callee); ok {
			capability := builder.capability(capabilityName)
			emit(semanticabi.OpHostEffect, capability, uint32(len(value.Arguments)))
			coverage.HostEffects = appendUnique(coverage.HostEffects, capabilityName)
		} else {
			emit(semanticabi.OpCall, uint32(len(value.Arguments)))
		}
	case *ast.MemberAccessExpression:
		name := "<dynamic-member>"
		if identifier, ok := value.Member.(*ast.Identifier); ok {
			name = identifier.Name
		}
		emit(semanticabi.OpMember, builder.symbol(name))
	case *ast.IfExpression:
		emit(semanticabi.OpIf, 0)
	case *ast.LoopExpression, *ast.WhileLoop, *ast.ForLoop:
		emit(semanticabi.OpLoop, 0)
	case *ast.BreakStatement, *ast.ContinueStatement:
		emit(semanticabi.OpBreak, 0)
	case *ast.ReturnStatement:
		emit(semanticabi.OpReturn)
	case *ast.MatchExpression:
		emit(semanticabi.OpMatch, uint32(len(value.Clauses)))
	case *ast.MatchClause:
		guard := uint32(0)
		if value.Guard != nil {
			guard = 1
		}
		emit(semanticabi.OpMatchClause, guard)
	case ast.Pattern:
		emit(semanticabi.OpPattern, builder.symbol(string(value.NodeType())))
	case *ast.RaiseStatement, *ast.RethrowStatement:
		emit(semanticabi.OpRaise)
	case *ast.ArrayLiteral:
		emit(semanticabi.OpArray, uint32(len(value.Elements)))
	case *ast.StringInterpolation:
		emit(semanticabi.OpInterpolate, uint32(len(value.Parts)))
	default:
		return fmt.Errorf("semanticabi shadow: unsupported node %s", node.NodeType())
	}
	return nil
}

func (builder *builder) symbol(value string) uint32 {
	if index, ok := builder.symbols[value]; ok {
		return index
	}
	index := uint32(len(builder.image.Strings))
	builder.image.Strings = append(builder.image.Strings, value)
	builder.symbols[value] = index
	return index
}

func (builder *builder) typeIndex(name string) uint32 {
	if index, ok := builder.types[name]; ok {
		return index
	}
	index := uint32(len(builder.image.Types))
	builder.image.Types = append(builder.image.Types, semanticabi.TypeRecord{Name: builder.symbol(name)})
	builder.types[name] = index
	return index
}

func (builder *builder) source(node ast.Node, callsite uint32) (uint32, error) {
	span := node.Span()
	positions := []int{span.Start.Line, span.Start.Column, span.End.Line, span.End.Column}
	for _, position := range positions {
		if position < 0 {
			return 0, fmt.Errorf("semanticabi shadow: negative source position for %s", node.NodeType())
		}
	}
	key := sourceKey{
		startLine: uint32(span.Start.Line), startColumn: uint32(span.Start.Column),
		endLine: uint32(span.End.Line), endColumn: uint32(span.End.Column), callsite: callsite,
	}
	if index, ok := builder.sources[key]; ok {
		return index, nil
	}
	index := uint32(len(builder.image.Sources))
	builder.image.Sources = append(builder.image.Sources, semanticabi.SourceRecord{
		File: builder.sourcePathIndex, StartLine: key.startLine, StartColumn: key.startColumn,
		EndLine: key.endLine, EndColumn: key.endColumn, Callsite: callsite,
	})
	builder.sources[key] = index
	return index, nil
}

func (builder *builder) constant(tag, aux uint32, data []byte) uint32 {
	key := fmt.Sprintf("%d:%d:%x", tag, aux, data)
	if index, ok := builder.constants[key]; ok {
		return index
	}
	index := uint32(len(builder.image.Constants))
	builder.image.Constants = append(builder.image.Constants, semanticabi.ConstantRecord{
		Tag: tag, Aux: aux, Data: append([]byte(nil), data...),
	})
	builder.constants[key] = index
	return index
}

func (builder *builder) capability(name string) uint32 {
	if index, ok := builder.capabilities[name]; ok {
		return index
	}
	index := uint32(len(builder.image.Capabilities))
	builder.image.Capabilities = append(builder.image.Capabilities, semanticabi.HostCapability{
		Name: builder.symbol(name), EffectKind: 1,
	})
	builder.capabilities[name] = index
	return index
}

func findFunction(module *ast.Module, name string) *ast.FunctionDefinition {
	for _, statement := range module.Body {
		function, ok := statement.(*ast.FunctionDefinition)
		if ok && function != nil && function.ID != nil && function.ID.Name == name {
			return function
		}
	}
	return nil
}

func modulePackageName(module *ast.Module) string {
	if module.Package == nil || len(module.Package.NamePath) == 0 {
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

func patternBindingName(pattern ast.Pattern) (string, bool) {
	identifier, ok := pattern.(*ast.Identifier)
	if !ok || identifier == nil {
		return "", false
	}
	return identifier.Name, true
}

func typeName(expression ast.TypeExpression) string {
	switch value := expression.(type) {
	case *ast.SimpleTypeExpression:
		if value.Name != nil {
			return value.Name.Name
		}
	case *ast.GenericTypeExpression:
		arguments := make([]string, len(value.Arguments))
		for index, argument := range value.Arguments {
			arguments[index] = typeName(argument)
		}
		return typeName(value.Base) + "<" + strings.Join(arguments, ",") + ">"
	case *ast.NullableTypeExpression:
		return typeName(value.InnerType) + "?"
	case *ast.ResultTypeExpression:
		return typeName(value.InnerType) + "!"
	case *ast.UnionTypeExpression:
		members := make([]string, len(value.Members))
		for index, member := range value.Members {
			members[index] = typeName(member)
		}
		return strings.Join(members, "|")
	case *ast.WildcardTypeExpression:
		return "_"
	}
	return string(expression.NodeType())
}

func hostCapability(callee ast.Expression) (string, bool) {
	name := expressionName(callee)
	switch name {
	case "print":
		return "able.host.print", true
	case "math.hypot", "able.math.hypot":
		return "able.math.hypot", true
	default:
		return "", false
	}
}

func expressionName(expression ast.Expression) string {
	switch value := expression.(type) {
	case *ast.Identifier:
		return value.Name
	case *ast.MemberAccessExpression:
		left := expressionName(value.Object)
		right := expressionName(value.Member)
		if left != "" && right != "" {
			return left + "." + right
		}
	}
	return ""
}

func integerAux(value *ast.IntegerType) uint32 {
	if value == nil {
		return 0
	}
	order := []ast.IntegerType{
		ast.IntegerTypeI8, ast.IntegerTypeI16, ast.IntegerTypeI32, ast.IntegerTypeI64, ast.IntegerTypeI128,
		ast.IntegerTypeU8, ast.IntegerTypeU16, ast.IntegerTypeU32, ast.IntegerTypeU64, ast.IntegerTypeU128,
	}
	for index, candidate := range order {
		if *value == candidate {
			return uint32(index + 1)
		}
	}
	return math.MaxUint32
}

func floatAux(value *ast.FloatType) uint32 {
	if value != nil && *value == ast.FloatTypeF32 {
		return 1
	}
	return 2
}

func signedMagnitude(value *big.Int) []byte {
	if value == nil || value.Sign() == 0 {
		return []byte{0}
	}
	abs := new(big.Int).Abs(value).Bytes()
	sign := byte(0)
	if value.Sign() < 0 {
		sign = 1
	}
	return append([]byte{sign}, abs...)
}

func appendUnique(values []string, value string) []string {
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}
