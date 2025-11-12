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

// Able returns the tree-sitter language for Able v10.
func Able() *sitter.Language {
	return sitter.NewLanguage(unsafe.Pointer(C.tree_sitter_able()))
}
