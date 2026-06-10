package language

// #cgo CFLAGS: -std=c11 -fPIC
// #include "../../../../../parser/tree-sitter-able/src/parser.c"
// #if __has_include("../../../../../parser/tree-sitter-able/src/scanner.c")
// #include "../../../../../parser/tree-sitter-able/src/scanner.c"
// #endif
import "C"

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	"unsafe"
)

// generatedGrammarRevision deliberately changes whenever the checked-in
// tree-sitter assets regenerate. Go's cgo cache does not track the parser.c
// included above, so this source-level stamp forces ordinary Go builds to
// relink the language with the matching generated grammar.
const generatedGrammarRevision = "2026-07-14-source-reexports"

// Able returns the tree-sitter language for Able v12.
func Able() *sitter.Language {
	return sitter.NewLanguage(unsafe.Pointer(C.tree_sitter_able()))
}
