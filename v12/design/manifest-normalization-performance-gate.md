# Manifest normalization performance gate

## Purpose

Manifest Normalization strengthens the portable interaction between file/text
processing, captured callables, and recoverable `Option`/`Result` handling. It
is intentionally a serial batch transformation rather than another router,
Regex parser, map aggregation, or concurrency workload.

The application reads 32 service manifests for 128 rounds. Twenty-four records
per round validate, optional owners receive region defaults, and a captured
normalizer emits environment buckets and checksums. Eight malformed records per
round take the recoverable error path. Able, Go, Python, and Ruby share the
input, algorithm, output, and verifier.

## Candidate admission lesson

The compiled main profile makes `String.to_builtin` material in a third unlike
application after K-Nucleotide and Policy Record Dispatch. That is a valid
evidence-invalidation trigger, but it does not make every plausible change
worth retaining.

The tested generic change avoided eager `fmt.Sprintf` construction for each
successful byte and preserved indexed diagnostics on failure. Repeated broad
measurements had mixed signs:

- K-Nucleotide improved about 1.5%;
- Policy Record Dispatch regressed about 2.9%;
- Manifest Normalization regressed about 1%.

The change and its generated-source test were therefore removed. Exact-leaf
breadth admits an experiment; it does not override the broad guard.

The bytecode main profile adds no new candidate. Its material descendants are
the already-reviewed call-frame, raw-integer, slot, member-cache, return, Go
map, and GC families. Aggregate dispatcher ancestry is not one semantic
optimization target.

## Consequence

Retain the portable application and current measurements. Retain no compiler,
runtime, VM, stdlib, language, or WASM performance code. The next application,
if coverage work continues, should target a remaining depth-two concurrency
triple with a shape unlike the existing routers and worker callback pipelines;
candidate admission remains independent.
