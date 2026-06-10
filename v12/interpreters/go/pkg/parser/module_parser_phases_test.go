package parser

import "testing"

func TestModuleParserPhaseObserver(t *testing.T) {
	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser: %v", err)
	}
	defer p.Close()

	source := []byte("fn main() { 1 }\n")
	var samples []ModuleParsePhaseSample
	p.SetPhaseObserver(func(sample ModuleParsePhaseSample) {
		samples = append(samples, sample)
	})
	if _, err := p.ParseModule(source); err != nil {
		t.Fatalf("ParseModule: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("phase samples = %d, want 1", len(samples))
	}
	if samples[0].SourceBytes != len(source) {
		t.Fatalf("source bytes = %d, want %d", samples[0].SourceBytes, len(source))
	}
	if samples[0].NativeParse <= 0 || samples[0].ASTMapping <= 0 {
		t.Fatalf("phase durations must be positive: %#v", samples[0])
	}
}
