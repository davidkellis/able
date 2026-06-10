# Three-family fresh scorecard (2026-07-12)

## Method

This scorecard uses the combined fresh-reference adapter for Base64, JSON, and
Monte Carlo Pi. Go 1.26.4, Ruby 4.0.5, Python 3.14.5, compiled Able, and
bytecode Able each used three CPU-2-pinned verifier-backed processes with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap. Monte Carlo
Pi is verified statistically and has nondeterministic reference stdout; every
other row has a stable verified hash.

| Benchmark | Mode | Able (s) | Go ratio | Ruby ratio | Python ratio | Status |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| Base64 | compiled | 2.4900 | 1.02x | 0.77x | 0.62x | verified 3/3 |
| Base64 | bytecode | 3.0400 | 1.24x | 0.94x | 0.75x | verified 3/3 |
| JSON | compiled | 0.7967 | 0.55x | 0.48x | 0.31x | verified 3/3 |
| JSON | bytecode | 0.9100 | 0.62x | 0.55x | 0.36x | verified 3/3 |
| Monte Carlo Pi | compiled | 0.2100 | 1.00x | 0.13x | 0.13x | verified 3/3 |
| Monte Carlo Pi | bytecode | 2.6400 | 12.62x | 1.62x | 1.68x | verified 3/3 |

## Decision

Compiled Able meets the current 95%-of-Go target across this feature-diverse
slice. Bytecode meets or exceeds the current Ruby/Python target on Base64 and
JSON. Monte Carlo Pi bytecode is the sole miss, but Base64 is host byte/MD5
work and JSON is host parsing work; neither supplies a matching numeric loop
or a repeated VM helper. Keep no runtime, compiler, or `able-stdlib` change.
Do not turn the Monte Carlo Pi row into a benchmark-specific float/random fast
path.

## Next recommendation

Refresh the same three-family references for Mandelbrot and compare its
bytecode profile with Monte Carlo Pi only if both are still material misses.
Why: Mandelbrot is an independent numeric escape-time loop and is the correct
test for a shared float/control-flow cost; the current text/host controls are
already healthy. The work entails fresh Go/Ruby/Python rows, multi-run Able
compiled/bytecode rows, then bounded profiles only for a concrete helper that
repeats in both numeric applications and remains neutral on Base64 and JSON.
