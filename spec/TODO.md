# Able v10 Spec TODOs

- [x] Integrate pattern + break semantics from `design/pattern-break-alignment.md` into the canonical specification (capture optional break label/value, literal/array/typed pattern behaviour, and shared error messages).
- [x] Document struct pattern destructuring rules once Go evaluator gains struct literal/member support and fixtures are in place.
- [x] Add v10 prose covering the shared executor contract (`proc_yield`, `proc_flush`, cancellation guarantees) now that Go and TypeScript runtimes align.
- [x] Capture method-set where-clause obligations in the spec, alongside fixtures, so higher-order interface calls (e.g., wrapper helpers) reflect the enforced diagnostics.
- [x] Clarify typed-pattern assignment semantics in the spec versus the warn-mode checker (mismatches yield runtime `Error` values while the checker remains advisory).
- [ ] Specify the `able.text.regex` module, including default code-point semantics, optional grapheme-aware execution, result types, and error reporting.
- [ ] Document the standard library `String`, `Grapheme`, and iteration helpers (byte, char, grapheme views) together with the byte-oriented indexing rules referenced by the spec.

See `spec/TODO_v11.md` for backlog items targeted at the next spec cycle.
