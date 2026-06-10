package semanticabi

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"
	"testing"
)

func TestCodecIsDeterministicAndPreservesSourceCallsites(t *testing.T) {
	image := testImage()
	first, err := Encode(image)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Encode(image)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated encodes differ")
	}
	decoded, err := Decode(first)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, image) {
		t.Fatalf("decoded image differs:\n got %#v\nwant %#v", decoded, image)
	}
	if decoded.Sources[1].Callsite != 0 {
		t.Fatalf("callsite = %d, want source 0", decoded.Sources[1].Callsite)
	}
	reencoded, err := Encode(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, reencoded) {
		t.Fatal("decode/re-encode is not byte-identical")
	}
}

func TestDecoderRejectsMalformedImagesDeterministically(t *testing.T) {
	valid, err := Encode(testImage())
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func([]byte) []byte{
		"truncated":      func(data []byte) []byte { return data[:len(data)-1] },
		"magic":          func(data []byte) []byte { data[0] ^= 0xff; return data },
		"section-kind":   func(data []byte) []byte { binary.LittleEndian.PutUint16(data[16:18], 8); return data },
		"section-offset": func(data []byte) []byte { binary.LittleEndian.PutUint32(data[20:24], 0); return data },
		"manifest-identity": func(data []byte) []byte {
			offset := sectionOffset(data, 0)
			data[offset+4] ^= 0xff
			return data
		},
		"unknown-opcode": func(data []byte) []byte {
			offset := sectionOffset(data, 6)
			binary.LittleEndian.PutUint16(data[offset+4:offset+6], 0xffff)
			return data
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			data := mutate(append([]byte(nil), valid...))
			_, firstErr := Decode(data)
			_, secondErr := Decode(data)
			if firstErr == nil || secondErr == nil {
				t.Fatal("malformed image was accepted")
			}
			if firstErr.Error() != secondErr.Error() {
				t.Fatalf("errors differ: %q vs %q", firstErr, secondErr)
			}
			if !strings.HasPrefix(firstErr.Error(), "semanticabi:") {
				t.Fatalf("error lacks stable prefix: %v", firstErr)
			}
		})
	}
}

func TestValidatorChecksIndicesTypesAndControlFlow(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Image)
		want   string
	}{
		{name: "source", mutate: func(image *Image) { image.Instructions[0].Source = 99 }, want: "source 99 out of range"},
		{name: "block", mutate: func(image *Image) { image.Instructions[0].Block = 1 }, want: "block 1 out of range"},
		{name: "operand-count", mutate: func(image *Image) { image.Instructions[0].Operands = nil }, want: "requires 2 operands"},
		{name: "symbol-operand", mutate: func(image *Image) { image.Instructions[0].Operands[0] = 99 }, want: "operand 0 (99) out of range"},
		{name: "type-operand", mutate: func(image *Image) { image.Instructions[2].Operands[0] = 99 }, want: "operand 0 (99) out of range"},
		{name: "control-target", mutate: func(image *Image) {
			image.Instructions[4] = Instruction{Opcode: OpBreak, Source: 0, Block: 0, Operands: []uint32{2}}
		}, want: "operand 0 (2) out of range"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			image := testImage()
			test.mutate(image)
			err := Validate(image)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func testImage() *Image {
	return &Image{
		Header:  CurrentHeader(41),
		Strings: []string{"bench.able", "bench", "main", "i32", "able.host.print"},
		Types:   []TypeRecord{{Name: 3}},
		Sources: []SourceRecord{
			{File: 0, StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1, Callsite: NoIndex},
			{File: 0, StartLine: 1, StartColumn: 5, EndLine: 1, EndColumn: 9, Callsite: 0},
		},
		Constants: []ConstantRecord{{Tag: TagKindInteger, Aux: 3, Data: []byte{0, 7}}},
		Packages:  []PackageRecord{{Name: 1, FunctionStart: 0, FunctionCount: 1}},
		Functions: []FunctionRecord{{
			Name: 2, Package: 0, Source: 0, RegisterCount: 1, BlockCount: 1,
			InstructionCount: 5, Flags: FunctionFlagShadowEligible,
		}},
		Instructions: []Instruction{
			{Opcode: OpDeclareFunction, Source: 0, Operands: []uint32{2, 0}},
			{Opcode: OpConstant, Source: 0, Operands: []uint32{0}},
			{Opcode: OpTypeRef, Source: 0, Operands: []uint32{0}},
			{Opcode: OpHostEffect, Source: 1, Operands: []uint32{0, 1}},
			{Opcode: OpReturn, Source: 0},
		},
		Capabilities: []HostCapability{{Name: 4, EffectKind: 1}},
	}
}

func sectionOffset(data []byte, sectionIndex int) int {
	entry := 16 + sectionIndex*12
	return int(binary.LittleEndian.Uint32(data[entry+4 : entry+8]))
}
