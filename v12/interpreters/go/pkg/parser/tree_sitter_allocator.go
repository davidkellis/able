package parser

/*
#include <stddef.h>

extern void ts_set_allocator(
	void *(*new_malloc)(size_t),
	void *(*new_calloc)(size_t, size_t),
	void *(*new_realloc)(void *, size_t),
	void (*new_free)(void *)
);

static void able_restore_tree_sitter_default_allocator(void) {
	ts_set_allocator(NULL, NULL, NULL, NULL);
}
*/
import "C"

import "sync"

var restoreTreeSitterDefaultAllocatorOnce sync.Once

// The Go binding installs C-to-Go-to-C allocator callbacks even when callers
// do not supply a custom allocator. Able does not replace Tree-sitter's
// allocator, so restore the library's documented native defaults before the
// first parser is created and avoid those callbacks on every parse allocation.
func restoreTreeSitterDefaultAllocator() {
	restoreTreeSitterDefaultAllocatorOnce.Do(func() {
		C.able_restore_tree_sitter_default_allocator()
	})
}
