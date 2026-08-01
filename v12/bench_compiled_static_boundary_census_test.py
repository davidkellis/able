#!/usr/bin/env python3
import hashlib
import json
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
V12 = ROOT / "v12"
REPORT = (
    V12
    / "docs/perf-baselines/2026-07-31-current-default-primitive-boxing-boundary-census-no-go.json"
)
FRONTIER = (
    V12
    / "docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json"
)


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class CompiledStaticBoundaryCensusTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.report = json.loads(REPORT.read_text())
        cls.manifest = json.loads((V12 / "bench-selection-manifest.json").read_text())

    def test_report_covers_current_compiled_selection(self) -> None:
        selected = self.manifest["modes"]["compiled"]
        self.assertEqual(len(selected), 66)
        census = self.report["coverage"]
        self.assertEqual(census["selected_rows"], 66)
        self.assertEqual(census["successful_rows"], 66)
        self.assertEqual(census["failed_rows"], 0)
        self.assertEqual(census["dependency_failed_rows"], 0)
        self.assertEqual(census["interpreter_linked_rows"], 0)

    def test_inputs_are_content_addressed(self) -> None:
        self.assertEqual(
            self.report["inputs"]["frontier_sha256"],
            sha256(FRONTIER),
        )

    def test_current_census_admits_no_named_or_cold_candidate(self) -> None:
        decision = self.report["decision"]
        self.assertEqual(decision["status"], "retain-no-code")
        self.assertIsNone(decision["admitted_category"])
        self.assertFalse(decision["runtime_profiles_run"])
        self.assertFalse(decision["ab_prototype_run"])

    def test_broad_identities_are_only_host_or_hashmap_boundaries(self) -> None:
        identities = self.report["cross_group_identity_review"]
        self.assertEqual(len(identities), 5)
        self.assertEqual(
            {item["class"] for item in identities},
            {"explicit-main-host-abi", "named-hashmap-kernel"},
        )
        self.assertEqual(
            self.report["primitive_encode_coverage"][
                "identities_in_at_least_three_workload_groups"
            ],
            5,
        )


if __name__ == "__main__":
    unittest.main()
