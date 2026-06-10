#!/usr/bin/env python3
"""Fast source-contract tests for isolated process timing metrics."""

from __future__ import annotations

import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
BENCH_PERF = SCRIPT_DIR / "bench_perf"


class TimingStreamContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.source = BENCH_PERF.read_text(encoding="utf-8")

    def test_time_metrics_use_a_dedicated_output_file(self) -> None:
        self.assertEqual(
            self.source.count('/usr/bin/time -p -o "$time_file"'),
            4,
        )
        self.assertGreaterEqual(
            self.source.count(
                """awk '$1 == "real" { print $2; exit }' "$time_file" """
                .rstrip()
            ),
            2,
        )

    def test_gc_trace_count_remains_on_program_stderr(self) -> None:
        self.assertGreaterEqual(
            self.source.count(
                """grep -cE '(^|[^[:alpha:]])gc [0-9]+ @' "$stderr_file" """
                .rstrip()
            ),
            2,
        )


if __name__ == "__main__":
    unittest.main()
