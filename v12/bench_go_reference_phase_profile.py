#!/usr/bin/env python3
"""Wrap a standalone Go reference so only its original main is profiled."""

from __future__ import annotations

import argparse
from pathlib import Path


WRAPPER = r'''package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
)

type referencePhaseStats struct {
	AllocatedBytes uint64 `json:"allocated_bytes"`
	Allocations    uint64 `json:"allocations"`
	Frees          uint64 `json:"frees"`
	GCCount        uint32 `json:"gc_count"`
}

func main() {
	var start runtime.MemStats
	statsPath := os.Getenv("ABLE_GO_REFERENCE_STATS")
	if statsPath != "" {
		runtime.ReadMemStats(&start)
	}

	var cpuFile *os.File
	cpuPath := os.Getenv("ABLE_GO_REFERENCE_CPU_PROFILE")
	if cpuPath != "" {
		if err := os.MkdirAll(filepath.Dir(cpuPath), 0o755); err != nil {
			panic(err)
		}
		var err error
		cpuFile, err = os.Create(cpuPath)
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			panic(err)
		}
	}

	calls := 1
	if raw := os.Getenv("ABLE_GO_REFERENCE_PROFILE_CALLS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			panic("ABLE_GO_REFERENCE_PROFILE_CALLS must be positive")
		}
		calls = parsed
	}
	if calls > 1 {
		originalStdout := os.Stdout
		discard, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			panic(err)
		}
		os.Stdout = discard
		for index := 1; index < calls; index++ {
			benchmarkMain()
		}
		os.Stdout = originalStdout
		if err := discard.Close(); err != nil {
			panic(err)
		}
	}
	benchmarkMain()

	if cpuFile != nil {
		pprof.StopCPUProfile()
		if err := cpuFile.Close(); err != nil {
			panic(err)
		}
	}
	if statsPath == "" {
		return
	}
	var end runtime.MemStats
	runtime.ReadMemStats(&end)
	payload, err := json.MarshalIndent(referencePhaseStats{
		AllocatedBytes: end.TotalAlloc - start.TotalAlloc,
		Allocations:    end.Mallocs - start.Mallocs,
		Frees:          end.Frees - start.Frees,
		GCCount:        end.NumGC - start.NumGC,
	}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Dir(statsPath), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(statsPath, append(payload, '\n'), 0o644); err != nil {
		panic(err)
	}
}
'''


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()

    source = args.source.resolve().read_text(encoding="utf-8")
    anchor = "func main() {"
    if source.count(anchor) != 1:
        raise ValueError(
            f"expected exactly one standalone Go main, found {source.count(anchor)}"
        )
    output_dir = args.output_dir.resolve()
    output_dir.mkdir(parents=True, exist_ok=True)
    (output_dir / "app.go").write_text(
        source.replace(anchor, "func benchmarkMain() {"),
        encoding="utf-8",
    )
    (output_dir / "profile_main.go").write_text(WRAPPER, encoding="utf-8")
    (output_dir / "go.mod").write_text(
        "module able-reference-profile\n\ngo 1.26\n",
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
