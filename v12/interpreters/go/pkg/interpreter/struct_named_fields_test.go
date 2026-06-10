package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestStructDefinitionNamedFieldIndexUsesDirectLookupForSmallDefinitions(t *testing.T) {
	def := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
			ast.FieldDef(ast.Ty("i32"), "z"),
			ast.FieldDef(ast.Ty("i32"), "w"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	structDefinitionFieldIndexCacheMu.Lock()
	delete(structDefinitionFieldIndexCache, def)
	structDefinitionFieldIndexCacheMu.Unlock()

	tests := []struct {
		name  string
		want  int
		found bool
	}{
		{name: "x", want: 0, found: true},
		{name: "y", want: 1, found: true},
		{name: "z", want: 2, found: true},
		{name: "w", want: 3, found: true},
		{name: "missing", want: 0, found: false},
	}

	for _, tc := range tests {
		got, ok := structDefinitionNamedFieldIndex(def, tc.name)
		if ok != tc.found || got != tc.want {
			t.Fatalf("structDefinitionNamedFieldIndex(%q) = (%d, %t), want (%d, %t)", tc.name, got, ok, tc.want, tc.found)
		}
	}
	structDefinitionFieldIndexCacheMu.RLock()
	_, ok := structDefinitionFieldIndexCache[def]
	structDefinitionFieldIndexCacheMu.RUnlock()
	if ok {
		t.Fatalf("small direct lookup should not seed shared field-index cache")
	}
}

func TestNewStructDefinitionValueOnlyBuildsLargeNamedFieldIndexMap(t *testing.T) {
	small := ast.StructDef(
		"Small",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "a"),
			ast.FieldDef(ast.Ty("i32"), "b"),
			ast.FieldDef(ast.Ty("i32"), "c"),
			ast.FieldDef(ast.Ty("i32"), "d"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if got := newStructDefinitionValue(small); got.NamedFieldIndices != nil {
		t.Fatalf("small named struct should use direct field lookup, got map with %d entries", len(got.NamedFieldIndices))
	}

	large := ast.StructDef(
		"Large",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "a"),
			ast.FieldDef(ast.Ty("i32"), "b"),
			ast.FieldDef(ast.Ty("i32"), "c"),
			ast.FieldDef(ast.Ty("i32"), "d"),
			ast.FieldDef(ast.Ty("i32"), "e"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	got := newStructDefinitionValue(large)
	if got.NamedFieldIndices == nil {
		t.Fatalf("large named struct should build field-index map")
	}
	if idx, ok := got.NamedFieldIndices["e"]; !ok || idx != 4 {
		t.Fatalf("large named struct index for e = (%d, %t), want (4, true)", idx, ok)
	}
}

func TestStructDefinitionNamedFieldIndexCachesLargeDefinitions(t *testing.T) {
	def := ast.StructDef(
		"Large",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "a"),
			ast.FieldDef(ast.Ty("i32"), "b"),
			ast.FieldDef(ast.Ty("i32"), "c"),
			ast.FieldDef(ast.Ty("i32"), "d"),
			ast.FieldDef(ast.Ty("i32"), "e"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	structDefinitionFieldIndexCacheMu.Lock()
	delete(structDefinitionFieldIndexCache, def)
	structDefinitionFieldIndexCacheMu.Unlock()

	got, ok := structDefinitionNamedFieldIndex(def, "e")
	if !ok || got != 4 {
		t.Fatalf("structDefinitionNamedFieldIndex(\"e\") = (%d, %t), want (4, true)", got, ok)
	}
	structDefinitionFieldIndexCacheMu.RLock()
	_, ok = structDefinitionFieldIndexCache[def]
	structDefinitionFieldIndexCacheMu.RUnlock()
	if !ok {
		t.Fatalf("large definition lookup should seed shared field-index cache")
	}
}
