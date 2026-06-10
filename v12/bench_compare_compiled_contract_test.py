#!/usr/bin/env python3
"""Fast source-contract tests for strict compiled comparison artifacts."""

from __future__ import annotations

import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
COMPARE = SCRIPT_DIR / "bench_compare_external"


class CompiledComparisonContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.source = COMPARE.read_text(encoding="utf-8")

    def test_compiled_build_arguments_are_forwarded_only_to_compiled_mode(self) -> None:
        self.assertIn('if [[ "$mode" == "compiled" ]]; then', self.source)
        self.assertIn(
            'bench_compiled_build_args+=("--compiled-build-arg=$compiled_build_arg")',
            self.source,
        )
        self.assertIn('"${bench_compiled_build_args[@]}"', self.source)

    def test_kept_comparison_retains_each_benchmark_artifact_directory(self) -> None:
        self.assertIn(
            'bench_artifact_args=(--workdir "$workdir/artifacts/$bench-$mode" --keep)',
            self.source,
        )
        self.assertIn('"${bench_artifact_args[@]}"', self.source)


if __name__ == "__main__":
    unittest.main()
