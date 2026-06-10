package semanticabi

import (
	"bytes"
	"fmt"
)

func Validate(image *Image) error {
	if image == nil {
		return fmt.Errorf("semanticabi: nil image")
	}
	if image.Header.SemanticVersion != SemanticVersion {
		return fmt.Errorf("semanticabi: unsupported semantic version %d", image.Header.SemanticVersion)
	}
	if image.Header.Identity != ManifestIdentity {
		return fmt.Errorf("semanticabi: semantic manifest identity mismatch")
	}
	if image.Header.ProgramID == 0 {
		return fmt.Errorf("semanticabi: program id zero is invalid")
	}
	for index, value := range image.Strings {
		if bytes.IndexByte([]byte(value), 0) >= 0 {
			return fmt.Errorf("semanticabi: string %d contains NUL", index)
		}
	}
	for index, record := range image.Types {
		if err := validateIndex("type name", index, record.Name, len(image.Strings)); err != nil {
			return err
		}
	}
	for index, record := range image.Sources {
		if err := validateIndex("source file", index, record.File, len(image.Strings)); err != nil {
			return err
		}
		if record.Callsite != NoIndex && int(record.Callsite) >= len(image.Sources) {
			return fmt.Errorf("semanticabi: source %d callsite %d out of range", index, record.Callsite)
		}
		if record.EndLine < record.StartLine || (record.EndLine == record.StartLine && record.EndColumn < record.StartColumn) {
			return fmt.Errorf("semanticabi: source %d has reversed span", index)
		}
	}
	for index, record := range image.Constants {
		if _, ok := KindByTag(record.Tag); !ok {
			return fmt.Errorf("semanticabi: constant %d has unknown tag %d", index, record.Tag)
		}
	}
	for index, record := range image.Packages {
		if err := validateIndex("package name", index, record.Name, len(image.Strings)); err != nil {
			return err
		}
		if err := validateRange("package function", index, record.FunctionStart, record.FunctionCount, len(image.Functions)); err != nil {
			return err
		}
	}
	for index, record := range image.Functions {
		if err := validateIndex("function name", index, record.Name, len(image.Strings)); err != nil {
			return err
		}
		if err := validateIndex("function package", index, record.Package, len(image.Packages)); err != nil {
			return err
		}
		if err := validateIndex("function source", index, record.Source, len(image.Sources)); err != nil {
			return err
		}
		if record.BlockCount == 0 {
			return fmt.Errorf("semanticabi: function %d has no control-flow blocks", index)
		}
		if record.Flags&FunctionFlagFlowValidated != 0 {
			if err := validateRange("function register", index, record.RegisterStart, record.RegisterCount, len(image.RegisterTypes)); err != nil {
				return err
			}
			if record.ParameterCount > record.RegisterCount {
				return fmt.Errorf("semanticabi: function %d has %d parameters but only %d registers", index, record.ParameterCount, record.RegisterCount)
			}
		}
		if err := validateRange("function instruction", index, record.InstructionStart, record.InstructionCount, len(image.Instructions)); err != nil {
			return err
		}
		for instructionIndex := record.InstructionStart; instructionIndex < record.InstructionStart+record.InstructionCount; instructionIndex++ {
			if err := validateInstruction(image, index, int(instructionIndex), record); err != nil {
				return err
			}
		}
		if record.Flags&FunctionFlagFlowValidated != 0 {
			if err := validateFlowFunction(image, index, record); err != nil {
				return err
			}
		}
	}
	for index, typeIndex := range image.RegisterTypes {
		if int(typeIndex) >= len(image.Types) {
			return fmt.Errorf("semanticabi: register type %d index %d out of range", index, typeIndex)
		}
	}
	for index, target := range image.CallTargets {
		if target.Kind < CallTargetLocal || target.Kind > CallTargetBuiltin {
			return fmt.Errorf("semanticabi: call target %d has invalid kind %d", index, target.Kind)
		}
		if target.PackageName != NoIndex && int(target.PackageName) >= len(image.Strings) {
			return fmt.Errorf("semanticabi: call target %d package %d out of range", index, target.PackageName)
		}
		if target.OwnerType != NoIndex && int(target.OwnerType) >= len(image.Types) {
			return fmt.Errorf("semanticabi: call target %d owner type %d out of range", index, target.OwnerType)
		}
		if int(target.Name) >= len(image.Strings) {
			return fmt.Errorf("semanticabi: call target %d name %d out of range", index, target.Name)
		}
		if target.ReturnType != NoIndex && int(target.ReturnType) >= len(image.Types) {
			return fmt.Errorf("semanticabi: call target %d return type %d out of range", index, target.ReturnType)
		}
	}
	for index, record := range image.Capabilities {
		if err := validateIndex("capability name", index, record.Name, len(image.Strings)); err != nil {
			return err
		}
		if record.EffectKind == 0 {
			return fmt.Errorf("semanticabi: capability %d has zero effect kind", index)
		}
	}
	return nil
}

