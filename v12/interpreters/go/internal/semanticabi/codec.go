package semanticabi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var imageMagic = [8]byte{'A', 'B', 'L', 'E', 'S', 'A', 'B', 'I'}

const (
	sectionHeader uint16 = iota + 1
	sectionStrings
	sectionTypes
	sectionSources
	sectionConstants
	sectionPackagesFunctions
	sectionInstructions
	sectionCapabilities
)

const (
	preambleSize      = 16
	sectionEntrySize  = 12
	sectionDirectory  = SectionCount * sectionEntrySize
	firstSectionStart = preambleSize + sectionDirectory
	maxCollectionSize = 1 << 24
	maxBlobSize       = 1 << 28
)

type section struct {
	kind    uint16
	payload []byte
}

func Encode(image *Image) ([]byte, error) {
	if err := Validate(image); err != nil {
		return nil, err
	}
	sections := []section{
		{kind: sectionHeader, payload: encodeHeader(image.Header)},
		{kind: sectionStrings, payload: encodeStrings(image.Strings)},
		{kind: sectionTypes, payload: encodeTypes(image.Types)},
		{kind: sectionSources, payload: encodeSources(image.Sources)},
		{kind: sectionConstants, payload: encodeConstants(image.Constants)},
		{kind: sectionPackagesFunctions, payload: encodePackagesFunctions(image.Packages, image.Functions, image.RegisterTypes, image.CallTargets)},
		{kind: sectionInstructions, payload: encodeInstructions(image.Instructions)},
		{kind: sectionCapabilities, payload: encodeCapabilities(image.Capabilities)},
	}
	total := firstSectionStart
	for _, current := range sections {
		total += len(current.payload)
		if total > maxBlobSize {
			return nil, fmt.Errorf("semanticabi: encoded image exceeds %d bytes", maxBlobSize)
		}
	}
	result := make([]byte, 0, total)
	result = append(result, imageMagic[:]...)
	result = binary.LittleEndian.AppendUint16(result, FormatVersion)
	result = binary.LittleEndian.AppendUint16(result, SectionCount)
	result = binary.LittleEndian.AppendUint32(result, 0)
	offset := firstSectionStart
	for _, current := range sections {
		result = binary.LittleEndian.AppendUint16(result, current.kind)
		result = binary.LittleEndian.AppendUint16(result, 0)
		result = binary.LittleEndian.AppendUint32(result, uint32(offset))
		result = binary.LittleEndian.AppendUint32(result, uint32(len(current.payload)))
		offset += len(current.payload)
	}
	for _, current := range sections {
		result = append(result, current.payload...)
	}
	return result, nil
}

func Decode(data []byte) (*Image, error) {
	if len(data) < firstSectionStart {
		return nil, fmt.Errorf("semanticabi: image truncated before section directory")
	}
	if !bytes.Equal(data[:len(imageMagic)], imageMagic[:]) {
		return nil, fmt.Errorf("semanticabi: invalid image magic")
	}
	if version := binary.LittleEndian.Uint16(data[8:10]); version != FormatVersion {
		return nil, fmt.Errorf("semanticabi: unsupported format version %d", version)
	}
	if count := binary.LittleEndian.Uint16(data[10:12]); count != SectionCount {
		return nil, fmt.Errorf("semanticabi: expected %d sections, got %d", SectionCount, count)
	}
	if binary.LittleEndian.Uint32(data[12:16]) != 0 {
		return nil, fmt.Errorf("semanticabi: nonzero preamble flags")
	}
	sections := make([][]byte, SectionCount)
	expectedOffset := firstSectionStart
	for index := 0; index < SectionCount; index++ {
		entryStart := preambleSize + index*sectionEntrySize
		entry := data[entryStart : entryStart+sectionEntrySize]
		kind := binary.LittleEndian.Uint16(entry[0:2])
		if kind != uint16(index+1) {
			return nil, fmt.Errorf("semanticabi: section %d has kind %d", index, kind)
		}
		if binary.LittleEndian.Uint16(entry[2:4]) != 0 {
			return nil, fmt.Errorf("semanticabi: section %d has nonzero flags", index)
		}
		offset := binary.LittleEndian.Uint32(entry[4:8])
		length := binary.LittleEndian.Uint32(entry[8:12])
		end := uint64(offset) + uint64(length)
		if int(offset) != expectedOffset {
			return nil, fmt.Errorf("semanticabi: section %d starts at %d, expected %d", index, offset, expectedOffset)
		}
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("semanticabi: section %d exceeds image bounds", index)
		}
		sections[index] = data[offset:uint32(end)]
		expectedOffset = int(end)
	}
	if expectedOffset != len(data) {
		return nil, fmt.Errorf("semanticabi: trailing bytes after final section")
	}

	image := &Image{}
	var err error
	if image.Header, err = decodeHeader(sections[0]); err != nil {
		return nil, err
	}
	if image.Strings, err = decodeStrings(sections[1]); err != nil {
		return nil, err
	}
	if image.Types, err = decodeTypes(sections[2]); err != nil {
		return nil, err
	}
	if image.Sources, err = decodeSources(sections[3]); err != nil {
		return nil, err
	}
	if image.Constants, err = decodeConstants(sections[4]); err != nil {
		return nil, err
	}
	if image.Packages, image.Functions, image.RegisterTypes, image.CallTargets, err = decodePackagesFunctions(sections[5]); err != nil {
		return nil, err
	}
	if image.Instructions, err = decodeInstructions(sections[6]); err != nil {
		return nil, err
	}
	if image.Capabilities, err = decodeCapabilities(sections[7]); err != nil {
		return nil, err
	}
	if err := Validate(image); err != nil {
		return nil, err
	}
	return image, nil
}

