# Compiled static-extern launcher (2026-07-11)

## Decision

Keep the generic static-launcher change. A fully compiled, fallback-free
application now uses the existing nil-interpreter `RegisterIn(nil, entryEnv)`
path even when it contains compiled Go extern bodies. Dynamic programs,
fallbacks, non-compiled functions, and non-seedable imports still use the
interpreter bootstrap path unchanged.

Previously, the static launcher selected a full
`interpreter.NewWithExecutor(...)` solely because `g.externBodies` was
non-empty. The generated externs are already compiled Go functions, and the
shared static registration/runtime path is explicitly nil-interpreter-capable.
The new path creates the standalone environment, registers print and OS args,
sets the executor kind, and runs the compiled entry with a nil interpreter.
This is a compiler-wide launch boundary, not an I/O, text, FASTA, or
benchmark-specific lowering.

## Evidence

All measurements use CPU `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, a
45-second process cap, and each external suite's verifier. The candidate and
restored baseline were each built once and launched ten times per application:

| Application | Baseline | Static-extern launcher | Change | Verification |
| --- | ---: | ---: | ---: | --- |
| I-Before-E | 0.134 s | 0.120 s | 10.4% faster | 10/10 each |
| Reverse Complement | 0.131 s | 0.117 s | 10.7% faster | 10/10 each |
| JSON control | 0.833 s | 0.772 s | 7.3% faster | 10/10 each |

Exact phase-allocation captures are attribution only. They confirm that the
same launch boundary is removed in both independent misses and in the control:

| Application | Prior bootstrap allocation | Static-extern launcher |
| --- | ---: | ---: |
| I-Before-E | 4,214,139 B / 12,621 allocs | 3,296,488 B / 10,927 allocs |
| Reverse Complement | 4,232,179 B / 12,691 allocs | 3,303,640 B / 11,009 allocs |
| JSON | 4,229,512 B / 12,867 allocs | 3,321,912 B / 11,213 allocs |

The roughly 0.9 MB / 1.7k-allocation reduction is the eliminated per-launch
interpreter construction and its dependent registration setup. It is separate
from the already-retained diagnostic-origin capacity reservation and does not
place any check on the bytecode VM hot path.

## Verification

- Generated-main coverage proves static host externs emit the nil-interpreter
  bridge path; dynamic launchers retain bootstrap generation.
- Static no-bootstrap fixture boundary and compiled-Go-extern tests pass.
- Bridge nil-environment/executor tests pass.
- I-Before-E, Reverse Complement, and JSON real external binaries all verify
  their output for the smoke run and every A/B timing launch.

No `able-stdlib` change is needed.

## Next recommendation

Refresh the fair pinned compiled Able-versus-Go scorecard for I-Before-E,
Reverse Complement, JSON, and Mandelbrot after this retained startup change.
That determines the real progress toward the compiler target and separates the
shared launcher gain from the remaining text, byte, and numeric application
costs. Use fresh rebuilt Go references and multiple verifier-backed Able
launches; profile another leaf only if it is newly material in two misses and
neutral on JSON.
