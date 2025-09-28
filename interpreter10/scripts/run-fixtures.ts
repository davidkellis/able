import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, V10 } from "../index";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../fixtures/ast");

type Manifest = {
  description?: string;
  entry?: string;
  expect?: {
    result?: { kind: string; value?: unknown };
    stdout?: string[];
    errors?: string[];
  };
};

type FixtureResult = { name: string; description?: string };

async function main() {
  const fixtures = await collectFixtures(FIXTURE_ROOT);
  if (fixtures.length === 0) {
    console.log("No fixtures found.");
    return;
  }

  const results: FixtureResult[] = [];

  for (const fixtureDir of fixtures) {
    const manifest = await readManifest(fixtureDir);
    const entry = manifest.entry ?? "module.json";
    const entryPath = path.join(fixtureDir, entry);
    const moduleAst = await readModule(entryPath);

    const interpreter = new V10.InterpreterV10();
    ensurePrint(interpreter);
    const stdout: string[] = [];
    let evaluationError: unknown;
    let result: V10.V10Value | undefined;
    interceptStdout(stdout, () => {
      try {
        result = interpreter.evaluate(moduleAst);
      } catch (err) {
        evaluationError = err;
      }
    });

    assertExpectations(fixtureDir, manifest.expect, result, stdout, evaluationError);
    results.push({ name: path.relative(FIXTURE_ROOT, fixtureDir), description: manifest.description });
  }

  for (const res of results) {
    const desc = res.description ? ` - ${res.description}` : "";
    console.log(`✓ ${res.name}${desc}`);
  }
  console.log(`Executed ${results.length} fixture(s).`);
}

async function collectFixtures(root: string): Promise<string[]> {
  const dirs: string[] = [];
  async function walk(current: string) {
    const entries = await fs.readdir(current, { withFileTypes: true });
    let hasModule = false;
    for (const entry of entries) {
      if (entry.isFile() && entry.name.endsWith(".json") && entry.name === "module.json") {
        hasModule = true;
      }
    }
    if (hasModule) {
      dirs.push(current);
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        await walk(path.join(current, entry.name));
      }
    }
  }
  await walk(root);
  return dirs.sort();
}

async function readManifest(dir: string): Promise<Manifest> {
  const manifestPath = path.join(dir, "manifest.json");
  try {
    const contents = await fs.readFile(manifestPath, "utf8");
    return JSON.parse(contents);
  } catch (err: any) {
    if (err.code === "ENOENT") return {};
    throw err;
  }
}

async function readModule(filePath: string): Promise<AST.Module> {
  const raw = JSON.parse(await fs.readFile(filePath, "utf8"));
  return hydrateNode(raw) as AST.Module;
}

function hydrateNode(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(hydrateNode);
  if (value && typeof value === "object") {
    const node = value as Record<string, unknown>;
    if (typeof node.type === "string") {
      switch (node.type) {
        case "IntegerLiteral":
          if (typeof node.value === "string") node.value = BigInt(node.value);
          break;
        case "FloatLiteral":
          if (typeof node.value === "string") node.value = Number(node.value);
          break;
        case "BooleanLiteral":
          if (typeof node.value === "string") node.value = node.value === "true";
          break;
        case "ArrayLiteral":
          node.elements = hydrateNode(node.elements) as unknown[];
          break;
        case "Module":
          node.imports = hydrateNode(node.imports) as unknown[];
          node.body = hydrateNode(node.body) as unknown[];
          break;
        default:
          for (const [key, val] of Object.entries(node)) {
            node[key] = hydrateNode(val) as never;
          }
          return node;
      }
    }
    for (const [key, val] of Object.entries(node)) {
      if (key !== "type") node[key] = hydrateNode(val) as never;
    }
    return node;
  }
  return value;
}

function assertExpectations(dir: string, expect: Manifest["expect"], result: V10.V10Value | undefined, stdout: string[], evaluationError: unknown) {
  if (!expect) {
    if (evaluationError) {
      throw evaluationError;
    }
    return;
  }
  if (expect.errors && expect.errors.length > 0) {
    if (!evaluationError) {
      throw new Error(`Fixture ${dir} expected errors ${JSON.stringify(expect.errors)} but evaluation succeeded`);
    }
    const message = extractErrorMessage(evaluationError);
    if (!expect.errors.includes(message)) {
      throw new Error(`Fixture ${dir} expected error message in ${JSON.stringify(expect.errors)}, got '${message}'`);
    }
    return;
  }
  if (evaluationError) {
    throw evaluationError;
  }
  if (expect.stdout) {
    if (JSON.stringify(stdout) !== JSON.stringify(expect.stdout)) {
      throw new Error(`Fixture ${dir} expected stdout ${JSON.stringify(expect.stdout)}, got ${JSON.stringify(stdout)}`);
    }
  }
  if (expect.result) {
    const { kind, value } = expect.result;
    if (!result) {
      throw new Error(`Fixture ${dir} expected result kind ${kind}, but evaluation produced no value`);
    }
    if (result.kind !== kind) {
      throw new Error(`Fixture ${dir} expected result kind ${kind}, got ${result.kind}`);
    }
    if (value !== undefined) {
      switch (result.kind) {
        case "string":
        case "bool":
        case "char":
        case "i32":
        case "f64":
          if ((result as any).value !== value) {
            throw new Error(`Fixture ${dir} expected value ${value}, got ${(result as any).value}`);
          }
          break;
        default:
          // For now only support primitive comparisons.
          break;
      }
    }
  }
}

function interceptStdout(buffer: string[], fn: () => void) {
  const original = console.log;
  console.log = (...args: unknown[]) => {
    buffer.push(args.join(" "));
  };
  try {
    fn();
  } finally {
    console.log = original;
  }
}

function ensurePrint(interpreter: V10.InterpreterV10) {
  const globals = interpreter.globals ?? (interpreter as any).globals;
  if (!globals) return;
  try {
    globals.define("print", {
      kind: "native_function",
      name: "print",
      arity: 1,
      impl: (_interp: any, args: any[]) => {
        console.log(args.map(formatValue).join(" "));
        return { kind: "nil", value: null };
      },
    });
  } catch {
    // ignore redefinition
  }
}

function formatValue(value: V10.V10Value): string {
  switch (value.kind) {
    case "string":
    case "char":
      return String(value.value);
    case "bool":
      return value.value ? "true" : "false";
    case "i32":
    case "f64":
      return String(value.value);
    default:
      return `[${value.kind}]`;
  }
}

function extractErrorMessage(err: unknown): string {
  if (!err) return "";
  if (typeof err === "string") return err;
  if (err instanceof Error) {
    const anyErr = err as any;
    if (anyErr.value && typeof anyErr.value === "object" && "message" in anyErr.value) {
      return String(anyErr.value.message);
    }
    return err.message;
  }
  if (typeof err === "object" && err) {
    const anyErr = err as any;
    if (typeof anyErr.message === "string") return anyErr.message;
  }
  return String(err);
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
