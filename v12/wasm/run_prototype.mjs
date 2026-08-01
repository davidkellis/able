import fs from "node:fs/promises";
import fsSync from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

import { createNodeSourceProvider } from "./node_source_provider.mjs";
import { buildSourceEvaluationRequest } from "./source_request.mjs";

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

  const request = await loadEvaluationRequest(args);

  const hostOutput = installHostOutput();
  const evaluate = await loadAbleWasmEvaluator(args.wasmPath);
  const responseRaw = evaluate(JSON.stringify(request));
  const response = JSON.parse(responseRaw);
  validateHostOutput(args, hostOutput);
  validateResponse(args, response);

  process.stdout.write(
    `${JSON.stringify({ request, response, hostOutput: hostOutput.snapshot() }, null, 2)}\n`,
  );
  // The Go module intentionally remains live to expose its JavaScript
  // callback. This single-request CLI must terminate explicitly after the
  // response has been printed.
  process.exit(args.expectResponseOK === null ? (response.ok ? 0 : 1) : 0);
}

function installHostOutput() {
  const stdout = [];
  const stderr = [];
  globalThis.able_host = {
    write_stdout(message) {
      stdout.push(String(message));
    },
    write_stderr(message) {
      stderr.push(String(message));
    },
  };
  return {
    snapshot() {
      return { stdout: [...stdout], stderr: [...stderr] };
    },
  };
}

function validateHostOutput(args, hostOutput) {
  const output = hostOutput.snapshot();
  const stdout = output.stdout.join("");
  const stderr = output.stderr.join("");
  const expectedStdout = decodeExpectedText(args.expectHostStdout);
  const expectedStderr = decodeExpectedText(args.expectHostStderr);
  if (expectedStdout !== null && stdout !== expectedStdout) {
    throw new Error(`host stdout mismatch: got ${JSON.stringify(stdout)}, want ${JSON.stringify(expectedStdout)}`);
  }
  if (expectedStderr !== null && stderr !== expectedStderr) {
    throw new Error(`host stderr mismatch: got ${JSON.stringify(stderr)}, want ${JSON.stringify(expectedStderr)}`);
  }
}

function validateResponse(args, response) {
  if (args.expectResponseOK !== null && response.ok !== args.expectResponseOK) {
    throw new Error(`response ok mismatch: got ${response.ok}, want ${args.expectResponseOK}`);
  }
}

function decodeExpectedText(value) {
  if (value === null) {
    return null;
  }
  return value
    .replaceAll("\\\\", "\\")
    .replaceAll("\\n", "\n")
    .replaceAll("\\r", "\r")
    .replaceAll("\\t", "\t");
}

async function loadEvaluationRequest(args) {
  if (args.moduleJSONPath) {
    const raw = await fs.readFile(args.moduleJSONPath, "utf8");
    try {
      return {
        execMode: args.execMode,
        setupModules: [],
        module: JSON.parse(raw),
        // Pre-parsed payloads retain the original bridge's diagnostics. Only
        // the browser source-loader path supplies source origins.
        entryOrigin: "",
      };
    } catch (err) {
      throw new Error(`decode module JSON ${args.moduleJSONPath}: ${err.message}`);
    }
  }

  const { createAbleParser, parseSourceToAstModule } = await import("./node_ast_parser.mjs");
  const parser = await createAbleParser(args.languageWasmPath);
  try {
    return await buildSourceEvaluationRequest({
      entryPath: args.sourcePath,
      moduleRoots: args.moduleRoots,
      execMode: args.execMode,
      sourceProvider: createNodeSourceProvider(),
      parseSource(source) {
        return parseSourceToAstModule(parser, source);
      },
    });
  } finally {
    parser.delete();
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
    moduleJSONPath: "",
    wasmPath: DEFAULT_WASM_PATH,
    languageWasmPath: DEFAULT_LANGUAGE_WASM_PATH,
    moduleRoots: [],
    execMode: "treewalker",
    expectHostStdout: null,
    expectHostStderr: null,
    expectResponseOK: null,
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
      case "--module-json":
        out.moduleJSONPath = resolveArg(argv, ++i, "--module-json");
        break;
      case "--wasm":
        out.wasmPath = resolveArg(argv, ++i, "--wasm");
        break;
      case "--language-wasm":
        out.languageWasmPath = resolveArg(argv, ++i, "--language-wasm");
        break;
      case "--module-root":
        out.moduleRoots.push(resolveArg(argv, ++i, "--module-root"));
        break;
      case "--exec-mode":
        out.execMode = resolveArg(argv, ++i, "--exec-mode");
        break;
      case "--expect-host-stdout":
        out.expectHostStdout = resolveArg(argv, ++i, "--expect-host-stdout");
        break;
      case "--expect-host-stderr":
        out.expectHostStderr = resolveArg(argv, ++i, "--expect-host-stderr");
        break;
      case "--expect-response-ok":
        out.expectResponseOK = resolveBooleanArg(argv, ++i, "--expect-response-ok");
        break;
      default:
        throw new Error(`unknown argument ${arg}`);
    }
  }

  out.sourcePath = path.resolve(out.sourcePath);
  if (out.moduleJSONPath) {
    out.moduleJSONPath = path.resolve(out.moduleJSONPath);
  }
  out.wasmPath = path.resolve(out.wasmPath);
  out.languageWasmPath = path.resolve(out.languageWasmPath);
  out.moduleRoots = out.moduleRoots.map((root) => path.resolve(root));
  return out;
}

function resolveArg(argv, index, flag) {
  if (index >= argv.length) {
    throw new Error(`missing value for ${flag}`);
  }
  return argv[index];
}

function resolveBooleanArg(argv, index, flag) {
  const value = resolveArg(argv, index, flag);
  if (value === "true") {
    return true;
  }
  if (value === "false") {
    return false;
  }
  throw new Error(`${flag} must be true or false`);
}

function printHelp() {
  process.stdout.write(`Usage: node run_prototype.mjs [options]

Options:
  --source <path>         Able source file to parse and execute.
  --module-json <path>    Pre-parsed fixture AST module JSON; skips tree-sitter.
  --wasm <path>           Path to the compiled ablewasm binary.
  --language-wasm <path>  Path to tree-sitter-able.wasm.
  --module-root <path>    Extra static-source root; may be repeated.
  --exec-mode <mode>      treewalker (default) or bytecode.
  --expect-host-stdout <text>
                         Require exact concatenated able_host stdout (use \\n).
  --expect-host-stderr <text>
                         Require exact concatenated able_host stderr (use \\n).
  --expect-response-ok <true|false>
                         Require the response success state.
  -h, --help              Show this help message.
`);
}

main().catch((err) => {
  process.stderr.write(`error: ${err.message}\n`);
  process.exitCode = 1;
});
