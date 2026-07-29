#!/usr/bin/env python3
"""Fast contract tests for the weighted three-family interaction frontier."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
GENERATOR = SCRIPT_DIR / "bench_feature_interaction_triples"
BASELINE_REMOVALS = [
    "lexical_blocks_bindings_patterns=backup_dedup",
    "types_nominals_generics_unions=backup_dedup",
    "expressions_arrays_text_files=backup_dedup",
    "functions_closures_callables=backup_dedup",
    "inherent_methods=backup_dedup",
    "option_result_exceptions=backup_dedup",
    "stdlib_protocols_regex=backup_dedup",
    "program_entry=backup_dedup",
    "lexical_blocks_bindings_patterns=concurrent_text_index",
    "control_flow=concurrent_text_index",
    "option_result_exceptions=concurrent_text_index",
    "lexical_blocks_bindings_patterns=concurrent_document_pipeline",
    "types_nominals_generics_unions=concurrent_document_pipeline",
    "expressions_arrays_text_files=concurrent_document_pipeline",
    "functions_closures_callables=concurrent_document_pipeline",
    "control_flow=concurrent_document_pipeline",
    "inherent_methods=concurrent_document_pipeline",
    "option_result_exceptions=concurrent_document_pipeline",
    "stdlib_protocols_regex=concurrent_document_pipeline",
    "concurrency=concurrent_document_pipeline",
    "program_entry=concurrent_document_pipeline",
    "lexical_blocks_bindings_patterns=manifest_normalization",
    "types_nominals_generics_unions=manifest_normalization",
    "expressions_arrays_text_files=manifest_normalization",
    "functions_closures_callables=manifest_normalization",
    "control_flow=manifest_normalization",
    "interfaces_implementations_dispatch=manifest_normalization",
    "option_result_exceptions=manifest_normalization",
    "stdlib_protocols_regex=manifest_normalization",
    "program_entry=manifest_normalization",
    "expressions_arrays_text_files=validated_job_pipeline",
    "program_entry=validated_job_pipeline",
    "lexical_blocks_bindings_patterns=binary_event_log",
    "types_nominals_generics_unions=binary_event_log",
    "expressions_arrays_text_files=binary_event_log",
    "functions_closures_callables=binary_event_log",
    "control_flow=binary_event_log",
    "interfaces_implementations_dispatch=binary_event_log",
    "option_result_exceptions=binary_event_log",
    "stdlib_protocols_regex=binary_event_log",
    "program_entry=binary_event_log",
    "lexical_blocks_bindings_patterns=sensor_calibration",
    "types_nominals_generics_unions=sensor_calibration",
    "expressions_arrays_text_files=sensor_calibration",
    "control_flow=sensor_calibration",
    "inherent_methods=sensor_calibration",
    "interfaces_implementations_dispatch=sensor_calibration",
    "option_result_exceptions=sensor_calibration",
    "stdlib_protocols_regex=sensor_calibration",
    "program_entry=sensor_calibration",
    "lexical_blocks_bindings_patterns=concurrent_stencil_reduction",
    "types_nominals_generics_unions=concurrent_stencil_reduction",
    "expressions_arrays_text_files=concurrent_stencil_reduction",
    "control_flow=concurrent_stencil_reduction",
    "inherent_methods=concurrent_stencil_reduction",
    "option_result_exceptions=concurrent_stencil_reduction",
    "concurrency=concurrent_stencil_reduction",
    "stdlib_protocols_regex=concurrent_stencil_reduction",
    "program_entry=concurrent_stencil_reduction",
    "lexical_blocks_bindings_patterns=concurrent_signal_dispatch",
    "types_nominals_generics_unions=concurrent_signal_dispatch",
    "expressions_arrays_text_files=concurrent_signal_dispatch",
    "control_flow=concurrent_signal_dispatch",
    "inherent_methods=concurrent_signal_dispatch",
    "interfaces_implementations_dispatch=concurrent_signal_dispatch",
    "option_result_exceptions=concurrent_signal_dispatch",
    "concurrency=concurrent_signal_dispatch",
    "stdlib_protocols_regex=concurrent_signal_dispatch",
    "program_entry=concurrent_signal_dispatch",
    "lexical_blocks_bindings_patterns=concurrent_transform_chain",
    "types_nominals_generics_unions=concurrent_transform_chain",
    "expressions_arrays_text_files=concurrent_transform_chain",
    "functions_closures_callables=concurrent_transform_chain",
    "control_flow=concurrent_transform_chain",
    "inherent_methods=concurrent_transform_chain",
    "option_result_exceptions=concurrent_transform_chain",
    "concurrency=concurrent_transform_chain",
    "stdlib_protocols_regex=concurrent_transform_chain",
    "program_entry=concurrent_transform_chain",
    "lexical_blocks_bindings_patterns=concurrent_policy_callbacks",
    "types_nominals_generics_unions=concurrent_policy_callbacks",
    "expressions_arrays_text_files=concurrent_policy_callbacks",
    "functions_closures_callables=concurrent_policy_callbacks",
    "control_flow=concurrent_policy_callbacks",
    "inherent_methods=concurrent_policy_callbacks",
    "interfaces_implementations_dispatch=concurrent_policy_callbacks",
    "option_result_exceptions=concurrent_policy_callbacks",
    "concurrency=concurrent_policy_callbacks",
    "stdlib_protocols_regex=concurrent_policy_callbacks",
    "program_entry=concurrent_policy_callbacks",
    "lexical_blocks_bindings_patterns=concurrent_stateful_pipeline",
    "types_nominals_generics_unions=concurrent_stateful_pipeline",
    "expressions_arrays_text_files=concurrent_stateful_pipeline",
    "functions_closures_callables=concurrent_stateful_pipeline",
    "control_flow=concurrent_stateful_pipeline",
    "interfaces_implementations_dispatch=concurrent_stateful_pipeline",
    "option_result_exceptions=concurrent_stateful_pipeline",
    "concurrency=concurrent_stateful_pipeline",
    "stdlib_protocols_regex=concurrent_stateful_pipeline",
    "program_entry=concurrent_stateful_pipeline",
    "lexical_blocks_bindings_patterns=concurrent_graph_visitors",
    "types_nominals_generics_unions=concurrent_graph_visitors",
    "expressions_arrays_text_files=concurrent_graph_visitors",
    "functions_closures_callables=concurrent_graph_visitors",
    "control_flow=concurrent_graph_visitors",
    "inherent_methods=concurrent_graph_visitors",
    "interfaces_implementations_dispatch=concurrent_graph_visitors",
    "option_result_exceptions=concurrent_graph_visitors",
    "concurrency=concurrent_graph_visitors",
    "stdlib_protocols_regex=concurrent_graph_visitors",
    "program_entry=concurrent_graph_visitors",
    "lexical_blocks_bindings_patterns=concurrent_tree_folds",
    "types_nominals_generics_unions=concurrent_tree_folds",
    "expressions_arrays_text_files=concurrent_tree_folds",
    "functions_closures_callables=concurrent_tree_folds",
    "control_flow=concurrent_tree_folds",
    "inherent_methods=concurrent_tree_folds",
    "interfaces_implementations_dispatch=concurrent_tree_folds",
    "option_result_exceptions=concurrent_tree_folds",
    "concurrency=concurrent_tree_folds",
    "stdlib_protocols_regex=concurrent_tree_folds",
    "program_entry=concurrent_tree_folds",
    "lexical_blocks_bindings_patterns=concurrent_audio_voices",
    "types_nominals_generics_unions=concurrent_audio_voices",
    "expressions_arrays_text_files=concurrent_audio_voices",
    "functions_closures_callables=concurrent_audio_voices",
    "control_flow=concurrent_audio_voices",
    "inherent_methods=concurrent_audio_voices",
    "interfaces_implementations_dispatch=concurrent_audio_voices",
    "option_result_exceptions=concurrent_audio_voices",
    "concurrency=concurrent_audio_voices",
    "stdlib_protocols_regex=concurrent_audio_voices",
    "program_entry=concurrent_audio_voices",
    "lexical_blocks_bindings_patterns=concurrent_packet_codecs",
    "types_nominals_generics_unions=concurrent_packet_codecs",
    "expressions_arrays_text_files=concurrent_packet_codecs",
    "functions_closures_callables=concurrent_packet_codecs",
    "control_flow=concurrent_packet_codecs",
    "inherent_methods=concurrent_packet_codecs",
    "interfaces_implementations_dispatch=concurrent_packet_codecs",
    "option_result_exceptions=concurrent_packet_codecs",
    "concurrency=concurrent_packet_codecs",
    "stdlib_protocols_regex=concurrent_packet_codecs",
    "program_entry=concurrent_packet_codecs",
    "lexical_blocks_bindings_patterns=concurrent_scene_tiles",
    "types_nominals_generics_unions=concurrent_scene_tiles",
    "expressions_arrays_text_files=concurrent_scene_tiles",
    "functions_closures_callables=concurrent_scene_tiles",
    "control_flow=concurrent_scene_tiles",
    "inherent_methods=concurrent_scene_tiles",
    "interfaces_implementations_dispatch=concurrent_scene_tiles",
    "option_result_exceptions=concurrent_scene_tiles",
    "concurrency=concurrent_scene_tiles",
    "stdlib_protocols_regex=concurrent_scene_tiles",
    "program_entry=concurrent_scene_tiles",
    "lexical_blocks_bindings_patterns=concurrent_state_machines",
    "types_nominals_generics_unions=concurrent_state_machines",
    "expressions_arrays_text_files=concurrent_state_machines",
    "functions_closures_callables=concurrent_state_machines",
    "control_flow=concurrent_state_machines",
    "inherent_methods=concurrent_state_machines",
    "interfaces_implementations_dispatch=concurrent_state_machines",
    "option_result_exceptions=concurrent_state_machines",
    "concurrency=concurrent_state_machines",
    "stdlib_protocols_regex=concurrent_state_machines",
    "program_entry=concurrent_state_machines",
]


class FeatureInteractionTripleTests(unittest.TestCase):
    def generate(self, *extra: str) -> dict[str, object]:
        with tempfile.TemporaryDirectory() as raw_dir:
            output = Path(raw_dir) / "triples.json"
            command = [str(GENERATOR)]
            for removal in BASELINE_REMOVALS:
                command.extend(["--baseline-remove", removal])
            command.extend(extra)
            command.extend(["--json-out", str(output)])
            result = subprocess.run(
                command,
                cwd=REPO_ROOT,
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            return json.loads(output.read_text(encoding="utf-8"))

    def test_current_frontier_raises_minimum_depth_to_eleven(self) -> None:
        report = self.generate()
        summary = report["summary"]
        self.assertEqual(summary["family_count"], 11)
        self.assertEqual(summary["triple_count"], 165)
        self.assertEqual(summary["baseline_zero_triples"], 0)
        self.assertEqual(summary["current_zero_triples"], 0)
        self.assertEqual(summary["baseline_min_depth"], 1)
        self.assertEqual(summary["current_min_depth"], 11)
        self.assertEqual(summary["baseline_depth_one_triples"], 8)
        self.assertEqual(summary["current_depth_one_triples"], 0)
        self.assertEqual(summary["improved_triples"], 165)
        self.assertAlmostEqual(
            summary["performance_frontier_total_excess_seconds"],
            226.856947,
            places=6,
        )

        by_families = {
            tuple(item["families"]): item for item in report["triples"]
        }
        for families in [
            (
                "concurrency",
                "expressions_arrays_text_files",
                "functions_closures_callables",
            ),
            ("concurrency", "functions_closures_callables", "program_entry"),
        ]:
            item = by_families[families]
            self.assertEqual(item["baseline_count"], 1)
            self.assertEqual(item["current_count"], 12)
            self.assertEqual(item["delta"], 11)
            self.assertEqual(
                item["portable_benchmarks"],
                [
                    "concurrent_audio_voices",
                    "concurrent_document_pipeline",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_transform_chain",
                    "concurrent_tree_folds",
                    "validated_job_pipeline",
                ],
            )

        for families in [
            (
                "concurrency",
                "expressions_arrays_text_files",
                "interfaces_implementations_dispatch",
            ),
            (
                "concurrency",
                "interfaces_implementations_dispatch",
                "program_entry",
            ),
        ]:
            item = by_families[families]
            self.assertEqual(item["baseline_count"], 2)
            self.assertEqual(item["current_count"], 12)
            self.assertEqual(item["delta"], 10)
            self.assertEqual(
                item["portable_benchmarks"],
                [
                    "concurrent_audio_voices",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_signal_dispatch",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_text_index",
                    "concurrent_tree_folds",
                    "validated_job_pipeline",
                ],
            )

        target = by_families[
            (
                "expressions_arrays_text_files",
                "functions_closures_callables",
                "option_result_exceptions",
            )
        ]
        self.assertEqual(target["baseline_count"], 2)
        self.assertEqual(target["current_count"], 16)
        self.assertEqual(target["delta"], 14)
        self.assertEqual(
            target["portable_benchmarks"],
            [
                "backup_dedup",
                "binary_event_log",
                "concurrent_audio_voices",
                "concurrent_document_pipeline",
                "concurrent_event_routing",
                "concurrent_graph_visitors",
                "concurrent_packet_codecs",
                "concurrent_policy_callbacks",
                "concurrent_scene_tiles",
                "concurrent_state_machines",
                "concurrent_stateful_pipeline",
                "concurrent_transform_chain",
                "concurrent_tree_folds",
                "manifest_normalization",
                "policy_record_dispatch",
                "validated_job_pipeline",
            ],
        )

        target_interaction = by_families[
            (
                "concurrency",
                "functions_closures_callables",
                "interfaces_implementations_dispatch",
            )
        ]
        self.assertEqual(target_interaction["baseline_count"], 3)
        self.assertEqual(target_interaction["current_count"], 11)
        self.assertEqual(target_interaction["delta"], 8)
        self.assertEqual(
            target_interaction["portable_benchmarks"],
            [
                "concurrent_audio_voices",
                "concurrent_event_routing",
                "concurrent_graph_visitors",
                "concurrent_packet_codecs",
                "concurrent_policy_callbacks",
                "concurrent_scene_tiles",
                "concurrent_state_machines",
                "concurrent_stateful_pipeline",
                "concurrent_tree_folds",
                "dependency_wave_validation",
                "validated_job_pipeline",
            ],
        )

        method_interaction = by_families[
            (
                "functions_closures_callables",
                "inherent_methods",
                "interfaces_implementations_dispatch",
            )
        ]
        self.assertEqual(method_interaction["baseline_count"], 4)
        self.assertEqual(method_interaction["current_count"], 11)
        self.assertEqual(method_interaction["delta"], 7)
        self.assertEqual(
            method_interaction["portable_benchmarks"],
            [
                "concurrent_audio_voices",
                "concurrent_event_routing",
                "concurrent_graph_visitors",
                "concurrent_packet_codecs",
                "concurrent_policy_callbacks",
                "concurrent_scene_tiles",
                "concurrent_state_machines",
                "concurrent_tree_folds",
                "dependency_wave_validation",
                "policy_record_dispatch",
                "validated_job_pipeline",
            ],
        )

    def test_concurrent_interaction_memberships_are_exact(self) -> None:
        report = self.generate()
        by_families = {
            tuple(item["families"]): item for item in report["triples"]
        }
        cases = {
            (
                "concurrency",
                "control_flow",
                "expressions_arrays_text_files",
            ): (1, 14, 13, [
                "concurrent_audio_voices",
                "concurrent_document_pipeline",
                "concurrent_event_routing",
                "concurrent_graph_visitors",
                "concurrent_packet_codecs",
                "concurrent_policy_callbacks",
                "concurrent_scene_tiles",
                "concurrent_signal_dispatch",
                "concurrent_state_machines",
                "concurrent_stateful_pipeline",
                "concurrent_stencil_reduction",
                "concurrent_text_index",
                "concurrent_transform_chain",
                "concurrent_tree_folds",
            ]),
            (
                "concurrency",
                "expressions_arrays_text_files",
                "lexical_blocks_bindings_patterns",
            ): (1, 14, 13, [
                "concurrent_audio_voices",
                "concurrent_document_pipeline",
                "concurrent_event_routing",
                "concurrent_graph_visitors",
                "concurrent_packet_codecs",
                "concurrent_policy_callbacks",
                "concurrent_scene_tiles",
                "concurrent_signal_dispatch",
                "concurrent_state_machines",
                "concurrent_stateful_pipeline",
                "concurrent_stencil_reduction",
                "concurrent_text_index",
                "concurrent_transform_chain",
                "concurrent_tree_folds",
            ]),
            (
                "concurrency",
                "expressions_arrays_text_files",
                "option_result_exceptions",
            ): (1, 15, 14, [
                "concurrent_audio_voices",
                "concurrent_document_pipeline",
                "concurrent_event_routing",
                "concurrent_graph_visitors",
                "concurrent_packet_codecs",
                "concurrent_policy_callbacks",
                "concurrent_scene_tiles",
                "concurrent_signal_dispatch",
                "concurrent_state_machines",
                "concurrent_stateful_pipeline",
                "concurrent_stencil_reduction",
                "concurrent_text_index",
                "concurrent_transform_chain",
                "concurrent_tree_folds",
                "validated_job_pipeline",
            ]),
        }
        for families, expected in cases.items():
            with self.subTest(families=families):
                item = by_families[families]
                self.assertEqual(item["baseline_count"], expected[0])
                self.assertEqual(item["current_count"], expected[1])
                self.assertEqual(item["delta"], expected[2])
                self.assertEqual(item["portable_benchmarks"], expected[3])

    def test_unknown_membership_is_rejected(self) -> None:
        result = subprocess.run(
            [
                str(GENERATOR),
                "--baseline-remove",
                "missing_family=concurrent_text_index",
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("unknown portable family", result.stderr)


if __name__ == "__main__":
    unittest.main()