func validateInstruction(image *Image, functionIndex, instructionIndex int, function FunctionRecord) error {
	instruction := image.Instructions[instructionIndex]
	descriptor, ok := OpByCode(instruction.Opcode)
	if !ok {
		return fmt.Errorf("semanticabi: instruction %d has unknown opcode %d", instructionIndex, instruction.Opcode)
	}
	if int(instruction.Source) >= len(image.Sources) {
		return fmt.Errorf("semanticabi: instruction %d source %d out of range", instructionIndex, instruction.Source)
	}
	if instruction.Block >= function.BlockCount {
		return fmt.Errorf("semanticabi: function %d instruction %d block %d out of range", functionIndex, instructionIndex, instruction.Block)
	}
	if descriptor.Variadic == 0 && len(instruction.Operands) != len(descriptor.Operands) {
		return fmt.Errorf("semanticabi: instruction %d opcode %s requires %d operands, got %d", instructionIndex, descriptor.Name, len(descriptor.Operands), len(instruction.Operands))
	}
	if descriptor.Variadic != 0 && len(instruction.Operands) < len(descriptor.Operands) {
		return fmt.Errorf("semanticabi: instruction %d opcode %s requires at least %d operands, got %d", instructionIndex, descriptor.Name, len(descriptor.Operands), len(instruction.Operands))
	}
	for operandIndex := range instruction.Operands {
		kind := descriptor.Variadic
		if operandIndex < len(descriptor.Operands) {
			kind = descriptor.Operands[operandIndex]
		}
		value := instruction.Operands[operandIndex]
		var limit int
		switch kind {
		case OperandImmediate:
			continue
		case OperandSymbol:
			limit = len(image.Strings)
		case OperandType:
			limit = len(image.Types)
		case OperandConstant:
			limit = len(image.Constants)
		case OperandBlock:
			limit = int(function.BlockCount)
		case OperandCapability:
			limit = len(image.Capabilities)
		case OperandRegister:
			limit = int(function.RegisterCount)
		case OperandCallTarget:
			limit = len(image.CallTargets)
		default:
			return fmt.Errorf("semanticabi: opcode %s has unknown operand class %d", descriptor.Name, kind)
		}
		if int(value) >= limit {
			return fmt.Errorf("semanticabi: instruction %d operand %d (%d) out of range", instructionIndex, operandIndex, value)
		}
	}
	return nil
}

func validateIndex(label string, owner int, value uint32, limit int) error {
	if int(value) >= limit {
		return fmt.Errorf("semanticabi: %s %d index %d out of range", label, owner, value)
	}
	return nil
}

func validateRange(label string, owner int, start, count uint32, limit int) error {
	end := uint64(start) + uint64(count)
	if end > uint64(limit) {
		return fmt.Errorf("semanticabi: %s %d range [%d,%d) out of range", label, owner, start, end)
	}
	return nil
}
