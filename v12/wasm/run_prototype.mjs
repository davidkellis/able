import fs from "node:fs/promises";
import fsSync from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

import { createAbleParser, parseSourceToAstModule } from "./ast_adapter.mjs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const DEFAULT_SOURCE_PATH = path.join(__dirname, "samples", "addition.able");
const DEFAULT_WASM_PATH = path.join(__dirname, "ablewasm.wasm");
const DEFAULT_LANGUAGE_WASM_PATH = path.join(
  __dirname,
  "../parser/tree-sitter-able/tree-sitter-able.wasm",
);

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    printHelp();
    return;
  }

  const parser = await createAbleParser(args.languageWasmPath);
  const source = await fs.readFile(args.sourcePath, "utf8");
  const moduleAst = parseSourceToAstModule(parser, source);
  const request = {
    execMode: args.execMode,
    module: moduleAst,
  };

  const evaluate = await loadAbleWasmEvaluator(args.wasmPath);
  const responseRaw = evaluate(JSON.stringify(request));
  const response = JSON.parse(responseRaw);

  process.stdout.write(
    `${JSON.stringify({ request, response }, null, 2)}\n`,
  );
  if (!response.ok) {
    process.exitCode = 1;
  }
}

async function loadAbleWasmEvaluator(wasmPath) {
  const require = createRequire(import.meta.url);
  const wasmExecPath = resolveWasmExecPath();
  require(wasmExecPath);

  if (typeof globalThis.Go !== "function") {
    throw new Error(`Go wasm runtime did not initialize from ${wasmExecPath}`);
  }

  const go = new globalThis.Go();
  const wasmBytes = await fs.readFile(wasmPath);
  const { instance } = await WebAssembly.instantiate(wasmBytes, go.importObject);
  go.run(instance);

  await waitForGlobalFunction("__able_eval_request_json");
  return globalThis.__able_eval_request_json;
}

function resolveWasmExecPath() {
  const goRoot = execFileSync("go", ["env", "GOROOT"], {
    encoding: "utf8",
  }).trim();
  const candidates = [
    path.join(goRoot, "lib", "wasm", "wasm_exec.js"),
    path.join(goRoot, "misc", "wasm", "wasm_exec.js"),
  ];
  for (const candidate of candidates) {
    try {
      fsSync.accessSync(candidate, fsSync.constants.R_OK);
      return candidate;
    } catch {
      // Continue to the next candidate.
    }
  }
  throw new Error(`unable to locate wasm_exec.js under GOROOT=${goRoot}`);
}

async function waitForGlobalFunction(name, timeoutMs = 3000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    const candidate = globalThis[name];
    if (typeof candidate === "function") {
      return;
    }
    await sleep(10);
  }
  throw new Error(`timed out waiting for global function ${name}`);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function parseArgs(argv) {
  const out = {
    sourcePath: DEFAULT_SOURCE_PATH,
    wasmPath: DEFAULT_WASM_PATH,
    languageWasmPath: DEFAULT_LANGUAGE_WASM_PATH,
    execMode: "treewalker",
    help: false,
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    switch (arg) {
      case "--help":
      case "-h":
        out.help = true;
        break;
      case "--source":
        out.sourcePath = resolveArg(argv, ++i, "--source");
        break;
      case "--wasm":
        out.wasmPath = resolveArg(argv, ++i, "--wasm");
        break;
      case "--language-wasm":
        out.languageWasmPath = resolveArg(argv, ++i, "--language-wasm");
        break;
      case "--exec-mode":
        out.execMode = resolveArg(argv, ++i, "--exec-mode");
        break;
      default:
        throw new Error(`unknown argument ${arg}`);
    }
  }

  out.sourcePath = path.resolve(out.sourcePath);
  out.wasmPath = path.resolve(out.wasmPath);
  out.languageWasmPath = path.resolve(out.languageWasmPath);
  return out;
}

function resolveArg(argv, index, flag) {
  if (index >= argv.length) {
    throw new Error(`missing value for ${flag}`);
  }
  return argv[index];
}

function printHelp() {
  process.stdout.write(`Usage: node run_prototype.mjs [options]

Options:
  --source <path>         Able source file to parse and execute.
  --wasm <path>           Path to the compiled ablewasm binary.
  --language-wasm <path>  Path to tree-sitter-able.wasm.
  --exec-mode <mode>      treewalker (default) or bytecode.
  -h, --help              Show this help message.
`);
}

main().catch((err) => {
  process.stderr.write(`error: ${err.message}\n`);
  process.exitCode = 1;
});
