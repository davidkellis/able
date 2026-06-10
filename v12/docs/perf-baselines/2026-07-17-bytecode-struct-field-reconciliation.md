# Bytecode named-struct field reconciliation — 2026-07-17

## Decision

Close the current named-struct field-index candidate and keep no production
change. Exact main-only counts confirm that all three target programs use the
same small-definition comparison machinery, but their successful field reads
do not form a broad three-program optimization shape. A genuinely shared
duplicate callable-field check was then removed experimentally; it preserved
semantics and allocation shape but regressed Iterator Collect by 6.80%, so it
was fully reverted.

No compiler, stdlib, fixture, language, or tree-walker change is part of this
tranche. Temporary counters and binaries are removed after this record is
written.

## Exact field census

An opt-in diagnostic reset after benchmark warmup and counted only measured
`main()` execution. Each program ran once in a separate process with the
canonical external `able-stdlib`, CPU 0, `GOMAXPROCS=1`, `GOGC=50`, and
`GOMEMLIMIT=1GiB`. The diagnostic classified definition identity, field,
index, operation, storage, lookup path, result, and value carrier.

| Workload | Material successful field work | Material misses |
| --- | --- | --- |
| Unicode Scalar Pipeline | `RawStringCharsIter`: 1,769,488 reads each of `bytes[0]`, `offset[1]`, and `len_bytes[2]`, plus 1,769,488 `offset[1]` writes; 3,538,976 successful i32 consumers | 8,196 positional direct-small misses, principally method probes on `StringBuilder` |
| Run-length encode | `RawStringCharsIter`: 960,048 reads each of `bytes[0]`, `offset[1]`, and `len_bytes[2]`, plus 960,048 `offset[1]` writes; 1,920,096 successful i32 consumers; 329,745 `StringBuilder.buffer[0]` reads | 659,500 positional direct-small misses, principally `StringBuilder.push_char` and `push_string` method probes |
| Iterator Collect | 62,000 successful callable reads of `__able_iterator_controller.yield[0]` | 80,052 positional direct-small misses, principally `LinkedListIterator.next` and `LinkedList.push_back` method probes |

All material positional accesses used the existing one-to-four-field direct
comparison path. None used `StructDefinitionValue.NamedFieldIndices` or the
shared large-definition map cache. All successful `offset` and `len_bytes`
values were ordinary small `runtime.IntegerValue` carriers. Unicode and
Run-length share the same successful iterator definition and fields, but
Iterator Collect does not. Conversely, all three repeat absent-field method
probes, across `StringBuilder`, `String`, `LinkedList`, and
`LinkedListIterator` definitions.

This rejects another small-definition map or inline-name cache: those shapes
were already slower in earlier tranches, and the present census has no map
misses to repair. It also rejects transporting fixed indices only into the
canonical string-iterator fast path because that would benefit one definition
in two related text programs rather than multiple definitions in three unlike
programs.

## Rejected generic candidate

`execCallMember(...)` first calls
`execCallMemberStructCallableField(...)` for a non-nil struct receiver. When
that check finds no callable field, the old control flow immediately calls
`bytecodeCanDirectMemberCall(...)`, which repeats the same named-field lookup.
The duplicate is exact in the census: for example, 32,004
`LinkedListIterator.next` call probes produce 64,012 field reads, while 174,848
Run-length `StringBuilder.push_char` call probes produce 349,703 reads.

The candidate remembered only that same-dispatch result and skipped the second
lookup for that non-nil receiver. It did not alter the separately cached
`next` precheck, cache keys, callable-field precedence, dynamic field values,
method rebinding, aliases, storage representation, or either interpreter's
language semantics. Focused callable-field and member-cache tests passed.

Repeated independent processes alternated order, retained every workstation
outlier, and used one measured call after warmup:

| Workload | Pairs | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Unicode Scalar Pipeline | 5 | 3.324677 s | 3.294812 s | 0.90% faster; noisy |
| Run-length encode | 10 | 991.028 ms | 941.868 ms | 4.96% faster, including the retained 1.370 s baseline outlier |
| Iterator Collect | 5 | 387.961 ms | 414.325 ms | 6.80% slower; reject |

Iterator Collect had identical `8,671,584 B/op` and `192,911 allocs/op` in
every baseline and candidate process, so its regression is not an allocation
artifact. Unicode and Run-length allocation counts varied only by the existing
small setup noise. The stable Iterator loss crosses the 5% broad guard and
rejects the candidate without spending more time on zero-exposure Array or
Split/Join controls. The source was restored before final verification.

## Closure

The field family is real but currently divides into two shapes:

- successful raw string-iterator state access, material in two related text
  programs; and
- absent callable-field probes before method dispatch, present across all
  three but sensitive to the attempted runtime control-flow rewrite.

Do not retry small-definition maps, inline cached field names, direct-small-i32
extraction, canonical string-iterator fixed-index helpers, storage-shape
bypasses, or the same-dispatch boolean shortcut without new cross-program
evidence.

## Next recommendation

Test lowering-time negative field proofs for statically typed member-call
sites, first across Run-length, Iterator Collect, Document Audit, and Lexical
Rollup.

Why: the exact census found repeated immutable facts—method names such as
`push_char`, `next`, and `push_back` are absent from their receiver
definitions—but the rejected runtime branch changed hot call-loop layout.
Lowering already transports positive definition/index plans for ordinary
member reads and writes. A negative proof could skip both callable-field scans
without adding a runtime map or a named-container special case, and the two
additional member-heavy applications can establish whether the category is
material beyond this small cohort.

What it entails: count candidate call sites by statically proven receiver
definition and field absence; require material exposure in at least three
unlike programs and multiple nominal definitions; then, only if admitted,
attach immutable negative-field metadata to the existing per-instruction
nominal member plan. Runtime validation must still confirm the receiver
definition and must fall back for dynamic/unknown receivers, callable fields,
aliases, interfaces, rebinding, and user structs. Compare repeated independent
processes with every outlier retained, and continue to defer WASM.
