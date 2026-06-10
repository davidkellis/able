package compiler

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

var compilerImportBindingSliceSink []staticImportBinding
var compilerImportableNameSetSink map[string]struct{}

func TestCompilerImportResolutionCachesInvalidateWhenBindingsGrow(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	expr := ast.Ty("RemoteThing")

	if sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias("app", "RemoteThing"); sourcePkg != "" || sourceName != "" {
		t.Fatalf("expected initial selector lookup to miss, got %q.%q", sourcePkg, sourceName)
	}
	if gen.importedSelectorAliasAppearsInTypeExpr("app", expr) {
		t.Fatal("expected initial type expression lookup to miss")
	}

	gen.addStaticImportBinding("app", staticImportBinding{
		Kind:          staticImportBindingSelector,
		SourcePackage: "example.types",
		LocalName:     "RemoteThing",
		SourceName:    "Thing",
	})

	sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias("app", "RemoteThing")
	if sourcePkg != "example.types" || sourceName != "Thing" {
		t.Fatalf("expected refreshed selector lookup, got %q.%q", sourcePkg, sourceName)
	}
	if !gen.importedSelectorAliasAppearsInTypeExpr("app", expr) {
		t.Fatal("expected refreshed type expression lookup to find the selector alias")
	}
}

func TestCompilerSourceReexportResolutionCacheInvalidatesWhenBindingsGrow(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})

	if sourcePkg, sourceName, ok := gen.sourceReexportSourceForName("facade", "Thing"); ok || sourcePkg != "" || sourceName != "" {
		t.Fatalf("expected initial re-export lookup to miss, got ok=%t %q.%q", ok, sourcePkg, sourceName)
	}
	gen.addSourceReexportBinding("facade", staticImportBinding{
		Kind:          staticImportBindingSelector,
		SourcePackage: "example.types",
		LocalName:     "Thing",
		SourceName:    "Thing",
	})

	sourcePkg, sourceName, ok := gen.sourceReexportSourceForName("facade", "Thing")
	if !ok || sourcePkg != "example.types" || sourceName != "Thing" {
		t.Fatalf("expected refreshed re-export lookup, got ok=%t %q.%q", ok, sourcePkg, sourceName)
	}
}

func TestCompilerTypeNormalizationCachesBySourceExpressionIdentity(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	expr := ast.Gen(ast.Ty("Array"), ast.Ty("i32"))

	firstPkg, first := gen.normalizeTypeExprContextForPackage("app", expr)
	secondPkg, second := gen.normalizeTypeExprContextForPackage("app", expr)
	if first == nil || firstPkg != secondPkg || first != second {
		t.Fatalf("expected repeated normalization to reuse the same result, got first=%T second=%T packages=%q/%q", first, second, firstPkg, secondPkg)
	}
	cacheKey := typeExprPackageCacheKey{Expr: expr, PackageName: "app"}
	if cached, ok := gen.typeExprPackageCache[cacheKey]; !ok || cached.NormalizedExpr != first {
		t.Fatal("expected the source-expression cache to retain the normalized result")
	}
}

func TestCompilerCollectedImportReadsDoNotCopyBindingSlices(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	gen.addStaticImportBinding("app", staticImportBinding{
		Kind:          staticImportBindingSelector,
		SourcePackage: "example.types",
		LocalName:     "Thing",
		SourceName:    "Thing",
	})
	gen.addSourceReexportBinding("facade", staticImportBinding{
		Kind:          staticImportBindingSelector,
		SourcePackage: "example.types",
		LocalName:     "Thing",
		SourceName:    "Thing",
	})

	if allocs := testing.AllocsPerRun(1000, func() {
		compilerImportBindingSliceSink = gen.staticImportsForPackage("app")
		compilerImportBindingSliceSink = gen.sourceReexportsForPackage("facade")
	}); allocs != 0 {
		t.Fatalf("expected collected import reads to allocate zero times, got %.2f", allocs)
	}
}

func TestCompilerImportableNameSetCacheInvalidatesWhenBindingsGrow(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	gen.typeAliases["source"] = map[string]ast.TypeExpression{"Thing": ast.Ty("i32")}
	gen.typeAliasPrivate["source"] = map[string]bool{}

	if names := gen.importableNameSet("facade"); len(names) != 0 {
		t.Fatalf("expected the initial facade export set to be empty, got %v", names)
	}
	gen.addSourceReexportBinding("facade", staticImportBinding{
		Kind:          staticImportBindingWildcard,
		SourcePackage: "source",
	})
	if _, ok := gen.importableNameSet("facade")["Thing"]; !ok {
		t.Fatal("expected the export-set cache to refresh after a re-export binding grows")
	}
}

func TestCompilerImportableNameSetCacheReusesReadOnlySet(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})
	gen.typeAliases["source"] = map[string]ast.TypeExpression{"Thing": ast.Ty("i32")}
	gen.typeAliasPrivate["source"] = map[string]bool{}
	compilerImportableNameSetSink = gen.importableNameSet("source")

	if allocs := testing.AllocsPerRun(1000, func() {
		compilerImportableNameSetSink = gen.importableNameSet("source")
	}); allocs != 0 {
		t.Fatalf("expected repeated export-set reads to allocate zero times, got %.2f", allocs)
	}
}
