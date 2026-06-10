# Bytecode wide-numeric five-application refresh

Date: 2026-07-21

## Decision

Keep no VM, compiler, canonical-stdlib, language, workload, reference, or WASM
change. Fresh current profiles confirm raw-integer extraction across Fixed
Width 128, Rational Series, Wide Integer Records, and the unrelated Reverse
Complement discriminator, but its concrete consumers differ and its general
carrier/extractor/store alternatives are already rejected by broad causal
wall-time evidence.

One new general primitive candidate advanced to implementation: direct casts
to language-reserved scalar primitives bypassed repeated runtime alias-target
canonicalization while user aliases and nominal types retained the existing
path. It passed cast and alias semantics and left allocation unchanged. It did
not pass the repeated broad wall-time gate. Eight-run arithmetic means were
favorable for the three wide targets but regressed Reverse Complement 2.17%; a
separate four-run CPU-0 subset regressed all three targets 2.66%-10.42%. The
candidate was fully removed because its result changed sign with workstation
contention and was not broad enough to retain.

## Reproducibility contract

A single current interpreter test binary was frozen before collection at
SHA-256
`6758a13355a1adeebe0984098679c3ad344b0f8cf1a8642694e873e3dd12d53e`.
Each application ran twice for main-only CPU profiles and twice without CPU
profiling for exact main-allocation counters. Every process used the canonical
external stdlib, `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB memory limit, and a
59-second test cap. Load, typecheck, lowering, and final forced collections
were outside the measured main window.

The three wide sources match the current five-run scorecard hashes. Distance
Field supplies float/native-call contrast; Reverse Complement supplies
primitive byte/Array/cast contrast. This prevents three related wide nominal
programs from satisfying breadth by themselves.

| Application | CPU runs | CPU mean | Merged samples |
| --- | --- | ---: | ---: |
| Fixed Width 128 | 7.304 s, 7.664 s | 7.484 s | 14.92 s |
| Rational Series | 3.591 s, 3.802 s | 3.696 s | 7.36 s |
| Wide Integer Records | 5.035 s, 5.283 s | 5.159 s | 10.25 s |
| Distance Field | 5.690 s, 5.659 s | 5.675 s | 11.32 s |
| Reverse Complement | 2.805 s, 3.167 s | 2.986 s | 5.96 s |

## Exact CPU intersection

The selection threshold was at least 1% flat CPU in at least three unlike
applications.

| Exact symbol | Breadth | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `(*bytecodeVM).runResumable` | 5 | 3.51%-8.56% | dispatcher parent already closed |
| `runtime.tryDeferToSpanScan` | 5 | 1.77%-7.91% | Go GC machinery |
| `bytecodeRawIntegerValueInfo` | 4 | 2.21%-4.19% | closed raw-integer family; consumers diverge |
| `internal/runtime/maps.ctrlGroup.matchH2` | 4 | 2.17%-7.21% | unrelated cache and boxing maps |
| `runtime.mapaccess2_faststr` | 4 | 1.24%-2.17% | unrelated type/member/environment maps |
| `appendSlotStackValueChecked` | 3 | 1.36%-2.01% | closed stack-carrier family |
| `execCastOpcode` | 3 | 1.01%-1.68% | parent; primitive target work advanced to A/B |
| `popCallFrameFields` | 3 | 1.17%-2.21% | closed call-frame family |
| `pushCallFrame` | 3 | 1.21%-1.94% | closed call-frame family |
| `aeshashbody` | 3 | 2.21%-3.71% | different string-key maps |
| `runtime.nextFreeFast` | 3 | 1.07%-2.08% | Go allocator machinery |

Raw-integer callers show why recurrence does not define one next operation:

- Fixed Width divides between checked UInt member operations, ordinary binary
  work, direct values, and coercion checks.
- Rational divides between casts, direct values, coercion checks, immediate
  comparisons, calls, and return work.
- Wide Records divides among signed/unsigned parsing, comparison, bitwise,
  member calls, and typed-pattern work.
- Reverse Complement is 68% immediate loop comparison and 24% primitive cast,
  with small byte-Array consumers.

The Go map matcher is likewise not one Able map. Fixed Width and Reverse are
dominated by boxed-integer maps, Rational divides among alias, generic-call,
bound-method, and string maps, and Wide Records divides among more than ten
type/member/environment/Array maps. AES hashing splits among member-method
keys in Rational, static/member/integer/environment keys in Wide Records, and
static-member identity keys in Distance Field.

## Exact main allocation

The two unprofiled allocation processes were stable to 16 bytes in Fixed
Width, Rational, and Distance; 328 bytes/eight objects in Wide Records; and
128 bytes/six objects in Reverse Complement.

| Application | Mean allocated bytes | Mean allocations | Mean frees | Mean GCs |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 1,348,450,616 | 31,911,044 | 30,753,250 | 40.5 |
| Rational Series | 105,204,664 | 1,012,479 | 922,593 | 6.0 |
| Wide Integer Records | 652,833,092 | 19,361,763 | 19,042,769 | 31.5 |
| Distance Field | 368,035,968 | 26,000,121 | 25,364,744 | 20.5 |
| Reverse Complement | 266,698,672 | 4,069,463 | 1,840,331 | 7.0 |

The previous exact allocation-owner gate remains consistent with these
counters: Fixed Width is dominated by big/wide integer values and positional
results; Rational by call environments and Rational construction; Wide
Records by parsing and wide values; Distance by raw-float transport; Reverse
by host-byte Array conversion and snapshots. No concrete allocator is shared
across three unlike applications.

## Primitive cast-target candidate

`canonicalRuntimeTypeExpression` accounts cumulatively for 1.27%-1.61% of the
three wide profiles and 2.85% of Reverse Complement. Most time lies in the
cached “does this expression reference an alias?” map lookup. Because Able's
lowercase scalar primitive names are reserved, a direct primitive cast target
cannot be a user alias or nominal binding. The candidate therefore returned
the original type expression for direct reserved scalar targets and preserved
canonical alias/nominal handling for every other type expression.

Focused numeric-cast, user-alias, and bytecode-cast tests passed. Exact A/B
allocation means differed by less than 0.0002%, confirming that this was a CPU
candidate rather than an allocation change.

The workstation was volatile, so the gate alternated order and used arithmetic
means, then tightened the final six runs to adjacent per-program A/B pairs.
The complete cohort has eight processes per binary for the primary programs
and Reverse, and six per binary for the unreached Distance guard. The final
four primary/Reverse runs additionally used the catalog's CPU-0 affinity.

| Application | Complete baseline mean | Complete candidate mean | Change | CPU-0 four-run change |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 8.553 s | 8.434 s | -1.39% | +10.42% |
| Rational Series | 4.404 s | 4.085 s | -7.24% | +2.66% |
| Wide Integer Records | 5.623 s | 5.409 s | -3.81% | +5.47% |
| Reverse Complement | 3.236 s | 3.306 s | +2.17% | -0.70% |
| Distance Field | 5.717 s | 5.637 s | -1.40% (six runs) | not extended |

The sign reversal is larger than the selected 1.27%-2.85% cumulative wall.
Keeping the change would mean trusting favorable contention in the complete
mean while ignoring an unrelated guard regression and the opposite CPU-0
target cohort. It was therefore reverted.

## Verification

The restored source passes focused cast/raw-integer coverage and the complete
bytecode family:

```text
go test ./pkg/interpreter -run 'TestCast|TestBytecodeVM_.*Cast|TestBytecodeVM_RawInteger' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.064s

go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  23.955s
```

The frozen binaries, raw profiles, and A/B reports are temporary and are
removed after frontier verification.

## Next recommendation

Refresh the `bytecode-float-numeric` ownership group across Distance Field,
RMS Norm, Mandelbrot, and Monte Carlo, with Matrix Multiply and one
non-floating call/Array discriminator.

Why: it is now the next-largest bytecode group at 15.317 target-excess
seconds. The latest wide refresh confirms that float work separates cleanly
from wide-integer work, while the current float evidence predates several
recent VM changes. The expanded group distinguishes native geometry,
reductions, fractal branching, stochastic arithmetic, and matrix loops.

What it entails: collect two current bounded main-only CPU and exact-allocation
processes per application, reconcile exact float/native/call owners, and admit
only a new concrete primitive operation present in at least three unlike
programs. Previously rejected raw-float carriers, sidecars, typed operand
lanes, native scalar returns, and producer-fusion designs stay closed unless a
new fact invalidates their causal wall-time failures. Any candidate must pass
repeated order-balanced means across all admitted targets and unrelated
integer/text guards. No Matrix-specific fusion, stdlib named-type shortcut,
benchmark opcode, or WASM work is admissible.
