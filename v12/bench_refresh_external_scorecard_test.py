#!/usr/bin/env python3
"""Fast dry-run tests for manifest-driven scorecard sample partitioning."""

from __future__ import annotations

import json
import shlex
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
REFRESH = SCRIPT_DIR / "bench_refresh_external_scorecard"
DEFAULT_MANIFEST = SCRIPT_DIR / "bench-selection-manifest.json"


def option(command: list[str], name: str) -> str:
    index = command.index(name)
    return command[index + 1]


class RefreshPartitionTests(unittest.TestCase):
    def dry_run(self, manifest: Path, output_dir: Path) -> list[list[str]]:
        result = subprocess.run(
            [
                str(REFRESH),
                "--dry-run",
                "--no-promote",
                "--tag",
                "partition-test",
                "--output-dir",
                str(output_dir),
                "--selection-manifest",
                str(manifest),
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        return [shlex.split(line[2:]) for line in result.stdout.splitlines() if line.startswith("+ ")]

    def assert_partition(self, manifest_path: Path) -> None:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        coverage = set(manifest["modes"]["compiled"])
        selected_bytecode = set(manifest["modes"]["bytecode"])
        with tempfile.TemporaryDirectory() as raw_output:
            commands = self.dry_run(manifest_path, Path(raw_output) / "reports")

        compare_runs: dict[tuple[str, str], int] = {}
        reference_runs: dict[str, int] = {}
        evidence_checks = 0
        for command in commands:
            if str(SCRIPT_DIR / "bench_compare_external") in command:
                mode = option(command, "--modes")
                runs = int(option(command, "-r"))
                for benchmark in option(command, "-b").split(","):
                    key = (benchmark, mode)
                    self.assertNotIn(key, compare_runs)
                    compare_runs[key] = runs
            if str(SCRIPT_DIR / "bench_refresh_interpreter_refs") in command:
                runs = int(option(command, "-r"))
                for benchmark in option(command, "-b").split(","):
                    self.assertNotIn(benchmark, reference_runs)
                    reference_runs[benchmark] = runs
            if str(SCRIPT_DIR / "bench_scorecard_evidence_check.py") in command:
                evidence_checks += 1
                self.assertEqual(int(option(command, "--require-runs")), 5)

        self.assertEqual(
            set(compare_runs),
            {(benchmark, mode) for benchmark in coverage for mode in ("compiled", "bytecode")},
        )
        for benchmark in coverage:
            self.assertEqual(compare_runs[(benchmark, "compiled")], 5)
            expected = 5 if benchmark in selected_bytecode else 1
            self.assertEqual(compare_runs[(benchmark, "bytecode")], expected)
            self.assertEqual(reference_runs[benchmark], expected)
        self.assertEqual(evidence_checks, 1)

    def test_default_manifest_drives_five_selected_and_one_status_probe(self) -> None:
        self.assert_partition(DEFAULT_MANIFEST)

    def test_partition_changes_when_reviewed_selection_changes(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest = json.loads(DEFAULT_MANIFEST.read_text(encoding="utf-8"))
            reviewed = manifest["modes"]["bytecode"]
            if "fib" in reviewed:
                reviewed.remove("fib")
            else:
                reviewed.append("fib")
            path = root / "selection.json"
            path.write_text(json.dumps(manifest), encoding="utf-8")
            self.assert_partition(path)


if __name__ == "__main__":
    unittest.main()