func encodeHeader(header Header) []byte {
	result := make([]byte, 0, 44)
	result = binary.LittleEndian.AppendUint32(result, header.SemanticVersion)
	result = append(result, header.Identity[:]...)
	return binary.LittleEndian.AppendUint64(result, header.ProgramID)
}

func decodeHeader(data []byte) (Header, error) {
	if len(data) != 44 {
		return Header{}, fmt.Errorf("semanticabi: header section has length %d", len(data))
	}
	header := Header{SemanticVersion: binary.LittleEndian.Uint32(data[:4])}
	copy(header.Identity[:], data[4:36])
	header.ProgramID = binary.LittleEndian.Uint64(data[36:44])
	return header, nil
}

func encodeStrings(values []string) []byte {
	encoder := newEncoder(len(values) * 8)
	encoder.count(len(values))
	for _, value := range values {
		encoder.bytes([]byte(value))
	}
	return encoder.data
}

func decodeStrings(data []byte) ([]string, error) {
	decoder := newDecoder("strings", data)
	count, err := decoder.count()
	if err != nil {
		return nil, err
	}
	result := make([]string, count)
	for index := range result {
		value, err := decoder.bytes()
		if err != nil {
			return nil, err
		}
		result[index] = string(value)
	}
	return result, decoder.finish()
}

func encodeTypes(values []TypeRecord) []byte {
	encoder := newEncoder(4 + len(values)*8)
	encoder.count(len(values))
	for _, value := range values {
		encoder.u32(value.Name)
		encoder.u32(value.Flags)
	}
	return encoder.data
}

func decodeTypes(data []byte) ([]TypeRecord, error) {
	decoder := newDecoder("types", data)
	count, err := decoder.count()
	if err != nil {
		return nil, err
	}
	result := make([]TypeRecord, count)
	for index := range result {
		if result[index].Name, err = decoder.u32(); err != nil {
			return nil, err
		}
		if result[index].Flags, err = decoder.u32(); err != nil {
			return nil, err
		}
	}
	return result, decoder.finish()
}

func encodeSources(values []SourceRecord) []byte {
	encoder := newEncoder(4 + len(values)*24)
	encoder.count(len(values))
	for _, value := range values {
		encoder.u32(value.File)
		encoder.u32(value.StartLine)
		encoder.u32(value.StartColumn)
		encoder.u32(value.EndLine)
		encoder.u32(value.EndColumn)
		encoder.u32(value.Callsite)
	}
	return encoder.data
}

func decodeSources(data []byte) ([]SourceRecord, error) {
	decoder := newDecoder("sources", data)
	count, err := decoder.count()
	if err != nil {
		return nil, err
	}
	result := make([]SourceRecord, count)
	for index := range result {
		fields := []*uint32{&result[index].File, &result[index].StartLine, &result[index].StartColumn, &result[index].EndLine, &result[index].EndColumn, &result[index].Callsite}
		for _, field := range fields {
			if *field, err = decoder.u32(); err != nil {
				return nil, err
			}
		}
	}
	return result, decoder.finish()
}

