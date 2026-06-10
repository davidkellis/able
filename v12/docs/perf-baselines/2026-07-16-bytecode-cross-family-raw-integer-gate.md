# Bytecode Cross-Family Raw-Integer Gate — 2026-07-16

## Decision

Keep no additional VM, compiler, runtime, canonical-stdlib, application, or
benchmark-source change from this tranche. Fresh Word Frequency, Regex Set,
Regex Stream, and K-Nucleotide profiles do repeat
`bytecodeRawIntegerValueInfo(...)`, but a new generic carrier-interface trial
traded a Word Frequency regression for regex-only gains. It was fully removed.

The profiles expose one separate shared allocation boundary worth diagnosing
next: generic positional named-struct construction. No struct optimization is
authorized yet because these profiles do not establish value lifetime or
escape behavior.

## Method

The runtime benchmark harness loaded and typechecked each source program once,
warmed `main`, then measured only repeated `main` calls. Runs used canonical
external `able-stdlib`, `ABLE_SOURCE_ROOT_ONLY=1`,
`ABLE_BENCH_SKIP_TYPECHECK=0`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`.

CPU profiles were collected with both `ABLE_BYTECODE_TRACE` and
`ABLE_BYTECODE_STATS` disabled. Trace/stats attribution and allocation
profiles used separate processes so their mutex, map, and sampling work could
not become CPU candidates. K-Nucleotide used one measured call after warmup;
its long bounded run is profile evidence, not a normal test lane.

Clean CPU-profile measurements were:

| Program | Measured runtime | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: |
| Word Frequency, 5x | 1,078,624,643 ns/op | 47,465,364 | 631,165 |
| Regex Set, 5x | 985,233,109 ns/op | 28,391,668 | 401,221 |
| Regex Stream, 5x | 827,190,941 ns/op | 29,185,171 | 434,842 |
| K-Nucleotide, 1x | 37,875,470,123 ns/op | 1,314,299,464 | 23,497,358 |

These are attribution runs, not replacements for the promoted five-process
external scorecard.

## CPU attribution

`bytecodeRawIntegerValueInfo(...)` is genuinely material in every clean
profile, but its direct-caller mix remains heterogeneous:

| Program | Extractor flat CPU | Leading direct callers |
| --- | ---: | --- |
| Word Frequency | 4.83% | raw slot store 46%, same-type small integer pair 23%, typed pattern 15% |
| Regex Set | 2.64% | same-type pair 23%, inline coercion check 23%, raw store 15%, Array slot/index 15% |
| Regex Stream | 2.43% | generic integer value 40%, raw store 30%, same-type pair 20% |
| K-Nucleotide | 5.45% | same-type pair 33%, immediate compare 23%, inline coercion 17%, raw store 17% |

The surrounding walls also differ. Word Frequency has concrete HashMap lookup
and type-match work; both regex programs are dominated by canonical Array-slot
and NFA member traffic; K-Nucleotide is led by binary/bitwise, call/return, and
string-key lookup work. This agrees with the earlier carrier diagnostics and
does not reopen type-switch reordering, caller-local prebranches, raw-cell
preservation, or frame/return work.

## Rejected generic carrier-interface trial

The trial gave every existing VM-private raw-integer carrier one common
private extraction method. `bytecodeRawIntegerValueInfo(...)` first performed
a single interface assertion and retained the ordinary boxed-integer fallback.
It did not change storage, materialization, numeric width/suffix semantics, or
the public `runtime.Value` ABI. Focused raw-integer/binary/store/type guards
passed.

Three independent benchmark invocations, each averaging five measured calls,
were compared with three matching invocations of the preserved restored
binary:

| Program | Candidate mean | Restored mean | Result |
| --- | ---: | ---: | ---: |
| Word Frequency | 1,111,358,449 ns/op | 1,081,256,777 ns/op | 2.78% slower |
| Regex Set | 992,386,373 ns/op | 1,035,876,960 ns/op | 4.20% faster |
| Regex Stream | 849,177,754 ns/op | 858,432,322 ns/op | 1.08% faster |

Allocation counts were unchanged. The Word Frequency regression fails the
broad admission bar, so K-Nucleotide candidate timing was unnecessary and the
trial was removed. Focused guards passed again on the restored tree.

## Allocation attribution

Allocation profiles include process-global small-integer cache initialization;
that repeated startup entry is not a per-program candidate. Below it,
`runtime.NewStructInstancePositionalSized(...)` repeats through
`execStructLiteralNamedFast(...)` in all four programs:

| Program | Sampled alloc space | Sampled alloc objects |
| --- | ---: | ---: |
| Word Frequency | 17.43% | 9.02% |
| Regex Set | 10.74% | 5.25% |
| Regex Stream | 12.33% | 6.20% |
| K-Nucleotide | 39.66% | 22.75% |

That is a shared generic nominal boundary, not evidence for a named struct,
container, UTF-8 result, regex thread, or frequency-record special case. The
current profiles show allocation sites but cannot prove whether the values are
immediately deconstructed, stored, returned, aliased, or mutated.

## Next recommendation

Attribute definition identity and immediate lifetime/consumer shape at the
shared positional-struct literal boundary across these four programs. Use
temporary off-timing counters to classify field read/pattern deconstruction,
return, collection insertion, aliasing, mutation, and frame escape, then remove
the counters. Advance only if at least three unlike programs share a provably
non-escaping pattern that supports a general nominal scalar-replacement or
frame-owned transport rule. This is the next lead because it is a large shared
allocation wall after the raw-integer mechanism failed, while lifetime proof
is required to preserve nominal identity and mutation semantics.
