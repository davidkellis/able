#!/usr/bin/env python3
"""Contract tests for ordered architecture-evidence refreshes."""

from __future__ import annotations

import hashlib
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
CHAIN = ROOT / "bench_architecture_evidence_chain"

FAKE_GENERATOR = """#!/usr/bin/env python3
import argparse
import hashlib
import json
from pathlib import Path


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def resolve(root, raw):
    path = Path(raw)
    return path if path.is_absolute() else root / path


parser = argparse.ArgumentParser()
parser.add_argument("--evidence", type=Path, required=True)
parser.add_argument("--json-out", type=Path)
parser.add_argument("--markdown-out", type=Path)
parser.add_argument("--check", action="store_true")
args = parser.parse_args()
root = Path.cwd()
evidence = json.loads(args.evidence.read_text())
input_text = (root / "input.txt").read_text()
decision = evidence["decision"]
if input_text == "decision-change\\n" and args.evidence.name == "evidence-b.json":
    decision = "changed"
report = {
    "decision": decision,
    "sources": {
        "evidence": {
            "path": args.evidence.resolve().relative_to(root).as_posix(),
            "sha256": digest(args.evidence),
        }
    },
}
for key, record in evidence["sources"].items():
    raw_path = record["path"] if isinstance(record, dict) else record
    source = resolve(root, raw_path)
    report["sources"][key] = {"path": raw_path, "sha256": digest(source)}
encoded = json.dumps(report, indent=2, sort_keys=True) + "\\n"
markdown = evidence["markdown"]
if input_text == "markdown-change\\n" and args.evidence.name == "evidence-b.json":
    markdown = "# changed\\n"
if args.json_out:
    args.json_out.write_text(encoded)
if args.markdown_out:
    args.markdown_out.write_text(markdown)
if args.check:
    if resolve(root, evidence["_checked_json"]).read_text() != encoded:
        raise SystemExit("checked JSON stale")
    if resolve(root, evidence["_checked_markdown"]).read_text() != markdown:
        raise SystemExit("checked Markdown stale")
"""


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class ArchitectureEvidenceChainTests(unittest.TestCase):
    def setUp(self) -> None:
        temp_parent = Path(os.environ.get("TMPDIR", "/var/tmp"))
        self.temporary = tempfile.TemporaryDirectory(
            prefix="able-architecture-evidence-chain-test-", dir=temp_parent
        )
        self.repo = Path(self.temporary.name)
        self.generator = self.repo / "fake_generator.py"
        self.generator.write_text(FAKE_GENERATOR, encoding="utf-8")
        self.input = self.repo / "input.txt"
        self.input.write_text("stable-one\n", encoding="utf-8")

        self.evidence_a = self.repo / "evidence-a.json"
        self.checked_a = self.repo / "checked-a.json"
        self.markdown_a = self.repo / "checked-a.md"
        self.markdown_a.write_text("# stable A\n", encoding="utf-8")
        self.write_evidence(
            self.evidence_a,
            self.checked_a,
            self.markdown_a,
            {"input": {"path": "input.txt", "sha256": digest(self.input)}},
        )
        self.generate_checked(self.evidence_a, self.checked_a, self.markdown_a)

        self.evidence_b = self.repo / "evidence-b.json"
        self.checked_b = self.repo / "checked-b.json"
        self.markdown_b = self.repo / "checked-b.md"
        self.markdown_b.write_text("# stable B\n", encoding="utf-8")
        self.write_evidence(
            self.evidence_b,
            self.checked_b,
            self.markdown_b,
            {
                "upstream": {
                    "path": "checked-a.json",
                    "sha256": digest(self.checked_a),
                }
            },
        )
        self.generate_checked(self.evidence_b, self.checked_b, self.markdown_b)

        self.manifest = self.repo / "manifest.json"
        self.write_manifest()

    def tearDown(self) -> None:
        self.temporary.cleanup()

    def write_evidence(
        self,
        path: Path,
        checked_json: Path,
        checked_markdown: Path,
        sources: dict[str, object],
    ) -> None:
        path.write_text(
            json.dumps(
                {
                    "_checked_json": checked_json.name,
                    "_checked_markdown": checked_markdown.name,
                    "decision": "retain",
                    "markdown": checked_markdown.read_text(encoding="utf-8"),
                    "sources": sources,
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )

    def generate_checked(
        self, evidence: Path, checked_json: Path, checked_markdown: Path
    ) -> None:
        result = subprocess.run(
            [
                sys.executable,
                str(self.generator),
                "--evidence",
                str(evidence),
                "--json-out",
                str(checked_json),
                "--markdown-out",
                str(checked_markdown),
            ],
            cwd=self.repo,
            text=True,
            capture_output=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stderr)

    def write_manifest(self, *, b_dependencies: list[str] | None = None) -> None:
        manifest = {
            "kind": "able-architecture-evidence-chain",
            "version": 1,
            "nodes": [
                {
                    "id": "a",
                    "generator": self.generator.name,
                    "evidence": self.evidence_a.name,
                    "checked_json": self.checked_a.name,
                    "checked_markdown": self.markdown_a.name,
                    "dependencies": [],
                },
                {
                    "id": "b",
                    "generator": self.generator.name,
                    "evidence": self.evidence_b.name,
                    "checked_json": self.checked_b.name,
                    "checked_markdown": self.markdown_b.name,
                    "dependencies": ["a"] if b_dependencies is None else b_dependencies,
                },
            ],
        }
        self.manifest.write_text(
            json.dumps(manifest, indent=2) + "\n", encoding="utf-8"
        )

    def run_chain(self, mode: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [
                sys.executable,
                str(CHAIN),
                "--repo-root",
                str(self.repo),
                "--manifest",
                str(self.manifest),
                mode,
            ],
            cwd=self.repo,
            text=True,
            capture_output=True,
            check=False,
        )

    def tracked_bytes(self) -> dict[Path, bytes]:
        return {
            path: path.read_bytes()
            for path in (
                self.evidence_a,
                self.checked_a,
                self.evidence_b,
                self.checked_b,
            )
        }

    def test_current_chain_checks_without_writes(self) -> None:
        before = self.tracked_bytes()
        result = self.run_chain("--check")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.tracked_bytes(), before)
        self.assertIn("2 node(s)", result.stdout)

    def test_refresh_propagates_hashes_in_topological_order(self) -> None:
        old_a = digest(self.checked_a)
        old_b = digest(self.checked_b)
        self.input.write_text("stable-two\n", encoding="utf-8")
        result = self.run_chain("--refresh")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotEqual(digest(self.checked_a), old_a)
        self.assertNotEqual(digest(self.checked_b), old_b)
        evidence_b = json.loads(self.evidence_b.read_text(encoding="utf-8"))
        self.assertEqual(
            evidence_b["sources"]["upstream"]["sha256"], digest(self.checked_a)
        )
        self.assertEqual(
            json.loads(self.checked_a.read_text(encoding="utf-8"))["decision"],
            "retain",
        )
        self.assertEqual(self.run_chain("--check").returncode, 0)

    def test_decision_drift_fails_closed_and_rolls_back(self) -> None:
        before = self.tracked_bytes()
        self.input.write_text("decision-change\n", encoding="utf-8")
        result = self.run_chain("--refresh")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("decision-bearing JSON changed", result.stderr)
        self.assertIn("rolled back", result.stderr)
        self.assertEqual(self.tracked_bytes(), before)

    def test_markdown_drift_fails_closed_and_rolls_back(self) -> None:
        before = self.tracked_bytes()
        self.input.write_text("markdown-change\n", encoding="utf-8")
        result = self.run_chain("--refresh")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Markdown changed", result.stderr)
        self.assertIn("rolled back", result.stderr)
        self.assertEqual(self.tracked_bytes(), before)

    def test_manifest_dependency_mismatch_is_rejected(self) -> None:
        self.write_manifest(b_dependencies=[])
        result = self.run_chain("--check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("dependency mismatch", result.stderr)


if __name__ == "__main__":
    unittest.main()
