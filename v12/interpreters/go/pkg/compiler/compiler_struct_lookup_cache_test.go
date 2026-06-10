package compiler

import "testing"

var compilerStructInfoSink *structInfo
var compilerStructInfoFoundSink bool

func TestCompilerStructInfoByNameCacheInvalidatesWhenStructsGrow(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	if info, ok := gen.structInfoByNameUnique("Thing"); ok || info != nil {
		t.Fatalf("expected initial lookup to miss, got ok=%t info=%v", ok, info)
	}

	expected := &structInfo{Name: "Thing", Package: "example", GoName: "Thing"}
	gen.structs["example.Thing"] = expected
	info, ok := gen.structInfoByNameUnique("Thing")
	if !ok || info != expected {
		t.Fatalf("expected lookup to refresh after the struct table grew, got ok=%t info=%v", ok, info)
	}

	gen.structs["other.Thing"] = &structInfo{Name: "Thing", Package: "other", GoName: "OtherThing"}
	if info, ok := gen.structInfoByNameUnique("Thing"); ok || info != nil {
		t.Fatalf("expected lookup to refresh to an ambiguous result, got ok=%t info=%v", ok, info)
	}
}

func TestCompilerStructInfoByNameCacheReusesResult(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	gen.structs["example.Thing"] = &structInfo{Name: "Thing", Package: "example", GoName: "Thing"}
	compilerStructInfoSink, compilerStructInfoFoundSink = gen.structInfoByNameUnique("Thing")

	if allocs := testing.AllocsPerRun(1000, func() {
		compilerStructInfoSink, compilerStructInfoFoundSink = gen.structInfoByNameUnique("Thing")
	}); allocs != 0 {
		t.Fatalf("expected repeated unique-name lookups to allocate zero times, got %.2f", allocs)
	}
}
