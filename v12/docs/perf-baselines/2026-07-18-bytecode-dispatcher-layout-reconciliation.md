# Bytecode Dispatcher Layout Reconciliation

Date: 2026-07-18

## Decision

Keep no bytecode VM, compiler, language, benchmark, fixture, or canonical
`able-stdlib` change from this tranche.

The exact rejected `JumpIfSlotNotNil` opcode candidate was reconstructed from
the preserved session patch and reproduced both recorded CLI fingerprints:

- baseline: `b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`;
- direct-opcode candidate:
  `b70d98ada7d7f58903f9591a810d005807ebc712f105b13cfe9e955d131c2def`.

Profiles and symbol layout confirm that the 9% Mandelbrot loss was not caused
by executing material slot-not-nil work. Adding the primary opcode enlarged
and moved generic VM code, changing performance in both directions across
unrelated programs with essentially unchanged allocation counts.

One general layout-stabilizing design was tested. It reused the existing
`JumpIfNotNil` dispatch case, marked its direct slot operand in existing
instruction metadata, and kept the rare quickening pass out of line. This
restored the primary dispatcher and jump-table sizes, and removed the
Mandelbrot regression, but ten-pair expansion showed repeatable 3%-4%
regressions in Reverse Complement, Option/Result Config, and Word Frequency.
The design therefore only changed which applications lost and was reverted.

No WASM work was performed.

## Exact layout attribution

All three CLI binaries were built from otherwise identical source and Go
toolchain state. Sizes and addresses below come from `go tool nm -size -sort
address` on those preserved binaries.

| Symbol/layout item | Baseline | Direct opcode | Existing-dispatch design |
| --- | ---: | ---: | ---: |
| `finalizeBytecodeProgramMetadata` size | 2,501 B | 2,735 B | 2,533 B |
| Mandelbrot fused-float helper address | `0xf402e0` | `0xf403c0` | `0xf40300` |
| `execJumpIfNotNil` size | 280 B | 308 B | 308 B |
| `runResumable` address | `0xf9a240` | `0xf9a340` | `0xf9a280` |
| `runResumable` size | 35,909 B | 35,941 B | 35,909 B |
| primary jump table | 1,144 B | 1,152 B | 1,144 B |

The direct opcode added 32 bytes to the primary dispatcher, eight bytes to its
jump table, and 234 bytes to the inlined metadata finalizer. Existing helper
addresses after that finalizer moved by 224 bytes before the new out-of-line
slot helper was reached. The existing-dispatch design kept the primary switch
and jump table at their baseline sizes; its CLI SHA-256 was
`41c216b4dea387e1bdb4a52d4418a9d393e0ad30db6e0ed0ee4703c6721003e0`.

This does not prove one cache-line alignment is intrinsically good or bad.
It does establish the engineering constraint: small Go source changes in or
before this 35.9 KB dispatcher can move unrelated hot code and materially
alter whole-program timing even when the new Able operation is almost absent.

## Main-only diagnostic profiles

