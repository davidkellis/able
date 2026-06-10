# K-Nucleotide Bytecode Selection Eligibility

## Decision

Keep K-Nucleotide bytecode in the reviewed strict-selection manifest. Five
fresh independent Able processes completed under the normal 90-second and
1-GiB guards, and the external verifier accepted every captured output. This
closes the only historically incomplete selected row before full cohort
collection.

This is an eligibility result, not a performance success or a promoted
scorecard. The bytecode interpreter remains much slower than both comparison
interpreters on this application.

## Source and contract identity

The retained five-run Python/Ruby reference artifact was reusable because its
foreign source hashes still matched the files in the external benchmark repo:

- Python: `d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806`
- Ruby: `de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745`

The canonical stdlib snapshot also remained identical: 69 Able source files,
tree hash
`785a6fd058c179379b1a153529fb340151a11b96d9014394cc40dbd87e1882ab`,
Git head `219eff222c28406487231713753641bc49ee5b9a`, with the same visible
dirty-state flag. The measured Able source hash was
`933749cb33f84a88274e010f7459d027be839a42162f0d559eca8a1920aa8a2a`.

The run retained the complete benchmark contract:

- input `knucleotide-input.fasta`, SHA-256
  `5d36fae47c10baed9a3aec44a559ae55daf0a159fb8b0b61bf5cae5c67a1ee30`;
- program arguments `knucleotide-input.fasta`;
- verifier `verify.rb`, SHA-256
  `5992ceafc2e064cc7f7e486adc8a0edf7066567383dd8b7c049088ec2c81d086`;
- combined contract SHA-256
  `1acf167c954333b6c8990a815597a71c956eece9efacbfc3b681aa75c0b5f438`.

## Five-process result

The Able bytecode samples were `38.56`, `40.05`, `39.51`, `39.24`, and
`38.48` seconds:

| Measure | Result |
| --- | ---: |
| Successful / attempted | 5 / 5 |
| Timeouts / failures | 0 / 0 |
| Verifier accepted | 5 / 5 |
| Mean | 39.1680 s |
| Median | 39.2400 s |
| Minimum / maximum | 38.4800 / 40.0500 s |
| Sample standard deviation | 0.6601 s |
| Coefficient of variation | 1.6854% |
| Able / Python | 33.52x |
| Able / Ruby | 32.51x |

The exact-run validator accepted five retained successful Able samples and
five retained successful samples for each reference interpreter. Python's
mean was 1.1685 seconds and Ruby's was 1.2047 seconds.

## Scope

No VM, compiler, stdlib, application, foreign reference, manifest, or current
promoted scorecard changed. The historical full-status scorecard still reports
the measurement it actually contains; it was not relabeled by splicing this
focused row into older measurements. The focused result and variance artifact
are the evidence used to admit the row to the next fresh cohorts.

Artifacts:

- `2026-07-15-k-nucleotide-bytecode-selection-eligibility-contract.json`
- `2026-07-15-k-nucleotide-bytecode-selection-eligibility.json`
- `2026-07-15-k-nucleotide-bytecode-selection-eligibility-comparison.md`
- `2026-07-15-k-nucleotide-bytecode-selection-eligibility-variance.json`
- `2026-07-15-k-nucleotide-bytecode-selection-eligibility-variance.md`

## Next recommendation

Collect two independent full five-run tagged scorecard cohorts with
`--no-promote`, using this exact reviewed selection manifest and canonical
stdlib source record, then run strict selected variance. Why: every selected
row now has direct evidence that it can complete the required five-process
gate, so cohort collection can measure real cross-application repeatability
instead of predictably failing on eligibility.

This entails keeping all 64 application/mode rows in timeout-preserving full
status, requiring five verifier-backed Able/reference samples for each of the
58 selected rows, checking identical manifest and stdlib identities plus
disjoint source reports, and promoting only after review. Profile or change
runtime/compiler code only when the completed cohorts show the same concrete
cost in at least three unlike applications; K-Nucleotide's map/text parents
alone remain insufficient evidence for an optimization.
