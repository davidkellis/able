The v12 Able stdlib is being rebuilt incrementally.

Current state: foundational modules are restored (core/errors, core/interfaces, core/options, core/iteration, core/numeric, concurrency/awaitable + await + future + channel + mutex, collections/array + enumerable + range + list + linked_list + vector + deque + queue + lazy_seq + hash_map + set + hash_set + bit_set + heap + tree_map + tree_set, text/string). The previous full implementation now lives under `v12/stdlib/quarantine/`.

Reintroduce modules one at a time from the quarantine copy into `v12/stdlib/src/`, fix parser/typechecker issues, add minimal tests (see `v12/stdlib/tests`), and keep the working set green (run `go test ./...` plus `./run_stdlib_tests.sh`) before adding more files.
