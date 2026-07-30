# Parser Go binding contract retained

Date: 2026-07-30

## Decision

Retain the grammar-root Go module as the sole module boundary for the v12
Tree-sitter grammar. Remove the obsolete nested `bindings/go/go.mod`, retain
the existing official `go-tree-sitter v0.24.0` dependency, record the complete
module sums, and run the binding test explicitly from the routine v12 test
runner.

This is a tooling and module-contract correction. It changes no grammar,
generated parser, binding implementation, Able parser behavior, AST,
interpreter, compiler, runtime, VM, canonical stdlib, language, benchmark,
fixture, frozen workspace, or WASM source.

## Cause

`tree-sitter.json` explicitly advertises the Go binding. The grammar-root
module and generated test agree on its identity:

- root module: `github.com/davidkellis/able`;
- binding package: `github.com/davidkellis/able/bindings/go`; and
- Go Tree-sitter runtime: `github.com/tree-sitter/go-tree-sitter v0.24.0`.

The nested module was an older, contradictory scaffold:

- module: `github.com/tree-sitter/tree-sitter-able`; and
- runtime: `github.com/smacker/go-tree-sitter`.

That nested boundary hid `bindings/go` from the grammar-root module and made
the generated test imports impossible to resolve as declared. Tree-sitter's
current initialization contract places `go.mod`, `bindings/go/binding.go`,
and `bindings/go/binding_test.go` under one grammar-root module. Its official
Go runtime is `tree-sitter/go-tree-sitter`.

## Retained changes

- deleted only `v12/parser/tree-sitter-able/bindings/go/go.mod`;
- retained the root module path and `go-tree-sitter v0.24.0`;
- added the tidy-required indirect `go-pointer v0.0.1` declaration;
- added the 36-line root `go.sum`; and
- added an explicit, one-minute
  `go test -timeout 1m -count=1 ./bindings/go` gate to
  `v12/run_all_tests.sh`.

The explicit package path is intentional. A broad `go test ./...` can return
success with a warning when a nested module hides every package; testing
`./bindings/go` instead makes that module-boundary regression fail.

## Verification

All build and scanner state lived under disk-backed `/var/tmp`.

| Check | Result | Wall | Peak RSS |
| --- | --- | ---: | ---: |
| Isolated grammar-root proof, cold Go state | pass | 21.22s including tidy/list/test | not recorded |
| Root `go mod tidy -diff` | empty | included below | included below |
| Root `go mod verify` | pass | included below | included below |
| Standalone binding, cold Go state | pass | 15.87s | 196,444 KB |
| Standalone binding, Go 1.26.5 | pass | 15.66s | 196,260 KB |
| Official Able parser packages | pass | 12.81s | 318,988 KB |
| Test-inclusive `govulncheck v1.6.0` | no vulnerabilities | 16.35s | 428,080 KB |
| Complete `./run_all_tests.sh` | pass | 973.55s | 4,651,560 KB |
| `bash -n v12/run_all_tests.sh` | pass | not measured | not measured |
| Scoped `git diff --check` | pass | not measured | not measured |

The complete runner passed every preflight, the standalone binding gate,
non-compiler packages, all 34 compiler batches, and the final bytecode fixture
pass with zero swaps. Its raw output SHA-256 was
`72dc51969c95d617f8e1c3d2bbb9e1552964b75e9c182922205dbfe63b6630f1`.
The known compiler batch aggregates that exceeded one minute retain their
prior named-test replays, whose individual maxima are below one minute.

The standalone security scan matched three root packages: the binding, its
external test package, and its test binary. It scanned:

- `github.com/davidkellis/able`;
- `github.com/mattn/go-pointer v0.0.1`;
- `github.com/tree-sitter/go-tree-sitter v0.24.0`; and
- the Go 1.26.5 standard library.

It reported zero vulnerabilities. Raw scan SHA-256:
`ac2fbe8e62be0ddae77b8fc19198ec6daf9d00145e24d2fb03f32fd040a92811`.

Retained module identities:

- `go.mod`:
  `d5c63a254931cd83e2b0cc415137cd076106facb235c0289437e9da2590bdab0`;
- `go.sum`:
  `a7d2dd524ce7f8e47eb33d9aa480d5011628376107c15e3a415cfef44d22fc44`.

Unchanged semantic-source identities:

- `bindings/go/binding.go`:
  `d92fb926123e23f3548e07257fe977d7b788b11354892414d43e35448c9b43c7`;
- `bindings/go/binding_test.go`:
  `36b9d9d04498363672e6202b84df6fccf1685928d06103d981ac1cbfc3eb5c1`;
- `src/parser.c`:
  `054369173a78160f15ebef87bced7c4717a4aaf85cfb04ded76eb6ab1ed4c6e3`.

No current performance closure is invalidated because no production,
semantic, dependency-version, benchmark, or language input changed.

## Cleanup

The isolated grammar copy, Go caches, downloaded scanner/toolchain, and raw
scanner output were removed after preserving the readable and
machine-readable evidence. The full-suite workspace and its final 148 KiB
Python cache were also removed. Total reclaimed task state is 3,117,804 KiB
(2.97 GiB). The final project cleanup
dry run is empty. No build state was placed in RAM-backed `/tmp`. All 34
deferred WASM files reproduce their prior line, byte, and SHA-256 identities.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: the dirty worktree now combines the unchanged 34-path deferred WASM
boundary with the retained post-publication security records, parser module
correction, test-runner gate, and handoff updates.

What it entails: expand every untracked path, classify the complete path set,
verify the parser module and all changed JSON/Go/shell/docs files, reproduce
the 34-path deferred complement, and record exact identities. Do not modify
WASM or resume unchanged performance experiments.

Why it matters: an exact review boundary prevents deferred work from entering
a future consolidation and makes the small parser correction independently
reviewable.

## References

- Tree-sitter parser initialization and generated binding layout:
  <https://tree-sitter.github.io/tree-sitter/cli/init.html>
- Tree-sitter's official Go runtime:
  <https://tree-sitter.github.io/tree-sitter/6-contributing.html>