func encodeConstants(values []ConstantRecord) []byte {
	encoder := newEncoder(4 + len(values)*16)
	encoder.count(len(values))
	for _, value := range values {
		encoder.u32(value.Tag)
		encoder.u32(value.Aux)
		encoder.bytes(value.Data)
	}
	return encoder.data
}

func decodeConstants(data []byte) ([]ConstantRecord, error) {
	decoder := newDecoder("constants", data)
	count, err := decoder.count()
	if err != nil {
		return nil, err
	}
	result := make([]ConstantRecord, count)
	for index := range result {
		if result[index].Tag, err = decoder.u32(); err != nil {
			return nil, err
		}
		if result[index].Aux, err = decoder.u32(); err != nil {
			return nil, err
		}
		if result[index].Data, err = decoder.bytes(); err != nil {
			return nil, err
		}
	}
	return result, decoder.finish()
}

func encodePackagesFunctions(packages []PackageRecord, functions []FunctionRecord, registerTypes []uint32, callTargets []CallTarget) []byte {
	encoder := newEncoder(16 + len(packages)*12 + len(functions)*40 + len(registerTypes)*4 + len(callTargets)*28)
	encoder.count(len(packages))
	for _, value := range packages {
		encoder.u32(value.Name)
		encoder.u32(value.FunctionStart)
		encoder.u32(value.FunctionCount)
	}
	encoder.count(len(functions))
	for _, value := range functions {
		encoder.u32(value.Name)
		encoder.u32(value.Package)
		encoder.u32(value.Source)
		encoder.u32(value.ParameterCount)
		encoder.u32(value.RegisterStart)
		encoder.u32(value.RegisterCount)
		encoder.u32(value.BlockCount)
		encoder.u32(value.InstructionStart)
		encoder.u32(value.InstructionCount)
		encoder.u32(value.Flags)
	}
	encoder.count(len(registerTypes))
	for _, value := range registerTypes {
		encoder.u32(value)
	}
	encoder.count(len(callTargets))
	for _, value := range callTargets {
		encoder.u32(uint32(value.Kind))
		encoder.u32(value.PackageName)
		encoder.u32(value.OwnerType)
		encoder.u32(value.Name)
		encoder.u32(value.Arity)
		encoder.u32(value.ReturnType)
		encoder.u32(value.Flags)
	}
	return encoder.data
}

