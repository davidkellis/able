# Able v10 Spec TODOs

- [x] Integrate pattern + break semantics from `design/pattern-break-alignment.md` into the canonical specification (capture optional break label/value, literal/array/typed pattern behaviour, and shared error messages).
- [x] Document struct pattern destructuring rules once Go evaluator gains struct literal/member support and fixtures are in place.

- [ ] Document functional update semantics for structs in examples after generics support lands. This is a v11 feature.
- [ ] Define `type` alias declarations (syntax, generics, evaluation) for the next spec revision. This is a v11 feature.
- [ ] Specify the safe member access operator (`?.`) and update AST/runtime semantics in the next spec revision. This is a v11 feature.