The exact baseline and direct-opcode test binaries each ran one warmed,
main-only call for Mandelbrot, Reverse Complement, Option/Result Config, Word
Frequency, Distance Field, and RMS Norm. Runs used CPU 0, `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, the canonical external stdlib, and the catalog's
source/input setup. These single captures are diagnostic attribution, not
selection evidence.

| Application | Baseline ns/op | Direct opcode ns/op | Change |
| --- | ---: | ---: | ---: |
| Mandelbrot | 6,488,897,587 | 7,078,833,590 | +9.09% |
| Reverse Complement | 6,153,396,403 | 6,257,136,551 | +1.69% |
| Option/Result Config | 836,368,783 | 839,952,359 | +0.43% |
| Word Frequency | 1,313,982,494 | 1,224,745,307 | -6.79% |
| Distance Field | 5,788,926,707 | 5,346,545,179 | -7.64% |
| RMS Norm | 4,276,592,186 | 4,361,680,888 | +1.99% |

The direct candidate adds only six allocations to Mandelbrot's 76,303,222,
two to Reverse Complement and Distance Field, and none to the other three at
the displayed precision. The mixed timing is therefore not an allocation or
GC-work explanation. CPU profiles likewise keep each application's existing
dominant leaves; only sample distribution and dispatcher ancestry move.

The existing-dispatch design's one diagnostic Mandelbrot run was
6,422,047,706 ns/op, 1.03% faster than the diagnostic baseline and far from
the direct opcode's 9.09% regression. That result admitted the design to the
repeated process gate but was not used to retain it.

## Repeated causal gate

Normal bytecode processes used the exact baseline and existing-dispatch CLI
binaries on CPU 0 with the same memory/GC contract and a 55-second cap. Every
run was paired in alternating order, every stdout passed the application's
public Ruby verifier, and every workstation outlier remains in the arithmetic
mean.

The initial gate ran five pairs for all six programs (60/60 verified). Because
the three apparent regressors were short or near the rejection boundary,
Reverse Complement, Option/Result Config, and Word Frequency received five
additional pairs. The final ledger contains 90/90 verified processes.

| Application | Samples/variant | Baseline mean | Existing-dispatch mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Mandelbrot | 5 | 6.840 s | 6.670 s | -2.49% |
| Reverse Complement | 10 | 7.077 s | 7.345 s | +3.79% |
| Option/Result Config | 10 | 0.877 s | 0.912 s | +3.99% |
| Word Frequency | 10 | 1.559 s | 1.606 s | +3.02% |
| Distance Field | 5 | 5.822 s | 5.540 s | -4.84% |
| RMS Norm | 5 | 4.650 s | 4.584 s | -1.42% |

The unweighted sum improves 0.63%, but three of six unlike applications
regress after expansion. That fails the explicit all-guards-neutral-or-better
bar and would again risk trading real-program performance for selected wins.
All samples and hashes are retained in
`2026-07-18-bytecode-dispatcher-layout-stabilized-causal-ab.tsv`.

## Restoration and verification

- The direct opcode, existing-dispatch marker, rare quickener, helper, and
  tests are fully removed.
- `rg` finds no `JumpIfSlotNotNil`, `QuickenSlotNotNil`, or slot-not-nil helper
  diagnostic in production/test source.
- The restored CLI exactly reproduces baseline SHA-256
  `b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`.
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passes in 30.391 seconds on the restored source.
- The canonical stdlib source state is unchanged from
  `2026-07-18-post-slot-not-nil-bytecode-stdlib-source-state.json`; this
  tranche made no change in the external repository.
- Preserved evidence is compact; binaries, raw profiles, object dumps, and
  scratch stdout/stderr are cleanup-only artifacts.

## Next recommendation

Run a full-selected-suite cold-opcode census before attempting another VM
optimization, then evaluate a stable two-tier dispatcher only if the census
finds a substantial globally cold tier.

Why: `runResumable` is a 35.9 KB primary dispatcher and contributes material
flat cost across all six profiles. Adding one rare case was enough to move
whole-program performance by 3%-9%, while merely hiding that case behind an
existing opcode still exchanged wins among programs. The next defensible move
is to shrink and stabilize the shared hot dispatcher based on aggregate
language-feature usage, not to add another isolated fusion or tune one helper's
address.

What it entails: collect bounded opcode counts for every selected bytecode
application, classify an opcode as cold only when it is immaterial across the
entire suite, and estimate the primary-switch/jump-table reduction available
from moving those cases to one out-of-line secondary handler. Prototype one
generic partition only if it removes a meaningful amount of primary code.
Then run order-balanced, verifier-backed guards across text, containers,
numeric, matching/error, concurrency, and iterator workloads, followed by the
complete selected bytecode scorecard. Revert if any family consistently
regresses. Continue to defer WASM.
