# Bytecode Runtime Opcode Census Closure (2026-07-12)

## Decision

Keep no VM, compiler, runtime, benchmark, fixture, or `able-stdlib` change.
The normal-build opcode census closes the stale follow-up after the dynamic
integer-box-cache result: it does not expose a new concrete VM leaf shared by
independent material misses. It also confirms that the prior recommendation to
start a canonical runtime-value-carrier design was stale—the complete design,
compiler-boundary audit, reachability sweep, and reverted broad `print` gate
are already recorded in `v12/design/canonical-runtime-value-architecture.md`.

Do not turn a common parent opcode into a candidate. `LoadSlot`, `Jump`, and
`Pop` occur in every ordinary program by design. A candidate needs the same
slow implementation leaf in independent applications, not merely the same
bytecode instruction name.

## Method

The existing `ABLE_BYTECODE_STATS=1` observer was enabled for a fresh
post-warmup one-call `BenchmarkBytecodeProgramRuntime` process. It records
normal bytecode operation and dispatch counters; production builds leave the
observer disabled. Every process used the canonical external stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, CPU 2 affinity, and a strict
60-second process guard. These are counts, not timing data: the observer uses
atomics and intentionally changes runtime cost.

| Application | Result | Relevant post-warmup evidence |
| --- | --- | --- |
| Base64 control | completed | 1,000,000 `CallMemberArraySlot`; 999,999 existing Array-member fast hits. Its normal CPU profile remains host codec/allocation work. |
| I-Before-E | completed | 174,454 named calls, with 172,826 inline-call hits; 347,653 member-cache hits versus 11 misses. |
| PiDigits control | completed | 439,905 named calls, with 438,910 inline-call hits; 406,414 member-cache hits versus 30 misses. Its normal profile remains BigInt host work. |
| Reverse Complement | completed | 12,133,703 `CallMemberArraySlot`; 12,133,687 Array-member fast hits. Its normal profile remains byte-array construction/boxing/stack work. |
| K-Nucleotide | counter run reached guard | Do not treat absent counters as zero. Its bounded normal CPU profile remains call-name/inline-return/raw-u64/map work and its cache-reuse probe is 95.77% dynamic-cache hits. |
| Mandelbrot | counter run reached guard | Do not treat absent counters as zero. Its bounded normal CPU profile remains the separately exhausted raw-float/condition-jump lane. |

Raw completed reports are retained in the sibling
`2026-07-12-bytecode-opcode-census/` directory. The K-Nucleotide and
Mandelbrot reports are deliberately absent rather than generated after a
longer run: extending an observer beyond the project's one-minute test rule
would not make its distorted elapsed time useful.

## Interpretation

The most frequent application-specific operations are already successful fast
dispatches, not misses:

- Base64 and Reverse Complement use the same generic Array-member fast path
  almost exclusively, yet their normal profiles are independent host-codec and
  byte/boxing/stack families.
- I-Before-E and PiDigits use the shared inline-call/member-cache machinery
  almost exclusively, yet their normal profiles are independent text/file and
  BigInt families.
- Reverse Complement's 30.4M `LoadSlot` operations and the other rows'
  `LoadSlot` counts are transport volume, not a proven shared cost; raw-slot,
  call-name, return, and cache-policy variants have already failed broad
  verifier-backed gates.

The observer therefore adds negative evidence: a further array-member,
inline-call, lookup-cache, or stack micro-optimization would optimize a path
that is already taking its existing fast implementation, or would re-open a
rejected broad candidate. The two capped counter rows do not weaken that
decision because their normal profiles already identify different, previously
tested implementation families.

## Next recommendation

Do not take another unchanged-source performance micro-tranche. Why: the full
22-application ledger, current CPU profiles, cache-reuse probe, dynamic-carrier
audit, and this census have exhausted the available broad candidates without a
repeated eligible leaf. The next work should return to language/runtime feature
completion; when a semantic change creates a new shared execution boundary,
add it to the cross-language application/fixture matrix and profile it against
the full guard set before optimizing it. This is necessary to avoid repeatedly
measuring the same disjoint paths or encoding an artificial benchmark rule.
