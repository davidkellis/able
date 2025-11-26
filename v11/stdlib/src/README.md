The v11 Able stdlib is being rebuilt incrementally.

Current state: foundational modules are restored (core/errors, core/interfaces, core/options, core/iteration, core/numeric, collections/array + enumerable + range + list + linked_list + vector + deque + queue + lazy_seq + hash_map + set + hash_set + bit_set + heap + tree_map + tree_set, text/string). The previous full implementation now lives under `v11/stdlib/quarantine/`.

Reintroduce modules one at a time from the quarantine copy into `v11/stdlib/src/`, fix parser/typechecker issues, add minimal tests (see `v11/stdlib/tests` and the TS ModuleLoader stdlib tests), and keep the working set green (run targeted `bun test` for stdlib tests and `go test ./...`) before adding more files.
