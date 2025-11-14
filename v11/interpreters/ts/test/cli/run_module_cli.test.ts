import { describe, expect, test } from "bun:test";
import { mkdtempSync, rmSync, writeFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import os from "node:os";
import { spawnSync } from "node:child_process";

const SCRIPT_PATH = path.resolve(__dirname, "../../scripts/run-module.ts");
const BUN_BIN = process.execPath;

describe("Able CLI (Bun prototype)", () => {
  test("run command executes Able module and prints stdout", () => {
    const result = runCli("run", {
      files: {
        "main.able": `
package cli_demo

fn main() -> void {
  print("Hello from CLI")
}
`,
      },
    });
    expect(result.status).toBe(0);
    expect(result.stderr).toBe("");
    expect(result.stdout).toContain("Hello from CLI");
  });

  test("check command reports typechecker diagnostics with package summary", () => {
    const result = runCli("check", {
      manifestName: "cli_test",
      files: {
        "main.able": `
package cli_fail

import cli_test.dep.{missing_symbol}

fn main() -> void {}
`,
        "dep.able": `
package dep

fn helper() -> void {}
`,
      },
    });
    expect(result.status).toBe(1);
    expect(result.stderr).toContain("typechecker:");
    expect(result.stderr).toContain("cli_fail");
    expect(result.stderr).toContain("package export summary");
  });

  test("run command merges multiple files belonging to the same package", () => {
    const result = runCli("run", {
      manifestName: "cli_multi",
      files: {
        "main.able": `
package cli_multi

fn main() -> void {
  print(helper_value())
}
`,
        "support.able": `
package cli_multi

fn helper_value() -> string {
  "helper-output"
}
`,
      },
    });
    expect(result.status).toBe(0);
    expect(result.stdout).toContain("helper-output");
    expect(result.stderr).toBe("");
  });

  test("run command loads packages discovered via search path env override", () => {
    const depDir = mkdtempSync(path.join(os.tmpdir(), "able-cli-dep-"));
    try {
      writeFixtureFile(depDir, "package.yml", "name: extras\n");
      writeFixtureFile(
        depDir,
        "helpers/greetings.able",
        `
fn greeting() -> string {
  "Hello from dependency"
}
`,
      );

      const result = runCli("run", {
        manifestName: "cli_root",
        files: {
          "main.able": `
package cli_root

import extras.helpers.{greeting}

fn main() -> void {
  print(greeting())
}
`,
        },
        env: {
          ABLE_MODULE_PATHS: depDir,
        },
      });

      expect(result.status).toBe(0);
      expect(result.stdout).toContain("Hello from dependency");
      expect(result.stderr).toBe("");
    } finally {
      rmSync(depDir, { recursive: true, force: true });
    }
  });

  test("run command loads packages discovered via ABLE_PATH alias", () => {
    const depDir = mkdtempSync(path.join(os.tmpdir(), "able-cli-alias-"));
    try {
      writeFixtureFile(depDir, "package.yml", "name: alias_pkg\n");
      writeFixtureFile(
        depDir,
        "helpers/greetings.able",
        `
fn greeting() -> string {
  "Hello from alias dependency"
}
`,
      );

      const result = runCli("run", {
        manifestName: "cli_root_alias",
        files: {
          "main.able": `
package cli_root_alias

import alias_pkg.helpers.{greeting}

fn main() -> void {
  print(greeting())
}
`,
        },
        env: {
          ABLE_PATH: depDir,
        },
      });

      expect(result.status).toBe(0);
      expect(result.stdout).toContain("Hello from alias dependency");
      expect(result.stderr).toBe("");
    } finally {
      rmSync(depDir, { recursive: true, force: true });
    }
  });

  test("run command respects ABLE_STD_LIB override order", () => {
    const stdlibRoot = mkdtempSync(path.join(os.tmpdir(), "able-cli-stdlib-"));
    try {
      writeFixtureFile(stdlibRoot, "package.yml", "name: able\n");
      writeFixtureFile(
        stdlibRoot,
        "src/custom.able",
        `
package custom

fn greeting() -> string { "Hello from custom stdlib" }
`,
      );

      const result = runCli("run", {
        manifestName: "cli_root_stdlib",
        files: {
          "main.able": `
package cli_root_stdlib

import able.custom.{greeting}

fn main() -> void {
  print(greeting())
}
`,
        },
        env: {
          ABLE_STD_LIB: path.join(stdlibRoot, "src"),
        },
      });

      expect(result.status).toBe(0);
      expect(result.stdout).toContain("Hello from custom stdlib");
      expect(result.stderr).toBe("");
    } finally {
      rmSync(stdlibRoot, { recursive: true, force: true });
    }
  });

  test("run command skips missing search paths with a warning", () => {
    const missingPath = path.join(os.tmpdir(), `able-missing-${Date.now()}`);
    const result = runCli("run", {
      files: {
        "main.able": `
package cli_warning

fn main() -> void {
  print("still works")
}
`,
      },
      env: {
        ABLE_MODULE_PATHS: missingPath,
      },
    });
    expect(result.status).toBe(0);
    expect(result.stdout).toContain("still works");
    expect(result.stderr).toContain("skipping search path");
  });

  test("run command reports diagnostics for missing import selectors", () => {
    const result = runCli("run", {
      manifestName: "cli_diag",
      files: {
        "main.able": `
package cli_diag.main

import cli_diag.support.{missing_symbol}

fn main() -> void {
  print(missing_symbol())
}
`,
        "support.able": `
package support

fn available() -> void {}
`,
      },
    });
    expect(result.status).toBe(1);
    expect(result.stderr).toContain("typechecker: package 'cli_diag.support' has no symbol 'missing_symbol'");
  });

  test("run command aborts under strict typecheck mode", () => {
    const result = runCli("run", {
      files: typecheckFailureProgram(),
    });
    expect(result.status).toBe(1);
    expect(result.stdout).toBe("");
    expect(result.stderr).toContain("typechecker:");
    expect(result.stderr).toContain("cli_warn");
  });

  test("run command continues when ABLE_TYPECHECK_FIXTURES=warn", () => {
    const result = runCli("run", {
      files: typecheckFailureProgram(),
      env: {
        ABLE_TYPECHECK_FIXTURES: "warn",
      },
    });
    expect(result.status).toBe(0);
    expect(result.stdout).toContain("hello from warn mode");
    expect(result.stderr).toContain("typechecker:");
    expect(result.stderr).toContain("cli_warn");
    expect(result.stderr).toContain("ABLE_TYPECHECK_FIXTURES=warn");
  });

  test("test command summarizes plan in list mode", () => {
    const result = runTestCli(["--list"]);
    expect(result.status).toBe(2);
    expect(result.stdout).toContain("able test plan (skeleton)");
    expect(result.stdout).toContain("list_only=true");
    expect(result.stderr).toContain("able test CLI skeleton");
  });

  test("test command parses filters, run options, and targets", () => {
    const result = runTestCli(
      [
        "--path",
        "pkg",
        "--exclude-path",
        "tmp",
        "--name",
        "spec",
        "--exclude-name",
        "skip",
        "--tag",
        "fast",
        "--exclude-tag",
        "flaky",
        "--format",
        "progress",
        "--fail-fast",
        "--repeat",
        "3",
        "--parallel",
        "2",
        "--shuffle",
        "123",
        "--dry-run",
        "--update-snapshots",
        ".",
      ],
      {
        workspace: {
          manifestName: "cli_tests",
          files: {
            "tests/example.test.able": `
package cli_tests.tests

fn main() -> void {}
`,
          },
        },
      },
    );
    expect(result.status).toBe(2);
    expect(result.stdout).toContain("targets=.");
    expect(result.stdout).toContain("include_paths=pkg");
    expect(result.stdout).toContain("exclude_tags=flaky");
    expect(result.stdout).toContain("format=progress");
    expect(result.stdout).toContain("repeat=3");
    expect(result.stdout).toContain("shuffle_seed=123");
    expect(result.stdout).toContain("list_only=true");
    expect(result.stdout).toContain("update_snapshots=true");
  });
});

type CliResult = { status: number | null; stdout: string; stderr: string };

type RunCliOptions = {
  files: Record<string, string>;
  manifestName?: string;
  entry?: string;
  env?: Record<string, string>;
};

type CLICommand = "run" | "check";
type RunTestCliOptions = {
  workspace?: {
    files: Record<string, string>;
    manifestName?: string;
  };
  env?: Record<string, string>;
};

function runCli(command: CLICommand, options: RunCliOptions): CliResult {
  const dir = mkdtempSync(path.join(os.tmpdir(), "able-cli-"));
  try {
    if (options.manifestName) {
      writeFixtureFile(dir, "package.yml", `name: ${options.manifestName}\n`);
    }
    for (const [relative, contents] of Object.entries(options.files)) {
      writeFixtureFile(dir, relative, contents);
    }
    const entryRelative = options.entry ?? "main.able";
    const entryPath = path.join(dir, entryRelative);
    const env = {
      ...process.env,
      ...(options.env ?? {}),
    };
    const result = spawnSync(BUN_BIN, ["run", SCRIPT_PATH, command, entryPath], {
      encoding: "utf8",
      env,
    });
    return {
      status: result.status,
      stdout: result.stdout ?? "",
      stderr: result.stderr ?? "",
    };
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

function writeFixtureFile(root: string, relative: string, contents: string): void {
  const destination = path.join(root, relative);
  mkdirSync(path.dirname(destination), { recursive: true });
  writeFileSync(destination, contents, "utf8");
}

function runTestCli(args: string[], options: RunTestCliOptions = {}): CliResult {
  const env = {
    ...process.env,
    ...(options.env ?? {}),
  };
  let cwd = process.cwd();
  let workspaceDir: string | undefined;

  if (options.workspace) {
    workspaceDir = mkdtempSync(path.join(os.tmpdir(), "able-cli-test-"));
    if (options.workspace.manifestName) {
      writeFixtureFile(workspaceDir, "package.yml", `name: ${options.workspace.manifestName}\n`);
    }
    for (const [relative, contents] of Object.entries(options.workspace.files)) {
      writeFixtureFile(workspaceDir, relative, contents);
    }
    cwd = workspaceDir;
  }

  const result = spawnSync(BUN_BIN, ["run", SCRIPT_PATH, "test", ...args], {
    encoding: "utf8",
    env,
    cwd,
  });

  if (workspaceDir) {
    rmSync(workspaceDir, { recursive: true, force: true });
  }

  return {
    status: result.status,
    stdout: result.stdout ?? "",
    stderr: result.stderr ?? "",
  };
}

function typecheckFailureProgram(): Record<string, string> {
  return {
    "main.able": `
package cli_warn

interface Show {
  fn to_string(self: Self) -> string
}

struct Point {
  value: i32
}

impl Show for Point {
  fn display(self: Self) -> string {
    "point"
  }
}

fn main() -> void {
  print("hello from warn mode")
}
`,
  };
}
