#!/usr/bin/env python3
"""Fast source-contract tests for bounded compiled preparation phases."""

from __future__ import annotations

import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
BENCH_PERF = SCRIPT_DIR / "bench_perf"


class CompiledPreparationPhaseTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.source = BENCH_PERF.read_text(encoding="utf-8")

    def test_emission_and_go_build_have_independent_bounds(self) -> None:
        self.assertIn('GO_CACHE="${GOCACHE:-$GO_DIR/.gocache}"', self.source)
        self.assertIn('echo ">>> Emitting compiled Go module"', self.source)
        self.assertIn(
            '"$ablec_bin" -main -pkg main -o "$compiled_out_dir"',
            self.source,
        )
        self.assertIn('echo ">>> Building compiled Go binary"', self.source)
        self.assertIn(
            'run_bounded_phase env GOCACHE="$GO_CACHE" \\\n'
            '      go build -C "$compiled_out_dir" -mod=mod -o "$compiled_bin" .',
            self.source,
        )

    def test_harness_does_not_hide_both_phases_inside_ablec_build(self) -> None:
        self.assertNotIn('"$ablec_bin" -build -o "$compiled_out_dir"', self.source)


if __name__ == "__main__":
    unittest.main()