func decodePackagesFunctions(data []byte) ([]PackageRecord, []FunctionRecord, []uint32, []CallTarget, error) {
	decoder := newDecoder("packages-functions", data)
	packageCount, err := decoder.count()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	packages := make([]PackageRecord, packageCount)
	for index := range packages {
		fields := []*uint32{&packages[index].Name, &packages[index].FunctionStart, &packages[index].FunctionCount}
		for _, field := range fields {
			if *field, err = decoder.u32(); err != nil {
				return nil, nil, nil, nil, err
			}
		}
	}
	functionCount, err := decoder.count()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	functions := make([]FunctionRecord, functionCount)
	for index := range functions {
		fields := []*uint32{&functions[index].Name, &functions[index].Package, &functions[index].Source, &functions[index].ParameterCount, &functions[index].RegisterStart, &functions[index].RegisterCount, &functions[index].BlockCount, &functions[index].InstructionStart, &functions[index].InstructionCount, &functions[index].Flags}
		for _, field := range fields {
			if *field, err = decoder.u32(); err != nil {
				return nil, nil, nil, nil, err
			}
		}
	}
	registerCount, err := decoder.count()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var registerTypes []uint32
	if registerCount != 0 {
		registerTypes = make([]uint32, registerCount)
	}
	for index := range registerTypes {
		if registerTypes[index], err = decoder.u32(); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	targetCount, err := decoder.count()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var callTargets []CallTarget
	if targetCount != 0 {
		callTargets = make([]CallTarget, targetCount)
	}
	for index := range callTargets {
		kind, kindErr := decoder.u32()
		if kindErr != nil {
			return nil, nil, nil, nil, kindErr
		}
		callTargets[index].Kind = CallTargetKind(kind)
		fields := []*uint32{&callTargets[index].PackageName, &callTargets[index].OwnerType, &callTargets[index].Name, &callTargets[index].Arity, &callTargets[index].ReturnType, &callTargets[index].Flags}
		for _, field := range fields {
			if *field, err = decoder.u32(); err != nil {
				return nil, nil, nil, nil, err
			}
		}
	}
	return packages, functions, registerTypes, callTargets, decoder.finish()
}

func encodeInstructions(values []Instruction) []byte {
	encoder := newEncoder(4 + len(values)*16)
	encoder.count(len(values))
	for _, value := range values {
		encoder.u16(value.Opcode)
		encoder.u16(value.Flags)
		encoder.u32(value.Source)
		encoder.u32(value.Block)
		encoder.count(len(value.Operands))
		for _, operand := range value.Operands {
			encoder.u32(operand)
		}
	}
	return encoder.data
}

func decodeInstructions(data []byte) ([]Instruction, error) {
	decoder := newDecoder("instructions", data)
	count, err := decoder.count()
	if err != nil {
		return nil, err
	}
	result := make([]Instruction, count)
	for index := range result {
		if result[index].Opcode, err = decoder.u16(); err != nil {
			return nil, err
		}
		if result[index].Flags, err = decoder.u16(); err != nil {
			return nil, err
		}
		if result[index].Source, err = decoder.u32(); err != nil {
			return nil, err
		}
		if result[index].Block, err = decoder.u32(); err != nil {
			return nil, err
		}
		operandCount, err := decoder.count()
		if err != nil {
			return nil, err
		}
		if operandCount != 0 {
			result[index].Operands = make([]uint32, operandCount)
		}
		for operandIndex := range result[index].Operands {
			if result[index].Operands[operandIndex], err = decoder.u32(); err != nil {
				return nil, err
			}
		}
	}
	return result, decoder.finish()
}

func encodeCapabilities(values []HostCapability) []byte {
	encoder := newEncoder(4 + len(values)*12)
	encoder.count(len(values))
	for _, value := range values {
		encoder.u32(value.Name)
		encoder.u32(value.EffectKind)
		encoder.u32(value.Flags)
	}
	return encoder.data
}

func decodeCapabilities(data []byte) ([]HostCapability, error) {
	decoder := newDecoder("capabilities", data)
	count, err := decoder.count()
	if err != nil {
		return nil, err
	}
	result := make([]HostCapability, count)
	for index := range result {
		fields := []*uint32{&result[index].Name, &result[index].EffectKind, &result[index].Flags}
		for _, field := range fields {
			if *field, err = decoder.u32(); err != nil {
				return nil, err
			}
		}
	}
	return result, decoder.finish()
}

type encoder struct{ data []byte }

func newEncoder(capacity int) *encoder { return &encoder{data: make([]byte, 0, capacity)} }
func (e *encoder) u16(value uint16)    { e.data = binary.LittleEndian.AppendUint16(e.data, value) }
func (e *encoder) u32(value uint32)    { e.data = binary.LittleEndian.AppendUint32(e.data, value) }
func (e *encoder) count(value int)     { e.u32(uint32(value)) }
func (e *encoder) bytes(value []byte) {
	e.count(len(value))
	e.data = append(e.data, value...)
}

type decoder struct {
	section  string
	data     []byte
	position int
}

func newDecoder(section string, data []byte) *decoder {
	return &decoder{section: section, data: data}
}

func (d *decoder) take(length int) ([]byte, error) {
	if length < 0 || d.position > len(d.data)-length {
		return nil, fmt.Errorf("semanticabi: %s section truncated at byte %d", d.section, d.position)
	}
	result := d.data[d.position : d.position+length]
	d.position += length
	return result, nil
}

func (d *decoder) u16() (uint16, error) {
	data, err := d.take(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(data), nil
}

func (d *decoder) u32() (uint32, error) {
	data, err := d.take(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

func (d *decoder) count() (int, error) {
	value, err := d.u32()
	if err != nil {
		return 0, err
	}
	if value > maxCollectionSize {
		return 0, fmt.Errorf("semanticabi: %s collection count %d exceeds limit", d.section, value)
	}
	return int(value), nil
}

func (d *decoder) bytes() ([]byte, error) {
	length, err := d.u32()
	if err != nil {
		return nil, err
	}
	if length > maxBlobSize {
		return nil, fmt.Errorf("semanticabi: %s blob length %d exceeds limit", d.section, length)
	}
	data, err := d.take(int(length))
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), data...), nil
}

func (d *decoder) finish() error {
	if d.position != len(d.data) {
		return fmt.Errorf("semanticabi: %s section has %d trailing bytes", d.section, len(d.data)-d.position)
	}
	return nil
}
